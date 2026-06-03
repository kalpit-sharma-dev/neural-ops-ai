# Test Coverage Roadmap — 80% Repo-Wide (Wave 4.4)

**ID:** NFR-05

## Current state

CI enforces **tiered** gates via `scripts/coverage-gate.sh`:

| Package | Minimum |
|---------|---------|
| `internal/ingestion/parser` | 80% |
| `internal/gateway/middleware` | 20% |
| `internal/gateway/auth` | 25% |
| `internal/security` | 55% |
| `internal/observability` | 25% (raised Wave 4) |

Full `go test ./...` runs on every PR but is not 80% gated repo-wide yet.

## Phased plan

| Phase | Target | Actions |
|-------|--------|---------|
| Q2 2026 | 40% aggregate core | Raise middleware → 35%, observability → 35% |
| Q3 2026 | 60% | Handler integration tests for phase5–8 APIs |
| Q4 2026 | 80% | Repo-wide gate in CI (`MIN_COVERAGE=80` all packages) |

## Priority packages for banks

1. `internal/gateway/middleware` — auth, governance, quota
2. `internal/finops` — tenant isolation, budgets
3. `internal/workflow` — ITSM steps
4. `internal/observability` — ABAC, residency

## Local check

```bash
cd backend && go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -1
bash ../scripts/coverage-gate.sh
```

## CI

Coverage gate runs in `backend-unit` job — failing PRs cannot merge.
