-- Phase 8: Deployments and incident correlations

CREATE TABLE IF NOT EXISTS deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    service VARCHAR(255) NOT NULL,
    version VARCHAR(128) NOT NULL,
    environment VARCHAR(64) NOT NULL DEFAULT 'production',
    deployed_by VARCHAR(255),
    deployed_at TIMESTAMPTZ NOT NULL,
    change_type VARCHAR(64) NOT NULL DEFAULT 'deploy',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deployments_tenant_deployed_at ON deployments(tenant_id, deployed_at);
CREATE INDEX IF NOT EXISTS idx_deployments_service ON deployments(service);

CREATE TABLE IF NOT EXISTS deployment_incident_correlations (
    deployment_id UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    correlation_strength DOUBLE PRECISION NOT NULL DEFAULT 0,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (deployment_id, incident_id)
);

CREATE INDEX IF NOT EXISTS idx_deployment_incident_correlations_incident ON deployment_incident_correlations(incident_id);
