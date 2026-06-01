package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// MetricPoint is one cloud metric sample.
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// MetricSeries is a cloud metric time series.
type MetricSeries struct {
	Provider string        `json:"provider"`
	Metric   string        `json:"metric"`
	Region   string        `json:"region"`
	Unit     string        `json:"unit"`
	Points   []MetricPoint `json:"points"`
	Source   string        `json:"source"`
}

// QueryMetrics fetches cloud vendor metrics using integration config or env credentials.
func QueryMetrics(ctx context.Context, provider, metric, region string, cfg map[string]string) (MetricSeries, error) {
	series := MetricSeries{
		Provider: provider,
		Metric:   metric,
		Region:   region,
		Unit:     "Count",
		Source:   "demo",
	}
	if region == "" {
		region = "us-east-1"
		series.Region = region
	}

	switch strings.ToLower(provider) {
	case "aws":
		if pts, err := queryAWSCloudWatch(ctx, metric, region, cfg); err == nil && len(pts) > 0 {
			series.Points = pts
			series.Source = "aws-cloudwatch"
			return series, nil
		}
	case "azure":
		if pts, err := queryAzureMonitor(ctx, metric, region, cfg); err == nil && len(pts) > 0 {
			series.Points = pts
			series.Source = "azure-monitor"
			return series, nil
		}
	case "gcp":
		if pts, err := queryGCPMonitoring(ctx, metric, region, cfg); err == nil && len(pts) > 0 {
			series.Points = pts
			series.Source = "gcp-monitoring"
			return series, nil
		}
	}

	series.Points = demoSeries(metric)
	return series, nil
}

func queryAWSCloudWatch(ctx context.Context, metric, region string, cfg map[string]string) ([]MetricPoint, error) {
	// Optional: Prometheus remote read URL in integration config
	if promURL := cfg["prometheusUrl"]; promURL != "" {
		return queryPrometheusInstant(ctx, promURL, metric)
	}
	accessKey := firstNonEmpty(cfg["accessKeyId"], os.Getenv("AWS_ACCESS_KEY_ID"))
	secretKey := firstNonEmpty(cfg["secretAccessKey"], os.Getenv("AWS_SECRET_ACCESS_KEY"))
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("aws credentials missing")
	}
	// Use CloudWatch GetMetricStatistics via AWS CLI-compatible HTTP proxy if configured
	if proxyURL := cfg["cloudwatchProxyUrl"]; proxyURL != "" {
		return queryHTTPMetricProxy(ctx, proxyURL, map[string]string{
			"provider": "aws", "metric": metric, "region": region,
		})
	}
	return nil, fmt.Errorf("configure cloudwatchProxyUrl or prometheusUrl")
}

func queryAzureMonitor(ctx context.Context, metric, region string, cfg map[string]string) ([]MetricPoint, error) {
	if proxyURL := cfg["monitorProxyUrl"]; proxyURL != "" {
		return queryHTTPMetricProxy(ctx, proxyURL, map[string]string{
			"provider": "azure", "metric": metric, "region": region,
		})
	}
	if promURL := cfg["prometheusUrl"]; promURL != "" {
		return queryPrometheusInstant(ctx, promURL, metric)
	}
	return nil, fmt.Errorf("azure monitor not configured")
}

func queryGCPMonitoring(ctx context.Context, metric, region string, cfg map[string]string) ([]MetricPoint, error) {
	if proxyURL := cfg["monitorProxyUrl"]; proxyURL != "" {
		return queryHTTPMetricProxy(ctx, proxyURL, map[string]string{
			"provider": "gcp", "metric": metric, "region": region,
		})
	}
	if promURL := cfg["prometheusUrl"]; promURL != "" {
		return queryPrometheusInstant(ctx, promURL, metric)
	}
	return nil, fmt.Errorf("gcp monitoring not configured")
}

func queryHTTPMetricProxy(ctx context.Context, baseURL string, params map[string]string) ([]MetricPoint, error) {
	u, _ := url.Parse(strings.TrimRight(baseURL, "/"))
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("proxy: %s", resp.Status)
	}
	body, _ := io.ReadAll(resp.Body)
	var series MetricSeries
	if err := json.Unmarshal(body, &series); err != nil {
		return nil, err
	}
	return series.Points, nil
}

func queryPrometheusInstant(ctx context.Context, promURL, metric string) ([]MetricPoint, error) {
	u := fmt.Sprintf("%s/api/v1/query?query=%s", strings.TrimRight(promURL, "/"), url.QueryEscape(metric))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var payload struct {
		Data struct {
			Result []struct {
				Value []any `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if len(payload.Data.Result) == 0 {
		return nil, fmt.Errorf("no data")
	}
	val := payload.Data.Result[0].Value
	if len(val) < 2 {
		return nil, fmt.Errorf("invalid prom response")
	}
	ts, _ := val[0].(float64)
	v, _ := val[1].(string)
	var f float64
	_, _ = fmt.Sscan(v, &f)
	return []MetricPoint{{Timestamp: time.Unix(int64(ts), 0).UTC(), Value: f}}, nil
}

func demoSeries(metric string) []MetricPoint {
	now := time.Now().UTC()
	out := make([]MetricPoint, 24)
	base := 42.0
	for i := 0; i < 24; i++ {
		out[i] = MetricPoint{
			Timestamp: now.Add(time.Duration(i-23) * time.Hour),
			Value:     base + float64(i%5)*3.1,
		}
	}
	_ = metric
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
