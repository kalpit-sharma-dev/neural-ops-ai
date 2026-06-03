# Gate E & Gate F — Enterprise Scale & FinOps Production

**Gate E:** Enterprise scale claims (multi-customer, SOC 2, SLO proof)  
**Gate F:** FinOps sold to finance with live billing reconciliation

---

## Gate E — Automated checks

```bash
./scripts/gate-ef-verify.sh --gate E
```

| # | Check | Evidence | Doc |
|---|-------|----------|-----|
| E1 | Gate D on **≥ 2** production customers | CRM / contracts | Roadmap § Gate E |
| E2 | SOC 2 Type I complete | Auditor letter | [SOC2_TYPE1_POLICY_INDEX.md](../compliance/SOC2_TYPE1_POLICY_INDEX.md) |
| E3 | SOC 2 Type II observation started | Audit plan | [SOC2_TYPE2_READINESS.md](../compliance/SOC2_TYPE2_READINESS.md) |
| E4 | SLOs met 3 consecutive months | Grafana SLO dashboard export | Ingress/gateway availability |
| E5 | k6 P95 &lt; 2s on search + finops | CI or quarterly k6 report | [POC_LOAD_TEST.md](./POC_LOAD_TEST.md), `.github/workflows/k6-bank-gate.yml` |
| E6 | Repo coverage roadmap on track | `COVERAGE_ROADMAP.md` | ≥ 60% phase target |
| E7 | Multi-region (if sold) | Residency API config | [MULTI_REGION_RESIDENCY.md](../runbooks/MULTI_REGION_RESIDENCY.md) |

**Sales:** May offer **multi-tenant cloud** only after Gate E (see roadmap §11).

---

## Gate F — FinOps production (finance sign-off)

```bash
export FINOPS_BILLING_MODE=live
export FINOPS_REQUIRE_POSTGRES=true
export FINOPS_ALLOW_SIMULATION=false
./scripts/gate-ef-verify.sh --gate F
```

| ID | Check | Pass criteria |
|----|-------|---------------|
| FIN-PROD-01 | AWS CUR live | `FINOPS_AWS_CUR_FILE` or S3 sync; ingest &gt; 0 lines |
| FIN-PROD-02 | GCP export | `FINOPS_GCP_BILLING_FILE` or hybrid scope documented |
| FIN-PROD-03 | Azure export | `FINOPS_AZURE_BILLING_FILE` or N/A with waiver |
| FIN-PROD-04 | Reconciliation | `driftPct` ≤ 1% per provider/month vs invoice |
| FIN-PROD-05 | Postgres only | No simulated bootstrap on empty tenant in prod |
| FIN-PROD-09 | Audit export | Finance received CSV; SIEM forward optional |

**Do not** claim FinOps parity in RFP until Gate F row FIN-PROD-01..04 are green for **that** bank's accounts.

---

## Combined sign-off

| Gate | Customer finance | Customer ops | NeuralOps |
|------|------------------|--------------|-----------|
| E | N/A | Platform lead | VP Engineering |
| F | Controller / FinOps lead | Cloud center of excellence | FinOps PM |
