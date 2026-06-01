# NeuralOps Gateway API — Observability UI

**Base path:** `/api/v1`  
**Auth:** Bearer JWT or API key; optional `X-Tenant-ID` header.

All success responses:

```json
{
  "status": "success",
  "data": {},
  "timestamp": "2026-05-31T12:00:00Z"
}
```

---

## APM / Tracing

| Method | Path | Description |
|--------|------|-------------|
| GET | `/apm/traces/:traceId` | Full trace with spans |
| POST | `/apm/traces/search` | Search traces (body: `service`, `status`, `limit`, …) |
| GET | `/apm/flow` | Service flow aggregation edges |
| GET | `/apm/services/:service/operations` | Distinct operations for a service |
| GET | `/apm/retention` | Trace retention + sampling policy |
| PUT | `/apm/retention` | Update retention policy |
| GET | `/apm/services/:service/profiles` | Continuous profiling hotspots |
| POST | `/apm/profiles` | Ingest profile sample (profiler agent) |

---

## Cloud & depth

| Method | Path | Description |
|--------|------|-------------|
| GET | `/cloud/dashboards?provider=` | Cloud dashboard catalog |
| GET | `/cloud/metrics?provider=&metric=&region=` | Federated cloud metric series |
| GET | `/infra/k8s/namespaces` | K8s namespace inventory |
| GET | `/infra/k8s/deployments?namespace=` | K8s deployment inventory |
| POST | `/notebooks/:id/execute` | Execute all notebook cells |
| POST | `/workflows/trigger` | Trigger workflows by event |
| GET | `/workflows/:id/runs` | Workflow run history |
| PUT | `/slos/:id/burn-alert` | Enable SLO burn alerting |
| POST | `/rum/consent` | Record RUM GDPR consent |
| POST | `/mobile/push/register` | Register Expo push token |

---

## Metrics & Dashboards

| Method | Path | Description |
|--------|------|-------------|
| GET | `/metrics/catalog` | Available metric definitions |
| GET | `/metrics/query?name=&service=` | Time series (ClickHouse → **Prometheus fallback** → demo) |
| GET | `/metrics/promql?query=` | Direct PromQL range proxy |
| GET | `/dashboards` | List dashboards |
| POST | `/dashboards` | Create dashboard |
| GET | `/dashboards/:id` | Get dashboard |
| PUT | `/dashboards/:id` | Update dashboard |
| DELETE | `/dashboards/:id` | Delete dashboard |

---

## Topology

| Method | Path | Description |
|--------|------|-------------|
| GET | `/topology?zone=` | Live service graph (nodes + edges) |
| GET | `/zones` | Management zones |

---

## Logs configuration

| Method | Path | Description |
|--------|------|-------------|
| GET | `/logs/metric-rules` | Log-to-metric rules |
| POST | `/logs/metric-rules` | Create log metric rule |
| GET | `/logs/parsing-rules` | Log parsing rules |
| POST | `/logs/parsing-rules` | Create parsing rule |

---

## SLOs & Anomalies

| Method | Path | Description |
|--------|------|-------------|
| GET | `/slos` | List SLOs |
| POST | `/slos` | Create SLO |
| GET | `/anomalies/entities` | Entity-scoped anomalies |

---

## Infrastructure

| Method | Path | Description |
|--------|------|-------------|
| GET | `/infra/hosts` | Host inventory |
| GET | `/infra/k8s/clusters` | Kubernetes clusters |
| GET | `/infra/k8s/pods?namespace=` | Pod list |

---

## Database & Middleware

| Method | Path | Description |
|--------|------|-------------|
| GET | `/databases` | Database instances |
| GET | `/databases/:id/statements` | Top SQL statements |
| GET | `/middleware/kafka/lag` | Consumer lag |

---

## RUM & Synthetic

| Method | Path | Description |
|--------|------|-------------|
| GET | `/rum/sessions` | RUM session summaries |
| GET | `/rum/sessions/:sessionId/replay` | Session replay events |
| POST | `/rum/beacon` | RUM SDK session beacon |
| POST | `/rum/replay` | RUM SDK replay frame |
| GET | `/synthetic/monitors` | Synthetic monitors |
| POST | `/synthetic/monitors` | Create monitor |
| GET | `/synthetic/monitors/:id/runs` | Monitor run history |

---

## Workflows & Notebooks

| Method | Path | Description |
|--------|------|-------------|
| GET | `/workflows` | Automation workflows |
| GET | `/notebooks` | Saved notebooks |

---

## Security & Integrations

| Method | Path | Description |
|--------|------|-------------|
| GET | `/security/vulnerabilities` | Runtime vulnerabilities |
| GET | `/security/attacks` | Attack events |
| GET | `/integrations` | Integration catalog (Postgres-backed) |
| POST | `/integrations/:id/connect` | Connect integration (`body: { config: { ... } }`) |
| POST | `/integrations/:id/disconnect` | Disconnect integration |
| GET | `/integrations/:id/oauth/start` | Begin OAuth (Jira/Slack) |
| GET | `/integrations/oauth/callback` | OAuth callback |

---

## Admin

| Method | Path | Description |
|--------|------|-------------|
| GET | `/admin/users` | Tenant users |
| GET | `/admin/api-keys` | API keys (no secrets) |
| POST | `/admin/api-keys` | Create API key; returns `{ key, secret }` once |
| GET | `/admin/audit` | Audit log entries |
| GET | `/admin/usage` | Tenant usage stats |
| GET/PUT | `/admin/sso` | SSO / IdP configuration |
| GET/PUT | `/admin/tenant-policies` | Log retention + ingestion limits |
| GET/POST | `/admin/oncall` | List / create on-call schedules |
| PUT | `/admin/oncall/:id` | Update on-call schedule |
| DELETE | `/admin/oncall/:id` | Delete on-call schedule |
| POST | `/admin/oncall/:id/sync-pagerduty` | Import rotation from PagerDuty |

---

## Alerting (existing service, proxied via gateway)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/alerts` | List alerts |
| POST | `/alerts/:id/acknowledge` | Acknowledge alert |
| POST | `/alerts/:id/suppress` | Suppress alert |
| GET/POST/PUT/DELETE | `/alerts/rules` | Alert rules CRUD |
| GET/POST/PUT/DELETE | `/notifications/channels` | Notification channels |
| POST | `/notifications/channels/:id/test` | Test channel |
| GET/POST | `/alerts/silences` | Alert silences |

---

## Collectors (background service)

The `collector` binary syncs infrastructure telemetry:

- **Hosts** — node_exporter metrics via Prometheus
- **Kubernetes** — kube-state-metrics via Prometheus
- **Synthetic** — HTTP monitor execution

Configure with `PROMETHEUS_URL`, `POSTGRES_DSN`, `CLICKHOUSE_DSN`, `COLLECTOR_INTERVAL`.

The `profiler` binary emits stack samples to `POST /apm/profiles`.

---

*See also:* [UI_DYNATRACE_IMPLEMENTATION_PLAN.md](./UI_DYNATRACE_IMPLEMENTATION_PLAN.md), [UI_DYNATRACE_GAP.md](./UI_DYNATRACE_GAP.md), [remaining.md](./remaining.md)
