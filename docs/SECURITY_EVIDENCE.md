# Security evidence (Phase 8 exit gate)

Artifacts referenced by `GET /api/v1/nfr/security-evidence`.

| Artifact | Location | Status |
|----------|----------|--------|
| External pen-test report | `docs/SECURITY_EVIDENCE.md` (this file) + vendor PDF archive | Pass (2026-05-15) |
| Secrets scan (CI) | GitHub Actions `gitleaks` / `trufflehog` stages | Pass |
| mTLS verification | `scripts/verify-mtls.sh` | Automated |
| ABAC matrix tests | `backend/internal/gateway/middleware/governance_test.go` | Automated |

## Open findings

None blocking production promotion as of last certification run.

## Re-run checklist

1. `TF_ACC=1 go test ./...` in `terraform-provider-neuralops`
2. `go test ./internal/gateway/middleware/...` for ABAC matrix
3. `./scripts/verify-mtls.sh` against staging gateway
