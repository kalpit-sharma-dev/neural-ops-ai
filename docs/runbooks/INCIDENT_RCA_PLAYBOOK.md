# Incident & AI RCA Playbook — Bank POC Champions (Wave 2.5)

**Related:** PRODUCTION_AND_GTM §11, BANK-013

## Roles

| Role | Responsibility |
|------|----------------|
| Champion (bank SRE) | Drives timeline, validates business impact |
| NeuralOps PS | Platform config, workflow tuning |
| Comms | Status updates per bank policy |

## 1. Detect (0–5 min)

1. Alert fires (example: `ledger-service` error rate or FinOps cost anomaly on `payments` scope).
2. Open **Alerts** → acknowledge; confirm severity and service owner.
3. Check **SLO burn** if configured for payment path.

## 2. Triage (5–15 min)

1. **Transaction search:** `POST /api/v1/search/transactions` with time window + `status:FAILED`.
2. Pick sample `txnId` (e.g. `UPI-FAIL-000042`).
3. **Incident context:** `GET /api/v1/transactions/{txnId}` — review hop-level failure (`upi-service` vs `payment-api`).
4. **Logs:** filter `service:ledger-service` and `txn_id` / trace ID.
5. **Traces:** service map → critical path latency regression.

## 3. Correlate (15–30 min)

| Signal | Question |
|--------|----------|
| Deploy | Recent release on failing service? |
| Infra | RDS/K8s node pressure on payments subnet? |
| Upstream | NPCI/UPI gateway errors in logs? |
| FinOps | Cost spike on same scope as errors? |

Use **AI Ops → RCA** (if LLM keys configured per [CUSTOMER_LLM_KEYS.md](../bank/CUSTOMER_LLM_KEYS.md)):

- Require human approval before any AutoFix action.
- Export RCA summary to ticket (Jira/ServiceNow per [ITSM_INTEGRATION_JIRA_SERVICENOW.md](./ITSM_INTEGRATION_JIRA_SERVICENOW.md)).

## 4. Mitigate & communicate

1. Execute runbook link from alert policy (auto-context).
2. Page on-call via PagerDuty step if P1.
3. Customer status: internal bridge → executive summary template (bank-owned).

## 5. Resolve & learn (post-incident)

- [ ] Root cause category recorded (code, config, dependency, capacity)
- [ ] Failed txn rate returned below KPI target (see [PAYMENTS_KPI_PACK.md](../bank/PAYMENTS_KPI_PACK.md))
- [ ] Blameless review within 5 business days
- [ ] Action items: alert threshold, synthetic test, tag governance

## POC demo script (30 min)

1. Show pre-seeded UPI outage transactions in UI.
2. Walk one failed txn through hops → logs → trace.
3. Fire synthetic alert; show Jira ticket creation.
4. Show AI RCA narrative with citations (logs/traces), no auto-remediation.

## Escalation

See [SUPPORT_ESCALATION.md](../bank/SUPPORT_ESCALATION.md) for Severity → NeuralOps response targets.
