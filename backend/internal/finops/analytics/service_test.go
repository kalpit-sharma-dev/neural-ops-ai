package analytics

import (
	"testing"
	"time"

	"github.com/neuralops/platform/internal/finops/domain"
)

func sampleItems() []domain.CostLineItem {
	now := time.Now().UTC()
	items := make([]domain.CostLineItem, 0, 14)
	for i := 0; i < 14; i++ {
		day := now.AddDate(0, 0, -i)
		amt := 100.0 + float64(i%3)*10
		items = append(items, domain.CostLineItem{
			BillingPeriod: day, IngestedAt: day, Team: "payments",
			EffectiveCost: amt, AmortizedCost: amt, ListCost: amt * 1.1,
			Tags: map[string]string{"team": "payments"},
		})
	}
	// spike for anomaly detection
	items = append(items, domain.CostLineItem{
		IngestedAt: now, Team: "payments", EffectiveCost: 500,
		Tags: map[string]string{"team": "payments"},
	})
	return items
}

func TestAggregateCosts(t *testing.T) {
	s := NewService()
	items := sampleItems()
	series := s.AggregateCosts(items, domain.CostsQuery{Scope: "payments"})
	if series.Total <= 0 {
		t.Fatalf("expected positive total, got %v", series.Total)
	}
	if len(series.Points) == 0 {
		t.Fatal("expected cost points")
	}
	if series.Forecast.P50 <= 0 {
		t.Fatal("expected forecast p50")
	}
	if series.TagCoverage <= 0 {
		t.Fatal("expected tag coverage")
	}
}

func TestDetectAnomalies(t *testing.T) {
	s := NewService()
	items := sampleItems()
	anomalies := s.DetectAnomalies("tenant-1", items)
	if len(anomalies) == 0 {
		t.Fatal("expected at least one anomaly from spike")
	}
	if anomalies[0].Severity == "" {
		t.Fatal("anomaly severity required")
	}
}

func TestBreakdown(t *testing.T) {
	s := NewService()
	items := sampleItems()
	root := s.Breakdown(items, "team")
	if root.AmountUSD <= 0 {
		t.Fatalf("expected breakdown total, got %v", root.AmountUSD)
	}
	if len(root.Children) == 0 {
		t.Fatal("expected breakdown children")
	}
}

func TestForecast(t *testing.T) {
	s := NewService()
	fc := s.Forecast(sampleItems(), "payments", "month")
	if fc.P50 <= 0 || fc.Upper < fc.Lower {
		t.Fatalf("invalid forecast: %+v", fc)
	}
}
