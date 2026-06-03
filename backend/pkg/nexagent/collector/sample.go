// Package collector defines the NEXAGENT telemetry collection contract and
// the sample model shared by host, kernel, and eBPF collectors.
package collector

import (
	"context"
	"time"
)

// Sample is a single normalized telemetry measurement emitted by a collector.
// It maps cleanly onto an OTLP metric data point (name + value + attributes).
type Sample struct {
	Name       string            `json:"name"`
	Value      float64           `json:"value"`
	Unit       string            `json:"unit,omitempty"`
	Kind       string            `json:"kind"` // gauge|counter
	Attributes map[string]string `json:"attributes,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
}

// Collector produces a batch of samples on each scrape. Implementations must be
// safe for repeated invocation and must not block longer than the scrape budget.
type Collector interface {
	// Name identifies the collector for diagnostics and per-source metrics.
	Name() string
	// Collect gathers a point-in-time batch. It must honor ctx cancellation.
	Collect(ctx context.Context) ([]Sample, error)
}
