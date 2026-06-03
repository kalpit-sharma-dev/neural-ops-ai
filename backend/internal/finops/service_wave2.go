// Wave 2 FinOps capabilities: imports, carbon, commitments, unit economics, reports, audit.
package finops

import (
	"context"
	"fmt"
	"time"

	"github.com/neuralops/platform/internal/finops/analytics"
	"github.com/neuralops/platform/internal/finops/connectors"
	"github.com/neuralops/platform/internal/finops/domain"
)

// ImportCustomCost ingests a custom/SaaS cost feed (REQ-FINOPS-004/005).
func (s *Service) ImportCustomCost(ctx context.Context, tenantID string, req domain.ImportRequest, actor string) (domain.ImportBatch, error) {
	batch, items, err := connectors.ImportCustomFeed(ctx, tenantID, req)
	if err != nil {
		return domain.ImportBatch{}, err
	}
	ver, _ := s.repo.GetImportVersion(ctx, tenantID, batch.DedupeKey)
	batch.Version = ver + 1
	existing, _ := s.repo.ListLineItems(ctx, tenantID, time.Time{}, time.Time{}, "", "")
	deduped := connectors.DedupeLineItems(existing, items)
	if err := s.repo.ReplaceLineItems(ctx, tenantID, deduped); err != nil {
		return domain.ImportBatch{}, err
	}
	_ = s.repo.SaveImportBatch(ctx, tenantID, batch)
	s.audit(ctx, tenantID, actor, "import", "cost_feed", batch.ID, map[string]string{"source": req.Source})
	return batch, nil
}

// ListSharedSplits returns shared cost split rules.
func (s *Service) ListSharedSplits(ctx context.Context, tenantID string) ([]domain.SharedSplitRule, error) {
	return s.repo.ListSharedSplits(ctx, tenantID)
}

// ListTagSuggestions returns ML-assisted tagging suggestions.
func (s *Service) ListTagSuggestions(ctx context.Context, tenantID string) ([]domain.TagSuggestion, error) {
	items, err := s.repo.ListLineItems(ctx, tenantID, time.Now().AddDate(0, -1, 0), time.Now(), "", "")
	if err != nil {
		return nil, err
	}
	suggestions := s.allocator.SuggestTags(tenantID, items)
	_ = s.repo.SaveTagSuggestions(ctx, tenantID, suggestions)
	return s.repo.ListTagSuggestions(ctx, tenantID)
}

// GetAnomalyWithCause returns anomaly detail with probable cause correlation.
func (s *Service) GetAnomalyWithCause(ctx context.Context, tenantID, id string) (domain.Anomaly, error) {
	a, err := s.repo.GetAnomaly(ctx, tenantID, id)
	if err != nil {
		return a, err
	}
	if a.ProbableCause == "" {
		a.ProbableCause = analytics.CorrelateProbableCause(a, nil)
	}
	return a, nil
}

// SubmitAnomalyFeedbackWithTuning records feedback and adjusts sensitivity (REQ-FINOPS-022).
func (s *Service) SubmitAnomalyFeedbackWithTuning(ctx context.Context, tenantID, id, feedback, actor string) (domain.Anomaly, error) {
	a, err := s.repo.UpdateAnomalyFeedback(ctx, tenantID, id, feedback)
	if err != nil {
		return a, err
	}
	sens, _ := s.repo.GetAnomalySensitivity(ctx, tenantID, a.Scope)
	sens = analytics.AdjustSensitivity(sens, feedback)
	sens.Scope = a.Scope
	_ = s.repo.SaveAnomalySensitivity(ctx, tenantID, sens)
	s.audit(ctx, tenantID, actor, "feedback", "anomaly", id, map[string]string{"feedback": feedback})
	return a, nil
}

// ListCommitments returns RI/SP/CUD inventory.
func (s *Service) ListCommitments(provider, status string) []domain.Commitment {
	return s.commitments.List(provider, status)
}

// ListCommitmentRecommendations returns purchase recommendations.
func (s *Service) ListCommitmentRecommendations() []domain.CommitmentRecommendation {
	return s.commitments.Recommendations()
}

// GetCarbonFootprint returns enhanced carbon report by dimension.
func (s *Service) GetCarbonFootprint(ctx context.Context, tenantID, scope, dimension string) (domain.CarbonFootprint, error) {
	items, err := s.repo.ListLineItems(ctx, tenantID, time.Now().AddDate(0, -1, 0), time.Now(), "", scope)
	if err != nil {
		return domain.CarbonFootprint{}, err
	}
	if dimension == "" {
		dimension = "team"
	}
	return s.carbon.Footprint(items, scope, dimension), nil
}

// ListCarbonRecommendations returns carbon reduction actions.
func (s *Service) ListCarbonRecommendations() []domain.CarbonRecommendation {
	return s.ListCarbonRecommendationsEnhanced()
}

// UnitEconomics computes cost per business metric.
func (s *Service) UnitEconomics(ctx context.Context, tenantID, metric, scope string) (domain.UnitEconomics, error) {
	items, err := s.repo.ListLineItems(ctx, tenantID, time.Now().AddDate(0, -1, 0), time.Now(), "", scope)
	if err != nil {
		return domain.UnitEconomics{}, err
	}
	if metric == "" {
		metric = "request"
	}
	return s.analytics.UnitEconomics(items, metric, scope), nil
}

// ListReports returns generated FinOps reports.
func (s *Service) ListReports(scope string) []domain.FinOpsReport {
	return s.reports.ListReports(scope)
}

// ListReportSchedules returns scheduled reports.
func (s *Service) ListReportSchedules(tenantID string) []domain.ReportSchedule {
	return s.reports.ListSchedules(tenantID)
}

// CreateReportSchedule creates a report schedule.
func (s *Service) CreateReportSchedule(ctx context.Context, tenantID string, sch domain.ReportSchedule, actor string) domain.ReportSchedule {
	created := s.reports.CreateSchedule(tenantID, sch)
	s.audit(ctx, tenantID, actor, "create", "report_schedule", created.ID, map[string]string{"name": created.Name})
	return created
}

// ListAuditLog returns FinOps audit entries.
func (s *Service) ListAuditLog(tenantID string, limit int) []domain.AuditEntry {
	return s.auditLog.List(tenantID, limit)
}

// ListReconciliation returns ingest snapshots for invoice reconciliation (FIN-PROD-04).
func (s *Service) ListReconciliation(ctx context.Context, tenantID string) ([]domain.IngestSnapshot, error) {
	return s.repo.ListIngestSnapshots(ctx, tenantID)
}

// ExportAuditCSV returns finance audit log as CSV (FIN-PROD-09).
func (s *Service) ExportAuditCSV(tenantID string, limit int) []byte {
	entries := s.auditLog.List(tenantID, limit)
	if limit <= 0 {
		limit = 1000
	}
	var b []byte
	b = append(b, "id,actor,action,entity_type,entity_id,created_at,detail\n"...)
	for _, e := range entries {
		detail := ""
		if e.Detail != nil {
			for k, v := range e.Detail {
				detail += k + "=" + v + ";"
			}
		}
		line := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%q\n",
			e.ID, e.Actor, e.Action, e.EntityType, e.EntityID,
			e.CreatedAt.Format(time.RFC3339), detail)
		b = append(b, line...)
	}
	return b
}

func (s *Service) audit(ctx context.Context, tenantID, actor, action, entityType, entityID string, detail map[string]string) {
	if s.auditLog != nil {
		s.auditLog.Record(ctx, tenantID, actor, action, entityType, entityID, detail)
	}
}

// RecommendationActionWithAudit applies lifecycle action with audit trail.
func (s *Service) RecommendationActionWithAudit(ctx context.Context, tenantID, id, action, actor string) (domain.Recommendation, error) {
	r, err := s.repo.UpdateRecommendationStatus(ctx, tenantID, id, action)
	if err != nil {
		return r, err
	}
	s.audit(ctx, tenantID, actor, action, "recommendation", id, map[string]string{"status": r.Status})
	return r, nil
}

// CreateAllocationRuleAudited creates rule with audit.
func (s *Service) CreateAllocationRuleAudited(ctx context.Context, tenantID string, r domain.AllocationRule, actor string) (domain.AllocationRule, error) {
	created, err := s.repo.CreateAllocationRule(ctx, tenantID, r)
	if err != nil {
		return created, err
	}
	s.audit(ctx, tenantID, actor, "create", "allocation_rule", created.ID, map[string]string{"name": created.Name})
	return created, nil
}

// CreateBudgetAudited creates budget with audit.
func (s *Service) CreateBudgetAudited(ctx context.Context, tenantID string, b domain.Budget, actor string) (domain.Budget, error) {
	created, err := s.repo.CreateBudget(ctx, tenantID, b)
	if err != nil {
		return created, err
	}
	s.audit(ctx, tenantID, actor, "create", "budget", created.ID, map[string]string{"name": created.Name})
	return created, nil
}

// UpdateBudgetAudited updates budget with audit.
func (s *Service) UpdateBudgetAudited(ctx context.Context, tenantID string, b domain.Budget, actor string) (domain.Budget, error) {
	updated, err := s.repo.UpdateBudget(ctx, tenantID, b)
	if err != nil {
		return updated, err
	}
	s.audit(ctx, tenantID, actor, "update", "budget", updated.ID, nil)
	return updated, nil
}

// DeleteBudgetAudited deletes budget with audit.
func (s *Service) DeleteBudgetAudited(ctx context.Context, tenantID, id, actor string) error {
	if err := s.repo.DeleteBudget(ctx, tenantID, id); err != nil {
		return err
	}
	s.audit(ctx, tenantID, actor, "delete", "budget", id, nil)
	return nil
}
