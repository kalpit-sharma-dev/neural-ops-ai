#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_SPEC="$ROOT_DIR/docs/openapi/gateway-v1.yaml"
TARGET_SPEC="$ROOT_DIR/backend/openapi/swagger.yaml"

if [[ ! -f "$SOURCE_SPEC" ]]; then
  echo "Source OpenAPI spec not found: $SOURCE_SPEC" >&2
  exit 1
fi

mkdir -p "$(dirname "$TARGET_SPEC")"
cp "$SOURCE_SPEC" "$TARGET_SPEC"

echo "Synced OpenAPI spec:"
echo "  source: $SOURCE_SPEC"
echo "  target: $TARGET_SPEC"
