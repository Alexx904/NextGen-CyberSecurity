-- 000001_init_schema.up.sql
-- NextGen CyberSecurity Platform - Initial Relational Schema

-- Abilitazione estensioni per generazione UUID e hashing
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- 1. TENANTS & USERS (Multi-Tenancy & Identity)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    logo_url TEXT,
    brand_colors JSONB NOT NULL DEFAULT '{"primary": "#0f172a", "secondary": "#2563eb", "accent": "#38bdf8"}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('SuperAdmin', 'TenantAdmin', 'SecOps', 'Auditor')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 2. TARGET SCOPE & ASSET MANAGEMENT
-- ============================================================================

CREATE TABLE IF NOT EXISTS target_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_target_groups_tenant_name UNIQUE (tenant_id, name)
);

CREATE TABLE IF NOT EXISTS targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    group_id UUID REFERENCES target_groups(id) ON DELETE SET NULL,
    target_type VARCHAR(50) NOT NULL CHECK (target_type IN ('IP', 'CIDR', 'RANGE', 'FQDN')),
    value VARCHAR(255) NOT NULL,
    description TEXT,
    is_blacklisted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 3. SCAN PROFILES, SCANS & SCHEDULES
-- ============================================================================

CREATE TABLE IF NOT EXISTS scan_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    nmap_args TEXT NOT NULL,
    description TEXT,
    timeout_seconds INTEGER NOT NULL DEFAULT 3600,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS scans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES scan_profiles(id) ON DELETE RESTRICT,
    trigger_type VARCHAR(50) NOT NULL DEFAULT 'MANUAL' CHECK (trigger_type IN ('MANUAL', 'SCHEDULED')),
    status VARCHAR(50) NOT NULL DEFAULT 'QUEUED' CHECK (status IN ('QUEUED', 'RUNNING', 'COMPLETED', 'FAILED', 'CANCELLED')),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    summary_stats JSONB NOT NULL DEFAULT '{}'::jsonb,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES scan_profiles(id) ON DELETE RESTRICT,
    cron_expression VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    next_run_at TIMESTAMPTZ,
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 4. CVE REPOSITORY & SCAN FINDINGS
-- ============================================================================

CREATE TABLE IF NOT EXISTS cve_entries (
    cve_id VARCHAR(50) PRIMARY KEY,
    cvss_score NUMERIC(3, 1),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'INFO')),
    description TEXT,
    cpe_match_criteria JSONB NOT NULL DEFAULT '[]'::jsonb,
    published_date TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scan_id UUID NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    host_ip VARCHAR(45) NOT NULL,
    hostname VARCHAR(255),
    port INTEGER NOT NULL,
    protocol VARCHAR(10) NOT NULL DEFAULT 'tcp',
    service_name VARCHAR(100),
    service_version VARCHAR(100),
    cve_id VARCHAR(50) REFERENCES cve_entries(cve_id) ON DELETE SET NULL,
    cvss_score NUMERIC(3, 1),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'INFO')),
    raw_banner TEXT,
    remediation TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 5. REPORT TEMPLATES (White-Label)
-- ============================================================================

CREATE TABLE IF NOT EXISTS report_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 6. INDEXES (Performance & Multi-Tenant Query Optimization)
-- ============================================================================

-- B-Tree Indexes per isolamento Tenant
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_target_groups_tenant_id ON target_groups(tenant_id);
CREATE INDEX IF NOT EXISTS idx_targets_tenant_id ON targets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_targets_tenant_type ON targets(tenant_id, target_type);
CREATE INDEX IF NOT EXISTS idx_targets_group_id ON targets(group_id);
CREATE INDEX IF NOT EXISTS idx_scans_tenant_status ON scans(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_scans_target_id ON scans(target_id);
CREATE INDEX IF NOT EXISTS idx_scans_profile_id ON scans(profile_id);
CREATE INDEX IF NOT EXISTS idx_schedules_tenant_active ON schedules(tenant_id, is_active);
CREATE INDEX IF NOT EXISTS idx_schedules_next_run ON schedules(next_run_at) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_findings_scan_id ON findings(scan_id);
CREATE INDEX IF NOT EXISTS idx_findings_tenant_severity ON findings(tenant_id, severity);
CREATE INDEX IF NOT EXISTS idx_findings_cve_id ON findings(cve_id);
CREATE INDEX IF NOT EXISTS idx_findings_host_port ON findings(host_ip, port);
CREATE INDEX IF NOT EXISTS idx_cve_entries_severity ON cve_entries(severity);
CREATE INDEX IF NOT EXISTS idx_report_templates_tenant_id ON report_templates(tenant_id);

-- GIN Indexes per campi JSONB (NFR-PERF-03)
CREATE INDEX IF NOT EXISTS idx_tenants_brand_colors ON tenants USING GIN (brand_colors);
CREATE INDEX IF NOT EXISTS idx_scans_summary_stats ON scans USING GIN (summary_stats);
CREATE INDEX IF NOT EXISTS idx_cve_entries_cpe_gin ON cve_entries USING GIN (cpe_match_criteria jsonb_path_ops);

-- ============================================================================
-- 7. SEED DATA (Default Scan Profiles)
-- ============================================================================

INSERT INTO scan_profiles (id, name, nmap_args, description, timeout_seconds)
VALUES 
    (gen_random_uuid(), 'Fast Discovery', '-T4 -F', 'Discovery rapida degli host attivi e delle top 100 porte TCP', 900),
    (gen_random_uuid(), 'Standard Service Scan', '-sV -T4', 'Scansione standard delle top 1000 porte TCP con rilevamento versioni dei servizi', 3600),
    (gen_random_uuid(), 'Full Deep Scan', '-p- -sV -O -T4', 'Scansione completa su 65535 porte TCP, OS detection e rilevamento servizi approfondito', 14400)
ON CONFLICT (name) DO NOTHING;
