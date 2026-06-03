# Bank Log & Telemetry Ingest (Wave 2.2)

**ID:** LOG-01  
**Audience:** Customer platform team + NeuralOps PS

## Supported paths

| Signal | Recommended agent | Landing |
|--------|-------------------|---------|
| App JSON logs | Fluent Bit → Kafka | `ingestion` topic |
| Syslog / mainframe bridge | Fluent Bit `syslog` input | Kafka |
| Traces + metrics | OpenTelemetry Collector | OTLP → gateway/collector |
| Legacy agents | Vector or Fluent Bit forward | Kafka |

## Architecture

```
Apps / K8s / VMs
    → Fluent Bit (DaemonSet) or OTEL Collector
    → Kafka (customer-managed or NeuralOps cluster)
    → ingestion-service → Elasticsearch / ClickHouse
    → gateway APIs (/api/v1/logs, /search, traces)
```

## Fluent Bit (Kubernetes) — minimal ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: neuralops-fluent-bit
data:
  fluent-bit.conf: |
    [SERVICE]
        Flush        1
        Log_Level    info
        Parsers_File parsers.conf

    [INPUT]
        Name              tail
        Path              /var/log/containers/*payments*.log
        Parser            docker
        Tag               kube.payments.*

    [FILTER]
        Name          record_modifier
        Match         kube.payments.*
        Record        tenant_id ${TENANT_ID}
        Record        service payments

    [OUTPUT]
        Name              kafka
        Match             *
        Brokers           ${KAFKA_BROKERS}
        Topics            neuralops.logs
        rdkafka.compression.type gzip
```

**Bank notes:**

- Scrub PAN/account numbers in Fluent Bit `lua` or `grep` exclude — see [PII_SCRUBBING_GUIDE.md](../bank/PII_SCRUBBING_GUIDE.md).
- Set `tenant_id` and `region` on every record for residency enforcement.

## OpenTelemetry Collector — traces + metrics

```yaml
receivers:
  otlp:
    protocols:
      grpc:
      http:
processors:
  batch:
  resource:
    attributes:
      - key: tenant.id
        value: ${TENANT_ID}
        action: upsert
exporters:
  otlp:
    endpoint: ${NEURALOPS_COLLECTOR_ENDPOINT}:4317
    tls:
      insecure: false
      cert_file: /etc/neuralops/client.crt
      key_file: /etc/neuralops/client.key
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch, resource]
      exporters: [otlp]
    metrics:
      receivers: [otlp]
      processors: [batch, resource]
      exporters: [otlp]
```

## Kafka topic sizing (POC)

| Volume | Partitions | Retention |
|--------|------------|-----------|
| 50 GB/day logs | 12+ | 24–72 h |
| 5M spans/day | 8+ | 24 h |

Validate with [POC_LOAD_TEST.md](../bank/POC_LOAD_TEST.md).

## Verification

```bash
# Health
curl -s "$GATEWAY/health"

# Sample log search (after ingest)
curl -s "$GATEWAY/api/v1/logs/search?q=service:ledger-service&limit=5" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT"
```

## Exit criteria (Gate C)

- [ ] ≥ 90% expected log volume ingested for 24 h without lag alert
- [ ] Trace IDs visible on payment path services
- [ ] PII scrubbing rules applied and spot-checked
