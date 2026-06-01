package analysis

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/ai"
	"github.com/neuralops/platform/internal/analysis/anomaly"
	"github.com/neuralops/platform/internal/analysis/cache"
	"github.com/neuralops/platform/internal/analysis/classifier"
	"github.com/neuralops/platform/internal/analysis/config"
	"github.com/neuralops/platform/internal/analysis/pipeline"
	"github.com/neuralops/platform/internal/analysis/repository"
	"github.com/neuralops/platform/internal/kafka"
	"github.com/neuralops/platform/internal/platform/db"
	"github.com/neuralops/platform/pkg/metrics"
	"go.uber.org/zap"
)

// App runs the AI analysis engine.
type App struct {
	cfg            *config.Config
	log            *zap.Logger
	pipeline       *pipeline.Pipeline
	logConsumer    *kafka.Consumer
	metricConsumer *kafka.Consumer
	producer       *kafka.Producer
	redis          *cache.RedisCache
	postgres       *repository.PostgresStore
}

// NewApp wires the analysis engine.
func NewApp(ctx context.Context, log *zap.Logger) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	llmClient, err := ai.NewLLMClient(cfg.LLMClientConfig())
	if err != nil {
		return nil, err
	}

	redisCache, err := cache.NewRedisCache(cfg.Redis.URL)
	if err != nil {
		log.Warn("redis unavailable, continuing without cache", zap.Error(err))
	}

	postgresStore, err := repository.NewPostgresStore(ctx, cfg.Postgres.DSN)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	esStore, err := repository.NewElasticsearchStore(cfg.Elasticsearch.URL, cfg.Elasticsearch.Index)
	if err != nil {
		log.Warn("elasticsearch unavailable, continuing without indexing", zap.Error(err))
		esStore = nil
	}

	qdrantStore := repository.NewQdrantStore(cfg.Qdrant.URL, cfg.Qdrant.Collection, cfg.Qdrant.VectorSize)
	if err := db.EnsureQdrantCollectionFromMigration(ctx, cfg.Qdrant.URL, ""); err != nil {
		log.Warn("qdrant collection migration skipped", zap.Error(err))
	} else if err := qdrantStore.EnsureCollection(ctx); err != nil {
		log.Warn("qdrant unavailable, continuing without embeddings", zap.Error(err))
		qdrantStore = nil
	}

	producer := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:    cfg.Kafka.Brokers,
		BatchSize:  500,
		FlushEvery: time.Second,
		DLQTopic:   "dlq-analysis",
	}, log)

	anomalyPublisher := pipeline.NewAnomalyPublisher(producer, cfg.Kafka.AnomalyTopic)
	var anomalyDetector *anomaly.Detector
	if redisCache != nil {
		anomalyDetector = anomaly.NewDetector(cfg.Analysis.AnomalyZScoreThreshold, redisCache, anomalyPublisher)
	}

	classifierService := classifier.NewService(
		llmClient,
		redisCache,
		cfg.Analysis.LLMConfidenceThreshold,
		cfg.LLM.Provider,
		func(method string) {
			// metrics recorded inside pipeline methods
		},
	)

	pipe := pipeline.New(log, pipeline.Config{
		LLM:             llmClient,
		LLMBackend:      cfg.LLM.Provider,
		Classifier:      classifierService,
		Cache:           redisCache,
		Postgres:        postgresStore,
		Elasticsearch:   esStore,
		Qdrant:          qdrantStore,
		Anomaly:         anomalyDetector,
		Producer:        producer,
		AnomalyTopic:    cfg.Kafka.AnomalyTopic,
		ExplanationTTL:  cfg.Redis.ExplanationTTL,
		RecurringWindow: cfg.Analysis.RecurringExceptionWindow,
		EmbedBatchSize:  cfg.Analysis.EmbedBatchSize,
	})
	pipe.StartEmbedFlusher(ctx)

	logConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers:     cfg.Kafka.Brokers,
		GroupID:     cfg.LogsConsumerGroup(),
		Topics:      cfg.Kafka.LogTopics,
		Concurrency: cfg.Kafka.Concurrency,
		ServiceName: "analysis",
	}, log, func(ctx context.Context, message kafka.Message) error {
		return pipe.ProcessLogMessage(ctx, message.Value)
	})

	metricConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers:     cfg.Kafka.Brokers,
		GroupID:     cfg.MetricsConsumerGroup(),
		Topics:      cfg.Kafka.MetricTopics,
		Concurrency: cfg.Kafka.Concurrency,
		ServiceName: "analysis",
	}, log, func(ctx context.Context, message kafka.Message) error {
		return pipe.ProcessMetricMessage(ctx, message.Value)
	})

	return &App{
		cfg:            cfg,
		log:            log,
		pipeline:       pipe,
		logConsumer:    logConsumer,
		metricConsumer: metricConsumer,
		producer:       producer,
		redis:          redisCache,
		postgres:       postgresStore,
	}, nil
}

// Run starts HTTP health endpoints and Kafka consumers.
func (a *App) Run(ctx context.Context) error {
	if a.cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	metrics.RegisterRoutes(router, "analysis")
	router.GET("/health", healthHandler("healthy"))
	router.GET("/ready", healthHandler("ready"))
	router.GET("/live", healthHandler("alive"))
	router.GET("/api/v1/health", healthHandler("healthy"))

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", a.cfg.Server.Port),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	go func() {
		a.log.Info("analysis log consumer started", zap.String("group", a.cfg.LogsConsumerGroup()))
		if err := a.logConsumer.Run(ctx); err != nil && ctx.Err() == nil {
			a.log.Error("log consumer stopped", zap.Error(err))
		}
	}()

	go func() {
		a.log.Info("analysis metric consumer started", zap.String("group", a.cfg.MetricsConsumerGroup()))
		if err := a.metricConsumer.Run(ctx); err != nil && ctx.Err() == nil {
			a.log.Error("metric consumer stopped", zap.Error(err))
		}
	}()

	a.log.Info("analysis service started", zap.Int("port", a.cfg.Server.Port))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Close releases resources.
func (a *App) Close() error {
	var err error
	if a.producer != nil {
		if closeErr := a.producer.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if a.redis != nil {
		if closeErr := a.redis.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	if a.postgres != nil {
		a.postgres.Close()
	}
	return err
}

func healthHandler(field string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"service": "analysis",
				field:     true,
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}
