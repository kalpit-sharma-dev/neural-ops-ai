#!/usr/bin/env bash
# Verify Kafka → materializer → Postgres alert score pipeline on a running stack.
set -euo pipefail

API_BASE="${API_BASE:-http://localhost:8080/api/v1}"
TENANT="${TENANT_ID:-default}"

policy_id="$(curl -fsS "${API_BASE}/alerts/policies" -H "X-Tenant-ID: ${TENANT}" | python -c "import json,sys; d=json.load(sys.stdin); print((d.get('data') or [{}])[0].get('id','ap-1'))" 2>/dev/null || echo ap-1)"

echo "Triggering alert policy ${policy_id}..."
curl -fsS -X POST "${API_BASE}/alerts/policies/${policy_id}/trigger" \
  -H "X-Tenant-ID: ${TENANT}" \
  -H "Content-Type: application/json" \
  -d '{"service":"payment-service","severity":"P1"}' >/dev/null

echo "Waiting for materialized score buckets (up to 60s)..."
for _ in $(seq 1 12); do
  scores="$(curl -fsS "${API_BASE}/alerts/policies/${policy_id}/scores" -H "X-Tenant-ID: ${TENANT}")"
  count="$(echo "${scores}" | python -c "import json,sys; print(len(json.load(sys.stdin).get('data') or []))" 2>/dev/null || echo 0)"
  if [ "${count}" -gt 0 ]; then
    echo "OK: ${count} alert score bucket(s) materialized for policy ${policy_id}"
    exit 0
  fi
  sleep 5
done

echo "FAIL: no score buckets after trigger — check KAFKA_BROKERS, materializer logs, and migration 000022" >&2
exit 1
