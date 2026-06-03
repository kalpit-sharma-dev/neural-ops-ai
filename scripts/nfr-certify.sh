#!/usr/bin/env bash
#
# nfr-certify.sh — run NFR (non-functional requirement) certification against a
# live stack and emit a signed JSON report.
#
# It exercises the platform's NFR endpoints, runs the live SLA + benchmark jobs,
# applies pass/fail gates (availability, p95 latency, error rate), and writes a
# report plus a SHA-256 checksum ("signature") so the artifact is tamper-evident
# in CI. Exits non-zero when any gate fails so the pipeline blocks a regressed
# release.
set -euo pipefail

API_URL="${NEURALOPS_API_URL:-http://localhost:8080/api/v1}"
TENANT="${NEURALOPS_TENANT:-default}"
OUT_DIR="${NFR_OUT_DIR:-./nfr-report}"
REPORT="${OUT_DIR}/nfr-certification.json"
SIGNATURE="${OUT_DIR}/nfr-certification.sha256"

# Gates (overridable via env).
MIN_AVAILABILITY="${NFR_MIN_AVAILABILITY:-99.9}"
MAX_P95_MS="${NFR_MAX_P95_MS:-300}"
MAX_ERROR_RATE="${NFR_MAX_ERROR_RATE:-1.0}"

mkdir -p "${OUT_DIR}"
auth=(-H "X-Tenant-ID: ${TENANT}")

echo "NFR certification against ${API_URL} (tenant=${TENANT})"

fetch() {
  curl -sf "${auth[@]}" "${API_URL}$1" 2>/dev/null || echo '{}'
}
post() {
  curl -sf "${auth[@]}" -X POST "${API_URL}$1" 2>/dev/null || echo '{}'
}

benchmarks="$(fetch /nfr/benchmarks)"
reliability="$(fetch /nfr/reliability)"
accessibility="$(fetch /nfr/accessibility)"
locales="$(fetch /nfr/i18n/locales)"
security_evidence="$(fetch /nfr/security-evidence)"
sla_run="$(post /nfr/sla/run)"
bench_run="$(post /nfr/benchmark/run)"

# Extract SLA metrics with jq when available, else fall back to grep/sed.
extract() {
  local json="$1" key="$2"
  if command -v jq >/dev/null 2>&1; then
    echo "$json" | jq -r "try (.data.$key) // (.$key) // empty"
  else
    echo "$json" | grep -o "\"$key\"[: ]*[0-9.]*" | head -1 | grep -o '[0-9.]*$'
  fi
}

availability="$(extract "$sla_run" availabilityPct)"; availability="${availability:-0}"
p95="$(extract "$sla_run" p95LatencyMs)"; p95="${p95:-9999}"
error_rate="$(extract "$sla_run" errorRatePct)"; error_rate="${error_rate:-100}"

pass=true
gate() {
  # gate <name> <actual> <op> <threshold>
  if awk "BEGIN{exit !($2 $3 $4)}"; then
    echo "  PASS  $1 ($2 $3 $4)"
  else
    echo "  FAIL  $1 ($2 not $3 $4)"
    pass=false
  fi
}

echo "Applying NFR gates:"
gate "availability"  "$availability" ">=" "$MIN_AVAILABILITY"
gate "p95_latency"   "$p95"          "<=" "$MAX_P95_MS"
gate "error_rate"    "$error_rate"   "<=" "$MAX_ERROR_RATE"

generated_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
cat > "${REPORT}" <<EOF
{
  "generatedAt": "${generated_at}",
  "apiUrl": "${API_URL}",
  "tenant": "${TENANT}",
  "gates": {
    "minAvailabilityPct": ${MIN_AVAILABILITY},
    "maxP95Ms": ${MAX_P95_MS},
    "maxErrorRatePct": ${MAX_ERROR_RATE}
  },
  "results": {
    "availabilityPct": ${availability},
    "p95LatencyMs": ${p95},
    "errorRatePct": ${error_rate}
  },
  "passed": ${pass},
  "evidence": {
    "benchmarks": ${benchmarks:-{}},
    "reliability": ${reliability:-{}},
    "accessibility": ${accessibility:-{}},
    "locales": ${locales:-{}},
    "securityEvidence": ${security_evidence:-{}},
    "slaRun": ${sla_run:-{}},
    "benchmarkRun": ${bench_run:-{}}
  }
}
EOF

# Sign the report so CI artifacts are tamper-evident.
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum "${REPORT}" | awk '{print $1}' > "${SIGNATURE}"
else
  shasum -a 256 "${REPORT}" | awk '{print $1}' > "${SIGNATURE}"
fi

echo "Report:    ${REPORT}"
echo "Signature: ${SIGNATURE} ($(cat "${SIGNATURE}"))"

if [ "${pass}" = "true" ]; then
  echo "NFR certification PASSED"
  exit 0
fi
echo "NFR certification FAILED — one or more gates not met" >&2
exit 1
