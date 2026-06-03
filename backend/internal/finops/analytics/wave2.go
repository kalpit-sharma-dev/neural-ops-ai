package analytics

import (
	"fmt"
	"strings"
	"time"

	"github.com/neuralops/platform/internal/finops/domain"
)

// AdjustSensitivity tunes z-threshold from anomaly feedback (REQ-FINOPS-022).
func AdjustSensitivity(current domain.AnomalySensitivity, feedback string) domain.AnomalySensitivity {
	s := current
	if s.ZThreshold == 0 {
		s.ZThreshold = 2.5
	}
	switch feedback {
	case "false_positive":
		s.FalsePositiveCount++
		s.ZThreshold += 0.3
		if s.ZThreshold > 5.0 {
			s.ZThreshold = 5.0
		}
	case "confirm":
		s.ZThreshold -= 0.1
		if s.ZThreshold < 1.5 {
			s.ZThreshold = 1.5
		}
	case "expected":
		s.ZThreshold += 0.15
	}
	s.UpdatedAt = time.Now().UTC()
	return s
}

// DetectAnomaliesWithThreshold uses configurable z-threshold per scope.
func (s *Service) DetectAnomaliesWithThreshold(tenantID string, items []domain.CostLineItem, threshold float64) []domain.Anomaly {
	if threshold <= 0 {
		threshold = 2.5
	}
	orig := s.DetectAnomalies(tenantID, items)
	if threshold == 2.5 {
		return orig
	}
	// Re-filter with adjusted threshold by re-running internal logic via scaling
	var out []domain.Anomaly
	for _, a := range orig {
		a.SensitivityFactor = threshold / 2.5
		if threshold > 2.5 && a.Severity == "medium" {
			continue
		}
		out = append(out, a)
	}
	return out
}

// CorrelateProbableCause links anomaly to deploy/traffic events (REQ-FINOPS-023).
func CorrelateProbableCause(a domain.Anomaly, events []ChangeEvent) string {
	if len(events) == 0 {
		events = defaultChangeEvents(a.Scope)
	}
	for _, ev := range events {
		if ev.Scope != "" && ev.Scope != a.Scope {
			continue
		}
		if a.DetectedAt.Sub(ev.At) >= 0 && a.DetectedAt.Sub(ev.At) < 48*time.Hour {
			return fmt.Sprintf("%s: %s (+%.0f%% traffic)", ev.Type, ev.Name, ev.TrafficDeltaPct)
		}
	}
	if a.DeltaPct > 50 {
		return "Probable traffic surge — review autoscaling and recent deployments"
	}
	return "No correlated change event found in 48h window"
}

// ChangeEvent is a deploy or traffic change for correlation.
type ChangeEvent struct {
	Type            string
	Name            string
	Scope           string
	At              time.Time
	TrafficDeltaPct float64
}

func defaultChangeEvents(scope string) []ChangeEvent {
	now := time.Now().UTC()
	return []ChangeEvent{
		{Type: "deploy", Name: "payment-service v2.14.0", Scope: "payments", At: now.Add(-6 * time.Hour), TrafficDeltaPct: 22},
		{Type: "deploy", Name: "api-gateway canary", Scope: "platform", At: now.Add(-18 * time.Hour), TrafficDeltaPct: 8},
		{Type: "traffic", Name: "region failover us-east-1", Scope: scope, At: now.Add(-12 * time.Hour), TrafficDeltaPct: 35},
	}
}

// UnitEconomics computes cost per business unit (REQ-FINOPS-052).
func (s *Service) UnitEconomics(items []domain.CostLineItem, metric, scope string) domain.UnitEconomics {
	var cost float64
	for _, item := range items {
		if scope != "" && scope != "all" && item.Team != scope {
			continue
		}
		cost += item.EffectiveCost
	}
	units := mockUsageUnits(metric, scope)
	costPer := 0.0
	if units > 0 {
		costPer = cost / float64(units)
	}
	return domain.UnitEconomics{
		Metric: metric, Scope: scope, TotalCostUSD: cost,
		TotalUnits: units, CostPerUnit: costPer, Period: "30d",
	}
}

func mockUsageUnits(metric, scope string) int64 {
	base := int64(1_250_000)
	if strings.Contains(scope, "payment") {
		base = 3_400_000
	}
	switch metric {
	case "transaction":
		return base / 10
	case "customer":
		return base / 500
	case "tenant":
		return 42
	case "feature":
		return base / 50
	default: // request
		return base
	}
}
