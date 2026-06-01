package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/neuralops/platform/internal/ingestion"
	"github.com/neuralops/platform/pkg/logger"
	"github.com/neuralops/platform/pkg/service"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.New("ingestion")
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer service.InitObservability(ctx, log, "ingestion")()

	app, err := ingestion.NewApp(ctx, log)
	if err != nil {
		log.Fatal("failed to initialize ingestion service", zap.Error(err))
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Error("failed to close ingestion service", zap.Error(err))
		}
	}()

	if err := app.Run(ctx); err != nil {
		log.Fatal("ingestion service stopped with error", zap.Error(err))
	}
}
