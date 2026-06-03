# FFIEC IT Examination Handbook — Mapping Worksheet (Wave 3.4)

**ID:** BANK-017 (optional)  
**Use:** Customer compliance workshop — not legal advice.

## How to use

1. Copy worksheet to customer GRC tool.
2. Map NeuralOps control **evidence** to FFIEC domain.
3. Mark **Customer** vs **NeuralOps** vs **Shared** responsibility.

## Sample mappings

| FFIEC domain | Examination topic | NeuralOps capability | Evidence | R/A |
|--------------|-------------------|----------------------|----------|-----|
| Development & Acquisition | Change management | CI, Helm upgrades | CI logs, change tickets | Shared |
| Cybersecurity | Access control | SSO, RBAC, ABAC | Auth checklist, ABAC tests | Shared |
| Cybersecurity | Monitoring | Logs, metrics, traces, alerts | Dashboards, alert policies | Shared |
| Operations | Business continuity | Backup/restore | BACKUP_RESTORE runbook | Customer infra |
| Operations | Incident response | Incidents, workflows, RCA | INCIDENT_RCA_PLAYBOOK | Shared |
| Outsourcing | Vendor management | Subprocessors | SUBPROCESSORS.md | NeuralOps |
| Audit | Logging & retention | Retention policies API | ADM-02 residency docs | Shared |

## Gaps to disclose honestly

| Topic | Status | Mitigation |
|-------|--------|------------|
| Live log tail | Roadmap — [LIVE_LOG_TAIL.md](../roadmap/LIVE_LOG_TAIL.md) | Batch search + export |
| AutoFix remediation | Partial | Human approval required |
| Shared SaaS multi-tenant | Not recommended for Tier 1 | Self-hosted SKU |

## Sign-off

| Role | Name | Date |
|------|------|------|
| Bank CISO delegate | | |
| NeuralOps SE | | |
