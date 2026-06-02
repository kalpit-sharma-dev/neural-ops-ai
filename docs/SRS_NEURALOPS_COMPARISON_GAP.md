# NeuralOps vs NEXOBS SRS — Comprehensive Comparison and Gap Plan

Date: 2026-06-02  
Inputs compared:
- `SRS_NextGen_Observability_Platform.md`
- `docs/openapi/gateway-v1.yaml`
- `frontend/src/router.tsx`
- `ARCHITECTURE.md`
- `API.md`, `docs/API.md`
- Platform docs in `docs/` and root runbooks

---

## 1) Executive Summary

NeuralOps already implements a strong cross-domain observability foundation: logs, traces/APM, metrics, dashboards, incidents, alerting, workflow automation, notebooks, RUM/synthetic, integrations, security signals, RBAC, SSO, and multi-tenant scaffolding.

Against the SRS, the largest remaining gaps are in:
- Deep agent/collector productization (single auto-instrumenting agent, fleet management, operator maturity)
- Unified cross-signal query language (NexQL equivalent)
- Advanced infra/network/cloud-finops breadth
- Full AppSec depth (RASP/SCA/CSPM/image scanning)
- Enterprise governance depth (ABAC, residency controls, MSP/white-label)
- Advanced AI features (causal graph explainability, forecasting, AutoFix safety framework, LLM workload observability)

---

## 2) Domain Coverage Snapshot

Legend:
- `Implemented`: clearly present in code/docs/runtime surface
- `Partial`: present but not full SRS depth
- `Missing`: no strong evidence in current code/docs

| SRS Domain | Status | Notes |
|---|---|---|
| Data collection & instrumentation (`REQ-COLL-*`) | Partial | OTel + ingest + agents exist, but not full NEXAGENT/fleet/operator scope |
| Metrics/monitoring/alerting (`REQ-MET-*`, `REQ-ALERT-*`) | Partial | Metrics catalog/query + PromQL + alerting exist; NexQL/unlimited cardinality and advanced routing/scoring missing |
| Tracing & APM (`REQ-APM-*`) | Partial | Strong tracing/service map/retention/profiling/DBM baseline; code-level/prod-grade sampling and service catalog depth incomplete |
| Logs & analytics (`REQ-LOG-*`) | Partial | Structured/semantic/AI search and parsing rules exist; live tail/tiering/pattern depth needs expansion |
| Infrastructure & cloud (`REQ-INF-*`) | Partial | Host/K8s/cloud pages and collectors exist; serverless/asset inventory/multi-cloud topology/finops-carbon depth incomplete |
| Network performance monitoring (`REQ-NPM-*`) | Missing/Partial | Minimal middleware/Kafka lag visibility; full NPM/flows/SNMP/SD-WAN/wireless missing |
| RUM & synthetic (`REQ-RUM-*`, `REQ-SYN-*`) | Partial | RUM beacon/replay + synthetic monitors exist; funnel/mobile synthetic/CI-private locations incomplete |
| Security observability (`REQ-SEC-*`) | Partial | Runtime vulnerability/attack signals exist; RASP/SCA/image/CSPM/SIEM-grade integration incomplete |
| Business observability | Partial | Dashboards/metrics/alerts support building blocks; domain-specific business KPI frameworks mostly custom/manual |
| AI/ML engine (`REQ-AI-*`) | Partial | AI chat/anomaly/RCA patterns exist; explainable causal AI, forecasting, AutoFix governance, LLM observability incomplete |
| Dashboards/reporting | Partial | Dashboard CRUD strong; scheduled executive reporting and advanced exports incomplete |
| Incident management/collab | Partial | Incident lifecycle + on-call/integrations exist; deeper war-room/collab workflow automation incomplete |
| Integrations/ecosystem (`REQ-INT-*`) | Partial | Jira/Slack/ServiceNow, OAuth, webhook, marketplace exist; broad CI/CD/BI/SIEM/provider/CLI parity incomplete |
| Platform admin & multi-tenancy (`REQ-ADM-*`) | Partial | RBAC, SSO, API keys, audit, usage, tenant policies exist; ABAC, residency and MSP/white-label depth missing |
| Non-functional requirements (`REQ-NFR-*`) | Partial | Good CI/tests/compose+k8s/mTLS/auth hardening; formal SLO/SLA, multi-region active-active, i18n/WCAG proofing incomplete |

---

## 3) What NeuralOps Already Does Well (SRS-Aligned)

- Unified platform with logs/metrics/traces/incidents/alerts/dashboards.
- API gateway with auth, RBAC, API keys, SSO paths, audit and usage surfaces.
- Trace/search/semantic/AI query endpoints and APM pages (`traces`, `service-flow`, `service-map`, `trace-compare`).
- Workflow automation (including branching graph persistence and execution) plus notebook execution.
- Security runtime signal pages (`/security/vulnerabilities`, `/security/attacks`) and detail drilldown.
- Marketplace and integration lifecycle (install/configure/connect/disconnect/oauth).
- RUM + session replay + synthetic monitor foundational APIs and UI.
- Cloud/infrastructure/Kubernetes/databases/middleware dedicated pages and data pipelines.
- OpenAPI contract, CI quality gates, integration tests, compose smoke verification, mTLS verification.

---

## 4) Comprehensive Missing Features Backlog (to reach SRS parity)

This section is intentionally action-oriented so you can implement systematically.

## 4.1 Collection / Agent Platform

### Missing or partial
- `REQ-COLL-001/003/004/007/016`: Production-grade single agent (`NEXAGENT`) with deep eBPF + auto-instrumentation + fleet controls.
- `REQ-COLL-009`: Full Kubernetes Operator for lifecycle.
- `REQ-COLL-008`: Air-gapped robust buffering and replay guarantees.
- `REQ-COLL-011/017`: IoT edge collector and visual collector pipeline composer.

### Build next
1. Agent management service (inventory, policy rollout, health, upgrades).
2. Agent CRDs + operator with canary rollout and policy reconciliation.
3. Collector pipeline DSL + UI (parse/enrich/drop/route/scrub).
4. Benchmark harness proving footprint targets.

## 4.2 Metrics, Query, Alerting

### Missing or partial
- `REQ-MET-003`: High-cardinality strategy with explicit query/index safeguards.
- `REQ-MET-006`: Unified cross-signal query language (NexQL equivalent).
- `REQ-MET-007`: Derived metrics engine.
- `REQ-ALERT-004/005/008`: Advanced routing trees, suppression windows, alert-fatigue scoring.

### Build next
1. Unified query service (metrics+logs+traces+events planner).
2. Derived-metric compiler and materialization strategy.
3. Alert policy graph (ownership, escalation, maintenance windows, dedupe keys).
4. Alert quality scoring model + feedback loop.

## 4.3 APM & Code Intelligence

### Missing or partial
- `REQ-APM-003`: Tail-based sampling governance.
- `REQ-APM-012`: Code-level visibility with blame/commit linkage.
- `REQ-APM-014`: Service catalog/ownership integration depth.
- `REQ-APM-015`: Broader runtime parity evidence.

### Build next
1. Sampling policy API + dynamic controls.
2. Span-to-repo mapping service (service -> repo -> team -> runbook).
3. Code-level error correlation (stack, release, commit, owner).

## 4.4 Log Analytics Depth

### Missing or partial
- `REQ-LOG-008`: Strong pattern clustering/outlier templates.
- `REQ-LOG-011`: True live tail stream UX/API.
- `REQ-LOG-012`: Tiered storage lifecycle and restore workflow.

### Build next
1. Live tail websocket stream with backpressure.
2. Pattern clustering job + explainable grouping metadata.
3. Hot/warm/cold tier policy engine with retrieval SLA.

## 4.5 Infra, Cloud, FinOps, Network

### Missing or partial
- `REQ-INF-010/011`: Cloud asset inventory and multi-cloud topology depth.
- `REQ-INF-012`: Serverless observability breadth.
- `REQ-INF-013/014`: Cost intelligence + carbon module.
- `REQ-NPM-*`: Dedicated network performance suite.

### Build next
1. Cloud inventory ingestors (AWS/GCP/Azure asset graph).
2. FinOps cost model + budget anomaly alerts.
3. NPM service: flow ingestion (NetFlow/sFlow/IPFIX), SNMP collectors, topology overlays.
4. Serverless traces/metrics/log correlation package.

## 4.6 Security (AppSec + SecObs)

### Missing or partial
- `REQ-SEC-001/002/003/004`: RASP, SCA, image scanning, CSPM.
- `REQ-SEC-007`: SIEM-grade push integrations and schema mapping.
- `REQ-SEC-008/009`: Network threat + cloud posture depth.

### Build next
1. Security findings domain model (vuln, exploitability, remediation state).
2. SCA and image scan ingestion connectors (Snyk/Trivy/Grype etc.).
3. CSPM policy engine and posture drift detection.
4. SIEM connector framework with bidirectional case sync.

## 4.7 AI / AIOps / LLM Observability

### Missing or partial
- `REQ-AI-004/005/006`: Explainable causal dependency graph with confidence and provenance.
- `REQ-AI-007/008/009`: Forecasting/capacity/predictive incidents.
- `REQ-AI-017/019/020`: Safe autonomous remediation + runbook generation + alert quality scoring.
- `REQ-AI-021/022`: Dedicated LLM workload and AI pipeline observability module.

### Build next
1. Causal graph service with evidence model and “why” explanations.
2. Forecasting pipeline (capacity, SLO burn prediction, incident risk).
3. AutoFix governance (approval gates, policy guardrails, blast-radius checks, rollback).
4. LLM telemetry schema (token, latency, model, prompt template, eval, hallucination signals).

## 4.8 Integrations & Ecosystem

### Missing or partial
- `REQ-INT-001/002/003`: Wider SCM/CI/CD/IDE ecosystem parity.
- `REQ-INT-011..014`: BI, warehouse, data-science export maturity.
- `REQ-INT-021/022`: First-class Terraform/Pulumi providers and official CLI.

### Build next
1. Connector SDK + certification model.
2. Export jobs and streaming sinks with schema contracts.
3. `neuralops` CLI (auth, query, workflows, incidents, dashboards, admin).
4. Terraform provider resources for alerts, dashboards, SLOs, integrations, users/policies.

## 4.9 Administration, Governance, Multi-Tenancy

### Missing or partial
- `REQ-ADM-003`: ABAC policy engine and enforcement.
- `REQ-ADM-007/008/009`: Residency, masking policy depth, per-signal retention governance.
- `REQ-ADM-011/012`: White-label and MSP console.

### Build next
1. Central policy decision point (RBAC + ABAC).
2. Data-governance controls (region locks, masking transforms, lifecycle policy registry).
3. MSP hierarchy with delegated admin and tenant switchboard.

## 4.10 Reporting, Collaboration, Business Observability

### Missing or partial
- Scheduled executive reporting, compliance report packs, SLA exports.
- Incident collaboration (war-room timeline, decision log, postmortem workflow templates).
- Business-observability reference models (domain KPI packs).

### Build next
1. Reporting service (scheduled PDF/CSV/email/webhook).
2. Incident collaboration workspace with timeline + action ownership + postmortem automation.
3. Domain packs (payments, e-commerce, SaaS, infra) with pre-built KPIs.

---

## 5) Priority Implementation Plan (Pragmatic)

## Wave 1 (P0 parity: highest value/risk)
- Unified query layer MVP (cross-signal search)
- Alert routing/suppression hardening + AI correlation quality
- Security baseline uplift: SCA + image findings ingest + SIEM export
- Agent/fleet management baseline and policy-controlled rollout
- Causal AI explainability for RCA paths

## Wave 2 (P1 enterprise scale)
- Forecasting/capacity planning and predictive incidents
- FinOps/cloud asset inventory and richer infra topology
- Terraform provider + CLI v1
- ABAC and data governance controls
- Reporting and collaboration automation

## Wave 3 (P2 differentiation)
- AutoFix with safe-guardrails
- LLM observability full module + AI pipeline observability
- Carbon intelligence and advanced NPM (SD-WAN/wireless)
- White-label + MSP control plane

---

## 6) Suggested Deliverable Structure for Execution

Create one epic per SRS domain with:
- Requirement mapping table (`REQ-*` -> story IDs)
- API contract additions (OpenAPI first)
- Data model changes + migrations
- UI screens and workflows
- Test strategy (unit/integration/e2e/load/security)
- Exit criteria (observable metrics and acceptance tests)

---

## 7) Requirement-ID Gap Pointers (High-confidence)

The following requirement families are substantially not-yet-complete in NeuralOps and should be treated as backlog anchors:

- Collection: `REQ-COLL-009`, `REQ-COLL-011`, `REQ-COLL-016`, `REQ-COLL-017`
- Metrics/Query: `REQ-MET-003`, `REQ-MET-006`, `REQ-MET-007`
- APM: `REQ-APM-012`, `REQ-APM-014`, `REQ-APM-015`
- Logs: `REQ-LOG-008`, `REQ-LOG-011`, `REQ-LOG-012`
- Infra/Cloud: `REQ-INF-010`, `REQ-INF-011`, `REQ-INF-013`, `REQ-INF-014`
- NPM: `REQ-NPM-001` through `REQ-NPM-006`
- RUM/Synthetic: `REQ-RUM-003`, `REQ-RUM-005`, `REQ-SYN-002`, `REQ-SYN-003`, `REQ-SYN-004`, `REQ-SYN-005`
- Security: `REQ-SEC-001`, `REQ-SEC-002`, `REQ-SEC-003`, `REQ-SEC-004`, `REQ-SEC-007`, `REQ-SEC-008`, `REQ-SEC-009`
- AI: `REQ-AI-005`, `REQ-AI-006`, `REQ-AI-007`, `REQ-AI-008`, `REQ-AI-009`, `REQ-AI-017`, `REQ-AI-019`, `REQ-AI-020`, `REQ-AI-021`, `REQ-AI-022`
- Integrations/Admin: `REQ-INT-011..014`, `REQ-INT-021`, `REQ-INT-022`, `REQ-ADM-003`, `REQ-ADM-007`, `REQ-ADM-011`, `REQ-ADM-012`

---

## 8) Notes and Caveats

- This comparison is evidence-based from repository artifacts and exposed routes/contracts; some capabilities may exist experimentally but are not yet productized or clearly documented.
- Several domains are present functionally but still “Partial” due to enterprise-grade depth expected in the SRS (scale, governance, ecosystem breadth, and hard SLO targets).

