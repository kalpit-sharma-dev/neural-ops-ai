-- Depth parity: integration OAuth state, on-call PagerDuty sync

CREATE TABLE IF NOT EXISTS integration_oauth_states (
    state VARCHAR(128) PRIMARY KEY,
    tenant_id VARCHAR(128) NOT NULL,
    integration_key VARCHAR(128) NOT NULL,
    redirect_uri TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '15 minutes'
);

CREATE INDEX IF NOT EXISTS idx_integration_oauth_expires ON integration_oauth_states(expires_at);

ALTER TABLE oncall_schedules
    ADD COLUMN IF NOT EXISTS pagerduty_schedule_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS pagerduty_synced_at TIMESTAMPTZ;

ALTER TABLE tenant_policies
    ADD COLUMN IF NOT EXISTS sso_client_secret TEXT;
