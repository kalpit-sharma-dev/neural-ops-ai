#!/usr/bin/env bash
# Gate D production verification — docs/bank/GATE_D_CLOSEOUT_RUNBOOK.md
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "=== Gate D verification (subset) ==="

# Gate C baseline first
if [[ -x scripts/gate-bc-verify.sh ]]; then
  GATEWAY="${GATEWAY:-http://localhost:8080}" \
  TOKEN="${TOKEN:-}" \
  TENANT="${TENANT:-default}" \
  ./scripts/gate-bc-verify.sh --gate C || exit 1
else
  echo "FAIL: gate-bc-verify.sh missing"
  exit 1
fi

FAIL=0
ok() { echo "[gate-D] OK: $*"; }
fail() { echo "[gate-D] FAIL: $*"; FAIL=1; }

# Production auth env (from shell or cluster — here we check exported prod-like vars)
ENVIRONMENT="${ENVIRONMENT:-production}" \
AUTH_DISABLED="${AUTH_DISABLED:-false}" \
AUTH_ALLOW_DEV_LOGIN="${AUTH_ALLOW_DEV_LOGIN:-false}" \
DEMO_MODE="${DEMO_MODE:-false}" \
OIDC_ENABLED="${OIDC_ENABLED:-true}" \
./scripts/verify-production-auth.sh && ok "production auth" || fail "production auth"

# FinOps prod settings when verifying customer prod
if [[ "${FINOPS_REQUIRE_POSTGRES:-}" == "true" ]]; then
  ok "FINOPS_REQUIRE_POSTGRES set"
fi
if [[ "${AUTOFIX_ENABLED:-false}" == "false" ]]; then
  ok "AUTOFIX disabled (bank default)"
else
  fail "AUTOFIX_ENABLED should be false in bank prod unless waived"
fi

# OpenAPI bank + prod paths
if [[ -f docs/openapi/gateway-v1.yaml ]]; then
  for marker in "/finops/reconciliation:" "/finops/audit/export:"; do
    grep -q "$marker" docs/openapi/gateway-v1.yaml && ok "openapi $marker" || fail "openapi $marker"
  done
fi

# Helm rollback dry-run when cluster access available
if command -v helm >/dev/null 2>&1 && command -v kubectl >/dev/null 2>&1; then
  if kubectl get ns "${HELM_NAMESPACE:-neuralops}" >/dev/null 2>&1; then
    ./scripts/helm-rollback-drill.sh --dry-run && ok "helm rollback dry-run" || fail "helm rollback dry-run"
  else
    echo "[gate-D] SKIP helm rollback (namespace not reachable)"
  fi
else
  echo "[gate-D] SKIP helm rollback (helm/kubectl not installed)"
fi

if [[ "$FAIL" -ne 0 ]]; then
  echo "Gate D automated checks failed. Complete manual items in GATE_D_CLOSEOUT_RUNBOOK.md"
  exit 1
fi
echo "Gate D automated checks passed. Complete legal, pen test, and sign-off tables manually."
