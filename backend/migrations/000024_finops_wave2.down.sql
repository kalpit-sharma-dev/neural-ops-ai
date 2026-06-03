DROP TABLE IF EXISTS finops_audit_log;
DROP TABLE IF EXISTS finops_report_schedules;
DROP TABLE IF EXISTS finops_carbon_samples;
DROP TABLE IF EXISTS finops_carbon_factors;
DROP TABLE IF EXISTS finops_commitments;
DROP TABLE IF EXISTS finops_anomaly_sensitivity;
DROP TABLE IF EXISTS finops_tag_suggestions;
DROP TABLE IF EXISTS finops_shared_splits;
DROP TABLE IF EXISTS finops_import_batches;

ALTER TABLE finops_anomalies DROP COLUMN IF EXISTS probable_cause;
ALTER TABLE finops_anomalies DROP COLUMN IF EXISTS sensitivity_factor;
ALTER TABLE finops_ingest_snapshots DROP COLUMN IF EXISTS version;
ALTER TABLE finops_recommendations DROP COLUMN IF EXISTS realized_savings_usd;
ALTER TABLE finops_recommendations DROP COLUMN IF EXISTS ticket_id;
ALTER TABLE finops_recommendations DROP COLUMN IF EXISTS ticket_url;
