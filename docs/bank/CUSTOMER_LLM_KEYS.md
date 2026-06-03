# Customer-Managed LLM Keys (BANK-016)

**Version:** 1.0  
**Date:** 2026-06-03

---

## 1. Why banks should use their own keys

- LLM provider appears as **customer's** subprocessor, not NeuralOps's.
- Data processing agreement stays between bank and Azure OpenAI / private endpoint.
- Keys can be revoked instantly by the bank.

---

## 2. Configuration (self-hosted / dedicated)

Set on **analysis** and **gateway** services (never commit to git):

| Variable | Description |
|----------|-------------|
| `LLM_PROVIDER` | `openai`, `azure`, `anthropic`, `ollama` |
| `OPENAI_API_KEY` | Customer key |
| `OPENAI_BASE_URL` | Azure OpenAI endpoint if applicable |
| `OPENAI_MODEL` | Approved model name (e.g. gpt-4o) |

Use Kubernetes External Secrets or Vault injection.

---

## 3. Network

- Prefer **private endpoint** (Azure OpenAI VNet, AWS PrivateLink) from analysis pods.
- Block outbound internet from analysis namespace except LLM endpoint.

---

## 4. Data sent to LLM

Typical RCA prompt includes:

- Incident title, service name, error excerpts (truncated)
- Correlated log lines (tenant-scoped, size-limited)

**Not sent by default:** full raw log archives, secrets, auth headers.

Configure prompt templates and max tokens per tenant (quota middleware).

---

## 5. Contract language

DPA §11: Vendor does not train models on Customer Data. Customer controls LLM provider via own key.

---

## 6. Verification

- [ ] Staging uses bank test key only
- [ ] No vendor-owned OpenAI key in prod `values-prod.yaml`
- [ ] LLM calls visible in bank's cloud billing dashboard

---

*See also [SECURITY_PACK.md](./SECURITY_PACK.md) §7.*
