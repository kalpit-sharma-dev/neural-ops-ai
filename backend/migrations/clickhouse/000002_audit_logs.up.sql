-- Phase 18: ClickHouse audit log replication for analytics

CREATE TABLE IF NOT EXISTS audit_logs (
    tenant_id String,
    user_id String,
    action String,
    resource_type String,
    resource_id String,
    ip String,
    user_agent String,
    result String,
    created_at DateTime64(3, 'UTC')
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (tenant_id, created_at, action);
