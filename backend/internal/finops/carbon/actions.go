package carbon

import (
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

// ApplyAction simulates or applies a carbon reduction recommendation (REQ-FINOPS-062).
func (s *Service) ApplyAction(req domain.CarbonActionRequest, rec domain.CarbonRecommendation) domain.CarbonActionResult {
	status := "simulated"
	appliedAt := time.Time{}
	if req.Action == "apply" {
		status = "applied"
		appliedAt = time.Now().UTC()
	} else if req.Action == "dismiss" {
		status = "dismissed"
	}
	return domain.CarbonActionResult{
		ID: uuid.NewString(), RecommendationID: rec.ID, Status: status,
		Co2eReductionKg: rec.Co2eReductionKg, CostDeltaUSD: rec.CostDeltaUSD,
		AppliedAt: appliedAt,
	}
}

// EnhancedRecommendations returns carbon actions with explicit cost trade-offs.
func (s *Service) EnhancedRecommendations() []domain.CarbonRecommendation {
	recs := s.Recommendations()
	for i := range recs {
		if recs[i].Description == "" {
			recs[i].Description = "Cost trade-off included — negative values indicate savings"
		}
	}
	return recs
}
