package budgets

import (
	"testing"

	"github.com/neuralops/platform/internal/finops/domain"
)

func TestEvaluateAlerts(t *testing.T) {
	budgets := []domain.Budget{
		{ID: "b1", Name: "Platform", ScopeValue: "platform", AmountUSD: 1000, Thresholds: []int{50, 80, 100}, BurnPct: 85},
	}
	alerts := EvaluateAlerts(budgets, domain.Forecast{})
	if len(alerts) < 2 {
		t.Fatalf("expected at least 2 threshold alerts, got %d", len(alerts))
	}
}

func TestEnrich(t *testing.T) {
	items := []domain.CostLineItem{
		{Team: "platform", EffectiveCost: 500},
		{Team: "payments", EffectiveCost: 200},
	}
	out := Enrich([]domain.Budget{{ScopeValue: "platform", AmountUSD: 1000}}, items)
	if out[0].SpendUSD != 500 || out[0].BurnPct != 50 {
		t.Fatalf("unexpected enrich: spend=%v burn=%v", out[0].SpendUSD, out[0].BurnPct)
	}
}
