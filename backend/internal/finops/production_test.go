package finops_test

import (
	"testing"

	"github.com/neuralops/platform/internal/finops"
)

func TestValidateProductionFinOpsRequiresPool(t *testing.T) {
	t.Setenv("FINOPS_REQUIRE_POSTGRES", "true")
	t.Setenv("FINOPS_BILLING_MODE", "live")
	t.Setenv("FINOPS_ALLOW_SIMULATION", "false")
	if err := finops.ValidateProductionFinOps(nil); err == nil {
		t.Fatal("expected error without pool")
	}
}

func TestValidateProductionFinOpsAllowsHybrid(t *testing.T) {
	t.Setenv("FINOPS_REQUIRE_POSTGRES", "true")
	t.Setenv("FINOPS_BILLING_MODE", "hybrid")
	t.Setenv("FINOPS_ALLOW_SIMULATION", "true")
	// pool nil still fails require postgres
	if err := finops.ValidateProductionFinOps(nil); err == nil {
		t.Fatal("expected pool required")
	}
}
