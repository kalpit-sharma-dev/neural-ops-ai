#!/usr/bin/env bash
# Verifies production-safe auth environment (PROD-AUTH-01 / Wave 0.5).
set -euo pipefail

fail=0
warn=0

check_eq() {
  local name="$1" actual="$2" expected="$3"
  if [[ "${actual:-}" != "$expected" ]]; then
    echo "FAIL: $name expected '$expected' got '${actual:-<unset>}'"
    fail=$((fail + 1))
  else
    echo "OK: $name=$expected"
  fi
}

check_not_eq() {
  local name="$1" actual="$2" forbidden="$3"
  if [[ "${actual:-}" == "$forbidden" ]]; then
    echo "FAIL: $name must not be '$forbidden'"
    fail=$((fail + 1))
  else
    echo "OK: $name is not '$forbidden'"
  fi
}

ENV_NAME="${ENVIRONMENT:-development}"
echo "== NeuralOps production auth verification (ENVIRONMENT=$ENV_NAME) =="

if [[ "$ENV_NAME" == "production" || "$ENV_NAME" == "staging" ]]; then
  check_eq "AUTH_DISABLED" "${AUTH_DISABLED:-false}" "false"
  check_eq "AUTH_ALLOW_DEV_LOGIN" "${AUTH_ALLOW_DEV_LOGIN:-}" "false"
  check_eq "DEMO_MODE" "${DEMO_MODE:-}" "false"

  if [[ "${OIDC_ENABLED:-false}" != "true" && "${SAML_ENABLED:-false}" != "true" ]]; then
    echo "FAIL: OIDC_ENABLED or SAML_ENABLED must be true for $ENV_NAME"
    fail=$((fail + 1))
  else
    echo "OK: SSO enabled (OIDC=${OIDC_ENABLED:-false} SAML=${SAML_ENABLED:-false})"
  fi

  if [[ -z "${JWT_PRIVATE_KEY_PEM:-}" && -z "${JWT_PRIVATE_KEY_FILE:-}" ]]; then
    echo "WARN: JWT private key not set — use External Secrets in real prod"
    warn=$((warn + 1))
  fi
else
  echo "SKIP: non-prod environment — only warnings"
  if [[ "${AUTH_DISABLED:-}" == "true" ]]; then
    echo "WARN: AUTH_DISABLED=true (acceptable for local dev only)"
    warn=$((warn + 1))
  fi
fi

echo ""
if [[ $fail -gt 0 ]]; then
  echo "Result: FAILED ($fail errors, $warn warnings)"
  exit 1
fi
echo "Result: PASSED ($warn warnings)"
exit 0
