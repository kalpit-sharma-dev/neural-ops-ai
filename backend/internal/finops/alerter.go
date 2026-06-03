package finops

import (
	"context"

	"github.com/neuralops/platform/internal/finops/domain"
)

// AnomalyAlerter routes cost anomalies through the platform alerting pipeline (REQ-FINOPS-021).
type AnomalyAlerter interface {
	RouteAnomaly(ctx context.Context, tenantID string, a domain.Anomaly)
}

// BudgetAlerter routes budget threshold alerts (REQ-FINOPS-050).
type BudgetAlerter interface {
	RouteBudgetAlert(ctx context.Context, tenantID string, budgetID, scope, severity, message string)
}

// SetAlerter configures optional alert routing after anomaly detection.
func (s *Service) SetAlerter(a AnomalyAlerter) {
	s.alerter = a
}

// SetBudgetAlerter configures budget threshold alert routing.
func (s *Service) SetBudgetAlerter(a BudgetAlerter) {
	s.budgetAlert = a
}
