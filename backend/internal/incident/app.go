package incident

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/ai"
	"github.com/neuralops/platform/internal/incident/config"
	"github.com/neuralops/platform/internal/incident/engine"
	"github.com/neuralops/platform/internal/incident/handler"
	"github.com/neuralops/platform/internal/incident/repository"
	"github.com/neuralops/platform/internal/kafka"
	"github.com/neuralops/platform/internal/platform/health"
	"github.com/neuralops/platform/pkg/metrics"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// App runs the incident engine.
type App struct {
	cfg    *config.Config
	log    *zap.Logger
	repo   *repository.Store
	router *gin.Engine
}

// NewApp wires the incident engine.
func NewApp(ctx context.Context, log *zap.Logger) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	repo, err := repository.NewStore(ctx, cfg.Postgres.DSN)
	if err != nil {
		return nil, err
	}

	var txnStore *repository.TransactionStore
	txnStore, err = repository.NewTransactionStore(ctx, cfg.ClickHouse.DSN, cfg.ClickHouse.TransactionTable)
	if err != nil {
		log.Warn("clickhouse transaction store unavailable", zap.Error(err))
	}

	llmClient, err := ai.NewLLMClient(cfg.LLMClientConfig())
	if err != nil {
		return nil, err
	}

	var redisClient *redis.Client
	if cfg.Redis.URL != "" {
		if opts, parseErr := redis.ParseURL(cfg.Redis.URL); parseErr == nil {
			redisClient = redis.NewClient(opts)
		}
	}

	incidentService := engine.NewService(log, cfg, repo, llmClient, redisClient)
	h := handler.New(log, incidentService, repo, txnStore, "00000000-0000-0000-0000-000000000002")

	logConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: cfg.Kafka.Brokers, GroupID: cfg.Kafka.ConsumerGroup + "-logs",
		Topics: cfg.Kafka.LogTopics, Concurrency: cfg.Kafka.Concurrency, ServiceName: "incident",
	}, log, func(ctx context.Context, message kafka.Message) error {
		return incidentService.HandleLog(ctx, message.Value)
	})

	anomalyConsumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: cfg.Kafka.Brokers, GroupID: cfg.Kafka.ConsumerGroup + "-anomalies",
		Topics: cfg.Kafka.AnomalyTopics, Concurrency: cfg.Kafka.Concurrency, ServiceName: "incident",
	}, log, func(ctx context.Context, message kafka.Message) error {
		return incidentService.HandleAnomaly(ctx, message.Value)
	})

	go func() { _ = logConsumer.Run(ctx) }()
	go func() { _ = anomalyConsumer.Run(ctx) }()

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	metrics.RegisterRoutes(router, "incident")

	healthChecker := health.NewChecker("incident", cfg.Server.Environment, health.DefaultVersion)
	healthChecker.Register("postgres", health.PingCheck(repo.Pool()))
	if redisClient != nil {
		healthChecker.Register("redis", health.RedisCheck(redisClient))
	}
	router.GET("/health", healthChecker.GinHandler())
	router.GET("/ready", healthChecker.GinHandler())
	router.GET("/live", simpleLiveHandler("incident"))
	h.RegisterRoutes(router)

	return &App{cfg: cfg, log: log, repo: repo, router: router}, nil
}

// Run starts the HTTP server until context cancellation.
func (a *App) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", a.cfg.Server.Port),
		Handler:           a.router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	a.log.Info("incident engine started", zap.Int("port", a.cfg.Server.Port))
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

func simpleLiveHandler(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{"service": service, "alive": true},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}
