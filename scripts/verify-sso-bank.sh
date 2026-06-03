#!/usr/bin/env bash
# Verifies bank SSO configuration for staging/production (BANK-007).
set -euo pipefail

fail=0
ENV_NAME="${ENVIRONMENT:-development}"

echo "== NeuralOps bank SSO verification (ENVIRONMENT=$ENV_NAME) =="

if [[ "$ENV_NAME" != "production" && "$ENV_NAME" != "staging" ]]; then
  echo "SKIP: SSO checks apply to staging/production only"
  exit 0
fi

if [[ "${AUTH_DISABLED:-false}" == "true" ]]; then
  echo "FAIL: AUTH_DISABLED must be false"
  fail=$((fail + 1))
fi

if [[ "${AUTH_ALLOW_DEV_LOGIN:-false}" == "true" ]]; then
  echo "FAIL: AUTH_ALLOW_DEV_LOGIN must be false"
  fail=$((fail + 1))
fi

if [[ "${OIDC_ENABLED:-false}" != "true" && "${SAML_ENABLED:-false}" != "true" ]]; then
  echo "FAIL: enable OIDC_ENABLED or SAML_ENABLED"
  fail=$((fail + 1))
else
  echo "OK: SSO provider enabled"
fi

if [[ "${OIDC_ENABLED:-false}" == "true" ]]; then
  for var in OIDC_ISSUER_URL OIDC_CLIENT_ID OIDC_REDIRECT_URL; do
    if [[ -z "${!var:-}" ]]; then
      echo "FAIL: $var required when OIDC enabled"
      fail=$((fail + 1))
    else
      echo "OK: $var set"
    fi
  done
fi

if [[ "${SAML_ENABLED:-false}" == "true" ]]; then
  for var in SAML_IDP_METADATA_URL SAML_SP_ENTITY_ID; do
    if [[ -z "${!var:-}" ]]; then
      echo "FAIL: $var required when SAML enabled"
      fail=$((fail + 1))
    else
      echo "OK: $var set"
    fi
  done
fi

if [[ $fail -gt 0 ]]; then
  echo "== SSO verification FAILED ($fail issues) =="
  exit 1
fi
echo "== SSO verification PASSED =="
