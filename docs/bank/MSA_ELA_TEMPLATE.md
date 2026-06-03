# Master Services Agreement / Enterprise License Agreement (Template)

**DRAFT — FOR LEGAL COUNSEL ONLY. NOT EXECUTABLE.**

**Product:** NeuralOps Enterprise (Self-Hosted)  
**Customer:** [Bank Legal Entity]  
**Effective date:** [DATE]

---

## 1. License grant

Vendor grants Customer a non-exclusive, non-transferable license to deploy NeuralOps software on Customer infrastructure for internal business purposes during the Term.

**Deployment:** Kubernetes (Helm chart) in Customer data centers or private cloud.

---

## 2. Order forms

Specific modules, log volume tiers, support level, and PS days defined in Order Form(s):

| SKU element | Example |
|-------------|---------|
| License tier | Up to [X] GB logs/day |
| Support | 8×5 or 24×7 |
| Professional services | [N] onboarding days |
| Modules | Observability core; FinOps optional |

---

## 3. Customer responsibilities

- Operate secure Kubernetes and data stores per [SECURITY_PACK.md](./SECURITY_PACK.md)
- Configure SSO (OIDC/SAML)
- Scrub PCI/PII from logs at source
- Maintain backups per [BACKUP_RESTORE.md](../runbooks/BACKUP_RESTORE.md)
- Provide access for PS during agreed windows

---

## 4. Support and SLA

- **Standard:** 8×5 email, P2 response [4] business hours
- **Premium:** 24×7, P1 response [1] hour (optional)

SLA credits apply only to Dedicated/Cloud SKUs with signed SLA appendix. Self-hosted: best-effort + support contract.

---

## 5. Fees

Annual license fee: $[AMOUNT] USD, invoiced [annually].  
Professional services: $[RATE]/day.  
Renewal: [auto-renew / negotiate].

---

## 6. Confidentiality

Mutual NDA terms [standard].

---

## 7. Warranty disclaimer

Software provided **as-is** except express warranties in SLA appendix. Vendor does not warrant uninterrupted operation of Customer infrastructure.

---

## 8. Limitation of liability

Cap: fees paid in prior 12 months. Exclusion of consequential damages [standard carve-outs for gross negligence].

---

## 9. Security incidents

Per DPA breach notification. Customer responsible for incident comms to regulators for data in Customer-controlled environment.

---

## 10. Term and termination

- Initial term: [1–3] years
- Termination for convenience: [90] days notice
- Upon termination: license ends; Customer deletes software; data per DPA

---

## 11. Governing law

[Laws of jurisdiction]

---

**Exhibits**

- A: Order Form  
- B: DPA ([DPA_TEMPLATE.md](./DPA_TEMPLATE.md))  
- C: Security Pack ([SECURITY_PACK.md](./SECURITY_PACK.md))  
- D: POC Scope ([POC_SCOPE_TEMPLATE.md](./POC_SCOPE_TEMPLATE.md)) — if applicable  

---

*Legal review required before any customer signature.*
