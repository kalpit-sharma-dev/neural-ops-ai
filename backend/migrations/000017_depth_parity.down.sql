ALTER TABLE tenant_policies DROP COLUMN IF EXISTS sso_client_secret;
ALTER TABLE oncall_schedules DROP COLUMN IF EXISTS pagerduty_synced_at;
ALTER TABLE oncall_schedules DROP COLUMN IF EXISTS pagerduty_schedule_id;
DROP TABLE IF EXISTS integration_oauth_states;
