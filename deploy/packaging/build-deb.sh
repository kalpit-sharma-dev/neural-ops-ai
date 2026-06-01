#!/usr/bin/env bash
# Build neuralops-oneagent .deb package (requires fpm: gem install fpm)
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
VERSION="${VERSION:-1.0.0}"
STAGE="${ROOT}/deploy/packaging/staging/deb"
rm -rf "$STAGE"
mkdir -p "$STAGE/opt/neuralops/oneagent" "$STAGE/etc/systemd/system" "$STAGE/DEBIAN"
cp -r "$ROOT/sdk/oneagent/"* "$STAGE/opt/neuralops/oneagent/"
cp "$ROOT/deploy/packaging/neuralops-oneagent.service" "$STAGE/etc/systemd/system/"
cat > "$STAGE/DEBIAN/control" <<EOF
Package: neuralops-oneagent
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: amd64
Maintainer: NeuralOps <ops@neuralops.ai>
Description: NeuralOps OneAgent auto-instrumentation
EOF
cat > "$STAGE/DEBIAN/postinst" <<'EOF'
#!/bin/sh
systemctl daemon-reload
systemctl enable neuralops-oneagent || true
EOF
chmod 755 "$STAGE/DEBIAN/postinst"
if command -v fpm >/dev/null 2>&1; then
  fpm -s dir -t deb -n neuralops-oneagent -v "$VERSION" -C "$STAGE" \
    --prefix / -p "$ROOT/deploy/packaging/dist/neuralops-oneagent_${VERSION}_amd64.deb" .
  echo "Built deploy/packaging/dist/neuralops-oneagent_${VERSION}_amd64.deb"
else
  dpkg-deb --build "$STAGE" "$ROOT/deploy/packaging/dist/neuralops-oneagent_${VERSION}_amd64.deb" 2>/dev/null || \
    echo "Install fpm or dpkg-deb to build .deb"
fi
