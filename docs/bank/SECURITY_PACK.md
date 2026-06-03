# NeuralOps — Enterprise Security Pack (BFSI)

**Version:** 1.0  
**Date:** 2026-06-03  
**Audience:** Bank security, GRC, architecture review  
**Status:** Draft for vendor risk assessment — not a SOC 2 attestation  

Export this document to PDF for customer distribution. Companion: [DATA_FLOW.md](./DATA_FLOW.md).

---

## 1. Product summary

NeuralOps is an **AI-native observability and incident intelligence platform** for logs, metrics, traces, incidents, and alerting. For banks, typical use cases include payment/UPI transaction correlation, service-level triage, and SSO-governed operator access.

**Deployment models for regulated buyers:**

| Model | Data location | Recommended |
|-------|---------------|-------------|
| Self-hosted Enterprise | Customer Kubernetes / data center | **Primary for banks** |
| Dedicated Cloud | Isolated VPC operated by vendor | Optional |
| Multi-tenant SaaS | Shared vendor infrastructure | Not recommended for first bank deal |

---

## 2. Architecture overview

```text
                    ┌─────────────────────────────────────┐
                    │  WAF / Ingress (TLS 1.2+)         │
                    └─────────────────┬───────────────────┘
                                      │
                    ┌─────────────────▼───────────────────┐
                    │  API Gateway                       │
                    │  OAuth2/OIDC/SAML, RBAC, ABAC,     │
                    │  rate limits, audit, tenant scope  │
                    └─────────────────┬───────────────────┘
          ┌───────────────────────────┼───────────────────────────┐
          │                           │                           │
   ┌──────▼──────┐            ┌───────▼───────┐           ┌───────▼───────┐
   │ Ingestion   │            │ Search        │           │ Incident      │
   │ Analysis    │            │ Alerting      │           │ Observability │
   └──────┬──────┘            └───────┬───────┘           └───────┬───────┘
          │                           │                           │
          └───────────────────────────┼───────────────────────────┘
                                      │
          ┌───────────────────────────▼───────────────────────────┐
          │  Kafka │ PostgreSQL │ Elasticsearch │ ClickHouse     │
          │  Redis │ Qdrant (optional semantic search)             │
          └───────────────────────────────────────────────────────┘
```

Canonical reference: [ARCHITECTURE.md](../../ARCHITECTURE.md).

---

## 3. Authentication and authorization

| Control | Implementation |
|---------|----------------|
| Human users | OIDC (e.g. Okta, Azure AD) or SAML 2.0 — **required in production** |
| Machine ingest | API keys with tenant binding; rotation supported |
| RBAC | `ADMIN`, `SRE`, `DEVELOPER`, `READONLY`, `ALERT_MANAGER` |
| ABAC | Service-scope filters on observability/FinOps queries (`X-Allowed-Services`) |
| Sessions | JWT access + refresh; configurable TTL |
| Dev/demo login | **Disabled** in production (`AUTH_ALLOW_DEV_LOGIN=false`) |

Production enforcement: gateway refuses unsafe config when `ENVIRONMENT=production` (see [PRODUCTION_AUTH_CHECKLIST.md](./PRODUCTION_AUTH_CHECKLIST.md)).

---

## 4. Encryption

| Layer | Standard |
|-------|----------|
| In transit | TLS 1.2+ on all external endpoints; optional mTLS east-west |
| At rest | Customer-managed KMS for RDS/EBS/S3 (self-hosted or dedicated cloud) |
| Secrets | Vault / AWS Secrets Manager / GCP Secret Manager / K8s External Secrets — never in git |

---

## 5. Tenant isolation

- Every API request carries `tenantId` (JWT claim or `X-Tenant-ID` for service accounts).
- Data stores scoped by tenant: ES index prefix, ClickHouse `tenant_id`, Postgres FK.
- Automated tests: FinOps scope ABAC, cross-tenant cost isolation (`finops_handlers_test.go`).
- **Recommendation:** Customer pen test includes cross-tenant read attempts.

---

## 6. Audit and logging

| Event type | Store | Retention (default) |
|------------|-------|---------------------|
| Admin actions (SSO, policy) | PostgreSQL + ClickHouse replica | Contract-defined |
| API audit middleware | Structured JSON logs | Centralized SIEM export |
| FinOps mutations | FinOps audit API | Exportable |

**Not logged:** passwords, API secrets, raw LLM prompts containing customer PII (configurable scrubbing).

---

## 7. Data handling and PII

- Customer controls what is sent in logs/traces (card data, PAN, credentials must be redacted at source).
- Collector-level scrubbing patterns supported — see [PII_SCRUBBING_GUIDE.md](./PII_SCRUBBING_GUIDE.md).
- LLM: customer-supplied API keys; no training on customer data (contractual) — see [CUSTOMER_LLM_KEYS.md](./CUSTOMER_LLM_KEYS.md).

---

## 8. Availability and operations (initial targets)

| Service | Target SLO (staging/prod agreement) |
|---------|-----------------------------------|
| Gateway API | 99.9% monthly availability |
| Log ingestion | 99.99% accepted (excl. client errors) |
| Search | p95 &lt; 2s for agreed retention window |

Runbooks: [docs/runbooks/](../runbooks/). DR: [BACKUP_RESTORE.md](../runbooks/BACKUP_RESTORE.md).

---

## 9. Compliance posture (honest)

| Attestation | Status |
|-------------|--------|
| SOC 2 Type II | **Roadmap** — not claimed until audit complete |
| ISO 27001 | Roadmap |
| PCI DSS (for NeuralOps as vendor) | N/A unless vendor processes cardholder data |
| Penetration test | Required before production promotion |

---

## 10. Subprocessors

See [SUBPROCESSORS.md](./SUBPROCESSORS.md). Customer LLM provider is typically **customer-controlled**.

---

## 11. Contact and evidence artifacts

| Artifact | Location |
|----------|----------|
| mTLS verification | `scripts/verify-mtls.sh` |
| OIDC verification | `scripts/verify-keycloak-oidc.sh` |
| Production auth check | `scripts/verify-production-auth.sh` |
| CI security gates | `.github/workflows/ci.yml`, [AUDIT_COMPLIANCE.md](../AUDIT_COMPLIANCE.md) |

**Legal:** [DPA_TEMPLATE.md](./DPA_TEMPLATE.md), [MSA_ELA_TEMPLATE.md](./MSA_ELA_TEMPLATE.md) — require legal review before execution.

---

*Internal owner: Security + Platform Engineering. Review before each bank RFP.*
