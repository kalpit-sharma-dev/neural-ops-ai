# Penetration Test Checklist (Wave 1.6)

**ID:** BANK-005

## Scheduling

| Milestone | Owner | Target |
|-----------|-------|--------|
| Vendor selected | Security / Procurement | T-8 weeks before Gate D |
| Staging scope agreed | CISO + NeuralOps SE | T-6 weeks |
| Test window | Customer + vendor | 5 business days |
| Remediation | Engineering | Critical: 7 days, High: 30 days |

**Scope:** Staging URL + self-hosted Helm chart version pinned; include SSO, API gateway, ingest path, admin APIs. Exclude production customer data — use synthetic seed only.

## Pre-test readiness

- [ ] [PRODUCTION_AUTH_CHECKLIST.md](./PRODUCTION_AUTH_CHECKLIST.md) — dev login off
- [ ] SSO enforced on staging
- [ ] Rate limits / WAF on ingress (customer)
- [ ] Test accounts: Admin, Developer (scoped), Viewer
- [ ] Emergency contact list — [SUPPORT_ESCALATION.md](./SUPPORT_ESCALATION.md)

## OWASP-oriented test areas

| Area | NeuralOps surface |
|------|-------------------|
| Injection | Search, FinOps filters, NexQL (if enabled) |
| Broken auth | OIDC/SAML, API keys, session fixation |
| SSRF | Integration webhooks, export destinations |
| IDOR | Tenant header `X-Tenant-ID`, transaction IDs |
| Misconfig | K8s RBAC, NetworkPolicy, public buckets |
| Sensitive data | PII in logs — [PII_SCRUBBING_GUIDE.md](./PII_SCRUBBING_GUIDE.md) |

## Automated baseline (before vendor)

```bash
# Secret scan (CI)
# gitleaks in .github/workflows

# Container scan (customer registry)
trivy image registry.example/neuralops-gateway:<tag>

# mTLS smoke
bash scripts/verify-mtls.sh
```

## Finding severity & remediation

| Sev | SLA | Gate D |
|-----|-----|--------|
| Critical | 7 days | Block production |
| High | 30 days | Risk acceptance signed by CISO |
| Medium/Low | Backlog | Track in ticket system |

## Deliverables

- [ ] Executive summary (customer-owned)
- [ ] Technical report with reproduction steps
- [ ] Retest letter after fixes
- [ ] Update [SECURITY_PACK.md](./SECURITY_PACK.md) appendix with test date (no findings detail in public docs)

## Status tracking

Mark complete in [BANK_PRODUCTION_READINESS_ROADMAP.md](../BANK_PRODUCTION_READINESS_ROADMAP.md) Wave 1.6 when retest passes.
