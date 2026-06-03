# Data Processing Agreement (Template)

**DRAFT — FOR LEGAL COUNSEL ONLY. NOT EXECUTABLE.**

NeuralOps / [Vendor Legal Entity]  
Customer: [Bank Legal Entity]  
Effective date: [DATE]

---

## 1. Definitions

- **Customer Data:** Log, metric, trace, incident, and configuration data submitted by Customer.
- **Personal Data:** Any Customer Data identifying natural persons, as defined by applicable law.
- **Services:** NeuralOps observability platform as described in the Order Form.

---

## 2. Roles

- Customer is **Data Controller** (or Processor to its end users, as applicable).
- Vendor is **Data Processor** processing Personal Data on documented instructions.

---

## 3. Processing instructions

Vendor shall process Personal Data only to:

1. Provide the Services per the Agreement and Order Form.
2. Follow Customer's documented configuration (retention, residency, access controls).
3. Comply with applicable law.

---

## 4. Security measures

Vendor implements measures described in [SECURITY_PACK.md](./SECURITY_PACK.md), including:

- Encryption in transit (TLS 1.2+)
- Encryption at rest (customer-managed keys for self-hosted)
- Access controls (SSO, RBAC, ABAC)
- Audit logging
- Tenant isolation

---

## 5. Subprocessors

Vendor may engage subprocessors listed in [SUBPROCESSORS.md](./SUBPROCESSORS.md). Vendor provides 30-day notice of material additions. Customer may object on reasonable grounds.

---

## 6. International transfers

Transfers outside [EEA / specified country] require [SCCs / BCR / other mechanism].

---

## 7. Data subject rights

Vendor assists Customer in responding to data subject requests within [30] days, to extent technically feasible.

---

## 8. Breach notification

Vendor notifies Customer without undue delay (target **72 hours**) after confirming a Personal Data breach affecting Customer Data.

---

## 9. Deletion and return

Upon termination, Vendor deletes or returns Customer Data per [DATA_FLOW.md](./DATA_FLOW.md) §7 within [30] days, except legal retention.

---

## 10. Audit rights

Customer may audit Vendor's compliance [annually / upon reasonable notice], or accept SOC 2 report when available.

---

## 11. AI / LLM processing

When AI features are enabled:

- Customer may supply its own LLM API keys.
- Vendor does not use Customer Data to train foundation models.
- Prompt content may be sent to LLM provider per Customer configuration.

---

## 12. Term

Co-terminous with Master Agreement.

---

**Signatures**

| Vendor | Customer |
|--------|----------|
| | |

---

*Replace bracketed fields. Do not send to banks without legal approval.*
