-- FinOps Wave 3: chargeback, scenarios, governance

CREATE TABLE IF NOT EXISTS finops_chargeback_statements (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    cost_center VARCHAR(128) NOT NULL,
    mode VARCHAR(16) NOT NULL DEFAULT 'showback',
    period VARCHAR(32) NOT NULL DEFAULT 'monthly',
    total_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    line_items JSONB NOT NULL DEFAULT '[]',
    export_url VARCHAR(512) NOT NULL DEFAULT '',
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_finops_chargeback_tenant ON finops_chargeback_statements (tenant_id, cost_center, generated_at DESC);

CREATE TABLE IF NOT EXISTS finops_scenario_runs (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    name VARCHAR(255) NOT NULL,
    scenario_type VARCHAR(64) NOT NULL,
    params JSONB NOT NULL DEFAULT '{}',
    baseline_cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    projected_cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    cost_delta_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    baseline_co2e_kg DOUBLE PRECISION NOT NULL DEFAULT 0,
    projected_co2e_kg DOUBLE PRECISION NOT NULL DEFAULT 0,
    co2e_delta_kg DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_finops_scenarios_tenant ON finops_scenario_runs (tenant_id, created_at DESC);

CREATE TABLE IF NOT EXISTS finops_governance_policies (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    name VARCHAR(255) NOT NULL,
    residency_region VARCHAR(64) NOT NULL DEFAULT '',
    allowed_scopes JSONB NOT NULL DEFAULT '[]',
    chargeback_mode VARCHAR(16) NOT NULL DEFAULT 'showback',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS finops_carbon_actions (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    recommendation_id VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'proposed',
    applied_at TIMESTAMPTZ,
    co2e_reduction_kg DOUBLE PRECISION NOT NULL DEFAULT 0,
    cost_delta_usd DOUBLE PRECISION NOT NULL DEFAULT 0
);
