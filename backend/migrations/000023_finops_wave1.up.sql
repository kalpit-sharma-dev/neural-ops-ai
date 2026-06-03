-- FinOps Wave 1: ingestion, allocation, anomalies, recommendations, budgets

CREATE TABLE IF NOT EXISTS finops_cost_line_items (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    billing_period DATE NOT NULL,
    provider VARCHAR(32) NOT NULL,
    account_id VARCHAR(128) NOT NULL DEFAULT '',
    service VARCHAR(128) NOT NULL DEFAULT '',
    region VARCHAR(64) NOT NULL DEFAULT '',
    resource_id VARCHAR(256) NOT NULL DEFAULT '',
    usage_type VARCHAR(128) NOT NULL DEFAULT '',
    quantity DOUBLE PRECISION NOT NULL DEFAULT 0,
    unit VARCHAR(32) NOT NULL DEFAULT '',
    amortized_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    list_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    effective_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    cost_view VARCHAR(16) NOT NULL DEFAULT 'amortized',
    tags JSONB NOT NULL DEFAULT '{}',
    team VARCHAR(128) NOT NULL DEFAULT '',
    environment VARCHAR(64) NOT NULL DEFAULT '',
    cost_center VARCHAR(128) NOT NULL DEFAULT '',
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_finops_line_items_tenant_period ON finops_cost_line_items (tenant_id, billing_period DESC);
CREATE INDEX IF NOT EXISTS idx_finops_line_items_scope ON finops_cost_line_items (tenant_id, team, service);

CREATE TABLE IF NOT EXISTS finops_allocation_rules (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    name VARCHAR(255) NOT NULL,
    dimension VARCHAR(64) NOT NULL,
    tag_key VARCHAR(128) NOT NULL,
    tag_value VARCHAR(256) NOT NULL DEFAULT '',
    priority INT NOT NULL DEFAULT 100,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_finops_alloc_rules_tenant ON finops_allocation_rules (tenant_id, priority);

CREATE TABLE IF NOT EXISTS finops_budgets (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    name VARCHAR(255) NOT NULL,
    scope_type VARCHAR(64) NOT NULL,
    scope_value VARCHAR(255) NOT NULL DEFAULT '',
    period VARCHAR(16) NOT NULL DEFAULT 'monthly',
    amount_usd DOUBLE PRECISION NOT NULL,
    thresholds JSONB NOT NULL DEFAULT '[50,80,100]',
    notify_policy_id VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_finops_budgets_tenant ON finops_budgets (tenant_id);

CREATE TABLE IF NOT EXISTS finops_anomalies (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    scope VARCHAR(128) NOT NULL,
    service VARCHAR(128) NOT NULL DEFAULT '',
    provider VARCHAR(32) NOT NULL DEFAULT '',
    delta_pct DOUBLE PRECISION NOT NULL,
    amount_usd DOUBLE PRECISION NOT NULL,
    severity VARCHAR(16) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    description TEXT NOT NULL DEFAULT '',
    alert_policy_id VARCHAR(64) NOT NULL DEFAULT '',
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    feedback VARCHAR(32) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_finops_anomalies_tenant ON finops_anomalies (tenant_id, detected_at DESC);

CREATE TABLE IF NOT EXISTS finops_recommendations (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    rec_type VARCHAR(64) NOT NULL,
    resource_id VARCHAR(256) NOT NULL,
    scope VARCHAR(128) NOT NULL DEFAULT '',
    title VARCHAR(512) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    projected_savings_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    risk_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_finops_recs_tenant ON finops_recommendations (tenant_id, status);

CREATE TABLE IF NOT EXISTS finops_ingest_snapshots (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    provider VARCHAR(32) NOT NULL,
    billing_period DATE NOT NULL,
    line_count BIGINT NOT NULL DEFAULT 0,
    total_amortized DOUBLE PRECISION NOT NULL DEFAULT 0,
    invoice_total DOUBLE PRECISION NOT NULL DEFAULT 0,
    drift_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    cost_view VARCHAR(16) NOT NULL DEFAULT 'amortized',
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, provider, billing_period, cost_view)
);
