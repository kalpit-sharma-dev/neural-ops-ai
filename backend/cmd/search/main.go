package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/neuralops/platform/internal/search"
	"github.com/neuralops/platform/pkg/logger"
	"github.com/neuralops/platform/pkg/service"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.New("search")
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer service.InitObservability(ctx, log, "search")()

	app, err := search.NewApp(ctx, log)
	if err != nil {
		log.Fatal("failed to initialize search service", zap.Error(err))
	}

	if err := app.Run(ctx); err != nil {
		log.Fatal("search service stopped with error", zap.Error(err))
	}
}
