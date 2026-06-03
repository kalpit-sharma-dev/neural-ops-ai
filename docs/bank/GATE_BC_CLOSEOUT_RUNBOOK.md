# Gate B & Gate C — Sales / PS Close-Out Runbook

**Audience:** Sales engineering, professional services, customer champion  
**Roadmap:** [BANK_PRODUCTION_READINESS_ROADMAP.md](../BANK_PRODUCTION_READINESS_ROADMAP.md) §10

This is the **single checklist** to declare staging production-ready (Gate B) and bank POC-ready (Gate C).

---

## Quick verify (automated)

```bash
chmod +x scripts/gate-bc-verify.sh
export GATEWAY=https://api.staging.customer.example
export TOKEN=$STAGING_TOKEN
export TENANT=poc-bank
./scripts/gate-bc-verify.sh --gate B   # or --gate C
```

---

## Gate B — Staging production

**Exit:** Safe to run a **internal** 4-week POC on this cluster (not yet customer contractual prod).

### B1 — Platform & data (Wave 1)

| # | Check | Evidence | Doc |
|---|-------|----------|-----|
| B1.1 | Helm staging/prod values applied | `helm list -n neuralops` | [values-staging.yaml](../../infra/helm/neuralops/values-staging.yaml) |
| B1.2 | Managed Postgres/Kafka/ES/CH/Redis | Connection test from gateway pod | [MANAGED_DATA_PLANE.md](../runbooks/MANAGED_DATA_PLANE.md) |
| B1.3 | SSO only; dev login off | `./scripts/verify-production-auth.sh` exit 0 | [PRODUCTION_AUTH_CHECKLIST.md](./PRODUCTION_AUTH_CHECKLIST.md) |
| B1.4 | External Secrets / K8s secrets | No plaintext DSN in ConfigMap | [external-secret.yaml](../../infra/helm/neuralops/templates/external-secret.yaml) |
| B1.5 | mTLS or mesh documented | `verify-mtls.sh` OR Istio PeerAuthentication applied | [MTLS_SERVICE_MESH.md](../runbooks/MTLS_SERVICE_MESH.md) |

### B2 — Operability

| # | Check | Evidence | Doc |
|---|-------|----------|-----|
| B2.1 | Backup + restore drill | Ticket + restore timestamp | [BACKUP_RESTORE.md](../runbooks/BACKUP_RESTORE.md) |
| B2.2 | 7-day soak started | 7× `staging-soak-check.sh --day N` logs | [STAGING_SOAK_7DAY.md](../runbooks/STAGING_SOAK_7DAY.md) |
| B2.3 | Grafana/Prometheus scraping gateway | Dashboard shows request rate | `infra/grafana/` |
| B2.4 | PII scrubbing guide shared with ingest team | Sign-off email | [PII_SCRUBBING_GUIDE.md](./PII_SCRUBBING_GUIDE.md) |

### B3 — Engineering gates

| # | Check | Evidence |
|---|-------|----------|
| B3.1 | CI green on `main` | GitHub Actions link |
| B3.2 | Tenant isolation tests | `go test ./internal/finops/... -run TestTenantIsolation` |
| B3.3 | Internal security checklist ≥ 80% | [PRODUCTION_AND_GTM.md](../PRODUCTION_AND_GTM.md) §15 |

**Gate B sign-off**

| Role | Name | Date |
|------|------|------|
| Platform lead | | |
| Security delegate | | |

---

## Gate C — Bank POC ready

**Prerequisite:** Gate B complete.

**Exit:** Legal may send security pack; customer champion can start POC week 1.

### C1 — Legal & security pack (Wave 0)

| # | Check | Evidence |
|---|-------|----------|
| C1.1 | Security pack + data flow sent | Email / portal upload |
| C1.2 | Subprocessor list included | [SUBPROCESSORS.md](./SUBPROCESSORS.md) |
| C1.3 | POC scope signed | [POC_SCOPE_TEMPLATE.md](./POC_SCOPE_TEMPLATE.md) filled |
| C1.4 | Counsel reviewed DPA/MSA templates | Internal legal ticket (not sent until Gate D) |

### C2 — POC technical package (Wave 2)

| # | Check | Evidence | Doc |
|---|-------|----------|-----|
| C2.1 | Payments KPI pack enabled | `POST .../kpi-packs/kpi-bfsi-payments/enable` | [PAYMENTS_KPI_PACK.md](./PAYMENTS_KPI_PACK.md) |
| C2.2 | Log ingest runbook handed to customer | Workshop deck | [BANK_LOG_INGEST.md](../runbooks/BANK_LOG_INGEST.md) |
| C2.3 | ITSM smoke | `./scripts/verify-itsm-integration.sh` | [ITSM_INTEGRATION_JIRA_SERVICENOW.md](../runbooks/ITSM_INTEGRATION_JIRA_SERVICENOW.md) |
| C2.4 | k6 POC load pass | k6 summary `http_req_duration p(95)<2000` | [POC_LOAD_TEST.md](./POC_LOAD_TEST.md) |
| C2.5 | Incident + RCA playbook walkthrough | 30 min demo completed | [INCIDENT_RCA_PLAYBOOK.md](../runbooks/INCIDENT_RCA_PLAYBOOK.md) |
| C2.6 | ABAC / developer scope tests in CI | Latest CI run | Roadmap 2.6 |
| C2.7 | Air-gap doc if restricted network | Included in pack if applicable | [AIR_GAPPED_DEPLOYMENT.md](./AIR_GAPPED_DEPLOYMENT.md) |

### C3 — API contract

| # | Check | Evidence |
|---|-------|----------|
| C3.1 | OpenAPI synced | `./scripts/sync-openapi.sh` |
| C3.2 | Contract tests pass | `make contract-test` && `make openapi-test` |

### C4 — FinOps (if in POC scope)

| # | Check | Evidence |
|---|-------|----------|
| C4.1 | FinOps smoke | `./scripts/verify-finops-production.sh` |
| C4.2 | Reconciliation API returns snapshots | `GET /api/v1/finops/reconciliation` |
| C4.3 | Do **not** claim invoice-grade FinOps until Gate F | Messaging per roadmap §11 |

### C5 — Safety defaults

| # | Check | Evidence |
|---|-------|----------|
| C5.1 | AutoFix disabled in staging/prod | `AUTOFIX_ENABLED=false` | [AUTOFIX_BANK_POLICY.md](./AUTOFIX_BANK_POLICY.md) |
| C5.2 | Customer LLM keys documented | Workshop | [CUSTOMER_LLM_KEYS.md](./CUSTOMER_LLM_KEYS.md) |

**Gate C sign-off**

| Role | Name | Date |
|------|------|------|
| Account executive | | |
| Customer champion (bank) | | |
| NeuralOps SE | | |

---

## After Gate C (preview Gate D)

Do not skip to production contract until:

- Pen test complete — [PENETRATION_TEST_CHECKLIST.md](./PENETRATION_TEST_CHECKLIST.md)
- Signed DPA + MSA
- Helm on customer K8s — [HELM_PS_PLAYBOOK.md](./HELM_PS_PLAYBOOK.md)
- Support escalation agreed — [SUPPORT_ESCALATION.md](./SUPPORT_ESCALATION.md)

---

## Appendix — command reference

```bash
./scripts/verify-production-auth.sh
./scripts/verify-mtls.sh
./scripts/verify-itsm-integration.sh
./scripts/verify-finops-production.sh
./scripts/staging-soak-check.sh --day 1
k6 run scripts/k6/bank-poc-load.js
make contract-test openapi-test
./scripts/sync-openapi.sh
```
