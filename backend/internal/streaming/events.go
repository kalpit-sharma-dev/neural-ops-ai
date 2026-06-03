package streaming

import (
	"encoding/json"
	"time"
)

const (
	EventMetricSample = "metric_sample"
	EventAlertSignal  = "alert_signal"
)

// StreamEvent is a Kafka message on neuralops.observability stream topics.
type StreamEvent struct {
	Type     string          `json:"type"`
	TenantID string          `json:"tenantId"`
	Payload  json.RawMessage `json:"payload"`
}

// MetricSamplePayload is one raw metric observation for derived materialization.
type MetricSamplePayload struct {
	MetricID  string    `json:"metricId"`
	Service   string    `json:"service"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// AlertSignalPayload is one alert evaluation signal for fatigue scoring.
type AlertSignalPayload struct {
	PolicyID  string    `json:"policyId"`
	Service   string    `json:"service"`
	Severity  string    `json:"severity"`
	Count     int64     `json:"count"`
	Timestamp time.Time `json:"timestamp"`
}

// Topics used by the materializer.
const (
	TopicObservabilityStream = "neuralops.observability.stream"
)

func ParseEvent(raw []byte) (StreamEvent, error) {
	var ev StreamEvent
	err := json.Unmarshal(raw, &ev)
	return ev, err
}
