-- Phase 8: ClickHouse metrics time-series (aligned with ingestion ClickHouseWriter)

CREATE TABLE IF NOT EXISTS metrics (
    tenant_id String DEFAULT 'default',
    service_name String,
    host String,
    pod String,
    metric_type String,
    value Float64,
    timestamp DateTime64(3, 'UTC'),
    labels Map(String, String)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, service_name, metric_type, timestamp);
