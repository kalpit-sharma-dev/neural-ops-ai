-- Distributed trace spans for APM / PurePath queries

CREATE TABLE IF NOT EXISTS trace_spans (
    tenant_id String DEFAULT 'default',
    trace_id String,
    span_id String,
    parent_id String,
    service String,
    operation String,
    start_time DateTime64(3, 'UTC'),
    duration_ms Int64,
    status String,
    tags Map(String, String)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(start_time)
ORDER BY (tenant_id, trace_id, start_time);
