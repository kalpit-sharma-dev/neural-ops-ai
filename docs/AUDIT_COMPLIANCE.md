# Audit-Grade Spec Compliance

This document maps **Phase 20 quality gates** and **Phase 18 auth hardening** from `cursor_prompt_ai_log_analyzer.md` to implemented artifacts and CI enforcement.

## Quality gates (Phase 20)

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Unit test coverage gate | Enforced | `scripts/coverage-gate.sh` — parser ≥80%, auth ≥25%, middleware ≥20%, security ≥55% |
| Integration tests (Docker) | Enforced | `backend/tests/integration/*` — seed→ES, incident→alert, auth, Kafka→analysis→ES |
| Load test harness | Available | `make loadtest` (100k logs/sec target) |
| CI on every PR/push | Enforced | `.github/workflows/ci.yml` |
| Playwright E2E | Enforced in CI | `frontend/e2e/` run against live compose stack (port 3000) |

## End-to-end pipeline tests

| Flow | Test file |
|------|-----------|
| Seed → Elasticsearch → search | `platform_flow_test.go`, `ingestion_pipeline_test.go` |
| Kafka raw log → analysis → ES search | `kafka_es_pipeline_test.go` |
| Incident → alert deduplication | `platform_flow_test.go` |
| Auth / RBAC (Postgres) | `auth_flow_test.go` |

CI job `backend-integration` runs all integration tests with `-tags=integration`.

## Auth & SSO (Phase 18)

| Capability | Implementation | Live verification |
|------------|----------------|-------------------|
| OIDC / Keycloak | `backend/internal/gateway/auth/oidc_flow.go`, realm import `infra/keycloak/neuralops-realm.json` | `scripts/verify-keycloak-oidc.sh` (discovery + password grant) |
| Production SAML | `github.com/crewjam/saml` via `saml_crewjam.go` when `SAML_METADATA_URL` set | `scripts/verify-saml-metadata.sh` |
| Dev SAML fallback | `saml.go` (unsigned XML parser) | Used when metadata URL unset |
| RBAC / permissions | `middleware/rbac.go`, `permission.go`, `developer_scope.go` | Unit tests in `middleware/*_test.go` |
| mTLS client (gateway → upstream) | `gateway/proxy/mtls.go` | Compose mounts `infra/certs/` |
| mTLS server verify | `platform/mtls/tls_config.go`, nginx `infra/mtls-proxy/` | `scripts/verify-mtls.sh` |

### Quickstart verification

After `bash scripts/quickstart.sh`, the stack automatically runs:

1. OIDC discovery + token grant against Keycloak
2. SAML SP metadata from gateway (`SAML_ENABLED=true`, crewjam)
3. mTLS proxy rejects unauthenticated clients on `:8443`

### Demo credentials

- Dev login: `demo@neuralops.ai`, tenant `00000000-0000-0000-0000-000000000002`
- Keycloak: `demo@neuralops.ai` / `demo1234`, client `neuralops-ui`

## CI pipeline summary

```
backend-unit        → go test ./... + coverage-gate.sh
backend-integration → testcontainers integration suite
frontend            → npm ci && build
compose-smoke       → full stack + OIDC + SAML + mTLS + dev login + Playwright
```

## Known limits (honest audit notes)

- Aggregate backend coverage is below 80% repo-wide; gates target high-value packages only.
- Full browser OIDC redirect flow is not automated in Playwright (token grant script covers IdP contract).
- Internal services still speak HTTP on the Docker network; mTLS termination is demonstrated via the nginx sidecar, not on every microservice port.
- SAML browser SSO round-trip against Keycloak is manual; metadata + crewjam ACS parsing are production-grade.

## Commands

```bash
make test-coverage          # unit + coverage gate
make test-integration       # Docker integration tests
bash scripts/quickstart.sh  # stack + OIDC/SAML/mTLS verify
cd frontend && npm run test:e2e
```
