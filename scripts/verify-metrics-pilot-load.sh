#!/usr/bin/env bash
# verify-metrics-pilot-load.sh — GAP-NFR-001 MET-01 metrics load at pilot volume.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASE_URL="${BASE_URL:-http://localhost:8080}"
TENANT="${K6_TENANT:-default}"
DURATION="${K6_METRICS_DURATION:-1m}"
RATE="${K6_METRICS_RATE:-30}"
EVIDENCE_DIR="${EVIDENCE_DIR:-${ROOT}/docs/evidence}"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
ARTIFACT="${EVIDENCE_DIR}/metrics-pilot-load-${STAMP}.json"

mkdir -p "${EVIDENCE_DIR}"
export BASE_URL K6_TENANT K6_METRICS_DURATION="${DURATION}" K6_METRICS_RATE="${RATE}" K6_SUMMARY_PATH="${ARTIFACT}"

if command -v k6 >/dev/null 2>&1; then
  k6 run "${ROOT}/scripts/k6/metrics-pilot-load.js"
else
  docker run --rm --network host \
    -v "${ROOT}/scripts/k6:/scripts" \
    -v "${EVIDENCE_DIR}:/evidence" \
    -e BASE_URL="${BASE_URL}" -e K6_TENANT="${TENANT}" \
    -e K6_METRICS_DURATION="${DURATION}" -e K6_METRICS_RATE="${RATE}" \
    -e K6_SUMMARY_PATH="/evidence/metrics-pilot-load-${STAMP}.json" \
    grafana/k6 run /scripts/metrics-pilot-load.js
fi

test -f "${ARTIFACT}"
sha256sum "${ARTIFACT}" | awk '{print $1}' > "${ARTIFACT%.json}.sha256"
echo "Metrics pilot evidence: ${ARTIFACT}"
