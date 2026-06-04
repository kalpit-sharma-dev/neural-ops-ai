#!/usr/bin/env bash
# verify-bank-ingest-load.sh — GAP-NFR-002: k6 bank ingest POC profile + signed evidence artifact.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASE_URL="${BASE_URL:-http://localhost:8080}"
TENANT="${K6_TENANT:-default}"
TOKEN="${K6_TOKEN:-${NEURALOPS_API_TOKEN:-}}"
DURATION="${K6_INGEST_DURATION:-2m}"
RATE="${K6_INGEST_RATE:-50}"
EVIDENCE_DIR="${EVIDENCE_DIR:-${ROOT}/docs/evidence}"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
ARTIFACT="${EVIDENCE_DIR}/bank-ingest-load-${STAMP}.json"
SUMMARY="${EVIDENCE_DIR}/bank-ingest-load-${STAMP}.sha256"

mkdir -p "${EVIDENCE_DIR}"

run_k6() {
  if command -v k6 >/dev/null 2>&1; then
    export BASE_URL K6_TENANT K6_TOKEN K6_INGEST_DURATION="${DURATION}" K6_INGEST_RATE="${RATE}"
    export K6_SUMMARY_PATH="${ARTIFACT}"
    k6 run "${ROOT}/scripts/k6/bank-ingest-load.js"
    return
  fi
  docker run --rm --network host \
    -v "${ROOT}/scripts/k6:/scripts" \
    -v "${EVIDENCE_DIR}:/evidence" \
    -e BASE_URL="${BASE_URL}" \
    -e K6_TENANT="${TENANT}" \
    -e K6_TOKEN="${TOKEN}" \
    -e K6_INGEST_DURATION="${DURATION}" \
    -e K6_INGEST_RATE="${RATE}" \
    -e K6_SUMMARY_PATH="/evidence/bank-ingest-load-${STAMP}.json" \
    grafana/k6 run /scripts/bank-ingest-load.js
}

echo "==> bank ingest load (${BASE_URL}, rate=${RATE}, duration=${DURATION})"
run_k6

if [ ! -f "${ARTIFACT}" ]; then
  echo "missing k6 summary artifact at ${ARTIFACT}" >&2
  exit 1
fi

sha256sum "${ARTIFACT}" | awk '{print $1}' > "${SUMMARY}"
echo "Evidence: ${ARTIFACT}"
echo "Checksum: ${SUMMARY}"

if command -v jq >/dev/null 2>&1; then
  passed="$(jq -r '.passed // false' "${ARTIFACT}")"
  if [ "${passed}" != "true" ]; then
    echo "ingest load gates failed — see ${ARTIFACT}" >&2
    exit 1
  fi
fi

echo "bank ingest load verification complete"
