#!/usr/bin/env bash
# Enforces minimum test coverage on core business-logic packages (Phase 20).
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND="${ROOT_DIR}/backend"
MIN_COVERAGE="${MIN_COVERAGE:-80}"
MIN_MIDDLEWARE="${MIN_MIDDLEWARE:-20}"
MIN_AUTH="${MIN_AUTH:-25}"
MIN_SECURITY="${MIN_SECURITY:-55}"

GATED_PACKAGES=(
  "./internal/ingestion/parser/..."
)

declare -A EXTENDED_MIN=(
  ["./internal/gateway/middleware/..."]="$MIN_MIDDLEWARE"
  ["./internal/gateway/auth/..."]="$MIN_AUTH"
  ["./internal/security/..."]="$MIN_SECURITY"
  ["./internal/notebook/..."]="25"
  ["./internal/mobile/..."]="25"
  ["./internal/observability/..."]="25"
)

REPORT_PACKAGES=(
  "./internal/correlation/graph/..."
  "./internal/correlation/deployment/..."
  "./internal/seed/..."
  "./internal/platform/mtls/..."
  "./internal/gateway/proxy/..."
)

cd "$BACKEND"
cover_file="$(mktemp)"
trap 'rm -f "$cover_file"' EXIT

ALL_PACKAGES=("${GATED_PACKAGES[@]}" "${!EXTENDED_MIN[@]}" "${REPORT_PACKAGES[@]}")
go test -coverprofile="$cover_file" -covermode=atomic "${ALL_PACKAGES[@]}" >/dev/null

total="$(go tool cover -func="$cover_file" | awk '/total:/ { gsub(/%/,"",$3); print $3 }')"
echo "Aggregate core coverage: ${total}%"

failed=0
for pkg in "${GATED_PACKAGES[@]}"; do
  pkg_cover="$(go test -coverprofile="$cover_file" -covermode=atomic "$pkg" 2>/dev/null | awk '/coverage:/ { gsub(/%/,""); print $5 }')"
  echo "  ${pkg}: ${pkg_cover}% (minimum ${MIN_COVERAGE}%)"
  awk -v total="$pkg_cover" -v min="$MIN_COVERAGE" 'BEGIN {
    if (total + 0 < min + 0) exit 1
  }' || failed=1
done

for pkg in "${!EXTENDED_MIN[@]}"; do
  min="${EXTENDED_MIN[$pkg]}"
  pkg_cover="$(go test -coverprofile="$cover_file" -covermode=atomic "$pkg" 2>/dev/null | awk '/coverage:/ { gsub(/%/,""); print $5 }')"
  echo "  ${pkg}: ${pkg_cover}% (minimum ${min}%)"
  awk -v total="$pkg_cover" -v min="$min" 'BEGIN {
    if (total + 0 < min + 0) exit 1
  }' || failed=1
done

if [[ "$failed" -ne 0 ]]; then
  echo "Coverage gate failed: one or more gated packages below threshold"
  exit 1
fi

echo "Coverage gate passed."
