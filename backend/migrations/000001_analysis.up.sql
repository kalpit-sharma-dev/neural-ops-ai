-- Analysis engine schema (also applied programmatically on startup)

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS error_classifications (
    id UUID PRIMARY KEY,
    log_id UUID NOT NULL,
    service VARCHAR(255) NOT NULL,
    category VARCHAR(64) NOT NULL,
    confidence DOUBLE PRECISION NOT NULL,
    reasoning TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS log_explanations (
    id UUID PRIMARY KEY,
    log_id UUID NOT NULL,
    service VARCHAR(255) NOT NULL,
    plain_english TEXT NOT NULL,
    technical_summary TEXT NOT NULL,
    severity_assessment VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS recommendations (
    id UUID PRIMARY KEY,
    incident_id UUID,
    type VARCHAR(64) NOT NULL,
    description TEXT NOT NULL,
    code_snippet TEXT,
    priority VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
