# Gate D — First Bank Production Close-Out

**Audience:** Professional services, customer platform team, legal, security  
**Prerequisite:** [Gate C](./GATE_BC_CLOSEOUT_RUNBOOK.md) signed  
**Roadmap:** [BANK_PRODUCTION_READINESS_ROADMAP.md](../BANK_PRODUCTION_READINESS_ROADMAP.md) § Gate D

**Exit:** Contractual production on **customer Kubernetes** (or dedicated VPC) with signed legal and operational readiness.

---

## Automated verify

```bash
chmod +x scripts/gate-d-verify.sh scripts/helm-rollback-drill.sh
export GATEWAY=https://api.production.bank.internal
export TOKEN=$PROD_TOKEN
export TENANT=prod-bank
export HELM_RELEASE=neuralops
export HELM_NAMESPACE=neuralops
./scripts/gate-d-verify.sh
```

---

## D1 — Legal & commercial

| # | Check | Evidence | Artifact |
|---|-------|----------|----------|
| D1.1 | Gate C complete | Signed Gate C table | [GATE_BC_CLOSEOUT_RUNBOOK.md](./GATE_BC_CLOSEOUT_RUNBOOK.md) |
| D1.2 | DPA executed | Counter-signed PDF | [DPA_TEMPLATE.md](./DPA_TEMPLATE.md) |
| D1.3 | MSA / ELA executed | Counter-signed PDF | [MSA_ELA_TEMPLATE.md](./MSA_ELA_TEMPLATE.md) |
| D1.4 | Subprocessors accepted | Email or contract exhibit | [SUBPROCESSORS.md](./SUBPROCESSORS.md) |
| D1.5 | Support SLA + escalation | Ticket + order form | [SUPPORT_ESCALATION.md](./SUPPORT_ESCALATION.md) |

Tracker: [DPA_MSA_SIGNOFF_TRACKER.md](./DPA_MSA_SIGNOFF_TRACKER.md)

---

## D2 — Security

| # | Check | Evidence | Artifact |
|---|-------|----------|----------|
| D2.1 | Pen test complete; no open **Critical** | Retest letter | [PENETRATION_TEST_CHECKLIST.md](./PENETRATION_TEST_CHECKLIST.md) |
| D2.2 | Production auth | `ENVIRONMENT=production ./scripts/verify-production-auth.sh` | [PRODUCTION_AUTH_CHECKLIST.md](./PRODUCTION_AUTH_CHECKLIST.md) |
| D2.3 | Customer SSO live (OIDC/SAML) | Login test + IdP app config | [SSO_SETUP_BANK.md](../runbooks/SSO_SETUP_BANK.md) |
| D2.4 | AutoFix disabled unless CISO waiver | `AUTOFIX_ENABLED=false` | [AUTOFIX_BANK_POLICY.md](./AUTOFIX_BANK_POLICY.md) |
| D2.5 | PII scrubbing on ingest path | Spot-check logs | [PII_SCRUBBING_GUIDE.md](./PII_SCRUBBING_GUIDE.md) |
| D2.6 | Bank security checklist ≥ 80% | Scored sheet | [BANK_PRODUCTION_SECURITY_CHECKLIST.md](./BANK_PRODUCTION_SECURITY_CHECKLIST.md) |

---

## D3 — Helm production deploy

| # | Check | Evidence | Artifact |
|---|-------|----------|----------|
| D3.1 | PS playbook days 1–5 complete | As-built diagram | [HELM_PS_PLAYBOOK.md](./HELM_PS_PLAYBOOK.md) |
| D3.2 | `values-prod.yaml` + customer overrides | `helm get values` output | [values-prod.yaml](../../infra/helm/neuralops/values-prod.yaml) |
| D3.3 | Image digest pinned (not `:latest`) | values `imageTag` / digest | |
| D3.4 | FINOPS live billing (if in scope) | CUR sync + reconciliation | [FINOPS_PRODUCTION.md](./FINOPS_PRODUCTION.md), [FINOPS_CUR_S3_SYNC.md](../runbooks/FINOPS_CUR_S3_SYNC.md) |
| D3.5 | Air-gap / restricted net (if applicable) | NetworkPolicy evidence | [AIR_GAPPED_DEPLOYMENT.md](./AIR_GAPPED_DEPLOYMENT.md) |
| D3.6 | Rollback drill **&lt; 15 min** | Timer + `helm history` | [HELM_ROLLBACK_DRILL.md](../runbooks/HELM_ROLLBACK_DRILL.md) |

```bash
./scripts/helm-rollback-drill.sh --dry-run   # then live in change window
```

---

## D4 — Data plane & DR

| # | Check | Evidence | Artifact |
|---|-------|----------|----------|
| D4.1 | Managed HA stores live | Connection tests | [MANAGED_DATA_PLANE.md](../runbooks/MANAGED_DATA_PLANE.md) |
| D4.2 | `FINOPS_REQUIRE_POSTGRES=true` (if FinOps prod) | Gateway env | [FINOPS_PRODUCTION.md](./FINOPS_PRODUCTION.md) |
| D4.3 | Backup restore drill post-go-live | Ticket within 30 days | [BACKUP_RESTORE.md](../runbooks/BACKUP_RESTORE.md) |
| D4.4 | No demo seed in prod namespace | `demoMode: false` | |

---

## D5 — Operability

| # | Check | Evidence |
|---|-------|----------|
| D5.1 | On-call rotation + PagerDuty/ServiceNow wired | Schedule export |
| D5.2 | Runbooks handed to customer NOC | Index in [README.md](./README.md) |
| D5.3 | Grafana dashboards imported | Screenshot / UID list |
| D5.4 | Change advisory board process for upgrades | Customer CAB ticket template |

---

## Gate D sign-off

| Role | Organization | Name | Date |
|------|--------------|------|------|
| Customer CISO delegate | Bank | | |
| Customer platform lead | Bank | | |
| NeuralOps account executive | NeuralOps | | |
| NeuralOps engineering lead | NeuralOps | | |

---

## After Gate D

- Begin SOC 2 Type II observation period — [SOC2_TYPE2_READINESS.md](../compliance/SOC2_TYPE2_READINESS.md)
- Plan Gate E after **2+** production customers and 3 months SLO data
- FinOps finance sign-off → [GATE_EF_CLOSEOUT_RUNBOOK.md](./GATE_EF_CLOSEOUT_RUNBOOK.md) Gate F
