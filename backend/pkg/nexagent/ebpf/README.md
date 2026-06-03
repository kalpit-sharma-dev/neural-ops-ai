# NEXAGENT eBPF collector (nested module)

This directory is a **separate Go module** so the root `backend/go.mod` does not
pull `github.com/cilium/ebpf`. The standard agent (`cmd/nexagent`) uses gopsutil
+ procfs only.

## Layout

| Path | Purpose |
|------|---------|
| `loader.go` | cilium/ebpf RTT collector (`-tags ebpf`) |
| `validate/` | Kernel/BTF/arch matrix (pure Go, no BPF load) |
| `cmd/fleet-validate/` | CLI that exits non-zero when host fails matrix |
| `bpf/tcp_rtt.bpf.c` | CO-RE eBPF program source |

## Build

```bash
cd backend/pkg/nexagent/ebpf
go test ./validate/...          # matrix unit tests (always)
go run ./cmd/fleet-validate   # host probe (Linux + BTF required to pass)

# Full eBPF loader (requires clang, libbpf, bpf2go):
bpftool btf dump file /sys/kernel/btf/vmlinux format c > bpf/vmlinux.h
make generate
go build -tags ebpf .
```

From repo root: `bash scripts/verify-ebpf-fleet.sh`

## Fleet matrix

- OS: Linux only
- Arch: amd64, arm64
- Kernel: ≥ 5.4 (CO-RE)
- BTF: `/sys/kernel/btf/vmlinux` present (`CONFIG_DEBUG_INFO_BTF=y`)

See `validate/matrix.go` for the authoritative rules and CI gates.

## Wiring into the agent

Replace the procfs `KernelCollector` in `cmd/nexagent/main.go` when building a
custom binary that links this module, or run a sidecar that merges eBPF samples
into the same spool envelope.
