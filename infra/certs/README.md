# Dev mTLS certificates

Generate with:

```bash
bash scripts/gen-mtls-certs.sh
```

Mount into the gateway container (see `infra/docker-compose.yml`):

- `INTERNAL_MTLS_CERT_FILE=/etc/neuralops/mtls/gateway-client.crt`
- `INTERNAL_MTLS_KEY_FILE=/etc/neuralops/mtls/gateway-client.key`
- `INTERNAL_MTLS_CA_FILE=/etc/neuralops/mtls/ca.crt`

When `INTERNAL_MTLS_ENABLED=true`, the gateway presents the client certificate on upstream HTTP(S) calls.
For full mutual TLS, terminate TLS on an internal proxy or enable HTTPS on backend services.

**Do not commit private keys to production** — use Vault/KMS in real deployments.
