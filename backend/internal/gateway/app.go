package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/ai"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/chat"
	"github.com/neuralops/platform/internal/gateway/config"
	"github.com/neuralops/platform/internal/gateway/dashboard"
	_ "github.com/neuralops/platform/internal/gateway/docs"
	gatewayhandler "github.com/neuralops/platform/internal/gateway/handler"
	gwmiddleware "github.com/neuralops/platform/internal/gateway/middleware"
	"github.com/neuralops/platform/internal/gateway/proxy"
	"github.com/neuralops/platform/internal/gateway/websocket"
	"github.com/neuralops/platform/internal/observability"
	"github.com/neuralops/platform/internal/platform/db"
	"github.com/neuralops/platform/internal/platform/health"
	"github.com/neuralops/platform/internal/security"
	"github.com/neuralops/platform/internal/tracequery"
	"github.com/neuralops/platform/pkg/metrics"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// App runs the API gateway.
type App struct {
	cfg    *config.Config
	log    *zap.Logger
	router *gin.Engine
	wsHub  *websocket.Hub
	pool   *pgxpool.Pool
}

// NewApp wires the API gateway.
func NewApp(ctx context.Context, log *zap.Logger) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	redisClient, err := gwmiddleware.NewRedisClient(cfg.Redis.URL)
	if err != nil {
		log.Warn("redis unavailable, rate limiting disabled", zap.Error(err))
	}

	var pgPool *pgxpool.Pool
	var auditRepo *security.AuditRepository
	var identityStore *auth.IdentityStore
	var jwtIssuer *auth.JWTIssuer
	var oidcFlow *auth.OIDCFlow
	if cfg.Postgres.DSN != "" {
		pgPool, err = db.NewPool(ctx, cfg.Postgres.DSN)
		if err != nil {
			log.Warn("postgres unavailable, audit logging disabled", zap.Error(err))
		} else if migrateErr := db.RunMigrations(ctx, pgPool, db.DefaultMigrationPaths()...); migrateErr != nil {
			log.Warn("gateway migrations failed", zap.Error(migrateErr))
		} else {
			auditRepo = security.NewAuditRepository(pgPool)
			if cfg.ClickHouse.DSN != "" {
				if chConn, chErr := openClickHouse(ctx, cfg.ClickHouse.DSN); chErr != nil {
					log.Warn("clickhouse unavailable, audit replication disabled", zap.Error(chErr))
				} else {
					auditRepo = auditRepo.WithClickHouse(chConn)
				}
			}
			identityStore = auth.NewIdentityStore(pgPool)
			auth.StartBackgroundCleanup(ctx, identityStore, 15*time.Minute)
		}
	}

	var authenticator *auth.Authenticator
	authenticator, err = auth.NewAuthenticator(*cfg, identityStore)
	if err != nil {
		return nil, err
	}
	jwtIssuer = authenticator.Issuer()
	var ssoManager *auth.SSOManager
	if cfg.Auth.OIDC.Enabled {
		oidcFlow, err = auth.NewOIDCFlow(ctx, cfg.Auth.OIDC)
		if err != nil {
			log.Warn("oidc flow unavailable", zap.Error(err))
		}
	}
	if pgPool != nil {
		ssoManager = auth.NewSSOManager(pgPool, cfg.Auth.OIDC, cfg.Tenant.DefaultTenant, oidcFlow)
		if err := ssoManager.ReloadAll(ctx); err != nil {
			log.Warn("tenant sso reload", zap.Error(err))
		}
	} else {
		ssoManager = auth.NewSSOManager(nil, cfg.Auth.OIDC, cfg.Tenant.DefaultTenant, oidcFlow)
	}

	var samlFlow auth.SAMLProvider
	if cfg.Auth.SAML.Enabled {
		samlFlow, err = auth.NewSAMLProvider(cfg.Auth.SAML)
		if err != nil {
			log.Warn("saml flow unavailable", zap.Error(err))
		}
	}

	llmClient, err := ai.NewLLMClient(cfg.LLMClientConfig())
	if err != nil {
		return nil, err
	}

	dashboardAgg := dashboard.NewAggregator(*cfg)
	chatSvc := chat.NewService(*cfg, llmClient)
	wsHub := websocket.NewHub(log, cfg.CORS.AllowedOrigins)
	wsHub.StartHeartbeat(60 * time.Second)

	upstreamTransport, err := proxy.BuildTransport(proxy.MTLSConfig{
		Enabled:    cfg.Auth.MTLS.Enabled,
		CertFile:   cfg.Auth.MTLS.CertFile,
		KeyFile:    cfg.Auth.MTLS.KeyFile,
		CAFile:     cfg.Auth.MTLS.CAFile,
		ServerName: cfg.Auth.MTLS.ServerName,
	})
	if err != nil {
		return nil, fmt.Errorf("mTLS transport: %w", err)
	}
	if cfg.Auth.MTLS.Enabled && upstreamTransport == http.DefaultTransport {
		log.Warn("mTLS enabled but client certificates missing; upstream calls will use plain HTTP transport")
	}

	ingestionProxy, err := proxy.NewWithTransport(cfg.Services.Ingestion, log, upstreamTransport)
	if err != nil {
		return nil, err
	}
	analysisProxy, err := proxy.NewWithTransport(cfg.Services.Analysis, log, upstreamTransport)
	if err != nil {
		return nil, err
	}
	incidentProxy, err := proxy.NewWithTransport(cfg.Services.Incident, log, upstreamTransport)
	if err != nil {
		return nil, err
	}
	searchProxy, err := proxy.NewWithTransport(cfg.Services.Search, log, upstreamTransport)
	if err != nil {
		return nil, err
	}
	alertingProxy, err := proxy.NewWithTransport(cfg.Services.Alerting, log, upstreamTransport)
	if err != nil {
		return nil, err
	}

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	healthChecker := health.NewChecker("gateway", cfg.Server.Environment, health.DefaultVersion)
	if pgPool != nil {
		healthChecker.Register("postgres", health.PingCheck(pgPool))
	}
	if redisClient != nil {
		healthChecker.Register("redis", health.RedisCheck(redisClient))
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gwmiddleware.CORS(cfg.CORS))
	router.Use(gwmiddleware.RequestLogger(log))
	router.Use(gwmiddleware.AuthRateLimit(cfg.AuthRateLimit, redisClient))
	router.Use(authenticator.Middleware())
	router.Use(gwmiddleware.Tenant(cfg.Tenant, identityStore))
	router.Use(gwmiddleware.DeveloperScope())
	router.Use(gwmiddleware.RateLimit(cfg.RateLimit, redisClient))
	router.Use(gwmiddleware.RBAC())
	router.Use(gwmiddleware.AuditLog(auditRepo, log))

	metrics.RegisterRoutes(router, "gateway")
	router.GET("/health", healthChecker.GinHandler())
	router.GET("/ready", healthChecker.GinHandler())
	router.GET("/live", simpleLiveHandler("gateway"))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	h := gatewayhandler.New(log, *cfg, dashboardAgg, chatSvc, wsHub)
	h.RegisterNativeRoutes(router)

	var chConn driver.Conn
	var spanStore *tracequery.SpanStore
	if cfg.ClickHouse.DSN != "" {
		if conn, chErr := openClickHouse(ctx, cfg.ClickHouse.DSN); chErr != nil {
			log.Warn("clickhouse unavailable for observability", zap.Error(chErr))
		} else {
			chConn = conn
		}
		if ss, spanErr := tracequery.NewSpanStore(ctx, cfg.ClickHouse.DSN); spanErr != nil {
			log.Warn("span store unavailable", zap.Error(spanErr))
		} else {
			spanStore = ss
		}
	}
	obsHandler := observability.NewHandler(observability.Deps{
		Log: log, Mem: observability.NewStore(), Spans: spanStore, PG: observability.NewPostgresRepo(pgPool),
		Collectors: observability.NewCollectorsRepo(pgPool),
		Prom:       observability.NewPromQLClient(cfg.Prometheus.URL),
		CH: chConn, Pool: pgPool, Identity: identityStore,
		SearchURL: cfg.Services.Search, SSOManager: ssoManager,
	})
	obsHandler.RegisterRoutes(router.Group("/api/v1"))
	authHandler := gatewayhandler.NewAuthHandler(log, *cfg, jwtIssuer, identityStore, ssoManager, samlFlow, authenticator, auditRepo)
	authHandler.RegisterRoutes(router)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/info", h.Info)

		// Ingestion exposes only exact POST endpoints; avoid /logs/*path and /metrics/*path
		// wildcards because observability registers /logs/metric-rules and /metrics/catalog.
		v1.POST("/logs", ingestionProxy.GinHandler())
		v1.POST("/metrics", ingestionProxy.GinHandler())
		v1.POST("/events", ingestionProxy.GinHandler())
		v1.POST("/traces", ingestionProxy.GinHandler())
		v1.POST("/webhooks/alerts", ingestionProxy.GinHandler())
		v1.POST("/webhooks/dynatrace", ingestionProxy.GinHandler())
		v1.POST("/webhooks/datadog", ingestionProxy.GinHandler())
		v1.POST("/webhooks/prometheus", ingestionProxy.GinHandler())

		v1.Any("/search", searchProxy.GinHandler())
		v1.Any("/search/*path", searchProxy.GinHandler())

		v1.Any("/incidents", incidentProxy.GinHandler())
		v1.Any("/incidents/*path", incidentProxy.GinHandler())
		v1.Any("/services/*path", incidentProxy.GinHandler())
		v1.Any("/transactions/*path", incidentProxy.GinHandler())

		v1.Any("/analysis", analysisProxy.GinHandler())
		v1.Any("/analysis/*path", analysisProxy.GinHandler())

		v1.Any("/alerts", alertingProxy.GinHandler())
		v1.Any("/alerts/*path", alertingProxy.GinHandler())
		v1.Any("/notifications/*path", alertingProxy.GinHandler())
		v1.Any("/escalation/*path", alertingProxy.GinHandler())
	}

	return &App{cfg: cfg, log: log, router: router, wsHub: wsHub, pool: pgPool}, nil
}

// Run starts the HTTP server until context cancellation.
func (a *App) Run(ctx context.Context) error {
	metrics.StartDBPoolReporter(ctx, "gateway", "postgres", a.pool)

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
		if a.pool != nil {
			a.pool.Close()
		}
	}()

	a.log.Info("api gateway started", zap.Int("port", a.cfg.Server.Port))
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

func openClickHouse(ctx context.Context, dsn string) (driver.Conn, error) {
	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse clickhouse dsn: %w", err)
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}
	if err := db.RunClickHouseMigrations(ctx, conn); err != nil {
		return nil, fmt.Errorf("clickhouse migrations: %w", err)
	}
	return conn, nil
}
