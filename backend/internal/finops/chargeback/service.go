package chargeback

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

// Service generates showback/chargeback statements (REQ-FINOPS-014).
type Service struct{}

// NewService creates chargeback service.
func NewService() *Service { return &Service{} }

// GenerateStatement builds exportable statement for a cost-center.
func (s *Service) GenerateStatement(items []domain.CostLineItem, costCenter, mode, period string) domain.ChargebackStatement {
	if mode == "" {
		mode = "showback"
	}
	if period == "" {
		period = "monthly"
	}
	byTeam := map[string]float64{}
	var total float64
	for _, item := range items {
		cc := item.CostCenter
		if cc == "" {
			cc = item.Team
		}
		if costCenter != "" && costCenter != "all" && cc != costCenter {
			continue
		}
		byTeam[item.Team] += item.EffectiveCost
		total += item.EffectiveCost
	}
	lines := make([]domain.ChargebackLine, 0, len(byTeam))
	for team, amt := range byTeam {
		pct := 0.0
		if total > 0 {
			pct = amt / total * 100
		}
		lines = append(lines, domain.ChargebackLine{
			Team: team, Service: team + "-services", AmountUSD: amt, AllocatedPct: pct,
		})
	}
	id := uuid.NewString()
	return domain.ChargebackStatement{
		ID: id, CostCenter: costCenter, Mode: mode, Period: period,
		TotalUSD: total, Lines: lines,
		ExportURL: fmt.Sprintf("/api/v1/finops/chargeback/statements/%s/export?format=csv", id),
		GeneratedAt: time.Now().UTC(),
	}
}

// ExportCSV renders statement as CSV bytes.
func (s *Service) ExportCSV(st domain.ChargebackStatement) string {
	out := "team,service,amount_usd,allocated_pct\n"
	for _, l := range st.Lines {
		out += fmt.Sprintf("%s,%s,%.2f,%.1f\n", l.Team, l.Service, l.AmountUSD, l.AllocatedPct)
	}
	out += fmt.Sprintf("total,,%.2f,100.0\n", st.TotalUSD)
	return out
}
