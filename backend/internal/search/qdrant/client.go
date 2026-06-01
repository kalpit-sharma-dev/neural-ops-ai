package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/neuralops/platform/internal/search/config"
	"github.com/neuralops/platform/internal/search/dto"
	"go.uber.org/zap"
)

// Client performs vector similarity search against Qdrant.
type Client struct {
	baseURL    string
	collection string
	httpClient *http.Client
	log        *zap.Logger
}

// NewClient creates a Qdrant search client.
func NewClient(cfg config.QdrantConfig, log *zap.Logger) *Client {
	return &Client{
		baseURL:    trimRightSlash(cfg.URL),
		collection: cfg.Collection,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		log:        log,
	}
}

// Search performs similarity search and returns log hits from payload metadata.
func (c *Client) Search(ctx context.Context, vector []float32, limit int, startTime, endTime *time.Time, tenantID string) ([]dto.LogHit, error) {
	if len(vector) == 0 {
		return nil, fmt.Errorf("empty query vector")
	}
	if limit <= 0 {
		limit = 50
	}

	filter := buildFilter(startTime, endTime, tenantID)
	body := map[string]any{
		"vector":       vector,
		"limit":        limit,
		"with_payload": true,
	}
	if filter != nil {
		body["filter"] = filter
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/collections/%s/points/search", c.baseURL, c.collection)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("qdrant search error: status=%d body=%s", res.StatusCode, string(raw))
	}

	var parsed searchResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}

	hits := make([]dto.LogHit, 0, len(parsed.Result))
	for _, point := range parsed.Result {
		hits = append(hits, mapPoint(point))
	}
	return hits, nil
}

type searchResponse struct {
	Result []searchPoint `json:"result"`
}

type searchPoint struct {
	ID      any            `json:"id"`
	Score   float64        `json:"score"`
	Payload map[string]any `json:"payload"`
}

func mapPoint(point searchPoint) dto.LogHit {
	payload := point.Payload
	id := fmt.Sprintf("%v", point.ID)
	if logID, ok := payload["log_id"].(string); ok && logID != "" {
		id = logID
	}
	return dto.LogHit{
		ID:             id,
		Score:          point.Score,
		Timestamp:      parseTime(stringValue(payload["timestamp"])),
		Service:        stringValue(payload["service"]),
		Severity:       stringValue(payload["severity"]),
		Message:        stringValue(payload["message"]),
		TraceID:        stringValue(payload["traceId"]),
		TxnID:          stringValue(payload["txnId"]),
		Classification: stringValue(payload["classification"]),
		Source:         "qdrant",
	}
}

func buildFilter(startTime, endTime *time.Time, tenantID string) map[string]any {
	must := make([]map[string]any, 0, 3)
	if tenantID != "" {
		must = append(must, map[string]any{
			"key":   "tenantId",
			"match": map[string]any{"value": tenantID},
		})
	}
	if startTime != nil || endTime != nil {
		rangeCond := map[string]any{}
		if startTime != nil {
			rangeCond["gte"] = startTime.UTC().Unix()
		}
		if endTime != nil {
			rangeCond["lte"] = endTime.UTC().Unix()
		}
		must = append(must, map[string]any{
			"key":   "timestamp",
			"range": rangeCond,
		})
	}
	if len(must) == 0 {
		return nil
	}
	return map[string]any{"must": must}
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

func trimRightSlash(value string) string {
	for len(value) > 0 && value[len(value)-1] == '/' {
		value = value[:len(value)-1]
	}
	return value
}
