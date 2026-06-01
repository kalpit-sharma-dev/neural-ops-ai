package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/neuralops/platform/internal/ai"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/platform/db"
	"github.com/neuralops/platform/internal/platform/tenant"
)

// ElasticsearchStore indexes enriched logs.
type ElasticsearchStore struct {
	client *elasticsearch.Client
	index  string
}

// NewElasticsearchStore creates an Elasticsearch repository.
func NewElasticsearchStore(url, index string) (*ElasticsearchStore, error) {
	return NewElasticsearchStoreAuth(url, index, "", "")
}

// NewElasticsearchStoreAuth creates an Elasticsearch repository with optional basic auth.
func NewElasticsearchStoreAuth(url, index, username, password string) (*ElasticsearchStore, error) {
	esCfg := elasticsearch.Config{Addresses: []string{url}}
	if username != "" {
		esCfg.Username = username
		esCfg.Password = password
	}
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("create elasticsearch client: %w", err)
	}

	store := &ElasticsearchStore{client: client, index: index}
	if err := db.ApplyElasticsearchLogTemplate(context.Background(), client, ""); err != nil {
		return nil, fmt.Errorf("apply logs template: %w", err)
	}
	if err := store.ensureIndex(context.Background()); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *ElasticsearchStore) ensureIndex(ctx context.Context) error {
	exists, err := s.client.Indices.Exists([]string{s.index})
	if err != nil {
		return err
	}
	defer exists.Body.Close()
	if exists.StatusCode == 200 {
		return nil
	}

	mapping := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"timestamp":           map[string]string{"type": "date"},
				"service":             map[string]string{"type": "keyword"},
				"environment":         map[string]string{"type": "keyword"},
				"severity":            map[string]string{"type": "keyword"},
				"message":             map[string]string{"type": "text"},
				"traceId":             map[string]string{"type": "keyword"},
				"txnId":               map[string]string{"type": "keyword"},
				"classification":      map[string]string{"type": "keyword"},
				"classification_conf": map[string]string{"type": "float"},
				"plain_english":       map[string]string{"type": "text"},
				"embedding_id":        map[string]string{"type": "keyword"},
			},
		},
	}
	body, _ := json.Marshal(mapping)
	res, err := s.client.Indices.Create(s.index, s.client.Indices.Create.WithBody(bytes.NewReader(body)))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() && res.StatusCode != 400 {
		return fmt.Errorf("create index: %s", res.String())
	}
	return nil
}

// IndexEnrichedLog indexes a log with AI enrichment fields under a tenant-specific index.
func (s *ElasticsearchStore) IndexEnrichedLog(ctx context.Context, tenantID string, entry domain.LogEntry, classification *domain.ErrorClassification, explanation *ai.LogExplanation, embeddingID string) error {
	if tenantID == "" {
		tenantID = "00000000-0000-0000-0000-000000000002"
	}
	indexName := tenant.LogsBootstrapIndex(tenantID)
	doc := map[string]any{
		"id":          entry.ID.String(),
		"tenantId":    tenantID,
		"timestamp":   entry.Timestamp.UTC().Format(time.RFC3339Nano),
		"service":     entry.Service,
		"environment": entry.Environment,
		"severity":    entry.Severity,
		"message":     entry.Message,
		"traceId":     entry.TraceID,
		"txnId":       entry.TxnID,
		"host":        entry.Host,
		"pod":         entry.Pod,
		"namespace":   entry.Namespace,
		"labels":      entry.Labels,
		"embedding_id": embeddingID,
	}
	if classification != nil {
		doc["classification"] = classification.Category
		doc["classification_conf"] = classification.Confidence
		doc["classification_reasoning"] = classification.Reasoning
	}
	if explanation != nil {
		doc["plain_english"] = explanation.PlainEnglish
		doc["technical_summary"] = explanation.TechnicalSummary
		doc["severity_assessment"] = explanation.SeverityAssessment
		doc["possible_causes"] = explanation.PossibleCauses
		doc["recommended_actions"] = explanation.RecommendedActions
	}
	if entry.ParsedStackTrace != nil {
		doc["stack_trace"] = entry.ParsedStackTrace
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	res, err := s.client.Index(indexName, bytes.NewReader(body), s.client.Index.WithDocumentID(entry.ID.String()), s.client.Index.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("index document: %s", res.String())
	}
	return nil
}
