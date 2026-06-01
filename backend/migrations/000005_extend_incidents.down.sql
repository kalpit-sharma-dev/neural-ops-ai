DROP TABLE IF EXISTS incident_services;
DROP TABLE IF EXISTS incident_events;
ALTER TABLE incidents DROP COLUMN IF EXISTS mttr_seconds;
ALTER TABLE incidents DROP COLUMN IF EXISTS deployment_id;
