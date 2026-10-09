-- ============================================================================
-- 000005: SSH Servers & Firewall Backups Schema
-- ============================================================================

CREATE TABLE IF NOT EXISTS ssh_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL DEFAULT 22 CHECK (port > 0 AND port <= 65535),
    username VARCHAR(100) NOT NULL,
    auth_type VARCHAR(20) NOT NULL CHECK (auth_type IN ('KEY', 'PASSWORD')),
    encrypted_secret TEXT NOT NULL,
    public_key_fingerprint VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'UNCHECKED' CHECK (status IN ('ONLINE', 'OFFLINE', 'UNCHECKED')),
    last_connected_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_ssh_servers_tenant_name UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_ssh_servers_tenant_id ON ssh_servers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ssh_servers_status ON ssh_servers(status);

CREATE TABLE IF NOT EXISTS firewall_backups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    server_id UUID NOT NULL REFERENCES ssh_servers(id) ON DELETE CASCADE,
    raw_iptables_rules TEXT NOT NULL,
    label VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_firewall_backups_tenant_id ON firewall_backups(tenant_id);
CREATE INDEX IF NOT EXISTS idx_firewall_backups_server_id ON firewall_backups(server_id);
CREATE INDEX IF NOT EXISTS idx_firewall_backups_created_at ON firewall_backups(created_at DESC);
