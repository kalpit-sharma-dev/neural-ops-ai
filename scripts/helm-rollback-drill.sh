#!/usr/bin/env bash
# Helm rollback drill helper (Gate D) — docs/runbooks/HELM_ROLLBACK_DRILL.md
set -euo pipefail

RELEASE="${HELM_RELEASE:-neuralops}"
NS="${HELM_NAMESPACE:-neuralops}"
DRY_RUN=false
TARGET_REV=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=true; shift ;;
    --revision) TARGET_REV="${2:-}"; shift 2 ;;
    *) shift ;;
  esac
done

if ! command -v helm >/dev/null 2>&1; then
  echo "helm not found"
  exit 1
fi

echo "== Helm rollback drill (release=$RELEASE namespace=$NS) =="
helm history "$RELEASE" -n "$NS" 2>/dev/null | tail -6 || {
  echo "FAIL: cannot read helm history (is release installed?)"
  exit 1
}

CURRENT=$(helm history "$RELEASE" -n "$NS" -o json 2>/dev/null | grep -o '"revision":[0-9]*' | tail -1 | grep -o '[0-9]*' || echo "1")
if [[ -z "$TARGET_REV" ]]; then
  if [[ "$CURRENT" -gt 1 ]]; then
    TARGET_REV=$((CURRENT - 1))
  else
    TARGET_REV=1
  fi
fi

echo "Current revision: $CURRENT"
echo "Target rollback revision: $TARGET_REV"

if [[ "$DRY_RUN" == "true" ]]; then
  echo "DRY RUN: would execute: helm rollback $RELEASE $TARGET_REV -n $NS --wait"
  echo "Rollback drill dry-run OK"
  exit 0
fi

START=$(date +%s)
helm rollback "$RELEASE" "$TARGET_REV" -n "$NS" --wait --timeout 10m
END=$(date +%s)
ELAPSED=$((END - START))
echo "Rollback completed in ${ELAPSED}s"

if [[ "$ELAPSED" -gt 900 ]]; then
  echo "WARN: elapsed exceeds 15 minute Gate D target"
  exit 1
fi
echo "Rollback drill OK (< 15 min)"
