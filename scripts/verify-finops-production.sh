#!/usr/bin/env bash
# FinOps production config smoke (Wave 5).
set -euo pipefail

GATEWAY="${GATEWAY:-http://localhost:8080}"
TENANT="${TENANT:-default}"
TOKEN="${TOKEN:-${NEURALOPS_API_TOKEN:-}}"

auth=()
if [[ -n "$TOKEN" ]]; then
  auth=(-H "Authorization: Bearer $TOKEN")
fi
hdr=(-H "X-Tenant-ID: $TENANT" "${auth[@]}")

echo "== FinOps production smoke =="
echo "Mode: ${FINOPS_BILLING_MODE:-simulated}"

code=$(curl -s -o /dev/null -w "%{http_code}" "${hdr[@]}" -X POST "$GATEWAY/api/v1/finops/ingest/run" || echo "000")
if [[ "$code" != "200" ]]; then
  echo "FAIL: ingest/run HTTP $code"
  exit 1
fi
echo "OK: POST /finops/ingest/run -> $code"

code=$(curl -s -o /dev/null -w "%{http_code}" "${hdr[@]}" "$GATEWAY/api/v1/finops/reconciliation" || echo "000")
if [[ "$code" != "200" ]]; then
  echo "FAIL: reconciliation HTTP $code"
  exit 1
fi
echo "OK: GET /finops/reconciliation -> $code"

code=$(curl -s -o /dev/null -w "%{http_code}" "${hdr[@]}" "$GATEWAY/api/v1/finops/audit/export?limit=10" || echo "000")
if [[ "$code" != "200" ]]; then
  echo "FAIL: audit/export HTTP $code"
  exit 1
fi
echo "OK: GET /finops/audit/export -> $code"

echo "FinOps production smoke completed."
