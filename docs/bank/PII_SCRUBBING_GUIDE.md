# PII & Sensitive Data Scrubbing (BANK-008 / REQ-COLL-015)

**Version:** 1.0  
**Date:** 2026-06-03

---

## 1. Principle

NeuralOps stores **what the customer sends**. Banks must scrub regulated data **before or at** collection. NeuralOps provides hooks; **the bank remains controller** for log content.

**Never ingest in clear text:**

- PAN, CVV, full card numbers
- Government IDs, full account numbers (mask to last 4)
- Passwords, OTPs, API secrets, private keys
- Unredacted session tokens

---

## 2. Recommended architecture

```text
Application logs → Fluent Bit / OTEL Collector
    │  regex/mask processors (PAN, email, phone)
    ▼
Optional NEXAGENT collector pipeline (drop/mask stage)
    ▼
NeuralOps ingestion API (TLS + API key)
```

---

## 3. Fluent Bit example (mask PAN-like sequences)

```ini
[FILTER]
    Name          rewrite_tag
    Match         kube.*
    Rule          $log ^.*$ scrubbed false

[FILTER]
    Name          lua
    Match         scrubbed
    script        mask_pan.lua
    call          mask_pan
```

`mask_pan.lua` (illustrative): replace `\b\d{13,19}\b` with `[REDACTED_PAN]`.

---

## 4. OpenTelemetry Collector processor

```yaml
processors:
  attributes/redact:
    actions:
      - key: http.request.header.authorization
        action: delete
      - key: user.password
        action: delete
```

---

## 5. NEXAGENT pipeline (roadmap)

Visual pipeline stages: `mask` → `drop` for fields matching bank policy IDs. Configure per fleet policy when pipeline UI is enabled (REQ-COLL-017).

---

## 6. Verification

- [ ] Security team signs off regex list for payment services
- [ ] Sample logs in staging scanned for PAN patterns (automated DLP scan)
- [ ] Pen test includes “accidental PAN in log” detection

---

## 7. FinOps note

Billing line items may include `accountId` — masked for non-admin roles in FinOps API (`MaskAccountID`).

---

*Owner: Platform + Customer security champion.*
