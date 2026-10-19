-- Removes channel configuration, observation baselines and delivery audit/history.
-- Back up these three tables before an application rollback. Test rollback only
-- on a named disposable database. This does not recall already sent messages.
DROP TRIGGER alert_scan_terminal ON scans;
DROP FUNCTION alert_on_scan();
DROP TRIGGER alert_device_change ON discovery_devices;
DROP FUNCTION alert_on_device();
DROP TRIGGER alert_finding_insert ON findings;
DROP FUNCTION alert_on_finding();
DROP FUNCTION alert_record_observation(UUID,TEXT,JSONB,JSONB,UUID);
DROP FUNCTION alert_device_signature(JSONB);
DROP TRIGGER IF EXISTS alert_company_identity ON target_groups;
DROP FUNCTION IF EXISTS alert_preserve_company_scope();
DROP TRIGGER IF EXISTS alert_channel_company ON alert_channels;
DROP FUNCTION IF EXISTS alert_validate_company_scope();
DROP TABLE alert_jobs;
DROP TABLE alert_observations;
DROP TABLE alert_channels;
DROP INDEX idx_target_groups_tenant_identity;
