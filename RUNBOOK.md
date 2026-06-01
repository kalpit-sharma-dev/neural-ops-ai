# Runbook

## Start the platform

```bash
docker compose -f infra/docker-compose.yml up -d --build
./scripts/quickstart.sh
```

## Seed demo data (50k logs, 500 transactions, 100 deployments)

```bash
cd backend
go run ./scripts/seed/main.go
```

## Health checks

| Component | Check |
|-----------|-------|
| Gateway | `curl http://localhost:8080/health` |
| All services | `curl http://localhost:8080/api/v1/info` |
| Prometheus | `http://localhost:9090/targets` |
| Grafana | `http://localhost:3001` (default admin/admin) |
| Jaeger | `http://localhost:16686` |

## Auth (development)

- Dev login: `demo@neuralops.ai` on tenant `00000000-0000-0000-0000-000000000002`
- API key: `demo-api-key`
- Compose sets `AUTH_DISABLED=false` with dev login enabled

## Incident response

1. Confirm active incidents in UI or `GET /api/v1/incidents?status=OPEN`
2. Inspect logs via Log Explorer or incident Logs tab
3. Review RCA and recommendations on incident detail page
4. Acknowledge and resolve when mitigated; audit entries are written to Postgres and ClickHouse

## Kafka lag

- Metric: `kafka_consumer_lag`
- Dashboard: Grafana → Kafka Ingestion
- If lag grows: scale analysis/correlation consumers, inspect DLQ topic

## Backup

- Postgres: snapshot `postgres_data` volume or logical dump
- Elasticsearch: snapshot repository (production)
- ClickHouse: `BACKUP` or replicated cluster

## Rollback

1. Deploy previous image tag via Kubernetes or Compose
2. Run down migrations only when explicitly approved
3. Verify `/ready` on gateway and downstream services
