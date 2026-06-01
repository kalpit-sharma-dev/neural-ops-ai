-- Phase 7: Alerting Service schema

CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    source VARCHAR(32) NOT NULL,
    alert_name VARCHAR(255) NOT NULL,
    service VARCHAR(255) NOT NULL,
    title VARCHAR(512) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(8) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'FIRING',
    fingerprint VARCHAR(128) NOT NULL,
    group_id UUID,
    occurrence_count INT NOT NULL DEFAULT 1,
    labels JSONB NOT NULL DEFAULT '{}'::jsonb,
    linked_incident_id UUID,
    ai_explanation TEXT,
    fired_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL,
    resolved_at TIMESTAMPTZ,
    acknowledged_at TIMESTAMPTZ,
    suppressed_until TIMESTAMPTZ,
    deduplicated BOOLEAN NOT NULL DEFAULT FALSE,
    escalation_level INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_service ON alerts(service);
CREATE INDEX IF NOT EXISTS idx_alerts_fingerprint ON alerts(fingerprint);
CREATE INDEX IF NOT EXISTS idx_alerts_fired_at ON alerts(fired_at);
CREATE INDEX IF NOT EXISTS idx_alerts_group_id ON alerts(group_id);

CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    source VARCHAR(32) NOT NULL,
    service_pattern VARCHAR(255),
    severity VARCHAR(8) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    labels JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notification_channels (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    channel_type VARCHAR(32) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS alert_silences (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    service_pattern VARCHAR(255),
    alert_name_pattern VARCHAR(255),
    reason TEXT,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS escalation_policies (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    service_pattern VARCHAR(255),
    levels JSONB NOT NULL DEFAULT '[]'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS oncall_schedules (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    team VARCHAR(255) NOT NULL,
    rotation JSONB NOT NULL DEFAULT '[]'::jsonb,
    timezone VARCHAR(64) NOT NULL DEFAULT 'UTC',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
