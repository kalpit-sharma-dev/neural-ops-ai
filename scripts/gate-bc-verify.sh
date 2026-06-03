#!/usr/bin/env bash
# Gate B / Gate C automated checks (docs/bank/GATE_BC_CLOSEOUT_RUNBOOK.md).
set -euo pipefail

GATE="${GATE:-http://localhost:8080}"
TOKEN="${TOKEN:-${NEURALOPS_API_TOKEN:-}}"
TENANT="${TENANT:-default}"
GATE_LEVEL="B"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --gate) GATE_LEVEL="${2^^}"; shift 2 ;;
    *) shift ;;
  esac
done

export GATEWAY="$GATE" TOKEN="$TOKEN" TENANT="$TENANT"
FAIL=0
ok() { echo "[gate-$GATE_LEVEL] OK: $*"; }
fail() { echo "[gate-$GATE_LEVEL] FAIL: $*"; FAIL=1; }

echo "=== Gate $GATE_LEVEL verification ==="
echo "Gateway: $GATE Tenant: $TENANT"

if curl -fsS "$GATE/health" >/dev/null 2>&1; then
  ok "/health"
else
  fail "/health"
fi

if [[ -x scripts/verify-production-auth.sh ]]; then
  ENVIRONMENT=staging AUTH_ALLOW_DEV_LOGIN=false OIDC_ENABLED=true \
    ./scripts/verify-production-auth.sh && ok "production-auth" || fail "production-auth"
else
  fail "verify-production-auth.sh missing"
fi

if [[ -x scripts/verify-sso-bank.sh ]]; then
  ENVIRONMENT=staging OIDC_ENABLED=true \
    ./scripts/verify-sso-bank.sh && ok "sso-bank" || fail "sso-bank"
fi

if [[ -x scripts/verify-itsm-integration.sh ]]; then
  ./scripts/verify-itsm-integration.sh && ok "itsm" || fail "itsm"
fi

if [[ -x scripts/verify-finops-production.sh ]]; then
  ./scripts/verify-finops-production.sh && ok "finops" || fail "finops"
fi

auth=()
[[ -n "$TOKEN" ]] && auth=(-H "Authorization: Bearer $TOKEN")
hdr=(-H "X-Tenant-ID: $TENANT" "${auth[@]}")

for path in \
  "/api/v1/finops/reconciliation" \
  "/api/v1/ai/llm/workloads" \
  "/api/v1/ai/llm/usage" \
  "/api/v1/admin/msp/tenants" \
  "/api/v1/logs/tiering" \
  "/api/v1/infra/serverless/functions" \
  "/api/v1/collectors/autoinstrumentation" \
  "/api/v1/collectors/discovery" \
  "/api/v1/collectors/supported-platforms" \
  "/api/v1/query/saved" \
  "/api/v1/logs/ingest-formats" \
  "/api/v1/apm/service-catalog" \
  "/api/v1/admin/signal-policies" \
  "/api/v1/security/threat-feed"; do
  code=$(curl -s -o /dev/null -w "%{http_code}" "${hdr[@]}" "$GATE$path" || echo "000")
  if [[ "$code" == "200" ]]; then
    ok "GET $path"
  else
    fail "GET $path HTTP $code"
  fi
done

code=$(curl -s -o /dev/null -w "%{http_code}" "${hdr[@]}" "$GATE/api/v1/finops/audit/export?limit=5" || echo "000")
if [[ "$code" == "200" ]]; then
  ok "GET /finops/audit/export"
else
  fail "GET /finops/audit/export HTTP $code"
fi

if [[ "$GATE_LEVEL" == "C" ]]; then
  if command -v k6 >/dev/null 2>&1; then
    k6 run --vus 2 --duration 10s scripts/k6/bank-poc-load.js && ok "k6 smoke" || fail "k6 smoke"
  else
    echo "[gate-C] SKIP k6 (not installed)"
  fi
fi

if [[ -f docs/openapi/gateway-v1.yaml ]]; then
  for spec_path in \
    "/finops/reconciliation:" \
    "/finops/audit/export:" \
    "/ai/llm/workloads:" \
    "/ai/llm/usage:" \
    "/query/explain:" \
    "/query/saved:" \
    "/collectors/autoinstrumentation:" \
    "/logs/tiering:" \
    "/infra/serverless/functions:"; do
    if grep -q "$spec_path" docs/openapi/gateway-v1.yaml; then
      ok "openapi $spec_path"
    else
      fail "openapi missing $spec_path"
    fi
  done
fi

if [[ "$FAIL" -ne 0 ]]; then
  echo "Gate $GATE_LEVEL verification failed."
  exit 1
fi
echo "Gate $GATE_LEVEL verification passed (automated subset; complete manual checklist in GATE_BC_CLOSEOUT_RUNBOOK.md)."
