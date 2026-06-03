-- SRS Phases 1-8 parity persistence (alert policies, AI, governance, RUM, exports, NFR, fleet)

CREATE TABLE IF NOT EXISTS alert_policies (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    service_pattern VARCHAR(255) NOT NULL DEFAULT '*',
    severity VARCHAR(16) NOT NULL DEFAULT 'P2',
    enabled BOOLEAN NOT NULL DEFAULT true,
    expression TEXT NOT NULL DEFAULT '',
    routes JSONB NOT NULL DEFAULT '[]',
    context JSONB NOT NULL DEFAULT '{}',
    dedupe_key VARCHAR(255) NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alert_policies_tenant ON alert_policies (tenant_id);

CREATE TABLE IF NOT EXISTS alert_suppressions (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    service_pattern VARCHAR(255) NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    created_by VARCHAR(255) NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alert_suppressions_tenant ON alert_suppressions (tenant_id);

CREATE TABLE IF NOT EXISTS derived_metrics (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    expression TEXT NOT NULL,
    unit VARCHAR(64) NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

CREATE TABLE IF NOT EXISTS ai_explanations (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    incident_id VARCHAR(128) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ai_explanations_incident ON ai_explanations (tenant_id, incident_id);

CREATE TABLE IF NOT EXISTS ai_forecasts (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS autofix_plans (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    incident_id VARCHAR(128) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS autofix_actions (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    plan_id VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS security_findings (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    title TEXT NOT NULL,
    category VARCHAR(32) NOT NULL,
    severity VARCHAR(16) NOT NULL,
    service VARCHAR(255) NOT NULL DEFAULT '',
    asset VARCHAR(512) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    exploitability VARCHAR(32) NOT NULL DEFAULT '',
    incident_id VARCHAR(128) NOT NULL DEFAULT '',
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_security_findings_tenant ON security_findings (tenant_id);

CREATE TABLE IF NOT EXISTS collector_fleet_agents (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    environment VARCHAR(64) NOT NULL DEFAULT 'prod',
    version VARCHAR(32) NOT NULL,
    target_version VARCHAR(32) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'healthy',
    policy_id VARCHAR(64) NOT NULL DEFAULT 'default',
    last_heartbeat_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_collector_fleet_tenant ON collector_fleet_agents (tenant_id);

CREATE TABLE IF NOT EXISTS rum_funnels (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS synthetic_browser_tests (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS synthetic_mobile_tests (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS synthetic_private_locations (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS business_kpi_packs (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    payload JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tenant_governance (
    tenant_id VARCHAR(128) PRIMARY KEY,
    abac JSONB NOT NULL DEFAULT '{}',
    residency JSONB NOT NULL DEFAULT '{}',
    branding JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS msp_tenants (
    id VARCHAR(64) PRIMARY KEY,
    parent_tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(128) NOT NULL,
    plan VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL,
    user_count INT NOT NULL DEFAULT 0,
    region VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS export_jobs (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    export_type VARCHAR(32) NOT NULL,
    destination TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nfr_certification (
    tenant_id VARCHAR(128) PRIMARY KEY,
    benchmarks JSONB NOT NULL DEFAULT '[]',
    reliability JSONB NOT NULL DEFAULT '[]',
    accessibility JSONB NOT NULL DEFAULT '{}',
    locales JSONB NOT NULL DEFAULT '[]',
    certification JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO tenant_governance (tenant_id, abac, residency, branding)
VALUES (
    'default',
    '{"enabled":true,"rules":[{"id":"abac-1","effect":"allow","action":"read","resource":"logs:*","condition":"user.department == resource.team"},{"id":"abac-2","effect":"deny","action":"export","resource":"pii:*","condition":"user.clearance < 3"}]}',
    '{"primaryRegion":"us-east-1","allowedRegions":["us-east-1","eu-west-1"],"piiStorageRegion":"us-east-1","crossBorderDenied":true}',
    '{"productName":"NeuralOps","logoUrl":"/brand/logo.svg","primaryColor":"#2563eb","accentColor":"#0ea5e9","supportEmail":"support@neuralops.ai","customDomain":"observe.acme-corp.com"}'
)
ON CONFLICT (tenant_id) DO NOTHING;
