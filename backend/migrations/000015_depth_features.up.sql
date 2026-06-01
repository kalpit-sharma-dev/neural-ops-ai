-- Production depth: APM retention, profiles, K8s API inventory, SLO burn alerts, workflows, cloud, RUM consent

CREATE TABLE IF NOT EXISTS trace_retention_policies (
    tenant_id VARCHAR(128) PRIMARY KEY,
    retention_days INT NOT NULL DEFAULT 30,
    head_sample_rate DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    tail_sample_rate DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS apm_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    service VARCHAR(255) NOT NULL,
    trace_id VARCHAR(128),
    span_id VARCHAR(128),
    function_name VARCHAR(512) NOT NULL,
    file_path VARCHAR(512),
    line_no INT NOT NULL DEFAULT 0,
    self_time_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    sample_count BIGINT NOT NULL DEFAULT 1,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_apm_profiles_service ON apm_profiles(tenant_id, service, recorded_at DESC);

CREATE TABLE IF NOT EXISTS collector_k8s_namespaces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'Active',
    pod_count INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

CREATE TABLE IF NOT EXISTS collector_k8s_deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    namespace VARCHAR(128) NOT NULL DEFAULT 'default',
    replicas INT NOT NULL DEFAULT 0,
    ready_replicas INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, namespace, name)
);

ALTER TABLE observability_slos
    ADD COLUMN IF NOT EXISTS burn_alert_threshold DOUBLE PRECISION NOT NULL DEFAULT 2.0;
ALTER TABLE observability_slos
    ADD COLUMN IF NOT EXISTS burn_alert_enabled BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS observability_workflow_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    workflow_id UUID NOT NULL REFERENCES observability_workflows(id) ON DELETE CASCADE,
    trigger_event VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'running',
    steps_log JSONB NOT NULL DEFAULT '[]',
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_wf ON observability_workflow_runs(workflow_id, started_at DESC);

CREATE TABLE IF NOT EXISTS observability_cloud_dashboards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    provider VARCHAR(32) NOT NULL,
    name VARCHAR(255) NOT NULL,
    region VARCHAR(128) NOT NULL DEFAULT 'global',
    metrics JSONB NOT NULL DEFAULT '[]',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS observability_integrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    integration_key VARCHAR(128) NOT NULL,
    name VARCHAR(255) NOT NULL,
    integration_type VARCHAR(64) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    connected BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, integration_key)
);

CREATE TABLE IF NOT EXISTS rum_consent_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    session_key VARCHAR(128) NOT NULL,
    consent_given BOOLEAN NOT NULL,
    consent_version VARCHAR(32) NOT NULL DEFAULT '1.0',
    ip_hash VARCHAR(64),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rum_consent_session ON rum_consent_logs(tenant_id, session_key);

INSERT INTO trace_retention_policies (tenant_id, retention_days, head_sample_rate, tail_sample_rate)
VALUES ('default', 30, 1.0, 1.0)
ON CONFLICT (tenant_id) DO NOTHING;

INSERT INTO observability_cloud_dashboards (tenant_id, provider, name, region, metrics)
SELECT 'default', 'aws', 'EC2 & Lambda Overview', 'us-east-1', '["CPUUtilization","Duration","Errors","Invocations"]'::jsonb
WHERE NOT EXISTS (SELECT 1 FROM observability_cloud_dashboards WHERE provider = 'aws' AND tenant_id = 'default');

INSERT INTO observability_cloud_dashboards (tenant_id, provider, name, region, metrics)
SELECT 'default', 'azure', 'App Service Health', 'eastus', '["HttpResponseTime","Http5xx","CpuPercentage"]'::jsonb
WHERE NOT EXISTS (SELECT 1 FROM observability_cloud_dashboards WHERE provider = 'azure' AND tenant_id = 'default');

INSERT INTO observability_cloud_dashboards (tenant_id, provider, name, region, metrics)
SELECT 'default', 'gcp', 'Cloud Run & GCE', 'us-central1', '["run.googleapis.com/request_count","compute.googleapis.com/instance/cpu/utilization"]'::jsonb
WHERE NOT EXISTS (SELECT 1 FROM observability_cloud_dashboards WHERE provider = 'gcp' AND tenant_id = 'default');
