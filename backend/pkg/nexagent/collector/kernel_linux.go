//go:build linux

package collector

import (
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"
	"time"
)

// KernelCollector reads low-overhead kernel networking counters directly from
// procfs (/proc/net/snmp, /proc/net/netstat). These expose TCP health signals
// (retransmits, resets, listen drops) that are otherwise invisible to userspace
// APM. This is the procfs baseline; the eBPF variant (see pkg/nexagent/ebpf)
// can replace it for per-flow RTT and syscall-level latency when compiled with
// a clang/libbpf toolchain.
type KernelCollector struct {
	hostname string
}

// NewKernelCollector builds the Linux procfs kernel collector.
func NewKernelCollector(hostname string) *KernelCollector {
	return &KernelCollector{hostname: hostname}
}

// Name implements Collector.
func (c *KernelCollector) Name() string { return "kernel" }

// Collect implements Collector, emitting selected TCP/IP kernel counters.
func (c *KernelCollector) Collect(ctx context.Context) ([]Sample, error) {
	now := time.Now().UTC()
	base := map[string]string{"host": c.hostname, "os": "linux", "source": "procfs"}
	out := make([]Sample, 0, 8)

	wanted := map[string]struct{}{
		"RetransSegs": {}, "InSegs": {}, "OutSegs": {}, "ActiveOpens": {},
		"PassiveOpens": {}, "EstabResets": {}, "AttemptFails": {},
	}
	if stats, err := parseProcNet("/proc/net/snmp", "Tcp", wanted); err == nil {
		for name, val := range stats {
			out = append(out, Sample{
				Name: "kernel.tcp." + toSnake(name), Value: float64(val), Unit: "1",
				Kind: "counter", Attributes: base, Timestamp: now,
			})
		}
	}

	wantedExt := map[string]struct{}{
		"ListenDrops": {}, "ListenOverflows": {}, "TCPLostRetransmit": {}, "TCPSynRetrans": {},
	}
	if stats, err := parseProcNet("/proc/net/netstat", "TcpExt", wantedExt); err == nil {
		for name, val := range stats {
			out = append(out, Sample{
				Name: "kernel.tcp." + toSnake(name), Value: float64(val), Unit: "1",
				Kind: "counter", Attributes: base, Timestamp: now,
			})
		}
	}

	return out, nil
}

// parseProcNet parses the two-line "Header: keys" / "Header: values" layout used
// by /proc/net/snmp and /proc/net/netstat, returning only the requested keys.
func parseProcNet(path, prefix string, wanted map[string]struct{}) (map[string]int64, error) {
	f, err := os.Open(path) //nolint:gosec // fixed procfs path
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]int64)
	scanner := bufio.NewScanner(f)
	var headers []string
	tag := prefix + ":"
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, tag) {
			continue
		}
		fields := strings.Fields(line)
		if len(headers) == 0 {
			headers = fields
			continue
		}
		for i := 1; i < len(fields) && i < len(headers); i++ {
			key := headers[i]
			if _, ok := wanted[key]; !ok {
				continue
			}
			if v, perr := strconv.ParseInt(fields[i], 10, 64); perr == nil {
				result[key] = v
			}
		}
		headers = nil // allow multiple prefixed blocks (netstat repeats)
	}
	return result, scanner.Err()
}

func toSnake(camel string) string {
	var b strings.Builder
	for i, r := range camel {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
