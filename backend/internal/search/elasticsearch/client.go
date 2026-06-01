package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/neuralops/platform/internal/platform/db"
	"github.com/neuralops/platform/internal/platform/tenant"
	"github.com/neuralops/platform/internal/search/config"
	"github.com/neuralops/platform/internal/search/dto"
	"go.uber.org/zap"
)

// Client executes Elasticsearch search operations.
type Client struct {
	es    *elasticsearch.Client
	index string
	log   *zap.Logger
}

// NewClient creates an Elasticsearch search client.
func NewClient(cfg config.ElasticsearchConfig, log *zap.Logger) (*Client, error) {
	esCfg := elasticsearch.Config{Addresses: []string{cfg.URL}}
	if cfg.Username != "" {
		esCfg.Username = cfg.Username
		esCfg.Password = cfg.Password
	}
	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("create elasticsearch client: %w", err)
	}
	client := &Client{es: es, index: cfg.ResolveIndex(), log: log}
	if err := db.ApplyElasticsearchLogTemplate(context.Background(), es, ""); err != nil {
		log.Warn("elasticsearch logs template setup skipped", zap.Error(err))
	}
	if err := client.EnsureIndexTemplate(context.Background(), cfg); err != nil {
		log.Warn("elasticsearch index template setup skipped", zap.Error(err))
	}
	if err := client.EnsureILMPolicy(context.Background(), cfg); err != nil {
		log.Warn("elasticsearch ILM setup skipped", zap.Error(err))
	}
	if err := client.EnsureRolloverAlias(context.Background(), cfg); err != nil {
		log.Warn("elasticsearch rollover alias setup skipped", zap.Error(err))
	}
	return client, nil
}

// SearchLogs executes a structured log search scoped to a tenant.
func (c *Client) SearchLogs(ctx context.Context, req dto.LogSearchRequest) (*dto.SearchResponse, error) {
	body := BuildQuery(req, true)
	return c.search(ctx, c.resolveIndexes(req.TenantID), body)
}

// SearchByTrace returns logs for a trace ID within a tenant.
func (c *Client) SearchByTrace(ctx context.Context, traceID, tenantID string, size int) (*dto.SearchResponse, error) {
	body := BuildTraceQuery(traceID, tenantID, size)
	return c.search(ctx, c.resolveIndexes(tenantID), body)
}

// SearchByIDs fetches logs by document IDs using terms query within a tenant scope.
func (c *Client) SearchByIDs(ctx context.Context, ids []string, tenantID string, size int) (*dto.SearchResponse, error) {
	if len(ids) == 0 {
		return &dto.SearchResponse{}, nil
	}
	if size <= 0 {
		size = len(ids)
	}
	body := map[string]any{
		"size": size,
		"query": map[string]any{
			"ids": map[string]any{"values": ids},
		},
	}
	return c.search(ctx, c.resolveIndexes(tenantID), body)
}

func (c *Client) resolveIndexes(tenantID string) string {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return c.index
	}
	alias := tenant.LogsAlias(tenantID)
	return alias + "," + c.index
}

func (c *Client) search(ctx context.Context, index string, body map[string]any) (*dto.SearchResponse, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex(index),
		c.es.Search.WithBody(bytes.NewReader(payload)),
		c.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch search error: %s", string(raw))
	}

	var parsed esSearchResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}

	response := &dto.SearchResponse{
		Total:  parsed.Hits.Total.Value,
		TookMs: parsed.Took,
		Hits:   make([]dto.LogHit, 0, len(parsed.Hits.Hits)),
	}

	for _, hit := range parsed.Hits.Hits {
		response.Hits = append(response.Hits, mapHit(hit))
	}
	if len(parsed.Hits.Hits) > 0 {
		response.SearchAfter = parsed.Hits.Hits[len(parsed.Hits.Hits)-1].Sort
	}
	response.Aggregations = mapAggregations(parsed.Aggregations)
	return response, nil
}

type esSearchResponse struct {
	Took         int64 `json:"took"`
	Hits         esHits `json:"hits"`
	Aggregations esAggs `json:"aggregations"`
}

type esHits struct {
	Total esTotal  `json:"total"`
	Hits  []esHit  `json:"hits"`
}

type esTotal struct {
	Value int64 `json:"value"`
}

type esHit struct {
	ID     string         `json:"_id"`
	Score  float64        `json:"_score"`
	Sort   []any          `json:"sort"`
	Source map[string]any `json:"_source"`
}

type esAggs struct {
	ByService  esTermsAgg `json:"by_service"`
	BySeverity esTermsAgg `json:"by_severity"`
	ByHour     esDateAgg  `json:"by_hour"`
}

type esTermsAgg struct {
	Buckets []esBucket `json:"buckets"`
}

type esDateAgg struct {
	Buckets []esDateBucket `json:"buckets"`
}

type esBucket struct {
	Key   any   `json:"key"`
	DocCount int64 `json:"doc_count"`
}

type esDateBucket struct {
	KeyAsString string `json:"key_as_string"`
	Key         int64  `json:"key"`
	DocCount    int64  `json:"doc_count"`
}

func mapHit(hit esHit) dto.LogHit {
	source := hit.Source
	return dto.LogHit{
		ID:             hit.ID,
		Score:          hit.Score,
		Timestamp:      parseTime(stringValue(source["timestamp"])),
		Service:        stringValue(source["service"]),
		Severity:       stringValue(source["severity"]),
		Message:        stringValue(source["message"]),
		TraceID:        stringValue(source["traceId"]),
		TxnID:          stringValue(source["txnId"]),
		Host:           stringValue(source["host"]),
		Pod:            stringValue(source["pod"]),
		Classification: stringValue(source["classification"]),
		PlainEnglish:   stringValue(source["plain_english"]),
		Source:         "elasticsearch",
	}
}

func mapAggregations(aggs esAggs) *dto.Aggregations {
	if len(aggs.ByService.Buckets) == 0 && len(aggs.BySeverity.Buckets) == 0 && len(aggs.ByHour.Buckets) == 0 {
		return nil
	}
	result := &dto.Aggregations{
		ByService:  mapTermBuckets(aggs.ByService.Buckets),
		BySeverity: mapTermBuckets(aggs.BySeverity.Buckets),
		ByHour:     make([]dto.Bucket, 0, len(aggs.ByHour.Buckets)),
	}
	for _, bucket := range aggs.ByHour.Buckets {
		key := bucket.KeyAsString
		if key == "" {
			key = fmt.Sprintf("%d", bucket.Key)
		}
		result.ByHour = append(result.ByHour, dto.Bucket{Key: key, Count: bucket.DocCount})
	}
	return result
}

func mapTermBuckets(buckets []esBucket) []dto.Bucket {
	out := make([]dto.Bucket, 0, len(buckets))
	for _, bucket := range buckets {
		out = append(out, dto.Bucket{Key: fmt.Sprintf("%v", bucket.Key), Count: bucket.DocCount})
	}
	return out
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	return time.Time{}
}
