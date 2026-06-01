package notebook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MetricPoint is one PromQL sample.
type MetricPoint struct {
	Timestamp time.Time
	Value     float64
}

// PromQuerier executes PromQL queries.
type PromQuerier interface {
	QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]MetricPoint, error)
	QueryInstant(ctx context.Context, query string) (float64, error)
}

// LogBackend searches logs via the search service HTTP API.
type LogBackend struct {
	baseURL    string
	httpClient *http.Client
}

// NewLogBackend creates a log search backend.
func NewLogBackend(baseURL string) *LogBackend {
	if strings.TrimSpace(baseURL) == "" {
		return nil
	}
	return &LogBackend{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

// Search runs a structured log query for a tenant.
func (b *LogBackend) Search(ctx context.Context, tenantID, query string) (string, error) {
	if b == nil {
		return "", fmt.Errorf("log search unavailable")
	}
	reqBody := parseLogQuery(query)
	reqBody.TenantID = tenantID
	if reqBody.Size <= 0 {
		reqBody.Size = 20
	}
	raw, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.baseURL+"/api/v1/search/logs", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)
	resp, err := b.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("search: %s", string(body))
	}
	var envelope struct {
		Data struct {
			Total int64 `json:"total"`
			Hits  []struct {
				Timestamp string `json:"timestamp"`
				Service   string `json:"service"`
				Severity  string `json:"severity"`
				Message   string `json:"message"`
			} `json:"hits"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return string(body), nil
	}
	if len(envelope.Data.Hits) == 0 {
		return fmt.Sprintf("0 hits for: %s", query), nil
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d hits (showing %d):\n", envelope.Data.Total, len(envelope.Data.Hits))
	for _, h := range envelope.Data.Hits {
		fmt.Fprintf(&sb, "[%s] %s %s: %s\n", h.Timestamp, h.Service, h.Severity, truncate(h.Message, 120))
	}
	return sb.String(), nil
}

type logSearchPayload struct {
	Query    string `json:"query"`
	Service  string `json:"service"`
	Severity string `json:"severity"`
	TenantID string `json:"tenantId"`
	Size     int    `json:"size"`
}

func parseLogQuery(q string) logSearchPayload {
	out := logSearchPayload{Query: q, Size: 20}
	for _, part := range strings.Fields(q) {
		if strings.Contains(part, ":") {
			k, v, _ := strings.Cut(part, ":")
			switch strings.ToLower(k) {
			case "service":
				out.Service = v
			case "severity", "status":
				out.Severity = strings.ToUpper(v)
			}
		}
	}
	return out
}

// RunPromQL executes PromQL and formats output.
func RunPromQL(ctx context.Context, prom PromQuerier, query string) (string, error) {
	if prom == nil {
		return "", fmt.Errorf("prometheus unavailable")
	}
	end := time.Now().UTC()
	start := end.Add(-1 * time.Hour)
	points, err := prom.QueryRange(ctx, query, start, end, time.Minute)
	if err != nil {
		val, instantErr := prom.QueryInstant(ctx, query)
		if instantErr != nil {
			return "", err
		}
		return fmt.Sprintf("instant value: %.4f", val), nil
	}
	if len(points) == 0 {
		return "no data", nil
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d points:\n", len(points))
	for i, p := range points {
		if i >= 10 {
			sb.WriteString("…\n")
			break
		}
		fmt.Fprintf(&sb, "%s = %.4f\n", p.Timestamp.Format(time.RFC3339), p.Value)
	}
	return sb.String(), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
