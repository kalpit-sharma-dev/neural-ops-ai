# SOC 2 Type I — Policy Index (Wave 3.3)

**ID:** BANK-011  
**Status:** Policy templates for customer diligence — **not** a certification claim.

## Trust Service Criteria mapping

| TSC | Policy document (customer-owned after edit) | NeuralOps artifact |
|-----|---------------------------------------------|-------------------|
| CC1 Control environment | Information Security Policy | [SECURITY_PACK.md](../bank/SECURITY_PACK.md) |
| CC2 Communication | Acceptable Use Policy | HR / customer |
| CC3 Risk assessment | Risk Register template | Export from GRC tool |
| CC4 Monitoring | Logging & monitoring standard | Grafana/Prometheus runbooks |
| CC5 Control activities | Change management procedure | CI/CD + Helm upgrade |
| CC6 Logical access | Access control policy | [PRODUCTION_AUTH_CHECKLIST.md](../bank/PRODUCTION_AUTH_CHECKLIST.md), RBAC/ABAC |
| CC7 System operations | Incident response plan | [INCIDENT_RCA_PLAYBOOK.md](../runbooks/INCIDENT_RCA_PLAYBOOK.md) |
| CC8 Change management | SDLC policy | `.github/workflows/ci.yml` |
| CC9 Risk mitigation | Vendor management | [SUBPROCESSORS.md](../bank/SUBPROCESSORS.md) |
| A1 Availability | DR/BCP | [BACKUP_RESTORE.md](../runbooks/BACKUP_RESTORE.md) |
| C1 Confidentiality | Encryption standard | SECURITY_PACK § encryption |
| PI1 Processing integrity | Data validation | Ingest schema + PII guide |

## Evidence collection (Type I readiness)

- [ ] Organization chart for platform team
- [ ] Last 90 days access reviews (SSO groups)
- [ ] Pen test report (Wave 1.6)
- [ ] Backup restore ticket
- [ ] Sample change ticket with CI link

## Auditor engagement

Type I: point-in-time design effectiveness.  
Type II (Wave 4): operating effectiveness over 6–12 months — start only after Gate D production stable.

## Disclaimer

Possession of this index does **not** imply SOC 2 certification. Update `[x]` checkboxes only after customer counsel and auditor agreement.
