#!/usr/bin/env bash
# Verifies production SAML SP metadata is exposed when SAML is enabled.
set -euo pipefail

GATEWAY_BASE="${GATEWAY_BASE:-http://localhost:8080}"
METADATA_URL="${GATEWAY_BASE}/api/v1/auth/saml/metadata"

echo "Checking SAML SP metadata: ${METADATA_URL}"
body="$(curl -fsS "${METADATA_URL}")"
echo "${body}" | grep -q 'EntityDescriptor'
echo "${body}" | grep -q 'AssertionConsumerService'
echo "SAML metadata verification passed."
