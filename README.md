# NeuralOps — AI-Powered Observability Platform

Production-grade monorepo for an AI-powered log analysis and incident management platform.

## Architecture

Full reference (C4, sequence, ER, component diagrams, section-wise flows):

- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** — system design
- **[docs/GAP_IMPLEMENTATION_PLAN.md](docs/GAP_IMPLEMENTATION_PLAN.md)** — partial/missing backlog (implement one-by-one)

```
backend/     Go microservices (gateway, ingestion, analysis, correlation, incident, search, alerting)
frontend/    React 18 + TypeScript + Vite dashboard
infra/       Docker Compose, Kubernetes (k8s/), Terraform
scripts/     Developer utilities
```

## Prerequisites

- Go 1.25+ (matches `backend/go.mod`)
- Node.js 20+
- Docker & Docker Compose

## Quick Start

```bash
# Start all infrastructure and services
docker compose -f infra/docker-compose.yml up -d --build

# Verify gateway health
curl http://localhost:8080/health

# Open the UI
open http://localhost:3000
```

Or use the helper script:

```bash
./scripts/quickstart.sh
```

## CLI & IaC (Phase 7)

### NeuralOps CLI

```bash
cd backend
go build -o neuralops ./cmd/neuralops
./neuralops version
./neuralops msp-tenants list
./neuralops abac-policies get --tenant default
./neuralops export logs --destination s3://bucket/path
```

### Terraform provider

```bash
cd terraform-provider-neuralops
go build -o terraform-provider-neuralops
# dev_overrides in ~/.terraformrc pointing at this directory
```

Resources: `neuralops_export_job`, `neuralops_alert_policy`, `neuralops_branding`.

### Pulumi

See [`providers/pulumi-neuralops/README.md`](providers/pulumi-neuralops/README.md) for the Terraform-bridge workflow.

### Hyperscale (operator, cloud SDKs, streaming)

See [`docs/HYPERSCALE.md`](docs/HYPERSCALE.md) for controller-runtime collector operator, AWS/GCP/Azure inventory, and the Kafka materializer service.

## Local Development

### Backend

```bash
cd backend
go mod tidy
go build ./...

# Run a single service
HTTP_PORT=8080 go run ./cmd/gateway
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

The Vite dev server proxies `/api` to `http://localhost:8080`.
For local smoke-test commands and dev-port fallback details (`5173`/`5174`), see [`docs/LOCAL_TEST_URLS.md`](docs/LOCAL_TEST_URLS.md).

## Services & Ports

| Service      | Port | Description              |
|-------------|------|--------------------------|
| Frontend    | 3000 | React UI (nginx in Docker) |
| Gateway     | 8080 | API gateway (main entry) |
| Ingestion   | 8081 | Log ingestion            |
| Analysis    | 8082 | AI analysis engine       |
| Correlation | 8083 | Correlation engine       |
| Incident    | 8084 | Incident management      |
| Search      | 8085 | Search service           |
| Alerting    | 8086 | Alerting service         |

## Infrastructure (Docker Compose)

| Component     | Port(s)        |
|--------------|----------------|
| PostgreSQL 16 | 5432          |
| ClickHouse    | 8123, 9000    |
| Elasticsearch 8 | 9200        |
| Kafka         | 9092, 29092   |
| Zookeeper     | 2181          |
| Qdrant        | 6333, 6334    |
| Redis         | 6379          |

## Health Endpoints

Every backend service exposes:

- `GET /health` — liveness
- `GET /ready` — readiness
- `GET /live` — alive probe
- `GET /metrics` — Prometheus metrics

## Module

Go module: `github.com/neuralops/platform`

## Documentation

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | System architecture and service boundaries |
| [API.md](API.md) | Top-level API reference and conventions |
| [RUNBOOK.md](RUNBOOK.md) | Operational procedures (startup, health checks, recovery) |
| [TROUBLESHOOTING.md](TROUBLESHOOTING.md) | Common local/dev failure modes and fixes |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contribution standards and development workflow |
| [CHANGELOG.md](CHANGELOG.md) | Notable product and platform changes |
| [docs/LOCAL_TEST_URLS.md](docs/LOCAL_TEST_URLS.md) | Local URLs to test UI, APIs, and observability after quickstart |
| [docs/openapi/gateway-v1.yaml](docs/openapi/gateway-v1.yaml) | Gateway API v1 OpenAPI contract (full gateway surface) |
| [docs/openapi/observability-v1.yaml](docs/openapi/observability-v1.yaml) | Observability-focused OpenAPI surface |
| [backend/openapi/swagger.yaml](backend/openapi/swagger.yaml) | Backend Swagger/OpenAPI YAML for gateway consumers |
| [docs/UI_IMPROVEMENT_ROADMAP.md](docs/UI_IMPROVEMENT_ROADMAP.md) | Phased, trackable UI/UX improvement backlog (checkboxes per task) |
| [docs/PRODUCTION_AND_GTM.md](docs/PRODUCTION_AND_GTM.md) | Production ship plan, enterprise sales, pricing, and onboarding guidelines |
| [docs/GCP_SETUP_GUIDE.md](docs/GCP_SETUP_GUIDE.md) | End-to-end GCP & GKE monitoring setup (projects, data plane, agents, FinOps) |
| [docs/BANK_PRODUCTION_READINESS_ROADMAP.md](docs/BANK_PRODUCTION_READINESS_ROADMAP.md) | Bank prod + sales checklist (waves 0–6, gates A–F); `./scripts/gate-verify.sh --gate C` |
| [docs/bank/README.md](docs/bank/README.md) | Bank customer pack index + gate close-out runbooks |
| [docs/bank/README.md](docs/bank/README.md) | Bank customer pack index (security, legal, runbooks) |
| [docs/UI_DYNATRACE_GAP.md](docs/UI_DYNATRACE_GAP.md) | UI feature gap analysis vs Dynatrace (parity matrix and roadmap) |
| [docs/UI_DYNATRACE_IMPLEMENTATION_PLAN.md](docs/UI_DYNATRACE_IMPLEMENTATION_PLAN.md) | Phase-wise plan to implement all Dynatrace UI gaps (12 phases) |
| [docs/AUDIT_COMPLIANCE.md](docs/AUDIT_COMPLIANCE.md) | CI quality gates and auth/mTLS audit traceability |
| [docs/FRONTEND_STACK.md](docs/FRONTEND_STACK.md) | Frontend architecture (TanStack Router, Tailwind v4) |

### OpenAPI Sync

- Canonical source: `docs/openapi/gateway-v1.yaml`
- Backend copy: `backend/openapi/swagger.yaml`
- Sync command:

```bash
./scripts/sync-openapi.sh
```

## Phase Roadmap

- **Phase 0** — Monorepo scaffold (current)
- **Phase 1** — Domain models
- **Phase 2+** — Ingestion, AI analysis, correlation, UI

## License

Proprietary — NeuralOps Platform
