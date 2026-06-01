#!/usr/bin/env bash
# Verifies Keycloak OIDC is reachable and matches gateway configuration.
set -euo pipefail

KEYCLOAK_BASE="${KEYCLOAK_BASE:-http://localhost:8088}"
REALM="${KEYCLOAK_REALM:-neuralops}"
GATEWAY_BASE="${GATEWAY_BASE:-http://localhost:8080}"
CLIENT_ID="${OIDC_CLIENT_ID:-neuralops-ui}"

DISCOVERY="${KEYCLOAK_BASE}/realms/${REALM}/.well-known/openid-configuration"

echo "Checking OIDC discovery: ${DISCOVERY}"
body="$(curl -fsS "${DISCOVERY}")"
echo "${body}" | grep -q '"issuer"'
echo "${body}" | grep -q '"authorization_endpoint"'
echo "${body}" | grep -q '"token_endpoint"'

echo "Checking gateway auth config..."
cfg="$(curl -fsS "${GATEWAY_BASE}/api/v1/auth/config")"
echo "${cfg}" | grep -q '"authEnabled":true' || echo "${cfg}" | grep -q '"authEnabled": true'

token_response="$(curl -fsS -X POST "${KEYCLOAK_BASE}/realms/${REALM}/protocol/openid-connect/token" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d "grant_type=password" \
  -d "client_id=${CLIENT_ID}" \
  -d "client_secret=${OIDC_CLIENT_SECRET:-neuralops-ui-secret}" \
  -d "username=demo@neuralops.ai" \
  -d "password=${KEYCLOAK_DEMO_PASSWORD:-demo1234}" \
  -d 'scope=openid profile email')"

echo "${token_response}" | grep -q '"access_token"'
echo "Keycloak OIDC verification passed."
