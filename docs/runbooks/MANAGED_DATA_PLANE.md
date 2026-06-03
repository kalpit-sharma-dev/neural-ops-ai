# Managed Data Plane — Postgres, Kafka, ES, ClickHouse, Redis (Wave 1.4)

**ID:** PROD-INFRA-01

## Principle

**Do not** run stateful data stores as single-replica in-cluster Deployments for bank production. Use customer-managed or cloud-managed services with HA, encryption at rest, and automated backups.

## Reference topology

| Store | Managed option | NeuralOps usage |
|-------|----------------|-----------------|
| **PostgreSQL 15+** | RDS / Cloud SQL / Azure DB | Tenants, SRS, FinOps, workflows, integrations |
| **Kafka 3.x** | MSK / Confluent / Event Hubs | Log/trace ingest pipeline |
| **Elasticsearch 8** | OpenSearch / Elastic Cloud | Log search, analytics |
| **ClickHouse** | CH Cloud / self HA cluster | Transactions, high-cardinality analytics |
| **Redis 7** | ElastiCache / Memorystore | Cache, rate limits, sessions |

## Connection secrets (Helm)

Create `neuralops-secrets` (or External Secrets) with:

| Key | Example |
|-----|---------|
| `DATABASE_URL` | `postgres://user:***@rds.internal:5432/neuralops?sslmode=require` |
| `KAFKA_BROKERS` | `b-1.msk.internal:9092,b-2...` |
| `ELASTICSEARCH_URL` | `https://es.internal:9200` |
| `CLICKHOUSE_DSN` | `clickhouse://ch.internal:9000/neuralops` |
| `REDIS_URL` | `rediss://cache.internal:6379` |

Wire via [external-secret.yaml](../../infra/helm/neuralops/templates/external-secret.yaml) and `values-prod.yaml` `secrets.create: false`.

## Sizing (staging starting point)

| Store | Staging | Production (POC scale) |
|-------|---------|------------------------|
| Postgres | db.r6g.large, 200 GB | Multi-AZ, 500 GB+, PITR |
| Kafka | 3 brokers, 500 GB total | 3+ AZ, retention per ingest runbook |
| ES | 3 data nodes, 1 TB | Hot/warm ILM per bank policy |
| ClickHouse | 2 replicas | Replication + backup to object store |
| Redis | 2 node cluster | TLS, auth, eviction policy `volatile-lru` |

## Terraform pointer

Customer-owned infra lives in **their** cloud account. Starter variable contract:

- See [infra/terraform/managed-data-plane/README.md](../../infra/terraform/managed-data-plane/README.md)
- Helm release in [infra/terraform/main.tf](../../infra/terraform/main.tf) — pass connection strings via sealed secrets, not Terraform state plaintext.

## Hardening checklist

- [ ] TLS for all data plane connections (`sslmode=require`, `rediss://`)
- [ ] Private subnets only; no public endpoints
- [ ] IAM / workload identity for backup jobs
- [ ] Encryption at rest (KMS)
- [ ] Backup RPO/RTO documented — [BACKUP_RESTORE.md](./BACKUP_RESTORE.md)

## Exit criteria (Gate B)

- [ ] All five stores are managed HA in staging
- [ ] Restore drill completed once per store class
- [ ] No demo seed jobs in production namespace
