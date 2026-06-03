#!/usr/bin/env bash
# Verifies ITSM integration endpoints are reachable (Wave 2.3).
set -euo pipefail

GATEWAY="${GATEWAY:-http://localhost:8080}"
TOKEN="${TOKEN:-${NEURALOPS_API_TOKEN:-}}"
TENANT="${TENANT:-default}"

auth=()
if [[ -n "$TOKEN" ]]; then
  auth=(-H "Authorization: Bearer $TOKEN")
fi
hdr=(-H "X-Tenant-ID: $TENANT" "${auth[@]}")

echo "== ITSM integration smoke =="
echo "Gateway: $GATEWAY Tenant: $TENANT"

code=$(curl -s -o /dev/null -w "%{http_code}" "${hdr[@]}" \
  "$GATEWAY/api/v1/integrations" || echo "000")
if [[ "$code" != "200" ]]; then
  echo "FAIL: list integrations HTTP $code"
  exit 1
fi
echo "OK: GET /api/v1/integrations -> $code"

for key in jira servicenow; do
  code=$(curl -s -o /dev/null -w "%{http_code}" "${hdr[@]}" \
    "$GATEWAY/api/v1/integrations/$key/oauth/start" || echo "000")
  # 302 redirect to IdP, or 400 if oauthClientId not configured yet
  if [[ "$code" == "302" || "$code" == "400" ]]; then
    echo "OK: $key oauth/start -> $code (configure client id for 302)"
  else
    echo "WARN: $key oauth/start unexpected HTTP $code"
  fi
done

code=$(curl -s -o /dev/null -w "%{http_code}" "${hdr[@]}" \
  "$GATEWAY/api/v1/workflows" || echo "000")
if [[ "$code" != "200" ]]; then
  echo "FAIL: list workflows HTTP $code"
  exit 1
fi
echo "OK: GET /api/v1/workflows -> $code"

echo "ITSM smoke completed."
