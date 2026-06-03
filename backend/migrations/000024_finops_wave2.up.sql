-- FinOps Wave 2: imports, commitments, carbon, reports, audit, lifecycle

ALTER TABLE finops_recommendations
    ADD COLUMN IF NOT EXISTS realized_savings_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS ticket_id VARCHAR(128) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS ticket_url VARCHAR(512) NOT NULL DEFAULT '';

ALTER TABLE finops_ingest_snapshots
    ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;

ALTER TABLE finops_anomalies
    ADD COLUMN IF NOT EXISTS probable_cause TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS sensitivity_factor DOUBLE PRECISION NOT NULL DEFAULT 1.0;

CREATE TABLE IF NOT EXISTS finops_import_batches (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    source VARCHAR(64) NOT NULL,
    format VARCHAR(16) NOT NULL DEFAULT 'json',
    line_count BIGINT NOT NULL DEFAULT 0,
    total_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 1,
    dedupe_key VARCHAR(256) NOT NULL DEFAULT '',
    imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, dedupe_key, version)
);
CREATE INDEX IF NOT EXISTS idx_finops_imports_tenant ON finops_import_batches (tenant_id, imported_at DESC);

CREATE TABLE IF NOT EXISTS finops_shared_splits (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    name VARCHAR(255) NOT NULL,
    resource_pattern VARCHAR(256) NOT NULL,
    mode VARCHAR(32) NOT NULL DEFAULT 'proportional',
    targets JSONB NOT NULL DEFAULT '[]',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS finops_tag_suggestions (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    resource_id VARCHAR(256) NOT NULL,
    suggested_key VARCHAR(128) NOT NULL,
    suggested_value VARCHAR(256) NOT NULL,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    spend_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_finops_tag_suggest_tenant ON finops_tag_suggestions (tenant_id, confidence DESC);

CREATE TABLE IF NOT EXISTS finops_anomaly_sensitivity (
    tenant_id VARCHAR(128) NOT NULL,
    scope VARCHAR(128) NOT NULL,
    z_threshold DOUBLE PRECISION NOT NULL DEFAULT 2.5,
    false_positive_count INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, scope)
);

CREATE TABLE IF NOT EXISTS finops_commitments (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    provider VARCHAR(32) NOT NULL,
    commitment_type VARCHAR(64) NOT NULL,
    region VARCHAR(64) NOT NULL DEFAULT '',
    coverage_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    utilization_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    monthly_commit_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_finops_commitments_tenant ON finops_commitments (tenant_id, expires_at);

CREATE TABLE IF NOT EXISTS finops_carbon_factors (
    id VARCHAR(64) PRIMARY KEY,
    provider VARCHAR(32) NOT NULL,
    region VARCHAR(64) NOT NULL,
    factor_version VARCHAR(32) NOT NULL DEFAULT '2026.1',
    kg_co2e_per_kwh DOUBLE PRECISION NOT NULL,
    renewable_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    methodology VARCHAR(128) NOT NULL DEFAULT 'GHG Protocol Scope 2',
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_finops_carbon_factors_region ON finops_carbon_factors (provider, region, factor_version);

CREATE TABLE IF NOT EXISTS finops_carbon_samples (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    resource_id VARCHAR(256) NOT NULL DEFAULT '',
    service VARCHAR(128) NOT NULL DEFAULT '',
    provider VARCHAR(32) NOT NULL,
    region VARCHAR(64) NOT NULL,
    co2e_kg DOUBLE PRECISION NOT NULL DEFAULT 0,
    kwh_estimate DOUBLE PRECISION NOT NULL DEFAULT 0,
    sampled_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_finops_carbon_samples_tenant ON finops_carbon_samples (tenant_id, sampled_at DESC);

CREATE TABLE IF NOT EXISTS finops_report_schedules (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    name VARCHAR(255) NOT NULL,
    scope VARCHAR(128) NOT NULL DEFAULT 'all',
    format VARCHAR(16) NOT NULL DEFAULT 'csv',
    cadence VARCHAR(32) NOT NULL DEFAULT 'weekly',
    delivery_channel VARCHAR(32) NOT NULL DEFAULT 'email',
    delivery_target VARCHAR(512) NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS finops_audit_log (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    actor VARCHAR(256) NOT NULL DEFAULT 'system',
    action VARCHAR(64) NOT NULL,
    entity_type VARCHAR(64) NOT NULL,
    entity_id VARCHAR(128) NOT NULL DEFAULT '',
    detail JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_finops_audit_tenant ON finops_audit_log (tenant_id, created_at DESC);
