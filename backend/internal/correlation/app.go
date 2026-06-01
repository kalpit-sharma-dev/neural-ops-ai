package correlation

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/correlation/config"
	"github.com/neuralops/platform/internal/correlation/deployment"
	"github.com/neuralops/platform/internal/correlation/engine"
	"github.com/neuralops/platform/internal/correlation/graph"
	"github.com/neuralops/platform/internal/correlation/repository"
	"github.com/neuralops/platform/internal/correlation/temporal"
	"github.com/neuralops/platform/internal/correlation/transaction"
	"github.com/neuralops/platform/internal/kafka"
	"github.com/neuralops/platform/internal/tracequery"
	"github.com/neuralops/platform/pkg/metrics"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// App runs the correlation engine.
type App struct {
	cfg    *config.Config
	log    *zap.Logger
	engine *engine.Engine
	repo   *repository.Store
}

// NewApp wires the correlation engine.
func NewApp(ctx context.Context, log *zap.Logger) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	repo, err := repository.NewStore(ctx, cfg.Postgres.DSN)
	if err != nil {
		return nil, err
	}

	clickhouseStore, chErr := repository.NewClickHouseStore(ctx, cfg.ClickHouse.DSN, cfg.ClickHouse.TransactionTable)
	if chErr != nil {
		log.Warn("clickhouse unavailable for transaction correlation", zap.Error(chErr))
	}

	var txnCorrelator *transaction.Correlator
	if clickhouseStore != nil {
		txnCorrelator = transaction.NewCorrelator(log, clickhouseStore)
	} else {
		txnCorrelator = transaction.NewCorrelator(log, nil)
	}

	var redisClient *redis.Client
	if cfg.Redis.URL != "" {
		if opts, parseErr := redis.ParseURL(cfg.Redis.URL); parseErr == nil {
			redisClient = redis.NewClient(opts)
		}
	}

	deploymentCorrelator := deployment.NewCorrelator(log, repo, redisClient, cfg.Correlation.DeploymentWatchWindow, cfg.Correlation.DeploymentEvalWindow)
	graphBuilder := graph.NewBuilder(log, repo)
	temporalAnalyzer := temporal.NewAnalyzer(log, repo, cfg.Correlation.TemporalWindow)

	var spanStore *tracequery.SpanStore
	if cfg.ClickHouse.DSN != "" {
		if ss, spanErr := tracequery.NewSpanStore(ctx, cfg.ClickHouse.DSN); spanErr != nil {
			log.Warn("clickhouse span store unavailable", zap.Error(spanErr))
		} else {
			spanStore = ss
		}
	}

	corrEngine := engine.New(log, deploymentCorrelator, graphBuilder, txnCorrelator, temporalAnalyzer, spanStore)

	deploymentCorrelator.StartEvaluator(ctx, time.Minute)
	graphBuilder.StartFlusher(ctx, cfg.Correlation.GraphFlushInterval)
	temporalAnalyzer.StartFlusher(ctx)
	if clickhouseStore != nil {
		txnCorrelator.StartFlusher(ctx, time.Minute, 2*time.Minute)
	}

	eventConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: cfg.Kafka.Brokers, GroupID: cfg.Kafka.ConsumerGroup + "-events",
		Topics: cfg.Kafka.EventTopics, Concurrency: cfg.Kafka.Concurrency, ServiceName: "correlation",
	}, log, func(ctx context.Context, message kafka.Message) error {
		return corrEngine.HandleEvent(ctx, message.Value)
	})

	traceConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: cfg.Kafka.Brokers, GroupID: cfg.Kafka.ConsumerGroup + "-traces",
		Topics: cfg.Kafka.TraceTopics, Concurrency: cfg.Kafka.Concurrency, ServiceName: "correlation",
	}, log, func(ctx context.Context, message kafka.Message) error {
		return corrEngine.HandleTrace(ctx, message.Value)
	})

	logConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: cfg.Kafka.Brokers, GroupID: cfg.Kafka.ConsumerGroup + "-logs",
		Topics: cfg.Kafka.LogTopics, Concurrency: cfg.Kafka.Concurrency, ServiceName: "correlation",
	}, log, func(ctx context.Context, message kafka.Message) error {
		return corrEngine.HandleLog(ctx, message.Value)
	})

	go func() { _ = eventConsumer.Run(ctx) }()
	go func() { _ = traceConsumer.Run(ctx) }()
	go func() { _ = logConsumer.Run(ctx) }()

	return &App{cfg: cfg, log: log, engine: corrEngine, repo: repo}, nil
}

// Run starts the HTTP server for health/metrics.
func (a *App) Run(ctx context.Context) error {
	if a.cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	metrics.RegisterRoutes(router, "correlation")
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

	a.log.Info("correlation engine started", zap.Int("port", a.cfg.Server.Port))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Close releases resources.
func (a *App) Close() {
	if a.repo != nil {
		a.repo.Close()
	}
}

func healthHandler(field string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{"service": "correlation", field: true},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}
