# Customer Incident Communications Template (BANK-020)

**Version:** 1.0  
**Owner:** Customer Success / Incident Commander

Use for **regulated bank customers** during Sev-1/Sev-2 platform incidents affecting observability or incident workflows.

---

## 1. Initial notification (within 30 min of Sev-1)

**Subject:** `[NeuralOps] Incident IN-{id} — {short title}`

```
Status: Investigating
Impact: {e.g. Log search latency elevated; alerting delayed}
Scope: {tenant/region/services}
Start time (UTC): {timestamp}
Customer action required: {None | SSO workaround | Pause ingest}

We are investigating elevated {symptom}. Your data remains encrypted at rest and in transit.
Updates every 30 minutes until mitigated.

Incident Commander: {name}
Support bridge: {dial-in or Slack channel}
```

---

## 2. Progress update

**Subject:** `[NeuralOps] Incident IN-{id} — Update {n}`

```
Status: {Investigating | Identified | Monitoring | Resolved}
Current findings: {root cause hypothesis or confirmed cause — no customer PII}
Mitigation: {what we changed}
ETA: {if known}
Next update: {time UTC}
```

---

## 3. Resolution

**Subject:** `[NeuralOps] Incident IN-{id} — Resolved`

```
Status: Resolved
Duration: {start} – {end} UTC
Root cause summary: {1–3 sentences — customer-safe}
Customer impact: {what they could/could not do}
Follow-up: Post-incident review within 5 business days; corrective actions shared under NDA
```

---

## 4. Status page mapping

| Internal severity | External status |
|-------------------|-----------------|
| Sev-1 | Major outage |
| Sev-2 | Degraded performance |
| Sev-3 | Minor / monitoring |

Coordinate with [SUPPORT_ESCALATION.md](./SUPPORT_ESCALATION.md) for 24×7 routing.

---

## Related

- [INCIDENT_RCA_PLAYBOOK.md](../runbooks/INCIDENT_RCA_PLAYBOOK.md)
- [GATE_D_CLOSEOUT_RUNBOOK.md](./GATE_D_CLOSEOUT_RUNBOOK.md)
