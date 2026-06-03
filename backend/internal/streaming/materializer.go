package streaming

import (
	"context"
	"encoding/json"
	"time"

	"github.com/neuralops/platform/internal/kafka"
	"go.uber.org/zap"
)

// Materializer consumes observability stream events and flushes aggregates.
type Materializer struct {
	agg      *Aggregator
	store    *Store
	log      *zap.Logger
	flushEvery time.Duration
}

// NewMaterializer creates a stream materializer.
func NewMaterializer(store *Store, log *zap.Logger) *Materializer {
	return &Materializer{
		agg:        NewAggregator(64),
		store:      store,
		log:        log,
		flushEvery: 5 * time.Second,
	}
}

// Handler returns a Kafka handler func.
func (m *Materializer) Handler() kafka.HandlerFunc {
	return func(ctx context.Context, msg kafka.Message) error {
		ev, err := ParseEvent(msg.Value)
		if err != nil {
			return err
		}
		tenant := ev.TenantID
		if tenant == "" {
			tenant = "default"
		}
		switch ev.Type {
		case EventMetricSample:
			var p MetricSamplePayload
			if err := json.Unmarshal(ev.Payload, &p); err != nil {
				return err
			}
			ts := p.Timestamp
			if ts.IsZero() {
				ts = time.Now().UTC()
			}
			m.agg.RecordMetric(tenant, p.MetricID, p.Service, p.Value, ts)
		case EventAlertSignal:
			var p AlertSignalPayload
			if err := json.Unmarshal(ev.Payload, &p); err != nil {
				return err
			}
			ts := p.Timestamp
			if ts.IsZero() {
				ts = time.Now().UTC()
			}
			count := p.Count
			if count <= 0 {
				count = 1
			}
			m.agg.RecordAlertSignal(tenant, p.PolicyID, p.Service, count, ts)
		}
		return nil
	}
}

// RunFlushLoop periodically drains aggregates to Postgres.
func (m *Materializer) RunFlushLoop(ctx context.Context) {
	ticker := time.NewTicker(m.flushEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			m.flush(context.Background())
			return
		case <-ticker.C:
			m.flush(ctx)
		}
	}
}

func (m *Materializer) flush(ctx context.Context) {
	metrics, alerts := m.agg.Drain()
	if err := m.store.FlushMetricBuckets(ctx, metrics); err != nil && m.log != nil {
		m.log.Warn("flush metric buckets", zap.Error(err))
	}
	if err := m.store.FlushAlertBuckets(ctx, alerts); err != nil && m.log != nil {
		m.log.Warn("flush alert buckets", zap.Error(err))
	}
}

// PublishEvent serializes a stream event for Kafka producers.
func PublishEvent(ev StreamEvent) ([]byte, error) {
	return json.Marshal(ev)
}
