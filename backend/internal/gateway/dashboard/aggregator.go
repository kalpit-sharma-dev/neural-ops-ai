package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/neuralops/platform/internal/gateway/config"
	"golang.org/x/sync/errgroup"
)

// Overview is the aggregated dashboard payload.
type Overview struct {
	ActiveIncidents       []IncidentSummary  `json:"activeIncidents"`
	ErrorRateLastHour     float64            `json:"errorRateLastHour"`
	ServiceHealth         []ServiceHealth    `json:"serviceHealth"`
	TopFailingServices    []FailingService   `json:"topFailingServices"`
	RecentAnomalies       []AnomalySummary   `json:"recentAnomalies"`
	DeploymentImpactScore float64            `json:"deploymentImpactScore"`
	LatencyP99Ms          float64            `json:"latencyP99Ms"`
	MTTRMinutes           float64            `json:"mttrMinutes"`
	GeneratedAt           time.Time          `json:"generatedAt"`
}

type IncidentSummary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Severity string `json:"severity"`
	Status   string `json:"status"`
	Service  string `json:"service,omitempty"`
}

type ServiceHealth struct {
	Service string  `json:"service"`
	Score   float64 `json:"score"`
}

type FailingService struct {
	Service    string `json:"service"`
	ErrorCount int64  `json:"errorCount"`
}

type AnomalySummary struct {
	Service   string    `json:"service"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	Timestamp time.Time `json:"timestamp"`
}

// Aggregator fetches dashboard data from downstream services.
type Aggregator struct {
	cfg    config.Config
	client *http.Client
}

// NewAggregator creates a dashboard aggregator.
func NewAggregator(cfg config.Config) *Aggregator {
	timeout := cfg.Dashboard.FetchTimeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &Aggregator{
		cfg:    cfg,
		client: &http.Client{Timeout: timeout},
	}
}

// FetchOverview aggregates dashboard metrics in parallel.
func (a *Aggregator) FetchOverview(ctx context.Context) (*Overview, error) {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.Dashboard.FetchTimeout)
	defer cancel()

	overview := &Overview{GeneratedAt: time.Now().UTC()}
	var (
		incidentsRaw []map[string]any
		searchRaw    map[string]any
	)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		body, err := a.get(groupCtx, a.cfg.Services.Incident+"/api/v1/incidents?status=OPEN&size=100")
		if err != nil {
			return err
		}
		var resp struct {
			Data []map[string]any `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return err
		}
		incidentsRaw = resp.Data
		return nil
	})
	group.Go(func() error {
		end := time.Now().UTC()
		start := end.Add(-1 * time.Hour)
		payload, _ := json.Marshal(map[string]any{
			"severity":  "ERROR",
			"startTime": start.Format(time.RFC3339),
			"endTime":   end.Format(time.RFC3339),
			"size":      0,
		})
		body, err := a.post(groupCtx, a.cfg.Services.Search+"/api/v1/search/logs", payload)
		if err != nil {
			return err
		}
		var resp struct {
			Data map[string]any `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return err
		}
		searchRaw = resp.Data
		return nil
	})

	if err := group.Wait(); err != nil {
		return overview, err
	}

	overview.ActiveIncidents = mapActiveIncidents(incidentsRaw)
	overview.ErrorRateLastHour, overview.TopFailingServices, overview.ServiceHealth = mapSearchStats(searchRaw)
	overview.RecentAnomalies = mapAnomalies(searchRaw)
	overview.DeploymentImpactScore = computeDeploymentImpact(overview.ActiveIncidents)
	overview.LatencyP99Ms = estimateLatencyP99(overview.ErrorRateLastHour, overview.ServiceHealth)
	overview.MTTRMinutes = estimateMTTR(incidentsRaw)
	return overview, nil
}

func (a *Aggregator) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

func (a *Aggregator) post(ctx context.Context, url string, payload []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

func mapActiveIncidents(raw []map[string]any) []IncidentSummary {
	out := make([]IncidentSummary, 0)
	for _, item := range raw {
		severity := strings.ToUpper(fmt.Sprintf("%v", item["severity"]))
		if severity != "P1" && severity != "P2" {
			continue
		}
		summary := IncidentSummary{
			ID:       fmt.Sprintf("%v", item["id"]),
			Title:    fmt.Sprintf("%v", item["title"]),
			Severity: severity,
			Status:   fmt.Sprintf("%v", item["status"]),
		}
		if services, ok := item["affectedServices"].([]any); ok && len(services) > 0 {
			summary.Service = fmt.Sprintf("%v", services[0])
		}
		out = append(out, summary)
	}
	return out
}

func mapSearchStats(raw map[string]any) (float64, []FailingService, []ServiceHealth) {
	total := toInt64(raw["total"])
	errorRate := float64(total)
	if errorRate > 0 {
		errorRate = errorRate / 60.0
	}

	aggs, _ := raw["aggregations"].(map[string]any)
	byService, _ := aggs["byService"].([]any)
	failures := make([]FailingService, 0)
	health := make([]ServiceHealth, 0)
	for _, bucket := range byService {
		entry, _ := bucket.(map[string]any)
		service := fmt.Sprintf("%v", entry["key"])
		count := toInt64(entry["count"])
		failures = append(failures, FailingService{Service: service, ErrorCount: count})
		score := 100.0
		if count > 0 {
			score = max(0, 100-float64(count)/10)
		}
		health = append(health, ServiceHealth{Service: service, Score: score})
	}
	sort.Slice(failures, func(i, j int) bool { return failures[i].ErrorCount > failures[j].ErrorCount })
	if len(failures) > 5 {
		failures = failures[:5]
	}
	return errorRate, failures, health
}

func mapAnomalies(raw map[string]any) []AnomalySummary {
	hits, _ := raw["hits"].([]any)
	out := make([]AnomalySummary, 0, min(len(hits), 5))
	for _, hit := range hits {
		entry, _ := hit.(map[string]any)
		out = append(out, AnomalySummary{
			Service:   fmt.Sprintf("%v", entry["service"]),
			Message:   fmt.Sprintf("%v", entry["message"]),
			Severity:  fmt.Sprintf("%v", entry["severity"]),
			Timestamp: parseTime(fmt.Sprintf("%v", entry["timestamp"])),
		})
		if len(out) >= 5 {
			break
		}
	}
	return out
}

func computeDeploymentImpact(incidents []IncidentSummary) float64 {
	if len(incidents) == 0 {
		return 0
	}
	score := float64(len(incidents)) * 10
	for _, incident := range incidents {
		if incident.Severity == "P1" {
			score += 20
		}
	}
	if score > 100 {
		return 100
	}
	return score
}

func toInt64(value any) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}

func parseTime(value string) time.Time {
	if value == "" || value == "<nil>" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	return time.Time{}
}

func estimateLatencyP99(errorRate float64, health []ServiceHealth) float64 {
	if len(health) == 0 {
		return 40 + errorRate*25
	}
	worst := 0.0
	for _, svc := range health {
		stress := (1 - svc.Score) * 200
		if stress > worst {
			worst = stress
		}
	}
	return max(35, worst+errorRate*15)
}

func estimateMTTR(incidents []map[string]any) float64 {
	var total float64
	var count float64
	for _, item := range incidents {
		if mttr, ok := item["mttr"].(float64); ok && mttr > 0 {
			total += mttr / 60.0
			count++
			continue
		}
		start := parseTime(fmt.Sprintf("%v", item["startTime"]))
		resolved := parseTime(fmt.Sprintf("%v", item["resolvedTime"]))
		if !start.IsZero() && !resolved.IsZero() {
			total += resolved.Sub(start).Minutes()
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / count
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
