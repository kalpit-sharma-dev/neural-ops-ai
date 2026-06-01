-- Production collectors: K8s, hosts, RUM, synthetic telemetry

CREATE TABLE IF NOT EXISTS collector_hosts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'up',
    cpu_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    memory_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    disk_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    zone VARCHAR(128) NOT NULL DEFAULT 'default',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

CREATE TABLE IF NOT EXISTS collector_k8s_clusters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    nodes INT NOT NULL DEFAULT 0,
    pods INT NOT NULL DEFAULT 0,
    health VARCHAR(32) NOT NULL DEFAULT 'healthy',
    namespace_count INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

CREATE TABLE IF NOT EXISTS collector_k8s_pods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    namespace VARCHAR(128) NOT NULL DEFAULT 'default',
    node VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'Running',
    cpu_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    memory_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    restarts INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, namespace, name)
);

CREATE TABLE IF NOT EXISTS rum_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    session_key VARCHAR(128) NOT NULL,
    user_id VARCHAR(255) NOT NULL DEFAULT '',
    page VARCHAR(512) NOT NULL DEFAULT '/',
    device VARCHAR(128) NOT NULL DEFAULT 'desktop',
    country VARCHAR(64) NOT NULL DEFAULT 'unknown',
    duration_ms BIGINT NOT NULL DEFAULT 0,
    errors INT NOT NULL DEFAULT 0,
    lcp DOUBLE PRECISION NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, session_key)
);

CREATE INDEX IF NOT EXISTS idx_rum_sessions_tenant ON rum_sessions(tenant_id, started_at DESC);

CREATE TABLE IF NOT EXISTS rum_replay_events (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    session_key VARCHAR(128) NOT NULL,
    seq INT NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rum_replay_session ON rum_replay_events(tenant_id, session_key, seq);

CREATE TABLE IF NOT EXISTS synthetic_monitors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    monitor_type VARCHAR(32) NOT NULL DEFAULT 'http',
    url TEXT NOT NULL,
    interval_label VARCHAR(32) NOT NULL DEFAULT '5m',
    locations JSONB NOT NULL DEFAULT '[]',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS synthetic_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    monitor_id UUID NOT NULL REFERENCES synthetic_monitors(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL,
    latency_ms BIGINT NOT NULL DEFAULT 0,
    location VARCHAR(128) NOT NULL DEFAULT 'default',
    ran_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_synthetic_runs_monitor ON synthetic_runs(monitor_id, ran_at DESC);
