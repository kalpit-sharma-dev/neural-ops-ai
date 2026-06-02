# NeuralOps SRS Parity — Phase-by-Phase Execution Backlog

Date: 2026-06-02  
Companion doc: `docs/SRS_NEURALOPS_COMPARISON_GAP.md`

This backlog converts the SRS gap analysis into executable implementation phases.

---

## How to use this file

- Treat each phase as a delivery milestone (2-4 weeks each).
- For every phase, complete all tracks:
  - API/OpenAPI
  - Backend services/domain logic
  - Data model/migrations
  - Frontend UX/pages/components
  - Test strategy and acceptance gates
- Keep `docs/openapi/gateway-v1.yaml` as canonical API contract and sync via `./scripts/sync-openapi.sh`.

---

## Phase 1 — Unified Query + Alerting Foundation (P0)

SRS targets:
- `REQ-MET-006`, `REQ-MET-007`
- `REQ-ALERT-003`, `REQ-ALERT-004`, `REQ-ALERT-005`, `REQ-ALERT-006`, `REQ-ALERT-007`

### API / OpenAPI
- Add unified query endpoints:
  - `POST /query/unified`
  - `POST /query/validate`
  - `GET /query/functions`
- Alert policy endpoints:
  - `GET/POST /alerts/policies`
  - `PUT/DELETE /alerts/policies/{id}`
  - `GET/POST /alerts/suppressions`
  - `DELETE /alerts/suppressions/{id}`
- Add alert context payload schemas (runbook, owner, topology links, trace/log pivots).

### Backend
- Build query planner that joins logs/metrics/traces/events.
- Add derived-metric expression evaluator.
- Implement alert policy engine:
  - routing trees, escalation timers, suppression windows, dedupe keys.
- Add correlation scoring improvements and policy explainability.

### DB / Migrations
- `alert_policies`, `alert_policy_routes`, `alert_suppressions`, `derived_metrics`.
- Indexes on tenant + status + route target + schedule windows.

### Frontend
- New pages:
  - Unified Query Workbench
  - Alert Policies
  - Suppression Rules
- Upgrade Alerts page with context cards and policy traceability.

### Tests / Exit criteria
- Unit: query parser/planner, alert routing logic.
- Integration: multi-condition alert + suppression + escalation.
- E2E: create policy -> trigger -> routed notification.
- Exit gate:
  - cross-signal query returns deterministic results under load,
  - policy engine supports AND/OR/NOT and route fallback.

---

## Phase 2 — Agent Fleet + Collector Platform (P0/P1)

SRS targets:
- `REQ-COLL-001`, `REQ-COLL-003`, `REQ-COLL-004`, `REQ-COLL-009`, `REQ-COLL-016`, `REQ-COLL-017`

### API / OpenAPI
- Fleet management endpoints:
  - `GET/POST /collectors/fleet`
  - `GET/PUT /collectors/fleet/{id}`
  - `POST /collectors/fleet/{id}/upgrade`
- Pipeline endpoints:
  - `GET/POST /collectors/pipelines`
  - `PUT /collectors/pipelines/{id}`
  - `POST /collectors/pipelines/{id}/validate`

### Backend
- Agent registry service (heartbeats, versions, policy assignment).
- Pipeline runtime for parse/enrich/drop/route/mask.
- Operator controller reconciler for K8s deployment and drift correction.

### DB / Migrations
- `collector_fleet`, `collector_heartbeats`, `collector_policies`, `collector_pipelines`.
- Versioned policy snapshots.

### Frontend
- Fleet console:
  - inventory, status, version drift, rollout actions.
- Pipeline builder:
  - visual stages + test-sample preview + validation.

### Tests / Exit criteria
- Integration: policy rollout to subset and rollback.
- Load: heartbeat scale + pipeline throughput benchmark.
- Exit gate:
  - controlled canary rollout,
  - fleet health dashboard and alerting functional.

---

## Phase 3 — Security Deepening (AppSec + SecObs) (P0/P1)

SRS targets:
- `REQ-SEC-001`, `REQ-SEC-002`, `REQ-SEC-003`, `REQ-SEC-004`, `REQ-SEC-005`, `REQ-SEC-007`

### API / OpenAPI
- New endpoints:
  - `GET /security/findings`
  - `GET /security/findings/{id}`
  - `POST /security/sca/import`
  - `POST /security/images/import`
  - `GET /security/cspm/posture`
  - `POST /security/siem/exports`

### Backend
- Findings domain model with severity, exploitability, asset, status, SLA.
- Ingestion adapters for SCA/image/CSPM tools.
- SIEM formatter/export worker (Splunk/Sentinel/Elastic schema adapters).

### DB / Migrations
- `security_findings`, `security_assets`, `security_posture_checks`, `security_exports`.
- Correlation tables to link finding <-> trace/log/incident.

### Frontend
- Security findings explorer and posture dashboard.
- Incident linkage panel from findings.
- SIEM export status UI.

### Tests / Exit criteria
- Contract tests for import/export schemas.
- Integration: finding correlation to incident creation.
- Exit gate:
  - findings triage workflow complete,
  - SIEM export reliably streams normalized signals.

---

## Phase 4 — AI Explainability + Forecasting + AutoFix Guardrails (P0/P1)

SRS targets:
- `REQ-AI-004`, `REQ-AI-005`, `REQ-AI-006`, `REQ-AI-007`, `REQ-AI-008`, `REQ-AI-017`, `REQ-AI-019`, `REQ-AI-020`

### API / OpenAPI
- Add:
  - `GET /ai/rca/{incidentId}`
  - `GET /ai/explanations/{id}`
  - `POST /ai/forecast`
  - `POST /ai/autofix/plan`
  - `POST /ai/autofix/execute`
  - `POST /ai/autofix/rollback`

### Backend
- Causal graph service with confidence and evidence provenance.
- Forecasting service (capacity, SLO burn, incident probability).
- AutoFix policy engine:
  - approval workflow,
  - blast radius policy,
  - action timeout + rollback hooks.

### DB / Migrations
- `ai_explanations`, `ai_forecasts`, `autofix_actions`, `autofix_approvals`, `autofix_rollbacks`.

### Frontend
- AI RCA panel with “why” tree and confidence bars.
- Forecast dashboards and capacity recommendations.
- AutoFix center (plan/approve/execute/rollback timeline).

### Tests / Exit criteria
- Safety tests for action policy constraints.
- Replay tests with historical incidents and expected RCA quality.
- Exit gate:
  - every AI recommendation has evidence + confidence,
  - no autonomous remediation without explicit policy pass.

---

## Phase 5 — Infra, Cloud Asset Graph, FinOps, NPM (P1/P2)

SRS targets:
- `REQ-INF-010`, `REQ-INF-011`, `REQ-INF-013`, `REQ-INF-014`
- `REQ-NPM-001` to `REQ-NPM-006`

### API / OpenAPI
- Asset graph:
  - `GET /cloud/assets`
  - `GET /cloud/topology`
- FinOps:
  - `GET /finops/costs`
  - `GET /finops/anomalies`
  - `GET /finops/carbon`
- NPM:
  - `GET /network/flows`
  - `GET /network/devices`
  - `GET /network/topology`
  - `GET /network/anomalies`

### Backend
- Cloud connectors (AWS/GCP/Azure inventory and relationships).
- Cost ingestion and anomaly detection.
- Network collectors (flows + SNMP) and topology processor.

### DB / Migrations
- `cloud_assets`, `cloud_relationships`, `finops_usage`, `finops_budgets`, `network_flows`, `network_devices`.

### Frontend
- Multi-cloud topology explorer.
- FinOps workspace (cost trends, budgets, anomaly alerts, carbon).
- NPM dashboards (latency/loss/jitter/hot links).

### Tests / Exit criteria
- Integration with mock cloud account + asset sync correctness.
- Load tests on flow ingestion and topology rendering.
- Exit gate:
  - usable cloud asset inventory and actionable cost anomaly alerts.

---

## Phase 6 — RUM/Synthetic/Business Observability Expansion (P1/P2)

SRS targets:
- `REQ-RUM-003`, `REQ-RUM-005`
- `REQ-SYN-002`, `REQ-SYN-003`, `REQ-SYN-004`, `REQ-SYN-005`
- Business observability requirements (section 14)

### API / OpenAPI
- RUM funnels:
  - `GET /rum/funnels`
  - `POST /rum/funnels`
- Synthetic expansion:
  - `POST /synthetic/browser-tests`
  - `POST /synthetic/mobile-tests`
  - `POST /synthetic/private-locations`
- Business KPI packs:
  - `GET /business/kpi-packs`
  - `POST /business/kpi-packs/{id}/enable`

### Backend
- Funnel computation engine and conversion attribution.
- Browser/mobile synthetic runner orchestration.
- Private location agent registration.
- KPI pack templates and connectors.

### DB / Migrations
- `rum_funnels`, `synthetic_browser_tests`, `synthetic_mobile_tests`, `synthetic_locations`, `business_kpi_packs`.

### Frontend
- Funnel analytics UI.
- Synthetic script editor and private location management.
- Business KPI pack library and setup wizard.

### Tests / Exit criteria
- E2E synthetic script execution from CI and private location.
- RUM-to-trace drilldown validation.
- Exit gate:
  - funnels and synthetic browser tests operational with alerts.

---

## Phase 7 — Integrations, CLI, IaC Providers, Enterprise Governance (P1/P2)

SRS targets:
- `REQ-INT-011..014`, `REQ-INT-021`, `REQ-INT-022`
- `REQ-ADM-003`, `REQ-ADM-007`, `REQ-ADM-011`, `REQ-ADM-012`

### API / OpenAPI
- Governance:
  - `GET/PUT /admin/abac-policies`
  - `GET/PUT /admin/data-residency`
  - `GET/PUT /admin/branding`
  - `GET /admin/msp/tenants`
- Export connectors:
  - `POST /exports/warehouse`
  - `POST /exports/bi`
  - `POST /exports/events`

### Backend
- ABAC policy engine integrated with existing RBAC middleware.
- Residency enforcement and data-routing policy checks.
- MSP tenant management control plane.
- Public CLI backend surfaces and signed provider APIs.

### DB / Migrations
- `abac_policies`, `residency_policies`, `msp_tenants`, `branding_themes`, `export_jobs`.

### Frontend
- ABAC policy UI and evaluator.
- Residency configuration and compliance views.
- MSP console and white-label controls.

### External deliverables
- Official `neuralops` CLI v1.
- Terraform provider v1 (plus Pulumi bridge or native provider).

### Tests / Exit criteria
- Security regression tests for RBAC+ABAC evaluation matrix.
- Provider acceptance tests (terraform apply/import/destroy).
- Exit gate:
  - enterprise governance controls and IaC automation production-ready.

---

## Phase 8 — NFR Hardening and Certification Track

SRS targets:
- `REQ-NFR-*` (performance, availability, reliability, accessibility, i18n, operability)

### Workstreams
- Performance certification:
  - ingestion throughput, query latency, alert latency benchmarks.
- Reliability:
  - HA/DR drills, multi-region failover tests.
- Security and compliance:
  - penetration test, secret management hardening, audit evidence packs.
- UX quality:
  - WCAG 2.1 AA verification and i18n framework rollout.

### Exit gate
- Signed benchmark report + SLO dashboards + runbooks + DR evidence.

---

## Cross-phase engineering rules

- API-first: no feature merge without OpenAPI update.
- Migration-safe: additive schema first, then backfills, then cutover.
- Feature flags: all major features behind tenant-aware flags until validated.
- Observability of observability: every new subsystem exports its own metrics/traces/logs.
- Security-by-default: no plaintext secrets; use pluggable secret providers.

---

## Suggested epic naming convention

- `SRS-PH1-QUERY-<topic>`
- `SRS-PH2-FLEET-<topic>`
- `SRS-PH3-SEC-<topic>`
- `SRS-PH4-AI-<topic>`
- `SRS-PH5-INFRA-<topic>`
- `SRS-PH6-RUMSYN-<topic>`
- `SRS-PH7-GOV-<topic>`
- `SRS-PH8-NFR-<topic>`

---

## Immediate next step (start tomorrow)

Start Phase 1 with three concrete epics:
1. Unified query API + parser/planner MVP
2. Alert policy routing/suppression domain and APIs
3. Frontend query workbench + alert policy UI

Ship these behind feature flags and validate via targeted smoke/e2e before broad rollout.

