-- Observability UI persistence (dashboards, SLOs, log rules, workflows, notebooks)

CREATE TABLE IF NOT EXISTS observability_dashboards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    description TEXT,
    shared BOOLEAN NOT NULL DEFAULT FALSE,
    tiles JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_obs_dashboards_tenant ON observability_dashboards(tenant_id);

CREATE TABLE IF NOT EXISTS observability_slos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    service VARCHAR(255) NOT NULL,
    sli_query TEXT NOT NULL,
    target DOUBLE PRECISION NOT NULL,
    window_days INT NOT NULL DEFAULT 30,
    error_budget DOUBLE PRECISION NOT NULL DEFAULT 100,
    burn_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'OK',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_obs_slos_tenant ON observability_slos(tenant_id);

CREATE TABLE IF NOT EXISTS observability_log_metric_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    pattern TEXT NOT NULL,
    service VARCHAR(255),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS observability_log_parsing_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    pattern TEXT NOT NULL,
    field VARCHAR(128) NOT NULL DEFAULT 'message',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS observability_workflows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    trigger_name VARCHAR(128) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    steps JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS observability_notebooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    cells JSONB NOT NULL DEFAULT '[]',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
