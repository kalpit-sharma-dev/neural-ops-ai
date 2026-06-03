# Auto-Instrumentation Matrix (COLL-03)

**SRS:** REQ-COLL-004  
**API:** `GET /api/v1/collectors/autoinstrumentation`

NeuralOps supports **OpenTelemetry-based** auto-instrumentation for bank POC workloads. This is not full zero-code parity with proprietary APM agents.

## Supported today (GA via OTEL)

| Language | Method |
|----------|--------|
| Java | `-javaagent:opentelemetry-javaagent.jar` |
| Go | OTEL SDK + `OTEL_SERVICE_NAME` |
| Python | `opentelemetry-instrument` wrapper |
| Node.js | `@opentelemetry/auto-instrumentations-node` |

## Bank deployment

1. Deploy NEXAGENT or sidecar with OTEL exporter → gateway `/api/v1/traces`
2. Set `OTEL_EXPORTER_OTLP_ENDPOINT` to customer gateway
3. Use [PII_SCRUBBING_GUIDE.md](./PII_SCRUBBING_GUIDE.md) on collector pipeline

## Preview / planned

- .NET auto-instrumentation (preview)
- Ruby (planned — manual SDK today)

Query live matrix: `curl -H "Authorization: Bearer $TOKEN" $GATEWAY/api/v1/collectors/autoinstrumentation`
