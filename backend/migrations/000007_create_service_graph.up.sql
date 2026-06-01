-- Phase 8: Service graph and health scores

ALTER TABLE service_dependencies ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(128) NOT NULL DEFAULT 'default';
ALTER TABLE service_dependencies ADD COLUMN IF NOT EXISTS p50_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE service_dependencies ADD COLUMN IF NOT EXISTS error_rate DOUBLE PRECISION NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS service_health (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    service_name VARCHAR(255) NOT NULL,
    health_score DOUBLE PRECISION NOT NULL DEFAULT 100,
    error_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
    latency_p99 DOUBLE PRECISION NOT NULL DEFAULT 0,
    anomaly_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, service_name)
);

CREATE INDEX IF NOT EXISTS idx_service_health_tenant_updated ON service_health(tenant_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_service_dependencies_tenant_source ON service_dependencies(tenant_id, source_service);
