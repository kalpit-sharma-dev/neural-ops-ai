package optimize

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

// GenerateStorageTiering adds lifecycle/storage recommendations (REQ-FINOPS-032).
func (e *Engine) GenerateStorageTiering(items []domain.CostLineItem) []domain.Recommendation {
	now := time.Now().UTC()
	var out []domain.Recommendation
	for _, item := range items {
		if item.EffectiveCost < 20 {
			continue
		}
		ut := strings.ToLower(item.UsageType + " " + item.Service)
		if strings.Contains(ut, "storage") || strings.Contains(ut, "s3") || strings.Contains(ut, "snapshot") {
			savings := item.EffectiveCost * 30 * 0.4
			out = append(out, domain.Recommendation{
				ID: uuid.NewString(), Type: "storage_tiering", ResourceID: item.ResourceID,
				Scope: item.Team, Title: "Transition to infrequent access tier",
				Description: "Object storage or snapshot eligible for IA/Glacier class — 40% savings estimate",
				ProjectedSavingsUSD: savings, RiskScore: 0.2, Status: "open",
				CreatedAt: now, UpdatedAt: now,
			})
		}
	}
	return out
}

// ApplyLifecycleAction handles recommendation lifecycle with ticket/autofix (REQ-FINOPS-033/034).
func ApplyLifecycleAction(r domain.Recommendation, action string) domain.Recommendation {
	now := time.Now().UTC()
	r.UpdatedAt = now
	switch action {
	case "acknowledge":
		r.Status = "acknowledged"
	case "in_progress":
		r.Status = "in_progress"
	case "dismiss":
		r.Status = "dismissed"
	case "realized":
		r.Status = "realized"
		if r.RealizedSavingsUSD == 0 {
			r.RealizedSavingsUSD = r.ProjectedSavingsUSD * 0.85
		}
	case "ticket":
		r.Status = "in_progress"
		r.TicketID = "FINOPS-" + strings.ToUpper(r.ID[:8])
		r.TicketURL = fmt.Sprintf("https://jira.example.com/browse/%s", r.TicketID)
	case "autofix":
		r.Status = "in_progress"
		r.TicketID = "AUTOFIX-" + strings.ToUpper(r.ID[:8])
		r.TicketURL = fmt.Sprintf("/ai/autofix/plans/%s", r.TicketID)
	default:
		r.Status = action
	}
	return r
}
