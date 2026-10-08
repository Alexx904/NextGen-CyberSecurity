-- ============================================================================
-- 000002: CVE Catalog Optimization Indexes
-- ============================================================================

-- Indice B-Tree per ranking CVSS Score (ricerche ordinate per severità e filtri MinCVSS)
CREATE INDEX IF NOT EXISTS idx_cve_entries_cvss ON cve_entries(cvss_score DESC NULLS LAST);

-- Indice B-Tree per tracking dell'ultimo aggiornamento e sync catalogo
CREATE INDEX IF NOT EXISTS idx_cve_entries_updated_at ON cve_entries(updated_at DESC);

