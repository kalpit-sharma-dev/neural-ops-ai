//go:build !linux

package collector

import "context"

// KernelCollector is a no-op outside Linux: kernel TCP counters via procfs and
// eBPF are Linux-only. The host collector still provides cross-platform vitals.
type KernelCollector struct {
	hostname string
}

// NewKernelCollector builds a no-op kernel collector on non-Linux platforms.
func NewKernelCollector(hostname string) *KernelCollector {
	return &KernelCollector{hostname: hostname}
}

// Name implements Collector.
func (c *KernelCollector) Name() string { return "kernel" }

// Collect returns no samples on non-Linux platforms.
func (c *KernelCollector) Collect(_ context.Context) ([]Sample, error) {
	return nil, nil
}
