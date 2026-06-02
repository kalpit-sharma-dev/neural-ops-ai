-- Per-tenant install state for marketplace apps & extensions. The catalog
-- itself is code-defined; this table records which extensions a tenant has
-- installed along with their (non-secret + secret) configuration.
CREATE TABLE IF NOT EXISTS marketplace_installs (
    tenant_id     VARCHAR(128) NOT NULL DEFAULT 'default',
    extension_key VARCHAR(128) NOT NULL,
    config        JSONB        NOT NULL DEFAULT '{}'::jsonb,
    installed_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, extension_key)
);
