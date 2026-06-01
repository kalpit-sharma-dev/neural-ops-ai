package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PromQLClient queries Prometheus HTTP API.
type PromQLClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPromQLClient creates a Prometheus query client.
func NewPromQLClient(baseURL string) *PromQLClient {
	if baseURL == "" {
		return nil
	}
	return &PromQLClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type promQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]any           `json:"values"`
			Value  []any             `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// QueryRange executes a PromQL range query and returns metric series points.
func (p *PromQLClient) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]MetricSeriesPoint, error) {
	if p == nil || p.baseURL == "" {
		return nil, fmt.Errorf("prometheus unavailable")
	}
	if end.IsZero() {
		end = time.Now().UTC()
	}
	if start.IsZero() {
		start = end.Add(-1 * time.Hour)
	}
	if step <= 0 {
		step = time.Minute
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("start", fmt.Sprintf("%d", start.Unix()))
	params.Set("end", fmt.Sprintf("%d", end.Unix()))
	params.Set("step", strconv.Itoa(int(step.Seconds())))

	reqURL := p.baseURL + "/api/v1/query_range?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("prometheus query failed: %s", string(body))
	}

	var parsed promQueryResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.Status != "success" || len(parsed.Data.Result) == 0 {
		return nil, fmt.Errorf("no prometheus data")
	}

	points := make([]MetricSeriesPoint, 0)
	for _, sample := range parsed.Data.Result[0].Values {
		if len(sample) < 2 {
			continue
		}
		ts, ok := sample[0].(float64)
		if !ok {
			continue
		}
		valStr := fmt.Sprint(sample[1])
		val, _ := strconv.ParseFloat(valStr, 64)
		points = append(points, MetricSeriesPoint{
			Timestamp: time.Unix(int64(ts), 0).UTC(),
			Value:     val,
		})
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("empty prometheus series")
	}
	return points, nil
}

// QueryInstant executes an instant PromQL query.
func (p *PromQLClient) QueryInstant(ctx context.Context, query string) (float64, error) {
	if p == nil || p.baseURL == "" {
		return 0, fmt.Errorf("prometheus unavailable")
	}
	params := url.Values{}
	params.Set("query", query)
	reqURL := p.baseURL + "/api/v1/query?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, err
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var parsed promQueryResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return 0, err
	}
	if len(parsed.Data.Result) == 0 || len(parsed.Data.Result[0].Value) < 2 {
		return 0, fmt.Errorf("no instant data")
	}
	val, _ := strconv.ParseFloat(fmt.Sprint(parsed.Data.Result[0].Value[1]), 64)
	return val, nil
}

// MetricToPromQL maps catalog metric names to PromQL expressions.
func MetricToPromQL(metric, service string) string {
	switch metric {
	case "latency_p95", "latency_p99":
		return fmt.Sprintf(`histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{service="%s"}[5m])) by (le)) * 1000`, service)
	case "error_rate":
		return fmt.Sprintf(`100 * sum(rate(http_requests_total{service="%s",status=~"5.."}[5m])) / clamp_min(sum(rate(http_requests_total{service="%s"}[5m])), 1)`, service, service)
	case "throughput":
		return fmt.Sprintf(`sum(rate(http_requests_total{service="%s"}[5m]))`, service)
	case "cpu_usage":
		return `100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
	default:
		return fmt.Sprintf(`sum(rate(http_requests_total{service="%s"}[5m]))`, service)
	}
}
