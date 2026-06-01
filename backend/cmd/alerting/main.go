package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/neuralops/platform/internal/alerting"
	"github.com/neuralops/platform/pkg/logger"
	"github.com/neuralops/platform/pkg/service"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.New("alerting")
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer service.InitObservability(ctx, log, "alerting")()

	app, err := alerting.NewApp(ctx, log)
	if err != nil {
		log.Fatal("failed to initialize alerting service", zap.Error(err))
	}
	defer app.Close()

	if err := app.Run(ctx); err != nil {
		log.Fatal("alerting service stopped with error", zap.Error(err))
	}
}
