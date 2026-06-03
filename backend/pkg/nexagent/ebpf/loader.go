//go:build linux && ebpf

// Package ebpf provides the optional eBPF-backed kernel collector for NEXAGENT.
//
// It is gated behind the `ebpf` build tag because it requires a clang/libbpf
// toolchain and bpf2go-generated bindings. Build it with:
//
//	go get github.com/cilium/ebpf@latest
//	make -C pkg/nexagent/ebpf        # runs bpf2go (go:generate) below
//	go build -tags ebpf ./...
//
// When compiled, EBPFCollector satisfies collector.Collector and can be swapped
// in for the procfs KernelCollector in cmd/nexagent/main.go.
package ebpf

import (
	"context"
	"fmt"
	"math"
	"net"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"github.com/neuralops/platform/pkg/nexagent/collector"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -target bpfel bpf bpf/tcp_rtt.bpf.c -- -I./bpf

// EBPFCollector loads the tcp_rtt CO-RE program and reads the per-flow RTT
// histogram, emitting p50/p95-style latency gauges per connection pair.
type EBPFCollector struct {
	hostname string
	objs     bpfObjects
	link     link.Link
}

// NewEBPFCollector loads and attaches the eBPF program. The caller must call
// Close to detach and release kernel resources.
func NewEBPFCollector(hostname string) (*EBPFCollector, error) {
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("remove memlock rlimit: %w", err)
	}
	var objs bpfObjects
	if err := loadBpfObjects(&objs, nil); err != nil {
		return nil, fmt.Errorf("load bpf objects: %w", err)
	}
	kp, err := link.Kprobe("tcp_rcv_established", objs.HandleTcpRcv, nil)
	if err != nil {
		objs.Close()
		return nil, fmt.Errorf("attach kprobe: %w", err)
	}
	return &EBPFCollector{hostname: hostname, objs: objs, link: kp}, nil
}

// Name implements collector.Collector.
func (c *EBPFCollector) Name() string { return "ebpf-tcp-rtt" }

// Collect reads the RTT histogram map and emits per-flow latency percentiles.
func (c *EBPFCollector) Collect(_ context.Context) ([]collector.Sample, error) {
	now := time.Now().UTC()
	out := make([]collector.Sample, 0, 16)

	var key bpfHistKey
	var hist bpfHist
	iter := c.objs.RttHist.Iterate()
	for iter.Next(&key, &hist) {
		p95us := percentileFromLog2Hist(hist.Slots[:], 0.95)
		attrs := map[string]string{
			"host":   c.hostname,
			"source": "ebpf",
			"saddr":  intToIP(key.Saddr),
			"daddr":  intToIP(key.Daddr),
		}
		out = append(out, collector.Sample{
			Name: "kernel.tcp.rtt.p95", Value: p95us / 1000.0, Unit: "ms",
			Kind: "gauge", Attributes: attrs, Timestamp: now,
		})
	}
	return out, iter.Err()
}

// Close detaches the program and frees maps.
func (c *EBPFCollector) Close() error {
	if c.link != nil {
		_ = c.link.Close()
	}
	return c.objs.Close()
}

// percentileFromLog2Hist reconstructs an approximate value at the given quantile
// from a power-of-two bucketed histogram.
func percentileFromLog2Hist(slots []uint64, q float64) float64 {
	var total uint64
	for _, s := range slots {
		total += s
	}
	if total == 0 {
		return 0
	}
	target := uint64(math.Ceil(float64(total) * q))
	var cum uint64
	for i, s := range slots {
		cum += s
		if cum >= target {
			return math.Pow(2, float64(i))
		}
	}
	return math.Pow(2, float64(len(slots)-1))
}

func intToIP(addr uint32) string {
	ip := make(net.IP, 4)
	ip[0] = byte(addr)
	ip[1] = byte(addr >> 8)
	ip[2] = byte(addr >> 16)
	ip[3] = byte(addr >> 24)
	return ip.String()
}
