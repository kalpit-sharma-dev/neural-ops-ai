package observability

import (
	"context"
	"strings"

	"github.com/neuralops/platform/internal/finops/domain"
	"github.com/neuralops/platform/internal/streaming"
)

// FinOpsCloudAdapter adapts Store to domain.CloudAssets.
type FinOpsCloudAdapter struct {
	Store *Store
}

// ListCloudAssets implements domain.CloudAssets.
func (a *FinOpsCloudAdapter) ListCloudAssets(provider string) []domain.CloudAsset {
	if a.Store == nil {
		return nil
	}
	assets := a.Store.ListCloudAssets(provider)
	out := make([]domain.CloudAsset, len(assets))
	for i, x := range assets {
		out[i] = domain.CloudAsset{
			ID: x.ID, Provider: x.Provider, Type: x.Type, Name: x.Name,
			Region: x.Region, AccountID: x.AccountID, Status: x.Status,
			Tags: x.Tags, MonthlyUSD: x.MonthlyUSD,
		}
	}
	return out
}

// FinOpsAnomalyAlerter routes FinOps anomalies through alert policies and streaming.
type FinOpsAnomalyAlerter struct {
	Mem    *Store
	Stream *streaming.Publisher
}

// RouteAnomaly evaluates alert policy and publishes notification signal.
func (a *FinOpsAnomalyAlerter) RouteAnomaly(ctx context.Context, tenantID string, an domain.Anomaly) {
	if a == nil || a.Mem == nil {
		return
	}
	policyID := an.AlertPolicyID
	if policyID == "" {
		policyID = "finops-cost-anomaly"
	}
	severity := finOpsAlertSeverity(an.Severity)
	res := EvaluateAlertPolicy(a.Mem, policyID, an.Service, severity)
	recordAlertTrigger(tenantID, res.Matched)
	if a.Stream != nil && res.Matched && !res.Suppressed {
		_ = a.Stream.PublishAlertSignal(ctx, tenantID, streaming.AlertSignalPayload{
			PolicyID: policyID, Service: an.Service, Severity: severity, Count: 1,
		})
	}
}

// RouteBudgetAlert routes budget threshold alerts (REQ-FINOPS-050).
func (a *FinOpsAnomalyAlerter) RouteBudgetAlert(ctx context.Context, tenantID, budgetID, scope, severity, message string) {
	if a == nil || a.Mem == nil {
		return
	}
	policyID := "finops-budget-alert"
	sev := finOpsAlertSeverity(severity)
	service := scope + "-budget"
	res := EvaluateAlertPolicy(a.Mem, policyID, service, sev)
	recordAlertTrigger(tenantID, res.Matched)
	if a.Stream != nil && res.Matched && !res.Suppressed {
		_ = a.Stream.PublishAlertSignal(ctx, tenantID, streaming.AlertSignalPayload{
			PolicyID: policyID, Service: service, Severity: sev, Count: 1,
		})
	}
}

// RouteStaleIngest routes alerts when billing ingest is stale (REQ §9 dogfooding).
func (a *FinOpsAnomalyAlerter) RouteStaleIngest(ctx context.Context, tenantID string, lagSeconds float64) {
	if a == nil || a.Mem == nil {
		return
	}
	policyID := "finops-ingest-stale"
	sev := "P2"
	if lagSeconds > 86400*2 {
		sev = "P1"
	}
	res := EvaluateAlertPolicy(a.Mem, policyID, "finops-ingest", sev)
	recordAlertTrigger(tenantID, res.Matched)
	if a.Stream != nil && res.Matched && !res.Suppressed {
		_ = a.Stream.PublishAlertSignal(ctx, tenantID, streaming.AlertSignalPayload{
			PolicyID: policyID, Service: "finops-ingest", Severity: sev, Count: 1,
		})
	}
}

// RouteCommitmentAlert routes commitment expiry/utilization alerts (REQ-FINOPS-042).
func (a *FinOpsAnomalyAlerter) RouteCommitmentAlert(ctx context.Context, tenantID string, alert domain.CommitmentAlert) {
	if a == nil || a.Mem == nil {
		return
	}
	policyID := alert.AlertPolicyID
	if policyID == "" {
		policyID = "finops-commitment-alert"
	}
	severity := finOpsAlertSeverity(alert.Severity)
	service := alert.Provider + "-commitments"
	res := EvaluateAlertPolicy(a.Mem, policyID, service, severity)
	recordAlertTrigger(tenantID, res.Matched)
	if a.Stream != nil && res.Matched && !res.Suppressed {
		_ = a.Stream.PublishAlertSignal(ctx, tenantID, streaming.AlertSignalPayload{
			PolicyID: policyID, Service: service, Severity: severity, Count: 1,
		})
	}
}

func finOpsAlertSeverity(s string) string {
	switch strings.ToLower(s) {
	case "high", "critical":
		return "P1"
	case "medium":
		return "P2"
	default:
		return "P3"
	}
}
