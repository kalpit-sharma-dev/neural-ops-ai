# Bank & BFSI Customer Pack

Artifacts for **Wave 0–6** and **Gates A–F** of [BANK_PRODUCTION_READINESS_ROADMAP.md](../BANK_PRODUCTION_READINESS_ROADMAP.md).

**Status matrix:** [IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md) — Done vs Ops vs Phase 2.

## Gate close-out (sales, PS, legal)

| Gate | Document | Script |
|------|----------|--------|
| B / C | [GATE_BC_CLOSEOUT_RUNBOOK.md](./GATE_BC_CLOSEOUT_RUNBOOK.md) | `./scripts/gate-verify.sh --gate B` |
| D | [GATE_D_CLOSEOUT_RUNBOOK.md](./GATE_D_CLOSEOUT_RUNBOOK.md) | `./scripts/gate-verify.sh --gate D` |
| E / F | [GATE_EF_CLOSEOUT_RUNBOOK.md](./GATE_EF_CLOSEOUT_RUNBOOK.md) | `./scripts/gate-verify.sh --gate E` or `--gate F` |

| Support | Document |
|---------|----------|
| Legal tracker | [DPA_MSA_SIGNOFF_TRACKER.md](./DPA_MSA_SIGNOFF_TRACKER.md) |
| Security scorecard | [BANK_PRODUCTION_SECURITY_CHECKLIST.md](./BANK_PRODUCTION_SECURITY_CHECKLIST.md) |
| Helm rollback | [../runbooks/HELM_ROLLBACK_DRILL.md](../runbooks/HELM_ROLLBACK_DRILL.md) |

## Send to bank security (before POC)

| Document | Purpose |
|----------|---------|
| [SECURITY_PACK.md](./SECURITY_PACK.md) | Architecture, auth, encryption, audit |
| [DATA_FLOW.md](./DATA_FLOW.md) | Data flow diagrams |
| [SUBPROCESSORS.md](./SUBPROCESSORS.md) | Subprocessor list |
| [POC_SCOPE_TEMPLATE.md](./POC_SCOPE_TEMPLATE.md) | Fill per deal |
| [AIR_GAPPED_DEPLOYMENT.md](./AIR_GAPPED_DEPLOYMENT.md) | Restricted network topology |

## Legal (counsel review required)

| Document | Purpose |
|----------|---------|
| [DPA_TEMPLATE.md](./DPA_TEMPLATE.md) | Data processing agreement |
| [MSA_ELA_TEMPLATE.md](./MSA_ELA_TEMPLATE.md) | Enterprise license |
| [DPA_MSA_SIGNOFF_TRACKER.md](./DPA_MSA_SIGNOFF_TRACKER.md) | Execution tracking |

## Engineering / operations

| Document | Purpose |
|----------|---------|
| [PRODUCTION_AUTH_CHECKLIST.md](./PRODUCTION_AUTH_CHECKLIST.md) | Prod auth env vars |
| [PII_SCRUBBING_GUIDE.md](./PII_SCRUBBING_GUIDE.md) | Log scrubbing |
| [CUSTOMER_LLM_KEYS.md](./CUSTOMER_LLM_KEYS.md) | Bank-owned LLM keys |
| [PAYMENTS_KPI_PACK.md](./PAYMENTS_KPI_PACK.md) | BFSI KPI enablement |
| [POC_LOAD_TEST.md](./POC_LOAD_TEST.md) | k6 POC volume profile |
| [FINOPS_PRODUCTION.md](./FINOPS_PRODUCTION.md) | Live billing + reconciliation |
| [AUTOFIX_BANK_POLICY.md](./AUTOFIX_BANK_POLICY.md) | AutoFix disabled by default |
| [MSP_WHITE_LABEL.md](./MSP_WHITE_LABEL.md) | MSP / branding APIs |
| [PENETRATION_TEST_CHECKLIST.md](./PENETRATION_TEST_CHECKLIST.md) | Pen test scope |
| [AIR_GAP_INSTALL.md](./AIR_GAP_INSTALL.md) | Offline license install |
| [HELM_PS_PLAYBOOK.md](./HELM_PS_PLAYBOOK.md) | 5–10 day Helm PS |
| [SUPPORT_ESCALATION.md](./SUPPORT_ESCALATION.md) | Sev / 24×7 |
| [../runbooks/FINOPS_CUR_S3_SYNC.md](../runbooks/FINOPS_CUR_S3_SYNC.md) | S3 CUR → PVC sync |
| [../runbooks/NPM_SUITE.md](../runbooks/NPM_SUITE.md) | NPM pilot scope |
| [../runbooks/BANK_LOG_INGEST.md](../runbooks/BANK_LOG_INGEST.md) | Fluent Bit / OTEL / Kafka |
| [../runbooks/ITSM_INTEGRATION_JIRA_SERVICENOW.md](../runbooks/ITSM_INTEGRATION_JIRA_SERVICENOW.md) | Jira / ServiceNow |
| [../runbooks/INCIDENT_RCA_PLAYBOOK.md](../runbooks/INCIDENT_RCA_PLAYBOOK.md) | POC incident walkthrough |
| [../runbooks/MTLS_SERVICE_MESH.md](../runbooks/MTLS_SERVICE_MESH.md) | East-west mTLS / Istio |
| [../runbooks/MANAGED_DATA_PLANE.md](../runbooks/MANAGED_DATA_PLANE.md) | RDS / MSK / ES / CH / Redis |
| [../runbooks/STAGING_SOAK_7DAY.md](../runbooks/STAGING_SOAK_7DAY.md) | Gate B soak |
| [../runbooks/BACKUP_RESTORE.md](../runbooks/BACKUP_RESTORE.md) | DR drill |
| [../runbooks/SSO_SETUP_BANK.md](../runbooks/SSO_SETUP_BANK.md) | OIDC/SAML setup |

## Compliance worksheets

| Document | Purpose |
|----------|---------|
| [../compliance/SOC2_TYPE1_POLICY_INDEX.md](../compliance/SOC2_TYPE1_POLICY_INDEX.md) | SOC 2 Type I index |
| [../compliance/SOC2_TYPE2_READINESS.md](../compliance/SOC2_TYPE2_READINESS.md) | Type II prep |
| [../compliance/FFIEC_CONTROL_MAPPING.md](../compliance/FFIEC_CONTROL_MAPPING.md) | FFIEC mapping |
| [../compliance/COVERAGE_ROADMAP.md](../compliance/COVERAGE_ROADMAP.md) | 80% coverage plan |
| [../roadmap/LIVE_LOG_TAIL.md](../roadmap/LIVE_LOG_TAIL.md) | Live tail commitment |
| [../roadmap/NEXQL_UNIFIED_QUERY.md](../roadmap/NEXQL_UNIFIED_QUERY.md) | Unified query roadmap |
| [../roadmap/LLM_OBSERVABILITY.md](../roadmap/LLM_OBSERVABILITY.md) | LLM usage APIs |

## Scripts

```bash
chmod +x scripts/gate-verify.sh scripts/gate-bc-verify.sh scripts/gate-d-verify.sh \
  scripts/gate-ef-verify.sh scripts/helm-rollback-drill.sh \
  scripts/verify-production-auth.sh scripts/verify-itsm-integration.sh \
  scripts/staging-soak-check.sh scripts/verify-finops-production.sh

./scripts/gate-verify.sh --gate C
./scripts/sync-openapi.sh
```

## Helm (Wave 1 / Gate D)

```bash
helm upgrade -i neuralops infra/helm/neuralops \
  -f infra/helm/neuralops/values.yaml \
  -f infra/helm/neuralops/values-prod.yaml
```

## Terraform (Wave 3.7)

Resources: `neuralops_alert_policy`, `neuralops_slo`, `neuralops_abac_policy`, `neuralops_msp_tenant`, `neuralops_finops_budget`, `neuralops_finops_allocation_rule`, `neuralops_branding`, `neuralops_export_job`.
