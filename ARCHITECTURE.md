# Architecture

NeuralOps is an AI-powered observability platform for log analysis, incident management, and service intelligence.

## High-level design

```
Clients (Web UI, API keys, webhooks)
        │
        ▼
   API Gateway ──► RBAC, auth, rate limits, audit, WebSocket
        │
        ├── Ingestion ──► Kafka ──► Analysis / Correlation
        ├── Search ──► Elasticsearch + Qdrant + ClickHouse
        ├── Incident ──► Postgres
        └── Alerting ──► Postgres + notifiers
```

## Services

| Service | Port | Responsibility |
|---------|------|----------------|
| Gateway | 8080 | Auth, proxy, dashboard aggregation, AI chat, WebSocket |
| Ingestion | 8081 | Log/metric/event/trace intake, Kafka publish |
| Analysis | 8082 | LLM classification, embeddings, anomaly scoring |
| Correlation | 8083 | Trace/transaction correlation |
| Incident | 8084 | Incident lifecycle, RCA, recommendations |
| Search | 8085 | Full-text, semantic, trace, transaction search |
| Alerting | 8086 | Alert ingestion, dedup, escalation, notifications |

## Data stores

- **PostgreSQL** — tenants, users, incidents, alerts, audit logs, auth sessions
- **Elasticsearch** — log documents (per-tenant indexes `tenant-{id}-logs-*`)
- **ClickHouse** — analytics logs, transactions, audit replication
- **Qdrant** — vector embeddings for semantic search
- **Redis** — rate limiting, rolling metrics
- **Kafka** — event bus between ingestion and processors

## Security model

- Authentication: internal JWT, API keys, OIDC PKCE with one-time SPA exchange codes
- Authorization: RBAC (`ADMIN`, `SRE`, `DEVELOPER`, `READ_ONLY`, `ALERT_MANAGER`)
- Tenant isolation: gateway injects `X-Tenant-ID`; search/analysis scope by tenant
- Audit: Postgres source of truth with ClickHouse replication

## Observability

- Prometheus metrics on `/metrics` for every service
- OpenTelemetry traces exported to Jaeger via OTLP
- Grafana dashboards for platform overview, Kafka ingestion, search/incidents, AI usage, gateway API

## Deployment targets

- Local/demo: `infra/docker-compose.yml`
- Kubernetes: `infra/k8s/` (Kustomize)
- Terraform bootstrap: `infra/terraform/`
