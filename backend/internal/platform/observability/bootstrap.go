package observability

import (
	"context"
	"os"

	"github.com/neuralops/platform/internal/platform/health"
	"github.com/neuralops/platform/internal/platform/otel"
	"github.com/neuralops/platform/pkg/metrics"
)

// Config configures platform observability bootstrap.
type Config struct {
	ServiceName string
	Version     string
	Environment string
}

// Bootstrap initializes tracing and runtime metrics collectors.
func Bootstrap(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = envOr("SERVICE_NAME", "neuralops")
	}
	if cfg.Environment == "" {
		cfg.Environment = envOr("ENVIRONMENT", "development")
	}
	if cfg.Version == "" {
		cfg.Version = envOr("SERVICE_VERSION", health.DefaultVersion)
	}

	metrics.RegisterRuntimeCollectors()
	return otel.Init(ctx, cfg.ServiceName, cfg.Version, cfg.Environment)
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
