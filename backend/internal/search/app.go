package search

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/ai"
	correlationrepo "github.com/neuralops/platform/internal/correlation/repository"
	"github.com/neuralops/platform/internal/search/config"
	"github.com/neuralops/platform/internal/search/elasticsearch"
	"github.com/neuralops/platform/internal/search/handler"
	"github.com/neuralops/platform/internal/search/qdrant"
	"github.com/neuralops/platform/internal/search/service"
	"github.com/neuralops/platform/internal/platform/health"
	"github.com/neuralops/platform/pkg/metrics"
	"go.uber.org/zap"
)

// App runs the search service.
type App struct {
	cfg      *config.Config
	log      *zap.Logger
	router   *gin.Engine
	txnStore *correlationrepo.ClickHouseStore
}

// NewApp wires the search service.
func NewApp(ctx context.Context, log *zap.Logger) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	esClient, err := elasticsearch.NewClient(cfg.Elasticsearch, log)
	if err != nil {
		log.Warn("elasticsearch client unavailable", zap.Error(err))
	}

	qdrantClient := qdrant.NewClient(cfg.Qdrant, log)

	var txnStore *correlationrepo.ClickHouseStore
	if cfg.ClickHouse.DSN != "" {
		txnStore, err = correlationrepo.NewClickHouseStore(ctx, cfg.ClickHouse.DSN, cfg.ClickHouse.TransactionTable)
		if err != nil {
			log.Warn("clickhouse transaction store unavailable", zap.Error(err))
		}
	}

	llmClient, err := ai.NewLLMClient(cfg.LLMClientConfig())
	if err != nil {
		return nil, err
	}

	svc := service.New(cfg, log, esClient, qdrantClient, txnStore, llmClient)
	h := handler.New(log, svc, "00000000-0000-0000-0000-000000000002")

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	metrics.RegisterRoutes(router, "search")

	healthChecker := health.NewChecker("search", cfg.Server.Environment, health.DefaultVersion)
	router.GET("/health", healthChecker.GinHandler())
	router.GET("/ready", healthChecker.GinHandler())
	router.GET("/live", simpleLiveHandler("search"))
	h.RegisterRoutes(router)

	return &App{cfg: cfg, log: log, router: router, txnStore: txnStore}, nil
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

	a.log.Info("search service started", zap.Int("port", a.cfg.Server.Port))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func simpleLiveHandler(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "success",
			"data":      gin.H{"service": service, "alive": true},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}
