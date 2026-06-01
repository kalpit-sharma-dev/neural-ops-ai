package ingestion

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/apm"
	"github.com/neuralops/platform/internal/ingestion/config"
	"github.com/neuralops/platform/internal/ingestion/connector"
	"github.com/neuralops/platform/internal/ingestion/handler"
	"github.com/neuralops/platform/internal/ingestion/parser"
	"github.com/neuralops/platform/internal/ingestion/ratelimit"
	"github.com/neuralops/platform/internal/ingestion/service"
	"github.com/neuralops/platform/internal/platform/db"
	"github.com/neuralops/platform/internal/platform/middleware"
	"github.com/neuralops/platform/internal/kafka"
	"github.com/neuralops/platform/internal/storage"
	"github.com/neuralops/platform/pkg/metrics"
	"go.uber.org/zap"
)

// App runs the ingestion HTTP server and background connectors.
type App struct {
	cfg       *config.Config
	log       *zap.Logger
	handler   *handler.Handler
	producer  *kafka.Producer
	metrics   *storage.ClickHouseWriter
	connectors *connector.Manager
}

// NewApp wires ingestion dependencies.
func NewApp(ctx context.Context, log *zap.Logger) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	enricher, err := parser.NewEnricher(
		cfg.Parser,
		cfg.Ingestion.IngestorID,
		cfg.Ingestion.Datacenter,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create enricher: %w", err)
	}

	limiter := ratelimit.New(cfg.RateLimit.PerTenant)

	producer := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:    cfg.Kafka.Brokers,
		BatchSize:  cfg.Kafka.BatchSize,
		FlushEvery: cfg.Kafka.FlushInterval,
		DLQTopic:   cfg.Topic(cfg.Kafka.DLQTopic),
	}, log)

	var metricsWriter *storage.ClickHouseWriter
	metricsWriter, err = storage.NewClickHouseWriter(ctx, storage.ClickHouseConfig{
		DSN:           cfg.ClickHouse.DSN,
		Table:         cfg.ClickHouse.MetricsTable,
		MaxBatchSize:  cfg.ClickHouse.MaxBatchSize,
		FlushInterval: cfg.ClickHouse.FlushInterval,
	}, log)
	if err != nil {
		log.Warn("clickhouse writer unavailable, metrics direct-write disabled", zap.Error(err))
		metricsWriter = nil
	}

	svc := service.New(cfg, log, enricher, limiter, producer, metricsWriter)
	if dsn := os.Getenv("POSTGRES_DSN"); dsn != "" {
		if pool, err := db.NewPool(ctx, dsn); err == nil {
			svc.SetPolicyStore(apm.NewPolicyStore(pool))
			log.Info("trace sampling policies enabled")
		}
	}
	h := handler.New(cfg, log, svc)
	connMgr := connector.NewManager(cfg.Connectors, log, svc)

	return &App{
		cfg:        cfg,
		log:        log,
		handler:    h,
		producer:   producer,
		metrics:    metricsWriter,
		connectors: connMgr,
	}, nil
}

// Run starts the ingestion service until context cancellation.
func (a *App) Run(ctx context.Context) error {
	if a.cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.ExtractTenant("default"))
	metrics.RegisterRoutes(router, "ingestion")

	router.GET("/health", a.healthHandler("healthy"))
	router.GET("/ready", a.healthHandler("ready"))
	router.GET("/live", a.healthHandler("alive"))

	a.handler.RegisterRoutes(router)

	a.connectors.Start(ctx)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", a.cfg.Server.Port),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			a.log.Error("server shutdown failed", zap.Error(err))
		}
	}()

	a.log.Info("ingestion service started",
		zap.Int("port", a.cfg.Server.Port),
		zap.Strings("kafka_brokers", a.cfg.Kafka.Brokers),
		zap.String("clickhouse", storage.ParseDSNHost(a.cfg.ClickHouse.DSN)),
	)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

// Close releases ingestion resources.
func (a *App) Close() error {
	var err error
	if a.metrics != nil {
		if closeErr := a.metrics.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if a.producer != nil {
		if closeErr := a.producer.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

func (a *App) healthHandler(field string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"service": "ingestion",
				field:     true,
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}
