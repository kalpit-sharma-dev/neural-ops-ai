# POC Load Test Profile (Wave 2.4)

**IDs:** NFR-02, NFR-03

## Default POC volume

| Signal | Rate | Duration |
|--------|------|----------|
| Log search API | 20 RPS | 5 min |
| FinOps breakdown | 10 RPS | 30 s |
| Trace/metrics ingest | Customer-specific | 1 h soak |

Adjust `VU` and `duration` to match signed POC scope ([POC_SCOPE_TEMPLATE.md](./POC_SCOPE_TEMPLATE.md)).

## Prerequisites

- Gateway reachable: `BASE_URL=https://staging.observe.customer.internal`
- Auth token with tenant header
- k6 installed (`go install go.k6.io/k6@latest` or CI image)

## Run

```bash
export BASE_URL=https://staging.observe.customer.internal
export K6_TOKEN=$POC_TOKEN
export K6_TENANT=poc-bank

k6 run scripts/k6/bank-poc-load.js
```

## Pass criteria

| Check | Threshold |
|-------|-----------|
| `http_req_duration` p95 | &lt; 2000 ms (search + finops) |
| HTTP 5xx rate | &lt; 0.1% |
| Ingest lag (Kafka consumer) | &lt; 60 s p95 during soak |

## CI (optional)

Add to `.github/workflows/ci.yml` as manual `workflow_dispatch` against staging — do not gate PRs on staging URL until environment exists.

## Evidence for Gate C

Attach k6 summary JSON + Grafana dashboard export to POC close-out deck.
