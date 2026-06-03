package collector

import (
	"context"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	gnet "github.com/shirou/gopsutil/v4/net"
)

// HostCollector gathers cross-platform host vitals (CPU, memory, network, load)
// using gopsutil. It is the always-on baseline collector on every OS.
type HostCollector struct {
	hostname string
	// cpuSampleWindow is how long Collect blocks to compute a CPU busy delta.
	cpuSampleWindow time.Duration
}

// NewHostCollector builds a host collector tagged with the given hostname.
func NewHostCollector(hostname string) *HostCollector {
	return &HostCollector{hostname: hostname, cpuSampleWindow: 200 * time.Millisecond}
}

// Name implements Collector.
func (c *HostCollector) Name() string { return "host" }

// Collect implements Collector, returning host CPU/memory/network/load gauges.
func (c *HostCollector) Collect(ctx context.Context) ([]Sample, error) {
	now := time.Now().UTC()
	base := map[string]string{"host": c.hostname, "os": runtime.GOOS}
	out := make([]Sample, 0, 8)

	if pct, err := cpu.PercentWithContext(ctx, c.cpuSampleWindow, false); err == nil && len(pct) > 0 {
		out = append(out, Sample{
			Name: "host.cpu.utilization", Value: round2(pct[0]), Unit: "percent",
			Kind: "gauge", Attributes: base, Timestamp: now,
		})
	}

	if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil && vm != nil {
		out = append(out,
			Sample{Name: "host.memory.utilization", Value: round2(vm.UsedPercent), Unit: "percent", Kind: "gauge", Attributes: base, Timestamp: now},
			Sample{Name: "host.memory.used", Value: float64(vm.Used), Unit: "By", Kind: "gauge", Attributes: base, Timestamp: now},
			Sample{Name: "host.memory.total", Value: float64(vm.Total), Unit: "By", Kind: "gauge", Attributes: base, Timestamp: now},
		)
	}

	if io, err := gnet.IOCountersWithContext(ctx, false); err == nil && len(io) > 0 {
		out = append(out,
			Sample{Name: "host.network.io.tx", Value: float64(io[0].BytesSent), Unit: "By", Kind: "counter", Attributes: base, Timestamp: now},
			Sample{Name: "host.network.io.rx", Value: float64(io[0].BytesRecv), Unit: "By", Kind: "counter", Attributes: base, Timestamp: now},
		)
	}

	// Load average is meaningful on Unix-like systems; gopsutil returns an error
	// on platforms without it, so we surface it best-effort.
	if avg, err := load.AvgWithContext(ctx); err == nil && avg != nil {
		out = append(out, Sample{
			Name: "host.cpu.load1", Value: round2(avg.Load1), Unit: "1",
			Kind: "gauge", Attributes: base, Timestamp: now,
		})
	}

	return out, nil
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
