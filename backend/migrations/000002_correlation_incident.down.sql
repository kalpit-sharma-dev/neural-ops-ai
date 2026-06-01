-- Rollback Phase 4 schema

DROP TABLE IF EXISTS incident_suppression_rules;
DROP TABLE IF EXISTS mttr_history;
DROP TABLE IF EXISTS incidents;
DROP TABLE IF EXISTS temporal_correlations;
DROP TABLE IF EXISTS service_dependencies;
DROP TABLE IF EXISTS deployment_correlations;
