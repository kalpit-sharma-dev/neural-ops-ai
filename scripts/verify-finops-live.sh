#!/usr/bin/env bash
# Live FinOps CUR/BQ/Azure wiring verification (FIN-PROD-01..04).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

GATEWAY="${GATEWAY:-http://localhost:8080}"
TENANT="${TENANT:-default}"
TOKEN="${TOKEN:-${NEURALOPS_API_TOKEN:-}}"

export FINOPS_BILLING_MODE="${FINOPS_BILLING_MODE:-live}"
export FINOPS_AWS_CUR_FILE="${FINOPS_AWS_CUR_FILE:-backend/internal/finops/connectors/testdata/aws_cur_sample.ndjson}"
export FINOPS_GCP_BILLING_FILE="${FINOPS_GCP_BILLING_FILE:-backend/internal/finops/connectors/testdata/gcp_billing_sample.ndjson}"
export FINOPS_AZURE_BILLING_FILE="${FINOPS_AZURE_BILLING_FILE:-backend/internal/finops/connectors/testdata/azure_billing_sample.ndjson}"
export FINOPS_AWS_INVOICE_USD="${FINOPS_AWS_INVOICE_USD:-730.75}"

auth=()
[[ -n "$TOKEN" ]] && auth=(-H "Authorization: Bearer $TOKEN")
hdr=(-H "X-Tenant-ID: $TENANT" "${auth[@]}")

echo "== FinOps live billing verification =="
echo "Mode: $FINOPS_BILLING_MODE"

code=$(curl -s -o /dev/null -w "%{http_code}" "${hdr[@]}" "$GATEWAY/api/v1/finops/billing/sources" || echo "000")
if [[ "$code" != "200" ]]; then
  echo "FAIL: GET /finops/billing/sources HTTP $code"
  exit 1
fi
echo "OK: GET /finops/billing/sources"

./scripts/verify-finops-production.sh

if command -v jq >/dev/null 2>&1; then
  body=$(curl -fsS "${hdr[@]}" "$GATEWAY/api/v1/finops/reconciliation" || echo "")
  drift=$(echo "$body" | jq '[.data[]?.driftPct // 0] | max // 0')
  echo "Reconciliation max drift: ${drift}%"
  awk -v d="$drift" 'BEGIN { if (d > 1.0) exit 1 }' || {
    echo "FAIL: drift ${drift}% exceeds 1%"
    exit 1
  }
  echo "OK: reconciliation drift within 1%"
fi

echo "FinOps live billing verification passed."
