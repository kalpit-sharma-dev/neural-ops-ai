#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
VERSION="${VERSION:-1.0.0}"
bash "$ROOT/deploy/packaging/build-deb.sh"
if command -v fpm >/dev/null 2>&1; then
  fpm -s deb -t rpm -p "$ROOT/deploy/packaging/dist/neuralops-oneagent-${VERSION}.x86_64.rpm" \
    "$ROOT/deploy/packaging/dist/neuralops-oneagent_${VERSION}_amd64.deb"
  echo "Built deploy/packaging/dist/neuralops-oneagent-${VERSION}.x86_64.rpm"
fi
