# NeuralOps — Data Flow (Bank Vendor Review)

**Version:** 1.0  
**Date:** 2026-06-03  

---

## 1. Ingestion flow

```text
Bank systems (apps, K8s, agents)
    │  OTLP / HTTP JSON / syslog / Kafka forwarder
    ▼
NEXAGENT or OTEL Collector (optional PII scrub)
    │  TLS
    ▼
API Gateway  POST /api/v1/logs|metrics|traces|events
    │  Auth: API key or mTLS; tenantId injected
    ▼
Ingestion service → Kafka topics
    ▼
Analysis (optional LLM classify) / direct indexers
    ▼
Elasticsearch (logs) │ ClickHouse (analytics) │ Postgres (metadata)
```

**Data classification:** Customer operational telemetry — may contain PII if customer does not scrub. **Not** payment card storage by NeuralOps.

---

## 2. Human user access flow

```text
Bank employee browser
    │  HTTPS
    ▼
Ingress / UI (React SPA)
    │  OIDC or SAML via customer IdP
    ▼
API Gateway — JWT session, RBAC, ABAC scope
    ▼
Search / Incidents / Dashboards / FinOps (scoped)
```

No bank credentials stored by NeuralOps when SSO is used.

---

## 3. AI explanation flow (optional)

```text
Incident or search context (tenant-scoped)
    ▼
Gateway AI chat / RCA endpoint
    │  Uses CUSTOMER_LLM_API_KEY when configured
    ▼
LLM provider (OpenAI / Azure OpenAI / Anthropic — customer contract)
    ▼
Explanation returned to UI (not used for model training per DPA)
```

---

## 4. Alert notification flow

```text
Alert engine evaluation
    ▼
Notification adapters (Slack, email, ServiceNow, Jira, PagerDuty)
    │  Outbound HTTPS from customer network or vendor VPC
    ▼
Bank ticketing / chat systems
```

---

## 5. FinOps billing flow (optional module)

```text
Cloud provider billing export (CUR / BigQuery / Azure Cost Management)
    │  Customer IAM role — read-only billing
    ▼
FinOps connectors (scheduled ingest)
    ▼
Normalized line items → PostgreSQL aggregates
    ▼
FinOps API + UI (ABAC scoped; account IDs masked for non-admin)
```

**Production requirement:** Live billing feeds — not simulated data. See [FINOPS_ENHANCEMENT_REQUIREMENTS.md](../FINOPS_ENHANCEMENT_REQUIREMENTS.md).

---

## 6. Data residency

- Self-hosted: **all data remains in customer-selected region/DC**.
- Dedicated cloud: single-tenant VPC in agreed region.
- Configure `X-Region-Target` and ABAC residency policies where multi-region gateway is enabled.

---

## 7. Data deletion (offboarding)

1. Export incidents, audit logs, alert history (JSON/CSV APIs).
2. Disable tenant and revoke API keys.
3. Delete tenant indices and Postgres rows per retention policy.
4. Confirm backup expiry per [BACKUP_RESTORE.md](../runbooks/BACKUP_RESTORE.md).

---

*Attach this document to [SECURITY_PACK.md](./SECURITY_PACK.md) for architecture review meetings.*
