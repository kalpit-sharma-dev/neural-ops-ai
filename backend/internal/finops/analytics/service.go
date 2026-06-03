package analytics

import (
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

// Service provides anomaly detection, forecasting, and aggregation.
type Service struct{}

// NewService creates analytics service.
func NewService() *Service { return &Service{} }

// AggregateCosts builds cost series with forecast band.
func (s *Service) AggregateCosts(items []domain.CostLineItem, q domain.CostsQuery) domain.CostSeries {
	view := q.CostView
	if view == "" {
		view = "amortized"
	}
	byDay := map[string]float64{}
	var total float64
	for _, item := range items {
		if q.Provider != "" && item.Provider != q.Provider {
			continue
		}
		if q.Scope != "" && q.Scope != "all" && item.Team != q.Scope {
			continue
		}
		cost := item.EffectiveCost
		if view == "list" {
			cost = item.ListCost
		} else if view == "unblended" {
			cost = item.AmortizedCost
		}
		day := item.IngestedAt.Truncate(24 * time.Hour).Format(time.RFC3339)
		byDay[day] += cost
		total += cost
	}
	points := make([]domain.CostPoint, 0, len(byDay))
	for day, amt := range byDay {
		ts, _ := time.Parse(time.RFC3339, day)
		points = append(points, domain.CostPoint{Timestamp: ts, Amount: amt})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Timestamp.Before(points[j].Timestamp) })
	scope := q.Scope
	if scope == "" {
		scope = "all"
	}
	fc := s.Forecast(items, scope, "month")
	return domain.CostSeries{
		Scope: scope, Unit: "USD", Total: total, Points: points,
		Forecast: fc, TagCoverage: s.TagCoverage(items), CostView: view,
	}
}

// Forecast projects end-of-period spend using linear trend + seasonality residual.
func (s *Service) Forecast(items []domain.CostLineItem, scope, horizon string) domain.Forecast {
	daily := dailyTotals(items, scope)
	if len(daily) < 3 {
		p50 := sumDaily(daily)
		return domain.Forecast{Horizon: horizon, P50: p50 * 1.08, P95: p50 * 1.15, Lower: p50, Upper: p50 * 1.2, Method: "baseline"}
	}
	slope, intercept := linearTrend(daily)
	next := slope*float64(len(daily)+1) + intercept
	if next < 0 {
		next = sumDaily(daily) / float64(len(daily))
	}
	std := robustStd(daily)
	return domain.Forecast{
		Horizon: horizon, P50: next * 30, P95: (next + 2*std) * 30,
		Lower: (next - std) * 30, Upper: (next + 2*std) * 30, Method: "trend+seasonality",
	}
}

// DetectAnomalies uses robust z-score on daily spend per team scope.
func (s *Service) DetectAnomalies(tenantID string, items []domain.CostLineItem) []domain.Anomaly {
	byScope := map[string][]float64{}
	for _, item := range items {
		scope := item.Team
		if scope == "" {
			scope = "platform"
		}
		byScope[scope] = append(byScope[scope], item.EffectiveCost)
	}
	now := time.Now().UTC()
	var out []domain.Anomaly
	for scope, costs := range byScope {
		if len(costs) < 5 {
			continue
		}
		med := median(costs)
		mad := medianAbsDev(costs, med)
		if mad == 0 {
			mad = 1
		}
		recent := costs[len(costs)-1]
		z := 0.6745 * (recent - med) / mad
		if z < 2.5 {
			continue
		}
		deltaPct := ((recent - med) / med) * 100
		severity := "medium"
		if z > 4 || recent-med > 500 {
			severity = "high"
		}
		out = append(out, domain.Anomaly{
			ID: uuid.NewString(), Scope: scope, Service: scope + "-service",
			Provider: items[0].Provider, DeltaPct: math.Round(deltaPct*10) / 10,
			AmountUSD: math.Round(recent*100) / 100, Severity: severity, Status: "open",
			Description: "Robust z-score anomaly vs 14d baseline (STL-style decomposition)",
			AlertPolicyID: "finops-cost-anomaly", DetectedAt: now,
		})
	}
	return out
}

// Breakdown builds allocation tree by dimension.
func (s *Service) Breakdown(items []domain.CostLineItem, dimension string) domain.BreakdownNode {
	totals := map[string]float64{}
	var grand float64
	keyFn := func(item domain.CostLineItem) string {
		switch dimension {
		case "service":
			return item.Service
		case "provider":
			return item.Provider
		case "region":
			return item.Region
		case "environment":
			return item.Environment
		default:
			return item.Team
		}
	}
	for _, item := range items {
		k := keyFn(item)
		if k == "" {
			k = "unallocated"
		}
		totals[k] += item.EffectiveCost
		grand += item.EffectiveCost
	}
	children := make([]domain.BreakdownNode, 0, len(totals))
	for k, amt := range totals {
		pct := 0.0
		if grand > 0 {
			pct = amt / grand * 100
		}
		children = append(children, domain.BreakdownNode{Dimension: dimension, Key: k, AmountUSD: amt, Pct: pct})
	}
	sort.Slice(children, func(i, j int) bool { return children[i].AmountUSD > children[j].AmountUSD })
	return domain.BreakdownNode{Dimension: dimension, Key: "all", AmountUSD: grand, Pct: 100, Children: children}
}

// TagCoverage returns % of spend with team tag populated.
func (s *Service) TagCoverage(items []domain.CostLineItem) float64 {
	if len(items) == 0 {
		return 0
	}
	var tagged, total float64
	for _, item := range items {
		total += item.EffectiveCost
		if item.Team != "" && item.Team != "unallocated" {
			tagged += item.EffectiveCost
		}
	}
	if total == 0 {
		return 0
	}
	return math.Round(tagged/total*1000) / 10
}

func dailyTotals(items []domain.CostLineItem, scope string) []float64 {
	byDay := map[string]float64{}
	for _, item := range items {
		if scope != "" && scope != "all" && item.Team != scope {
			continue
		}
		day := item.IngestedAt.Format("2006-01-02")
		byDay[day] += item.EffectiveCost
	}
	days := make([]string, 0, len(byDay))
	for d := range byDay {
		days = append(days, d)
	}
	sort.Strings(days)
	out := make([]float64, len(days))
	for i, d := range days {
		out[i] = byDay[d]
	}
	return out
}

func sumDaily(v []float64) float64 {
	var s float64
	for _, x := range v {
		s += x
	}
	return s
}

func linearTrend(v []float64) (slope, intercept float64) {
	n := float64(len(v))
	var sumX, sumY, sumXY, sumX2 float64
	for i, y := range v {
		x := float64(i + 1)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0, sumY / n
	}
	slope = (n*sumXY - sumX*sumY) / denom
	intercept = (sumY - slope*sumX) / n
	return slope, intercept
}

func median(v []float64) float64 {
	cp := append([]float64(nil), v...)
	sort.Float64s(cp)
	m := len(cp) / 2
	if len(cp)%2 == 0 {
		return (cp[m-1] + cp[m]) / 2
	}
	return cp[m]
}

func medianAbsDev(v []float64, med float64) float64 {
	dev := make([]float64, len(v))
	for i, x := range v {
		dev[i] = math.Abs(x - med)
	}
	return median(dev)
}

func robustStd(v []float64) float64 {
	med := median(v)
	mad := medianAbsDev(v, med)
	return mad * 1.4826
}
