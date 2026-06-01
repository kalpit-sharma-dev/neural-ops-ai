-- P0-P3: mobile push tokens, tenant policies, on-call UI seed

CREATE TABLE IF NOT EXISTS mobile_push_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(128) NOT NULL DEFAULT 'default',
    user_id VARCHAR(255) NOT NULL DEFAULT '',
    expo_push_token VARCHAR(512) NOT NULL,
    platform VARCHAR(32) NOT NULL DEFAULT 'unknown',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, expo_push_token)
);

CREATE TABLE IF NOT EXISTS tenant_policies (
    tenant_id VARCHAR(128) PRIMARY KEY,
    log_retention_days INT NOT NULL DEFAULT 90,
    ingestion_rate_limit INT NOT NULL DEFAULT 100000,
    sso_provider VARCHAR(64) NOT NULL DEFAULT 'oidc',
    sso_metadata_url TEXT,
    sso_client_id VARCHAR(255),
    sso_issuer_url TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO tenant_policies (tenant_id, log_retention_days, sso_provider, sso_issuer_url)
VALUES ('default', 90, 'oidc', 'http://localhost:8088/realms/neuralops')
ON CONFLICT (tenant_id) DO NOTHING;

INSERT INTO oncall_schedules (id, tenant_id, team, rotation, timezone, enabled)
SELECT uuid_generate_v4(), 'default', 'platform-sre',
  '[{"name":"Primary","email":"sre@neuralops.ai","after":"0m"},{"name":"Manager","email":"manager@neuralops.ai","after":"30m"}]'::jsonb,
  'UTC', true
WHERE NOT EXISTS (SELECT 1 FROM oncall_schedules WHERE tenant_id = 'default' AND team = 'platform-sre');
