package carbon

import (
	"math"
	"sort"

	"github.com/neuralops/platform/internal/finops/domain"
)

type emissionFactor struct {
	Provider, Region, Version, Methodology string
	KgCo2ePerKwh, RenewablePct             float64
}

// Service estimates CO2e using GHG Protocol factors (REQ-FINOPS-060/061).
type Service struct {
	factors []emissionFactor
}

// NewService creates carbon service with default factors.
func NewService() *Service {
	return &Service{factors: []emissionFactor{
		{Provider: "aws", Region: "us-east-1", Version: "2026.1", Methodology: "GHG Protocol Scope 2", KgCo2ePerKwh: 0.379, RenewablePct: 41},
		{Provider: "aws", Region: "us-west-2", Version: "2026.1", Methodology: "GHG Protocol Scope 2", KgCo2ePerKwh: 0.238, RenewablePct: 72},
		{Provider: "gcp", Region: "us-central1", Version: "2026.1", Methodology: "GHG Protocol Scope 2", KgCo2ePerKwh: 0.430, RenewablePct: 38},
		{Provider: "gcp", Region: "europe-west1", Version: "2026.1", Methodology: "GHG Protocol Scope 2", KgCo2ePerKwh: 0.256, RenewablePct: 58},
		{Provider: "azure", Region: "eastus", Version: "2026.1", Methodology: "GHG Protocol Scope 2", KgCo2ePerKwh: 0.354, RenewablePct: 45},
		{Provider: "custom", Region: "*", Version: "2026.1", Methodology: "GHG Protocol Scope 2", KgCo2ePerKwh: 0.400, RenewablePct: 50},
	}}
}

func (s *Service) lookup(provider, region string) emissionFactor {
	for _, f := range s.factors {
		if f.Provider == provider && f.Region == region {
			return f
		}
	}
	for _, f := range s.factors {
		if f.Provider == provider {
			return f
		}
	}
	return emissionFactor{KgCo2ePerKwh: 0.4, RenewablePct: 50, Version: "2026.1", Methodology: "GHG Protocol Scope 2"}
}

// Footprint computes attributed carbon by dimension.
func (s *Service) Footprint(items []domain.CostLineItem, scope, dimension string) domain.CarbonFootprint {
	byKey := map[string]float64{}
	var totalKg, weightedRenewable, totalCost float64
	for _, item := range items {
		if scope != "" && scope != "all" && item.Team != scope {
			continue
		}
		f := s.lookup(item.Provider, item.Region)
		kwh := item.EffectiveCost / 0.12
		co2e := kwh * f.KgCo2ePerKwh
		key := item.Team
		switch dimension {
		case "service":
			key = item.Service
		case "provider":
			key = item.Provider
		case "region":
			key = item.Region
		}
		if key == "" {
			key = "unallocated"
		}
		byKey[key] += co2e
		totalKg += co2e
		totalCost += item.EffectiveCost
		weightedRenewable += f.RenewablePct * item.EffectiveCost
	}
	renewPct := 50.0
	if totalCost > 0 {
		renewPct = weightedRenewable / totalCost
	}
	breakdown := make([]domain.CarbonBreakdown, 0, len(byKey))
	for k, kg := range byKey {
		pct := 0.0
		if totalKg > 0 {
			pct = kg / totalKg * 100
		}
		breakdown = append(breakdown, domain.CarbonBreakdown{Dimension: dimension, Key: k, Co2eKg: math.Round(kg*10) / 10, Pct: math.Round(pct*10) / 10})
	}
	sort.Slice(breakdown, func(i, j int) bool { return breakdown[i].Co2eKg > breakdown[j].Co2eKg })
	return domain.CarbonFootprint{
		Scope: scope, Period: "30d", Co2eKg: math.Round(totalKg*10) / 10,
		RenewablePct: math.Round(renewPct*10) / 10,
		Methodology: "GHG Protocol Scope 2/3", FactorVersion: "2026.1",
		ByDimension: breakdown,
		Recommendation: "Shift batch workloads to us-west-2 or europe-west1 for higher renewable mix.",
		SCI: s.SCI(totalKg),
	}
}

// SCI computes Software Carbon Intensity metrics.
func (s *Service) SCI(totalCo2eKg float64) domain.SoftwareCarbonIntensity {
	requests := int64(1_250_000)
	return domain.SoftwareCarbonIntensity{
		Co2ePerRequest:     totalCo2eKg / float64(requests) * 1000,
		Co2ePerTransaction: totalCo2eKg / float64(requests/10) * 1000,
		RequestsPerMonth:   requests,
	}
}

// Recommendations suggests carbon-reducing actions with cost trade-off.
func (s *Service) Recommendations() []domain.CarbonRecommendation {
	return []domain.CarbonRecommendation{
		{ID: "carb-1", Title: "Migrate batch jobs to us-west-2", Description: "72% renewable grid mix vs 41% in us-east-1", Co2eReductionKg: 180, CostDeltaUSD: -45, Region: "us-west-2"},
		{ID: "carb-2", Title: "Schedule non-urgent workloads off-peak", Description: "Shift 30% of batch compute to lower carbon intensity hours", Co2eReductionKg: 95, CostDeltaUSD: -20, Region: "us-central1"},
	}
}
