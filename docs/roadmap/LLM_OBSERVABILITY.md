# LLM Workload Observability (Wave 6.4)

**ID:** AI-03

## Preview APIs (available)

| Endpoint | Description |
|----------|-------------|
| `GET /api/v1/ai/llm/workloads` | Per-model latency, tokens, cost |
| `GET /api/v1/ai/llm/usage?window=24h` | Aggregate tokens and budget burn |

## Roadmap

| Milestone | Target | Deliverable |
|-----------|--------|-------------|
| M1 | Q3 2026 | OTEL spans for LLM calls (`gen_ai.*` attributes) |
| M2 | Q4 2026 | Per-tenant cost caps + alerts |
| M3 | Q1 2027 | Prompt/version registry + quality scoring |

## Bank requirements

- Customer-owned API keys — [CUSTOMER_LLM_KEYS.md](../bank/CUSTOMER_LLM_KEYS.md)
- No prompt content in logs (PII scrubbing)
- AutoFix **off** for LLM-driven remediation by default

## Instrumentation (planned)

```go
// OTEL: gen_ai.system, gen_ai.request.model, gen_ai.usage.input_tokens
```

Interim: use gateway AI chat metrics + manual workload rows in UI.
