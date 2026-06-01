package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/neuralops/platform/internal/analysis"
	"github.com/neuralops/platform/pkg/logger"
	"github.com/neuralops/platform/pkg/service"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.New("analysis")
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer service.InitObservability(ctx, log, "analysis")()

	app, err := analysis.NewApp(ctx, log)
	if err != nil {
		log.Fatal("failed to initialize analysis service", zap.Error(err))
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Error("failed to close analysis service", zap.Error(err))
		}
	}()

	if err := app.Run(ctx); err != nil {
		log.Fatal("analysis service stopped with error", zap.Error(err))
	}
}
