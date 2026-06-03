package scenarios

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

// Engine models what-if FinOps scenarios (REQ-FINOPS-053).
type Engine struct{}

// NewEngine creates scenario engine.
func NewEngine() *Engine { return &Engine{} }

// Run evaluates projected cost and carbon deltas.
func (e *Engine) Run(req domain.ScenarioRequest, baselineCost, baselineCo2e float64) domain.ScenarioResult {
	if baselineCost == 0 {
		baselineCost = 10000
	}
	if baselineCo2e == 0 {
		baselineCo2e = 1200
	}
	costMult, co2Mult, summary := e.factors(req)
	projectedCost := baselineCost * costMult
	projectedCo2e := baselineCo2e * co2Mult
	return domain.ScenarioResult{
		ID: uuid.NewString(), Name: req.Name, ScenarioType: req.ScenarioType,
		BaselineCostUSD: baselineCost, ProjectedCostUSD: projectedCost,
		CostDeltaUSD: projectedCost - baselineCost,
		BaselineCo2eKg: baselineCo2e, ProjectedCo2eKg: projectedCo2e,
		Co2eDeltaKg: projectedCo2e - baselineCo2e,
		Summary: summary, CreatedAt: time.Now().UTC(),
	}
}

func (e *Engine) factors(req domain.ScenarioRequest) (costMult, co2Mult float64, summary string) {
	costMult, co2Mult = 1.0, 1.0
	switch req.ScenarioType {
	case "region_migration":
		from := req.Params["from_region"]
		to := req.Params["to_region"]
		costMult = 0.92
		co2Mult = 0.78
		summary = fmt.Sprintf("Migrate workloads from %s to %s: ~8%% cost reduction, ~22%% CO₂e reduction (higher renewable mix)", from, to)
	case "instance_family":
		family := req.Params["target_family"]
		costMult = 0.85
		co2Mult = 0.95
		summary = fmt.Sprintf("Move to %s instance family: ~15%% compute savings with minimal carbon change", family)
	case "commitment_purchase":
		term := req.Params["term_months"]
		costMult = 0.82
		co2Mult = 1.0
		summary = fmt.Sprintf("Purchase %s-month commitment: ~18%% effective discount, carbon neutral", term)
	default:
		summary = "Custom scenario: review params for projected impact"
	}
	if strings.Contains(req.Scope, "payment") {
		costMult *= 1.05
	}
	return costMult, co2Mult, summary
}
