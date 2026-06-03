# Support Escalation & 24×7 Option (Wave 3.5)

**IDs:** BANK-019, BANK-020

## Severity definitions

| Sev | Example | Target response | Target update |
|-----|---------|-----------------|----------------|
| S1 | Platform down, no ingest | 15 min | 30 min |
| S2 | Degraded search, POC blocked | 1 h | 2 h |
| S3 | Non-critical defect | 1 business day | Daily |
| S4 | How-to / feature request | 2 business days | Weekly |

*24×7 option:* contractual amendment — P1 only, named customer tenants.

## Escalation path

```
L1 Customer NOC / champion
  → L2 NeuralOps support@neuralops.ai (ticket ID required)
    → L3 Engineering on-call (S1/S2 only, page via PagerDuty)
      → L4 Engineering leadership + account team
```

## Required ticket fields

- Tenant ID / environment URL
- Time zone + incident start (UTC)
- Trace ID or `txnId` sample
- Recent change (deploy, Helm, policy)
- Attach k6 summary or `/health` output if perf-related

## Status page

- **Dedicated customers:** optional private status (RSS/email) — customer-hosted.
- **SaaS:** publish component list matching gateway, ingestion, search, alerting.

## Data handling for support

- No production credentials in tickets.
- Use time-bounded **support read-only** SSO role where available.
- Air-gap: customer uploads log bundle via secure file exchange — see [AIR_GAPPED_DEPLOYMENT.md](./AIR_GAPPED_DEPLOYMENT.md).

## Commercial

Enterprise ELA should reference severity table and maintenance windows (e.g. Sun 02:00–06:00 local, pre-approved).
