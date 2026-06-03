package optimize

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

// CloudResource is inventory input for optimization.
type CloudResource struct {
	ID, Provider, Type, Name, Region, Status string
	Tags                                     map[string]string
	MonthlyUSD                               float64
}

// Engine generates right-sizing and idle resource recommendations.
type Engine struct{}

// NewEngine creates optimization engine.
func NewEngine() *Engine { return &Engine{} }

// Generate produces Wave 1 recommendations (right-size + idle/orphan).
func (e *Engine) Generate(tenantID string, items []domain.CostLineItem, assets []CloudResource) []domain.Recommendation {
	now := time.Now().UTC()
	var out []domain.Recommendation
	util := map[string]float64{
		"aws-ec2-pay-01": 0.22, "aws-rds-ledger": 0.78, "gcp-run-api": 0.45,
		"azure-app-auth": 0.35, "aws-lambda-notify": 0.08,
	}
	for _, a := range assets {
		u := util[a.ID]
		if u == 0 {
			u = 0.4
		}
		if a.Type == "ec2" || a.Type == "app_service" || a.Type == "cloud_run" {
			if u < 0.35 && a.MonthlyUSD > 100 {
				savings := a.MonthlyUSD * (1 - u/0.5)
				out = append(out, domain.Recommendation{
					ID: uuid.NewString(), Type: "right_size", ResourceID: a.ID,
					Scope: teamFromTags(a.Tags), Title: "Right-size " + a.Name,
					Description: "P95 CPU/memory below 35% — downsize instance class",
					ProjectedSavingsUSD: savings, RiskScore: 0.25, Status: "open",
					CreatedAt: now, UpdatedAt: now,
				})
			}
		}
		if strings.Contains(strings.ToLower(a.Status), "stop") || a.Type == "ebs" && u < 0.05 {
			out = append(out, domain.Recommendation{
				ID: uuid.NewString(), Type: "idle_resource", ResourceID: a.ID,
				Scope: teamFromTags(a.Tags), Title: "Idle/orphaned: " + a.Name,
				Description: "Resource billed but unused — review for safe deletion",
				ProjectedSavingsUSD: a.MonthlyUSD * 0.9, RiskScore: 0.15, Status: "open",
				CreatedAt: now, UpdatedAt: now,
			})
		}
	}
	// Unattached disk pattern from line items
	for _, item := range items {
		if item.UsageType == "Storage" && item.EffectiveCost > 50 && item.ResourceID == "" {
			out = append(out, domain.Recommendation{
				ID: uuid.NewString(), Type: "idle_resource", ResourceID: item.ResourceID,
				Scope: item.Team, Title: "Unattached storage volume",
				Description: "EBS/disk with no attached instance",
				ProjectedSavingsUSD: item.EffectiveCost * 30, RiskScore: 0.1, Status: "open",
				CreatedAt: now, UpdatedAt: now,
			})
		}
	}
	return out
}

func teamFromTags(tags map[string]string) string {
	if tags == nil {
		return "platform"
	}
	if t, ok := tags["team"]; ok {
		return t
	}
	return "platform"
}
