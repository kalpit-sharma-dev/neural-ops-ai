//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/analysis/classifier"
	"github.com/neuralops/platform/internal/analysis/pipeline"
	"github.com/neuralops/platform/internal/analysis/repository"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/ingestion/dto"
	kafkpkg "github.com/neuralops/platform/internal/kafka"
	"github.com/neuralops/platform/internal/platform/tenant"
	searchconfig "github.com/neuralops/platform/internal/search/config"
	searchdto "github.com/neuralops/platform/internal/search/dto"
	searchelastic "github.com/neuralops/platform/internal/search/elasticsearch"
	"github.com/neuralops/platform/internal/seed"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/elasticsearch"
	kafkatest "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

func TestKafkaAnalysisElasticsearchPipeline(t *testing.T) {
	skipUnlessDocker(t)
	ctx := context.Background()

	kafkaContainer, err := kafkatest.Run(ctx, "confluentinc/cp-kafka:7.6.1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = kafkaContainer.Terminate(ctx) })

	brokers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, brokers)

	esContainer, err := elasticsearch.Run(ctx, "docker.elastic.co/elasticsearch/elasticsearch:8.11.0",
		elasticsearch.WithPassword("changeme"),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/").WithPort("9200/tcp").WithStartupTimeout(3*time.Minute),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = esContainer.Terminate(ctx) })

	esURL := esContainer.Settings.Address
	password := esContainer.Settings.Password
	if password == "" {
		password = "changeme"
	}

	index := tenant.LogsBootstrapIndex(seed.DemoTenantID)
	esStore, err := repository.NewElasticsearchStoreAuth(esURL, index, "elastic", password)
	require.NoError(t, err)

	topic := "raw-logs-audit"
	conn, err := kafkago.DialLeader(ctx, "tcp", brokers[0], topic, 0)
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	producer := kafkpkg.NewProducer(kafkpkg.ProducerConfig{
		Brokers:    brokers,
		BatchSize:  10,
		FlushEvery: 100 * time.Millisecond,
	}, zap.NewNop())
	t.Cleanup(func() { _ = producer.Close() })

	pipe := pipeline.New(zap.NewNop(), pipeline.Config{
		Classifier:    classifier.NewService(nil, nil, 0.5, "", nil),
		Elasticsearch: esStore,
		Producer:      producer,
	})

	logID := uuid.New()
	message := "audit pipeline unique marker " + logID.String()
	enriched := dto.EnrichedLog{
		LogEntry: domain.LogEntry{
			ID:        logID,
			Timestamp: time.Now().UTC(),
			Service:   "upi-service",
			Severity:  domain.LogSeverityError,
			Message:   message,
		},
		IngestionMetadata: dto.IngestionMetadata{TenantID: seed.DemoTenantID, Source: "integration-test"},
	}
	payload, err := json.Marshal(enriched)
	require.NoError(t, err)

	require.NoError(t, pipe.ProcessLogMessage(ctx, payload))
	require.NoError(t, producer.Close())

	time.Sleep(2 * time.Second)

	searchClient, err := searchelastic.NewClient(searchconfig.ElasticsearchConfig{
		URL:          esURL,
		Index:        index,
		InitialIndex: index,
		Username:     "elastic",
		Password:     password,
	}, zap.NewNop())
	require.NoError(t, err)

	result, err := searchClient.SearchLogs(ctx, searchdto.LogSearchRequest{
		Query:    message,
		TenantID: seed.DemoTenantID,
		Size:     5,
	})
	require.NoError(t, err)
	require.Greater(t, result.Total, int64(0))
}
