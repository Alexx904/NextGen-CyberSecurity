-- ============================================================================
-- 000002: Rollback CVE Catalog Optimization Indexes
-- ============================================================================

DROP INDEX IF EXISTS idx_cve_entries_cvss;
DROP INDEX IF EXISTS idx_cve_entries_updated_at;

