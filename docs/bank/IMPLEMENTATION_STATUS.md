# Bank Production — Implementation Status Matrix

**Updated:** 2026-06-03  
**Closure tracker:** [OPEN_POINTS_CLOSURE.md](./OPEN_POINTS_CLOSURE.md)

Legend: ✅ Repo complete | 🟡 Execute on staging/customer | ⬜ Ops-only (legal/audit/vendor)

---

## Summary

| Category | Status |
|----------|--------|
| Section 7 code backlog | ✅ **Closed in repo** (APIs, UI, tests, scripts) |
| Wave 0–3 artifacts | ✅ Complete |
| Wave 4–6 | ✅ Repo complete; SOC2 Type II + coverage 80% ongoing |
| Gates B–F | 🟡 Automated scripts pass; manual sign-offs remain |
| FinOps live billing | 🟡 Customer CUR wiring (FIN-PROD-01–04) |

---

## Remaining ops-only (cannot code)

| ID | Owner |
|----|-------|
| BANK-005 | Security — pen test vendor |
| BANK-003/004 sign-off | Legal |
| BANK-011/012 | GRC — SOC 2 audit |
| NFR-01 | SRE — 7-day soak execution |
| FIN-PROD-01–04 | Finance + customer cloud |
| Gate D–F | Leadership manual checklists |

Verify repo closure: `./scripts/gate-verify.sh --gate C`
