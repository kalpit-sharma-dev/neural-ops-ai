//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	esv8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/alerting/config"
	"github.com/neuralops/platform/internal/alerting/model"
	"github.com/neuralops/platform/internal/alerting/notifier"
	"github.com/neuralops/platform/internal/alerting/pipeline"
	"github.com/neuralops/platform/internal/alerting/repository"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/incident/engine"
	incidentrepo "github.com/neuralops/platform/internal/incident/repository"
	searchconfig "github.com/neuralops/platform/internal/search/config"
	"github.com/neuralops/platform/internal/search/dto"
	searchelastic "github.com/neuralops/platform/internal/search/elasticsearch"
	"github.com/neuralops/platform/internal/seed"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/elasticsearch"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

type recordingNotifier struct {
	sends int
}

func (r *recordingNotifier) Send(_ context.Context, _ model.AlertRecord, _ *domain.Incident) error {
	r.sends++
	return nil
}

func (r *recordingNotifier) Type() string { return "test" }

func TestLogIngestionToSearchFlow(t *testing.T) {
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

	now := time.Now().UTC()
	logs := seed.GenerateAllLogs(now, 500)
	require.NoError(t, seed.LoadElasticsearchAuth(ctx, esURL, "neuralops-logs-000001", "elastic", password, logs))

	es, err := esv8.NewClient(esv8.Config{
		Addresses: []string{esURL},
		Username:  "elastic",
		Password:  password,
	})
	require.NoError(t, err)
	_, err = es.Indices.Refresh(es.Indices.Refresh.WithIndex("neuralops-logs-000001"))
	require.NoError(t, err)

	searchClient, err := searchelastic.NewClient(searchconfig.ElasticsearchConfig{
		URL:          esURL,
		Index:        "neuralops-logs-000001",
		InitialIndex: "neuralops-logs-000001",
	}, zap.NewNop())
	require.NoError(t, err)

	resp, err := searchLogs(ctx, es, searchClient, dto.LogSearchRequest{
		Query: "SocketTimeoutException",
		Size:  10,
	})
	require.NoError(t, err)
	require.Greater(t, resp.Total, int64(0))

	respService, err := searchLogs(ctx, es, searchClient, dto.LogSearchRequest{
		Service: "upi-service",
		Size:    20,
	})
	require.NoError(t, err)
	require.Greater(t, respService.Total, int64(0))
}

func searchLogs(ctx context.Context, es *esv8.Client, _ *searchelastic.Client, req dto.LogSearchRequest) (*dto.SearchResponse, error) {
	body, err := json.Marshal(searchelastic.BuildQuery(req, true))
	if err != nil {
		return nil, err
	}
	res, err := es.Search(
		es.Search.WithContext(ctx),
		es.Search.WithIndex("neuralops-logs-000001"),
		es.Search.WithBody(bytes.NewReader(body)),
		es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	return &dto.SearchResponse{Total: parsed.Hits.Total.Value}, nil
}

func TestIncidentDedupAndAlertDispatch(t *testing.T) {
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

	incidentStore, err := incidentrepo.NewStore(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(incidentStore.Close)

	alertStore, err := repository.NewStore(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(alertStore.Close)

	fingerprint := engine.BuildFingerprint("upi-service", domain.ErrorCategoryTimeout, time.Now().UTC(), 5*time.Minute)
	incident := &domain.Incident{
		ID:               uuid.New(),
		Title:            "UPI timeout burst",
		Summary:          "Integration test incident",
		Severity:         domain.IncidentSeverityP1,
		Status:           domain.IncidentStatusOpen,
		AffectedServices: []string{"upi-service"},
		StartTime:        time.Now().UTC(),
	}
	require.NoError(t, incidentStore.CreateIncident(ctx, incident, fingerprint, "upi-service", string(domain.ErrorCategoryTimeout), seed.DemoTenantID))

	existing, err := incidentStore.FindActiveByFingerprint(ctx, fingerprint)
	require.NoError(t, err)
	require.NotNil(t, existing)

	recorder := &recordingNotifier{}
	registry := notifier.NewRegistry()
	registry.Register("test", recorder)

	processor := pipeline.NewProcessor(config.AlertingConfig{
		DefaultTenant: seed.DemoTenantID,
		GroupWindow:   5 * time.Minute,
		DemoMode:      false,
	}, zap.NewNop(), alertStore, nil, registry, nil)

	incoming := model.IncomingAlert{
		Source:      "PROMETHEUS",
		AlertName:   "HighErrorRate",
		Service:     "upi-service",
		Title:       "UPI error rate above threshold",
		Description: "Error rate exceeded threshold",
		Severity:    "P1",
		FiredAt:     time.Now().UTC(),
		Labels:      map[string]string{"env": "production"},
	}

	first, err := processor.Process(ctx, seed.DemoTenantID, []model.IncomingAlert{incoming})
	require.NoError(t, err)
	require.Len(t, first, 1)
	require.Equal(t, 1, recorder.sends)

	second, err := processor.Process(ctx, seed.DemoTenantID, []model.IncomingAlert{incoming})
	require.NoError(t, err)
	require.Len(t, second, 1)
	require.True(t, second[0].Deduplicated)
	require.Equal(t, 2, second[0].OccurrenceCount)
	require.Equal(t, 1, recorder.sends, "deduplicated alert must not dispatch again")
}
