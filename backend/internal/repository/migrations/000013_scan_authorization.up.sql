ALTER TABLE scans ADD COLUMN consent_snapshot JSONB;
ALTER TABLE schedules ADD COLUMN consent_snapshot JSONB;

-- No cascading inventory FK: consent_snapshot evidence outlives mutable inventory.
CREATE TABLE scan_authorization_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    object_type TEXT NOT NULL CHECK (object_type IN ('scans', 'schedules')),
    object_id UUID NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    evidence JSONB NOT NULL,
    UNIQUE (object_type, object_id)
);
CREATE INDEX scan_authorization_audit_tenant_time ON scan_authorization_audit (tenant_id, recorded_at DESC, id);

CREATE FUNCTION reject_scan_authorization_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'scan consent_snapshot evidence is append-only';
END;
$$;
CREATE TRIGGER scan_authorization_audit_immutable
BEFORE UPDATE OR DELETE OR TRUNCATE ON scan_authorization_audit
FOR EACH STATEMENT EXECUTE FUNCTION reject_scan_authorization_mutation();
REVOKE UPDATE, DELETE, TRUNCATE ON scan_authorization_audit FROM PUBLIC;

CREATE FUNCTION record_scan_authorization() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        IF NEW.consent_snapshot IS DISTINCT FROM OLD.consent_snapshot THEN
            RAISE EXCEPTION 'scan consent_snapshot snapshot cannot be changed';
        END IF;
        RETURN NEW;
    END IF;
    IF NEW.consent_snapshot IS NOT NULL THEN
        IF NEW.consent_snapshot->>'statement' IS DISTINCT FROM 'scan-authorization-v1'
           OR NEW.consent_snapshot->>'accepted' IS DISTINCT FROM 'true'
           OR (NEW.consent_snapshot->>'tenant_id')::uuid IS DISTINCT FROM NEW.tenant_id
           OR NEW.consent_snapshot->>'user_id' IS NULL
           OR (NEW.consent_snapshot->>'user_id')::uuid = '00000000-0000-0000-0000-000000000000'::uuid
           OR NEW.consent_snapshot->>'accepted_at' IS NULL THEN
            RAISE EXCEPTION 'invalid consent_snapshot identity';
        END IF;
        INSERT INTO scan_authorization_audit (tenant_id, user_id, object_type, object_id, evidence)
        VALUES (NEW.tenant_id, (NEW.consent_snapshot->>'user_id')::uuid, TG_TABLE_NAME, NEW.id, NEW.consent_snapshot);
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER scans_authorization_record BEFORE INSERT OR UPDATE ON scans
FOR EACH ROW EXECUTE FUNCTION record_scan_authorization();
CREATE TRIGGER schedules_authorization_record BEFORE INSERT OR UPDATE ON schedules
FOR EACH ROW EXECUTE FUNCTION record_scan_authorization();
