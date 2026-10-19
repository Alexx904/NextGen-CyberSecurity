CREATE TABLE assessment_assets (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    group_id UUID NOT NULL,
    name VARCHAR(200) NOT NULL CHECK (length(trim(name)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, group_id, name),
    FOREIGN KEY (tenant_id,group_id) REFERENCES target_groups(tenant_id,id) ON DELETE RESTRICT
);
CREATE TABLE assessment_issues (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    asset_id UUID NOT NULL,
    protocol VARCHAR(3) NOT NULL CHECK (protocol IN ('tcp','udp')),
    port INTEGER NOT NULL CHECK (port BETWEEN 1 AND 65535),
    issue_key VARCHAR(200) NOT NULL CHECK (length(trim(issue_key)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, asset_id, protocol, port, issue_key),
    FOREIGN KEY (tenant_id, asset_id) REFERENCES assessment_assets(tenant_id,id) ON DELETE RESTRICT
);
CREATE TABLE assessment_issue_events (
    tenant_id UUID NOT NULL,
    issue_id UUID NOT NULL,
    version INTEGER NOT NULL CHECK (version > 0),
    payload JSONB NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
    finding_id UUID,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (tenant_id, issue_id, version),
    UNIQUE (tenant_id, finding_id),
    FOREIGN KEY (tenant_id, issue_id) REFERENCES assessment_issues(tenant_id,id) ON DELETE RESTRICT
);
CREATE TABLE assessment_control_evidence (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    group_id UUID NOT NULL,
    control_id VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    FOREIGN KEY (tenant_id,group_id) REFERENCES target_groups(tenant_id,id) ON DELETE RESTRICT
);
CREATE INDEX assessment_evidence_scope ON assessment_control_evidence(tenant_id,group_id,recorded_at,id);
CREATE INDEX assessment_issue_scope ON assessment_issues(tenant_id,asset_id);
CREATE INDEX assessment_assets_scope ON assessment_assets(tenant_id,group_id);
CREATE FUNCTION assessment_audit_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'Assessment audit records are append-only'; END;
$$;
CREATE TRIGGER assessment_events_immutable BEFORE UPDATE OR DELETE OR TRUNCATE ON assessment_issue_events
FOR EACH STATEMENT EXECUTE FUNCTION assessment_audit_immutable();
CREATE TRIGGER assessment_evidence_immutable BEFORE UPDATE OR DELETE OR TRUNCATE ON assessment_control_evidence
FOR EACH STATEMENT EXECUTE FUNCTION assessment_audit_immutable();
