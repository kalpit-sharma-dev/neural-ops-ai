# NeuralOps — Subprocessor List

**Version:** 1.0  
**Date:** 2026-06-03  
**Status:** Draft — publish on website before EU/UK customers  

Update this list when infrastructure or LLM defaults change.

---

## 1. Scope

Subprocessors are third parties that process **customer personal data** on behalf of NeuralOps when operating **NeuralOps Cloud** or **Dedicated Cloud**. For **self-hosted Enterprise**, the customer operates infrastructure; subprocessors below apply only to optional vendor-operated components (e.g. support systems, license validation if enabled).

---

## 2. Infrastructure subprocessors (Cloud / Dedicated)

| Subprocessor | Purpose | Data processed | Location |
|--------------|---------|----------------|----------|
| Amazon Web Services (or equivalent) | Compute, networking, storage | Telemetry, metadata, audit | Region per contract |
| Managed PostgreSQL (RDS / Cloud SQL / Azure DB) | Relational data | Users, incidents, alerts, config | Same region |
| Managed Kafka (MSK / Confluent) | Event streaming | Log/metric event payloads | Same region |
| Elasticsearch / OpenSearch service | Log search | Log content | Same region |
| ClickHouse Cloud or self-managed | Analytics | Aggregates, audit replica | Same region |
| Redis (ElastiCache / Memorystore) | Rate limits, cache | Tenant IDs, counters | Same region |

*Exact providers depend on deployment manifest — customer approves list in order form.*

---

## 3. Optional subprocessors

| Subprocessor | Purpose | When used |
|--------------|---------|-----------|
| OpenAI / Anthropic / Azure OpenAI | AI explanations | Only if customer enables vendor-managed LLM key |
| Stripe | Subscription billing | SaaS SKU only |
| Email provider (SendGrid / SES) | Alert email delivery | If email channel enabled |
| PagerDuty / Slack / Jira / ServiceNow | Notifications | Customer integration endpoints |

**Bank recommendation:** Use **customer-managed LLM keys** and **customer-owned** Slack/ServiceNow — removes LLM vendor from NeuralOps subprocessor list.

---

## 4. NeuralOps personnel access

- **Support access:** Break-glass only, time-bound, customer-approved for Dedicated/self-hosted PS engagements.
- **No** routine access to production log content without ticket authorization.

---

## 5. Change notification

NeuralOps will notify customers **30 days** before adding a new subprocessor that processes personal data (per DPA).

---

*Legal must approve before external publication.*
