#!/usr/bin/env bash
# Gate E / Gate F verification — docs/bank/GATE_EF_CLOSEOUT_RUNBOOK.md
set -euo pipefail

GATE="${GATE:-http://localhost:8080}"
TOKEN="${TOKEN:-${NEURALOPS_API_TOKEN:-}}"
TENANT="${TENANT:-default}"
LEVEL="E"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --gate) LEVEL="${2^^}"; shift 2 ;;
    *) shift ;;
  esac
done

export GATEWAY="$GATE" TOKEN="$TOKEN" TENANT="$TENANT"
FAIL=0
ok() { echo "[gate-$LEVEL] OK: $*"; }
fail() { echo "[gate-$LEVEL] FAIL: $*"; FAIL=1; }

echo "=== Gate $LEVEL verification ==="

if [[ "$LEVEL" == "E" || "$LEVEL" == "F" ]]; then
  if [[ -x scripts/gate-d-verify.sh ]]; then
    echo "[gate-$LEVEL] Assuming Gate D complete (run gate-d-verify separately in prod)"
  fi
fi

auth=()
[[ -n "$TOKEN" ]] && auth=(-H "Authorization: Bearer $TOKEN")
hdr=(-H "X-Tenant-ID: $TENANT" "${auth[@]}")

if [[ "$LEVEL" == "E" ]]; then
  if command -v k6 >/dev/null 2>&1; then
    if k6 run --vus 5 --duration 30s scripts/k6/bank-poc-load.js 2>/dev/null; then
      ok "k6 bank profile (30s)"
    else
      fail "k6 bank profile"
    fi
  else
    echo "[gate-E] SKIP k6 not installed"
  fi

  for doc in \
    "docs/compliance/SOC2_TYPE1_POLICY_INDEX.md" \
    "docs/compliance/SOC2_TYPE2_READINESS.md" \
    "docs/compliance/COVERAGE_ROADMAP.md"; do
    [[ -f "$doc" ]] && ok "doc $(basename "$doc")" || fail "missing $doc"
  done
fi

if [[ "$LEVEL" == "F" ]]; then
  MODE="${FINOPS_BILLING_MODE:-}"
  if [[ "$MODE" == "live" || "$MODE" == "hybrid" ]]; then
    ok "FINOPS_BILLING_MODE=$MODE"
  else
    fail "FINOPS_BILLING_MODE must be live or hybrid for Gate F (got '$MODE')"
  fi

  if [[ "${FINOPS_REQUIRE_POSTGRES:-}" == "true" ]]; then
    ok "FINOPS_REQUIRE_POSTGRES"
  else
    fail "set FINOPS_REQUIRE_POSTGRES=true for Gate F"
  fi

  if [[ -x scripts/verify-finops-production.sh ]]; then
    ./scripts/verify-finops-production.sh && ok "finops production smoke" || fail "finops smoke"
  fi

  # Reconciliation drift check (JSON)
  if command -v jq >/dev/null 2>&1; then
    body=$(curl -fsS "${hdr[@]}" "$GATE/api/v1/finops/reconciliation" 2>/dev/null || echo "")
    if echo "$body" | jq -e '.status == "success"' >/dev/null 2>&1; then
      drift=$(echo "$body" | jq '[.data[]?.driftPct // 0] | max // 0')
      if awk -v d="$drift" 'BEGIN { exit !(d <= 1.0) }'; then
        ok "reconciliation max drift ${drift}%"
      else
        fail "reconciliation drift ${drift}% exceeds 1%"
      fi
    else
      fail "reconciliation API"
    fi
  else
    echo "[gate-F] SKIP drift parse (jq not installed)"
  fi
fi

if [[ "$FAIL" -ne 0 ]]; then
  exit 1
fi
echo "Gate $LEVEL automated checks passed."
