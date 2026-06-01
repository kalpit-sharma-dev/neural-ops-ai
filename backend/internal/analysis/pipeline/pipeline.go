package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/ai"
	"github.com/neuralops/platform/internal/analysis/anomaly"
	"github.com/neuralops/platform/internal/analysis/cache"
	"github.com/neuralops/platform/internal/analysis/classifier"
	analysismetrics "github.com/neuralops/platform/internal/analysis/metrics"
	"github.com/neuralops/platform/internal/analysis/repository"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/ingestion/dto"
	"github.com/neuralops/platform/internal/kafka"
	"go.uber.org/zap"
)

// Pipeline processes logs and metrics from Kafka.
type Pipeline struct {
	log          *zap.Logger
	classifier   *classifier.Service
	llm          ai.LLMClient
	llmBackend   string
	cache        *cache.RedisCache
	postgres     *repository.PostgresStore
	elasticsearch *repository.ElasticsearchStore
	qdrant       *repository.QdrantStore
	anomaly      *anomaly.Detector
	producer     *kafka.Producer
	anomalyTopic string
	explanationTTL time.Duration
	recurringWindow int
	embedBatchSize  int

	embedMu     sync.Mutex
	embedBuffer []embedItem
}

type embedItem struct {
	entry   domain.LogEntry
	payload map[string]any
}

// Config configures the analysis pipeline.
type Config struct {
	LLM             ai.LLMClient
	LLMBackend      string
	Classifier      *classifier.Service
	Cache           *cache.RedisCache
	Postgres        *repository.PostgresStore
	Elasticsearch   *repository.ElasticsearchStore
	Qdrant          *repository.QdrantStore
	Anomaly         *anomaly.Detector
	Producer        *kafka.Producer
	AnomalyTopic    string
	ExplanationTTL  time.Duration
	RecurringWindow int
	EmbedBatchSize  int
}

// New creates an analysis pipeline.
func New(log *zap.Logger, cfg Config) *Pipeline {
	if cfg.EmbedBatchSize <= 0 {
		cfg.EmbedBatchSize = 100
	}
	if cfg.RecurringWindow <= 0 {
		cfg.RecurringWindow = 100
	}
	if cfg.ExplanationTTL <= 0 {
		cfg.ExplanationTTL = time.Hour
	}
	return &Pipeline{
		log:             log,
		classifier:      cfg.Classifier,
		llm:             cfg.LLM,
		llmBackend:      cfg.LLMBackend,
		cache:           cfg.Cache,
		postgres:        cfg.Postgres,
		elasticsearch:   cfg.Elasticsearch,
		qdrant:          cfg.Qdrant,
		anomaly:         cfg.Anomaly,
		producer:        cfg.Producer,
		anomalyTopic:    cfg.AnomalyTopic,
		explanationTTL:  cfg.ExplanationTTL,
		recurringWindow: cfg.RecurringWindow,
		embedBatchSize:  cfg.EmbedBatchSize,
	}
}

// ProcessLogMessage handles a Kafka log message.
func (p *Pipeline) ProcessLogMessage(ctx context.Context, raw []byte) error {
	start := time.Now()
	defer func() {
		analysismetrics.ObserveAnalysisLatency("log", time.Since(start))
	}()

	var enriched dto.EnrichedLog
	if err := json.Unmarshal(raw, &enriched); err != nil {
		var entry domain.LogEntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			return fmt.Errorf("unmarshal log: %w", err)
		}
		enriched.LogEntry = entry
	}

	entry := enriched.LogEntry
	classification, err := p.classifier.Classify(ctx, entry.Message)
	if err != nil {
		return err
	}
	entry.ClassifiedError = classification

	var explanation *ai.LogExplanation
	if entry.IsErrorSeverity() && classification != nil {
		explanation, err = p.generateExplanation(ctx, entry, classification)
		if err != nil {
			p.log.Warn("explanation generation failed", zap.Error(err))
		}
	}

	if entry.ParsedStackTrace != nil {
		if err := p.analyzeStackTrace(ctx, *entry.ParsedStackTrace, entry.Service); err != nil {
			p.log.Warn("stack trace analysis failed", zap.Error(err))
		}
	}

	embeddingID := entry.ID.String()
	tenantID := enriched.IngestionMetadata.TenantID
	if tenantID == "" {
		tenantID = "00000000-0000-0000-0000-000000000002"
	}
	if err := p.queueEmbedding(entry, tenantID); err != nil {
		p.log.Warn("queue embedding failed", zap.Error(err))
	}

	if p.postgres != nil {
		if err := p.postgres.SaveClassification(ctx, entry.ID, entry.Service, classification); err != nil {
			p.log.Warn("save classification failed", zap.Error(err))
		}
		if err := p.postgres.SaveExplanation(ctx, entry.ID, entry.Service, explanation); err != nil {
			p.log.Warn("save explanation failed", zap.Error(err))
		}
	}

	if p.elasticsearch != nil {
		if err := p.elasticsearch.IndexEnrichedLog(ctx, tenantID, entry, classification, explanation, embeddingID); err != nil {
			p.log.Warn("index enriched log failed", zap.Error(err))
		}
	}

	if p.cache != nil {
		_ = p.cache.UpdateServiceErrorRate(ctx, entry.Service, entry.IsErrorSeverity())
	}

	if classification != nil {
		analysismetrics.IncAnalysisProcessed(entry.Service, string(classification.Category))
	}

	enrichedPayload, _ := json.Marshal(map[string]any{
		"log":            entry,
		"classification": classification,
		"explanation":    explanation,
		"embeddingId":    embeddingID,
	})
	return p.producer.Publish(ctx, kafka.Message{
		Topic: "enriched-logs",
		Key:   kafka.PartitionKey(entry.Service, string(entry.Environment)),
		Value: enrichedPayload,
	})
}

// ProcessMetricMessage handles a Kafka metric message.
func (p *Pipeline) ProcessMetricMessage(ctx context.Context, raw []byte) error {
	start := time.Now()
	defer func() {
		analysismetrics.ObserveAnalysisLatency("metric", time.Since(start))
	}()

	var metric domain.Metric
	if err := json.Unmarshal(raw, &metric); err != nil {
		return fmt.Errorf("unmarshal metric: %w", err)
	}

	if p.anomaly == nil {
		return nil
	}

	result, err := p.anomaly.AnalyzeMetric(ctx, metric)
	if err != nil {
		return err
	}
	if result != nil && result.Detected {
		analysismetrics.IncAnalysisProcessed(metric.ServiceName, "ANOMALY")
	}
	return nil
}

// FlushEmbeddings processes queued embedding batches.
func (p *Pipeline) FlushEmbeddings(ctx context.Context) error {
	p.embedMu.Lock()
	if len(p.embedBuffer) == 0 {
		p.embedMu.Unlock()
		return nil
	}
	batch := p.embedBuffer
	p.embedBuffer = nil
	p.embedMu.Unlock()

	if p.llm == nil || p.qdrant == nil {
		return nil
	}

	texts := make([]string, 0, len(batch))
	logIDs := make([]uuid.UUID, 0, len(batch))
	payloads := make([]map[string]any, 0, len(batch))
	for _, item := range batch {
		texts = append(texts, item.entry.Message)
		logIDs = append(logIDs, item.entry.ID)
		payloads = append(payloads, item.payload)
	}

	start := time.Now()
	analysismetrics.IncLLMAPICall(p.llmBackend, "Embed")
	vectors, err := p.llm.Embed(ctx, texts)
	analysismetrics.ObserveLLMLatency(time.Since(start))
	if err != nil {
		return err
	}

	_, err = p.qdrant.UpsertEmbeddings(ctx, logIDs, vectors, payloads)
	return err
}

// StartEmbedFlusher periodically flushes embedding batches.
func (p *Pipeline) StartEmbedFlusher(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				_ = p.FlushEmbeddings(context.Background())
				return
			case <-ticker.C:
				if err := p.FlushEmbeddings(ctx); err != nil {
					p.log.Warn("flush embeddings failed", zap.Error(err))
				}
			}
		}
	}()
}

func (p *Pipeline) generateExplanation(ctx context.Context, entry domain.LogEntry, classification *domain.ErrorClassification) (*ai.LogExplanation, error) {
	cacheKey := hashKey(string(classification.Category) + ":" + entry.Service + ":" + entry.Message)
	if p.cache != nil {
		if cached, err := p.cache.GetExplanation(ctx, entry.Service, cacheKey); err == nil && cached != nil {
			return cached, nil
		}
	}

	if p.llm == nil {
		return nil, fmt.Errorf("llm not configured")
	}

	start := time.Now()
	analysismetrics.IncLLMAPICall(p.llmBackend, "Explain")
	explanation, err := p.llm.Explain(ctx, entry.Message, classification)
	analysismetrics.ObserveLLMLatency(time.Since(start))
	if err != nil {
		return nil, err
	}

	if p.cache != nil {
		_ = p.cache.SetExplanation(ctx, entry.Service, cacheKey, explanation, p.explanationTTL)
	}
	return explanation, nil
}

func (p *Pipeline) analyzeStackTrace(ctx context.Context, trace domain.StackTrace, service string) error {
	fingerprint := hashKey(trace.ExceptionType + ":" + trace.Hotspot)
	if p.cache != nil {
		if count, err := p.cache.TrackException(ctx, service, fingerprint, p.recurringWindow); err == nil && count > 1 {
			p.log.Info("recurring exception detected", zap.String("service", service), zap.Int64("count", count), zap.String("fingerprint", fingerprint))
		}
	}

	if p.llm == nil {
		return nil
	}

	start := time.Now()
	analysismetrics.IncLLMAPICall(p.llmBackend, "SummarizeStackTrace")
	_, err := p.llm.SummarizeStackTrace(ctx, trace)
	analysismetrics.ObserveLLMLatency(time.Since(start))
	return err
}

func (p *Pipeline) queueEmbedding(entry domain.LogEntry, tenantID string) error {
	p.embedMu.Lock()
	defer p.embedMu.Unlock()

	p.embedBuffer = append(p.embedBuffer, embedItem{
		entry: entry,
		payload: map[string]any{
			"log_id":    entry.ID.String(),
			"tenantId":  tenantID,
			"service":   entry.Service,
			"severity":  entry.Severity,
			"timestamp": entry.Timestamp.UTC().Format(time.RFC3339),
		},
	})

	if len(p.embedBuffer) >= p.embedBatchSize {
		go func() {
			_ = p.FlushEmbeddings(context.Background())
		}()
	}
	return nil
}

func hashKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// AnomalyPublisher publishes anomalies to Kafka.
type AnomalyPublisher struct {
	producer *kafka.Producer
	topic    string
}

// NewAnomalyPublisher creates an anomaly publisher.
func NewAnomalyPublisher(producer *kafka.Producer, topic string) *AnomalyPublisher {
	return &AnomalyPublisher{producer: producer, topic: topic}
}

// PublishAnomaly publishes an anomaly event.
func (p *AnomalyPublisher) PublishAnomaly(ctx context.Context, item domain.AnomalyDetection) error {
	payload, err := anomaly.AnomalyEventJSON(item)
	if err != nil {
		return err
	}
	return p.producer.Publish(ctx, kafka.Message{
		Topic: p.topic,
		Key:   kafka.PartitionKey(item.ServiceName, string(item.MetricType)),
		Value: payload,
	})
}
