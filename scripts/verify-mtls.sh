#!/usr/bin/env bash
# Verifies mTLS proxy rejects unauthenticated clients and accepts gateway client cert.
set -euo pipefail

PROXY_URL="${MTLS_PROXY_URL:-https://localhost:8443}"
CERT_DIR="${MTLS_CERT_DIR:-infra/certs}"

echo "Expecting mTLS proxy to reject requests without client certificate..."
if curl -kfsS "${PROXY_URL}/health" >/dev/null 2>&1; then
  echo "mTLS proxy accepted unauthenticated request (unexpected)"
  exit 1
fi
echo "Unauthenticated request rejected as expected."

echo "Expecting gateway client certificate to be accepted..."
curl -kfsS \
  --cert "${CERT_DIR}/gateway-client.crt" \
  --key "${CERT_DIR}/gateway-client.key" \
  --cacert "${CERT_DIR}/ca.crt" \
  "${PROXY_URL}/health" | grep -q ok
echo "mTLS client verification passed."
