# Network Performance Monitoring (NPM) Suite (Wave 6.2)

**IDs:** NPM-01..04

## Capabilities

| Feature | API | Status |
|---------|-----|--------|
| Topology graph | `GET /api/v1/network/topology` | ✅ Seed + UI |
| Flow anomalies | `GET /api/v1/network/anomalies` | ✅ |
| eBPF flows | NexAgent pilot | 🟡 Linux hosts only |
| SD-WAN tunnels | `GET /api/v1/network/npm/sdwan` | ✅ API |
| Wireless links | `GET /api/v1/network/npm/wireless` | ✅ API |
| NetFlow | Ingest via Fluent Bit `flow` input | Customer config |

## Bank POC scope

1. Map **payments path**: `fw-edge → sw-core → pay-subnet → ledger`.
2. Correlate NPM anomaly with APM trace latency on same window.
3. Optional: deploy NexAgent on 2–3 payment subnet hosts (eBPF).

## eBPF pilot checklist

- [ ] Kernel 5.10+ with BTF
- [ ] `scripts/verify-ebpf-fleet.sh` green in CI matrix
- [ ] Read-only CAP_BPF / privileged DaemonSet per bank policy

## Sales language

Full Datadog NPM parity not claimed; position as **path analysis + eBPF pilot** with roadmap for NetFlow/SD-WAN depth.
