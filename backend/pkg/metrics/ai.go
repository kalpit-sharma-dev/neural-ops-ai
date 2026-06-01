package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	aiAPIRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_api_requests_total",
			Help: "Total AI/LLM API requests",
		},
		[]string{"provider", "operation", "status"},
	)

	aiAPILatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_api_latency_seconds",
			Help:    "AI/LLM API latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider", "operation"},
	)
)

func init() {
	prometheus.MustRegister(aiAPIRequests, aiAPILatency)
}

// RecordAIRequest records AI provider usage for cost/latency dashboards.
func RecordAIRequest(provider, operation, status string, duration time.Duration) {
	aiAPIRequests.WithLabelValues(provider, operation, status).Inc()
	aiAPILatency.WithLabelValues(provider, operation).Observe(duration.Seconds())
}
