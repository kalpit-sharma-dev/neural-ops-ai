//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/neuralops/platform/internal/platform/tenant"
	"github.com/neuralops/platform/internal/search/dto"
	searchelastic "github.com/neuralops/platform/internal/search/elasticsearch"
	searchconfig "github.com/neuralops/platform/internal/search/config"
	"github.com/neuralops/platform/internal/seed"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/elasticsearch"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

func TestIngestionToSearchPipeline(t *testing.T) {
	skipUnlessDocker(t)
	ctx := context.Background()

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
	logs := seed.GenerateAllLogs(time.Now().UTC(), 200)
	require.NoError(t, seed.LoadElasticsearchAuth(ctx, esURL, index, "elastic", password, logs))

	searchClient, err := searchelastic.NewClient(searchconfig.ElasticsearchConfig{
		URL:          esURL,
		Index:        index,
		InitialIndex: index,
	}, zap.NewNop())
	require.NoError(t, err)

	result, err := searchClient.SearchLogs(ctx, dto.LogSearchRequest{
		Query:    "SocketTimeoutException",
		TenantID: seed.DemoTenantID,
		Size:     10,
	})
	require.NoError(t, err)
	require.Greater(t, result.Total, int64(0))
}
