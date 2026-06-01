package alerting

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/ai"
	"github.com/neuralops/platform/internal/alerting/config"
	"github.com/neuralops/platform/internal/alerting/escalation"
	"github.com/neuralops/platform/internal/alerting/handler"
	"github.com/neuralops/platform/internal/alerting/notifier"
	"github.com/neuralops/platform/internal/alerting/oncall"
	"github.com/neuralops/platform/internal/alerting/pipeline"
	"github.com/neuralops/platform/internal/alerting/repository"
	"github.com/neuralops/platform/internal/alerting/service"
	"github.com/neuralops/platform/internal/platform/health"
	"github.com/neuralops/platform/pkg/metrics"
	"go.uber.org/zap"
)

// App runs the alerting service.
type App struct {
	cfg    *config.Config
	log    *zap.Logger
	repo   *repository.Store
	router *gin.Engine
}

// NewApp wires the alerting service.
func NewApp(ctx context.Context, log *zap.Logger) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	repo, err := repository.NewStore(ctx, cfg.Postgres.DSN)
	if err != nil {
		return nil, err
	}

	llmClient, err := ai.NewLLMClient(cfg.LLMClientConfig())
	if err != nil {
		return nil, err
	}

	registry := notifier.NewRegistry()
	schedule, err := oncall.LoadSchedule(cfg.Oncall.SchedulesFile)
	if err != nil {
		log.Warn("oncall schedule unavailable", zap.Error(err))
		schedule = &oncall.Schedule{}
	}

	processor := pipeline.NewProcessor(cfg.Alerting, log, repo, llmClient, registry, notifier.NewMobilePushNotifier(repo.Pool()))
	svc := service.New(cfg.Alerting.DefaultTenant, processor, repo, registry)
	if err := svc.ReloadNotifiers(ctx, cfg.Alerting.DefaultTenant); err != nil {
		log.Warn("failed to reload notifiers", zap.Error(err))
	}

	escalationEngine := escalation.NewEngine(cfg.Alerting, log, repo, registry, schedule)
	go escalationEngine.Run(ctx)

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	metrics.RegisterRoutes(router, "alerting")

	healthChecker := health.NewChecker("alerting", cfg.Server.Environment, health.DefaultVersion)
	healthChecker.Register("postgres", health.PingCheck(repo.Pool()))
	router.GET("/health", healthChecker.GinHandler())
	router.GET("/ready", healthChecker.GinHandler())
	router.GET("/live", simpleLiveHandler("alerting"))

	h := handler.New(log, svc, cfg.Alerting.DefaultTenant)
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

	a.log.Info("alerting service started", zap.Int("port", a.cfg.Server.Port))
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
			"status":    "success",
			"data":      gin.H{"service": service, "alive": true},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}
