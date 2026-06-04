#!/usr/bin/env bash
# verify-gap-plan.sh — engineering acceptance runner for GAP_IMPLEMENTATION_PLAN (code items).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

echo "== Gap plan engineering verification =="

echo "-- backend build --"
(cd backend && go build ./...)

echo "-- observability + contract tests --"
(cd backend && go test ./internal/observability/... ./tests/contract/... -count=1)

echo "-- coverage gate --"
bash scripts/coverage-gate.sh

echo "-- frontend build --"
(cd frontend && npm run build)

echo "-- ops tracker docs --"
bash scripts/ops/verify-gap-ops-readiness.sh

echo "OK: engineering gap plan verification passed"
echo "Optional (requires running stack):"
echo "  bash scripts/run-ui-actions-smoke.sh"
echo "  bash scripts/verify-metrics-pilot-load.sh"
echo "  bash scripts/verify-bank-ingest-load.sh"
echo "  bash scripts/verify-ebpf-fleet.sh"
