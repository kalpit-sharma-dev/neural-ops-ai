// Materializer consumes observability stream Kafka topics and persists derived metrics + alert scores.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/kafka"
	"github.com/neuralops/platform/internal/platform/db"
	"github.com/neuralops/platform/internal/streaming"
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

	brokers := splitEnv("KAFKA_BROKERS", "kafka:9092")
	dsn := os.Getenv("POSTGRES_DSN")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var pool *pgxpool.Pool
	if dsn != "" {
		var err error
		pool, err = db.NewPool(ctx, dsn)
		if err != nil {
			log.Fatal("postgres", zap.Error(err))
		}
		if err := db.RunMigrations(ctx, pool, db.DefaultMigrationPaths()...); err != nil {
			log.Warn("migrations", zap.Error(err))
		}
	}

	mat := streaming.NewMaterializer(streaming.NewStore(pool), log)
	go mat.RunFlushLoop(ctx)

	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers:     brokers,
		GroupID:     envOr("KAFKA_GROUP", "neuralops-materializer"),
		Topics:      []string{streaming.TopicObservabilityStream},
		Concurrency: envInt("MATERIALIZER_CONCURRENCY", 32),
		ServiceName: "materializer",
	}, log, mat.Handler())
	if consumer == nil {
		log.Fatal("kafka consumer not configured")
	}
	log.Info("materializer started", zap.Strings("brokers", brokers))
	if err := consumer.Run(ctx); err != nil {
		log.Fatal("consumer", zap.Error(err))
	}
}

func splitEnv(key, def string) []string {
	v := os.Getenv(key)
	if v == "" {
		v = def
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil || n <= 0 {
		return def
	}
	return n
}
