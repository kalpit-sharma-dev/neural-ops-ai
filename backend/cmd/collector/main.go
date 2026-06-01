package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/neuralops/platform/internal/collector"
	"github.com/neuralops/platform/internal/observability"
	"github.com/neuralops/platform/internal/platform/db"
	"github.com/neuralops/platform/pkg/config"
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewProduction()
	defer func() { _ = log.Sync() }()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := loadConfig()
	if cfg.PostgresDSN == "" {
		log.Fatal("POSTGRES_DSN required")
	}

	pool, err := db.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal("postgres connect failed", zap.Error(err))
	}
	defer pool.Close()
	if err := db.RunMigrations(ctx, pool, db.DefaultMigrationPaths()...); err != nil {
		log.Warn("migrations", zap.Error(err))
	}

	promURL := collector.NormalizePrometheusURL(cfg.PrometheusURL)
	prom := observability.NewPromQLClient(promURL)
	repo := observability.NewCollectorsRepo(pool)
	collector.SeedDefaultMonitors(ctx, repo, cfg.TenantID)

	k8s := collector.NewK8sCollector(prom, pool, cfg.TenantID)
	hosts := collector.NewHostCollector(prom, pool, cfg.TenantID)
	synthetic := collector.NewSyntheticRunner(repo, cfg.TenantID)
	sloEval := collector.NewSLOEvaluator(pool, prom, cfg.TenantID, config.Getenv("ALERTING_URL", "http://alerting:8086"))
	retention := collector.NewRetentionSync(cfg.ClickHouseDSN, pool)
	observability.SeedDemoProfiles(ctx, pool, cfg.TenantID)

	interval := cfg.Interval
	if interval <= 0 {
		interval = 60 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	run := func() {
		if err := hosts.Sync(ctx); err != nil {
			log.Warn("host sync", zap.Error(err))
		}
		if err := k8s.Sync(ctx); err != nil {
			log.Warn("k8s sync", zap.Error(err))
		}
		if err := synthetic.RunAll(ctx); err != nil {
			log.Warn("synthetic run", zap.Error(err))
		}
		if err := sloEval.EvaluateAll(ctx); err != nil {
			log.Warn("slo evaluate", zap.Error(err))
		}
		if err := retention.Sync(ctx, cfg.TenantID); err != nil {
			log.Warn("retention sync", zap.Error(err))
		}
	}
	run()

	log.Info("collector started", zap.String("prometheus", promURL), zap.Duration("interval", interval))
	for {
		select {
		case <-ctx.Done():
			log.Info("collector stopped")
			return
		case <-ticker.C:
			run()
		}
	}
}

type collectorConfig struct {
	PostgresDSN    string
	ClickHouseDSN  string
	PrometheusURL  string
	TenantID       string
	Interval       time.Duration
}

func loadConfig() collectorConfig {
	interval, _ := time.ParseDuration(config.Getenv("COLLECTOR_INTERVAL", "60s"))
	return collectorConfig{
		PostgresDSN:   config.Getenv("POSTGRES_DSN", "postgres://neuralops:neuralops@localhost:5432/neuralops?sslmode=disable"),
		ClickHouseDSN: config.Getenv("CLICKHOUSE_DSN", "clickhouse://localhost:9000/neuralops"),
		PrometheusURL: config.Getenv("PROMETHEUS_URL", "http://localhost:9090"),
		TenantID:      config.Getenv("TENANT_ID", "00000000-0000-0000-0000-000000000002"),
		Interval:      interval,
	}
}
