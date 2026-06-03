# FinOps Module — Enhancement Requirements (NeuralOps / NEXOBS)

Status: **Complete** (Waves 1–3 implemented, 2026-06-03)
Date: 2026-06-03
Owners: Platform / Cloud Cost Intelligence squad
Traceability: extends SRS `REQ-INF-013` (Cloud Cost Intelligence) and `REQ-INF-014` (Carbon Emissions); see `SRS_NextGen_Observability_Platform.md` §10.5 and §25 (Pricing).

### Implementation status

| Wave | Requirements | Status |
|---|---|---|
| **Wave 1 (P0)** | 001–003, 010–011, 020–021, 030–031, 050–051, 071 | **Complete** — migrations `000023`, connectors, allocation, anomalies+alerting, recommendations, budgets+forecast, ABAC |
| **Wave 2 (P1)** | 004–005, 012–013, 022–023, 032–034, 040–041, 052, 060–061, 070, 072 | **Complete** — migration `000024`, imports, shared splits, ML tags, lifecycle, commitments, unit economics, carbon/SCI, reports, audit |
| **Wave 3 (P2)** | 014, 042, 053, 062, 073 | **Complete** — migration `000025`, chargeback, scenarios, commitment alerts, carbon actions, Terraform provider, governance |

**Notes:** Billing connectors use simulated CUR/BigQuery/Azure fixtures when cloud credentials are absent; Postgres hybrid store falls back to in-memory seed. Scheduled daily ingest runs via gateway (`FINOPS_INGEST_CADENCE`, default 24h) with stale-ingest alerting. List endpoints support optional `page`/`size`/`sort` pagination. k6 load script: `scripts/k6/finops-breakdown.js`.

---

## 1. Purpose & Scope

The current FinOps capability is a thin, read-only slice: cost-trend series, static cost anomalies, and a single carbon-footprint summary, served from an in-memory store. This document specifies the requirements to evolve FinOps into a **production-grade Cloud Cost Intelligence + Sustainability module** that closes the gap to `REQ-INF-013`/`REQ-INF-014` and delivers the "Deep FinOps with waste detection and right-sizing AI" differentiator promised in the SRS competitive analysis (§5.5).

In scope:
- Multi-cloud cost ingestion, normalization, allocation, anomaly detection, optimization, budgeting/forecasting, commitment management, unit economics, and carbon accounting.
- Persistent storage, REST APIs (v1), UI, alerting integration, RBAC/ABAC, observability, and tests.

Out of scope (tracked elsewhere):
- General metrics/log/trace ingestion (covered by core observability).
- Billing of NeuralOps itself (covered by §25 Pricing/Metering).

---

## 2. Current State (Baseline)

| Area | Current implementation | Limitation |
|---|---|---|
| API | `GET /finops/costs`, `GET /finops/anomalies`, `GET /finops/carbon` (`backend/internal/observability/phase5_handlers.go`) | Read-only, no write/config, no pagination, no time-range params |
| Data model | `FinOpsCostSeries`, `FinOpsCostPoint`, `FinOpsCostAnomaly`, `FinOpsCarbonFootprint` (`types.go`) | No resource-level granularity, allocation tags, commitments, or budgets |
| Store | In-memory seeded data (`phase5_store.go`); budget = `total * 1.05` | Not persistent, no real billing ingestion, naive budget |
| Anomalies | Static seeded list | No detection algorithm, no feedback, no alert routing |
| Carbon | Single static footprint | No per-resource/region attribution, no methodology |
| UI | FinOps tab in `CloudFinOpsNetwork.tsx` | Sparkline + lists only; no drilldown, allocation, or recommendations workflow |

---

## 3. Goals & Non-Functional Targets

- **G1**: Ingest real billing data from AWS (CUR), GCP (BigQuery Billing Export), and Azure (Cost Management Exports) with daily reconciliation.
- **G2**: Attribute ≥ 95% of cloud spend to an owner (team/service/environment) via tag/label rules + ML-assisted allocation of untagged spend.
- **G3**: Detect cost anomalies with precision ≥ 0.85 and surface them within 24h of the billing data landing.
- **G4**: Surface actionable savings (right-sizing, idle, commitment coverage) with quantified monthly USD impact and one-click ticketing.
- **G5**: Carbon accounting aligned to the GHG Protocol / SCI methodology, attributable by service, region, and provider.

Non-functional (inherits SRS §NFR):
- Cost dashboard P95 query latency < 2s for 13-month ranges over 1M resource-days.
- Ingestion pipeline handles ≥ 50M billing line items/day per tenant.
- All endpoints multi-tenant isolated, ABAC/residency enforced (reuse gateway middleware).

---

## 4. New Requirements

Requirement IDs use the `REQ-FINOPS-*` namespace and map back to SRS `REQ-INF-013/014`. Priority: P0 (must), P1 (should), P2 (could).

### 4.1 Data Ingestion & Normalization

- **REQ-FINOPS-001** [P0] The platform SHALL ingest cloud billing data from AWS Cost and Usage Reports (CUR), GCP BigQuery billing export, and Azure Cost Management exports via scheduled connectors with configurable cadence (default daily).
- **REQ-FINOPS-002** [P0] The platform SHALL normalize heterogeneous billing schemas into a canonical **FOCUS-aligned** cost line-item model (provider, account, service, region, resourceId, usageType, quantity, unit, amortizedCost, listCost, effectiveCost, tags, billingPeriod).
- **REQ-FINOPS-003** [P0] The platform SHALL support **amortized vs unblended vs blended** cost views and reconcile ingested totals against the provider invoice with a drift tolerance alert (> 1%).
- **REQ-FINOPS-004** [P1] The platform SHALL support ingestion of **custom/SaaS cost feeds** (Datadog, Snowflake, Kubernetes chargeback, on-prem) via a generic CSV/JSON cost-import API.
- **REQ-FINOPS-005** [P1] The platform SHALL deduplicate and version billing snapshots so late-arriving/restated provider data is reconciled idempotently.

### 4.2 Cost Allocation & Attribution

- **REQ-FINOPS-010** [P0] The platform SHALL allocate spend to organizational dimensions (team, service, environment, cost-center, product) using tag/label mapping rules with priority ordering.
- **REQ-FINOPS-011** [P0] The platform SHALL compute **Kubernetes cost allocation** (namespace, workload, pod) by combining cloud resource cost with cluster utilization (CPU/memory/GPU/storage/network) — including idle/unallocated cost split.
- **REQ-FINOPS-012** [P1] The platform SHALL provide **shared cost splitting** (proportional, even, weighted) for untagged/shared resources (e.g., NAT, data transfer, support).
- **REQ-FINOPS-013** [P1] The platform SHALL provide **ML-assisted tagging** suggestions for untagged spend and report tag-coverage % as a governance KPI.
- **REQ-FINOPS-014** [P2] The platform SHALL support **showback and chargeback** modes with exportable statements per cost-center.

### 4.3 Anomaly Detection & Alerting

- **REQ-FINOPS-020** [P0] The platform SHALL detect cost anomalies using seasonality-aware models (e.g., STL/Prophet-style decomposition + robust z-score) per allocation dimension and resource.
- **REQ-FINOPS-021** [P0] The platform SHALL route cost anomalies through the existing **alerting/notification pipeline** (alert policies, suppression, Slack/Jira/ServiceNow) with severity derived from absolute USD and % delta.
- **REQ-FINOPS-022** [P1] The platform SHALL support an **anomaly feedback loop** (confirm / false-positive / expected) that tunes future sensitivity, reusing the alert-feedback pattern.
- **REQ-FINOPS-023** [P1] The platform SHALL correlate cost anomalies with deploy/change events and traffic metrics to surface probable cause (e.g., release X, traffic surge, region failover).

### 4.4 Optimization & Waste Detection

- **REQ-FINOPS-030** [P0] The platform SHALL generate **right-sizing recommendations** for compute (EC2/GCE/VM, RDS/CloudSQL, containers) using P95/P99 utilization, with projected monthly savings and risk score.
- **REQ-FINOPS-031** [P0] The platform SHALL detect **idle/orphaned resources** (unattached disks, idle load balancers, stopped-but-billed, old snapshots, unused IPs) with savings estimate and safe-delete guidance.
- **REQ-FINOPS-032** [P1] The platform SHALL recommend **storage tiering / lifecycle** optimizations (object storage class transitions, snapshot retention).
- **REQ-FINOPS-033** [P1] The platform SHALL track each recommendation through a **lifecycle** (open → acknowledged → in-progress → realized → dismissed) and measure realized vs projected savings.
- **REQ-FINOPS-034** [P1] The platform SHALL allow one-click creation of a remediation ticket (Jira/ServiceNow) or an AutoFix plan (guardrailed) from a recommendation.

### 4.5 Commitments (RI / Savings Plans / CUDs)

- **REQ-FINOPS-040** [P1] The platform SHALL report **Reserved Instance / Savings Plan / Committed Use Discount** coverage, utilization, and expiry.
- **REQ-FINOPS-041** [P1] The platform SHALL provide **commitment purchase recommendations** with break-even and risk analysis based on steady-state usage.
- **REQ-FINOPS-042** [P2] The platform SHALL alert on **expiring or under-utilized commitments**.

### 4.6 Budgets, Forecasting & Unit Economics

- **REQ-FINOPS-050** [P0] The platform SHALL support **budgets** scoped to any allocation dimension with absolute and percent-of thresholds and multi-stage alerts (50/80/100/forecasted-overrun).
- **REQ-FINOPS-051** [P0] The platform SHALL **forecast** end-of-period spend per scope using trend + seasonality, with confidence intervals, replacing the current naive `total * 1.05` budget heuristic.
- **REQ-FINOPS-052** [P1] The platform SHALL compute **unit economics** metrics (cost per request/transaction/customer/tenant/feature) by joining cost with business/usage metrics.
- **REQ-FINOPS-053** [P2] The platform SHALL support **what-if scenario modeling** (e.g., region migration, instance family change, commitment purchase) with projected cost and carbon deltas.

### 4.7 Carbon & Sustainability

- **REQ-FINOPS-060** [P1] The platform SHALL estimate **CO2e per resource** using a documented methodology (GHG Protocol scope 2/3, grid carbon intensity by region) with versioned emission factors.
- **REQ-FINOPS-061** [P1] The platform SHALL report carbon footprint by service, region, provider, and time, plus **Software Carbon Intensity (SCI)** per unit of work.
- **REQ-FINOPS-062** [P2] The platform SHALL recommend **carbon-reducing actions** (region shift to higher renewable mix, schedule shifting) with cost trade-off shown alongside.

### 4.8 Reporting, Governance & Access

- **REQ-FINOPS-070** [P1] The platform SHALL provide **scheduled FinOps reports** (PDF/CSV, email/webhook) per scope, reusing the platform reporting service.
- **REQ-FINOPS-071** [P0] The platform SHALL enforce **RBAC + ABAC** so cost data is visible only to authorized scopes/tenants, honoring data residency.
- **REQ-FINOPS-072** [P1] The platform SHALL emit an **audit log** for budget changes, allocation-rule edits, and recommendation actions.
- **REQ-FINOPS-073** [P2] The platform SHALL expose FinOps entities via the **Terraform provider** (`neuralops_finops_budget`, `neuralops_finops_allocation_rule`).

---

## 5. API Design (v1, API-first)

All endpoints are tenant-scoped, versioned under `/api/v1/finops`, return the platform standard success/error envelope, and support `page`, `size`, `sort`, `filter`, and `from`/`to` time-range params where applicable.

### 5.1 Costs & Allocation
```
GET    /api/v1/finops/costs?scope&provider&granularity&from&to&groupBy   # enhanced
GET    /api/v1/finops/costs/breakdown?dimension&from&to                  # allocation tree
GET    /api/v1/finops/allocation/rules
POST   /api/v1/finops/allocation/rules
PUT    /api/v1/finops/allocation/rules/{id}
DELETE /api/v1/finops/allocation/rules/{id}
GET    /api/v1/finops/kubernetes/cost?cluster&namespace&from&to
POST   /api/v1/finops/imports                                            # custom cost feed
```

### 5.2 Anomalies
```
GET    /api/v1/finops/anomalies?scope&severity&status&from&to            # enhanced + filters
GET    /api/v1/finops/anomalies/{id}
POST   /api/v1/finops/anomalies/{id}/feedback                            # confirm|false_positive|expected
```

### 5.3 Optimization & Commitments
```
GET    /api/v1/finops/recommendations?type&scope&status
GET    /api/v1/finops/recommendations/{id}
POST   /api/v1/finops/recommendations/{id}/actions                       # acknowledge|ticket|autofix|dismiss
GET    /api/v1/finops/commitments?provider&status
GET    /api/v1/finops/commitments/recommendations
```

### 5.4 Budgets, Forecast, Unit Economics
```
GET    /api/v1/finops/budgets
POST   /api/v1/finops/budgets
PUT    /api/v1/finops/budgets/{id}
DELETE /api/v1/finops/budgets/{id}
GET    /api/v1/finops/forecast?scope&horizon
GET    /api/v1/finops/unit-economics?metric&scope&from&to
POST   /api/v1/finops/scenarios                                          # what-if modeling
```

### 5.5 Carbon & Reporting
```
GET    /api/v1/finops/carbon?scope&dimension&from&to                     # enhanced
GET    /api/v1/finops/carbon/recommendations
GET    /api/v1/finops/reports
POST   /api/v1/finops/reports/schedule
```

Backward compatibility: existing `GET /finops/costs|anomalies|carbon` responses remain valid; new fields are additive and new query params are optional.

---

## 6. Data Model & Migrations

New persistent tables (Postgres; Goose migration `0000xx_finops_enhancement.up.sql`). Use explicit columns, PK/FK, unique + check constraints, and indexes on `(tenant_id, billing_period)`, `(tenant_id, scope)`, and time columns. Avoid `SELECT *`.

- `finops_cost_line_items` — canonical FOCUS line items (partitioned by `billing_period`).
- `finops_allocation_rules` — tag/label → dimension mapping, priority, mode.
- `finops_budgets` — scope, period, thresholds, notification policy ref.
- `finops_anomalies` — detected anomalies + status + feedback (replaces static seed).
- `finops_recommendations` — type, target, projected/realized savings, lifecycle status.
- `finops_commitments` — RI/SP/CUD coverage, utilization, expiry.
- `finops_carbon_factors` — versioned emission factors per region/provider.
- `finops_carbon_samples` — per-resource CO2e samples.

Raw line-item analytics SHOULD use the existing object-storage/Iceberg lake (per SRS data-tiering); Postgres holds aggregates, config, and operational state. Reuse the **streaming materializer** (`backend/cmd/materializer`) to pre-aggregate daily cost roll-ups and anomaly score buckets.

---

## 7. Architecture (HLD/LLD pointers)

- **Connectors** (`backend/internal/finops/connectors/`): AWS/GCP/Azure billing ingestors built on existing `cloudinventory` SDK auth (`NEURALOPS_CLOUD_*`).
- **Normalizer** → canonical model → object-storage lake + Postgres aggregates.
- **Allocation engine** (`internal/finops/allocation`): rule evaluation + K8s utilization join.
- **Anomaly/forecast service** (`internal/finops/analytics`): seasonality models, reuse Kafka topic `neuralops.observability.stream` for derived cost metrics.
- **Recommendation engine** (`internal/finops/optimize`): right-sizing/idle/commitment analyzers.
- **Carbon service** (`internal/finops/carbon`): factor lookup + SCI.
- **Handlers** (`internal/observability/` or new `internal/finops/handler`) wired into gateway with ABAC/residency middleware.

Hexagonal boundaries: domain logic behind interfaces; connectors and persistence are adapters injected via DI. No hardcoded credentials — use the existing secret sources.

---

## 8. Security Design

- AuthN via existing gateway (OAuth2/OIDC/JWT/session). AuthZ via RBAC + ABAC scope filters on every query (`REQ-FINOPS-071`).
- Validate all headers/query/path/body; parameterized SQL only (no injection).
- Cost data is sensitive financial data: encrypt at rest, mask account IDs for non-privileged roles, never log raw billing payloads or cloud credentials.
- Residency: route/store per `X-Region-Target` and tenant residency policy.
- Audit all mutations (`REQ-FINOPS-072`).

---

## 9. Observability Design

- Structured JSON logs with `traceId/spanId/correlationId/tenantId` on ingestion and query paths.
- Prometheus metrics: `finops_ingest_line_items_total`, `finops_ingest_lag_seconds`, `finops_reconciliation_drift_pct`, `finops_anomaly_detected_total`, `finops_recommendation_savings_usd`, `finops_query_latency_seconds`.
- OpenTelemetry traces across connector → normalizer → store → API.
- Dogfooding: the module monitors its own ingestion freshness and alerts on stale billing data.

---

## 10. Test Strategy

- **Unit**: normalizer schema mapping, allocation rule precedence, anomaly scoring, forecast math, carbon factor lookup. Target ≥ 90% on `internal/finops/*`.
- **Integration**: ingest fixture CUR/BigQuery/Azure exports → assert canonical totals + reconciliation drift.
- **Contract**: `observability_contract_test.go`, `bank_openapi_contract_test.go`; canonical `docs/openapi/gateway-v1.yaml` synced via `./scripts/sync-openapi.sh` (includes `/finops/reconciliation`, `/finops/audit/export`).
- **E2E/UI**: FinOps tab drilldown, budget CRUD, recommendation action, anomaly feedback (extend `frontend` smoke specs).
- **Load**: k6 on `/finops/costs/breakdown` at 1M resource-days, P95 < 2s.
- **Security**: ABAC scope-isolation tests (tenant A cannot read tenant B cost), injection fuzz on filter params.

---

## 11. Frontend Enhancements

Evolve `frontend/src/pages/CloudFinOpsNetwork.tsx` FinOps tab into sub-tabs:
- **Overview** (trend, forecast band, budget burn-down).
- **Allocation** (cost breakdown tree by team/service/namespace, tag coverage %).
- **Anomalies** (filters, detail drawer, feedback actions).
- **Optimization** (recommendations table, projected/realized savings, ticket/AutoFix actions).
- **Commitments** (coverage/utilization/expiry).
- **Sustainability** (carbon by dimension, SCI, recommendations with cost trade-off).

Reuse existing `Card`, `Badge`, `Select`, `PageStates`, React Query patterns, and i18n keys.

---

## 12. Phased Delivery

| Wave | Requirements | Outcome |
|---|---|---|
| **Wave 1 (P0)** | 001–003, 010–011, 020–021, 030–031, 050–051, 071 | Real ingestion, allocation, anomaly alerting, right-sizing/idle, budgets+forecast, secured |
| **Wave 2 (P1)** | 004–005, 012–013, 022–023, 032–034, 040–041, 052, 060–061, 070, 072 | Optimization lifecycle, commitments, unit economics, carbon, reports |
| **Wave 3 (P2)** | 014, 042, 053, 062, 073 | Chargeback, scenarios, carbon actions, Terraform, advanced governance |

---

## 13. Acceptance Criteria (Exit)

- ≥ 95% of ingested spend attributed to an owner; reconciliation drift < 1% vs invoice.
- Anomaly precision ≥ 0.85 on labeled validation set; alerts delivered via existing pipeline.
- At least three recommendation types live with realized-savings tracking.
- Budgets with forecast-based early warning replace the `total * 1.05` heuristic.
- Carbon reporting attributable by service/region/provider with documented methodology.
- All new endpoints in OpenAPI, contract-tested, ABAC-isolated, and observable via metrics/traces.
