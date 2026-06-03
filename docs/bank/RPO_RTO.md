# RPO / RTO Commitments (BANK-010)

**Version:** 1.0  
**Owner:** Platform Engineering + SRE  
**Sign-off:** Engineering leadership (Gate B)

---

## 1. Service objectives

| Tier | Systems | RPO | RTO | Notes |
|------|---------|-----|-----|-------|
| **Tier 0** | PostgreSQL (metadata, incidents, FinOps) | **5 min** | **1 h** | PITR via managed Postgres |
| **Tier 1** | Kafka (ingest pipeline) | N/A (replay) | **1 h** | RF ≥ 3, multi-AZ |
| **Tier 1** | Elasticsearch (log search) | **24 h** | **4 h** | Daily snapshots |
| **Tier 1** | ClickHouse (analytics, transactions) | **24 h** | **4 h** | Native backup to object storage |
| **Tier 2** | Redis (rate limit / quota) | **15 min** | **30 min** | Recreate from config |
| **Tier 2** | Helm values / GitOps | **0** | **30 min** | Source of truth in Git |

NeuralOps **gateway and stateless services** recover in **≤ 15 minutes** via Helm rollback ([HELM_ROLLBACK_DRILL.md](../runbooks/HELM_ROLLBACK_DRILL.md)).

---

## 2. Measurement method

- **RPO:** Maximum data loss window between last durable backup/replication point and failure time.
- **RTO:** Time from incident declaration to restored service passing `/ready` and smoke tests in [GATE_BC_CLOSEOUT_RUNBOOK.md](./GATE_BC_CLOSEOUT_RUNBOOK.md).

Quarterly drills recorded in [BACKUP_RESTORE.md](../runbooks/BACKUP_RESTORE.md).

---

## 3. Customer-specific overrides

Bank contracts may tighten Tier 0 to **RPO 1 min / RTO 30 min** with:

- Cross-region Postgres read replica + automated failover
- Hot Elasticsearch/ClickHouse standby cluster
- Documented in SOW appendix; not default self-hosted SKU

---

## 4. Engineering sign-off

| Role | Name | Date | Signature |
|------|------|------|-----------|
| VP Engineering | | | |
| Head of SRE | | | |
| CISO delegate | | | |

---

## Related

- [BACKUP_RESTORE.md](../runbooks/BACKUP_RESTORE.md)
- [SECURITY_PACK.md](./SECURITY_PACK.md)
