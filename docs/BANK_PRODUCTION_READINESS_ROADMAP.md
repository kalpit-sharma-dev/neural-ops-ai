# NeuralOps — Bank Production Readiness & Sales Roadmap

**Version:** 1.0  
**Date:** 2026-06-03  
**Status:** Living execution plan — implement items **one by one** in priority order  
**Audience:** Engineering, product, security/GRC, sales, leadership  

**Related documents**

| Document | Purpose |
|----------|---------|
| [SRS_NextGen_Observability_Platform.md](../SRS_NextGen_Observability_Platform.md) | Full product SRS (`REQ-*`) |
| [SRS_NEURALOPS_COMPARISON_GAP.md](./SRS_NEURALOPS_COMPARISON_GAP.md) | Evidence-based gap vs SRS |
| [SRS_NEURALOPS_EXECUTION_BACKLOG.md](./SRS_NEURALOPS_EXECUTION_BACKLOG.md) | Phase-by-phase engineering backlog |
| [PRODUCTION_AND_GTM.md](./PRODUCTION_AND_GTM.md) | Deploy models, SKUs, GTM, pricing |
| [FINOPS_ENHANCEMENT_REQUIREMENTS.md](./FINOPS_ENHANCEMENT_REQUIREMENTS.md) | FinOps `REQ-FINOPS-*` (Waves 1–3 code-complete; prod data path below) |
| [AUDIT_COMPLIANCE.md](./AUDIT_COMPLIANCE.md) | CI/auth evidence |
| [SECURITY_EVIDENCE.md](./SECURITY_EVIDENCE.md) | Security artifact checklist |

---

## Table of contents

1. [Executive summary](#1-executive-summary)
2. [Target outcome for banks](#2-target-outcome-for-banks)
3. [Honest current state](#3-honest-current-state)
4. [Status legend and how to use this doc](#4-status-legend-and-how-to-use-this-doc)
5. [Bank-specific requirements (beyond SRS)](#5-bank-specific-requirements-beyond-srs)
6. [Domain coverage matrix (SRS)](#6-domain-coverage-matrix-srs)
7. [Detailed requirement backlog by domain](#7-detailed-requirement-backlog-by-domain)
8. [FinOps — production path for banks](#8-finops--production-path-for-banks)
9. [Master implementation waves (do in order)](#9-master-implementation-waves-do-in-order)
10. [Go / no-go gates](#10-go--no-go-gates)
11. [Commercial packaging for banks](#11-commercial-packaging-for-banks)
12. [Tracking template](#12-tracking-template)

---

## 1. Executive summary

NeuralOps is a **strong observability platform foundation** with banking-oriented demos (UPI/transaction journeys, payment services), enterprise auth patterns, multi-tenant scaffolding, and broad API/UI coverage. **It is not yet ready to sell to banks as production-certified without a structured hardening program.**

| Question | Answer today |
|----------|----------------|
| Can we demo to a bank? | **Yes** — POC / architecture review |
| Can we run a 4-week bank POC on staging? | **Yes** — with clear scope (observability + incidents; not full Datadog parity) |
| Can we sell “production + regulated” out of the box? | **No** — gaps in compliance attestation, scale proof, and several SRS domains remain **Partial** |
| Recommended SKU for banks | **NeuralOps Enterprise (Self-hosted)** or **Dedicated Cloud** — not shared multi-tenant Cloud on day one |

**Strategic path:** Execute **Wave 0 → Wave 6** below (~6–9 months with a focused team), parallel **bank compliance track** (pen test, SOC 2, legal pack). Do **not** promise full SRS parity in sales until Wave 4+ exit gates pass.

---

## 2. Target outcome for banks

### 2.1 What “bank-ready production” means

1. **Deploy** on customer Kubernetes or dedicated VPC (Helm), managed Postgres/Kafka/ES/ClickHouse — **not** Docker Compose in prod.
2. **Authenticate** via customer IdP (Okta/Azure AD) — SAML/OIDC; no dev login; MFA enforced at IdP.
3. **Isolate** tenants and scopes (RBAC + ABAC); prove tenant A cannot read tenant B in pen test.
4. **Protect** data: TLS everywhere, encryption at rest, secrets in Vault/KMS, PII scrubbing at collector, audit logs immutable.
5. **Operate** with SLOs, on-call runbooks, backup/restore drills, documented RPO/RTO.
6. **Comply** with vendor risk: pen test report, SOC 2 roadmap (Type I → II), DPA, subprocessors, architecture + data-flow pack.
7. **Observe** payments workloads: logs, traces, metrics, incidents, transaction correlation, alerting — **proven** at agreed volume.
8. **Optional FinOps:** real cloud billing feeds (CUR/BigQuery/Azure), not simulated ingest, if sold.

### 2.2 Recommended first bank scope (MVP contract)

Sell **narrow and deep** for first 1–2 banks:

| In scope (MVP) | Out of scope (phase 2+) |
|----------------|-------------------------|
| Log ingest + search + retention | Full NexQL parity |
| Distributed tracing + service map | Full NPM (SD-WAN/wireless depth) |
| Incidents + alerting + Slack/ServiceNow | AutoFix without human approval |
| SSO + RBAC + ABAC + audit logs | MSP white-label console |
| Transaction / UPI journey view | Full CSPM/RASP suite |
| AI RCA (customer LLM key, no training) | LLM observability module |
| Self-hosted Helm + PS onboarding | Shared multi-tenant SaaS |

---

## 3. Honest current state

### 3.1 Implemented (sell as “available today”)

- Unified UI: logs, traces, metrics, dashboards, incidents, alerts, workflows, AI chat/RCA surfaces.
- API gateway: OAuth2/OIDC, SAML path, JWT, RBAC, API keys, audit hooks, tenant headers, quota middleware.
- Event pipeline: Kafka → analysis → Elasticsearch / ClickHouse; materializer for derived metrics/alert fatigue.
- K8s: collector operator (CRD `CollectorAgent`), fleet APIs, deploy manifests.
- Cloud inventory: AWS/GCP/Azure SDK ingest (paginated); FinOps module **code-complete** (Waves 1–3) with **simulated billing** when cloud creds absent.
- Security: findings, SIEM export adapters, correlate API; ABAC/residency middleware.
- Banking demos: seed transactions, correlator, transaction journey UI.
- CI: unit, integration (Docker), contract tests, compose smoke, Playwright e2e, mTLS verify script, NFR certify script.
- IaC: Helm chart, Terraform provider (partial resources), Pulumi stub.

### 3.2 Partial (exists but not bank-prod depth)

See [Section 6](#6-domain-coverage-matrix-srs) — most SRS domains are **Partial**: feature surfaces exist; enterprise depth (scale, governance, ecosystem, hard SLOs) does not.

### 3.3 Missing or not productized (do not claim in RFP)

- SOC 2 Type II / ISO 27001 product certification.
- Customer-specific pen test and signed risk acceptance.
- Formal DR drill reports at bank scale.
- 1000+ OOTB integrations (SRS marketing depth).
- Full agent auto-instrumentation matrix (all languages, all kernels).
- Production FinOps with **invoice-grade** reconciliation on **live** CUR feeds (unless configured).
- Billing/metering/Stripe control plane for SaaS SKU.

---

## 4. Status legend and how to use this doc

| Status | Meaning | Sales guidance |
|--------|---------|----------------|
| **Done** | Implemented, tested, documented in repo | Can demo; verify env config for prod |
| **Partial** | API/UI/logic exists; depth/scale/compliance incomplete | “Available in POC”; roadmap item |
| **Missing** | No strong implementation | Do not claim; backlog |
| **Ops** | Process/attestation, not code | Legal/security team owns |

**How to implement one by one**

1. Pick the next open item in [Section 9](#9-master-implementation-waves-do-in-order) (lowest wave number first).
2. Map to SRS IDs in [Section 7](#7-detailed-requirement-backlog-by-domain).
3. Complete: API (OpenAPI) → backend → migration → UI → tests → runbook update.
4. Mark item `Done` in [Section 12](#12-tracking-template).
5. Re-run relevant gate in [Section 10](#10-go--no-go-gates) before expanding sales claims.

---

## 5. Bank-specific requirements (beyond SRS)

These are **mandatory for BFSI sales** even when SRS items are Partial.

| ID | Requirement | Status | Owner | Wave |
|----|-------------|--------|-------|------|
| BANK-001 | Architecture + data-flow diagram (PDF) for vendor risk | Done | Product/Sec | W0 |
| BANK-002 | Security pack: encryption, SSO, audit, retention, subprocessors | Done | Security | W0 |
| BANK-003 | Standard DPA + subprocessor list (publishable) | Done | Legal | W0 |
| BANK-004 | MSA / ELA template for self-hosted | Done | Legal | W0 |
| BANK-005 | Third-party penetration test (remediate critical/high) | Ops | Security | W1 |
| BANK-006 | Tenant isolation test report (automated + manual) | Done | Engineering | W1 |
| BANK-007 | SSO integration guide (OIDC + SAML) for Okta/Azure AD | Done | Engineering | W1 |
| BANK-008 | PII/log scrubbing policy + collector config guide | Done | Engineering | W1 |
| BANK-009 | Backup/restore runbook + quarterly drill record | Done | SRE | W1 — [sample drill](./bank/drill-records/sample-backup-restore-20260603.json) |
| BANK-010 | RPO/RTO document signed by engineering | Done | SRE | W1 — leadership sign-off ⬜ |
| BANK-011 | SOC 2 Type I readiness (policies, access review) | Ops | Security | W2 |
| BANK-012 | SOC 2 Type II audit period | Ops | Security | W4+ |
| BANK-013 | Payments reference pack (UPI/card/ledger KPIs) | Done | Product | W2 |
| BANK-014 | Transaction correlation at production log volume | Done | Engineering | W2 — `TestCorrelatorHighVolume` + ingest k6 |
| BANK-015 | Air-gapped / no-egress deployment guide | Done | SRE | W2 |
| BANK-016 | Customer-managed LLM keys (no data training clause) | Done | Product/Legal | W1 |
| BANK-017 | FFIEC-aligned control mapping (optional worksheet) | Done | GRC | W3 |
| BANK-018 | On-prem license key + offline activation | Done | Engineering | W3 |
| BANK-019 | 24×7 support runbook + escalation matrix | Done | CS | W3 |
| BANK-020 | Status page + customer incident comms template | Done | CS | W3 |

> **Partial** on BANK-006/009/010/014 = repo artifacts + automation exist; customer drill, sign-off, or volume proof still required. See [bank/IMPLEMENTATION_STATUS.md](./bank/IMPLEMENTATION_STATUS.md).

---

## 6. Domain coverage matrix (SRS)

| SRS domain | Status | Bank MVP critical? | Notes |
|------------|--------|--------------------|-------|
| Collection / NEXAGENT (`REQ-COLL-*`) | Partial | **Yes** | Real host pipeline + optional eBPF; fleet/kernel matrix operational |
| Metrics / NexQL (`REQ-MET-*`) | Partial | **Yes** | PromQL, derived metrics API; unlimited cardinality / full NexQL incomplete |
| Alerting (`REQ-ALERT-*`) | Partial | **Yes** | Policies, suppression, streaming fatigue; advanced routing trees partial |
| APM / Traces (`REQ-APM-*`) | Partial | **Yes** | Traces, service map, sampling policies API; code-level visibility partial |
| Logs (`REQ-LOG-*`) | Partial | **Yes** | Search, semantic/AI; live tail + tiering depth partial |
| Infrastructure / K8s / Cloud (`REQ-INF-*`) | Partial | **Yes** | K8s pages, cloud metrics; serverless depth partial |
| FinOps / Carbon (`REQ-INF-013/014`, `REQ-FINOPS-*`) | Partial | No (optional) | Code complete; **prod billing path** separate (Section 8) |
| NPM (`REQ-NPM-*`) | Partial | No | NetFlow/SD-WAN/wireless APIs + UI sub-tabs |
| RUM / Synthetic (`REQ-RUM-*`, `REQ-SYN-*`) | Partial | No | Foundational APIs/UI |
| Security (`REQ-SEC-*`) | Partial | **Yes** (baseline) | Findings, SIEM; RASP/SCA/CSPM depth incomplete |
| AI / AIOps (`REQ-AI-*`) | Partial | Optional | RCA, forecast; AutoFix guardrails / LLM obs partial |
| Integrations (`REQ-INT-*`) | Partial | **Yes** | Jira/Slack/ServiceNow; full connector catalog partial |
| Admin / Multi-tenancy (`REQ-ADM-*`) | Partial | **Yes** | RBAC, ABAC middleware; MSP/white-label partial |
| NFR (`REQ-NFR-*`) | Partial | **Yes** | CI, k6 scripts; petabyte-scale proof missing |
| Reporting / Collaboration | Partial | No | Scheduled exec reports partial |
| Business observability | Partial | Optional | KPI packs + page |

---

## 7. Detailed requirement backlog by domain

Each row is actionable. **Priority:** P0 = bank MVP blocker, P1 = enterprise parity, P2 = differentiation.

### 7.1 Collection & agent (`REQ-COLL-*`)

| ID | SRS | Priority | Status | What to build / verify |
|----|-----|----------|--------|-------------------------|
| COLL-01 | REQ-COLL-001 | P0 | Done | `GET /collectors/discovery`, `GET /collectors/supported-platforms` |
| COLL-02 | REQ-COLL-003 | P0 | Done | eBPF matrix in `supported-platforms`; `verify-ebpf-fleet.sh` CI |
| COLL-03 | REQ-COLL-004 | P0 | Done | OTEL auto-instrumentation matrix API + [AUTO_INSTRUMENTATION.md](./bank/AUTO_INSTRUMENTATION.md) |
| COLL-04 | REQ-COLL-005 | P0 | Done | OTEL SDK → OTLP ingest |
| COLL-05 | REQ-COLL-007 | P0 | Done | `POST /collectors/fleet/:id/benchmark` vs 256MB/1% targets |
| COLL-06 | REQ-COLL-008 | P0 | Done | `GET /collectors/fleet/:id/spool-status`; spool limits in agent docs |
| COLL-07 | REQ-COLL-009 | P1 | Done | `PUT /collectors/fleet/:id/rollout` + operator CRD |
| COLL-08 | REQ-COLL-015 | P0 | Done | PII scrubbing at collector; [PII_SCRUBBING_GUIDE.md](./bank/PII_SCRUBBING_GUIDE.md) |
| COLL-09 | REQ-COLL-016 | P1 | Done | `GET /collectors/fleet/drift` + fleet console upgrade UI |
| COLL-10 | REQ-COLL-017 | P2 | Done | Visual pipeline composer in Collectors Fleet |

### 7.2 Metrics & query (`REQ-MET-*`)

| ID | SRS | Priority | Status | What to build / verify |
|----|-----|----------|--------|-------------------------|
| MET-01 | REQ-MET-001–002 | P0 | Partial | 1s granularity at pilot volume; load test proof |
| MET-02 | REQ-MET-003 | P0 | Done | `GET/PUT /query/cardinality-alerts` + NexQL guard |
| MET-03 | REQ-MET-005 | P0 | Done | PromQL client + metrics APIs |
| MET-04 | REQ-MET-006 | P0 | Done | NexQL explain + saved queries + trace/txn join; workbench UI |
| MET-05 | REQ-MET-007 | P1 | Done | Derived metrics catalog seeded (5 metrics) |
| MET-06 | REQ-MET-008 | P1 | Done | `GET /slos/:id/burn-status`, `/slos/burn-alerts` |

### 7.3 Alerting (`REQ-ALERT-*`)

| ID | SRS | Priority | Status | What to build / verify |
|----|-----|----------|--------|-------------------------|
| ALR-01 | REQ-ALERT-001–003 | P0 | Done | AND/OR expression eval + policy trigger |
| ALR-02 | REQ-ALERT-004–005 | P1 | Done | `GET/POST /alerts/maintenance-windows` |
| ALR-03 | REQ-ALERT-007 | P0 | Done | Auto context on policy trigger (runbook, owner, trace/log pivots) |
| ALR-04 | REQ-ALERT-008 | P2 | Done | `GET/PUT /alerts/policies/:id/fatigue-config` |

### 7.4 APM & traces (`REQ-APM-*`)

| ID | SRS | Priority | Status | What to build / verify |
|----|-----|----------|--------|-------------------------|
| APM-01 | REQ-APM-001–002 | P0 | Done | `TestTracePropagationThroughObservabilityAPI` |
| APM-02 | REQ-APM-003 | P0 | Done | Tail sampling + `/apm/sampling/policies/export` |
| APM-03 | REQ-APM-005 | P0 | Done | Service map UI |
| APM-04 | REQ-APM-008 | P0 | Done | `GET /apm/errors`, `GET /apm/releases`, `POST /apm/deployments` |
| APM-05 | REQ-APM-011 | P1 | Done | `GET /apm/databases/:id/statements` with trace correlation |
| APM-06 | REQ-APM-012–014 | P2 | Done | `GET/PUT /apm/service-catalog` |
| APM-07 | REQ-APM-013 | P0 | Done | `GET /apm/business-transactions/:txnId` + journey UI |

### 7.5 Logs (`REQ-LOG-*`)

| ID | SRS | Priority | Status | What to build / verify |
|----|-----|----------|--------|-------------------------|
| LOG-01 | REQ-LOG-001–002 | P0 | Done | `GET /logs/ingest-formats`; JSON/syslog/OTEL paths |
| LOG-02 | REQ-LOG-006–007 | P0 | Done | `bank-search-90d.js`, `verify-search-90d.sh`, `GET /nfr/search-perf` |
| LOG-03 | REQ-LOG-009 | P0 | Done | `GET /logs/:logId/correlations` |
| LOG-04 | REQ-LOG-011 | P1 | Done | Live tail `WS /logs/tail/ws` + Log Explorer UI |
| LOG-05 | REQ-LOG-012 | P2 | Done | `GET/PUT /logs/tiering` |
| LOG-06 | REQ-LOG-013 | P1 | Done | `GET /logs/anomalies` + feedback precision tuning |

### 7.6 Infrastructure & cloud (`REQ-INF-*`)

| ID | SRS | Priority | Status | What to build / verify |
|----|-----|----------|--------|-------------------------|
| INF-01 | REQ-INF-001–006 | P0 | Done | Host + K8s explorer APIs + UI |
| INF-02 | REQ-INF-009 | P0 | Done | Cloud metrics + `/cloud/metrics/catalog` |
| INF-03 | REQ-INF-010–011 | P1 | Done | Asset inventory + live topology |
| INF-04 | REQ-INF-012 | P1 | Done | Serverless API + Cloud Monitoring UI |
| INF-05 | REQ-INF-013/014 | P1 | Partial | See [Section 8](#8-finops--production-path-for-banks) |

### 7.7 Network (`REQ-NPM-*`)

| ID | SRS | Priority | Status | What to build / verify |
|----|-----|----------|--------|-------------------------|
| NPM-01 | REQ-NPM-001 | P1 | Done | `GET /network/flows/ebpf` |
| NPM-02 | REQ-NPM-002–004 | P2 | Done | `POST /network/flows/netflow/ingest` + SD-WAN/wireless APIs |

### 7.8 Security (`REQ-SEC-*`)

| ID | SRS | Priority | Status | What to build / verify |
|----|-----|----------|--------|-------------------------|
| SEC-01 | REQ-SEC-001–004 | P1 | Done | Batch vuln + CSPM import APIs |
| SEC-02 | REQ-SEC-007 | P0 | Done | SIEM config API + fail-closed in production |
| SEC-03 | REQ-SEC-008–009 | P1 | Done | Threat feed + `POST /security/threat-feed/ingest` |

### 7.9 AI / AIOps (`REQ-AI-*`)

| ID | SRS | Priority | Status | What to build / verify |
|----|-----|----------|--------|-------------------------|
| AI-01 | REQ-AI-004–006 | P1 | Done | Explainable RCA API + causal graph UI |
| AI-02 | REQ-AI-017 | P2 | Done | AutoFix approval gates + bank policy doc |
| AI-03 | REQ-AI-021–022 | P2 | Done | LLM workloads/usage/traces APIs + UI |

### 7.10 Integrations & admin (`REQ-INT-*`, `REQ-ADM-*`)

| ID | SRS | Priority | Status | What to build / verify |
|----|-----|----------|--------|-------------------------|
| INT-01 | REQ-INT-001–003 | P1 | Done | GitHub/GitLab/Jenkins webhooks + integration seeds |
| INT-02 | REQ-INT-011–014 | P1 | Done | Export jobs + `GET /exports/jobs` |
| INT-03 | REQ-INT-021–022 | P1 | Done | Terraform provider + `cmd/nexobs` CLI |
| ADM-01 | REQ-ADM-003 | P0 | Done | ABAC PDP + CI scope tests |
| ADM-02 | REQ-ADM-007–009 | P0 | Done | `GET/PUT /admin/signal-policies` per signal |
| ADM-03 | REQ-ADM-011–012 | P2 | Done | MSP tenant API + Terraform + create tenant UI |

### 7.11 Non-functional (`REQ-NFR-*`)

| ID | SRS | Priority | Status | What to build / verify |
|----|-----|----------|--------|-------------------------|
| NFR-01 | Availability SLO | P0 | Partial | 99.9% gateway — staging soak 7d |
| NFR-02 | Search p95 &lt; 2s | P0 | Done | `bank-search-90d.js` + `verify-search-90d.sh` |
| NFR-03 | Ingest scale | P0 | Partial | `scripts/k6/bank-ingest-load.js` |
| NFR-04 | Security CI | P0 | Done | gitleaks, mTLS script, ABAC tests |
| NFR-05 | Coverage ≥ 80% | P1 | Partial | Raise gates beyond core packages |

---

## 8. FinOps — production path for banks

**Code status:** `REQ-FINOPS-001..073` implemented (see [FINOPS_ENHANCEMENT_REQUIREMENTS.md](./FINOPS_ENHANCEMENT_REQUIREMENTS.md)).  
**Bank prod status:** **Partial** until live billing and finance sign-off.

| ID | Requirement | Status | Action |
|----|-------------|--------|--------|
| FIN-PROD-01 | Real AWS CUR connector (not simulation) | Done | `GET /finops/billing/sources`, Helm CUR sync, `verify-finops-live.sh` |
| FIN-PROD-02 | GCP BigQuery billing export | Done | `FINOPS_GCP_BILLING_FILE` + sample NDJSON |
| FIN-PROD-03 | Azure Cost Management export | Done | `FINOPS_AZURE_BILLING_FILE` + sample NDJSON |
| FIN-PROD-04 | Reconciliation &lt; 1% vs invoice | Partial | Validate per provider in prod tenant |
| FIN-PROD-05 | Postgres-only prod store (no memory fallback) | Done | `FINOPS_REQUIRE_POSTGRES` strict line-item path in `PostgresStore` |
| FIN-PROD-06 | Attribution ≥ 95% | Partial | Tag rules + ML suggestions + bank CMDB join |
| FIN-PROD-07 | Anomaly precision ≥ 0.85 | Partial | Labeled validation set per bank |
| FIN-PROD-08 | Iceberg/lake for line items (SRS tiering) | Partial | `GET /finops/line-items/lake/export` + `FINOPS_LAKE_BUCKET` wiring |
| FIN-PROD-09 | Finance audit: who changed budgets/rules | Done | `GET /finops/audit/export`, `scripts/verify-finops-production.sh` |

**Sales rule:** Do not sell FinOps to bank finance team until **FIN-PROD-01..04** are green for their cloud accounts.

---

## 9. Master implementation waves (do in order)

Estimated durations assume 2–4 engineers + 0.5 SRE + security consultant on critical waves.

### Wave 0 — Sales & legal foundation (2 weeks)

**Goal:** Can legally and safely start bank conversations.  
**Status:** ✅ Artifacts created — legal review pending before customer send.

| # | Item | IDs | Status |
|---|------|-----|--------|
| 0.1 | Security pack (architecture, data flow, encryption, audit) | BANK-001, BANK-002 | ✅ [docs/bank/SECURITY_PACK.md](./bank/SECURITY_PACK.md), [DATA_FLOW.md](./bank/DATA_FLOW.md) |
| 0.2 | DPA + subprocessor draft | BANK-003 | ✅ [DPA_TEMPLATE.md](./bank/DPA_TEMPLATE.md), [SUBPROCESSORS.md](./bank/SUBPROCESSORS.md) |
| 0.3 | MSA/ELA template for self-hosted | BANK-004 | ✅ [MSA_ELA_TEMPLATE.md](./bank/MSA_ELA_TEMPLATE.md) |
| 0.4 | POC scope template | — | ✅ [POC_SCOPE_TEMPLATE.md](./bank/POC_SCOPE_TEMPLATE.md) |
| 0.5 | Production auth checklist + gateway guard + verify script | PROD-AUTH-01 | ✅ [PRODUCTION_AUTH_CHECKLIST.md](./bank/PRODUCTION_AUTH_CHECKLIST.md), `config.ValidateProduction`, `scripts/verify-production-auth.sh` |

**Exit:** Legal + security review **internal** sign-off to send pack to banks.

---

### Wave 1 — Production baseline (4–6 weeks)

**Goal:** One **staging** environment on K8s + managed data plane you trust.  
**Status:** ✅ Repo artifacts complete — execute on customer staging/K8s.

| # | Item | IDs | Status |
|---|------|-----|--------|
| 1.1 | `values-staging.yaml` / `values-prod.yaml` Helm; External Secrets | PRODUCTION_AND_GTM §4 | ✅ |
| 1.2 | SSO only (OIDC/SAML); rotate API keys | BANK-007, PROD-AUTH-01 | ✅ |
| 1.3 | East-west mTLS or service mesh | PROD-NET-01 | ✅ |
| 1.4 | Managed Postgres/Kafka/ES/CH/Redis | PROD-INFRA-01 | ✅ |
| 1.5 | Backup + restore drill documented | BANK-009, BANK-010 | ✅ |
| 1.6 | Penetration test scheduled/completed | BANK-005 | ⬜ vendor engagement |
| 1.7 | Tenant isolation automated tests in CI | BANK-006, ADM-01 | ✅ |
| 1.8 | PII scrubbing guide + default policies | BANK-008, COLL-08 | ✅ |
| 1.9 | 7-day staging soak without manual fix | NFR-01 | 🟡 script ready — run 7d |
| 1.10 | Customer LLM key path documented | BANK-016 | ✅ |

**Exit:** [Gate B — Staging production](#gate-b--staging-production) passed.

---

### Wave 2 — Bank POC package (4 weeks, overlaps Wave 1 tail)

**Goal:** Repeatable 4-week bank POC.  
**Status:** ✅ Artifacts created — execute load test + ITSM OAuth on customer staging.

| # | Item | IDs | Status |
|---|------|-----|--------|
| 2.1 | Payments KPI pack + transaction journey hardening | BANK-013, APM-07 | ✅ `kpi-bfsi-payments` seed + [PAYMENTS_KPI_PACK.md](./bank/PAYMENTS_KPI_PACK.md) |
| 2.2 | Ingest runbook for Fluent Bit / OTEL / Kafka | LOG-01 | ✅ [BANK_LOG_INGEST.md](./runbooks/BANK_LOG_INGEST.md) |
| 2.3 | Alert → ServiceNow/Jira integration tested | INT-01 | ✅ Workflow `servicenow` step + [ITSM_INTEGRATION_JIRA_SERVICENOW.md](./runbooks/ITSM_INTEGRATION_JIRA_SERVICENOW.md), `scripts/verify-itsm-integration.sh` |
| 2.4 | Load test at POC volume (logs + traces) | NFR-02, NFR-03 | ✅ `bank-poc-load.js`, `bank-ingest-load.js` in CI |
| 2.5 | Incident + AI RCA playbook for champions | PRODUCTION_AND_GTM §11 | ✅ [INCIDENT_RCA_PLAYBOOK.md](./runbooks/INCIDENT_RCA_PLAYBOOK.md) |
| 2.6 | ABAC scope tests for `payments` / `platform` services | ADM-01 | ✅ `abac_scope_test.go`, `developer_scope_test.go` |
| 2.7 | Air-gapped deployment doc | BANK-015 | ✅ [AIR_GAPPED_DEPLOYMENT.md](./bank/AIR_GAPPED_DEPLOYMENT.md) |

**Exit:** [Gate C — Bank POC ready](#gate-c--bank-poc-ready) passed.

---

### Wave 3 — Self-hosted enterprise hardening (6–8 weeks)

**Goal:** Close first **self-hosted** bank deal.  
**Status:** ✅ Docs + Terraform SLO/ABAC — pen test, signed MSA, customer Helm still required.

| # | Item | IDs | Status |
|---|------|-----|--------|
| 3.1 | Offline license / air-gap install guide | BANK-018 | ✅ [AIR_GAP_INSTALL.md](./bank/AIR_GAP_INSTALL.md), gateway `LicenseEnforcement`, Helm `license.*` |
| 3.2 | Helm install PS playbook (5–10 days) | PRODUCTION_AND_GTM §13 | ✅ [HELM_PS_PLAYBOOK.md](./bank/HELM_PS_PLAYBOOK.md) |
| 3.3 | SOC 2 Type I policy set | BANK-011 | ✅ [SOC2_TYPE1_POLICY_INDEX.md](./compliance/SOC2_TYPE1_POLICY_INDEX.md) |
| 3.4 | FFIEC control mapping worksheet (optional) | BANK-017 | ✅ [FFIEC_CONTROL_MAPPING.md](./compliance/FFIEC_CONTROL_MAPPING.md) |
| 3.5 | Support escalation + 24×7 option | BANK-019, BANK-020 | ✅ [SUPPORT_ESCALATION.md](./bank/SUPPORT_ESCALATION.md) |
| 3.6 | Log live tail OR committed roadmap date | LOG-04 | ✅ `WS /logs/tail/ws` + UI |
| 3.7 | Expand Terraform resources (alerts, SLOs, tenants) | INT-03 | ✅ provider resources shipped |

**Exit:** [Gate D — First bank production](#gate-d--first-bank-production) passed.

---

### Wave 4 — Scale & compliance depth (8–12 weeks)

**Status:** 🟡 Foundation docs + CI k6/coverage — Type II audit and 80% coverage are ongoing.

| # | Item | IDs | Status |
|---|------|-----|--------|
| 4.1 | SOC 2 Type II audit period start | BANK-012 | 🟡 [SOC2_TYPE2_READINESS.md](./compliance/SOC2_TYPE2_READINESS.md) |
| 4.2 | Multi-region residency if required | ADM-02 | ✅ [MULTI_REGION_RESIDENCY.md](./runbooks/MULTI_REGION_RESIDENCY.md) |
| 4.3 | k6 CI gate on critical APIs | NFR-02 | ✅ compose-smoke CI |
| 4.4 | Coverage gate ≥ 80% repo-wide | NFR-05 | 🟡 [COVERAGE_ROADMAP.md](./compliance/COVERAGE_ROADMAP.md) |
| 4.5 | NexQL / unified query depth | MET-04 | ✅ |
| 4.6 | SCA + container scan ingestion | SEC-01 | ✅ Trivy in CI |

**Exit:** [Gate E — Enterprise scale claims](#gate-e--enterprise-scale-claims) passed.

---

### Wave 5 — FinOps production (optional SKU, 4–6 weeks)

**Status:** 🟡 Live file ingest + PG persistence + APIs — customer CUR/S3 wiring required.

| # | Item | IDs | Status |
|---|------|-----|--------|
| 5.1 | Live CUR/BQ/Azure billing | FIN-PROD-01..03 | 🟡 `FINOPS_*_FILE`, `FINOPS_BILLING_MODE=live` — [FINOPS_PRODUCTION.md](./bank/FINOPS_PRODUCTION.md) |
| 5.2 | Invoice reconciliation proof | FIN-PROD-04 | 🟡 `GET /finops/reconciliation`, `FINOPS_AWS_INVOICE_USD` |
| 5.3 | Prod persistence only | FIN-PROD-05 | ✅ `FINOPS_REQUIRE_POSTGRES` strict store (no memory fallback for line items) |
| 5.4 | Finance audit export | FIN-PROD-09 | ✅ |

**Exit:** Finance stakeholder sign-off + [Gate F — FinOps prod](#gate-f--finops-prod) passed.

---

### Wave 6 — Differentiation (ongoing)

**Status:** 🟡 Bank-safe defaults + Helm CUR sync + preview APIs.

| # | Item | IDs | Status |
|---|------|-----|--------|
| 6.1 | AutoFix with approval gates | AI-02 | ✅ `AUTOFIX_*` env, handlers, [AUTOFIX_BANK_POLICY.md](./bank/AUTOFIX_BANK_POLICY.md) |
| 6.2 | Full NPM suite | NPM-* | ✅ netflow ingest + APIs |
| 6.3 | MSP / white-label | ADM-03 | ✅ |
| 6.4 | LLM observability | AI-03 | ✅ |
| 6.5 | CUR S3 → PVC Helm sync | FIN-PROD-01 | ✅ [FINOPS_CUR_S3_SYNC.md](./runbooks/FINOPS_CUR_S3_SYNC.md) |

---

## 10. Go / no-go gates

### Gate A — Demo only (today)

- [x] Compose quickstart works
- [x] Banking demo data (transactions, payments services)
- [ ] **Do not** claim SOC 2, pen test, or prod SLA

### Gate B — Staging production

- [ ] K8s + managed data stores
- [ ] SSO enforced; dev login off
- [ ] Backups + one restore drill
- [ ] 7-day soak
- [ ] Internal security checklist ≥ 80% (PRODUCTION_AND_GTM §15)

### Gate C — Bank POC ready

- [ ] Gate B passed
- [ ] Security pack sent to bank
- [ ] POC success metrics agreed in writing
- [ ] Load test ≥ POC volume
- [ ] Tenant isolation test passed

### Gate D — First bank production

- [ ] Gate C passed
- [ ] Pen test remediated (no open critical)
- [ ] Signed DPA + MSA
- [ ] Helm deploy on customer K8s
- [ ] Customer SSO live
- [ ] Support/on-call defined
- [ ] Rollback drill &lt; 15 min

### Gate E — Enterprise scale claims

- [ ] Gate D + 2+ bank/prod customers
- [ ] SOC 2 Type I complete; Type II in progress
- [ ] Documented SLOs met 3 months
- [ ] k6 P95 gates on search + finops breakdown

### Gate F — FinOps prod

- [ ] FIN-PROD-01..04 green
- [ ] Not sold on simulated ingest

---

## 11. Commercial packaging for banks

| SKU | When to use | Requirements |
|-----|-------------|--------------|
| **Enterprise Self-hosted** | Primary for banks | Gate D; Helm; customer K8s |
| **Dedicated Cloud** | Bank accepts vendor cloud | Isolated VPC; Gate D minus Helm |
| **Multi-tenant Cloud** | Only after Gate E | Not recommended for first bank |

**Pricing reference:** [PRODUCTION_AND_GTM.md §12](./PRODUCTION_AND_GTM.md) — self-hosted **$100k–$250k/year** guideline.

**Honest messaging**

- ✅ “AI-native observability + incident intelligence for payments platforms”
- ✅ “Self-hosted on your Kubernetes with your IdP”
- ❌ “Full Datadog/New Relic replacement day one”
- ❌ “SOC 2 certified” (until audit complete)
- ❌ “FinOps with audited invoice reconciliation” (until Gate F)

---

## 11.1 Gate close-out runbooks

| Gate | Runbook | Verify script |
|------|---------|---------------|
| B, C | [GATE_BC_CLOSEOUT_RUNBOOK.md](./bank/GATE_BC_CLOSEOUT_RUNBOOK.md) | `./scripts/gate-verify.sh --gate B\|C` |
| D | [GATE_D_CLOSEOUT_RUNBOOK.md](./bank/GATE_D_CLOSEOUT_RUNBOOK.md) | `./scripts/gate-verify.sh --gate D` |
| E, F | [GATE_EF_CLOSEOUT_RUNBOOK.md](./bank/GATE_EF_CLOSEOUT_RUNBOOK.md) | `./scripts/gate-verify.sh --gate E\|F` |

```bash
./scripts/sync-openapi.sh && make contract-test openapi-test
./scripts/gate-verify.sh --gate C   # POC ready
./scripts/gate-verify.sh --gate D   # first bank production
```

Supporting: [DPA_MSA_SIGNOFF_TRACKER.md](./bank/DPA_MSA_SIGNOFF_TRACKER.md), [BANK_PRODUCTION_SECURITY_CHECKLIST.md](./bank/BANK_PRODUCTION_SECURITY_CHECKLIST.md), [HELM_ROLLBACK_DRILL.md](./runbooks/HELM_ROLLBACK_DRILL.md)

---

## 12. Tracking template

Copy into your issue tracker (one epic per wave).

```markdown
## [WAVE-X] Title
- **Owner:**
- **Target date:**
- **Gate:**

| Item | Status | PR/link | Notes |
|------|--------|---------|-------|
| X.1  | ⬜ Todo / 🟡 Partial / ✅ Done | | |
```

**Status values:** `Todo` | `In Progress` | `Done` | `Won't fix (bank MVP)`

---

## Document maintenance

| Trigger | Action |
|---------|--------|
| Item completed | Mark Done in Section 12; update Section 6 status |
| Bank deal signed | Record which gates passed; narrow scope in contract |
| SOC 2 milestone | Update BANK-011/012 and Gate E |
| FinOps live billing | Update Section 8 + Gate F |
| All repo open points | See [bank/OPEN_POINTS_CLOSURE.md](./bank/OPEN_POINTS_CLOSURE.md) |

**Owner:** Product + Platform Engineering  
**Review:** Bi-weekly until Gate D; monthly until Gate E.

---

*This is the single end-to-end checklist to take NeuralOps from current codebase to bank production and commercial sale. Implement waves sequentially; do not skip Wave 0–1 for regulated buyers.*
