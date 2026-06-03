# FinOps AWS CUR — S3 Sync & Helm Mount (Wave 5 + 6)

## Overview

Production FinOps reads billing NDJSON from a **PVC** mounted on `gateway`, synced nightly from the customer CUR S3 bucket.

```
AWS CUR S3 → CronJob (aws s3 sync) → PVC → FINOPS_AWS_CUR_FILE → ingest API
```

## Helm enable

```yaml
# values-prod.yaml
finopsCur:
  enabled: true
  s3:
    bucket: acme-billing-cur
    prefix: cur/NeuralOps/Monthly/
    region: us-east-1
    roleArn: arn:aws:iam::111122223333:role/neuralops-cur-reader
  curFile: /var/lib/neuralops/cur/latest.ndjson
```

```bash
helm upgrade -i neuralops infra/helm/neuralops \
  -f infra/helm/neuralops/values.yaml \
  -f infra/helm/neuralops/values-prod.yaml \
  --set finopsCur.s3.bucket=acme-billing-cur
```

## IRSA (EKS)

1. Create IAM policy: `s3:GetObject`, `s3:ListBucket` on CUR bucket/prefix.
2. Annotate `neuralops` ServiceAccount with role ARN.
3. Set `finopsCur.s3.roleArn` in values.

## CUR → NDJSON ETL

CUR arrives as Parquet/CSV. Customer pipeline must emit FOCUS-aligned NDJSON (see [FINOPS_PRODUCTION.md](../bank/FINOPS_PRODUCTION.md)):

- Option A: AWS Glue job writes `latest.ndjson` to same S3 prefix; sync job copies it.
- Option B: In-cluster converter CronJob after `aws s3 sync` (customer-maintained).

## Verify

```bash
kubectl logs -n neuralops job/$(kubectl get jobs -n neuralops -o name | grep finops-cur | head -1 | cut -d/ -f2)
kubectl exec -n neuralops deploy/gateway -- ls -la /var/lib/neuralops/cur
curl -X POST https://api.example/api/v1/finops/ingest/run -H "Authorization: Bearer $TOKEN"
curl https://api.example/api/v1/finops/reconciliation -H "Authorization: Bearer $TOKEN"
```

## GCP / Azure

Use file mount with `FINOPS_GCP_BILLING_FILE` / `FINOPS_AZURE_BILLING_FILE` from separate PVCs (same pattern).
