package observability

import (
	"testing"
	"time"
)

func TestMetricToPromQL(t *testing.T) {
	q := MetricToPromQL("error_rate", "payment-service")
	if q == "" {
		t.Fatal("expected promql query")
	}
	if MetricToPromQL("throughput", "api-gateway") == "" {
		t.Fatal("expected throughput query")
	}
}

func TestNewPromQLClientNil(t *testing.T) {
	if NewPromQLClient("") != nil {
		t.Fatal("empty url should return nil client")
	}
}

func TestPromQLClientQueryRangeUnavailable(t *testing.T) {
	c := NewPromQLClient("http://127.0.0.1:1")
	_, err := c.QueryRange(t.Context(), "up", time.Now().Add(-time.Hour), time.Now(), time.Minute)
	if err == nil {
		t.Fatal("expected error for unreachable prometheus")
	}
}
