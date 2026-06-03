# Hyperscale production runbook

## 1. Collector operator (controller-runtime)

Deploy CRD, RBAC, and operator:

```bash
kubectl apply -f deploy/kubernetes/operator/crd-collectoragent.yaml
kubectl apply -f deploy/kubernetes/operator/rbac.yaml
kubectl apply -f deploy/kubernetes/operator/deployment.yaml
```

Create a `CollectorAgent` CR; the reconciler syncs to `NEURALOPS_API_URL` and supports canary via `spec.canaryVersion`.

Legacy file-based loop: set `OPERATOR_LEGACY=true` and `COLLECTOR_RECONCILE_SPEC=/config/reconcile.json`.

Build:

```bash
cd backend && go build -o collector-operator ./cmd/collector-operator
```

## 2. Cloud inventory (AWS / GCP / Azure SDKs)

```bash
export NEURALOPS_CLOUD_LIVE=true
export NEURALOPS_CLOUD_MAX_PAGES=500
export NEURALOPS_CLOUD_PAGE_SIZE=100
# AWS
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=...
# GCP
export GCP_PROJECT_ID=my-project
export GOOGLE_APPLICATION_CREDENTIALS=/path/sa.json
# Azure
export AZURE_SUBSCRIPTION_ID=...
```

`GET /api/v1/cloud/assets` uses paginated EC2/RDS, GCE aggregated list, and Azure VM list-all pager.

## 3. Streaming materialization

Kafka topic: `neuralops.observability.stream`

Run materializer (64 shards, 5s flush):

```bash
export KAFKA_BROKERS=localhost:9092
export POSTGRES_DSN=postgres://...
export MATERIALIZER_CONCURRENCY=32
go run ./cmd/materializer
```

Docker Compose includes a `materializer` service (depends on Kafka + Postgres + gateway). After `docker compose up`, verify end-to-end:

```bash
bash scripts/verify-materialization.sh
```

This triggers an alert policy, waits for Kafka consumption + Postgres flush, and asserts score buckets exist.

Gateway publishes samples when `KAFKA_BROKERS` is set (derived metric create, alert policy trigger).

Read APIs:

- `GET /api/v1/metrics/derived/{id}/samples`
- `GET /api/v1/alerts/policies/{id}/scores`

Migration: `000022_streaming_materialization.up.sql`
