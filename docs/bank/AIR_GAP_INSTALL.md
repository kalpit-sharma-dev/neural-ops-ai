# Offline License & Air-Gap Install (Wave 3.1)

**ID:** BANK-018

## License file

1. Sales issues **offline license** JSON (signed by NeuralOps licensing service).
2. Create Secret before Helm install:

```bash
kubectl create secret generic neuralops-license \
  --from-file=license.json=./license-offline.json \
  -n neuralops
```

3. Helm values:

```yaml
license:
  enabled: true
  secretName: neuralops-license
  key: license.json
```

> Gateway validates license when `NEURALOPS_LICENSE_PATH` and `NEURALOPS_LICENSE_SECRET` are set. Helm: `license.enabled=true`.

## Install sequence (no internet)

| Day | Activity |
|-----|----------|
| 0 | Registry mirror + image vulnerability scan |
| 1 | Data plane (Postgres, Kafka, ES, CH) + backups |
| 2 | `helm install` core + ingress + SSO |
| 3 | Ingest agents (Fluent Bit / OTEL) |
| 4 | SSO + RBAC + ABAC policies |
| 5 | POC load test + DR drill |

## Upgrade bundle

1. Receive `neuralops-<version>-airgap.tar.gz` + SHA256 manifest.
2. Load images: `ctr images import` or `docker load`.
3. `helm upgrade` with unchanged `values-prod.yaml` + migration job.

## AI in air-gap

- Set `AI_ENABLED=false` or point to on-prem inference only.
- Document model allowlist in security pack appendix.

## Related

- [AIR_GAPPED_DEPLOYMENT.md](./AIR_GAPPED_DEPLOYMENT.md)
- [HELM_PS_PLAYBOOK.md](./HELM_PS_PLAYBOOK.md)
