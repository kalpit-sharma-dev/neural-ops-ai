# No Data Training Clause (BANK-016 supplement)

**Version:** 1.0  
**Owner:** Legal + Product

This clause supplements the [DPA_TEMPLATE.md](./DPA_TEMPLATE.md) for bank customers using AI features.

---

## Contract language (insert into DPA / SOW)

> **Customer data and AI.** Customer Content (logs, traces, metrics, incident data, and prompts sent to AI features) is processed solely to provide the Services. NeuralOps does **not** use Customer Content to train foundation models or shared machine-learning models. When Customer configures a **Customer-Managed LLM Endpoint** (see [CUSTOMER_LLM_KEYS.md](./CUSTOMER_LLM_KEYS.md)), inference requests are sent directly to Customer's provider; NeuralOps retains only operational metadata (latency, token counts, error rates) required for service health.

---

## Technical controls

| Control | Implementation |
|---------|----------------|
| Customer LLM keys | `LLM_API_KEY` / provider URL per tenant; no shared vendor key in prod |
| Prompt retention | Configurable; default 7 days for RCA audit |
| Air-gap AI off | `AI_ENABLED=false` in Helm values |
| AutoFix | Disabled by default in bank prod (`AUTOFIX_ENABLED=false`) |

---

## Related

- [CUSTOMER_LLM_KEYS.md](./CUSTOMER_LLM_KEYS.md)
- [AUTOFIX_BANK_POLICY.md](./AUTOFIX_BANK_POLICY.md)
