# Multi-Region Data Residency (Wave 4.2)

**ID:** ADM-02

## API surfaces

| API | Purpose |
|-----|---------|
| `GET/PUT /api/v1/admin/data-residency` | Primary region, allowed regions, cross-border deny |
| Ingest `X-Region` header | Evaluated by governance middleware |

## Configuration example

```json
{
  "primaryRegion": "eu-west-1",
  "allowedRegions": ["eu-west-1", "eu-central-1"],
  "piiStorageRegion": "eu-west-1",
  "crossBorderDenied": true
}
```

Terraform: use `neuralops` provider admin APIs or UI **Settings → Governance**.

## Deployment patterns

| Pattern | When |
|---------|------|
| **Single region** | Default POC |
| **Active-passive DR** | Secondary region read replica; failover runbook |
| **In-region only** | EU bank — no US replicas |

## Per-region stacks

```
eu-west-1:  Helm release neuralops-eu + EU data plane
us-east-1:  Helm release neuralops-us (separate tenants)
```

Do **not** replicate tenant Postgres across borders when `crossBorderDenied: true`.

## Verification

```bash
curl -s "$GATEWAY/api/v1/admin/data-residency" -H "Authorization: Bearer $TOKEN" | jq .
# Ingest with disallowed region should 403
curl -s -X POST "$GATEWAY/api/v1/ingest/logs" \
  -H "X-Region: us-east-1" -H "X-Tenant-ID: eu-tenant" ...
```

## Related

- [FFIEC_CONTROL_MAPPING.md](../compliance/FFIEC_CONTROL_MAPPING.md)
- [DATA_FLOW.md](../bank/DATA_FLOW.md)
