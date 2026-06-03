#!/usr/bin/env bash
# Unified gate verifier: B | C | D | E | F
# Usage: ./scripts/gate-verify.sh --gate D
set -euo pipefail

GATE_LEVEL="C"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --gate) GATE_LEVEL="${2^^}"; shift 2 ;;
    *) shift ;;
  esac
done

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

case "$GATE_LEVEL" in
  B|C)
    exec ./scripts/gate-bc-verify.sh --gate "$GATE_LEVEL"
    ;;
  D)
    exec ./scripts/gate-d-verify.sh
    ;;
  E|F)
    exec ./scripts/gate-ef-verify.sh --gate "$GATE_LEVEL"
    ;;
  *)
    echo "Unknown gate: $GATE_LEVEL (use B, C, D, E, or F)"
    exit 1
    ;;
esac
