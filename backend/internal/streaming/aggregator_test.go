package streaming

import (
	"testing"
	"time"
)

func TestAggregatorDrain(t *testing.T) {
	a := NewAggregator(4)
	now := time.Now().UTC()
	a.RecordMetric("default", "dm-1", "payment", 10, now)
	a.RecordMetric("default", "dm-1", "payment", 20, now)
	a.RecordAlertSignal("default", "ap-1", "payment", 5, now)
	metrics, alerts := a.Drain()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric bucket, got %d", len(metrics))
	}
	if metrics[0].Value != 15 {
		t.Fatalf("expected avg 15, got %v", metrics[0].Value)
	}
	if len(alerts) != 1 || alerts[0].SignalCount != 5 {
		t.Fatalf("unexpected alerts: %+v", alerts)
	}
}
