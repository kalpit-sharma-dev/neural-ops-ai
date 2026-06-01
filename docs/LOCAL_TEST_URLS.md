# Local test URLs

After running `bash scripts/quickstart.sh`, use these links to verify the NeuralOps stack on your machine. All URLs assume the default `infra/docker-compose.yml` port mappings on **localhost**.

## Quickstart rebuild behavior

By default, `scripts/quickstart.sh`:

1. Stops the stack and **removes compose-built images** (`docker compose down --rmi local`)
2. Rebuilds **all images without Docker cache** (`--no-cache`) so UI and API changes always apply
3. Starts containers with **`--force-recreate`**
4. Prunes **dangling** images left from old builds

Faster repeat runs (uses layer cache):

```bash
QUICKSTART_USE_CACHE=1 bash scripts/quickstart.sh
```

If the UI still looks old in the browser, hard-refresh: **Ctrl+Shift+R** (Windows/Linux) or **Cmd+Shift+R** (macOS).

## Primary application

| What | URL |
|------|-----|
| Web UI | http://localhost:3000 |
| API Gateway (health) | http://localhost:8080/health |
| API info | http://localhost:8080/api/v1/info |
| Swagger UI | http://localhost:8080/swagger/index.html |
| Gateway Prometheus metrics | http://localhost:8080/metrics |

## Authentication

| What | URL | Credentials / notes |
|------|-----|---------------------|
| Keycloak admin console | http://localhost:8088 | `admin` / `admin` |
| Keycloak realm (neuralops) | http://localhost:8088/realms/neuralops | OIDC issuer for the UI |
| Dev login (UI) | http://localhost:3000 | Email: `demo@neuralops.ai` |
| OIDC callback (gateway) | http://localhost:8080/api/v1/auth/oidc/callback | Used by SSO redirect flow |

**Dev tenant ID:** `00000000-0000-0000-0000-000000000002`  
**Keycloak demo user password:** `demo1234`

## Backend services (health endpoints)

| Service | URL |
|---------|-----|
| Ingestion | http://localhost:8081/health |
| Analysis | http://localhost:8082/health |
| Correlation | http://localhost:8083/health |
| Incident | http://localhost:8084/health |
| Search | http://localhost:8085/health |
| Alerting | http://localhost:8086/health |
| mTLS proxy (ingestion) | https://localhost:8443/health |

## Observability stack

| What | URL | Credentials / notes |
|------|-----|---------------------|
| Grafana | http://localhost:3001 | `admin` / `admin` |
| Prometheus | http://localhost:9090 | Targets, PromQL, alerts |
| Jaeger UI | http://localhost:16686 | Distributed traces |
| Pyroscope | http://localhost:4040 | Continuous profiling |
| kube-state-metrics (dev mock) | http://localhost:8087/metrics | Static K8s metrics for local compose |

## Data stores (smoke checks)

| What | URL |
|------|-----|
| Elasticsearch | http://localhost:9200 |
| ClickHouse HTTP | http://localhost:8123/ping |
| Qdrant | http://localhost:6333 |

## UI routes

| Page | URL |
|------|-----|
| Dashboard (Command Center) | http://localhost:3000/ |
| Log Explorer | http://localhost:3000/logs |
| Incidents | http://localhost:3000/incidents |
| Trace Explorer | http://localhost:3000/traces |
| Compare Traces | http://localhost:3000/traces/compare |
| Service Flow | http://localhost:3000/service-flow |
| Metrics | http://localhost:3000/metrics |
| Dashboards | http://localhost:3000/dashboards |
| Transactions | http://localhost:3000/transactions |
| Anomalies | http://localhost:3000/anomalies |
| SLOs | http://localhost:3000/slos |
| Service Map | http://localhost:3000/service-map |
| Infrastructure | http://localhost:3000/infrastructure |
| Kubernetes | http://localhost:3000/kubernetes |
| Databases | http://localhost:3000/databases |
| Middleware | http://localhost:3000/middleware |
| RUM | http://localhost:3000/rum |
| Synthetic | http://localhost:3000/synthetic |
| AI Assistant | http://localhost:3000/ai-chat |
| Cloud | http://localhost:3000/cloud |
| Workflows | http://localhost:3000/workflows |
| Notebooks | http://localhost:3000/notebooks |
| Security | http://localhost:3000/security |
| Marketplace | http://localhost:3000/marketplace |
| Integrations | http://localhost:3000/integrations |
| Alerts | http://localhost:3000/alerts |
| Settings | http://localhost:3000/settings |
| Design system | http://localhost:3000/design-system |

## Sample API calls (via gateway)

These require authentication unless auth is disabled. Sign in via the UI first, or use dev login.

| What | URL |
|------|-----|
| Auth config | http://localhost:8080/api/v1/auth/config |
| List dashboards | http://localhost:8080/api/v1/dashboards |
| Metric catalog | http://localhost:8080/api/v1/metrics/catalog |
| Topology | http://localhost:8080/api/v1/topology |
| List incidents | http://localhost:8080/api/v1/incidents |

## Internal ports (optional tooling)

Exposed on localhost for debugging with external clients:

| Service | Port |
|---------|------|
| PostgreSQL | 5432 |
| Redis | 6379 |
| Kafka (internal listener) | 9092 |
| Kafka (host listener) | 29092 |
| ClickHouse native | 9000 |
| Zookeeper | 2181 |
| OTLP gRPC (Jaeger) | 4317 |

## Quick verification commands

```bash
curl -fsS http://localhost:8080/health
curl -fsS http://localhost:8123/ping
curl -fsS -o /dev/null -w "%{http_code}\n" http://localhost:3000/
docker ps --filter name=neuralops --format "table {{.Names}}\t{{.Status}}"
```

## Related docs

- [README.md](../README.md) — project overview
- [RUNBOOK.md](../RUNBOOK.md) — operations
- [TROUBLESHOOTING.md](../TROUBLESHOOTING.md) — common failures
- [API.md](../API.md) — API reference

## Stop the stack

```bash
docker compose -f infra/docker-compose.yml down
```
