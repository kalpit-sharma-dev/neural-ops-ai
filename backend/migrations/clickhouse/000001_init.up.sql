-- Phase 8: ClickHouse analytics schema

CREATE TABLE IF NOT EXISTS logs (
    tenant_id String DEFAULT 'default',
    timestamp DateTime64(3, 'UTC'),
    service String,
    environment String,
    severity String,
    message String,
    trace_id String,
    txn_id String,
    host String,
    pod String,
    classification String,
    explanation String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, service, severity, timestamp);

CREATE TABLE IF NOT EXISTS transactions (
    tenant_id String DEFAULT 'default',
    txn_id String,
    txn_type String,
    status String,
    hops String,
    total_latency_ms Int64,
    failed_at String,
    created_at DateTime64(3, 'UTC')
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (tenant_id, txn_id, created_at);

CREATE TABLE IF NOT EXISTS anomalies (
    tenant_id String DEFAULT 'default',
    timestamp DateTime64(3, 'UTC'),
    service String,
    metric_type String,
    score Float64,
    baseline Float64,
    actual Float64,
    anomaly_type String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, service, timestamp);

CREATE TABLE IF NOT EXISTS transaction_journeys (
    txn_id String,
    txn_type String,
    status String,
    failed_at String,
    total_latency_ms Int64,
    retry_count Int32,
    hops String,
    started_at DateTime64(3, 'UTC'),
    completed_at Nullable(DateTime64(3, 'UTC'))
) ENGINE = MergeTree()
ORDER BY (txn_id, started_at);
