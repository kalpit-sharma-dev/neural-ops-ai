# Helm Professional Services Playbook (Wave 3.2)

**Duration:** 5–10 business days (self-hosted enterprise)

## Prerequisites

| Item | Owner |
|------|-------|
| K8s 1.28+ cluster, ≥ 3 worker nodes | Customer |
| Managed Postgres 15+, Kafka, ES, ClickHouse, Redis | Customer or NeuralOps PS |
| DNS + TLS cert (corp CA) | Customer |
| SSO metadata (OIDC/SAML) | Customer IAM |
| `values-prod.yaml` filled | Joint |

## Day-by-day

### Day 1 — Foundation

```bash
helm upgrade -i neuralops infra/helm/neuralops \
  -n neuralops --create-namespace \
  -f infra/helm/neuralops/values.yaml \
  -f infra/helm/neuralops/values-prod.yaml
```

- Verify `/health`, `/ready`
- External Secrets or in-cluster Secrets for DB/Kafka

### Day 2 — Data & migrations

- Run goose migrations against Postgres
- ClickHouse init job
- Seed **disabled** in prod

### Day 3 — Auth

- [SSO_SETUP_BANK.md](../runbooks/SSO_SETUP_BANK.md)
- `./scripts/verify-production-auth.sh`
- Map RBAC groups → Admin / SRE / Developer

### Day 4 — Ingest & integrations

- [BANK_LOG_INGEST.md](../runbooks/BANK_LOG_INGEST.md)
- ITSM OAuth smoke: `./scripts/verify-itsm-integration.sh`

### Day 5 — Observability of NeuralOps

- Import Grafana dashboards from `infra/grafana/`
- Prometheus scrape annotations on gateway

### Days 6–10 — Hardening (optional)

- NetworkPolicy / service mesh mTLS
- Backup restore drill
- k6 POC profile on staging

## Handoff package

- [ ] As-built diagram (namespaces, data stores)
- [ ] Runbook index ([bank/README.md](./README.md))
- [ ] On-call rotation + [SUPPORT_ESCALATION.md](./SUPPORT_ESCALATION.md)
- [ ] Rollback: `helm rollback neuralops <rev>`

## Rollback SLA

Target **&lt; 15 minutes** to previous Helm revision with compatible DB migration (forward-only migrations require restore from backup — document in change plan).
