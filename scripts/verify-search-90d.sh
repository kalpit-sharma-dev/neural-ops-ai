#!/usr/bin/env bash
# 90-day log search performance gate (LOG-02 / NFR-02).
set -euo pipefail

GATEWAY="${GATEWAY:-http://localhost:8080}"
TOKEN="${TOKEN:-${NEURALOPS_API_TOKEN:-}}"
TENANT="${TENANT:-default}"
RETENTION="${RETENTION_DAYS:-90}"

auth=()
[[ -n "$TOKEN" ]] && auth=(-H "Authorization: Bearer $TOKEN")
hdr=(-H "X-Tenant-ID: $TENANT" "${auth[@]}")

echo "== 90d search perf gate (retention=${RETENTION}d) =="

code=$(curl -s -o /dev/null -w "%{http_code}" "${hdr[@]}" \
  "$GATEWAY/api/v1/nfr/search-perf?retentionDays=$RETENTION" || echo "000")
if [[ "$code" != "200" ]]; then
  echo "FAIL: GET /nfr/search-perf HTTP $code"
  exit 1
fi
echo "OK: GET /nfr/search-perf"

if command -v k6 >/dev/null 2>&1; then
  BASE_URL="$GATEWAY" K6_TENANT="$TENANT" K6_TOKEN="$TOKEN" RETENTION_DAYS="$RETENTION" \
    k6 run scripts/k6/bank-search-90d.js
  echo "OK: k6 bank-search-90d"
else
  echo "SKIP: k6 not installed — run: k6 run scripts/k6/bank-search-90d.js"
fi

echo "90d search perf gate completed."
