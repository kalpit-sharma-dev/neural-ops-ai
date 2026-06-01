-- Phase 18: One-time OIDC SPA exchange codes (avoid tokens in URL)

CREATE TABLE IF NOT EXISTS auth_exchange_codes (
    code_hash VARCHAR(64) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_exchange_codes_expires_at ON auth_exchange_codes(expires_at);
