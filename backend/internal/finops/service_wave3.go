package finops

import (
	"context"
	"time"

	"github.com/neuralops/platform/internal/finops/domain"
)

// GenerateChargebackStatement creates showback/chargeback export (REQ-FINOPS-014).
func (s *Service) GenerateChargebackStatement(ctx context.Context, tenantID, costCenter, mode, period string) (domain.ChargebackStatement, error) {
	if mode == "" {
		mode = s.governance.GetChargebackMode(tenantID)
	}
	items, err := s.repo.ListLineItems(ctx, tenantID, time.Now().AddDate(0, -1, 0), time.Now(), "", "")
	if err != nil {
		return domain.ChargebackStatement{}, err
	}
	st := s.chargeback.GenerateStatement(items, costCenter, mode, period)
	if err := s.repo.SaveChargebackStatement(ctx, tenantID, st); err != nil {
		return st, err
	}
	return st, nil
}

// ListChargebackStatements returns historical statements.
func (s *Service) ListChargebackStatements(ctx context.Context, tenantID, costCenter string) ([]domain.ChargebackStatement, error) {
	return s.repo.ListChargebackStatements(ctx, tenantID, costCenter)
}

// ExportChargebackStatement returns CSV export for a statement.
func (s *Service) ExportChargebackStatement(ctx context.Context, tenantID, id string) (string, domain.ChargebackStatement, error) {
	st, err := s.repo.GetChargebackStatement(ctx, tenantID, id)
	if err != nil {
		return "", st, err
	}
	return s.chargeback.ExportCSV(st), st, nil
}

// RunScenario models what-if cost/carbon impact (REQ-FINOPS-053).
func (s *Service) RunScenario(ctx context.Context, tenantID string, req domain.ScenarioRequest, actor string) (domain.ScenarioResult, error) {
	items, _ := s.repo.ListLineItems(ctx, tenantID, time.Now().AddDate(0, -1, 0), time.Now(), "", req.Scope)
	var baselineCost float64
	for _, item := range items {
		baselineCost += item.EffectiveCost
	}
	fp := s.carbon.Footprint(items, req.Scope, "team")
	result := s.scenarios.Run(req, baselineCost, fp.Co2eKg)
	_ = s.repo.SaveScenarioRun(ctx, tenantID, result)
	s.audit(ctx, tenantID, actor, "run_scenario", "scenario", result.ID, map[string]string{"type": req.ScenarioType})
	return result, nil
}

// ListScenarios returns recent scenario runs.
func (s *Service) ListScenarios(ctx context.Context, tenantID string, limit int) ([]domain.ScenarioResult, error) {
	return s.repo.ListScenarioRuns(ctx, tenantID, limit)
}

// ListCommitmentAlerts detects expiring/under-utilized commitments (REQ-FINOPS-042).
func (s *Service) ListCommitmentAlerts(provider string) []domain.CommitmentAlert {
	cmts := s.commitments.List(provider, "")
	return s.commitments.DetectAlerts(cmts)
}

// RouteCommitmentAlerts sends commitment alerts through alert pipeline.
func (s *Service) RouteCommitmentAlerts(ctx context.Context, tenantID string) []domain.CommitmentAlert {
	alerts := s.ListCommitmentAlerts("")
	if s.alerter != nil {
		for _, a := range alerts {
			if ca, ok := s.alerter.(CommitmentAlerter); ok {
				ca.RouteCommitmentAlert(ctx, tenantID, a)
			}
		}
	}
	return alerts
}

// CommitmentAlerter routes commitment alerts.
type CommitmentAlerter interface {
	RouteCommitmentAlert(ctx context.Context, tenantID string, a domain.CommitmentAlert)
}

// ApplyCarbonAction simulates or applies carbon reduction (REQ-FINOPS-062).
func (s *Service) ApplyCarbonAction(ctx context.Context, tenantID string, req domain.CarbonActionRequest, actor string) (domain.CarbonActionResult, error) {
	var rec domain.CarbonRecommendation
	for _, r := range s.carbon.EnhancedRecommendations() {
		if r.ID == req.RecommendationID {
			rec = r
			break
		}
	}
	if rec.ID == "" {
		rec = domain.CarbonRecommendation{ID: req.RecommendationID}
	}
	result := s.carbon.ApplyAction(req, rec)
	s.audit(ctx, tenantID, actor, req.Action, "carbon_action", result.ID, map[string]string{"rec": req.RecommendationID})
	return result, nil
}

// ListCarbonRecommendationsEnhanced returns carbon actions with trade-offs.
func (s *Service) ListCarbonRecommendationsEnhanced() []domain.CarbonRecommendation {
	return s.carbon.EnhancedRecommendations()
}

// ListGovernancePolicies returns FinOps governance config.
func (s *Service) ListGovernancePolicies(tenantID string) []domain.GovernancePolicy {
	return s.governance.List(tenantID)
}

// UpdateAllocationRuleAudited updates rule with audit trail.
func (s *Service) UpdateAllocationRuleAudited(ctx context.Context, tenantID string, r domain.AllocationRule, actor string) (domain.AllocationRule, error) {
	updated, err := s.repo.UpdateAllocationRule(ctx, tenantID, r)
	if err != nil {
		return updated, err
	}
	s.audit(ctx, tenantID, actor, "update", "allocation_rule", updated.ID, map[string]string{"name": updated.Name})
	return updated, nil
}

// DeleteAllocationRuleAudited deletes rule with audit trail.
func (s *Service) DeleteAllocationRuleAudited(ctx context.Context, tenantID, id, actor string) error {
	if err := s.repo.DeleteAllocationRule(ctx, tenantID, id); err != nil {
		return err
	}
	s.audit(ctx, tenantID, actor, "delete", "allocation_rule", id, nil)
	return nil
}

// UpsertGovernancePolicy updates governance settings.
func (s *Service) UpsertGovernancePolicy(ctx context.Context, tenantID string, p domain.GovernancePolicy, actor string) domain.GovernancePolicy {
	updated := s.governance.Upsert(tenantID, p)
	s.audit(ctx, tenantID, actor, "upsert", "governance_policy", updated.ID, map[string]string{"name": updated.Name})
	return updated
}
