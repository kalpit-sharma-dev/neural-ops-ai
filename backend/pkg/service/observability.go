package service

import (
	"context"

	"github.com/neuralops/platform/internal/platform/observability"
	"go.uber.org/zap"
)

// InitObservability bootstraps OTEL and Prometheus runtime collectors.
func InitObservability(ctx context.Context, log *zap.Logger, serviceName string) func() {
	shutdown, err := observability.Bootstrap(ctx, observability.Config{ServiceName: serviceName})
	if err != nil {
		log.Warn("observability bootstrap failed", zap.Error(err))
		return func() {}
	}
	return func() {
		if err := shutdown(context.Background()); err != nil {
			log.Warn("observability shutdown failed", zap.Error(err))
		}
	}
}
