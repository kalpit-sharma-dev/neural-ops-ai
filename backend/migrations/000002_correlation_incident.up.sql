-- Phase 4: Correlation & Incident Engine schema

CREATE TABLE IF NOT EXISTS deployment_correlations (
    id UUID PRIMARY KEY,
    deployment_id UUID NOT NULL,
    service VARCHAR(255) NOT NULL,
    version VARCHAR(128) NOT NULL,
    deployed_at TIMESTAMPTZ NOT NULL,
    correlation_level VARCHAR(16) NOT NULL,
    error_rate_before DOUBLE PRECISION NOT NULL DEFAULT 0,
    error_rate_after DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deployment_correlations_service ON deployment_correlations(service);
CREATE INDEX IF NOT EXISTS idx_deployment_correlations_deployed_at ON deployment_correlations(deployed_at);

CREATE TABLE IF NOT EXISTS service_dependencies (
    id UUID PRIMARY KEY,
    source_service VARCHAR(255) NOT NULL,
    target_service VARCHAR(255) NOT NULL,
    call_count BIGINT NOT NULL DEFAULT 0,
    p99_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source_service, target_service)
);

CREATE TABLE IF NOT EXISTS temporal_correlations (
    id UUID PRIMARY KEY,
    window_start TIMESTAMPTZ NOT NULL,
    window_end TIMESTAMPTZ NOT NULL,
    affected_services TEXT[] NOT NULL,
    common_cause_hint TEXT,
    infra_events TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY,
    title VARCHAR(512) NOT NULL,
    summary TEXT NOT NULL,
    severity VARCHAR(8) NOT NULL,
    status VARCHAR(32) NOT NULL,
    affected_services TEXT[] NOT NULL,
    root_service VARCHAR(255),
    error_category VARCHAR(64),
    fingerprint VARCHAR(128) NOT NULL,
    blast_radius TEXT[],
    start_time TIMESTAMPTZ NOT NULL,
    resolved_time TIMESTAMPTZ,
    acknowledged_at TIMESTAMPTZ,
    mttr_ns BIGINT DEFAULT 0,
    timeline JSONB DEFAULT '[]'::jsonb,
    recommendations JSONB DEFAULT '[]'::jsonb,
    root_cause_analysis JSONB,
    resolution_notes TEXT,
    tenant_id VARCHAR(128) DEFAULT 'default',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_incidents_fingerprint_active
    ON incidents(fingerprint)
    WHERE status IN ('OPEN', 'INVESTIGATING');

CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incidents(severity);
CREATE INDEX IF NOT EXISTS idx_incidents_start_time ON incidents(start_time);

CREATE TABLE IF NOT EXISTS mttr_history (
    id UUID PRIMARY KEY,
    incident_id UUID NOT NULL REFERENCES incidents(id),
    service VARCHAR(255) NOT NULL,
    team VARCHAR(255),
    mttr_ns BIGINT NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS incident_suppression_rules (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    service_pattern VARCHAR(255),
    error_category VARCHAR(64),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
