#!/usr/bin/env bash
# Verifies mTLS proxy rejects unauthenticated clients and accepts gateway client cert.
set -euo pipefail

PROXY_URL="${MTLS_PROXY_URL:-https://localhost:8443}"
CERT_DIR="${MTLS_CERT_DIR:-infra/certs}"
PROXY_HOST="${MTLS_PROXY_HOST:-localhost}"
PROXY_PORT="${MTLS_PROXY_PORT:-8443}"

verify_with_curl() {
  echo "Expecting mTLS proxy to reject requests without client certificate..."
  if curl -kfsS "${PROXY_URL}/health" >/dev/null 2>&1; then
    echo "mTLS proxy accepted unauthenticated request (unexpected)"
    return 1
  fi
  echo "Unauthenticated request rejected as expected."

  echo "Expecting gateway client certificate to be accepted..."
  curl -kfsS \
    --cert "${CERT_DIR}/gateway-client.crt" \
    --key "${CERT_DIR}/gateway-client.key" \
    --cacert "${CERT_DIR}/ca.crt" \
    "${PROXY_URL}/health" | grep -q ok
  echo "mTLS client verification passed."
}

verify_with_openssl() {
  if ! command -v openssl >/dev/null 2>&1; then
    echo "openssl is required for mTLS verification on this platform, but was not found."
    return 1
  fi

  echo "Expecting gateway client certificate to be accepted..."
  printf 'GET /health HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n' "${PROXY_HOST}" \
    | openssl s_client \
      -connect "${PROXY_HOST}:${PROXY_PORT}" \
      -servername "${PROXY_HOST}" \
      -cert "${CERT_DIR}/gateway-client.crt" \
      -key "${CERT_DIR}/gateway-client.key" \
      -CAfile "${CERT_DIR}/ca.crt" 2>/dev/null \
    | tr -d '\r' \
    | grep -q "^ok$"
  echo "mTLS client verification passed."
}

echo "Expecting mTLS proxy to reject requests without client certificate..."
if curl -kfsS "${PROXY_URL}/health" >/dev/null 2>&1; then
  echo "mTLS proxy accepted unauthenticated request (unexpected)"
  exit 1
fi
echo "Unauthenticated request rejected as expected."

# On Windows Git Bash, curl usually uses Schannel and may fail to import PEM cert/key.
# Use OpenSSL only for the client-cert leg; keep the unauthenticated check on curl.
if curl -V 2>/dev/null | grep -qi schannel; then
  verify_with_openssl
else
  echo "Expecting gateway client certificate to be accepted..."
  curl -kfsS \
    --cert "${CERT_DIR}/gateway-client.crt" \
    --key "${CERT_DIR}/gateway-client.key" \
    --cacert "${CERT_DIR}/ca.crt" \
    "${PROXY_URL}/health" | grep -q ok
  echo "mTLS client verification passed."
fi
