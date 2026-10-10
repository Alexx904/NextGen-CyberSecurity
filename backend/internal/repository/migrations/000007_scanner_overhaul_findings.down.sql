-- 000007_scanner_overhaul_findings.down.sql

DELETE FROM scan_profiles WHERE name = 'Specialized UDP Critical Services';

DROP INDEX IF EXISTS idx_scan_profiles_invasiveness;
ALTER TABLE scan_profiles DROP COLUMN IF EXISTS invasiveness;

DROP INDEX IF EXISTS idx_findings_port_state;
DROP INDEX IF EXISTS idx_findings_detection_method;
DROP INDEX IF EXISTS idx_findings_confidence;
ALTER TABLE findings
    DROP COLUMN IF EXISTS confidence,
    DROP COLUMN IF EXISTS detection_method,
    DROP COLUMN IF EXISTS cpe,
    DROP COLUMN IF EXISTS port_state;

