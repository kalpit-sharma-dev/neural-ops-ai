#!/usr/bin/env bash
# Daily staging soak health check (Wave 1.9).
set -euo pipefail

GATEWAY="${GATEWAY:-http://localhost:8080}"
PROMETHEUS="${PROMETHEUS:-}"
DAY=""
FAIL=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --day) DAY="${2:-?}"; shift 2 ;;
    *) shift ;;
  esac
done

log() { echo "[soak] $*"; }
fail() { log "FAIL: $*"; FAIL=1; }
ok() { log "OK: $*"; }

log "=== Staging soak check (day ${DAY:-n/a}) $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

if curl -fsS "${GATEWAY}/health" | grep -qE 'ok|healthy|success'; then
  ok "GET /health"
else
  fail "GET /health"
fi

if curl -fsS "${GATEWAY}/ready" >/dev/null 2>&1; then
  ok "GET /ready"
else
  fail "GET /ready (optional endpoint)"
fi

	if curl -fsS "${GATEWAY}/api/v1/search/logs?q=service:gateway&limit=1" \
  -H "X-Tenant-ID: default" 2>/dev/null | grep -q '"status":"success"'; then
  ok "log search API (/search/logs)"
else
  fail "log search API (auth may be required on staging)"
fi

if [[ -n "$PROMETHEUS" ]]; then
  up="$(curl -fsS "${PROMETHEUS}/api/v1/query" --data-urlencode 'query=up{job="gateway"}' \
    | grep -o '"value":\[[^]]*\]' | tail -1 || true)"
  if echo "$up" | grep -q '"1"'; then
    ok "Prometheus gateway up"
  else
    fail "Prometheus gateway up query"
  fi
fi

if [[ "$FAIL" -ne 0 ]]; then
  log "Soak check completed with failures"
  exit 1
fi
log "Soak check passed"
exit 0
