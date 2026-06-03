# Backup & Restore Runbook (BANK-009 / BANK-010)

**Version:** 1.0  
**Owner:** Platform SRE  
**Review:** Quarterly drill required for Gate B

---

## 1. Scope

| System | Backup method | RPO target | RTO target |
|--------|---------------|----------|------------|
| PostgreSQL | Automated PITR (RDS/cloud) | 5 min | 1 h |
| Elasticsearch | Daily snapshot → object storage | 24 h | 4 h |
| ClickHouse | Native backup / S3 | 24 h | 4 h |
| Kafka | RF ≥ 3, multi-AZ | N/A (replay) | 1 h |
| Helm values / Git | GitOps repo | 0 | 30 min |

---

## 2. PostgreSQL restore (RDS example)

1. Identify restore point (incident time − 5 min).
2. Restore to new RDS instance from snapshot/PITR.
3. Update `POSTGRES_DSN` in External Secret.
4. Run gateway migrations: `go run ./cmd/migrate` (or init job).
5. Verify `/ready` on gateway and incident list API.
6. Cut over connection string; decommission old instance after 48h.

**Drill record:** Date, operator, actual RTO, issues.

---

## 3. Elasticsearch restore

1. Restore snapshot to new cluster or index prefix.
2. Re-point `ELASTICSEARCH_URL` in search service.
3. Run smoke: `GET /api/v1/search?q=health&tenant=default`.

---

## 4. ClickHouse restore

1. Restore from backup per vendor docs.
2. Verify audit replication and transaction analytics queries.

---

## 5. Kafka

- Increase consumer lag alert threshold during broker recovery.
- Replay from retention if ingestion duplicated — use idempotent ingest keys.

---

## 6. Quarterly drill checklist

- [ ] Restore Postgres to isolated instance — **tested**
- [ ] Restore one ES snapshot index — **tested**
- [ ] Document actual RTO vs targets
- [ ] Update [BANK_PRODUCTION_READINESS_ROADMAP.md](../BANK_PRODUCTION_READINESS_ROADMAP.md) Gate B

---

## 7. Bank customer (self-hosted)

Customer operates backups. NeuralOps PS provides this runbook as **recommended procedure** during onboarding (Wave 3.2).

---

*Last drill: [DATE] [PASS/FAIL]*
