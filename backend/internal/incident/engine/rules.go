package engine

import (
	"strings"

	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/platform/db"
)

// ServiceMetrics holds rolling metrics for incident rule evaluation.
type ServiceMetrics struct {
	TotalLogs   int64
	ErrorLogs   int64
	P99LatencyMs float64
	AnomalyScore float64
	PaymentDown  bool
}

// ErrorRate returns the current error rate.
func (m ServiceMetrics) ErrorRate() float64 {
	if m.TotalLogs == 0 {
		return 0
	}
	return float64(m.ErrorLogs) / float64(m.TotalLogs)
}

// EvaluateSeverity applies incident creation rules.
func EvaluateSeverity(service string, metrics ServiceMetrics, paymentServices []string) domain.IncidentSeverity {
	if metrics.PaymentDown || (isPaymentService(service, paymentServices) && metrics.ErrorRate() >= 0.5) {
		return domain.IncidentSeverityP1
	}
	if metrics.ErrorRate() > 0.5 {
		return domain.IncidentSeverityP1
	}
	if metrics.ErrorRate() > 0.2 || metrics.P99LatencyMs > 5000 {
		return domain.IncidentSeverityP2
	}
	if metrics.ErrorRate() > 0.05 || metrics.AnomalyScore > 0.8 {
		return domain.IncidentSeverityP3
	}
	if metrics.AnomalyScore > 0 {
		return domain.IncidentSeverityP4
	}
	return ""
}

func isPaymentService(service string, paymentServices []string) bool {
	if db.IsPaymentService(service) {
		return true
	}
	service = strings.ToLower(service)
	for _, candidate := range paymentServices {
		if service == strings.ToLower(candidate) {
			return true
		}
	}
	return false
}

// ShouldCreateIncident returns true when severity warrants incident creation.
func ShouldCreateIncident(severity domain.IncidentSeverity) bool {
	return severity != ""
}
