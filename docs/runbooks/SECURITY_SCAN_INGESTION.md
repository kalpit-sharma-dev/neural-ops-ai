# SCA & Container Scan Ingestion (Wave 4.6)

**ID:** SEC-01

## Objective

Surface **dependency** and **image** vulnerabilities inside NeuralOps security views and route critical findings to alerts.

## Supported sources (integrate via API / webhook)

| Scanner | Ingest method |
|---------|---------------|
| Trivy | CI JSON → `POST /api/v1/security/vulnerabilities` (batch) |
| Grype | Same schema |
| Snyk | Webhook adapter (customer middleware) |
| AWS Inspector / Azure Defender | Cloud inventory sync — partial in `cloudinventory` |

## Normalized finding schema

```json
{
  "source": "trivy",
  "assetType": "container_image",
  "assetId": "neuralops/gateway:1.2.3",
  "cveId": "CVE-2024-XXXX",
  "severity": "HIGH",
  "package": "openssl",
  "fixedVersion": "3.0.14",
  "detectedAt": "2026-06-01T12:00:00Z"
}
```

## CI pipeline (customer registry)

```bash
trivy image --format json -o trivy.json "$IMAGE"
curl -s -X POST "$GATEWAY/api/v1/security/vulnerabilities/batch" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @"$(jq '{findings: .Results}' trivy.json)"
```

> Batch endpoint: `POST /api/v1/security/vulnerabilities/batch` (see handler + OpenAPI).

## Alert policy

Create alert policy `service_pattern: '*'` route to security channel when `severity=critical` count &gt; 0 for `assetType=container_image`.

## Gate E

- [ ] Weekly scan on all NeuralOps images in customer registry
- [ ] Critical CVE SLA 7 days
- [ ] Dashboard panel imported (Grafana or in-app security correlation)

## Related

- `backend/internal/observability/security_correlation.go`
- `.github/workflows/ci.yml` — add Trivy job (optional `continue-on-error` until GA)
