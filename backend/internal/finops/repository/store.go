package repository

import (
	"context"
	"time"

	"github.com/neuralops/platform/internal/finops/domain"
)

// Store persists FinOps operational state.
type Store interface {
	SaveLineItems(ctx context.Context, tenantID string, items []domain.CostLineItem) error
	ListLineItems(ctx context.Context, tenantID string, from, to time.Time, provider, scope string) ([]domain.CostLineItem, error)
	SaveIngestSnapshot(ctx context.Context, tenantID string, s domain.IngestSnapshot) error

	ListAllocationRules(ctx context.Context, tenantID string) ([]domain.AllocationRule, error)
	CreateAllocationRule(ctx context.Context, tenantID string, r domain.AllocationRule) (domain.AllocationRule, error)
	UpdateAllocationRule(ctx context.Context, tenantID string, r domain.AllocationRule) (domain.AllocationRule, error)
	DeleteAllocationRule(ctx context.Context, tenantID, id string) error

	ListBudgets(ctx context.Context, tenantID string) ([]domain.Budget, error)
	CreateBudget(ctx context.Context, tenantID string, b domain.Budget) (domain.Budget, error)
	UpdateBudget(ctx context.Context, tenantID string, b domain.Budget) (domain.Budget, error)
	DeleteBudget(ctx context.Context, tenantID, id string) error

	ListAnomalies(ctx context.Context, tenantID, scope, severity, status string) ([]domain.Anomaly, error)
	GetAnomaly(ctx context.Context, tenantID, id string) (domain.Anomaly, error)
	SaveAnomaly(ctx context.Context, tenantID string, a domain.Anomaly) error
	UpdateAnomalyFeedback(ctx context.Context, tenantID, id, feedback string) (domain.Anomaly, error)

	ListRecommendations(ctx context.Context, tenantID, recType, scope, status string) ([]domain.Recommendation, error)
	GetRecommendation(ctx context.Context, tenantID, id string) (domain.Recommendation, error)
	SaveRecommendation(ctx context.Context, tenantID string, r domain.Recommendation) error
	UpdateRecommendationStatus(ctx context.Context, tenantID, id, action string) (domain.Recommendation, error)

	// Wave 2
	SaveImportBatch(ctx context.Context, tenantID string, b domain.ImportBatch) error
	GetImportVersion(ctx context.Context, tenantID, dedupeKey string) (int, error)
	ReplaceLineItems(ctx context.Context, tenantID string, items []domain.CostLineItem) error
	ListSharedSplits(ctx context.Context, tenantID string) ([]domain.SharedSplitRule, error)
	SaveSharedSplits(ctx context.Context, tenantID string, rules []domain.SharedSplitRule) error
	ListTagSuggestions(ctx context.Context, tenantID string) ([]domain.TagSuggestion, error)
	SaveTagSuggestions(ctx context.Context, tenantID string, s []domain.TagSuggestion) error
	GetAnomalySensitivity(ctx context.Context, tenantID, scope string) (domain.AnomalySensitivity, error)
	SaveAnomalySensitivity(ctx context.Context, tenantID string, s domain.AnomalySensitivity) error
	ListIngestSnapshots(ctx context.Context, tenantID string) ([]domain.IngestSnapshot, error)

	// Wave 3
	SaveChargebackStatement(ctx context.Context, tenantID string, st domain.ChargebackStatement) error
	GetChargebackStatement(ctx context.Context, tenantID, id string) (domain.ChargebackStatement, error)
	ListChargebackStatements(ctx context.Context, tenantID, costCenter string) ([]domain.ChargebackStatement, error)
	SaveScenarioRun(ctx context.Context, tenantID string, r domain.ScenarioResult) error
	ListScenarioRuns(ctx context.Context, tenantID string, limit int) ([]domain.ScenarioResult, error)
}
