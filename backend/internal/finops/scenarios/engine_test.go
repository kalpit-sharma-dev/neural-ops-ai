package scenarios

import (
	"testing"

	"github.com/neuralops/platform/internal/finops/domain"
)

func TestRunRegionMigration(t *testing.T) {
	e := NewEngine()
	r := e.Run(domain.ScenarioRequest{
		Name: "US West migration", ScenarioType: "region_migration", Scope: "platform",
		Params: map[string]string{"from_region": "us-east-1", "to_region": "us-west-2"},
	}, 10000, 1200)
	if r.CostDeltaUSD >= 0 {
		t.Fatalf("expected cost reduction, got delta %v", r.CostDeltaUSD)
	}
	if r.Co2eDeltaKg >= 0 {
		t.Fatalf("expected co2 reduction, got delta %v", r.Co2eDeltaKg)
	}
}
