-- Phase 8: Incident extensions

ALTER TABLE incidents ADD COLUMN IF NOT EXISTS deployment_id UUID;
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS mttr_seconds BIGINT DEFAULT 0;

CREATE TABLE IF NOT EXISTS incident_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    event_type VARCHAR(64) NOT NULL,
    description TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_incident_events_incident_id ON incident_events(incident_id);
CREATE INDEX IF NOT EXISTS idx_incident_events_occurred_at ON incident_events(occurred_at);

CREATE TABLE IF NOT EXISTS incident_services (
    incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    service_name VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL CHECK (role IN ('PRIMARY', 'DOWNSTREAM')),
    PRIMARY KEY (incident_id, service_name)
);

CREATE INDEX IF NOT EXISTS idx_incident_services_service_name ON incident_services(service_name);
