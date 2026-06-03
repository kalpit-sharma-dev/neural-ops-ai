# 7-Day Staging Soak (Wave 1.9)

**ID:** NFR-01

## Goal

Prove staging runs **7 consecutive days** without manual pod restarts, data loss, or SLO breach — prerequisite for Gate B.

## Prerequisites

- Wave 1.1–1.5 complete (Helm staging, SSO, backups configured)
- Alerting routed to internal channel (not customer P1)
- `demoMode: false`, `authAllowDevLogin: false`

## Monitoring during soak

| Signal | Tool | Alert threshold |
|--------|------|-----------------|
| Gateway availability | Prometheus `up{job="gateway"}` | &lt; 99.9% rolling 24h |
| Error rate | `http_requests_total{status=~"5.."}` | &gt; 0.1% 15m |
| Kafka lag | Consumer lag metric | &gt; 60s p95 |
| Pod restarts | `kube_pod_container_status_restarts_total` | Any CrashLoop |
| Disk | PVC / managed store | &gt; 80% |

Import dashboards from `infra/grafana/dashboards/`.

## Daily script

```bash
chmod +x scripts/staging-soak-check.sh
GATEWAY=https://api.staging.neuralops.example \
PROMETHEUS=http://prometheus.monitoring:9090 \
./scripts/staging-soak-check.sh --day 3
```

Append output to soak log file in ticket `SOAK-YYYYMMDD`.

## Synthetic load (optional, low rate)

```bash
# 1 RPS background — do not exceed staging capacity plan
k6 run --vus 2 --duration 24h scripts/k6/bank-poc-load.js
```

Prefer scheduled CronJob in staging namespace.

## Failure handling

| Event | Action |
|-------|--------|
| Single pod restart (OOM) | Log, fix limit, **restart soak clock** |
| Planned Helm upgrade | **Pause** soak; resume after 24h clean |
| Data store failover test | Pre-approved; does not reset clock if RTO &lt; 5 min |

## Exit criteria

- [ ] 7× daily `staging-soak-check.sh` all green
- [ ] Zero unplanned manual `kubectl delete pod` during window
- [ ] One backup restore drill passed during or immediately after soak
- [ ] Post-soak review notes attached to Gate B checklist

## Gate B linkage

Update [BANK_PRODUCTION_READINESS_ROADMAP.md](../BANK_PRODUCTION_READINESS_ROADMAP.md) § Gate B when complete.
