package finops

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/finops/connectors"
)

// ValidateProductionFinOps enforces FIN-PROD-05 settings when FINOPS_REQUIRE_POSTGRES=true.
func ValidateProductionFinOps(pool *pgxpool.Pool) error {
	cfg := connectors.LoadBillingConfig()
	if !cfg.RequirePostgres {
		return nil
	}
	if pool == nil {
		return fmt.Errorf("FINOPS_REQUIRE_POSTGRES=true but database pool is unavailable")
	}
	if cfg.Mode == connectors.BillingModeSimulated && !cfg.AllowSimulationFallback {
		return fmt.Errorf("production FinOps requires FINOPS_BILLING_MODE=live or hybrid with billing files configured")
	}
	return nil
}

// AllowAutoSimulatedIngest returns whether empty tenants may bootstrap via simulated ingest.
func AllowAutoSimulatedIngest() bool {
	cfg := connectors.LoadBillingConfig()
	return cfg.Mode == connectors.BillingModeSimulated || cfg.AllowSimulationFallback
}
