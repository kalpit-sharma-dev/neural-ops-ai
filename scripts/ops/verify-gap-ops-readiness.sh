#!/usr/bin/env bash
# verify-gap-ops-readiness.sh — checks Wave 0 ops tracker artifacts exist (execution remains customer/SRE).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TRACKERS=(
  "docs/bank/DPA_MSA_SIGNOFF_TRACKER.md"
  "docs/bank/PENETRATION_TEST_CHECKLIST.md"
  "docs/compliance/SOC2_TYPE1_POLICY_INDEX.md"
  "docs/runbooks/BACKUP_RESTORE.md"
  "docs/runbooks/STAGING_SOAK_7DAY.md"
  "docs/bank/GATE_BC_CLOSEOUT_RUNBOOK.md"
  "docs/bank/FINOPS_PRODUCTION.md"
)
SCRIPTS=(
  "scripts/staging-soak-check.sh"
  "scripts/ops/verify-gap-ops-readiness.sh"
)

missing=0
for f in "${TRACKERS[@]}"; do
  if [ -f "${ROOT}/${f}" ]; then
    echo "OK  ${f}"
  else
    echo "MISSING ${f}" >&2
    missing=$((missing + 1))
  fi
done

if [ "${missing}" -gt 0 ]; then
  exit 1
fi

for f in "${SCRIPTS[@]}"; do
  if [ -f "${ROOT}/${f}" ]; then
    echo "OK  ${f}"
  else
    echo "MISSING ${f}" >&2
    missing=$((missing + 1))
  fi
done

if [ "${missing}" -gt 0 ]; then
  exit 1
fi
echo "Wave 0 ops documentation trackers present (sign-off is manual)"
