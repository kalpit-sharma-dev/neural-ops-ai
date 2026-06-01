# Troubleshooting

## Gateway returns 401

- Confirm `AUTH_DISABLED` is not `true` if you expect auth
- Use dev login (`demo@neuralops.ai`) or API key header `X-API-Key: demo-api-key`
- Check token expiry; refresh via `POST /api/v1/auth/refresh`

## No logs in Log Explorer

- Run seed: `go run ./backend/scripts/seed/main.go`
- Verify Elasticsearch: `curl http://localhost:9200/_cat/indices`
- Ensure tenant index exists: `tenant-*-logs-*`
- Check search service health on port 8085

## Kafka consumer lag high

- Inspect Grafana Kafka Ingestion dashboard
- Restart analysis/correlation pods or containers
- Verify Kafka health: `docker compose -f infra/docker-compose.yml ps kafka`

## ClickHouse / Postgres connection errors

- Wait for healthchecks after `docker compose up`
- Confirm DSNs in gateway environment match compose service names

## Frontend cannot reach API

- Dev: Vite proxies `/api` to `localhost:8080`
- Docker: use frontend container with gateway URL configured
- CORS: gateway `configs/gateway.yaml` allowed origins must include UI origin

## Integration tests skipped

- Tests require Docker: `docker info` must succeed
- Run with tag: `go test -tags=integration ./tests/integration/...`

## OIDC login fails

- OIDC requires IdP configuration (`OIDC_ENABLED`, issuer, client ID/secret)
- For local demos use dev login instead of SSO

## Build failures

```bash
cd backend && go mod tidy && go build ./...
cd frontend && npm ci && npm run build
```
