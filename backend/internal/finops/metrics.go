package finops

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ingestLineItemsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "finops_ingest_line_items_total",
		Help: "FinOps billing line items ingested",
	}, []string{"tenant", "provider"})
	ingestLagSeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "finops_ingest_lag_seconds",
		Help: "Seconds since last successful FinOps ingest",
	}, []string{"tenant"})
	reconciliationDriftPct = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "finops_reconciliation_drift_pct",
		Help: "Invoice reconciliation drift percentage",
	}, []string{"tenant", "provider"})
	anomalyDetectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "finops_anomaly_detected_total",
		Help: "Cost anomalies detected",
	}, []string{"tenant", "severity"})
	recommendationSavingsUSD = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "finops_recommendation_savings_usd",
		Help: "Projected monthly savings from open recommendations",
	}, []string{"tenant", "type"})
	queryLatencySeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "finops_query_latency_seconds",
		Help:    "FinOps API query latency",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})
	budgetAlertTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "finops_budget_alert_total",
		Help: "Budget threshold alerts fired",
	}, []string{"tenant", "threshold"})
)

func recordIngest(tenant, provider string, count int, driftPct float64) {
	if tenant == "" {
		tenant = "default"
	}
	ingestLineItemsTotal.WithLabelValues(tenant, provider).Add(float64(count))
	reconciliationDriftPct.WithLabelValues(tenant, provider).Set(driftPct)
	ingestLagSeconds.WithLabelValues(tenant).Set(0)
}

func recordAnomaly(tenant, severity string) {
	if tenant == "" {
		tenant = "default"
	}
	anomalyDetectedTotal.WithLabelValues(tenant, severity).Inc()
}

func recordRecommendationSavings(tenant, recType string, usd float64) {
	if tenant == "" {
		tenant = "default"
	}
	recommendationSavingsUSD.WithLabelValues(tenant, recType).Set(usd)
}

func observeQuery(operation string, seconds float64) {
	queryLatencySeconds.WithLabelValues(operation).Observe(seconds)
}

// ObserveQuery records FinOps API query latency for Prometheus (REQ §9).
func ObserveQuery(operation string, seconds float64) {
	observeQuery(operation, seconds)
}

func recordBudgetAlert(tenant, threshold string) {
	if tenant == "" {
		tenant = "default"
	}
	budgetAlertTotal.WithLabelValues(tenant, threshold).Inc()
}

func recordIngestLag(tenant string, lagSeconds float64) {
	if tenant == "" {
		tenant = "default"
	}
	ingestLagSeconds.WithLabelValues(tenant).Set(lagSeconds)
}
