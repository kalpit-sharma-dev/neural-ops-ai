# Bank POC — Scope & Success Criteria Template

**Version:** 1.0  
**Customer:** [Bank name]  
**Dates:** [Start] – [End] (recommended 4 weeks)  
**NeuralOps SKU:** POC on staging / dedicated tenant (not production until Gate D)

---

## 1. Participants

| Role | Name | Organization |
|------|------|--------------|
| Customer technical lead | | Bank |
| Customer security | | Bank |
| NeuralOps engineering | | Vendor |
| NeuralOps account | | Vendor |

---

## 2. In scope

| # | Capability | Notes |
|---|------------|-------|
| 1 | Log ingest (JSON/OTEL) for [N] services | e.g. payment-api, upi-service, ledger |
| 2 | Full-text + structured log search | 30-day retention in POC |
| 3 | Distributed tracing for same services | OTEL propagation |
| 4 | Service map / dependency view | |
| 5 | Incidents + alert rules (error rate, latency) | |
| 6 | AI RCA with **bank-provided LLM key** | No vendor OpenAI key |
| 7 | SSO (OIDC or SAML) test IdP | Okta / Azure AD test app |
| 8 | RBAC roles for 5–10 pilot users | |
| 9 | Transaction / UPI journey view | Banking differentiator |
| 10 | Audit log export sample | |

---

## 3. Out of scope (phase 2)

- Full Datadog/Splunk replacement
- FinOps with live invoice reconciliation (unless explicitly added)
- Multi-region active-active
- 24×7 vendor SLA
- SOC 2 Type II attestation (share roadmap only)
- AutoFix without human approval
- Production cutover

---

## 4. Customer prerequisites (Week 0)

- [ ] IdP metadata (SAML) or OIDC client credentials
- [ ] Sample log formats (3 services)
- [ ] Egress allowlist for NeuralOps endpoints (if applicable)
- [ ] Named security contact for questionnaire
- [ ] Agreed POC success metrics (Section 6)

---

## 5. Weekly plan

| Week | Customer | NeuralOps |
|------|----------|-----------|
| 1 | Provide logs; SSO test users | Connect ingest; validate parsing |
| 2 | Connect staging SSO | Tune alerts; dashboards |
| 3 | Run parallel with existing tool | AI RCA on 2 sample incidents |
| 4 | Measure MTTR / alert noise | Success report + commercial proposal |

---

## 6. Success criteria (agree upfront)

**POC passes if ANY TWO of:**

| Metric | Baseline | Target |
|--------|----------|--------|
| Duplicate alerts | [Current]/week | ≥ 20% reduction |
| Triage time (sample incidents) | [X] min | ≥ 30% faster |
| SSO + ingest | N/A | ≥ [Y] GB/day ingested with &lt; 2 min search lag |
| Transaction traceability | N/A | End-to-end UPI path visible for sample txn |

---

## 7. Security

- [ ] [SECURITY_PACK.md](./SECURITY_PACK.md) reviewed
- [ ] DPA negotiation started if POC passes
- [ ] No production customer PII in vendor demo tenant without approval

---

## 8. Exit outcomes

| Outcome | Next step |
|---------|-----------|
| **Pass** | Enterprise license proposal + Gate D planning |
| **Partial** | Extended POC or narrowed scope |
| **Fail** | Document gaps; no production commitment |

---

**Signatures (non-binding POC charter)**

Customer: _________________ Date: _______  
NeuralOps: _________________ Date: _______
