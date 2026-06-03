# Live Log Tail — Product Commitment (Wave 3.6)

**ID:** LOG-04  
**SRS:** REQ-LOG-011

## Commitment

| Milestone | Target | Deliverable |
|-----------|--------|-------------|
| Design complete | Q3 2026 | WebSocket API spec + backpressure model |
| Private preview | Q4 2026 | Tenant-scoped tail behind feature flag |
| GA (enterprise) | Q1 2027 | UI live tail + RBAC/ABAC on `logs:tail` |

## Interim POC workaround

- Near-real-time: log search with 5–15 s refresh + saved queries
- Export job for forensic slices — `POST /api/v1/exports/logs`
- Fluent Bit forward to customer SIEM for true streaming

## Technical notes (planned)

- Gateway WebSocket `/api/v1/logs/tail`
- Max 500 concurrent tails per tenant; rate limit per user
- Residency: tail only from region-permitted indices

## Sales language

Until GA: **do not** claim parity with Datadog/Splunk live tail. Use committed dates above in enterprise roadmap appendix.
