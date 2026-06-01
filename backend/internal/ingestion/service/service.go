package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/ingestion/config"
	"github.com/neuralops/platform/internal/ingestion/dto"
	"github.com/neuralops/platform/internal/ingestion/parser"
	"github.com/neuralops/platform/internal/ingestion/ratelimit"
	"github.com/neuralops/platform/internal/ingestion/validator"
	"github.com/neuralops/platform/internal/kafka"
	"github.com/neuralops/platform/internal/apm"
	"github.com/neuralops/platform/internal/security"
	"github.com/neuralops/platform/internal/storage"
	"go.uber.org/zap"
)

// Service orchestrates ingestion, enrichment, and publishing.
type Service struct {
	cfg       *config.Config
	log       *zap.Logger
	enricher  *parser.Enricher
	limiter   *ratelimit.Limiter
	producer  *kafka.Producer
	metrics   *storage.ClickHouseWriter
	policies  *apm.PolicyStore
}

// SetPolicyStore attaches trace sampling policies.
func (s *Service) SetPolicyStore(store *apm.PolicyStore) {
	s.policies = store
}

// New creates an ingestion service.
func New(
	cfg *config.Config,
	log *zap.Logger,
	enricher *parser.Enricher,
	limiter *ratelimit.Limiter,
	producer *kafka.Producer,
	metrics *storage.ClickHouseWriter,
) *Service {
	return &Service{
		cfg:      cfg,
		log:      log,
		enricher: enricher,
		limiter:  limiter,
		producer: producer,
		metrics:  metrics,
	}
}

// IngestLogs ingests one or more log records.
func (s *Service) IngestLogs(ctx context.Context, requests []dto.LogIngestRequest, source string) (accepted, rejected int, err error) {
	for _, req := range requests {
		if !s.limiter.Allow(tenantID(req.TenantID), plan(req.Plan)) {
			rejected++
			continue
		}

		if err := validator.ValidateLog(&req, s.cfg.Ingestion.TimestampSkew); err != nil {
			rejected++
			s.log.Warn("log validation failed", zap.Error(err))
			continue
		}

		enriched := s.enricher.Enrich(&req, source)
		enriched.LogEntry.Message = security.MaskMessage(enriched.LogEntry.Message)
		payload, marshalErr := json.Marshal(enriched)
		if marshalErr != nil {
			rejected++
			s.log.Error("marshal enriched log failed", zap.Error(marshalErr))
			continue
		}

		key := kafka.PartitionKey(enriched.Service, string(enriched.Environment))
		if pubErr := s.producer.Publish(ctx, kafka.Message{
			Topic: s.cfg.Topic("raw-logs"),
			Key:   key,
			Value: payload,
		}); pubErr != nil {
			return accepted, rejected, pubErr
		}
		accepted++
	}
	return accepted, rejected, nil
}

// IngestMetrics ingests metrics to Kafka and ClickHouse.
func (s *Service) IngestMetrics(ctx context.Context, requests []dto.MetricIngestRequest) (accepted, rejected int, err error) {
	for _, req := range requests {
		if !s.limiter.Allow(tenantID(req.TenantID), plan(req.Plan)) {
			rejected++
			continue
		}

		if err := validator.ValidateMetric(&req); err != nil {
			rejected++
			continue
		}

		metric := domain.Metric{
			ServiceName: req.ServiceName,
			Host:        req.Host,
			Pod:         req.Pod,
			MetricType:  domain.MetricType(strings.ToUpper(req.MetricType)),
			Value:       req.Value,
			Timestamp:   req.Timestamp.UTC(),
			Labels:      req.Labels,
		}

		payload, marshalErr := json.Marshal(metric)
		if marshalErr != nil {
			rejected++
			continue
		}

		key := kafka.PartitionKey(metric.ServiceName, "metrics")
		if pubErr := s.producer.Publish(ctx, kafka.Message{
			Topic: s.cfg.Topic("raw-metrics"),
			Key:   key,
			Value: payload,
		}); pubErr != nil {
			return accepted, rejected, pubErr
		}

		if s.metrics != nil {
			if writeErr := s.metrics.Write(storage.MetricRow{
				TenantID:    tenantID(req.TenantID),
				ServiceName: metric.ServiceName,
				Host:        metric.Host,
				Pod:         metric.Pod,
				MetricType:  metric.MetricType,
				Value:       metric.Value,
				Timestamp:   metric.Timestamp,
				Labels:      metric.Labels,
			}); writeErr != nil {
				s.log.Warn("clickhouse metric write failed", zap.Error(writeErr))
			}
		}

		accepted++
	}
	return accepted, rejected, nil
}

// IngestEvents ingests deployment and configuration events.
func (s *Service) IngestEvents(ctx context.Context, requests []dto.EventIngestRequest) (accepted, rejected int, err error) {
	for _, req := range requests {
		if !s.limiter.Allow(tenantID(req.TenantID), plan(req.Plan)) {
			rejected++
			continue
		}
		if err := validator.ValidateEvent(&req); err != nil {
			rejected++
			continue
		}

		eventID := uuid.New()
		if req.ID != "" {
			if parsed, parseErr := uuid.Parse(req.ID); parseErr == nil {
				eventID = parsed
			}
		}

		deployment := domain.Deployment{
			ID:          eventID,
			Service:     req.Service,
			Version:     req.Version,
			DeployedAt:  req.DeployedAt.UTC(),
			DeployedBy:  req.DeployedBy,
			ChangeType:  domain.ChangeType(strings.ToUpper(req.ChangeType)),
			Environment: domain.Environment(req.Environment),
			CreatedAt:   time.Now().UTC(),
		}

		payload, marshalErr := json.Marshal(deployment)
		if marshalErr != nil {
			rejected++
			continue
		}

		key := kafka.PartitionKey(deployment.Service, string(deployment.Environment))
		if pubErr := s.producer.Publish(ctx, kafka.Message{
			Topic: s.cfg.Topic("raw-events"),
			Key:   key,
			Value: payload,
		}); pubErr != nil {
			return accepted, rejected, pubErr
		}
		accepted++
	}
	return accepted, rejected, nil
}

// IngestTraces ingests distributed trace spans with head/tail sampling.
func (s *Service) IngestTraces(ctx context.Context, requests []dto.TraceSpanRequest) (accepted, rejected int, err error) {
	sampled := make(map[string]bool)
	for _, req := range requests {
		if !s.limiter.Allow(tenantID(req.TenantID), plan(req.Plan)) {
			rejected++
			continue
		}
		if err := validator.ValidateTrace(&req); err != nil {
			rejected++
			continue
		}

		isError := strings.EqualFold(req.Status, "ERROR") || strings.EqualFold(req.Status, "error")
		rate := 1.0
		if s.policies != nil {
			p := s.policies.Get(ctx, tenantID(req.TenantID))
			rate = p.HeadSampleRate
			if isError {
				rate = p.TailSampleRate
			}
		}
		if _, ok := sampled[req.TraceID]; !ok {
			sampled[req.TraceID] = apm.ShouldSample(req.TraceID, isError, rate)
		}
		if !sampled[req.TraceID] {
			rejected++
			continue
		}

		payload, marshalErr := json.Marshal(req)
		if marshalErr != nil {
			rejected++
			continue
		}

		key := kafka.PartitionKey(req.Service, req.TraceID)
		if pubErr := s.producer.Publish(ctx, kafka.Message{
			Topic: s.cfg.Topic("raw-traces"),
			Key:   key,
			Value: payload,
		}); pubErr != nil {
			return accepted, rejected, pubErr
		}
		accepted++
	}
	return accepted, rejected, nil
}

// IngestAlertWebhook ingests third-party alert webhooks as events.
func (s *Service) IngestAlertWebhook(ctx context.Context, req dto.AlertWebhookRequest) error {
	if !s.limiter.Allow(tenantID(req.TenantID), plan(req.Plan)) {
		return domain.NewValidationError("rate_limit", "tenant rate limit exceeded")
	}

	alert := domain.Alert{
		ID:           uuid.New(),
		Source:       domain.AlertSource(strings.ToUpper(req.Source)),
		Title:        req.Title,
		Description:  req.Description,
		Severity:     domain.IncidentSeverity(strings.ToUpper(req.Severity)),
		FiredAt:      req.FiredAt.UTC(),
		Labels:       req.Labels,
		Deduplicated: false,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	payload, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	return s.producer.Publish(ctx, kafka.Message{
		Topic: s.cfg.Topic("raw-events"),
		Key:   kafka.PartitionKey(req.Source, req.Severity),
		Value: payload,
	})
}

func tenantID(id string) string {
	if id == "" {
		return "default"
	}
	return id
}

func plan(planName string) string {
	if planName == "" {
		return "startup"
	}
	return strings.ToLower(planName)
}
