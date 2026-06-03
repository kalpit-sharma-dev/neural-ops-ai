package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	srsQueryTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "neuralops_observability_unified_query_total",
		Help: "Unified cross-signal queries executed",
	}, []string{"tenant"})
	srsAlertTriggerTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "neuralops_observability_alert_trigger_total",
		Help: "Alert policy trigger evaluations",
	}, []string{"tenant", "matched"})
	srsGovernanceDenied = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "neuralops_observability_governance_denied_total",
		Help: "ABAC or residency denials at gateway",
	}, []string{"reason"})
)

func recordUnifiedQuery(tenant string) {
	if tenant == "" {
		tenant = "default"
	}
	srsQueryTotal.WithLabelValues(tenant).Inc()
}

func recordAlertTrigger(tenant string, matched bool) {
	if tenant == "" {
		tenant = "default"
	}
	flag := "false"
	if matched {
		flag = "true"
	}
	srsAlertTriggerTotal.WithLabelValues(tenant, flag).Inc()
}

func RecordGovernanceDenied(reason string) {
	if reason == "" {
		reason = "unknown"
	}
	srsGovernanceDenied.WithLabelValues(reason).Inc()
}
