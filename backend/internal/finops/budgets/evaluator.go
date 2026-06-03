package budgets

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

// Alert is a budget threshold crossing (REQ-FINOPS-050).
type Alert struct {
	ID          string  `json:"id"`
	BudgetID    string  `json:"budgetId"`
	BudgetName  string  `json:"budgetName"`
	Scope       string  `json:"scope"`
	Threshold   int     `json:"threshold"`
	BurnPct     float64 `json:"burnPct"`
	ForecastPct float64 `json:"forecastPct,omitempty"`
	Severity    string  `json:"severity"`
	Message     string  `json:"message"`
}

// Enrich computes spend and burn for budgets from line items.
func Enrich(budgets []domain.Budget, items []domain.CostLineItem) []domain.Budget {
	out := make([]domain.Budget, len(budgets))
	copy(out, budgets)
	for i := range out {
		var spend float64
		for _, item := range items {
			if item.Team == out[i].ScopeValue || out[i].ScopeValue == "all" {
				spend += item.EffectiveCost
			}
		}
		out[i].SpendUSD = spend
		if out[i].AmountUSD > 0 {
			out[i].BurnPct = spend / out[i].AmountUSD * 100
		}
	}
	return out
}

// EvaluateAlerts checks 50/80/100 thresholds and forecast overrun.
func EvaluateAlerts(budgets []domain.Budget, forecast domain.Forecast) []Alert {
	var out []Alert
	for _, b := range budgets {
		if b.AmountUSD <= 0 {
			continue
		}
		burn := b.BurnPct
		forecastPct := 0.0
		if forecast.P50 > 0 {
			forecastPct = forecast.P50 / b.AmountUSD * 100
		}
		thresholds := b.Thresholds
		if len(thresholds) == 0 {
			thresholds = []int{50, 80, 100}
		}
		for _, th := range thresholds {
			if burn >= float64(th) {
				out = append(out, newAlert(b, th, burn, forecastPct, false))
			}
		}
		if forecastPct >= 100 {
			out = append(out, newAlert(b, 100, burn, forecastPct, true))
		}
	}
	return out
}

func newAlert(b domain.Budget, th int, burn, forecastPct float64, forecastOverrun bool) Alert {
	severity := "info"
	if th >= 100 || forecastOverrun {
		severity = "high"
	} else if th >= 80 {
		severity = "medium"
	}
	msg := fmt.Sprintf("Budget %s at %.0f%% of $%.0f cap (threshold %d%%)", b.Name, burn, b.AmountUSD, th)
	if forecastOverrun {
		msg = fmt.Sprintf("Budget %s forecast to exceed cap: projected %.0f%% of $%.0f", b.Name, forecastPct, b.AmountUSD)
		th = 100
	}
	return Alert{
		ID: uuid.NewString(), BudgetID: b.ID, BudgetName: b.Name, Scope: b.ScopeValue,
		Threshold: th, BurnPct: burn, ForecastPct: forecastPct, Severity: severity, Message: msg,
	}
}

// StaleIngestLag returns seconds since last ingest for freshness monitoring.
func StaleIngestLag(lastIngest time.Time) float64 {
	if lastIngest.IsZero() {
		return 86400
	}
	return time.Since(lastIngest).Seconds()
}
