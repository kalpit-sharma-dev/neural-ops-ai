# Open Points Closure Tracker

**Updated:** 2026-06-03  
**Purpose:** Single view of what is **closed in repo** vs **requires customer/ops execution**.

**Engineering backlog (implement one-by-one):** [GAP_IMPLEMENTATION_PLAN.md](../GAP_IMPLEMENTATION_PLAN.md) · [GAP_IMPLEMENTATION_STATUS.md](../GAP_IMPLEMENTATION_STATUS.md)

---

## Closed in repo (no further engineering)

All Section 7 items marked **Done** in [BANK_PRODUCTION_READINESS_ROADMAP.md](../BANK_PRODUCTION_READINESS_ROADMAP.md) have API/UI/tests/scripts in this repository.

Verify with:

```bash
./scripts/gate-verify.sh --gate C
cd backend && go test ./internal/observability/... ./tests/contract/... -count=1
```

---

## Repo-ready — ops/customer to execute

| ID | Action | Owner | Script / doc |
|----|--------|-------|--------------|
| BANK-003/004 | Legal sign DPA + MSA | Legal | [DPA_MSA_SIGNOFF_TRACKER.md](./DPA_MSA_SIGNOFF_TRACKER.md) |
| BANK-005 | Schedule pen test; remediate findings | Security | [PENETRATION_TEST_CHECKLIST.md](./PENETRATION_TEST_CHECKLIST.md) |
| BANK-009 | Run quarterly backup/restore drill | SRE | [BACKUP_RESTORE.md](../runbooks/BACKUP_RESTORE.md), `scripts/bank-drill-record.sh` |
| BANK-010 | Engineering leadership sign RPO/RTO | SRE | [RPO_RTO.md](./RPO_RTO.md) |
| BANK-011/012 | SOC 2 Type I / II audit | Security/GRC | [SOC2_TYPE1_POLICY_INDEX.md](../compliance/SOC2_TYPE1_POLICY_INDEX.md) |
| BANK-014 | Run correlator at agreed log volume | Engineering | `scripts/k6/bank-ingest-load.js` |
| NFR-01 | 7-day staging soak | SRE | `scripts/staging-soak-check.sh` |
| FIN-PROD-01–04 | Wire live CUR/BQ/Azure + invoice sign-off | Finance | [FINOPS_PRODUCTION.md](./FINOPS_PRODUCTION.md) |
| Gate B–F | Manual checklist sign-offs | Leadership | Gate closeout runbooks |

---

## Depth pack (optional engineering — added)

| Artifact | Closes |
|----------|--------|
| `GET /finops/billing/sources` | FIN-PROD-01..03 live file health |
| `scripts/verify-finops-live.sh` | Gate F automated smoke with sample CUR/GCP/Azure |
| `scripts/k6/bank-search-90d.js` | LOG-02 / NFR-02 90d search p95 |
| `scripts/verify-search-90d.sh` | 90d search gate runner |
| `GET /nfr/search-perf` | Search perf evidence API |
| `GET /apm/databases/:id/statements` | APM-05 DBM with traceId |
| `GET /network/flows/ebpf` | NPM-01 eBPF flow samples |

---

## Cannot close without external proof

- Petabyte-scale ingest at contract volume
- 80% repo-wide coverage gate
- SOC 2 / ISO certification claims
- FinOps invoice reconciliation sign-off (Gate F)
