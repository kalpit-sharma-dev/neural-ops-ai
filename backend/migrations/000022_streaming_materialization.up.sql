-- Streaming materialization for derived metrics and alert fatigue scoring

CREATE TABLE IF NOT EXISTS derived_metric_samples (
    tenant_id VARCHAR(128) NOT NULL,
    metric_id VARCHAR(64) NOT NULL,
    service VARCHAR(255) NOT NULL DEFAULT '',
    bucket_ts TIMESTAMPTZ NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    sample_count BIGINT NOT NULL DEFAULT 1,
    PRIMARY KEY (tenant_id, metric_id, service, bucket_ts)
);
CREATE INDEX IF NOT EXISTS idx_derived_metric_samples_ts ON derived_metric_samples (tenant_id, bucket_ts DESC);

CREATE TABLE IF NOT EXISTS alert_score_buckets (
    tenant_id VARCHAR(128) NOT NULL,
    policy_id VARCHAR(64) NOT NULL,
    service VARCHAR(255) NOT NULL,
    bucket_ts TIMESTAMPTZ NOT NULL,
    signal_count BIGINT NOT NULL DEFAULT 0,
    fatigue_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, policy_id, service, bucket_ts)
);
CREATE INDEX IF NOT EXISTS idx_alert_score_buckets_ts ON alert_score_buckets (tenant_id, bucket_ts DESC);
