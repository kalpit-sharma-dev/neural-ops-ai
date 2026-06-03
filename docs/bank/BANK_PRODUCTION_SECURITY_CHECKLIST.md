# Bank Production Security Checklist (Gate D)

Maps [PRODUCTION_AND_GTM.md](../PRODUCTION_AND_GTM.md) §15 to bank artifacts. Score **≥ 80%** (weighted) before Gate D sign-off.

**Scoring:** Done = 1, Partial = 0.5, Todo = 0. Minimum 80% of applicable items.

## Infrastructure (weight 25%)

| # | Item | Status | Evidence |
|---|------|--------|----------|
| I1 | K8s private subnets | | |
| I2 | Managed Postgres/Kafka/ES/CH/Redis | | [MANAGED_DATA_PLANE.md](../runbooks/MANAGED_DATA_PLANE.md) |
| I3 | Ingress TLS + HSTS | | |
| I4 | Secrets in secret manager | | External Secrets |
| I5 | Backup + restore tested | | [BACKUP_RESTORE.md](../runbooks/BACKUP_RESTORE.md) |
| I6 | Monitoring + paging | | Grafana/PagerDuty |

## Application (weight 20%)

| # | Item | Status | Evidence |
|---|------|--------|----------|
| A1 | Dev auth disabled | | `verify-production-auth.sh` |
| A2 | Health/readiness probes | | `kubectl describe pod` |
| A3 | HPA + resource limits | | Helm values |
| A4 | Migrations on deploy | | Job logs |
| A5 | AutoFix off or waived | | [AUTOFIX_BANK_POLICY.md](./AUTOFIX_BANK_POLICY.md) |

## Security (weight 30%)

| # | Item | Status | Evidence |
|---|------|--------|----------|
| S1 | SSO enforced | | [SSO_SETUP_BANK.md](../runbooks/SSO_SETUP_BANK.md) |
| S2 | API key rotation process | | Runbook |
| S3 | Audit logging | | Gateway audit + FinOps export |
| S4 | NetworkPolicies / mesh mTLS | | [MTLS_SERVICE_MESH.md](../runbooks/MTLS_SERVICE_MESH.md) |
| S5 | CI dependency scan (Trivy) | | GitHub Actions |
| S6 | Pen test no Critical open | | [PENETRATION_TEST_CHECKLIST.md](./PENETRATION_TEST_CHECKLIST.md) |
| S7 | Tenant isolation tests CI | | `tenant_isolation_test.go` |
| S8 | ABAC configured | | Enterprise governance UI |

## Legal / commercial (weight 15%)

| # | Item | Status | Evidence |
|---|------|--------|----------|
| L1 | DPA executed | | [DPA_MSA_SIGNOFF_TRACKER.md](./DPA_MSA_SIGNOFF_TRACKER.md) |
| L2 | MSA executed | | |
| L3 | Subprocessors disclosed | | [SUBPROCESSORS.md](./SUBPROCESSORS.md) |
| L4 | Support SLA | | [SUPPORT_ESCALATION.md](./SUPPORT_ESCALATION.md) |

## Data governance (weight 10%)

| # | Item | Status | Evidence |
|---|------|--------|----------|
| D1 | PII scrubbing | | [PII_SCRUBBING_GUIDE.md](./PII_SCRUBBING_GUIDE.md) |
| D2 | Data residency configured | | [MULTI_REGION_RESIDENCY.md](../runbooks/MULTI_REGION_RESIDENCY.md) |
| D3 | Customer LLM keys | | [CUSTOMER_LLM_KEYS.md](./CUSTOMER_LLM_KEYS.md) |

## Score

| Section | Score / max |
|---------|-------------|
| Infrastructure | /6 |
| Application | /5 |
| Security | /8 |
| Legal | /4 |
| Data governance | /3 |
| **Total %** | |

**Signed:** __________________ Date: __________
