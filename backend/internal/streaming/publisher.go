package streaming

import (
	"context"
	"encoding/json"
	"time"

	"github.com/neuralops/platform/internal/kafka"
)

// Publisher emits observability stream events to Kafka.
type Publisher struct {
	producer *kafka.Producer
}

// NewPublisher wraps a Kafka producer.
func NewPublisher(p *kafka.Producer) *Publisher {
	return &Publisher{producer: p}
}

func (p *Publisher) publish(ctx context.Context, ev StreamEvent) error {
	if p == nil || p.producer == nil {
		return nil
	}
	raw, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	return p.producer.Publish(ctx, kafka.Message{
		Topic: TopicObservabilityStream,
		Key:   ev.TenantID,
		Value: raw,
	})
}

// PublishMetricSample enqueues a derived metric sample.
func (p *Publisher) PublishMetricSample(ctx context.Context, tenantID string, payload MetricSamplePayload) error {
	if payload.Timestamp.IsZero() {
		payload.Timestamp = time.Now().UTC()
	}
	body, _ := json.Marshal(payload)
	return p.publish(ctx, StreamEvent{Type: EventMetricSample, TenantID: tenantID, Payload: body})
}

// PublishAlertSignal enqueues an alert scoring signal.
func (p *Publisher) PublishAlertSignal(ctx context.Context, tenantID string, payload AlertSignalPayload) error {
	if payload.Timestamp.IsZero() {
		payload.Timestamp = time.Now().UTC()
	}
	body, _ := json.Marshal(payload)
	return p.publish(ctx, StreamEvent{Type: EventAlertSignal, TenantID: tenantID, Payload: body})
}
