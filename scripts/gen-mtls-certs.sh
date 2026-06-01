#!/usr/bin/env bash
# Generates a dev PKI for internal mTLS (CA + gateway client cert + service server cert).
set -euo pipefail

# Git Bash on Windows rewrites /CN=... OpenSSL subjects unless disabled.
if [[ -n "${MSYSTEM:-}" || "${OSTYPE:-}" == msys* ]]; then
  export MSYS_NO_PATHCONV=1
  export MSYS2_ARG_CONV_EXCL="*"
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CERT_DIR="${ROOT_DIR}/infra/certs"
DAYS=825
CN_CA="NeuralOps Dev CA"

mkdir -p "$CERT_DIR"

required_files=(
  ca.crt
  ca.key
  gateway-client.crt
  gateway-client.key
  internal-server.crt
  internal-server.key
)
have_all=true
for file in "${required_files[@]}"; do
  if [[ ! -f "${CERT_DIR}/${file}" ]]; then
    have_all=false
    break
  fi
done

if [[ "$have_all" == "true" && "${FORCE:-0}" != "1" ]]; then
  echo "Certificates already exist in ${CERT_DIR} (set FORCE=1 to regenerate)"
  exit 0
fi

if [[ "$have_all" != "true" ]]; then
  echo "Incomplete mTLS PKI in ${CERT_DIR}; regenerating..."
  rm -f "${CERT_DIR}/ca.key" "${CERT_DIR}/ca.crt" "${CERT_DIR}/gateway-client."* "${CERT_DIR}/internal-server."*
fi

# Run openssl from CERT_DIR so paths with spaces work on Windows/Git Bash.
(
  cd "$CERT_DIR"

  openssl genrsa -out ca.key 4096
  openssl req -x509 -new -nodes -key ca.key -sha256 -days "$DAYS" \
    -subj "/CN=${CN_CA}" -out ca.crt

  openssl genrsa -out gateway-client.key 2048
  openssl req -new -key gateway-client.key \
    -subj "/CN=neuralops-gateway-client" -out gateway-client.csr
  openssl x509 -req -in gateway-client.csr -CA ca.crt -CAkey ca.key \
    -CAcreateserial -out gateway-client.crt -days "$DAYS" -sha256

  openssl genrsa -out internal-server.key 2048
  openssl req -new -key internal-server.key \
    -subj "/CN=neuralops-internal" -out internal-server.csr
  printf "subjectAltName=DNS:ingestion,DNS:ingestion-mtls,DNS:search,DNS:incident,DNS:analysis,DNS:localhost\n" > server-ext.cnf
  openssl x509 -req -in internal-server.csr -CA ca.crt -CAkey ca.key \
    -CAcreateserial -out internal-server.crt -days "$DAYS" -sha256 \
    -extfile server-ext.cnf

  rm -f ./*.csr ./ca.srl ./server-ext.cnf
  chmod 600 ./*.key
)

echo "Generated mTLS PKI in ${CERT_DIR}"
echo "  CA:              ca.crt"
echo "  Gateway client:  gateway-client.crt + gateway-client.key"
echo "  Internal server: internal-server.crt + internal-server.key"
