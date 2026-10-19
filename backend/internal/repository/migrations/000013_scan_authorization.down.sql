-- Destructive only to the new consent_snapshot evidence. Back up before rollback.
DROP TRIGGER schedules_authorization_record ON schedules;
DROP TRIGGER scans_authorization_record ON scans;
DROP FUNCTION record_scan_authorization();
DROP TABLE scan_authorization_audit;
DROP FUNCTION reject_scan_authorization_mutation();
ALTER TABLE schedules DROP COLUMN consent_snapshot;
ALTER TABLE scans DROP COLUMN consent_snapshot;
