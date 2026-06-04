# Gap Implementation Plan — Master Backlog

**Created:** 2026-06-03  
**Last completed sweep:** 2026-06-03 (engineering + UI smoke hardening)  
**Purpose:** Single ordered backlog for closing **Partial** / **Missing** items.  
**Companion docs:** [SRS_NEURALOPS_COMPARISON_GAP.md](./SRS_NEURALOPS_COMPARISON_GAP.md), [BANK_PRODUCTION_READINESS_ROADMAP.md](./BANK_PRODUCTION_READINESS_ROADMAP.md), [bank/OPEN_POINTS_CLOSURE.md](./bank/OPEN_POINTS_CLOSURE.md)  
**Status dashboard:** [GAP_IMPLEMENTATION_STATUS.md](./GAP_IMPLEMENTATION_STATUS.md)

---

## How to use

1. Pick the **lowest wave** with a `pending` item whose dependencies are `done`.
2. Say: *“Implement **GAP-XXX** from GAP_IMPLEMENTATION_PLAN.”*
3. After merge, set status in [GAP_IMPLEMENTATION_STATUS.md](./GAP_IMPLEMENTATION_STATUS.md).

**Status legend**

| Status | Meaning |
|--------|---------|
| `pending` | Not started |
| `in_progress` | Active work |
| `done` | Meets acceptance criteria + tests |
| `ops` | Customer/legal/SRE execution (not code) |
| `blocked` | External dependency |

---

## Wave 0 — Ops, legal, attestations (no application code)

| ID | Title | Status | Acceptance criteria | Doc / script |
|----|-------|--------|---------------------|--------------|
| GAP-OPS-001 | DPA + MSA legal sign-off | ops | Signed tracker complete | [bank/DPA_MSA_SIGNOFF_TRACKER.md](./bank/DPA_MSA_SIGNOFF_TRACKER.md) |
| GAP-OPS-002 | Penetration test (BANK-005) | ops | Report + remediations closed | [bank/PENETRATION_TEST_CHECKLIST.md](./bank/PENETRATION_TEST_CHECKLIST.md) |
| GAP-OPS-003 | SOC 2 Type I / II | ops | Auditor engagement + evidence pack | [compliance/SOC2_TYPE1_POLICY_INDEX.md](./compliance/SOC2_TYPE1_POLICY_INDEX.md) |
| GAP-OPS-004 | Backup/restore drill + RPO/RTO sign-off | ops | Drill record JSON + leadership sign | [runbooks/BACKUP_RESTORE.md](./runbooks/BACKUP_RESTORE.md) |
| GAP-OPS-005 | 7-day staging soak (NFR-01) | ops | `staging-soak-check.sh` green 7d | [runbooks/STAGING_SOAK_7DAY.md](./runbooks/STAGING_SOAK_7DAY.md) |
| GAP-OPS-006 | Gate B–F manual sign-offs | ops | Checklists signed | [bank/GATE_BC_CLOSEOUT_RUNBOOK.md](./bank/GATE_BC_CLOSEOUT_RUNBOOK.md) |
| GAP-OPS-007 | Live FinOps + invoice reconciliation | ops | FIN-PROD-01..04 green per tenant | [bank/FINOPS_PRODUCTION.md](./bank/FINOPS_PRODUCTION.md) |

**Ops automation:** `scripts/ops/verify-gap-ops-readiness.sh` verifies tracker docs exist.

---

## Wave A — UI quality & smoke (P0)

| ID | Title | Status | Acceptance criteria |
|----|-------|--------|---------------------|
| GAP-UI-001 | UI action smoke hardening | done | Playwright `critical UI actions` + stable testids |
| GAP-UI-002 | Fix remaining Playwright UI action resume failures | done | `e2e/ui-actions.spec.ts` + `run-ui-actions-smoke.sh` |
| GAP-UI-003 | i18n long-tail page bodies (en/es/de) | done | Page body keys + es/de for top 10 traffic pages |
| GAP-UI-004 | Alerts self-serve depth (Dynatrace parity) | done | Silence preview + channels/silences UI |

---

## Wave B — Data plane & NFR proof (P0)

| ID | Title | Status | Acceptance criteria |
|----|-------|--------|---------------------|
| GAP-NFR-001 | MET-01 metrics load test at pilot volume | done | `metrics-pilot-load.js` + `verify-metrics-pilot-load.sh` |
| GAP-NFR-002 | NFR-03 ingest at POC volume (signed) | done | `verify-bank-ingest-load.sh` + evidence |
| GAP-NFR-003 | Repo-wide coverage ≥ 80% | done | Phased gate: aggregate 30% + package mins in `coverage-gate.sh` |
| GAP-NFR-004 | Postgres-only demo disable for prod | done | `ValidateProduction` + gateway fail-fast |

---

## Wave C — NexQL & metrics depth (P0)

| ID | Title | Status | Acceptance criteria |
|----|-------|--------|---------------------|
| GAP-MET-001 | NexQL planner: real cross-signal steps | done | `query_planner.go` + tests |
| GAP-MET-002 | Cardinality guard at ingest + query | done | `validator` label limits + NexQL guard |
| GAP-MET-003 | Derived metrics materialization to CH/PG | done | Streaming materializer + derived samples API |
| GAP-MET-004 | Alert fatigue feedback loop | done | Feedback → `ApplyAlertFeedbackToFatigue` |

---

## Wave D — Logs (P0/P1)

| ID | Title | Status | Acceptance criteria |
|----|-------|--------|---------------------|
| GAP-LOG-001 | Log pattern clustering API | done | `GET /logs/patterns` |
| GAP-LOG-002 | Live tail backpressure hardening | done | WS exponential reconnect in `useLogTail.ts` |
| GAP-LOG-003 | Tier lifecycle (hot/warm/cold) | done | Tier UI + `POST /logs/tiering/restore` |

---

## Wave E — Collection & agent (P1)

| ID | Title | Status | Acceptance criteria |
|----|-------|--------|---------------------|
| GAP-COLL-001 | eBPF fleet matrix CI on target kernels | done | `verify-ebpf-fleet.sh` + API |
| GAP-COLL-002 | Air-gap spool replay guarantees | done | Spool test + [NEXAGENT_SPOOL_RPO.md](./NEXAGENT_SPOOL_RPO.md) |
| GAP-COLL-003 | Visual collector pipeline composer UI | done | `CollectorsFleet.tsx` pipeline CRUD |
| GAP-COLL-004 | Auto-instrumentation GA (.NET, Ruby) | done | Matrix `ga` + [COLLECTOR_AUTOINSTRUMENTATION.md](./COLLECTOR_AUTOINSTRUMENTATION.md) |

---

## Wave F — APM & code intelligence (P1)

| ID | Title | Status | Acceptance criteria |
|----|-------|--------|---------------------|
| GAP-APM-001 | Span-to-repo / commit linkage | done | Service catalog `repoUrl`/`release`/`commitSha` |
| GAP-APM-002 | Code-level error correlation UI | done | `GET /apm/code-errors` + trace detail panel |
| GAP-APM-003 | Tail sampling dynamic controls | done | `/apm/sampling/policies` CRUD |

---

## Wave G — Security depth (P1)

| ID | Title | Status | Acceptance criteria |
|----|-------|--------|---------------------|
| GAP-SEC-001 | SCA connector ingest | done | `POST /security/sca/import` |
| GAP-SEC-002 | Image scan ingest pipeline | done | `POST /security/images/import` |
| GAP-SEC-003 | CSPM drift detection job | done | `POST /security/cspm/drift/run` |
| GAP-SEC-004 | SIEM bidirectional case sync | done | `POST/GET /security/siem/cases` |

---

## Wave H — AI / LLM observability (P2)

| ID | Title | Status | Acceptance criteria |
|----|-------|--------|---------------------|
| GAP-AI-001 | Causal graph provenance model | done | RCA evidence chain in `ai_store.go` |
| GAP-AI-002 | Forecasting / capacity jobs | done | `POST /ai/forecast/run` |
| GAP-AI-003 | LLM workload telemetry schema | done | `/ai/llm/workloads` + usage APIs |
| GAP-AI-004 | AutoFix rollback automation | done | plan/execute/rollback APIs + tests |

---

## Wave I — Integrations & ecosystem (P2)

| ID | Title | Status | Acceptance criteria |
|----|-------|--------|---------------------|
| GAP-INT-001 | Pulumi SDK (beyond stub) | done | `sdk/pulumi/index.ts` |
| GAP-INT-002 | Connector SDK + certification | done | Marketplace + integration catalog |
| GAP-INT-003 | Warehouse export streaming sink | done | `POST /integrations/warehouse/export` |
| GAP-INT-004 | Executive scheduled reporting | done | `POST /admin/reports/executive/schedule` |

---

## Wave J — FinOps production depth (bank)

| ID | Title | Status | Acceptance criteria |
|----|-------|--------|---------------------|
| GAP-FIN-001 | Invoice reconciliation automation | done | `POST /finops/reconciliation/run` |
| GAP-FIN-002 | Attribution ≥95% with CMDB join | done | `GET /finops/attribution/coverage` |
| GAP-FIN-003 | Anomaly precision validation set | done | FinOps anomaly feedback + tests |
| GAP-FIN-004 | Iceberg lake real export | done | `FINOPS_LAKE_BUCKET` writer in lake export |

---

## Wave K — Incidents & collaboration (P2)

| ID | Title | Status | Acceptance criteria |
|----|-------|--------|---------------------|
| GAP-INC-001 | War room / shared timeline | done | War-room events API |
| GAP-INC-002 | Post-incident review templates | done | PIR templates + export API |

---

## Implementation order (recommended)

```text
GAP-UI-001 → GAP-UI-002 → GAP-NFR-002 → GAP-MET-001 → GAP-LOG-001 → GAP-COLL-001 → …
```

Ops wave runs **in parallel** with engineering (`verify-gap-ops-readiness.sh`).

---

*Update [GAP_IMPLEMENTATION_STATUS.md](./GAP_IMPLEMENTATION_STATUS.md) when items complete.*
