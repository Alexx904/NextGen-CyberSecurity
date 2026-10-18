CREATE TABLE discovery_sessions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subnet TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    duration_seconds DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (duration_seconds >= 0),
    UNIQUE (tenant_id, id)
);

CREATE INDEX idx_discovery_sessions_tenant_started
    ON discovery_sessions (tenant_id, started_at DESC, id DESC);

CREATE TABLE discovery_devices (
    tenant_id UUID NOT NULL,
    session_id UUID NOT NULL,
    ip INET NOT NULL,
    device JSONB NOT NULL CHECK (jsonb_typeof(device) = 'object'),
    PRIMARY KEY (tenant_id, session_id, ip),
    FOREIGN KEY (tenant_id, session_id)
        REFERENCES discovery_sessions(tenant_id, id) ON DELETE CASCADE
);
