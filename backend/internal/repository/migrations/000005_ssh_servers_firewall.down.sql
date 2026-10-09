-- ============================================================================
-- 000005 DOWN: Drop SSH Servers & Firewall Backups
-- ============================================================================

DROP TABLE IF EXISTS firewall_backups CASCADE;
DROP TABLE IF EXISTS ssh_servers CASCADE;
