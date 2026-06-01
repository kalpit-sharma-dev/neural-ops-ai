-- Phase 8: Error pattern catalog

CREATE TABLE IF NOT EXISTS error_patterns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    pattern_hash VARCHAR(128) NOT NULL,
    category VARCHAR(64) NOT NULL,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    sample_message TEXT NOT NULL,
    first_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    occurrence_count BIGINT NOT NULL DEFAULT 1,
    llm_explanation JSONB NOT NULL DEFAULT '{}'::jsonb,
    UNIQUE (tenant_id, pattern_hash)
);

CREATE INDEX IF NOT EXISTS idx_error_patterns_tenant_last_seen ON error_patterns(tenant_id, last_seen);
CREATE INDEX IF NOT EXISTS idx_error_patterns_category ON error_patterns(category);
