# NeuralOps — Production Ship & Enterprise Sales Guidelines

**Version:** 1.0  
**Last updated:** 2026-05-31  
**Audience:** Engineering, platform/SRE, product, sales, customer success, leadership  
**Related docs:** [AUDIT_COMPLIANCE.md](./AUDIT_COMPLIANCE.md), [README.md](../README.md), [infra/helm/neuralops/values.yaml](../infra/helm/neuralops/values.yaml)

---

## Table of contents

1. [Executive summary](#1-executive-summary)
2. [Current platform state](#2-current-platform-state)
3. [Production deployment models](#3-production-deployment-models)
4. [Phase A — Production baseline](#4-phase-a--production-baseline)
5. [Phase B — Multi-tenant SaaS](#5-phase-b--multi-tenant-saas)
6. [Phase C — Scale and cost optimization](#6-phase-c--scale-and-cost-optimization)
7. [Security and compliance guidelines](#7-security-and-compliance-guidelines)
8. [Operability and SRE guidelines](#8-operability-and-sre-guidelines)
9. [Product packaging and SKUs](#9-product-packaging-and-skus)
10. [Go-to-market strategy](#10-go-to-market-strategy)
11. [Sales and POC playbook](#11-sales-and-poc-playbook)
12. [Pricing and packaging](#12-pricing-and-packaging)
13. [Customer onboarding guidelines](#13-customer-onboarding-guidelines)
14. [Team, budget, and timeline](#14-team-budget-and-timeline)
15. [Production readiness checklist](#15-production-readiness-checklist)
16. [Risks and mitigations](#16-risks-and-mitigations)
17. [Appendix — reference commands](#17-appendix--reference-commands)

---

## 1. Executive summary

NeuralOps is an **AI-native observability and incident intelligence platform**. The codebase supports a full demo stack (Docker Compose), CI quality gates, Helm/Kubernetes artifacts, and enterprise auth patterns (OIDC, SAML, RBAC, mTLS demo).

**Strategic split:**

| Environment | Purpose |
|-------------|---------|
| **Docker Compose** | Local dev, sales demos, CI smoke tests |
| **Kubernetes + managed data plane** | Production and paying customers |

**Two parallel tracks:**

1. **Ship to production** — harden security, operability, multi-tenancy, billing, and compliance.
2. **Sell to other companies** — package SKUs, run design-partner pilots, convert to paid Cloud SaaS, Dedicated, or Self-hosted Enterprise.

**Target first milestone:** One production-grade staging environment + 2 design-partner pilots + 1 paid customer within ~90 days of starting Phase A.

---

## 2. Current platform state

### What is production-ready today

| Capability | Evidence |
|------------|----------|
| Microservices architecture | Gateway, ingestion, analysis, correlation, incident, search, alerting |
| Event pipeline | Kafka → analysis → Elasticsearch / ClickHouse |
| Auth | OIDC (Keycloak), SAML (crewjam), RBAC, API keys, dev login (disable in prod) |
| Frontend | React UI, TanStack Router, Tailwind v4 |
| Deploy artifacts | Helm chart (`infra/helm/neuralops/`), Terraform hook, HPA/PDB |
| Observability | Prometheus, Grafana dashboards, Jaeger/OTEL |
| CI | Unit, integration, compose-smoke, Playwright E2E |
| Audit traceability | [AUDIT_COMPLIANCE.md](./AUDIT_COMPLIANCE.md) |

### Known gaps before enterprise sales

- Aggregate test coverage below 80% repo-wide (tiered gates only on core packages).
- Internal services use plain HTTP on Docker network; mTLS demonstrated via nginx sidecar only.
- No tenant control plane, billing, or usage metering UI.
- No formal SOC 2, pen test, or DR drill documentation.
- Compose healthchecks and dev cert scripts are tuned for local/demo — not prod SLOs.
- LLM cost controls and per-tenant quotas not enforced in product.

**Rule:** Never sell “Docker Compose on a single VM” as production. Sell **managed SaaS** or **Helm on customer K8s**.

---

## 3. Production deployment models

### Recommended: Managed data plane + K8s compute

Run NeuralOps services on **Kubernetes** (EKS, GKE, or AKS). Use **managed** infrastructure for stateful systems:

| Component | Dev (Compose) | Production (recommended) |
|-----------|---------------|---------------------------|
| Postgres | Container | RDS / Cloud SQL / Azure Database |
| Kafka | Confluent in Compose | MSK / Confluent Cloud |
| Elasticsearch | Container | Elastic Cloud / OpenSearch Service |
| ClickHouse | Container | ClickHouse Cloud or dedicated cluster |
| Redis | Container | ElastiCache / Memorystore |
| Qdrant | Container | Qdrant Cloud or in-cluster with PVC + backup |
| Keycloak | Container (demo) | Customer IdP (Okta, Azure AD) or dedicated Keycloak |
| Ingress | localhost | ALB / NGINX Ingress + cert-manager + WAF |
| Secrets | Files / env | Vault / AWS Secrets Manager / GCP Secret Manager |

### Deployment topology (reference)

```text
                    ┌─────────────────────────────────────┐
                    │  CDN / WAF / Ingress (TLS)          │
                    └─────────────────┬───────────────────┘
                                      │
                    ┌─────────────────▼───────────────────┐
                    │  Gateway (auth, rate limit, proxy)   │
                    └─────────────────┬───────────────────┘
          ┌───────────────────────────┼───────────────────────────┐
          │                           │                           │
   ┌──────▼──────┐            ┌───────▼───────┐           ┌───────▼───────┐
   │ Ingestion   │            │ Analysis      │           │ Search        │
   │ Correlation │            │ Incident      │           │ Alerting      │
   └──────┬──────┘            └───────┬───────┘           └───────┬───────┘
          │                           │                           │
          └───────────────────────────┼───────────────────────────┘
                                      │
          ┌───────────────────────────▼───────────────────────────┐
          │  Kafka │ Postgres │ ES │ ClickHouse │ Redis │ Qdrant   │
          └───────────────────────────────────────────────────────┘
```

### GitOps

- Store environment config in a **separate infra repo** or `infra/environments/{staging,prod}/`.
- Use **Argo CD** or **Flux** to sync Helm releases.
- Never commit secrets; use External Secrets Operator + cloud secret store.

---

## 4. Phase A — Production baseline

**Duration:** 4–6 weeks  
**Goal:** One staging environment you trust for pilot customers.

### 4.1 Infrastructure tasks

- [ ] Choose primary cloud (AWS recommended for MSK + RDS maturity).
- [ ] Complete Terraform modules: VPC, EKS/GKE, RDS, MSK, OpenSearch, Redis, S3 for backups.
- [ ] Create `infra/helm/neuralops/values-staging.yaml` and `values-prod.yaml`.
- [ ] Configure cert-manager + Let's Encrypt or ACM on Ingress.
- [ ] Set up External Secrets for all DSNs, API keys, OIDC client secrets.
- [ ] Pin container images by digest in prod (not `:latest`).

### 4.2 Application hardening

- [ ] Disable in production:
  - `AUTH_ALLOW_DEV_LOGIN`
  - `AUTH_DISABLED`
  - `DEMO_MODE`
  - `CLICKHOUSE_SKIP_USER_SETUP`
- [ ] Require SSO (OIDC or SAML) for all human users.
- [ ] Enable mTLS or service mesh (Istio/Linkerd) for east-west traffic.
- [ ] Set resource requests/limits on every Deployment (Helm values).
- [ ] Configure pod disruption budgets (already scaffolded in `infra/k8s/autoscaling.yaml`).
- [ ] Run DB migrations via init job or dedicated migration pipeline (Flyway/Goose already in backend).

### 4.3 Observability

- [ ] Export metrics to Prometheus; configure alerts (Kafka lag, error rate, LLM latency).
- [ ] Ship logs to centralized logging (not only container stdout).
- [ ] OTEL traces to Jaeger/Tempo with sampling (10% in prod).
- [ ] Define SLOs (see [Section 8](#8-operability-and-sre-guidelines)).

### 4.4 Quality gates before Phase A exit

- [ ] Staging runs **7 consecutive days** without manual intervention.
- [ ] Load test at expected pilot volume (`make loadtest` baseline + customer-specific profile).
- [ ] Rollback drill: deploy bad image → rollback in **&lt; 15 minutes**.
- [ ] Restore drill: Postgres PITR + ES snapshot restore documented and tested once.

**Phase A exit criteria:** Staging sign-off from engineering + security checklist (Section 15) at 80%+.

---

## 5. Phase B — Multi-tenant SaaS

**Duration:** 6–10 weeks (after Phase A)  
**Goal:** Safely host multiple companies on one platform.

### 5.1 Tenant isolation

Every request and stored record must carry `tenantId`:

| Layer | Isolation mechanism |
|-------|---------------------|
| API Gateway | JWT/API key → tenant claim; reject cross-tenant paths |
| Kafka | Tenant ID in message metadata; optional topic prefix per tier |
| Elasticsearch | Tenant-scoped indices (`tenant.LogsBootstrapIndex`) |
| ClickHouse | `tenant_id` column + row policies |
| Postgres | `tenant_id` FK on all tenant-owned rows |
| Redis | Key prefix per tenant |
| Qdrant | Payload filter on `tenantId` |
| LLM usage | Per-tenant token budget and rate limit |

### 5.2 Control plane (build or extend)

Minimum admin capabilities:

- Create / suspend / delete tenant
- Configure IdP (OIDC issuer, SAML metadata URL)
- Issue and revoke API keys
- Set quotas: logs/day, retention days, users, AI credits
- Export tenant data (GDPR)
- View usage dashboard

**Implementation options:**

1. New `admin` service + admin UI (recommended for SaaS scale).
2. Extend gateway with `/api/v1/admin/*` + internal-only Ingress (faster MVP).

### 5.3 Billing and metering

Integrate **Stripe** (or Chargebee):

| Meter | Source |
|-------|--------|
| Logs ingested (GB) | Ingestion metrics / ClickHouse |
| Log retention (GB-month) | ES + ClickHouse storage |
| LLM tokens | Analysis metrics (`llm_api_calls_total`) |
| Active users | Postgres `users` table |
| API requests | Gateway Prometheus |

Emit usage records daily; invoice monthly. Hard-cap or soft-cap overages per contract.

### 5.4 Phase B exit criteria

- [ ] Provision tenant #2 in **&lt; 1 hour** (automated).
- [ ] Tenant A cannot read tenant B data (pen test or automated isolation test).
- [ ] Usage visible in admin UI and matches Stripe test invoices.

---

## 6. Phase C — Scale and cost optimization

**Duration:** Ongoing after first paying customers.

- Autoscale ingestion and analysis on Kafka consumer lag (KEDA).
- LLM: cache explanations by fingerprint; batch embeddings; cheaper model for classify fallback.
- Elasticsearch ILM: hot → warm → delete per tenant tier.
- ClickHouse TTL on raw metrics and audit tables.
- Multi-region: only when **3+ enterprise customers** require data residency outside primary region.
- Cost dashboards: $/tenant/month (infra + LLM).

---

## 7. Security and compliance guidelines

### 7.1 Authentication and authorization

| Requirement | Production guideline |
|-------------|---------------------|
| Human users | SSO only (OIDC or SAML); no dev login |
| Machine users | API keys with rotation policy (90 days) |
| Roles | Admin, SRE, Developer, ReadOnly, AlertManager — document in customer admin guide |
| Session | JWT access + refresh; revoke on logout |
| IdP | Customer-owned Okta/Azure AD/Google; Keycloak only for demo |

### 7.2 Data protection

- **Encryption in transit:** TLS 1.2+ everywhere (Ingress, DB connections, Kafka if supported).
- **Encryption at rest:** Cloud-managed keys (KMS) for RDS, S3, EBS.
- **Secrets:** Never in git; rotate on compromise; least-privilege IAM.
- **PII / logs:** Customer responsible for log content; provide redaction hooks and document what NeuralOps stores.
- **Audit logs:** Insert-only Postgres + ClickHouse replication (already implemented); retain 1–7 years per contract.

### 7.3 Network

- Private subnets for data plane; no public IPs on databases.
- NetworkPolicies: only gateway → services → data stores required paths.
- WAF on public Ingress (OWASP rules, rate limiting).
- Optional: IP allowlist for Self-hosted Enterprise.

### 7.4 Compliance roadmap

| Milestone | Target | Notes |
|-----------|--------|-------|
| Security overview doc | Phase A | 2-page PDF for sales |
| Penetration test | Before first enterprise close | Third-party, remediate critical/high |
| SOC 2 Type I readiness | Phase B | Policies, access reviews, change management |
| SOC 2 Type II | 6–12 months post Type I | Requires audit period |
| GDPR DPA | Before EU customers | Standard DPA + subprocessors list |
| HIPAA BAA | Only if healthcare vertical | Dedicated tier + BAAs |

### 7.5 Subprocessors (maintain public list)

Typical list for Cloud SaaS:

- Cloud provider (AWS/GCP/Azure)
- LLM provider (OpenAI / Anthropic) — disclose data sent for explanation
- Email (if alerting uses email)
- Stripe (billing)
- Optional: Elastic Cloud, Confluent, etc.

---

## 8. Operability and SRE guidelines

### 8.1 Service level objectives (initial)

| Service | SLI | SLO (monthly) |
|---------|-----|---------------|
| Gateway API | Availability + p99 latency | 99.9%, p99 &lt; 500ms |
| Log ingestion | Accepted / rejected ratio | 99.99% accepted (excl. customer 4xx) |
| Search | p95 query latency | p95 &lt; 2s for 90 days of data |
| Incident pipeline | Time from log to incident | p95 &lt; 5 min (async) |
| AI explanation | p95 LLM latency | p95 &lt; 30s (with cache hit target 60%+) |

### 8.2 Alerting (on-call)

Page on:

- Gateway 5xx rate &gt; 1% for 5 min
- Kafka consumer lag &gt; threshold per group
- Postgres connection pool exhausted
- Elasticsearch cluster red
- Any pod CrashLoopBackOff in prod namespace

### 8.3 Runbooks (required documents)

Create `docs/runbooks/` entries for:

1. Ingestion stall (Kafka down, disk full)
2. Search degraded (ES yellow/red)
3. Analysis backlog (LLM rate limit, Qdrant down)
4. Auth/SSO failure (IdP cert expiry)
5. Deploy rollback procedure
6. Database restore procedure

### 8.4 Release process

1. PR → CI green (unit, integration, compose-smoke).
2. Deploy to staging automatically.
3. Soak 24h for major releases.
4. Prod deploy: business hours, on-call present, feature flags for risky changes.
5. Semantic versioning for API (`/api/v1/` stable; breaking changes → v2).

### 8.5 Backup and DR

| System | Backup | RPO | RTO (target) |
|--------|--------|-----|--------------|
| Postgres | Automated PITR | 5 min | 1 h |
| Elasticsearch | Daily snapshot to S3 | 24 h | 4 h |
| ClickHouse | Backup to S3 / native backup | 24 h | 4 h |
| Kafka | Replication factor ≥ 2 in prod | N/A | 1 h |
| Config/Helm | Git | 0 | 30 min |

Quarterly: run restore drill and document results.

---

## 9. Product packaging and SKUs

### 9.1 Three commercial SKUs

#### SKU 1 — NeuralOps Cloud (SaaS)

- **Buyer:** Mid-market, 50–500 engineers, fast POC
- **Deployment:** Multi-tenant on NeuralOps infrastructure
- **Includes:** SSO, RBAC, standard retention, email support
- **Contract:** Monthly or annual subscription + usage overages

#### SKU 2 — NeuralOps Dedicated Cloud

- **Buyer:** Fintech, healthcare, regulated SaaS
- **Deployment:** Single-tenant VPC/cluster; optional private link
- **Includes:** Custom IdP, SLA 99.9%, dedicated support channel, custom retention
- **Contract:** Annual minimum commit

#### SKU 3 — NeuralOps Enterprise (Self-hosted)

- **Buyer:** Banks, government, air-gapped environments
- **Deployment:** Customer Kubernetes; Helm chart + license key
- **Includes:** Annual license, 8×5 or 24×7 support option, PS onboarding days
- **Contract:** Enterprise license agreement (ELA)

### 9.2 Feature matrix (sales reference)

| Feature | Cloud | Dedicated | Self-hosted |
|---------|-------|-----------|-------------|
| SSO (OIDC/SAML) | ✓ | ✓ | ✓ |
| RBAC | ✓ | ✓ | ✓ |
| AI explanations | ✓ | ✓ | ✓ (customer LLM key) |
| Custom retention | Tiered | ✓ | ✓ |
| Data residency | Region select | ✓ | Customer DC |
| SLA 99.9% | Optional | ✓ | Optional |
| Source code access | ✗ | ✗ | ✗ |
| Helm / K8s manifests | N/A | N/A | ✓ |

### 9.3 What not to sell

- Single-node Docker Compose in customer production
- Shared demo tenant credentials
- “Unlimited” LLM without metering
- SOC 2 compliance before audit is complete (say “SOC 2 in progress” honestly)

---

## 10. Go-to-market strategy

### 10.1 Positioning

**Category:** AI-native observability and incident intelligence.

**Tagline:** *Turn noisy logs into actionable incidents with plain-English root cause — before on-call burns out.*

**Differentiators:**

1. AI explanations tied to incident workflow (not generic chat on logs).
2. Correlation: logs + metrics + deployments + traces in one narrative.
3. Developer-scoped access (service-level RBAC).
4. Deploy flexibly: SaaS or self-hosted for regulated buyers.

### 10.2 Ideal customer profile (ICP)

| Attribute | Target |
|-----------|--------|
| Company size | 50–500 engineers |
| Verticals | Fintech, payments, B2B SaaS, platform teams |
| Pain | High MTTR, alert fatigue, junior staff can't triage JVM/Go traces |
| Stack | Already shipping logs (Fluent Bit, Filebeat, OTEL) to HTTP/Kafka |
| Budget | $50k–$300k/year observability spend |

### 10.3 Buyer personas

| Persona | Cares about |
|---------|-------------|
| **VP Engineering / CTO** | MTTR, team productivity, cost vs Datadog |
| **Head of SRE / Platform** | Integration effort, Kafka/ES compatibility, SSO |
| **Security / GRC** | SSO, audit logs, data residency, DPA |
| **Procurement** | Annual contract, clear usage metrics, exit/data export |

### 10.4 90-day GTM timeline

| Week | Engineering | Sales / marketing |
|------|-------------|-------------------|
| 1–2 | Phase A kickoff; staging env | Finalize positioning; 2-page security PDF |
| 3–4 | Staging hardening | Recruit 3–5 design partners |
| 5–6 | Phase A exit | POC playbook; demo recording (15 min) |
| 7–8 | Phase B start (tenant admin MVP) | First POC integrations |
| 9–10 | Billing integration (Stripe test) | Security questionnaires for active deals |
| 11–12 | Prod cutover prep | Close 1 paid customer; publish case study |

### 10.5 Marketing assets checklist

- [ ] Website: product, pricing (or “contact sales”), security page
- [ ] Architecture diagram (PDF)
- [ ] 15-minute demo video (seed data → incident → AI explanation)
- [ ] Comparison sheet vs “logs only” and vs “legacy APM”
- [ ] One customer case study (after first pilot)
- [ ] Docs portal: install (self-hosted), API reference, SSO setup

---

## 11. Sales and POC playbook

### 11.1 Discovery questions

1. How many logs per day? Peak burst?
2. Current stack (Splunk, Datadog, ELK, Grafana Loki)?
3. SSO provider? SAML required?
4. MTTR today for P1 incidents?
5. Data residency requirements?
6. Who is on-call and how many services?

### 11.2 Standard demo flow (30 minutes)

1. **Dashboard** — error rates, active incidents (2 min)
2. **Search** — natural language / structured query (3 min)
3. **Incident** — open P1, timeline, correlated deployment (5 min)
4. **AI chat** — “Why did UPI fail?” plain-English answer (5 min)
5. **Alerting** — deduplicated alert, no duplicate pages (3 min)
6. **Admin** — SSO, RBAC roles, API key ingestion (5 min)
7. **Q&A + next steps** (7 min)

Use: `bash scripts/quickstart.sh` locally or dedicated demo SaaS tenant (never prod customer data in demo).

### 11.3 POC structure (4 weeks)

| Week | Customer action | NeuralOps action |
|------|-----------------|------------------|
| 1 | Provide 1–3 services’ logs (syslog/JSON/OTEL) | Connect ingestion; validate parsing |
| 2 | Connect SSO (OIDC test app) | Configure tenant + RBAC |
| 3 | Run parallel with existing tool | Tune alerts, AI prompts, dashboards |
| 4 | Measure MTTR on 2 real incidents | Success report + commercial proposal |

**POC success criteria (define upfront):**

- ≥ 20% reduction in duplicate alerts, or
- ≥ 30% faster triage on agreed sample incidents, or
- Successful SSO + ingestion of ≥ X GB/day

### 11.4 Security review pack (send before POC)

1. Architecture diagram + data flow
2. List of data stored and retention defaults
3. Encryption (transit + rest)
4. SSO configuration guide
5. Subprocessors list
6. Sample DPA
7. Audit log description

### 11.5 Objection handling

| Objection | Response |
|-----------|----------|
| “We already have Datadog.” | NeuralOps focuses on **incident intelligence and AI triage**, integrates alongside existing ingestion. |
| “LLM data privacy?” | Customer can use **their API key**; no training on customer data; DPA available. |
| “Build vs buy?” | TCO of building Kafka+ES+AI pipeline is 3–5 engineers × 12+ months. |
| “Too early stage.” | Offer design-partner pricing + dedicated support + influence roadmap. |

---

## 12. Pricing and packaging

**Note:** Adjust after 2–3 POCs. Ranges below are guidelines for B2B SaaS.

### 12.1 Cloud SaaS (annual contract recommended)

| Tier | Monthly (USD) | Included |
|------|---------------|----------|
| **Starter** | $2,000 – $5,000 | 100 GB logs/mo, 30-day retention, 5 users, shared cluster |
| **Growth** | $8,000 – $15,000 | 500 GB/mo, 90-day retention, SSO, 25 users, AI credit bundle |
| **Enterprise** | $25,000+ | Custom retention, SLA, dedicated support, PS days |

### 12.2 Usage overages

| Metric | Typical overage |
|--------|-----------------|
| Log ingestion | $0.50 – $1.50 / GB |
| Extra retention | $0.10 – $0.30 / GB-month |
| LLM / AI credits | $0.002 – $0.01 / 1K tokens (pass-through + margin) |
| Additional users | $50 – $150 / user / month |

### 12.3 Dedicated Cloud

- **Floor:** $40k – $80k / year minimum commit
- Includes: isolated VPC, 99.9% SLA, named CSM, quarterly business review

### 12.4 Self-hosted Enterprise

- **License:** $100k – $250k / year (depends on log volume tier and support level)
- **Support:** 8×5 standard; 24×7 premium (+30–50%)
- **Professional services:** $2k – $3k / day for onboarding (typically 5–10 days)

### 12.5 Design partner pricing

- **First 3–5 customers:** 50–100% discount year 1 in exchange for case study, logo, and product feedback.
- Cap LLM usage in contract to control cost during pilot.

---

## 13. Customer onboarding guidelines

### 13.1 Pre-onboarding checklist

- [ ] Signed order form or pilot agreement
- [ ] Technical contact + security contact named
- [ ] IdP metadata (SAML) or OIDC client details exchanged
- [ ] Log format samples received
- [ ] Egress IPs documented if allowlist required

### 13.2 Day 0 — Tenant provision

1. Create tenant record (UUID, name, plan tier).
2. Configure IdP; test SSO login with customer admin.
3. Issue ingestion API key; share `POST /api/v1/logs` endpoint.
4. Enable Grafana folder or shared dashboards.
5. Schedule kickoff call (60 min).

### 13.3 Week 1 — Integration

- Validate log parsing for top 3 services.
- Confirm logs appear in search within 2 minutes of ingestion.
- Configure first alert rules (error rate, latency).
- Train champions: search, incidents, AI chat.

### 13.4 Week 2–4 — Value proof

- Review first incidents created automatically.
- Tune false-positive rate on alerts.
- Document MTTR baseline vs post-NeuralOps.
- Handoff to customer success with success plan.

### 13.5 Offboarding / export

- Provide JSON/CSV export of incidents, alert history, audit logs.
- 30-day data deletion after contract end (unless legal hold).
- Document in DPA.

---

## 14. Team, budget, and timeline

### 14.1 Minimum team

| Role | FTE | Responsibility |
|------|-----|----------------|
| Platform / SRE | 0.5–1.0 | K8s, prod env, on-call |
| Backend engineer | 1.0 | Control plane, billing, hardening |
| Frontend engineer | 0.5 | Admin UI, SSO UX |
| Founder / PM | 0.5 | Roadmap, design partners |
| Sales / CS | 0.5–1.0 | POCs, demos, renewals |
| Security (consultant) | Project | Pen test, SOC 2 prep |

### 14.2 Budget estimates (USD, year 1)

| Item | Range |
|------|-------|
| Cloud infra (staging + prod) | $3k – $15k / month |
| LLM API (demos + pilots) | $500 – $5k / month |
| Penetration test | $15k – $40k one-time |
| SOC 2 Type I | $30k – $80k one-time |
| Legal (MSA, DPA templates) | $5k – $15k one-time |
| CRM + marketing site | $500 – $2k / month |

### 14.3 Master timeline (Gantt overview)

```text
Month 1–1.5   Phase A (prod baseline)     ||||||||||||||
Month 2–3     Design partner POCs         ||||||||||||||||
Month 2–4     Phase B (multi-tenant)      ||||||||||||||||
Month 3       First paid customer         ★
Month 4–6     SOC 2 readiness             ||||||||||||||||
Month 6+      Phase C (scale)             ongoing
```

---

## 15. Production readiness checklist

Use this before calling any environment **production** or accepting paying traffic.

### Infrastructure

- [ ] Kubernetes cluster in private subnets
- [ ] Managed Postgres, Kafka, ES, Redis provisioned
- [ ] Ingress TLS valid; HSTS enabled
- [ ] Secrets in secret manager (not ConfigMap)
- [ ] Backups configured and restore tested
- [ ] Monitoring + paging wired to on-call

### Application

- [ ] Dev auth disabled
- [ ] All services have health/readiness probes
- [ ] Resource limits set; HPA configured
- [ ] Migrations run automatically on deploy
- [ ] Feature flags for AI and experimental paths

### Security

- [ ] SSO enforced for UI
- [ ] API keys rotatable
- [ ] Audit logging enabled
- [ ] NetworkPolicies applied
- [ ] Dependency scan in CI (no critical CVEs unpatched)

### Legal / commercial

- [ ] MSA and DPA templates ready
- [ ] Subprocessors page published
- [ ] Support SLA defined
- [ ] Incident communication process (status page optional)

### Sales readiness

- [ ] Demo tenant separate from prod
- [ ] POC playbook documented
- [ ] Pricing approved internally
- [ ] Security pack PDF ready to send

---

## 16. Risks and mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Crowded observability market | Hard to differentiate | Lead with AI incident narrative + fast POC |
| LLM cost spike | Margin erosion | Per-tenant caps, caching, usage billing |
| Tenant data leak | Legal, trust loss | Isolation tests, pen test, bug bounty later |
| Key person dependency | Delivery stall | Document runbooks; hire SRE early |
| Customer expects Datadog parity | Lost deals | Integrate alongside existing tools; don’t rip-and-replace day 1 |
| Compose used in prod | Outages, no SLA | Contract forbids; only Helm/SaaS supported |

---

## 17. Appendix — reference commands

### Local demo (sales / dev)

```bash
bash scripts/gen-mtls-certs.sh
bash scripts/quickstart.sh
# UI: http://localhost:3000
# API: http://localhost:8080/health
```

### Quality gates

```bash
make test-coverage
make test-integration   # requires Docker
cd frontend && npm run test:e2e
```

### Helm deploy (staging/prod skeleton)

```bash
helm upgrade --install neuralops infra/helm/neuralops \
  -f infra/helm/neuralops/values.yaml \
  -f infra/helm/neuralops/values-prod.yaml \
  -n neuralops --create-namespace
```

### Auth verification (after deploy)

```bash
bash scripts/verify-keycloak-oidc.sh
bash scripts/verify-saml-metadata.sh
bash scripts/verify-mtls.sh
```

---

## Document maintenance

| Trigger | Action |
|---------|--------|
| New SKU or pricing change | Update Section 9 and 12 |
| SOC 2 milestone | Update Section 7.4 |
| New cloud region | Update Section 3 and subprocessors |
| Major architecture change | Update diagrams and Section 3 |

**Owner:** Product + Platform Engineering  
**Review cadence:** Monthly until first 5 customers; quarterly thereafter.

---

*This document is the canonical guideline for shipping NeuralOps to production and selling to enterprise customers. For technical audit traceability, see [AUDIT_COMPLIANCE.md](./AUDIT_COMPLIANCE.md).*
