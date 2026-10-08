-- ============================================================================
-- 000003: Drop Unused report_templates Table (Cleanup Dead Schema)
-- ============================================================================

DROP INDEX IF EXISTS idx_report_templates_tenant_id;
DROP TABLE IF EXISTS report_templates CASCADE;

