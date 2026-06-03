package commitments

import (
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

// Service reports RI/SP/CUD coverage and purchase recommendations (REQ-FINOPS-040/041).
type Service struct{}

// NewService creates commitment service.
func NewService() *Service { return &Service{} }

// List returns seeded commitment inventory.
func (s *Service) List(provider, status string) []domain.Commitment {
	now := time.Now().UTC()
	all := []domain.Commitment{
		{ID: "cmt-1", Provider: "aws", CommitmentType: "savings_plan", Region: "us-east-1", CoveragePct: 72, UtilizationPct: 88, MonthlyCommitUSD: 4200, ExpiresAt: now.AddDate(0, 8, 0), Status: "active"},
		{ID: "cmt-2", Provider: "aws", CommitmentType: "reserved_instance", Region: "us-east-1", CoveragePct: 45, UtilizationPct: 62, MonthlyCommitUSD: 1800, ExpiresAt: now.AddDate(0, 3, 0), Status: "active"},
		{ID: "cmt-3", Provider: "gcp", CommitmentType: "cud", Region: "us-central1", CoveragePct: 58, UtilizationPct: 91, MonthlyCommitUSD: 3100, ExpiresAt: now.AddDate(1, 0, 0), Status: "active"},
		{ID: "cmt-4", Provider: "azure", CommitmentType: "reserved_vm", Region: "eastus", CoveragePct: 38, UtilizationPct: 55, MonthlyCommitUSD: 2200, ExpiresAt: now.AddDate(0, 1, 0), Status: "expiring"},
	}
	out := make([]domain.Commitment, 0)
	for _, c := range all {
		if provider != "" && c.Provider != provider {
			continue
		}
		if status != "" && c.Status != status {
			continue
		}
		out = append(out, c)
	}
	return out
}

// Recommendations suggests commitment purchases based on steady-state usage.
func (s *Service) Recommendations() []domain.CommitmentRecommendation {
	return []domain.CommitmentRecommendation{
		{ID: uuid.NewString(), Provider: "aws", CommitmentType: "compute_savings_plan", TermMonths: 12, BreakEvenMonths: 4.2, MonthlySavings: 890, RiskScore: 0.2, Description: "Steady EC2/Fargate usage supports 1yr compute SP — 18% effective discount"},
		{ID: uuid.NewString(), Provider: "gcp", CommitmentType: "cud", TermMonths: 12, BreakEvenMonths: 5.1, MonthlySavings: 620, RiskScore: 0.25, Description: "Cloud Run baseline supports 1yr CUD in us-central1"},
		{ID: uuid.NewString(), Provider: "azure", CommitmentType: "reserved_vm", TermMonths: 36, BreakEvenMonths: 8.0, MonthlySavings: 1100, RiskScore: 0.35, Description: "App Service steady load — 3yr RI maximizes discount"},
	}
}
