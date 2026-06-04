package validator

import (
	"testing"
	"time"

	"github.com/neuralops/platform/internal/ingestion/dto"
)

func TestValidateMetricLabelCardinality(t *testing.T) {
	req := &dto.MetricIngestRequest{
		ServiceName: "pay", MetricType: "CPU", Value: 1,
		Timestamp: time.Now().UTC(),
		Labels:    map[string]string{"env": "prod"},
	}
	if err := ValidateMetric(req); err != nil {
		t.Fatal(err)
	}
	labels := make(map[string]string, 40)
	for i := 0; i < 40; i++ {
		labels["k"+string(rune('a'+i))] = "v"
	}
	req.Labels = labels
	if err := ValidateMetric(req); err == nil {
		t.Fatal("expected cardinality error")
	}
}
