-- 000006_scan_profile_categories.down.sql

DROP INDEX IF EXISTS idx_scan_profiles_tenant_id;
DROP INDEX IF EXISTS idx_scan_profiles_category;

DELETE FROM scan_profiles WHERE is_custom = true;
DELETE FROM scan_profiles WHERE name IN (
    'Stealth SYN Penetration Audit',
    'Firewall & Filter Bypass Check',
    'OS & System Daemon Flaws Audit',
    'Critical Services & RCE Surface',
    'Web & API Surface Audit',
    'Full Web Stack Deep Versioning',
    'Fast Ping & Host Sweep (LAN)',
    'Legacy & Insecure Protocol Check'
);

ALTER TABLE scan_profiles
    DROP COLUMN IF EXISTS tenant_id,
    DROP COLUMN IF EXISTS is_custom,
    DROP COLUMN IF EXISTS category;

