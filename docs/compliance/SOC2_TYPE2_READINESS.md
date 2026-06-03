# SOC 2 Type II — Audit Period Readiness (Wave 4.1)

**ID:** BANK-012

## Difference from Type I

| Type I | Type II |
|--------|---------|
| Design at a point in time | Operating effectiveness over **6–12 months** |
| Policy existence | Evidence of consistent execution |

Start Type II only after [Gate D](../BANK_PRODUCTION_READINESS_ROADMAP.md) production stable ≥ 30 days.

## 12-month evidence calendar

| Month | Evidence to collect |
|-------|---------------------|
| 1–3 | Access reviews, change tickets, incident logs, backup drills |
| 4–6 | Pen test retest, vendor reviews, training records |
| 7–9 | Soak/post-mortem samples, capacity reviews |
| 10–12 | Population sampling for auditor (tickets, logs, configs) |

## Control owners

| Control | Owner | System |
|---------|-------|--------|
| Logical access | Customer IAM + NeuralOps RBAC export | SSO groups |
| Change mgmt | Engineering | GitHub PR + Helm releases |
| Monitoring | SRE | Prometheus/Grafana |
| Incident | On-call | PagerDuty + [INCIDENT_RCA_PLAYBOOK.md](../runbooks/INCIDENT_RCA_PLAYBOOK.md) |
| Vendor | Legal | [SUBPROCESSORS.md](../bank/SUBPROCESSORS.md) |

## NeuralOps deliverables to customer

- [ ] Annual penetration test summary (customer-owned report)
- [ ] Subprocessor change notification process
- [ ] Security patch SLA for self-hosted images
- [ ] Coordinated incident notification template

## Index

Policies: [SOC2_TYPE1_POLICY_INDEX.md](./SOC2_TYPE1_POLICY_INDEX.md)

**Disclaimer:** Type II requires a licensed CPA firm; this document is preparation only.
