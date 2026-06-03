# NexQL & Unified Query Workbench (Wave 4.5)

**ID:** MET-04  
**SRS:** REQ-MET-006, REQ-MET-004

## Vision

Single workbench for logs, metrics, traces, and FinOps — planner selects optimal store (ES, ClickHouse, Prometheus) with cardinality guards.

## Current (partial)

- PromQL metrics APIs
- Log search + analytics endpoints
- Query planner stubs in `backend/internal/observability/query_planner.go`
- Cardinality limits in `query_cardinality.go`

## Roadmap

| Milestone | Target | Deliverable |
|-----------|--------|-------------|
| M1 | Q3 2026 | NexQL parser + explain API |
| M2 | Q4 2026 | Cross-signal join (trace_id, txn_id) |
| M3 | Q1 2027 | UI workbench with saved queries |

## Bank POC interim

Use separate tabs (Logs, Metrics, Traces) with shared `traceId` / `txnId` filters — documented in [INCIDENT_RCA_PLAYBOOK.md](../runbooks/INCIDENT_RCA_PLAYBOOK.md).

## Sales language

Do not claim full NexQL parity until M2 GA.
