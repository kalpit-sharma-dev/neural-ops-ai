# Air-Gapped & Restricted Network Deployment (Wave 2.7)

**ID:** BANK-015

## When to use

- No outbound internet from observability cluster
- Regulator-mandated data localization
- SWIFT / cardholder environments with deny-all egress

## Topology

```
Customer K8s (air-gap)
  ├── NeuralOps Helm release (gateway, ingestion, UI, workers)
  ├── Managed Postgres / Kafka / ES / CH (in-zone)
  ├── Private container registry (mirrored images)
  └── Optional: customer LLM endpoint (VPC-only) or AI disabled
```

## Image supply

1. Mirror `ghcr.io/neuralops/*` (or vendor registry) to internal registry.
2. Update `values-prod.yaml`:

```yaml
image:
  registry: registry.bank.internal/neuralops
  pullSecrets:
    - name: bank-registry-pull
```

3. `imagePullPolicy: IfNotPresent` — pre-load on nodes for DR.

## License & updates

- Offline license file mounted via Secret (see [AIR_GAP_INSTALL.md](./AIR_GAP_INSTALL.md) Wave 3.1).
- Patch cadence: quarterly bundle delivered on encrypted media; verify checksums in change window.

## Data plane

- No SaaS telemetry egress; disable external webhooks except bank-approved ITSM endpoints.
- LLM: [CUSTOMER_LLM_KEYS.md](./CUSTOMER_LLM_KEYS.md) — model endpoint must stay in trust zone.

## Ingest without cloud APIs

- Fluent Bit / OTEL only — [BANK_LOG_INGEST.md](../runbooks/BANK_LOG_INGEST.md)
- FinOps live billing **disabled** until CUR export files are sneaker-netted (see FIN-PROD-01 in roadmap).

## Security controls

| Control | Air-gap note |
|---------|----------------|
| SSO | Bank IdP reachable only inside corp network |
| Secrets | Vault / K8s Secrets — no External Secrets cloud backend |
| Updates | Signed bundles + internal change advisory |
| Support | Jump host or customer-shared log bundle — [SUPPORT_ESCALATION.md](./SUPPORT_ESCALATION.md) |

## Verification checklist

- [ ] `helm template` renders with internal registry + no `external-secrets` cloud provider
- [ ] `./scripts/verify-production-auth.sh` passes with `AUTH_ALLOW_DEV_LOGIN=false`
- [ ] Egress NetworkPolicy denies `0.0.0.0/0` except DNS + ITSM allowlist
- [ ] Backup/restore drill on in-zone Postgres — [BACKUP_RESTORE.md](../runbooks/BACKUP_RESTORE.md)
