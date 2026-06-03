# Production Authentication Checklist (PROD-AUTH-01)

**Wave 0.5 / Wave 1.2** — Required before staging or bank production accepts traffic.

---

## Environment variables (gateway)

| Variable | Production | Staging | Local dev |
|----------|------------|---------|-----------|
| `ENVIRONMENT` | `production` | `staging` | `development` |
| `AUTH_DISABLED` | `false` | `false` | `false` |
| `AUTH_ALLOW_DEV_LOGIN` | `false` | `false` | `true` (optional) |
| `DEMO_MODE` | `false` | `false` | `true` (optional) |
| `OIDC_ENABLED` or `SAML_ENABLED` | **At least one `true`** | Same | Optional |
| `JWT_PRIVATE_KEY_PEM` / files | Required (not default dev keys) | Required | Dev keys OK |

Gateway **refuses to start** when `ENVIRONMENT=production` and unsafe auth flags are set (see `config.ValidateProduction`).

---

## Frontend

| Variable | Production |
|----------|------------|
| `VITE_DEMO_MODE` | unset or `false` |
| Demo banner | Must not appear |

---

## Helm

Use `values-prod.yaml` or `values-staging.yaml`:

```yaml
config:
  demoMode: false
  authDisabled: false
  authAllowDevLogin: false
auth:
  oidc:
    enabled: true  # or saml.enabled: true
```

---

## Verification

```bash
# Against running gateway or env file
./scripts/verify-production-auth.sh

# Or at gateway startup — fails fast on misconfiguration
ENVIRONMENT=production AUTH_ALLOW_DEV_LOGIN=true go run ./cmd/gateway
# Expected: error exit
```

---

## Manual checks

- [ ] No shared `demo@neuralops.ai` credentials communicated to bank users
- [ ] API keys rotated from any POC/demo keys
- [ ] IdP MFA enforced at identity provider
- [ ] Session timeout aligned with bank policy (JWT TTL)
- [ ] CORS `allowed_origins` lists only bank UI domains

---

## Sign-off

| Role | Name | Date |
|------|------|------|
| Platform engineering | | |
| Security | | |
