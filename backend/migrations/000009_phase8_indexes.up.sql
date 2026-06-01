-- Phase 8: Composite indexes for tenant/time/service queries

CREATE INDEX IF NOT EXISTS idx_incidents_tenant_start_time ON incidents(tenant_id, start_time DESC);
CREATE INDEX IF NOT EXISTS idx_incidents_tenant_severity ON incidents(tenant_id, severity);

CREATE INDEX IF NOT EXISTS idx_alerts_tenant_fired_at ON alerts(tenant_id, fired_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_tenant_fingerprint ON alerts(tenant_id, fingerprint);
CREATE INDEX IF NOT EXISTS idx_alerts_tenant_service_severity ON alerts(tenant_id, service, severity);

CREATE INDEX IF NOT EXISTS idx_deployments_tenant_service_deployed ON deployments(tenant_id, service, deployed_at DESC);

CREATE INDEX IF NOT EXISTS idx_error_patterns_tenant_category ON error_patterns(tenant_id, category);
