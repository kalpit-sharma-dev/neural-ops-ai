package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/neuralops/platform/internal/incident"
	"github.com/neuralops/platform/pkg/logger"
	"github.com/neuralops/platform/pkg/service"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.New("incident")
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer service.InitObservability(ctx, log, "incident")()

	app, err := incident.NewApp(ctx, log)
	if err != nil {
		log.Fatal("failed to initialize incident engine", zap.Error(err))
	}
	defer app.Close()

	if err := app.Run(ctx); err != nil {
		log.Fatal("incident engine stopped with error", zap.Error(err))
	}
}
