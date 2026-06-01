DROP TABLE IF EXISTS service_health;
ALTER TABLE service_dependencies DROP COLUMN IF EXISTS error_rate;
ALTER TABLE service_dependencies DROP COLUMN IF EXISTS p50_latency_ms;
ALTER TABLE service_dependencies DROP COLUMN IF EXISTS tenant_id;
