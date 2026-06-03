# FinOps Production Enablement (Wave 5)

**IDs:** FIN-PROD-01..05, FIN-PROD-09  
**Prerequisite:** [Gate D](../BANK_PRODUCTION_READINESS_ROADMAP.md) for platform; finance sign-off for **Gate F**.

## Environment variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `FINOPS_BILLING_MODE` | `simulated`, `live`, `hybrid` | `live` |
| `FINOPS_REQUIRE_POSTGRES` | Disable memory-only prod (`FIN-PROD-05`) | `true` |
| `FINOPS_ALLOW_SIMULATION` | Hybrid fallback when file missing | `false` in prod |
| `FINOPS_AWS_CUR_FILE` | NDJSON/JSON CUR export path (FIN-PROD-01) | `/data/cur/june.ndjson` |
| `FINOPS_AWS_INVOICE_USD` | Invoice total for reconciliation (FIN-PROD-04) | `1250000.00` |
| `FINOPS_GCP_BILLING_FILE` | BigQuery export NDJSON (FIN-PROD-02) | `/data/gcp/billing.ndjson` |
| `FINOPS_AZURE_BILLING_FILE` | Cost Management export (FIN-PROD-03) | `/data/azure/billing.ndjson` |
| `NEURALOPS_CLOUD_LIVE` | Live inventory for simulation path | `true` |

## Billing file format (NDJSON)

```json
{"provider":"aws","accountId":"111122223333","service":"AmazonEC2","region":"us-east-1","resourceId":"i-abc","amortizedCost":420.5,"billingPeriod":"2026-06-01","tags":{"team":"payments"}}
```

Sample: `backend/internal/finops/connectors/testdata/aws_cur_sample.ndjson`

## Helm (gateway / observability)

```yaml
config:
  finopsBillingMode: live
  finopsRequirePostgres: true
```

Mount CUR files via PVC or sync from S3 (customer job).

## APIs

| Endpoint | FIN-PROD |
|----------|----------|
| `POST /api/v1/finops/ingest/run` | Trigger ingest |
| `GET /api/v1/finops/reconciliation` | Drift vs invoice (04) |
| `GET /api/v1/finops/audit/export?limit=500` | Finance audit CSV (09) |
| `GET /api/v1/finops/ingest/status` | Freshness |

## Reconciliation gate

- Drift **≤ 1%** per provider/month vs cloud invoice
- Alert on `driftAlert=true` in reconciliation response

## Verification

```bash
chmod +x scripts/verify-finops-production.sh
FINOPS_AWS_CUR_FILE=backend/internal/finops/connectors/testdata/aws_cur_sample.ndjson \
FINOPS_AWS_INVOICE_USD=730.75 \
FINOPS_BILLING_MODE=live \
./scripts/verify-finops-production.sh
```

## Sales rule

Do not sell live FinOps to finance until reconciliation is green for **all** in-scope cloud accounts for 2 consecutive months.
