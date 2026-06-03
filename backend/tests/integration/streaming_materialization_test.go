//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/kafka"
	"github.com/neuralops/platform/internal/platform/db"
	"github.com/neuralops/platform/internal/streaming"
	kafkamod "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStreamingMaterializationAlertPipeline(t *testing.T) {
	skipUnlessDocker(t)
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("neuralops"),
		postgres.WithUsername("neuralops"),
		postgres.WithPassword("neuralops"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp").WithStartupTimeout(2*time.Minute)),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, db.RunMigrations(ctx, pool, db.DefaultMigrationPaths()...))

	kafkaContainer, err := kafkamod.Run(ctx, "confluentinc/cp-kafka:7.6.1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = kafkaContainer.Terminate(ctx) })

	brokers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err)

	log := zap.NewNop()
	mat := streaming.NewMaterializer(streaming.NewStore(pool), log)
	matCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go mat.RunFlushLoop(matCtx)

	producer := kafka.NewProducer(kafka.ProducerConfig{Brokers: brokers}, log)
	require.NotNil(t, producer)
	t.Cleanup(func() { _ = producer.Close() })
	pub := streaming.NewPublisher(producer)

	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers:     brokers,
		GroupID:     "test-materializer",
		Topics:      []string{streaming.TopicObservabilityStream},
		Concurrency: 4,
		ServiceName: "test-materializer",
	}, log, mat.Handler())
	require.NotNil(t, consumer)

	consumeCtx, stopConsume := context.WithTimeout(ctx, 30*time.Second)
	defer stopConsume()
	done := make(chan error, 1)
	go func() { done <- consumer.Run(consumeCtx) }()

	time.Sleep(2 * time.Second)
	require.NoError(t, pub.PublishAlertSignal(ctx, "default", streaming.AlertSignalPayload{
		PolicyID: "ap-test", Service: "payment-service", Severity: "P1", Count: 3,
	}))

	require.Eventually(t, func() bool {
		var n int
		err := pool.QueryRow(ctx, `
SELECT COUNT(*) FROM alert_score_buckets
WHERE tenant_id = 'default' AND policy_id = 'ap-test'`).Scan(&n)
		return err == nil && n > 0
	}, 25*time.Second, 500*time.Millisecond)

	stopConsume()
	_ = <-done

	var signalCount int64
	var fatigue float64
	err = pool.QueryRow(ctx, `
SELECT signal_count, fatigue_score FROM alert_score_buckets
WHERE tenant_id = 'default' AND policy_id = 'ap-test' AND service = 'payment-service'
ORDER BY bucket_ts DESC LIMIT 1`).Scan(&signalCount, &fatigue)
	require.NoError(t, err)
	require.GreaterOrEqual(t, signalCount, int64(1))
	require.Greater(t, fatigue, 0.0)
}

func TestStreamingMaterializationDirectHandler(t *testing.T) {
	skipUnlessDocker(t)
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("neuralops"),
		postgres.WithUsername("neuralops"),
		postgres.WithPassword("neuralops"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp").WithStartupTimeout(2*time.Minute)),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, db.RunMigrations(ctx, pool, db.DefaultMigrationPaths()...))

	log := zap.NewNop()
	mat := streaming.NewMaterializer(streaming.NewStore(pool), log)

	payload, _ := json.Marshal(streaming.MetricSamplePayload{
		MetricID: "dm-test", Service: "checkout", Value: 0.42,
	})
	raw, _ := streaming.PublishEvent(streaming.StreamEvent{
		Type: streaming.EventMetricSample, TenantID: "default", Payload: payload,
	})
	require.NoError(t, mat.Handler()(ctx, kafka.Message{Value: raw}))

	matCtx, cancel := context.WithCancel(ctx)
	go mat.RunFlushLoop(matCtx)
	t.Cleanup(cancel)

	require.Eventually(t, func() bool {
		var n int
		err := pool.QueryRow(ctx, `
SELECT COUNT(*) FROM derived_metric_samples
WHERE tenant_id = 'default' AND metric_id = 'dm-test'`).Scan(&n)
		return err == nil && n > 0
	}, 15*time.Second, 300*time.Millisecond)
}
