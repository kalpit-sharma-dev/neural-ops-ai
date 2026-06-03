package allocation

import (
	"testing"

	"github.com/neuralops/platform/internal/finops/domain"
)

func TestApplyAllocationRules(t *testing.T) {
	e := NewEngine()
	items := []domain.CostLineItem{
		{Tags: map[string]string{"team": "payments"}, ResourceID: "r1", EffectiveCost: 100},
		{Tags: map[string]string{"env": "prod"}, ResourceID: "r2", EffectiveCost: 50},
	}
	rules := []domain.AllocationRule{
		{Dimension: "team", TagKey: "team", Priority: 1, Enabled: true},
		{Dimension: "environment", TagKey: "env", Priority: 2, Enabled: true},
	}
	e.Apply("tenant", items, rules)
	if items[0].Team != "payments" {
		t.Fatalf("expected team payments, got %s", items[0].Team)
	}
	if items[1].Environment != "prod" {
		t.Fatalf("expected env prod, got %s", items[1].Environment)
	}
	if items[1].Team == "" {
		t.Fatal("expected inferred team for untagged item")
	}
}

func TestKubernetesCost(t *testing.T) {
	e := NewEngine()
	items := []domain.CostLineItem{
		{Service: "ec2", Team: "platform", ResourceID: "wl-a", EffectiveCost: 200},
		{Service: "ec2", Team: "platform", ResourceID: "wl-b", EffectiveCost: 100},
	}
	rows := e.KubernetesCost(items, "prod-cluster", "")
	if len(rows) == 0 {
		t.Fatal("expected k8s cost rows")
	}
	var total float64
	for _, r := range rows {
		total += r.TotalUSD
	}
	if total <= 0 {
		t.Fatal("expected positive k8s total")
	}
}
