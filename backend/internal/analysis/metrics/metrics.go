package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	analysisProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "analysis_processed_total",
		Help: "Total number of analyzed records",
	}, []string{"service", "classification"})

	analysisLatencySeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "analysis_latency_seconds",
		Help:    "Analysis pipeline latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"analysis_type"})

	llmAPICallsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "llm_api_calls_total",
		Help: "Total number of LLM API calls",
	}, []string{"backend", "method"})

	llmAPILatencySeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "llm_api_latency_seconds",
		Help:    "LLM API latency in seconds",
		Buckets: prometheus.DefBuckets,
	})
)

// IncAnalysisProcessed increments processed counter.
func IncAnalysisProcessed(service, classification string) {
	analysisProcessedTotal.WithLabelValues(service, classification).Inc()
}

// ObserveAnalysisLatency records analysis latency.
func ObserveAnalysisLatency(analysisType string, duration time.Duration) {
	analysisLatencySeconds.WithLabelValues(analysisType).Observe(duration.Seconds())
}

// IncLLMAPICall increments LLM API call counter.
func IncLLMAPICall(backend, method string) {
	llmAPICallsTotal.WithLabelValues(backend, method).Inc()
}

// ObserveLLMLatency records LLM API latency.
func ObserveLLMLatency(duration time.Duration) {
	llmAPILatencySeconds.Observe(duration.Seconds())
}
