// @title NeuralOps API Gateway
// @version 1.0
// @description Single entry point for NeuralOps observability platform
// @BasePath /api/v1
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/neuralops/platform/internal/gateway"
	"github.com/neuralops/platform/pkg/logger"
	"github.com/neuralops/platform/pkg/service"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.New("gateway")
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer service.InitObservability(ctx, log, "gateway")()

	app, err := gateway.NewApp(ctx, log)
	if err != nil {
		log.Fatal("failed to initialize api gateway", zap.Error(err))
	}

	if err := app.Run(ctx); err != nil {
		log.Fatal("api gateway stopped with error", zap.Error(err))
	}
}
