# Cursor AI Prompt — AI Log Analyzer Observability Platform
## Full-Stack Build: Go Backend + React Frontend

---

> **HOW TO USE THIS PROMPT IN CURSOR**
> Open Cursor → New Chat (Composer mode) → Paste the relevant section.
> Work phase by phase. Each section is a self-contained Cursor prompt.

---

---

# ═══════════════════════════════════════════
# PHASE 0 — PROJECT SCAFFOLD & MONOREPO SETUP
# ═══════════════════════════════════════════

```
Create a production-grade monorepo for an AI-powered observability platform called "NeuralOps".

Monorepo structure:
neuralops/
├── backend/                   # Go microservices
│   ├── cmd/
│   │   ├── ingestion/         # Log ingestion service
│   │   ├── analysis/          # AI analysis engine
│   │   ├── correlation/       # Correlation engine
│   │   ├── incident/          # Incident management
│   │   ├── search/            # Search service
│   │   ├── alerting/          # Alerting service
│   │   └── gateway/           # API Gateway (main entry)
│   ├── internal/
│   │   ├── domain/            # Domain models
│   │   ├── repository/        # DB interfaces
│   │   ├── kafka/             # Kafka producer/consumer
│   │   ├── ai/                # AI/LLM client (Go)
│   │   ├── storage/           # Storage adapters
│   │   └── middleware/        # Auth, logging, tracing
│   ├── pkg/
│   │   ├── config/
│   │   ├── logger/
│   │   └── metrics/
│   ├── migrations/
│   ├── go.mod
│   └── go.sum
├── frontend/                  # React app
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── hooks/
│   │   ├── store/
│   │   ├── api/
│   │   └── styles/
│   ├── package.json
│   └── vite.config.ts
├── infra/
│   ├── docker-compose.yml
│   ├── k8s/
│   └── terraform/
├── scripts/
└── README.md

Use:
- Go 1.22+ with modules
- Vite + React 18 + TypeScript
- Docker Compose for local dev

Generate the root README.md, go.mod with module name "github.com/neuralops/platform", and the full docker-compose.yml including:
  - Apache Kafka + Zookeeper
  - PostgreSQL 16
  - ClickHouse
  - Elasticsearch 8
  - Qdrant (vector DB)
  - Redis
  - All backend Go services (one container each)
  - React frontend (nginx)

Each service in docker-compose should have health checks, env vars, and depends_on correctly set.
```

---

# ═══════════════════════════════════════════
# PHASE 1 — BACKEND: DOMAIN MODELS & CORE
# ═══════════════════════════════════════════

```
In the Go backend (backend/internal/domain/), define all domain models as Go structs with JSON tags, validation tags, and database tags.

Required domain models:

1. LogEntry
   - ID (uuid)
   - Timestamp (time.Time)
   - Service (string)
   - Environment (string — prod/staging/dev)
   - Severity (enum: DEBUG/INFO/WARN/ERROR/FATAL/CRITICAL)
   - Message (string)
   - TraceID (string)
   - TxnID (string, for banking transaction tracking)
   - Thread (string)
   - Host (string)
   - Pod (string)
   - Namespace (string)
   - Labels (map[string]string)
   - RawJSON (jsonb)
   - EmbeddingVector ([]float32, for semantic search)
   - ParsedStackTrace (*StackTrace)
   - ClassifiedError (*ErrorClassification)
   - Severity
   - Timestamp

2. StackTrace
   - Language (enum: Java/Golang/Python/NodeJS/DotNet)
   - ExceptionType (string)
   - ExceptionMessage (string)
   - Frames ([]StackFrame)
   - Hotspot (string — identified code hotspot)

3. StackFrame
   - ClassName / Package
   - MethodName / FunctionName
   - FileName
   - LineNumber
   - IsThirdParty (bool)

4. ErrorClassification
   - Category (enum: TIMEOUT/DB_DEADLOCK/MEMORY_LEAK/CONNECTION_EXHAUSTION/DNS_ISSUE/SSL_ISSUE/RETRY_STORM/KAFKA_LAG/DEPLOYMENT_ISSUE/DEPENDENCY_FAILURE/THREAD_STARVATION/CIRCUIT_BREAKER_OPEN/RATE_LIMITING/AUTH_FAILURE/UNKNOWN)
   - Confidence (float64, 0-1)
   - Reasoning (string)

5. Incident
   - ID (uuid)
   - Title (string)
   - Summary (string — AI-generated)
   - Severity (P1/P2/P3/P4)
   - Status (OPEN/INVESTIGATING/RESOLVED/SUPPRESSED)
   - AffectedServices ([]string)
   - RootCauseAnalysis (*RootCause)
   - CorrelatedLogIDs ([]uuid)
   - CorrelatedTraceIDs ([]uuid)
   - CorrelatedDeploymentID (*uuid)
   - BlastRadius ([]string)
   - StartTime, ResolvedTime
   - MTTR (duration)
   - Recommendations ([]Recommendation)
   - Timeline ([]IncidentEvent)

6. RootCause
   - FirstFailingService (string)
   - RootCauseDescription (string)
   - Evidence ([]Evidence)
   - DeploymentCorrelation (*DeploymentCorrelation)
   - InfraCorrelation (*InfraCorrelation)
   - Confidence (float64)

7. Recommendation
   - Type (enum: CONFIG/TIMEOUT/RETRY/AUTOSCALING/DB_TUNING/JVM/CIRCUIT_BREAKER)
   - Description (string)
   - CodeSnippet (string, optional)
   - Priority (HIGH/MEDIUM/LOW)

8. Deployment
   - ID (uuid)
   - Service (string)
   - Version (string)
   - DeployedAt (time.Time)
   - DeployedBy (string)
   - ChangeType (DEPLOYMENT/CONFIG_CHANGE/SCALING/FEATURE_TOGGLE)
   - Environment (string)

9. Alert
   - ID (uuid)
   - Source (PROMETHEUS/DYNATRACE/CLOUD/CUSTOM)
   - Title, Description
   - Severity
   - FiredAt, ResolvedAt
   - Labels (map[string]string)
   - LinkedIncidentID (*uuid)
   - Deduplicated (bool)

10. Transaction (Banking-specific)
    - TxnID (string — UPI/NEFT ref)
    - TxnType (UPI/NEFT/RTGS/IMPS)
    - Status (SUCCESS/FAILED/PARTIAL/PENDING)
    - Hops ([]TxnHop — service journey)
    - TotalLatencyMs (int64)
    - FailedAt (string — which service)
    - RetryCount (int)

11. TxnHop
    - ServiceName, SpanID, TraceID
    - LatencyMs (int64)
    - Status, ErrorMessage

12. Metric
    - ServiceName, Host, Pod
    - MetricType (CPU/MEMORY/LATENCY/THROUGHPUT/DB_CONNECTIONS/THREAD_POOL/GC_PAUSE)
    - Value (float64)
    - Timestamp

13. AnomalyDetection
    - ServiceName, MetricType
    - Score (float64, 0-1)
    - Detected (bool)
    - Baseline, Actual (float64)
    - DetectedAt (time.Time)
    - AnomalyType (SPIKE/DROP/TREND/SEASONAL)

Place all models in backend/internal/domain/models.go.
Create backend/internal/domain/enums.go for all enums as typed constants.
Create backend/internal/domain/errors.go with domain-specific error types.
```

---

# ═══════════════════════════════════════════
# PHASE 2 — BACKEND: INGESTION SERVICE
# ═══════════════════════════════════════════

```
Build the Log Ingestion Service in Go at backend/cmd/ingestion/.

Architecture:
  HTTP server (Gin framework) → Kafka producer → ClickHouse writer

Requirements:

1. HTTP Endpoints (REST):
   POST /api/v1/logs         — single log or batch (JSON array, max 10,000)
   POST /api/v1/metrics      — metrics ingestion
   POST /api/v1/events       — deployment/config events
   POST /api/v1/traces       — distributed trace spans
   GET  /api/v1/health       — health check

2. Kafka Producer:
   - Topic: "raw-logs", "raw-metrics", "raw-events", "raw-traces"
   - Use segmentio/kafka-go library
   - Async batched writes (configurable flush interval, max batch size)
   - Producer error handling with dead letter queue topic "dlq-logs"
   - Message key = service+environment for partition locality

3. Log Parser & Enrichment (before writing to Kafka):
   - Parse Java stack traces (identify exception class, message, frames)
   - Parse Golang panic output
   - Parse Python tracebacks
   - Parse Node.js error stacks
   - Extract traceId / txnId from log message via configurable regex patterns
   - Detect log severity from message if not provided
   - GeoIP enrichment for host IPs (optional)
   - Add ingestion metadata: receivedAt, ingestorID, datacenter

4. Multi-source connectors as separate goroutines:
   - File tail connector (inotify-based)
   - Syslog UDP/TCP receiver (port 5140)
   - Fluentd/Fluent Bit forward protocol receiver (port 24224)
   - HTTP webhook receiver for Dynatrace/Datadog/Prometheus alerts

5. Rate limiting:
   - Per-tenant rate limiter using token bucket (golang.org/x/time/rate)
   - Configurable per plan tier (Startup: 10k/sec, Business: 100k/sec, Enterprise: unlimited)

6. Validation:
   - Required fields: timestamp, service, message
   - Timestamp must be within ±1 hour of now (reject stale/future logs)
   - Max message length: 64KB

7. ClickHouse direct writer (for high-throughput metrics):
   - Use clickhouse-go/v2 driver
   - Async batch insert to "metrics" table
   - Flush every 1 second or 50,000 rows (whichever first)

8. Configuration via Viper:
   - kafka.brokers, kafka.topic_prefix
   - clickhouse.dsn
   - rate_limit.per_tenant
   - parser.stack_trace_languages

Use Uber's zap for structured logging throughout. Include full unit tests for the parser.
```

---

# ═══════════════════════════════════════════
# PHASE 3 — BACKEND: AI ANALYSIS ENGINE
# ═══════════════════════════════════════════

```
Build the AI Analysis Engine in Go at backend/cmd/analysis/.

This service consumes from Kafka, performs AI analysis, and writes results to PostgreSQL + Elasticsearch.

Architecture: Kafka Consumer → Analysis Pipeline → PostgreSQL + Elasticsearch + Qdrant

1. Kafka Consumer (segmentio/kafka-go):
   - Consumer groups: "analysis-errors", "analysis-metrics"
   - Consume from: "raw-logs", "enriched-logs"
   - Configurable concurrency (worker pool pattern)

2. Error Classification Pipeline (Go):
   - Rule-based first pass (regex + keyword matching) for high-confidence categories:
     * SocketTimeoutException → TIMEOUT
     * Deadlock → DB_DEADLOCK
     * OOMKilled / OutOfMemoryError → MEMORY_LEAK
     * connection refused / connection reset → CONNECTION_EXHAUSTION
     * etc.
   - If confidence < 0.8, call LLM API for classification
   - Cache classifications in Redis (key: hash(errorMessage), TTL: 1 hour)

3. LLM API Client (Go HTTP client, NO Python):
   - Support multiple backends via interface:
     type LLMClient interface {
         Classify(ctx context.Context, logMsg string) (*ErrorClassification, error)
         Explain(ctx context.Context, logMsg string) (*LogExplanation, error)
         SummarizeStackTrace(ctx context.Context, trace StackTrace) (*StackSummary, error)
         GenerateRCA(ctx context.Context, incident IncidentContext) (*RootCause, error)
         GenerateRecommendation(ctx context.Context, rca RootCause) ([]Recommendation, error)
         AnswerQuery(ctx context.Context, query string, context []LogEntry) (string, error)
     }
   - Implement OpenAI-compatible backend (works with OpenAI, Azure OpenAI, Ollama)
   - Implement Anthropic Claude backend
   - Implement Ollama self-hosted backend
   - Retry with exponential backoff (3 retries, 2s initial delay)
   - Request timeout: 30s for explanations, 60s for RCA

4. Log Explanation Generator:
   Input: log message + error classification
   Output:
     {
       "plain_english": "The application timed out waiting for a DB response",
       "technical_summary": "...",
       "possible_causes": ["DB overloaded", "network issue", "connection pool exhausted"],
       "recommended_actions": ["Check DB slow query log", "Inspect connection pool metrics"],
       "severity_assessment": "HIGH"
     }
   - Use structured prompting with JSON output
   - Cache by (errorClassification + serviceType) pair

5. Stack Trace Analyzer:
   - Language detection (Java/Go/Python/Node/DotNet)
   - Frame parsing for each language
   - Identify hotspot: first non-library frame
   - Detect recurring exceptions (compare with last 100 from same service)
   - LLM summary of what the code was doing when it failed

6. Embedding Generation (for semantic search):
   - Generate 1536-dim embeddings for each log message
   - Use text-embedding-3-small (OpenAI) or nomic-embed-text (Ollama)
   - Write vectors to Qdrant collection "log_embeddings"
   - Batch embed: 100 logs per API call

7. Anomaly Detection (pure Go, no Python):
   - Maintain rolling statistics per (service, metric) in Redis
   - Z-score based detection: flag if |z| > 3
   - EWMA (Exponentially Weighted Moving Average) for trend detection
   - Seasonality: compare with same hour yesterday, same hour last week
   - Publish anomaly events to Kafka topic "anomalies"

8. Write results back to:
   - PostgreSQL: incidents, error_classifications, recommendations
   - Elasticsearch: enriched logs (with AI fields: classification, explanation, embedding_id)
   - Redis: real-time stats (service error rates, anomaly scores)

Use structured logging (zap) and expose Prometheus metrics:
  - analysis_processed_total (counter, labels: service, classification)
  - analysis_latency_seconds (histogram, labels: analysis_type)
  - llm_api_calls_total (counter, labels: backend, method)
  - llm_api_latency_seconds (histogram)
```

---

# ═══════════════════════════════════════════
# PHASE 4 — BACKEND: CORRELATION & INCIDENT ENGINE
# ═══════════════════════════════════════════

```
Build the Correlation Engine and Incident Engine in Go.

CORRELATION ENGINE (backend/cmd/correlation/):

1. Deployment Correlation:
   - Monitor Kafka topic "raw-events" for deployment events
   - After each deployment, watch for error rate increase in next 15 minutes
   - Correlation logic:
     * Error rate increased > 2x baseline within 10 min of deployment → HIGH correlation
     * Error rate increased 1.5-2x → MEDIUM
   - Store deployment-error correlations in PostgreSQL

2. Service Dependency Graph:
   - Build real-time service call graph from distributed traces
   - Use directed adjacency list: map[string][]DependencyEdge
   - Store in PostgreSQL table "service_dependencies" with edge weights (call count, p99 latency)
   - Update graph every minute via background goroutine
   - Detect blast radius: BFS from first failing service

3. Transaction Correlation (Banking):
   - Group log entries by txnId
   - Build transaction journey: ordered list of services touched
   - Identify failure point: first service in journey with error
   - Track partial success: some hops succeeded before failure
   - Store transaction journeys in ClickHouse (high volume)

4. Temporal Correlation:
   - Sliding window analysis: 5-minute windows
   - Detect: multiple services failing within same window → likely common cause
   - Correlate with infra events: pod restarts, node failures, network issues

INCIDENT ENGINE (backend/cmd/incident/):

1. Incident Creation Rules:
   - P1: Error rate > 50% for any service, or any payment service down
   - P2: Error rate > 20%, or latency p99 > 5s
   - P3: Error rate > 5%, or anomaly score > 0.8
   - P4: Anomaly detected, informational

2. Deduplication:
   - Fingerprint incidents by (rootService, errorCategory, timeWindow)
   - If same fingerprint active incident exists → append, don't create new
   - Suppression rules configurable per tenant

3. AI-Generated Incident Summary:
   - On new incident: call LLM with context (correlated logs, metrics, deployment info)
   - Generate:
     * Executive summary (2-3 sentences for non-technical)
     * Technical RCA
     * Blast radius
     * Recommended immediate actions
   - Store in PostgreSQL incidents table

4. Incident Timeline Builder:
   - Chronological list of events: errors, metric spikes, deployments, alerts
   - Auto-generated narrative: "At 10:05, deployment v2.3.1 was pushed. 3 minutes later..."

5. MTTR Tracker:
   - Record created_at, acknowledged_at, resolved_at
   - Calculate MTTR per service, per team, per period
   - Store historical MTTR for trending

6. REST API endpoints:
   GET    /api/v1/incidents                   — list with filters (status, severity, service, time range)
   GET    /api/v1/incidents/:id               — incident detail with full timeline
   POST   /api/v1/incidents/:id/acknowledge   — acknowledge incident
   POST   /api/v1/incidents/:id/resolve       — resolve with resolution notes
   GET    /api/v1/incidents/:id/recommendations — get AI recommendations
   GET    /api/v1/incidents/:id/timeline      — detailed timeline
   GET    /api/v1/services/dependency-map     — service graph
   GET    /api/v1/transactions/:txnId         — transaction journey

Use PostgreSQL with pgx/v5 driver. Include database migrations using golang-migrate.
```

---

# ═══════════════════════════════════════════
# PHASE 5 — BACKEND: SEARCH SERVICE
# ═══════════════════════════════════════════

```
Build the Search Service in Go at backend/cmd/search/.

1. Full-Text Search (Elasticsearch):
   - Index mapping for logs with fields: timestamp, service, severity, message, traceId, txnId, host, pod, classification
   - Query builder supporting:
     * Full-text (message content)
     * Exact match (service, severity, traceId, txnId)
     * Regex search
     * Time range filter
     * Label/tag filters
   - Pagination with search_after (not offset-based, for performance)
   - Aggregations: error count by service, by severity, by hour

2. Semantic Search (Qdrant vector DB):
   - Convert user query to embedding (same model as ingestion)
   - Qdrant similarity search in "log_embeddings" collection
   - Hybrid search: combine Qdrant results with Elasticsearch BM25
   - Natural language queries like: "payment failures after deployment" 

3. Transaction Search:
   - ClickHouse query for transaction journeys
   - Search by: txnId, txnType, status, date range, failed service

4. AI-Powered Natural Language to Query:
   - Parse queries like: "Show UPI failures yesterday between 10pm and 11pm"
   - Use LLM to extract: service="upi-service", severity="ERROR", timeRange=yesterday(22:00-23:00)
   - Convert to Elasticsearch query DSL

5. Search API endpoints:
   POST /api/v1/search/logs           — structured log search
   POST /api/v1/search/semantic       — semantic/NLP search
   GET  /api/v1/search/trace/:traceId — get all logs for a trace
   GET  /api/v1/search/txn/:txnId     — get transaction journey
   POST /api/v1/search/ai             — AI natural language search

6. Elasticsearch index management:
   - ILM policy: hot (7 days SSD), warm (30 days), cold (90 days), delete (configurable)
   - Index template with correct mappings (disable dynamic mapping)
   - Rollover at 50GB or 7 days

Use olivere/elastic/v7 or elastic/go-elasticsearch/v8 client. Include request validation and error handling.
```

---

# ═══════════════════════════════════════════
# PHASE 6 — BACKEND: API GATEWAY
# ═══════════════════════════════════════════

```
Build the API Gateway in Go at backend/cmd/gateway/ using Gin framework.

This is the single entry point for the React frontend and external API clients.

1. Routing — proxy/aggregate to internal services:
   /api/v1/logs/*       → ingestion service
   /api/v1/metrics/*    → ingestion service
   /api/v1/search/*     → search service
   /api/v1/incidents/*  → incident service
   /api/v1/analysis/*   → analysis service
   /api/v1/alerts/*     → alerting service
   /api/v1/chat/*       → analysis service (AI chat)
   /api/v1/dashboard/*  → aggregated from multiple services

2. Authentication middleware:
   - JWT validation (RS256) for API tokens
   - API key authentication (for ingestion clients)
   - SSO via OIDC (support Okta, Google, Azure AD)
   - Extract tenantID and userID from token, inject to downstream headers

3. Authorization middleware (RBAC):
   Roles: ADMIN, SRE, DEVELOPER, READONLY, ALERT_MANAGER
   - ADMIN: all access
   - SRE: all read + incident management
   - DEVELOPER: read logs/incidents for their services only
   - READONLY: read-only all
   - ALERT_MANAGER: manage alerts/on-call

4. Tenant isolation:
   - All DB queries MUST include tenant_id filter
   - Inject X-Tenant-ID header to all downstream services
   - Middleware validates tenant subscription status

5. Dashboard aggregation endpoint:
   GET /api/v1/dashboard/overview — single endpoint that parallel-fetches:
     * Active incidents (P1/P2)
     * Overall error rate (last 1h)
     * Service health scores
     * Top 5 failing services
     * Recent anomalies
     * Deployment impact score
   Use errgroup for parallel fetching with 3s timeout

6. AI Chat endpoint:
   POST /api/v1/chat/query
   Body: { "question": "Why did payment failures spike yesterday?" }
   - Fetch relevant context: last 24h incidents, error trends, deployment events
   - Call LLM with context + question
   - Stream response using SSE (Server-Sent Events)

7. WebSocket endpoint for real-time dashboard:
   WS /api/v1/ws/realtime
   - Push events: new incidents, anomalies, resolved alerts
   - Use gorilla/websocket
   - Ping/pong keepalive every 30s
   - Per-connection subscriptions: { services: [...], severities: [...] }

8. Rate limiting per tenant using Redis sliding window.

9. Request/Response logging middleware using zap (log: method, path, status, latency, tenantId).

10. Swagger/OpenAPI docs auto-generated using swaggo/swag.

Include CORS configuration for the React frontend.
```

---

# ═══════════════════════════════════════════
# PHASE 7 — BACKEND: ALERTING SERVICE
# ═══════════════════════════════════════════

```
Build the Alerting Service in Go at backend/cmd/alerting/.

1. Alert ingestion from external sources:
   - Webhook receiver for Prometheus Alertmanager (POST /webhook/prometheus)
   - Webhook receiver for Dynatrace Problems API (POST /webhook/dynatrace)
   - Webhook receiver for AWS CloudWatch (POST /webhook/aws)
   - Normalize all alerts to internal Alert domain model

2. Alert processing pipeline:
   a) Deduplication: fingerprint by (source, alertname, service, labels hash)
      - If same fingerprint, update count + last_seen, don't create duplicate
   b) Grouping: group related alerts within 5-minute window, same service
   c) Enrichment: match alert to active incident, add AI explanation
   d) Suppression: maintenance windows, configured silences

3. Notification dispatchers (implement as interface):
   type Notifier interface {
       Send(ctx context.Context, alert Alert, incident *Incident) error
   }
   Implement:
   - EmailNotifier (use net/smtp + HTML templates)
   - SlackNotifier (Slack Incoming Webhooks + Block Kit)
   - TeamsNotifier (Microsoft Teams Adaptive Cards)
   - PagerDutyNotifier (PagerDuty Events API v2)
   - SMSNotifier (Twilio API)
   - WebhookNotifier (generic HTTP POST)

4. Escalation policies:
   - Define escalation chains per service/team
   - If P1 not acknowledged in 5 min → escalate to L2
   - If not acknowledged in 15 min → escalate to manager
   - Track acknowledgment via POST /api/v1/alerts/:id/acknowledge

5. On-call schedule integration:
   - Simple on-call rotation config (JSON)
   - Determine current on-call person for routing

6. REST API:
   GET    /api/v1/alerts                    — list alerts with filters
   GET    /api/v1/alerts/:id               — alert detail
   POST   /api/v1/alerts/:id/acknowledge   — ack
   POST   /api/v1/alerts/:id/suppress      — suppress with duration
   GET    /api/v1/alerts/rules             — list alert rules
   POST   /api/v1/alerts/rules             — create alert rule
   GET    /api/v1/notifications/channels   — list notification channels
   POST   /api/v1/notifications/channels   — create channel (Slack/Email/etc)

Use PostgreSQL for alert storage. Include database migrations.
```

---

# ═══════════════════════════════════════════
# PHASE 8 — DATABASE SCHEMAS
# ═══════════════════════════════════════════

```
Create all database migrations using golang-migrate format at backend/migrations/.

POSTGRESQL TABLES (primary transactional store):

001_create_tenants.sql:
- tenants (id uuid PK, name, plan tier, created_at, settings jsonb)
- users (id uuid PK, tenant_id FK, email, role, sso_sub, created_at)

002_create_incidents.sql:
- incidents (id, tenant_id, title, summary, severity, status, affected_services[], root_cause jsonb, recommendations jsonb[], timeline jsonb[], blast_radius[], deployment_id, created_at, acknowledged_at, resolved_at, mttr_seconds)
- incident_events (id, incident_id, event_type, description, metadata jsonb, occurred_at)
- incident_services (incident_id, service_name, role: PRIMARY/DOWNSTREAM)

003_create_deployments.sql:
- deployments (id, tenant_id, service, version, environment, deployed_by, deployed_at, change_type, metadata jsonb)
- deployment_incident_correlations (deployment_id, incident_id, correlation_strength float, detected_at)

004_create_alerts.sql:
- alerts (id, tenant_id, source, title, description, severity, status, labels jsonb, fired_at, resolved_at, incident_id, fingerprint, occurrence_count)
- alert_rules (id, tenant_id, name, condition jsonb, threshold, window_seconds, severity, enabled)
- notification_channels (id, tenant_id, type, name, config jsonb, enabled)
- escalation_policies (id, tenant_id, service, steps jsonb)

005_create_service_graph.sql:
- service_dependencies (id, tenant_id, source_service, target_service, call_count, p50_latency_ms, p99_latency_ms, error_rate, updated_at)
- service_health (id, tenant_id, service_name, health_score float, error_rate float, latency_p99 float, anomaly_score float, updated_at)

006_create_error_classifications.sql:
- error_patterns (id, tenant_id, pattern_hash, category, confidence, sample_message, first_seen, last_seen, occurrence_count, llm_explanation jsonb)

CLICKHOUSE TABLES (time-series analytics):
- logs (tenant_id, timestamp, service, environment, severity, message, trace_id, txn_id, host, pod, classification, explanation jsonb) — ENGINE = MergeTree() PARTITION BY toYYYYMM(timestamp)
- metrics (tenant_id, timestamp, service, host, metric_type, value) — ENGINE = MergeTree()
- transactions (tenant_id, txn_id, txn_type, status, hops jsonb, total_latency_ms, failed_at, created_at) — ENGINE = MergeTree()
- anomalies (tenant_id, timestamp, service, metric_type, score, baseline, actual, anomaly_type) — ENGINE = MergeTree()

ELASTICSEARCH INDEX TEMPLATES:
- logs-{YYYY.MM.dd} with mappings for all LogEntry fields
- Disable dynamic mapping, set keyword for high-cardinality fields
- Enable _source compression

QDRANT COLLECTIONS:
- log_embeddings: vector size=1536, distance=Cosine, payload: {log_id, service, timestamp, severity}

Include both .up.sql and .down.sql for each migration.
Add indexes: tenant_id+timestamp, service+severity+timestamp, trace_id, txn_id, fingerprint.
```

---

# ═══════════════════════════════════════════
# PHASE 9 — FRONTEND: REACT SETUP & DESIGN SYSTEM
# ═══════════════════════════════════════════

```
Set up the React frontend for NeuralOps at frontend/ with a cutting-edge design system.

TECH STACK:
- React 18 + TypeScript
- Vite 5
- TanStack Query v5 (data fetching + caching)
- TanStack Router (file-based routing)
- Zustand (global state)
- Recharts (charts/graphs)
- React Flow (service dependency map)
- Framer Motion (animations)
- Tailwind CSS v4
- Radix UI primitives (accessible components)
- date-fns (date handling)
- Zod (validation)
- react-hot-toast (notifications)

DESIGN SYSTEM — "Neural Dark" aesthetic:
Define in src/styles/design-tokens.css:

Colors:
  --bg-base: #070B14          /* deepest background */
  --bg-surface: #0D1424       /* card surfaces */
  --bg-elevated: #131B2E      /* modals, dropdowns */
  --bg-highlight: #1A2540     /* hover states */
  --border-subtle: #1E2D47    /* subtle borders */
  --border-strong: #2A3D5C    /* strong borders */
  
  --accent-primary: #3B82F6   /* electric blue */
  --accent-secondary: #6366F1  /* indigo */
  --accent-glow: rgba(59,130,246,0.15)
  
  --severity-critical: #EF4444
  --severity-p1: #F97316
  --severity-p2: #F59E0B
  --severity-p3: #3B82F6
  --severity-p4: #6B7280
  
  --status-healthy: #10B981
  --status-degraded: #F59E0B
  --status-down: #EF4444
  --status-unknown: #6B7280
  
  --text-primary: #F1F5F9
  --text-secondary: #94A3B8
  --text-muted: #475569
  
  --success: #10B981
  --warning: #F59E0B
  --error: #EF4444
  --info: #3B82F6

Typography:
  --font-display: 'Space Grotesk', sans-serif    /* headings */
  --font-mono: 'JetBrains Mono', monospace       /* log lines, code */
  --font-body: 'Inter', sans-serif               /* body text */

Spacing scale: 4px base unit (4, 8, 12, 16, 24, 32, 48, 64)

Effects:
  --glow-blue: 0 0 20px rgba(59,130,246,0.3)
  --glow-red: 0 0 20px rgba(239,68,68,0.3)
  --glass: backdrop-filter: blur(12px); background: rgba(13,20,36,0.8)

Build these base components in src/components/ui/:

1. Badge — variants: critical/p1/p2/p3/p4/healthy/degraded/down/info
2. Card — glassmorphism style, subtle border, hover glow
3. Metric card — large number, trend indicator, sparkline
4. Status dot — animated pulse for live/critical states
5. Severity indicator — colored left-border bar
6. Timeline item — dot + line + content
7. Code block — JetBrains Mono, syntax-highlighted log lines, copy button
8. Skeleton loaders — shimmer animation for all card types
9. Empty state — illustrated empty states with CTA
10. Button — primary/secondary/ghost/danger variants
11. SearchInput — with keyboard shortcut badge (⌘K)
12. Select, DateRangePicker, MultiSelect
13. Tooltip, Popover (Radix based)
14. Modal/Dialog (Radix based)
15. DataTable — sortable, filterable, virtual scroll for large datasets

Create src/api/client.ts with Axios instance:
- Base URL from env
- JWT token injection
- Request/response interceptors
- Auto-refresh token on 401
- Error normalization

Create src/store/ with Zustand stores:
- useAuthStore (user, tenant, token)
- useRealtimeStore (websocket connection, live incidents)
- useFilterStore (global log filters: service, severity, time range)
```

---

# ═══════════════════════════════════════════
# PHASE 10 — FRONTEND: MAIN DASHBOARD PAGE
# ═══════════════════════════════════════════

```
Build the Main Dashboard page at src/pages/Dashboard.tsx.

This is the primary screen SREs see on login. Design it with maximum information density while maintaining clarity.

LAYOUT: 3-column grid, responsive to 2-col on medium, 1-col on mobile.

TOP BAR (sticky):
- NeuralOps logo (neural network icon in electric blue)
- Global environment selector: ALL / PROD / STAGING / DEV
- Global time range picker: Last 1h / 6h / 24h / 7d / Custom
- Search bar (opens command palette on click — ⌘K)
- Notification bell with unread count
- User avatar + tenant name

SIDEBAR (collapsible, 240px):
Icons + labels for:
- Dashboard (grid icon)
- Incidents (flame icon)
- Log Explorer (terminal icon)
- Trace Explorer (git-branch icon)
- Transaction Journey (route icon)
- Anomaly Detection (activity icon)
- Service Map (share-2 icon)
- AI Assistant (cpu icon)
- Alerts (bell icon)
- Settings (settings icon)

Active state: electric blue left border + subtle blue bg

MAIN CONTENT — Dashboard widgets:

ROW 1 — KPI Cards (4 across):
1. Active Incidents: large number (color: red if P1/P2 active), breakdown P1/P2/P3/P4 dots
2. Error Rate: percentage + trend arrow + sparkline (last 24h)
3. Avg Latency p99: milliseconds + trend + sparkline
4. MTTR Today: duration format + vs yesterday comparison

ROW 2 — Wide + Narrow:
Left (60%): "Incident Timeline" — horizontal scrollable timeline showing last 24h of incidents as colored bars, deployments as vertical dashed lines. Click any bar to open incident panel.
Right (40%): "Top Failing Services" — ranked list with:
  - Service name + health indicator dot
  - Error rate bar (colored by severity)
  - Latency p99
  - Trend arrow

ROW 3 — Full width:
"Error Rate by Service" — stacked area chart (Recharts), last 24h, each service a different color. Hover shows breakdown tooltip.

ROW 4 — 3 equal columns:
Left: "Active Anomalies" — list of detected anomalies with score bars, animated pulse for new ones
Center: "Recent Deployments" — timeline with deployment name, version, deployer, correlation risk badge
Right: "AI Insights" — auto-generated bullets: "UPI service error rate up 23% — possible connection exhaustion"

Real-time updates via WebSocket:
- New incident notification: toast + animate incident count
- Resolved incident: green toast
- New anomaly: pulse animation on anomaly widget

Use TanStack Query for all data fetching with 30s refetch interval for dashboard data.
Use Framer Motion for:
- Page load stagger animation
- KPI card number counting animation
- Incident timeline scroll

Make the dashboard feel like a premium Bloomberg Terminal meets Vercel Analytics.
```

---

# ═══════════════════════════════════════════
# PHASE 11 — FRONTEND: LOG EXPLORER PAGE
# ═══════════════════════════════════════════

```
Build the Log Explorer page at src/pages/LogExplorer.tsx.

This is a Kibana/Datadog-style log viewer but with AI superpowers.

LAYOUT:
- Left panel (300px): filters sidebar
- Main panel: log stream + search bar
- Right panel (400px, slide-in): log detail

LEFT PANEL — Filters:
- Service multi-select (with service health dots)
- Severity checkboxes (DEBUG/INFO/WARN/ERROR/FATAL) with count badges
- Environment pills
- Time range picker
- Host / Pod filter
- "Has stack trace" toggle
- "Has AI explanation" toggle
- Quick filters: "Show only errors", "Show only my services"

MAIN PANEL:

Search bar (full width, prominent):
- Placeholder: "Search logs... or try 'payment failures after 10pm' (AI)"
- Toggle button: [Text] [Regex] [AI/NLP]
- Keyboard shortcuts shown inline

Log stream table (virtual scroll — must handle 100,000 rows):
Columns:
  1. Severity badge (color-coded, icon)
  2. Timestamp (relative + absolute on hover)
  3. Service (pill, clickable to filter)
  4. Message (truncated, with highlighted search matches)
  5. Trace ID (monospace, copy button on hover)
  6. AI icon (if explanation available, click to show)

Log line styling:
- FATAL/CRITICAL: red left border + subtle red bg tint
- ERROR: orange left border
- WARN: yellow left border
- INFO: no accent
- DEBUG: muted gray

Stack trace lines: monospace font, indented, collapsible

"Live tail" toggle: auto-scroll to new logs, pause on hover

LOG DETAIL PANEL (right slide-in):
When user clicks a log line, show:
1. Full raw JSON (collapsible)
2. "AI Explanation" section:
   - Plain English explanation
   - Possible causes (bulleted)
   - Recommended actions (bulleted)
3. Stack Trace Analysis (if present):
   - Parsed frames in readable format
   - Hotspot highlighted in yellow
   - "Recurring?" badge (seen X times in last hour)
4. Related logs:
   - Same traceId (show trace journey)
   - Same txnId (show transaction journey)
   - Similar errors (semantic similarity)
5. Linked incident (if any)
6. "Open in AI Chat" button

Bottom bar:
- Total results count
- Query performance (Xs)
- Export options: JSON / CSV / Share link

Use windowing (react-virtual or TanStack Virtual) for the log table.
Animate the right panel sliding in with Framer Motion.
```

---

# ═══════════════════════════════════════════
# PHASE 12 — FRONTEND: INCIDENT DETAIL PAGE
# ═══════════════════════════════════════════

```
Build the Incident Detail page at src/pages/IncidentDetail.tsx.

This page is what an SRE opens during an outage. It needs to be dense with information and actionable.

LAYOUT: Full-page, no sidebar (focus mode)

TOP SECTION:
- Breadcrumb: Incidents > [Incident Title]
- Incident title (large, bold)
- Severity badge (P1/P2/P3/P4) + Status badge (OPEN/INVESTIGATING/RESOLVED)
- Time: "Opened 23 minutes ago · Unacknowledged"
- CTA buttons: [Acknowledge] [Resolve] [Escalate] [Share]

TABS below top section:
[Overview] [Timeline] [Logs] [Traces] [Metrics] [AI Analysis] [Recommendations]

OVERVIEW TAB:
Left 60%:
- "AI Root Cause Analysis" card:
  * Animated typing effect as AI summary appears
  * First failing service (highlighted in red)
  * Blast radius list (affected downstream services)
  * Confidence score progress bar
  * Evidence items (list of correlated log snippets, deployment, metric spike)
- Deployment correlation box (if detected):
  "Deployment v2.3.1 pushed at 10:05 PM — 94% correlation with this incident"

Right 40%:
- Incident stats: MTTR clock (counting up), affected services count, error count
- Service health cards for each affected service
- On-call information: who is on-call, contact buttons

TIMELINE TAB:
Vertical timeline (similar to a git log):
- Each event: timestamp, icon (deployment/error/metric/alert), description
- Color-coded: deployments=purple, errors=red, metric=orange, alerts=yellow
- Filter timeline: show/hide event types
- "Mark as root cause" action on any event

LOGS TAB:
Embedded log explorer filtered to incident time range + affected services.
Show only ERROR/FATAL by default, toggle to show all.

METRICS TAB:
Side-by-side Recharts for each affected service:
- Error rate (red area chart)
- Latency p99 (orange line)
- CPU/Memory (if anomaly detected)
Vertical red dashed line at incident start time.

AI ANALYSIS TAB:
Full RCA in readable format:
1. What happened (executive summary)
2. Technical root cause (detailed)
3. Contributing factors
4. Why it propagated (cascade explanation)
5. What prevented faster detection (gap analysis)

RECOMMENDATIONS TAB:
Actionable cards, each with:
- Priority badge (HIGH/MEDIUM/LOW)
- Recommendation title
- Detailed description
- Code/config snippet (if applicable, in a copy-able code block)
- "Mark as Applied" checkbox

All data loaded via TanStack Query.
WebSocket subscription for live updates while incident is OPEN.
Keyboard shortcut: A to acknowledge, R to resolve, Esc to go back.
```

---

# ═══════════════════════════════════════════
# PHASE 13 — FRONTEND: SERVICE DEPENDENCY MAP
# ═══════════════════════════════════════════

```
Build the Service Dependency Map at src/pages/ServiceMap.tsx using React Flow.

This is a live, animated visualization of the microservice topology.

VISUAL DESIGN:
- Dark canvas with subtle grid background
- Services as rectangular nodes with rounded corners
- Color coding based on health: green/yellow/red border glow
- Animated flowing lines between services for active traffic
- Line thickness = relative traffic volume
- Line color: green=healthy, yellow=elevated latency, red=errors flowing

NODE DESIGN (custom React Flow node):
Each service node shows:
- Service name (bold)
- Health status dot (animated pulse if degraded/down)
- Error rate percentage
- p99 latency (ms)
- Requests/sec

Mini sparkline in bottom of node for error rate trend.

EDGE LABELS:
On hover, show edge tooltip:
- Calls/sec
- p99 latency
- Error rate on this edge
- "View traces" link

CONTROLS:
Top-left panel:
- [Zoom In] [Zoom Out] [Fit to Screen] [Export PNG]
- Layout toggle: [Layered] [Force-Directed] [Circular]
- Filter: show only services with errors

Top-right:
- Time range for the data
- "Highlight critical path" button — traces path from entry to most-erroring service

INCIDENT OVERLAY:
If active incident selected:
- Blast radius services glow red
- First failing service has pulsing red ring
- Arrow showing propagation direction
- Error count badges on affected edges

CLICK BEHAVIOR:
Click node → slide-in panel with service details:
- Full metric charts
- Recent incidents
- Deployments
- Dependencies list
- "View logs" shortcut

Use React Flow 11+ with custom node components.
Animate edges with animated SVG dash offset (traffic flow effect).
Fetch graph data every 60 seconds (auto-refresh).
```

---

# ═══════════════════════════════════════════
# PHASE 14 — FRONTEND: AI CHAT ASSISTANT PAGE
# ═══════════════════════════════════════════

```
Build the AI Chat Assistant page at src/pages/AIAssistant.tsx.

This is like having a senior SRE colleague who knows everything about your system.

LAYOUT:
Full-height split:
- Left 30%: conversation history + suggested questions
- Right 70%: active conversation

LEFT PANEL:
"Recent Conversations" list (each with timestamp + first question truncated)
"Suggested Questions" section with example queries grouped by category:
  Incidents: "What caused the payment failures last night?"
  Performance: "Which service has the worst latency this week?"
  Trends: "Is our error rate improving over the past month?"
  Banking: "Show me all failed UPI transactions today"
  Deployments: "Did any deployments cause incidents this week?"

Click a suggested question → populates input and sends

RIGHT PANEL — Chat interface:

Messages:
- User messages: right-aligned, blue bubble
- AI messages: left-aligned, dark card with subtle blue left border
- AI avatar: neural network icon

AI message formatting:
- Markdown rendering (headers, bold, lists, code blocks)
- Embedded mini log table (when showing log results)
- Embedded mini chart (when showing metrics over time)
- Clickable service names → navigate to service page
- Clickable incident IDs → open incident detail

"Sources used" expandable section below each AI message:
Shows what data AI analyzed: "Analyzed 1,247 logs, 3 incidents, 1 deployment from last 24h"

Input area (bottom):
- Textarea with auto-resize
- Send button with Ctrl+Enter shortcut
- "Include context" toggles: [Last 24h] [All Services] [Production Only]
- Character/token count indicator

Streaming response:
- Use SSE (Server-Sent Events) to stream AI response word by word
- Show typing indicator (animated dots) while waiting
- "Stop generating" button during streaming

Loading state:
- "Searching logs..." → "Analyzing patterns..." → "Generating response..."
- These messages update as backend processes

Example conversation flow:
User: "Why did payment failures increase yesterday evening?"
AI: (streaming) "Based on my analysis of your systems...
    📊 I found 2,847 payment failures between 9:45 PM and 11:30 PM
    
    **Root Cause:** Deployment v2.3.1 of UPI Routing Service at 9:42 PM introduced 
    a blocking DB call in the payment flow...
    
    **Evidence:**
    - Error rate increased from 0.2% to 34% within 8 minutes of deployment
    - Thread pool exhaustion detected on upi-service pods (all 3 pods hit max threads)
    - Downstream services (Payment API, Notification Service) failed as a result
    
    **Recommendation:** Rollback v2.3.1 or apply the connection pool fix in PR #4521"

This page should feel like ChatGPT meets Datadog AI.
```

---

# ═══════════════════════════════════════════
# PHASE 15 — FRONTEND: TRANSACTION JOURNEY PAGE
# ═══════════════════════════════════════════

```
Build the Transaction Journey page at src/pages/TransactionJourney.tsx.

Critical for banking use cases — visualize every hop of a financial transaction.

LAYOUT:
- Top: search bar for txnId / traceId
- Middle: transaction flow visualization
- Bottom: hop-level log details

SEARCH:
Large centered input with placeholder "Enter Transaction ID, UPI Ref, NEFT UTR..."
Recent searches shown below (from localStorage)

TRANSACTION FLOW VISUALIZATION:
Horizontal swim-lane diagram showing transaction hops:

[API Gateway] → [Auth Service] → [UPI Service] → [CBS/Ledger] → [Notification]
   12ms              8ms             145ms           234ms           18ms

Each service box shows:
- Service name
- Latency (color-coded: green <100ms, yellow 100-500ms, red >500ms)
- Status: ✓ SUCCESS / ✗ FAILED / ⚠ PARTIAL
- Retry count (if retried)

Failed hop:
- Red border + red background tint
- Error message shown below box
- "First failure" indicator arrow

Arrow between hops:
- Color = status of the call
- Shows direction of call

SUMMARY CARDS (above visualization):
- Total TxnID
- Type (UPI/NEFT/RTGS)
- Final Status (large colored badge)
- Total Latency
- Failure Point

HOP DETAIL TABLE (below visualization):
Click any hop → expand inline:
- Full log lines from that service for this txnId
- Span ID, Trace ID (link to trace explorer)
- Request/Response payload (masked if PAN/account data)
- Timestamps (received, processed, responded)

FILTERS:
Filter failed transactions by: service, txn type, time range, error category

Animate the flow diagram:
- Draw arrows sequentially with CSS animation (left to right)
- Pulse the failed hop with red glow

Handle edge cases:
- Transaction not found: "No transaction found for ID xxx"
- Partial data: show available hops with "Data unavailable" for missing ones
```

---

# ═══════════════════════════════════════════
# PHASE 16 — FIGMA-EQUIVALENT UI SPECIFICATION
# ═══════════════════════════════════════════

```
Implement a complete Figma-equivalent design specification as a React component at src/pages/DesignSystem.tsx.

This is a living style guide / component gallery.

PAGE SECTIONS:

1. COLORS
   Display all design tokens as swatches with hex values and variable names.
   Show: backgrounds, surfaces, accents, severity colors, status colors, text colors.

2. TYPOGRAPHY
   Show all type styles:
   - Display XL: 48px Space Grotesk Bold — "NeuralOps Platform"
   - Display L: 36px Space Grotesk Bold
   - Heading 1: 28px Space Grotesk SemiBold
   - Heading 2: 22px Space Grotesk SemiBold
   - Heading 3: 18px Space Grotesk Medium
   - Body Large: 16px Inter Regular
   - Body: 14px Inter Regular
   - Body Small: 12px Inter Regular
   - Label: 12px Inter SemiBold uppercase
   - Mono: 13px JetBrains Mono Regular
   Each with line height, letter spacing specs.

3. SPACING & GRID
   Visual spacing scale: 4px increments
   Column grid example: 12-col, 24px gutter, 32px margin

4. ICONS
   Lucide React icon grid for all used icons with names

5. COMPONENTS GALLERY:
   
   5a. Badges & Status Indicators:
       - All severity badges (CRITICAL/P1/P2/P3/P4)
       - Status dots (healthy/degraded/down) with pulse animation
       - Environment pills (PROD/STAGING/DEV)
   
   5b. Cards:
       - Metric card (with sparkline)
       - Incident card
       - Alert card
       - Service health card
   
   5c. Data Display:
       - Log line variants (each severity)
       - Code block with copy button
       - Stack trace display
       - Timeline event
   
   5d. Charts:
       - Area chart (error rate over time)
       - Bar chart (error count by service)
       - Sparkline (inline metric)
       - Anomaly score gauge
   
   5e. Navigation:
       - Sidebar (collapsed + expanded state)
       - Breadcrumb
       - Tab bar
   
   5f. Forms & Inputs:
       - Search input
       - Select dropdown
       - Date range picker
       - Toggle/Switch
       - Checkbox group
   
   5g. Feedback:
       - Toast notifications (success/error/warning/info)
       - Empty state illustrations
       - Loading skeleton
       - Error boundary fallback
       - Progress bar (AI analysis progress)

6. LAYOUT PATTERNS:
   - Dashboard grid example
   - Log explorer split-panel
   - Incident detail full page

Include a "Copy Tailwind" button on each component to copy class names.
Add a theme toggle in the top-right (Dark [default] / Light).

This page should be comprehensive enough to hand to any developer joining the team.
```

---

# ═══════════════════════════════════════════
# PHASE 17 — ANOMALY DETECTION DASHBOARD
# ═══════════════════════════════════════════

```
Build the Anomaly Detection dashboard at src/pages/AnomalyDetection.tsx.

LAYOUT:
- Top: anomaly score overview + active anomaly count
- Center: metric panels with anomaly visualization
- Bottom: anomaly history table

ANOMALY SCORE GAUGES (top row):
For each environment:
- Circular gauge (0-100 "neural anomaly score")
- Color transitions: green (0-30) → yellow (30-60) → red (60-100)
- Animated fill
- "Normal / Elevated / Critical" label

SERVICE ANOMALY LIST:
Cards for each service showing:
- Service name + health dot
- Anomaly type (SPIKE/DROP/TREND/SEASONAL)
- Detected metric (CPU/LATENCY/ERROR_RATE/etc)
- "Baseline: 45ms | Actual: 230ms | +411%"
- When detected (relative time)
- Status badge: NEW / INVESTIGATING / EXPECTED / FALSE_POSITIVE

Click card → expand with:
- Time-series chart with anomaly period highlighted in red
- Baseline band (shaded gray)
- Anomaly window (red background)

ANOMALY DRILL-DOWN CHART:
Large chart area (Recharts):
- Multi-line for multiple services
- Anomaly markers (red dot + tooltip)
- Toggle services on/off
- Zoom on anomaly period button

HISTORY TABLE:
Sortable/filterable table of past anomalies:
Columns: Time | Service | Metric | Score | Deviation | Duration | Outcome (True/False Positive)
Allow marking anomalies as "expected" (scheduled maintenance, known pattern) to improve ML model.

Real-time anomaly feed (right sidebar):
Live stream of anomaly events as they're detected.
Each with small sparkline and score badge.
Auto-scroll with pause on hover.
```

---

# ═══════════════════════════════════════════
# PHASE 18 — SECURITY, AUTH & MULTI-TENANCY
# ═══════════════════════════════════════════

```
Implement security and multi-tenancy across the Go backend.

1. JWT Authentication (backend/internal/middleware/auth.go):
   - RS256 JWT validation using golang-jwt/jwt/v5
   - Extract claims: userId, tenantId, email, role, plan
   - Inject into request context
   - Support both Authorization: Bearer header and X-API-Key header
   - API keys stored hashed (SHA-256) in PostgreSQL with tenant association

2. OIDC/SSO Integration:
   - Use coreos/go-oidc/v3
   - Support Okta, Google Workspace, Azure AD
   - PKCE flow
   - After OIDC callback: create/update user, issue internal JWT
   - SAML support: implement SP-initiated flow using crewjam/saml

3. RBAC Middleware (backend/internal/middleware/rbac.go):
   Define permission matrix:
   
   | Permission           | ADMIN | SRE | DEVELOPER | READONLY | ALERT_MANAGER |
   |---------------------|-------|-----|-----------|----------|---------------|
   | logs.read           | ✓     | ✓   | ✓ (own)   | ✓        | -             |
   | incidents.read      | ✓     | ✓   | ✓ (own)   | ✓        | -             |
   | incidents.write     | ✓     | ✓   | -         | -        | -             |
   | alerts.manage       | ✓     | ✓   | -         | -        | ✓             |
   | settings.write      | ✓     | -   | -         | -        | -             |
   
   Implement as middleware: checkPermission(permission string) gin.HandlerFunc

4. Tenant isolation:
   - All repository methods accept tenantID as first parameter
   - Never query without WHERE tenant_id = $1
   - Separate Elasticsearch index prefix per tenant: "tenant-{id}-logs-*"
   - Separate Qdrant payload filter per tenant

5. Data masking (for PII/PAN compliance):
   Implement in ingestion pipeline (Go):
   - PAN masking: regex [0-9]{12,19} → first 6 + ******* + last 4
   - Aadhaar masking: XXXX-XXXX-1234
   - Account number masking
   - Email masking: j***@gmail.com
   - Phone masking: +91-98****1234
   - Configurable masking rules per tenant

6. Audit logging:
   - Every API call logged to audit table: userId, tenantId, action, resourceType, resourceId, ip, userAgent, timestamp, result
   - Immutable (insert-only)
   - Stored in PostgreSQL + replicated to ClickHouse for analytics

7. Encryption:
   - All secrets (notification channel configs, OIDC client secrets) encrypted at rest using AES-256-GCM
   - Encryption key from environment (support Vault/KMS in production)
   - TLS termination at gateway (mTLS for internal service communication)
```

---

# ═══════════════════════════════════════════
# PHASE 19 — OBSERVABILITY OF THE PLATFORM ITSELF
# ═══════════════════════════════════════════

```
Add self-observability to the NeuralOps platform (dogfooding — the platform monitors itself).

1. Prometheus metrics exposition:
   Every Go service exposes GET /metrics in Prometheus format.
   
   Standard metrics (use prometheus/client_golang):
   - HTTP request duration histogram (labels: method, path, status, service)
   - HTTP requests total counter
   - Active goroutines gauge
   - Memory usage gauges (heap, stack, GC)
   - Database connection pool stats
   - Kafka consumer lag per topic/partition
   
   Service-specific metrics (as defined in each service phase above).

2. Distributed tracing:
   Instrument every service with OpenTelemetry Go SDK.
   - Auto-instrument: Gin HTTP, Kafka, PostgreSQL, Elasticsearch calls
   - Export to Jaeger (or OTLP endpoint)
   - Propagate trace context via W3C TraceContext headers
   - Sample rate: 10% in production, 100% in dev

3. Structured logging everywhere:
   Use uber-go/zap in all services.
   Standard fields on every log:
   - service, version, environment, host, traceId, spanId, userId, tenantId
   - Log levels: configurable per service via env var

4. Health check endpoints:
   Every service: GET /health
   Returns:
   {
     "status": "healthy" | "degraded" | "down",
     "version": "1.2.3",
     "uptime_seconds": 3600,
     "checks": {
       "postgres": "healthy",
       "kafka": "healthy",
       "elasticsearch": "healthy"
     }
   }

5. Grafana dashboards (as JSON provisioning files in infra/grafana/dashboards/):
   - NeuralOps Platform Overview (all services health)
   - Kafka Consumer Lag
   - AI API Usage & Costs
   - Log Ingestion Rate
   - Search Latency
   - Incident Engine Throughput
```

---

# ═══════════════════════════════════════════
# PHASE 20 — MVP TESTING & SEED DATA
# ═══════════════════════════════════════════

```
Create comprehensive testing and seed data for the NeuralOps platform.

1. Go Unit Tests (use standard testing + testify):
   - backend/internal/domain/ — model validation tests
   - backend/cmd/ingestion/ — parser tests (Java/Go/Python/Node stack traces)
   - backend/cmd/analysis/ — error classification tests (rule-based)
   - backend/cmd/correlation/ — deployment correlation logic tests
   - backend/cmd/incident/ — deduplication fingerprint tests
   - Test coverage target: 80% for core business logic

2. Integration Tests:
   Use testcontainers-go to spin up PostgreSQL/Redis/Elasticsearch in tests.
   Test: full log ingestion → enrichment → search flow.
   Test: incident creation → deduplication → notification dispatch.

3. Seed Data Generator (backend/scripts/seed/main.go):
   Generate realistic banking observability data:
   
   Services to seed: api-gateway, auth-service, upi-service, payment-api, ledger-service, notification-service, cbs-adapter, fraud-service
   
   Generate:
   a) 50,000 normal log entries across all services (last 7 days)
   b) Simulate "UPI outage incident":
      - Deployment v2.3.1 of upi-service at T=0
      - Normal logs until T+3min
      - Gradual SocketTimeoutException increase (T+3 to T+5)
      - Thread pool exhaustion errors (T+5 to T+8)
      - Payment API cascade failures (T+8 to T+12)
      - 1,247 failed UPI transactions
      - Auto-resolved at T+45min (rollback)
   c) 30 historical incidents (mix of severities, some resolved quickly, some long)
   d) 100 deployments (60 clean, 30 caused minor issues, 10 caused incidents)
   e) 500 banking transactions (400 success, 100 failures at various hops)
   
   Output: SQL insert scripts + Elasticsearch bulk index + ClickHouse CSV.

4. Frontend E2E Tests (Playwright):
   - Login flow
   - Dashboard loads without errors
   - Log search returns results
   - Incident detail page loads
   - AI chat sends message and gets response

5. Load test script (backend/scripts/loadtest/main.go):
   - Simulate 100k logs/sec ingestion for 60 seconds
   - Measure: ingestion latency p99, Kafka lag, ClickHouse insert rate
   - Print report at end

6. Demo mode:
   Add DEMO_MODE=true env flag that:
   - Uses seed data instead of real data
   - Shows a "Demo Mode" banner in UI
   - Disables actual alert sending
   This makes it easy to show to potential customers.

Create a scripts/quickstart.sh that:
1. Checks prerequisites (Docker, Go, Node)
2. Runs docker-compose up -d
3. Waits for services to be healthy
4. Runs migrations
5. Seeds demo data
6. Starts all services
7. Opens browser to http://localhost:3000
8. Prints login credentials: demo@neuralops.ai / Demo@1234
```

---

## TECHNOLOGY VERSIONS REFERENCE

| Component | Version | Notes |
|-----------|---------|-------|
| Go | 1.22+ | Required for range-over-func |
| Gin | v1.9+ | HTTP framework |
| segmentio/kafka-go | v0.4+ | Kafka client |
| pgx/v5 | v5.5+ | PostgreSQL driver |
| clickhouse-go/v2 | v2.20+ | ClickHouse driver |
| go-elasticsearch/v8 | v8.x | Elasticsearch |
| qdrant-go | latest | Vector DB |
| golang-jwt/jwt/v5 | v5+ | JWT |
| uber-go/zap | v1.27+ | Logging |
| viper | v1.18+ | Config |
| OpenTelemetry Go | v1.24+ | Tracing |
| prometheus/client_golang | v1.19+ | Metrics |
| React | 18.x | Frontend |
| TypeScript | 5.x | Type safety |
| Vite | 5.x | Build tool |
| TanStack Query | v5 | Data fetching |
| TanStack Router | v1 | Routing |
| Zustand | v4 | State |
| Framer Motion | v11 | Animations |
| React Flow | v11 | Service map |
| Recharts | v2.x | Charts |
| Tailwind CSS | v4 | Styling |
| Radix UI | latest | Primitives |

## CURSOR TIPS

1. **Work phase by phase** — don't paste everything at once.
2. **After each phase**, verify the service compiles: `cd backend && go build ./...`
3. **After Phase 9**, verify frontend runs: `cd frontend && npm run dev`
4. **Use `@Codebase`** in Cursor to let it understand the whole project context.
5. **For complex AI prompts** (LLM client, RCA generation), ask Cursor to generate the prompt templates separately first.
6. **Docker Compose** — always run `docker-compose up -d` before testing backend services.
7. **Migrations** — run `make migrate-up` after each schema phase.

---

*Generated for NeuralOps — AI-Powered Observability Platform*
*Based on AI_Log_Analyzer_PRS_v2 | All backend in Go | UI in React + TypeScript*
