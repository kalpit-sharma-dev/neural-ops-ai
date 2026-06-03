# BFSI Payments KPI Pack (Wave 2.1)

**IDs:** BANK-013, APM-07  
**API:** `GET /api/v1/business/kpi-packs`, `POST /api/v1/business/kpi-packs/kpi-bfsi-payments/enable`

## Purpose

Pre-built business observability for bank POCs: payment success, UPI latency, failed-txn rate, and chargebacks — aligned to seeded transaction journeys (`ledger-service`, `payment-api`, `upi-service`).

## Enable for POC tenant

```bash
curl -s -X POST "$GATEWAY/api/v1/business/kpi-packs/kpi-bfsi-payments/enable" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT"
```

## UI

1. **Observability → Business observability** — enable **BFSI payments KPI pack**.
2. **Incidents → Transaction search** — correlate `txnId` with traces/logs (seed: `UPI-FAIL-*`, `TXN-DEMO-*`).
3. **APM → Service map** — filter `payments`, `upi-service`, `ledger-service`.

## POC success metrics (suggested)

| KPI | Target | Source |
|-----|--------|--------|
| Payment success rate | ≥ 99.5% | KPI pack + txn aggregate |
| UPI p95 | &lt; 800 ms | APM + synthetic checkout |
| Failed txn rate during incident | Detect &lt; 5 min | Alerts on `ledger-service` / `payment-api` |

## Instrumentation checklist

- [ ] OTEL service names match CMDB: `payment-api`, `upi-service`, `ledger-service`.
- [ ] Log field `txn_id` (or `txnId`) on payment path.
- [ ] Business transaction ID propagated on HTTP headers (`X-Txn-Id`) for trace correlation.
- [ ] RUM funnel optional for digital channels (see `fn-checkout` seed funnel).

## Demo data

Run seed/load so ClickHouse has UPI outage transactions:

```bash
cd backend && go run ./scripts/seed/...   # or compose seed profile per README
```

## Related

- [POC_SCOPE_TEMPLATE.md](./POC_SCOPE_TEMPLATE.md)
- [INCIDENT_RCA_PLAYBOOK.md](../runbooks/INCIDENT_RCA_PLAYBOOK.md)
- [BANK_PRODUCTION_READINESS_ROADMAP.md](../BANK_PRODUCTION_READINESS_ROADMAP.md) Wave 2.1
