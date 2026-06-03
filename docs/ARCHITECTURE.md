# NeuralOps Platform — Architecture Reference

**Audience:** Architecture, platform engineering, security, and SRE teams  
**Scope:** Full stack — backend microservices, API gateway, observability UI APIs, data plane, agents, deployment  
**Status:** Living document aligned with repository layout as of 2026-06

---

## Table of contents

1. [Executive summary](#1-executive-summary)
2. [System context (C4 Level 1)](#2-system-context-c4-level-1)
3. [Container architecture (C4 Level 2)](#3-container-architecture-c4-level-2)
4. [Deployment architecture](#4-deployment-architecture)
5. [Backend structure](#5-backend-structure)
6. [API Gateway — component view](#6-api-gateway--component-view)
7. [Observability UI API layer](#7-observability-ui-api-layer)
8. [Frontend architecture](#8-frontend-architecture)
9. [Data architecture & ER model](#9-data-architecture--er-model)
10. [Section-wise flows](#10-section-wise-flows)
11. [Security architecture](#11-security-architecture)
12. [Platform observability](#12-platform-observability)
13. [API surface map](#13-api-surface-map)
14. [Related documents](#14-related-documents)

---

## 1. Executive summary

**NeuralOps** is an AI-powered observability platform for log analysis, distributed tracing, incident management, alerting, FinOps, and unified operations dashboards. The system is delivered as a **monorepo** with:

| Layer | Technology | Role |
|-------|------------|------|
| **UI** | React 18, TypeScript, Vite, TanStack Router/Query | Operator console; code-split routes; talks to gateway only |
| **Edge** | Go **API Gateway** (`:8080`) | Auth, RBAC, tenant context, rate limits, reverse proxy, **observability UI APIs**, WebSockets |
| **Services** | Go microservices (`:8081–8086`) | Ingestion, analysis, correlation, incident, search, alerting |
| **Data** | PostgreSQL, ClickHouse, Elasticsearch, Kafka, Redis, Qdrant | OLTP, analytics, logs/search, events, cache, vectors |
| **Agents** | `nexagent`, `collector-operator`, `materializer` | Host telemetry, K8s collector CRD, streaming aggregation |
| **IaC / Ops** | Docker Compose, K8s manifests, Terraform | Local demo, hyperscale paths |

**Design principles:** API-first, tenant isolation, secure-by-default gateway middleware, event-driven ingest pipeline, and a **fat gateway surface** for UI features (`internal/observability`) alongside **thin domain microservices** for core pipelines.

---

## 2. System context (C4 Level 1)

Shows who uses the system and what external systems connect.

```mermaid
C4Context
  title System Context — NeuralOps

  Person(operator, "Platform Operator", "SRE, DevOps, FinOps, Security")
  Person(developer, "Application Developer", "Uses traces, logs, dashboards")
  Person(admin, "Tenant Admin", "Users, SSO, policies, exports")

  System(neuralops, "NeuralOps Platform", "Observability, incidents, AI analysis, FinOps")

  System_Ext(idp, "Identity Provider", "OIDC / SAML (Okta, Azure AD, etc.)")
  System_Ext(cloud, "Cloud Providers", "AWS, GCP, Azure billing & inventory APIs")
  System_Ext(pagerduty, "PagerDuty / Slack / Jira", "Notifications & ITSM")
  System_Ext(prometheus, "Prometheus / Datadog", "Metrics & alert webhooks")
  System_Ext(cur, "Billing Exports", "CUR, BigQuery, Azure Cost Management")

  Rel(operator, neuralops, "Uses web UI & API")
  Rel(developer, neuralops, "Ingests telemetry, searches logs")
  Rel(admin, neuralops, "Configures tenant, ABAC, exports")
  Rel(neuralops, idp, "SSO login")
  Rel(neuralops, cloud, "Asset inventory, cost ingest")
  Rel(neuralops, pagerduty, "Oncall sync, notifications")
  Rel(neuralops, prometheus, "Webhooks, PromQL")
  Rel(neuralops, cur, "FinOps line items")
```

---

## 3. Container architecture (C4 Level 2)

Runtime containers and primary data flows.

```mermaid
C4Container
  title Container Diagram — NeuralOps (Docker Compose reference)

  Person(user, "User", "Browser")

  Container_Boundary(frontend, "Presentation") {
    Container(spa, "Frontend SPA", "React/Vite", "Static nginx :3000, proxies /api to gateway")
  }

  Container_Boundary(edge, "Edge") {
    Container(gateway, "API Gateway", "Go/Gin", "Auth, observability APIs, proxy :8080")
  }

  Container_Boundary(services, "Microservices") {
    Container(ingestion, "Ingestion", "Go", ":8081 → Kafka")
    Container(analysis, "Analysis", "Go", ":8082 LLM & scoring")
    Container(correlation, "Correlation", "Go", ":8083")
    Container(incident, "Incident", "Go", ":8084")
    Container(search, "Search", "Go", ":8085")
    Container(alerting, "Alerting", "Go", ":8086")
    Container(materializer, "Materializer", "Go", "Kafka → Postgres aggregates")
    Container(operator, "Collector Operator", "Go/K8s", "CollectorAgent CRD")
    Container(nexagent, "NexAgent", "Go/eBPF", "Host/collector agent")
  }

  Container_Boundary(data, "Data Plane") {
    ContainerDb(pg, "PostgreSQL", "Tenants, incidents, alerts, FinOps, UI config")
    ContainerDb(ch, "ClickHouse", "Analytics, traces, audit replica")
    ContainerDb(es, "Elasticsearch", "Log documents per tenant")
    ContainerDb(kafka, "Kafka", "Ingest & streaming topics")
    ContainerDb(redis, "Redis", "Rate limits, cache")
    ContainerDb(qdrant, "Qdrant", "Embeddings / semantic search")
  }

  Rel(user, spa, "HTTPS")
  Rel(spa, gateway, "REST /api/v1, WebSocket")
  Rel(gateway, ingestion, "POST /logs|metrics|traces|events")
  Rel(gateway, search, "GET/POST /search")
  Rel(gateway, incident, "/incidents, /services")
  Rel(gateway, analysis, "/analysis")
  Rel(gateway, alerting, "/alerts (subset)")
  Rel(gateway, pg, "SQL")
  Rel(gateway, ch, "Traces, analytics")
  Rel(ingestion, kafka, "Publish")
  Rel(analysis, kafka, "Consume")
  Rel(materializer, kafka, "Consume")
  Rel(materializer, pg, "Write aggregates")
  Rel(search, es, "Query")
  Rel(search, qdrant, "Vector search")
  Rel(incident, pg, "CRUD")
  Rel(alerting, pg, "CRUD")
```

### Request routing rule (important)

The gateway mounts **two classes** of `/api/v1` routes:

| Class | Handler | Examples |
|-------|---------|----------|
| **Native / Observability** | `internal/observability` on gateway | `/dashboards`, `/slos`, `/finops/*`, `/collectors/*`, `/query/unified` |
| **Proxied** | Reverse proxy to microservices | `POST /logs`, `/search/*`, `/incidents/*`, `/analysis/*`, legacy `/alerts` |

Wildcards are avoided where they would collide (e.g. `/logs/metric-rules` vs ingestion `POST /logs`).

---

## 4. Deployment architecture

### 4.1 Local / demo (Docker Compose)

```mermaid
flowchart TB
  subgraph Host["Developer machine"]
    Browser["Browser :3000"]
    GW["Gateway :8080"]
    FE["Frontend nginx :3000"]
    Browser --> FE
    FE -->|/api proxy| GW
  end

  subgraph Infra["infra/docker-compose.yml"]
    PG[(Postgres)]
    CH[(ClickHouse)]
    ES[(Elasticsearch)]
    KF[Kafka]
    RD[(Redis)]
    QD[(Qdrant)]
    JG[Jaeger]
    PR[Prometheus]
    GF[Grafana]
  end

  subgraph Svc["Go services"]
    ING[Ingestion]
    SRCH[Search]
    INC[Incident]
    ALT[Alerting]
    ANA[Analysis]
    MAT[Materializer]
  end

  GW --> PG & CH & RD
  GW --> ING & SRCH & INC & ALT & ANA
  ING --> KF
  MAT --> KF & PG
  SRCH --> ES & QD
```

### 4.2 Kubernetes (production-oriented)

| Artifact | Path | Purpose |
|----------|------|---------|
| Collector Operator | `deploy/kubernetes/operator/` | CRD `CollectorAgent`, RBAC, deployment |
| Materializer | `deploy/kubernetes/materializer/` | Streaming aggregation workers |
| Fleet profiling | `deploy/kubernetes/fleet-profiling.yaml` | Profiling ConfigMap pattern |

See [`HYPERSCALE.md`](HYPERSCALE.md) for operator, cloud SDK inventory, and materializer operations.

### 4.3 Build & delivery

| Component | Build | Runtime image |
|-----------|-------|---------------|
| Gateway + observability | `backend/Dockerfile` (multi-stage) | Single binary per service target |
| Frontend | `frontend/Dockerfile` | nginx serves `dist/` |
| Microservices | Same Dockerfile, `CMD` per service | `gateway`, `ingestion`, … |

---

## 5. Backend structure

```
backend/
├── cmd/                    # Service entrypoints
│   ├── gateway/            # API gateway (main UI + API entry)
│   ├── ingestion/          # Telemetry intake
│   ├── analysis/           # AI / classification
│   ├── correlation/        # Trace & dependency correlation
│   ├── incident/           # Incident lifecycle API
│   ├── search/             # Log & semantic search
│   ├── alerting/           # Alert rules & notifications
│   ├── materializer/       # Kafka → Postgres streaming
│   ├── collector-operator/ # K8s operator
│   ├── collector/          # Synthetic / health collector runner
│   ├── nexagent/           # Host agent (eBPF scaffold)
│   └── neuralops/          # CLI (admin, export, health)
├── internal/
│   ├── gateway/            # Auth, proxy, middleware, dashboard, chat, WS
│   ├── observability/      # UI-facing REST (APM, FinOps, SRS, NFR, …)
│   ├── finops/             # Cost domain service (used by observability HTTP)
│   ├── streaming/          # Kafka publisher, materializer logic
│   ├── ai/                 # LLM client
│   ├── tracequery/         # ClickHouse span store
│   ├── cloudinventory/     # AWS/GCP/Azure SDK collectors
│   ├── operator/           # controller-runtime reconciler
│   └── platform/           # DB, health, tenant helpers
├── pkg/nexagent/           # Agent pipeline, eBPF, buffer
├── migrations/             # Goose SQL (Postgres)
└── tests/                  # Contract & integration tests
```

### Microservice responsibilities

| Service | Port | Primary responsibility |
|---------|------|------------------------|
| **Gateway** | 8080 | Security edge, UI API surface, proxy composition |
| **Ingestion** | 8081 | Validate & publish logs/metrics/traces/events to Kafka |
| **Analysis** | 8082 | LLM classification, embeddings, anomaly hints |
| **Correlation** | 8083 | Service graph, deployment & temporal correlation |
| **Incident** | 8084 | Incidents, RCA artifacts, recommendations |
| **Search** | 8085 | Elasticsearch + Qdrant + ClickHouse queries |
| **Alerting** | 8086 | Alert ingestion, dedup, escalation, notify |

---

## 6. API Gateway — component view

```mermaid
flowchart LR
  subgraph Middleware["Middleware chain (order)"]
    M1[CORS]
    M2[Request logger]
    M3[License enforcement]
    M4[Auth rate limit]
    M5[Authenticator JWT/API key/OIDC]
    M6[Tenant + Developer scope]
    M7[Rate limit + Tenant quota]
    M8[RBAC]
    M9[Governance ABAC / residency]
    M10[Multi-region router]
    M11[Audit log]
  end

  subgraph Handlers["Handlers"]
    AUTH[Auth routes /auth/*]
    NAT[Native: dashboard, chat, WS, /info]
    OBS[Observability /api/v1/*]
    PRX[Proxies: ingestion, search, incident, analysis, alerting]
  end

  Client --> M1 --> M2 --> M3 --> M4 --> M5 --> M6 --> M7 --> M8 --> M9 --> M10 --> M11
  M11 --> AUTH & NAT & OBS & PRX
```

| Package | Responsibility |
|---------|----------------|
| `internal/gateway/auth` | JWT, API keys, OIDC PKCE, SAML, sessions, identity store |
| `internal/gateway/middleware` | RBAC, governance, quotas, audit, license |
| `internal/gateway/proxy` | mTLS-capable reverse proxy to upstream services |
| `internal/gateway/dashboard` | Aggregated dashboard payload for home screen |
| `internal/gateway/chat` | AI chat streaming to LLM |
| `internal/gateway/websocket` | Real-time hub + log tail |
| `internal/observability` | Large REST surface for UI features (see §7) |

---

## 7. Observability UI API layer

The UI does **not** call microservices directly. Almost all product screens use **`/api/v1`** on the gateway, implemented in `internal/observability`.

### 7.1 Route registration (logical domains)

```mermaid
flowchart TB
  RR[RegisterRoutes]
  RR --> CORE[registerCoreRoutes]
  RR --> DOM[registerDomainRoutes]

  CORE --> APM["/apm traces, flow"]
  CORE --> Q["/query unified, NexQL"]
  CORE --> COL["/collectors fleet, pipelines"]
  CORE --> MET["/metrics /dashboards"]
  CORE --> LOG["/logs rules, tiering"]
  CORE --> SEC["/security /integrations"]
  CORE --> ADM["/admin users, keys"]

  DOM --> CN[registerCloudNetworkRoutes]
  DOM --> EXP[registerExperienceRoutes]
  DOM --> GOV[registerGovernanceRoutes]
  DOM --> NFR[registerNFRRoutes]
  DOM --> FIN[registerFinOpsRoutes]
  DOM --> SRS[registerSRSRoutes]
  DOM --> EXT[RegisterExtendedRoutes]
  DOM --> DEP[RegisterDepthRoutes]
  DOM --> X[registerExtensionRoutes]
```

| Registrar | Domain | Examples |
|-----------|--------|----------|
| `registerCoreRoutes` | Core product | Traces, dashboards, SLOs, infra, security, marketplace |
| `registerCloudNetworkRoutes` | Cloud & NPM | `/cloud/assets`, `/network/flows`, FinOps summary |
| `registerFinOpsRoutes` | FinOps detail | Costs, budgets, chargeback, scenarios |
| `registerExperienceRoutes` | RUM & synthetic | Funnels, browser/mobile tests, KPI packs |
| `registerGovernanceRoutes` | Enterprise admin | ABAC, residency, MSP, exports |
| `registerNFRRoutes` | Certification | Benchmarks, reliability, a11y |
| `registerSRSRoutes` | SRS parity | Collector upgrade, alert scores, derived metrics |
| `RegisterExtendedRoutes` | Integrations | Cloud metrics query, SSO, on-call, OAuth |
| `registerExtensionRoutes` | Roadmap / depth | eBPF flows, DBM, billing sources, CI webhooks |

### 7.2 Observability component diagram

```mermaid
flowchart TB
  subgraph HTTP["observability.Handler"]
    ROUTES[routes.go]
    H_CORE[handler.go - core pages]
    H_FIN[finops_handlers.go]
    H_CLOUD[phase5 cloud/network]
    H_EXP[phase6 experience]
    H_GOV[phase7 governance]
    H_NFR[phase8 NFR]
    H_AI[ai_handlers.go]
    H_NEX[nexql_handlers.go]
    H_EXT[extended_handlers.go]
    H_ROAD[roadmap_partial_handlers.go]
  end

  subgraph Stores["Persistence"]
    MEM[Store in-memory + seeds]
    PG_REPO[PostgresRepo / SRS / Collectors]
    CH[ClickHouse spans]
    FIN_SVC[finops.Service]
  end

  ROUTES --> HTTP
  HTTP --> MEM & PG_REPO & CH & FIN_SVC
```

**Persistence strategy:** Demo-ready in-memory `Store` seeds power most screens without Postgres; when `POSTGRES_DSN` is set, repositories overlay tenants, SRS, collectors, FinOps line items, and governance policies.

---

## 8. Frontend architecture

### 8.1 Stack

| Concern | Choice |
|---------|--------|
| UI framework | React 18 + TypeScript |
| Build | Vite 5 |
| Routing | TanStack Router (code-split `lazyPage`) |
| Server state | TanStack Query |
| Client state | Zustand (auth, theme) |
| Styling | Tailwind CSS 4 |
| Charts / graph | Recharts, XYFlow |
| Session replay | rrweb |

### 8.2 Frontend container (C4)

```mermaid
flowchart TB
  subgraph Browser
    APP[App.tsx AuthBootstrap]
    ROUTER[TanStack Router]
    PAGES[Lazy pages /pages/*]
    API[api/* axios client]
    STORE[store/authStore, themeStore]
  end

  APP --> ROUTER --> PAGES
  PAGES --> API
  API -->|"/api/v1 + Bearer"| GW[API Gateway]
  APP --> STORE
```

### 8.3 Directory layout

```
frontend/src/
├── api/           # Typed clients (incidents, logs, finops, …)
├── pages/         # Route screens (50+ lazy-loaded)
├── components/    # UI, layout, domain widgets
├── routes/        # lazyPage + chunk recovery
├── store/         # Zustand stores
├── hooks/         # Shared React hooks
├── i18n/          # Messages (en/es/de)
└── router.tsx     # Route tree & auth guards
```

### 8.4 Auth flow (UI)

```mermaid
sequenceDiagram
  participant U as User
  participant SPA as Frontend
  participant GW as Gateway /auth

  U->>SPA: Open app
  SPA->>GW: GET /auth/config
  alt OIDC enabled
    U->>GW: Login redirect
    GW-->>SPA: Callback with exchange code
    SPA->>GW: POST /auth/token (exchange)
  else Local/JWT
    U->>SPA: Email/password
    SPA->>GW: POST /auth/login
  end
  GW-->>SPA: access + refresh tokens
  SPA->>GW: GET /auth/me (Bearer)
  Note over SPA: apiClient injects Authorization + X-Tenant-ID
```

### 8.5 Production serving

| Mode | Port | Notes |
|------|------|-------|
| Dev (`npm run dev`) | 5173 | Vite proxies `/api` → `:8080` |
| Docker | 3000 | nginx serves `dist/`, proxies `/api/` to `gateway:8080` |

Lazy routes load hashed chunks (e.g. `Incidents-*.js`). After deploy, users may need hard refresh; `lazyPage.tsx` auto-reloads once on chunk load failure.

---

## 9. Data architecture & ER model

### 9.1 Store roles

| Store | Workload | Tenant isolation |
|-------|----------|------------------|
| **PostgreSQL** | System of record: users, incidents, alerts, FinOps, UI entities | `tenant_id` column / FK |
| **ClickHouse** | High-volume logs analytics, traces, audit replica | Partitioning by tenant / time |
| **Elasticsearch** | Full-text log search | Index per tenant pattern |
| **Kafka** | Ingest fan-out, streaming samples | Headers / topic naming |
| **Redis** | Rate limiting, ephemeral metrics | Key prefix by tenant |
| **Qdrant** | Vector embeddings | Collection per tenant |

### 9.2 Core ER diagram (PostgreSQL — simplified)

```mermaid
erDiagram
  TENANTS ||--o{ USERS : has
  TENANTS ||--o{ INCIDENTS : owns
  TENANTS ||--o{ ALERTS : owns
  TENANTS ||--o{ ALERT_RULES : configures
  TENANTS ||--o{ FINOPS_COST_LINE_ITEMS : billed
  TENANTS ||--o{ OBSERVABILITY_DASHBOARDS : stores
  TENANTS ||--o{ OBSERVABILITY_SLOS : defines

  INCIDENTS ||--o{ MTTR_HISTORY : tracks
  ALERTS }o--o| INCIDENTS : may_link

  TENANTS {
    uuid id PK
    string name
    string plan_tier
    jsonb settings
  }

  USERS {
    uuid id PK
    uuid tenant_id FK
    string email
    string role
  }

  INCIDENTS {
    uuid id PK
    string tenant_id
    string title
    string severity
    string status
    timestamptz start_time
    jsonb timeline
  }

  ALERTS {
    uuid id PK
    string tenant_id
    string fingerprint
    string severity
    string status
    uuid linked_incident_id FK
  }

  FINOPS_COST_LINE_ITEMS {
    string id PK
    string tenant_id
    date billing_period
    string provider
    float effective_cost
    jsonb tags
  }

  OBSERVABILITY_DASHBOARDS {
    uuid id PK
    string tenant_id
    string name
    jsonb tiles
  }

  OBSERVABILITY_SLOS {
    uuid id PK
    string tenant_id
    string service
    float target
    float burn_rate
  }
```

### 9.3 FinOps ER (extension)

```mermaid
erDiagram
  TENANTS ||--o{ FINOPS_COST_LINE_ITEMS : ingests
  TENANTS ||--o{ FINOPS_BUDGETS : sets
  TENANTS ||--o{ FINOPS_ALLOCATION_RULES : defines
  TENANTS ||--o{ FINOPS_ANOMALIES : detects

  FINOPS_BUDGETS {
    string id PK
    string tenant_id
    string scope_type
    float amount_usd
  }

  FINOPS_ALLOCATION_RULES {
    string id PK
    string tenant_id
    string tag_key
    int priority
  }
```

Migrations: `backend/migrations/000023_finops_wave1.up.sql` through `000025_finops_wave3.up.sql`.

### 9.4 Streaming materialization

Topic: `neuralops.observability.stream`  
Consumers: `cmd/materializer` → Postgres tables (`000022_streaming_materialization.up.sql`)  
Producers: Gateway observability (alert policy triggers, derived metrics)

---

## 10. Section-wise flows

### 10.1 Log ingestion

```mermaid
sequenceDiagram
  participant Agent as Collector / App
  participant GW as Gateway
  participant ING as Ingestion
  participant K as Kafka
  participant ANA as Analysis
  participant ES as Elasticsearch

  Agent->>GW: POST /api/v1/logs (Bearer, X-Tenant-ID)
  GW->>ING: Proxy (mTLS optional)
  ING->>ING: Validate, normalize
  ING->>K: Publish log event
  K->>ANA: Consume (classification, embeddings)
  ANA->>ES: Index document (tenant-scoped)
```

### 10.2 Unified search (UI workbench)

```mermaid
sequenceDiagram
  participant UI as Log Explorer / Query Workbench
  participant GW as Gateway
  participant OBS as observability.Handler
  participant SRCH as Search service
  participant ES as Elasticsearch
  participant CH as ClickHouse

  UI->>GW: POST /api/v1/query/unified
  GW->>OBS: UnifiedQuery handler
  alt Complex cross-signal
    OBS->>SRCH: Proxy or delegated query
    SRCH->>ES: Full-text
    SRCH->>CH: Analytics fallback
  else In-memory / CH only
    OBS->>CH: Local query planner
  end
  OBS-->>UI: Envelope data + rows
```

### 10.3 Incident lifecycle

```mermaid
sequenceDiagram
  participant UI as Incidents UI
  participant GW as Gateway
  participant INC as Incident service
  participant PG as PostgreSQL
  participant ANA as Analysis

  UI->>GW: GET /api/v1/incidents
  GW->>INC: Proxy
  INC->>PG: List by tenant
  INC-->>UI: Incident list

  Note over ANA,INC: Correlation engine may open/update incidents from alerts & traces

  UI->>GW: GET /api/v1/incidents/{id}
  GW->>INC: Proxy
  INC->>PG: Detail + timeline
  opt RCA requested
    UI->>GW: POST /api/v1/analysis/...
    GW->>ANA: LLM root-cause assist
  end
```

### 10.4 Alerting & notification

```mermaid
sequenceDiagram
  participant Prom as Prometheus/Webhook
  participant GW as Gateway
  participant ING as Ingestion
  participant ALT as Alerting
  participant PG as PostgreSQL
  participant UI as Alerts UI

  Prom->>GW: POST /api/v1/webhooks/prometheus
  GW->>ING: Forward
  ING->>ALT: Internal route / alert event
  ALT->>PG: Dedup by fingerprint, insert alert
  ALT->>ALT: Escalation policy evaluation

  UI->>GW: GET /api/v1/alerts/policies (observability)
  Note over GW: UI policies on gateway; firing alerts may use alerting svc

  UI->>GW: PATCH acknowledge via proxy /api/v1/alerts/{id}/acknowledge
  GW->>ALT: Proxy
```

### 10.5 FinOps cost ingest & dashboard

```mermaid
sequenceDiagram
  participant Sched as Gateway scheduler
  participant FIN as finops.Service
  participant Cloud as AWS/GCP/Azure APIs
  participant PG as PostgreSQL
  participant UI as CloudFinOps UI

  Sched->>FIN: IngestAll(tenantId) periodic
  FIN->>Cloud: Billing export / inventory
  FIN->>PG: Upsert finops_cost_line_items
  UI->>GW: GET /api/v1/finops/costs?scope=
  GW->>FIN: GetCosts(tenant scoped)
  FIN->>PG: Aggregate line items
  FIN-->>UI: Cost series + forecast
```

### 10.6 AI chat (operations copilot)

```mermaid
sequenceDiagram
  participant UI as AI Chat page
  participant GW as Gateway chat handler
  participant LLM as OpenAI-compatible API
  participant SRCH as Search (tools)

  UI->>GW: POST /api/v1/chat (stream)
  GW->>SRCH: Optional RAG context fetch
  GW->>LLM: Completion with tools
  loop SSE chunks
    LLM-->>GW: token
    GW-->>UI: chunk
  end
```

### 10.7 Collector operator (Kubernetes)

```mermaid
sequenceDiagram
  participant Ops as Platform admin
  participant K8s as Kubernetes API
  participant OP as collector-operator
  participant GW as Gateway
  participant Agent as Collector pod

  Ops->>K8s: Apply CollectorAgent CRD
  K8s->>OP: Reconcile CR
  OP->>GW: Register fleet / pipeline config
  OP->>Agent: Deploy DaemonSet / sidecar
  Agent->>GW: POST telemetry (logs/traces)
```

### 10.8 Streaming materialization

```mermaid
sequenceDiagram
  participant UI as Alert Policies UI
  participant GW as Gateway
  participant K as Kafka
  participant MAT as Materializer
  participant PG as PostgreSQL

  UI->>GW: POST trigger evaluation
  GW->>K: Publish stream event
  MAT->>K: Consumer group
  MAT->>PG: Upsert score buckets / samples
  UI->>GW: GET /api/v1/alerts/policies/{id}/scores
  GW->>PG: Read materialized view
```

### 10.9 Authentication & governance (request path)

```mermaid
sequenceDiagram
  participant UI as Frontend
  participant GW as Gateway middleware
  participant GOV as GovernanceService
  participant OBS as Observability handler

  UI->>GW: API request + Bearer + X-Tenant-ID
  GW->>GW: Authenticate & RBAC
  GW->>GOV: EvaluateABAC(role, path)
  alt Denied
    GOV-->>UI: 403 GOV001
  else Allowed
    GW->>OBS: Handler
    OBS-->>UI: 200 success envelope
  end
```

---

## 11. Security architecture

| Control | Implementation |
|---------|----------------|
| **Authentication** | JWT access/refresh, API keys, OIDC PKCE, SAML option |
| **Authorization** | RBAC roles: `ADMIN`, `SRE`, `DEVELOPER`, `READ_ONLY`, `ALERT_MANAGER` |
| **ABAC / residency** | `GovernanceService` + middleware `GOV001`/`GOV002` |
| **Tenant isolation** | `X-Tenant-ID` + context; FinOps & search scoped by tenant |
| **Rate limiting** | Redis-backed global + per-tenant quotas |
| **Audit** | Postgres `audit` tables + optional ClickHouse replication |
| **mTLS** | Optional upstream proxy to microservices |
| **Secrets** | Env / K8s secrets (no keys in repo) |

Evidence and checklist: [`docs/SECURITY_EVIDENCE.md`](SECURITY_EVIDENCE.md).

---

## 12. Platform observability

| Signal | Tool | Endpoint |
|--------|------|----------|
| Metrics | Prometheus | Each service `/metrics` |
| Traces | OpenTelemetry → Jaeger | `OTEL_EXPORTER_OTLP_ENDPOINT` |
| Dashboards | Grafana | Compose stack in `infra/` |
| Logs | Structured JSON (Zap) | `traceId`, `tenantId` fields |

---

## 13. API surface map

| Prefix | Owner | Purpose |
|--------|-------|---------|
| `/health`, `/ready`, `/live` | Gateway | Probes |
| `/auth/*` | Gateway | Login, OIDC, refresh, me |
| `/api/v1/info` | Gateway | Build info |
| `/api/v1/logs` (POST) | Ingestion proxy | Log ingest |
| `/api/v1/search/*` | Search proxy | Search & traces |
| `/api/v1/incidents/*` | Incident proxy | Incidents API |
| `/api/v1/analysis/*` | Analysis proxy | AI analysis |
| `/api/v1/alerts` (subset) | Alerting proxy | Legacy alerting API |
| `/api/v1/*` (majority) | Observability | UI dashboards, FinOps, SRS, NFR, etc. |
| `/swagger/*` | Gateway | OpenAPI UI |

Contract tests: `backend/tests/contract/observability_contract_test.go`  
OpenAPI: `backend/openapi/swagger.yaml`, `docs/openapi/gateway-v1.yaml`

---

## 14. Related documents

| Document | Content |
|----------|---------|
| [`../ARCHITECTURE.md`](../ARCHITECTURE.md) | Short architecture summary |
| [`../README.md`](../README.md) | Quick start & ports |
| [`SRS_NEURALOPS_COMPARISON_GAP.md`](SRS_NEURALOPS_COMPARISON_GAP.md) | SRS traceability |
| [`BANK_PRODUCTION_READINESS_ROADMAP.md`](BANK_PRODUCTION_READINESS_ROADMAP.md) | Production gates |
| [`HYPERSCALE.md`](HYPERSCALE.md) | Operator, materializer, cloud SDK |
| [`FINOPS_ENHANCEMENT_REQUIREMENTS.md`](FINOPS_ENHANCEMENT_REQUIREMENTS.md) | FinOps depth |
| [`SECURITY_EVIDENCE.md`](SECURITY_EVIDENCE.md) | Security controls evidence |
| [`LOCAL_TEST_URLS.md`](LOCAL_TEST_URLS.md) | Dev URLs & smoke tests |

---

## Document maintenance

When adding a new **UI screen**, expect: `frontend/src/pages/*`, `frontend/src/api/*`, and a registrar in `backend/internal/observability/routes.go` (or domain handler file).

When adding a new **microservice endpoint**, expect: service `internal/*/handler`, gateway proxy route in `internal/gateway/app.go`, and contract test update.

**Diagram legend:** Mermaid diagrams render in GitHub, GitLab, VS Code, and most Markdown viewers. For formal C4 exports, import the Mermaid sources into Structurizr or IcePanel if your team standardizes on those tools.
