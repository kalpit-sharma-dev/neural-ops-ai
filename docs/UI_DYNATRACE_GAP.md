# NeuralOps UI vs Dynatrace — Feature Gap Analysis

**Version:** 1.2  
**Last updated:** 2026-05-31 (vendor parity pass — Pyroscope/eBPF, OneAgent packaging, multi-tenant SSO, ServiceNow OAuth, full OpenAPI)  
**Audience:** Product, engineering, sales, customer success  
**Related docs:** [FRONTEND_STACK.md](./FRONTEND_STACK.md), [PRODUCTION_AND_GTM.md](./PRODUCTION_AND_GTM.md), [README.md](../README.md)

---

## Table of contents

1. [Purpose](#1-purpose)
2. [Executive summary](#2-executive-summary)
3. [NeuralOps UI inventory](#3-neuralops-ui-inventory)
4. [Feature parity matrix](#4-feature-parity-matrix)
5. [Gap analysis by Dynatrace product area](#5-gap-analysis-by-dynatrace-product-area)
6. [Where NeuralOps is differentiated](#6-where-neuralops-is-differentiated)
7. [Recommended UI roadmap (priority order)](#7-recommended-ui-roadmap-priority-order)
8. [Sales positioning notes](#8-sales-positioning-notes)
9. [Appendix — route reference](#9-appendix--route-reference)
10. [Full implementation plan](#10-full-implementation-plan)

---

## 1. Purpose

This document compares the **NeuralOps React UI** (`frontend/src`) to **Dynatrace** as a reference observability platform. It is intended for:

- Product planning and roadmap prioritization
- Honest sales and POC scoping (what to demo vs what to defer)
- Engineering estimates for UI parity work

**Scope:** UI and user-facing workflows only. Backend capabilities that exist without a UI surface are noted where relevant but are not counted as “shipped” for end users.

**Baseline:** NeuralOps UI routes and components as of 2026-05-31. Dynatrace capabilities are described at the product-module level (APM, Smartscape, Davis, RUM, etc.), not tied to a specific Dynatrace release.

---

## 2. Executive summary

NeuralOps UI today is an **AI-native observability command center with broad Dynatrace-style surface area at MVP+ depth** — not yet a full OneAgent-grade replacement.

| Dimension | NeuralOps | Dynatrace |
|-----------|-----------|-----------|
| **Core strength** | Log search + AI explanations, incident intelligence, streaming AI assistant | Full-stack observability with entity model (PurePath, Smartscape, Davis) |
| **Overlap** | Logs, incidents, APM (waterfall/compare), metrics, topology, alerting, RUM/synthetic shells, admin | Same areas with deeper entity graph, agents, and mature workflows |
| **Largest remaining gaps** | Fleet-wide agent packaging in customer clusters; App Store published builds; multi-cloud credential federation at scale | — |
| **Differentiation** | Generative AI on logs/incidents, plain-English RCA, banking/txn journey demo | Entity-centric Davis AI, OneAgent auto-instrumentation, mature ecosystem |

**Bottom line for buyers:** NeuralOps competes on **incident speed and log intelligence** with credible **breadth for POCs**; Dynatrace wins on **production depth, agent coverage, and enterprise ecosystem maturity**.

---

## 3. NeuralOps UI inventory

Routes are defined in `frontend/src/router.tsx`. Primary navigation is in `frontend/src/components/layout/Sidebar.tsx`.

| Route | Screen | Status | Dynatrace analogue |
|-------|--------|--------|-------------------|
| `/` | Dashboard (Command Center) | **Implemented** | Home / dashboards |
| `/incidents` | Incidents list | **Implemented** | Problems |
| `/incidents/$id` | Incident detail (tabs) | **Implemented** | Problem detail + Davis RCA |
| `/logs` | Log Explorer | **Implemented** | Log monitoring |
| `/logs/settings` | Log rules & retention | **Partial** | Log processing rules |
| `/traces` | Trace Explorer | **Partial** | PurePath / distributed traces |
| `/traces/$traceId` | Trace detail (waterfall + flame) | **Partial** | PurePath detail |
| `/traces/compare` | Trace compare | **Partial** | Trace regression analysis |
| `/service-flow` | Service flow | **Partial** | Service flow |
| `/entities/$type/$id` | Entity shell | **Partial** | Smartscape entity |
| `/transactions` | Transaction Journey | **Implemented** (domain demo) | Business flow / PurePath lite |
| `/metrics` | Metrics explorer | **Partial** | Data explorer |
| `/dashboards`, `/dashboards/$id` | Custom dashboards + builder | **Partial** | Dashboards |
| `/anomalies` | Anomaly Detection | **Partial** | Davis anomalies |
| `/slos` | SLO management | **Partial** | SLO / error budgets |
| `/service-map` | Service Map | **Partial** | Smartscape / service flow |
| `/infrastructure` | Host / infra | **Partial** | Infrastructure |
| `/kubernetes` | Kubernetes | **Partial** | Kubernetes app |
| `/databases` | Database monitoring | **Partial** | Database insights |
| `/middleware` | Middleware (Kafka, etc.) | **Partial** | Technology-specific views |
| `/rum` | Real User Monitoring | **Partial** | Browser RUM |
| `/rum/sessions/$sessionId/replay` | Session replay | **Partial** | Session replay |
| `/synthetic` | Synthetic monitors | **Partial** | Synthetic |
| `/ai-chat` | AI Assistant | **Implemented** | Davis Copilot / Notebooks |
| `/workflows` | Workflows | **Partial** | Automation |
| `/notebooks` | Notebooks | **Partial** | Notebooks |
| `/security`, `/security/attacks/$id` | Application security | **Partial** | Application security |
| `/integrations` | Integration wizards | **Partial** | Integrations |
| `/marketplace` | Extension catalog | **Partial** | Hub / marketplace |
| `/alerts` | Alerts (rules, channels, silences) | **Partial** | Alerting configuration |
| `/settings/*` | Admin (users, keys, audit, usage) | **Partial** | Admin / platform settings |
| `/health` | System Health | **Implemented** | Internal ops (not end-user Dynatrace feature) |
| `/design-system` | Design System | **Dev-only** | N/A |
| `/login`, `/auth/callback` | Auth | **Implemented** | SSO login flows |

### Incident detail tabs (`/incidents/$id`)

Overview, Timeline, Logs, Traces, Metrics, AI Analysis, Recommendations — **implemented** with acknowledge/resolve, MTTR clock, WebSocket-driven updates.

### Cross-cutting UI

| Capability | Status |
|------------|--------|
| Environment + time range filters (top bar) | Implemented |
| Command palette (⌘K navigation) | Implemented |
| WebSocket realtime (incidents, anomalies) | Implemented |
| OIDC login / logout | Implemented |
| Live log tail | Implemented |
| Log export (JSON/CSV) + share links | Implemented |
| Grafana dashboards | External (not in-app) |

---

## 4. Feature parity matrix

Legend: **Full** = comparable UX depth · **Partial** = exists but materially weaker · **None** = no UI · **Stub** = placeholder only

| Category | NeuralOps UI | Dynatrace UI |
|----------|--------------|--------------|
| Log search & tail | **Full** | Full |
| AI on logs / incidents | **Full** (generative) | Full (Davis, entity-based) |
| Incident / problem management | **Full** | Full |
| Distributed tracing (waterfall) | **Partial** (waterfall, flame, compare, span→logs) | Full (PurePath) |
| Trace search (ID lookup) | **Partial** | Full |
| Service topology (live) | **Partial** (topology API, zones filter) | Full (Smartscape) |
| Metrics explorer | **Partial** (PromQL proxy + charts) | Full |
| Anomaly detection UI | **Partial** | Full |
| Custom dashboards | **Partial** (create + drag-reorder builder) | Full |
| Alert rule builder | **Partial** (CRUD + edit rules/channels) | Full |
| Notification channels UI | **Partial** (Slack/PagerDuty/webhook) | Full |
| Infrastructure / hosts | **Partial** (collector-backed lists) | Full |
| Kubernetes monitoring | **Partial** (Prometheus/collector sync) | Full |
| Cloud (AWS/Azure/GCP) | **Partial** (federation API + demo fallback) | Full |
| RUM (browser/mobile) | **Partial** (SDK + sessions + replay + native mobile SDK) | Full |
| Synthetic monitoring | **Partial** (HTTP checks via collector) | Full |
| Database monitoring | **Partial** (statement views) | Full |
| SLO / SLI management | **Partial** (create SLIs + burn alerts) | Full |
| Admin (users, SSO, tokens) | **Partial** (multi-tenant SSO reload, users, API keys) | Full |
| Workflow / remediation automation | **Partial** (editor + Jira/Slack/PagerDuty execution) | Full |
| Application security | **Partial** (attack list + detail drill) | Full |
| Session replay | **Partial** (rrweb + consent banner) | Full |
| Mobile on-call app | **Partial** (Expo + store-ready SDK + server push) | Full |
| Extension marketplace | **Partial** (catalog UI) | Full |
| Continuous profiling | **Partial** (Pyroscope + eBPF Beyla in compose) | Full |
| OneAgent / auto-instrumentation | **Partial** (Node/Go/Java/.NET SDK + deb/rpm/k8s) | Full |

---

## 5. Gap analysis by Dynatrace product area

### 5.1 Application Performance Monitoring (APM / PurePath)

**NeuralOps has:** Trace explorer, trace detail with waterfall + flame graph, trace compare, service flow, span→logs drill with time window, ClickHouse span store, entity shell at `/entities/$type/$id`.

**Still missing vs Dynatrace:**

- Waterfall / flame graph trace viewer (span tree, timing bars, critical path within trace) — **shipped MVP+**
- Service flow view (request path with per-hop latency and error rates) — **partial**
- Compare traces (slow vs fast, regression analysis) — **shipped MVP+**
- Automatic request naming and endpoint grouping UI
- Code-level drill-down (method hotspots, continuous profiling)
- Unified one-click drill: trace → logs → metrics → host (partially present only in incident context)
- Trace retention and sampling policy UI

**Impact:** Highest gap for teams evaluating NeuralOps as an **APM replacement**. Log-centric workflows are strong; **request-level performance analysis is not**.

---

### 5.2 Smartscape & service topology

**NeuralOps has:** Interactive React Flow service map with topology API fallback, management zone filter, entity drill links, critical path highlight, PNG export, service detail panel.

**Still missing vs Dynatrace:**

- Auto-discovered topology from live telemetry (OneAgent / OTEL entity graph)
- Real-time health propagation on edges and nodes
- Drill-down from map to host, process, container, or database entity
- Vertical topology (data center, host, process, service layers)
- Dependency change over time (before/after deploy)
- Management zones and team-scoped topology views

**Impact:** Service map reads as **demo/visualization** today, not operational Smartscape.

---

### 5.3 Davis AI vs NeuralOps AI

| Capability | Dynatrace (Davis) | NeuralOps |
|------------|-------------------|-----------|
| Root cause on entities | Host, service, process, DB | Logs, incidents, chat context |
| Automatic problem open/close | Yes | Semi-automatic incidents |
| Anomaly baselines per metric | Per-entity in UI | Dashboard-level aggregates |
| Causal analysis graph | Entity-centric | Narrative + timeline in incident |
| Copilot / natural language | Yes | **AI Assistant** (`/ai-chat`) — streaming, sources |
| Automation / workflows | Yes | **Partial** (`/workflows` create + list) |
| Notebooks | Yes | **Partial** (`/notebooks` create + list) |

**Impact:** NeuralOps wins on **plain-English log/incident intelligence**; Dynatrace wins on **cross-stack entity causality** without requiring log search.

---

### 5.4 Log monitoring

**NeuralOps has:** Text, regex, and AI search; filters (service, severity, host, pod, environment); live tail; virtualized list; detail panel; stack trace detection; AI explanations; export and share.

**Missing vs Dynatrace:**

- Log metrics (count-based alerts from log patterns) in UI
- Log processing / parsing rule editor
- Log buckets / retention policy UI
- Deep integration: click log line → open exact span in PurePath waterfall — **partial** (log→trace link exists; span→logs with window shipped)

**Impact:** **Low gap** — core log UX is competitive; integration with traces is the main weakness.

---

### 5.5 Infrastructure, Kubernetes, and cloud

**NeuralOps has (MVP+):** `/infrastructure`, `/kubernetes` screens backed by collector + Prometheus sync; host/pod as log filter fields and entity types.

**Still missing vs Dynatrace:**

- Host list and detail (CPU, memory, disk, network)
- Process monitoring
- Container views
- Kubernetes: clusters, nodes, namespaces, workloads, pods, events
- Cloud vendor dashboards (AWS, Azure, GCP)
- Network monitoring

**Note:** Host/pod appear as **log filter fields**, not as monitored infrastructure entities.

**Impact:** **Blocker** for platform/SRE buyers who expect Dynatrace “Infrastructure” or “Kubernetes” apps.

---

### 5.6 Real User Monitoring (RUM) and Synthetic

**NeuralOps has (MVP+):** `/rum` sessions, `sdk/rum/neuralops-rum.js` beacon, `/rum/sessions/$sessionId/replay`, `/synthetic` with collector HTTP checks.

**Still missing vs Dynatrace:**

- Browser RUM (page load, JS errors, Core Web Vitals, geo, devices)
- Mobile RUM (iOS/Android)
- User session analysis and funnel views
- Session replay
- Synthetic monitors (HTTP/browser tests, global locations, schedules, run history)

**Impact:** **Blocker** for frontend/product teams and SLA monitoring from the user’s perspective.

---

### 5.7 Database and middleware monitoring

**NeuralOps has (MVP+):** `/databases`, `/middleware` with technology-oriented summary views.

**Still missing vs Dynatrace:**

- Database statement analysis, wait events, connection pool views
- Technology-specific dashboards (Postgres, Redis, Kafka as first-class entities)
- Message queue depth and consumer lag UI (Kafka lag may exist in Grafana/backend, not as a dedicated NeuralOps screen)

---

### 5.8 Alerting and notifications

**NeuralOps has:** `/alerts` with Active, Rules, Channels, Silences, History tabs; rule/channel CRUD + edit; alert detail modal with incident link; backend alerting service.

**Still missing vs Dynatrace:**

- Alert rule builder (thresholds, log events, metric events)
- Notification channel configuration (Slack, PagerDuty, email, webhooks)
- Alert silencing and maintenance windows
- Alert history, deduplication, and escalation policies
- SLO-based alerting UI

**Note:** Alerting **backend service** exists; **management UI is MVP+** — escalation policies and SLO burn alerts remain.

**Impact:** **Medium** — ops can self-serve basic alert config; enterprise escalation/SLO alerting still gaps.

---

### 5.9 Dashboards, metrics, and analytics

**NeuralOps has:** Command Center dashboard; `/metrics` explorer; `/dashboards` with create + drag-reorder builder; `/slos` with create form; Grafana in `infra/grafana/` (external).

**Still missing vs Dynatrace:**

- Custom dashboard builder (drag-drop tiles, sharing, favorites)
- Metric browser / data explorer (DQL-style querying in UI)
- Notebooks for ad-hoc analysis
- Scheduled reports / PDF export
- SLO / SLI definition and error budget UI
- Business KPI / BizEvents dashboards (beyond Transaction Journey)

---

### 5.10 Admin, multi-tenancy, and governance

**NeuralOps has:** `/settings` hub with users, API keys (create + one-time secret), audit log, usage; OIDC login; tenant name in top bar.

**Still missing vs Dynatrace:**

- User, team, and role management UI
- SSO / IdP configuration UI
- API token lifecycle management
- Management zones / environment grouping
- Data retention and ingestion policy UI per tenant
- Audit log viewer
- License and usage metering UI

**Note:** RBAC and auth patterns exist on the **backend**; governance is largely **API/ops-config**, not self-serve UI.

---

### 5.11 Security and application protection

**NeuralOps has:** `/security` attack list, `/security/attacks/$id` detail with trace/log/entity drill.

**Still missing vs Dynatrace:**

- Runtime application security (vulnerabilities, attack detection)
- Security problem types and dashboards
- Attack analytics and blocking workflows

---

### 5.12 Mobile app and ecosystem

**NeuralOps has:** `mobile/` Expo app (Dashboard, Alerts, Incidents); `/marketplace` catalog; `/integrations` connect wizard.

**Still missing vs Dynatrace:**

- Native mobile client for on-call
- In-product extension marketplace
- Integration wizards (Jira, ServiceNow, CI/CD) in UI

---

## 6. Where NeuralOps is differentiated

Use these in positioning; do not claim Dynatrace parity here.

1. **AI-native log search** — natural language and regex modes with plain-English explanations on log lines.
2. **Incident intelligence** — AI Analysis and Recommendations tabs, typing narrative, actionable recs.
3. **Streaming AI Assistant** — copilot-style Q&A with cited sources (`/ai-chat`).
4. **Transaction Journey** — business-oriented hop view (e.g. UPI/payments demo narrative).
5. **Unified dark command-center UX** — purpose-built for incident response, not generic dashboard sprawl.

---

## 7. Recommended UI roadmap (priority order)

Prioritized by **customer impact** and **Dynatrace displacement** value.

> **2026-05-31 update:** All P0–P3 items in the table below are **shipped**. GTM: `mobile/store/SUBMISSION.md`, `deploy/kubernetes/fleet-profiling.yaml`.

| Priority | Initiative | Rationale | Suggested scope |
|----------|------------|-----------|-----------------|
| P0 | **Trace waterfall + service flow** | Largest APM gap; unblocks “replace Dynatrace for tracing” conversations | Span tree, timing bars, trace ↔ log links |
| P0 | **Alert rules + channels UI** | Alerts page is a stub; ops cannot self-serve | CRUD rules, Slack/PagerDuty/webhook, silence windows |
| P1 | **Metrics explorer** | No standalone metrics UI today | Service/metric picker, time series, compare |
| P1 | **Live service map** | Map must reflect real telemetry | OTEL/service graph API, health on nodes/edges |
| P1 | **Settings / admin** | Enterprise buyers expect self-serve | Users, roles, API keys, SSO read-only status |
| P2 | **Custom dashboards** | Or deep Grafana embed with SSO | Tile builder or iframe + variable sync |
| P2 | **SLO management** | Enterprise SLAs | SLI definitions, error budgets, burn alerts |
| P3 | **Kubernetes / host views** | Platform team requirement | Node/pod list, resource charts, log jump |
| P3 | **RUM / synthetic** | Only if GTM targets frontend teams | Separate agent + new UI modules |

---

## 8. Sales positioning notes

### Safe to demo today

- Log Explorer (search, tail, AI, export, trace drill)
- Incidents list and detail (timeline, AI analysis, recommendations)
- AI Assistant
- Trace waterfall, flame graph, compare, span→logs drill
- Metrics explorer, custom dashboards (builder), SLOs
- Alerts (rules, channels, silences, history)
- Service Map with topology API and zone filter
- Entity pages, infrastructure/K8s/RUM/synthetic shells
- Anomaly Detection (overview-driven)
- Transaction Journey (when txn IDs exist in demo data)
- Security attacks, integrations, marketplace catalog
- Settings (users, API keys, audit)

### Do not promise without roadmap commitment

- OneAgent-grade auto-instrumentation and code-level profiling
- Live Smartscape from production agents at Dynatrace depth
- rrweb-grade session replay + GDPR tooling
- Executable workflows and notebook cells
- Cloud vendor dashboards (AWS/Azure/GCP)
- Mobile push notifications and app-store on-call app
- SLO burn-rate alerting and full escalation policies

### Suggested talk track

> “NeuralOps accelerates **mean time to understand** through AI on logs and incidents, with **credible breadth** across APM, metrics, alerting, and topology for POCs. Dynatrace remains the reference for **production-grade agent coverage and ecosystem depth**. We win when the buyer’s pain is log noise, slow RCA, and incident war rooms — not when they need full cloud-native infra parity on day one.”

---

## 9. Appendix — route reference

| Path | Component | File |
|------|-----------|------|
| `/` | Dashboard | `frontend/src/pages/Dashboard.tsx` |
| `/incidents` | Incidents | `frontend/src/pages/Incidents.tsx` |
| `/incidents/$id` | Incident detail | `frontend/src/pages/IncidentDetail.tsx` |
| `/logs` | Log Explorer | `frontend/src/pages/LogExplorer.tsx` |
| `/logs/settings` | Log settings | `frontend/src/pages/LogSettings.tsx` |
| `/traces` | Trace Explorer | `frontend/src/pages/TraceExplorer.tsx` |
| `/traces/$traceId` | Trace detail | `frontend/src/pages/TraceDetail.tsx` |
| `/traces/compare` | Trace compare | `frontend/src/pages/TraceCompare.tsx` |
| `/service-flow` | Service flow | `frontend/src/pages/ServiceFlow.tsx` |
| `/entities/$type/$id` | Entity page | `frontend/src/pages/EntityPage.tsx` |
| `/metrics` | Metrics explorer | `frontend/src/pages/MetricsExplorer.tsx` |
| `/dashboards` | Dashboard list | `frontend/src/pages/Dashboards.tsx` |
| `/dashboards/$id` | Dashboard view/builder | `frontend/src/pages/DashboardView.tsx` |
| `/slos` | SLOs | `frontend/src/pages/SLOs.tsx` |
| `/transactions` | Transaction Journey | `frontend/src/pages/TransactionJourney.tsx` |
| `/anomalies` | Anomaly Detection | `frontend/src/pages/AnomalyDetection.tsx` |
| `/service-map` | Service Map | `frontend/src/pages/ServiceMap.tsx` |
| `/infrastructure` | Infrastructure | `frontend/src/pages/Infrastructure.tsx` |
| `/kubernetes` | Kubernetes | `frontend/src/pages/Kubernetes.tsx` |
| `/databases` | Databases | `frontend/src/pages/Databases.tsx` |
| `/middleware` | Middleware | `frontend/src/pages/Middleware.tsx` |
| `/rum` | RUM | `frontend/src/pages/RUM.tsx` |
| `/rum/sessions/$sessionId/replay` | Session replay | `frontend/src/pages/SessionReplay.tsx` |
| `/synthetic` | Synthetic | `frontend/src/pages/Synthetic.tsx` |
| `/ai-chat` | AI Assistant | `frontend/src/pages/AIChat.tsx` |
| `/workflows` | Workflows | `frontend/src/pages/Workflows.tsx` |
| `/notebooks` | Notebooks | `frontend/src/pages/Notebooks.tsx` |
| `/security` | Security | `frontend/src/pages/Security.tsx` |
| `/security/attacks/$id` | Attack detail | `frontend/src/pages/SecurityAttackDetail.tsx` |
| `/integrations` | Integrations | `frontend/src/pages/Integrations.tsx` |
| `/marketplace` | Marketplace | `frontend/src/pages/Marketplace.tsx` |
| `/alerts` | Alerts | `frontend/src/pages/Alerts.tsx` |
| `/settings` | Settings hub | `frontend/src/pages/Settings.tsx` |
| `/settings/users` | Users | `frontend/src/pages/SettingsUsers.tsx` |
| `/settings/api-keys` | API keys | `frontend/src/pages/SettingsApiKeys.tsx` |
| `/settings/audit` | Audit log | `frontend/src/pages/SettingsAudit.tsx` |
| `/settings/usage` | Usage | `frontend/src/pages/SettingsUsage.tsx` |

---

## 10. Full implementation plan

For the **phase-by-phase engineering plan** (12 phases, backend + frontend tasks, timelines, exit criteria, and gap traceability), see:

**[UI_DYNATRACE_IMPLEMENTATION_PLAN.md](./UI_DYNATRACE_IMPLEMENTATION_PLAN.md)**

---

**Maintainers:** Update this document when adding routes or shipping parity features. Bump **Last updated** and **Version** on substantive changes.
