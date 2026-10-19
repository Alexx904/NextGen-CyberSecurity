CREATE UNIQUE INDEX idx_target_groups_tenant_identity ON target_groups(tenant_id,id);

CREATE TABLE alert_channels (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    company_id UUID,
    name VARCHAR(160) NOT NULL CHECK (length(trim(name)) > 0),
    kind TEXT NOT NULL CHECK (kind IN ('slack','teams','telegram','webhook','smtp')),
    enabled BOOLEAN NOT NULL DEFAULT false,
    events TEXT[] NOT NULL CHECK (cardinality(events) BETWEEN 1 AND 7 AND events <@ ARRAY['finding.high','finding.critical','host.discovered','host.changed','port.discovered','scan.completed','scan.failed']::TEXT[]),
    encrypted_config TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tenant_id,id),
    FOREIGN KEY (tenant_id,company_id) REFERENCES target_groups(tenant_id,id) ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED
);
CREATE INDEX idx_alert_channels_tenant ON alert_channels(tenant_id,created_at,id) WHERE deleted_at IS NULL;

CREATE FUNCTION alert_validate_company_scope() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  IF NEW.company_id IS NULL THEN RETURN NEW; END IF;
  PERFORM pg_advisory_xact_lock(hashtextextended(NEW.tenant_id::TEXT,0));
  IF NOT EXISTS(SELECT 1 FROM target_groups WHERE tenant_id=NEW.tenant_id AND id=NEW.company_id AND parent_id IS NULL) THEN
    RAISE EXCEPTION 'Alert channel scope must be a root company of its tenant' USING ERRCODE='23514',CONSTRAINT='alert_company_scope';
  END IF;
  RETURN NEW;
END;
$$;
CREATE TRIGGER alert_channel_company BEFORE INSERT OR UPDATE OF company_id ON alert_channels FOR EACH ROW EXECUTE FUNCTION alert_validate_company_scope();

CREATE FUNCTION alert_preserve_company_scope() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  IF NEW.parent_id IS NOT DISTINCT FROM OLD.parent_id OR NEW.parent_id IS NULL THEN RETURN NEW; END IF;
  PERFORM pg_advisory_xact_lock(hashtextextended(NEW.tenant_id::TEXT,0));
  IF EXISTS(SELECT 1 FROM alert_channels WHERE tenant_id=NEW.tenant_id AND company_id=NEW.id AND deleted_at IS NULL) THEN
    RAISE EXCEPTION 'Remove or reassign alert channels before reparenting company' USING ERRCODE='23514',CONSTRAINT='alert_company_scope';
  END IF;
  RETURN NEW;
END;
$$;
CREATE TRIGGER alert_company_identity BEFORE UPDATE OF parent_id ON target_groups FOR EACH ROW EXECUTE FUNCTION alert_preserve_company_scope();

CREATE TABLE alert_observations (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_key TEXT NOT NULL,
    signature JSONB NOT NULL,
    PRIMARY KEY(tenant_id,entity_key)
);

CREATE TABLE alert_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    channel_id UUID NOT NULL,
    event_id UUID NOT NULL,
    payload JSONB NOT NULL CHECK(jsonb_typeof(payload)='object'),
    state TEXT NOT NULL DEFAULT 'pending' CHECK(state IN ('pending','sending','sent','dead','cancelled')),
    attempts INTEGER NOT NULL DEFAULT 0 CHECK(attempts BETWEEN 0 AND 6),
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    lease_until TIMESTAMPTZ,
    claim_token UUID,
    result_code VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(channel_id,event_id),
    FOREIGN KEY(tenant_id,channel_id) REFERENCES alert_channels(tenant_id,id) ON DELETE CASCADE
);
CREATE INDEX idx_alert_jobs_due ON alert_jobs(next_attempt_at,id) WHERE state IN ('pending','sending');

-- Existing history is a baseline, never retroactively sent to new channels.
INSERT INTO alert_observations(tenant_id,entity_key,signature)
SELECT DISTINCT tenant_id,'host:'||host_ip,'{}'::JSONB FROM findings;
INSERT INTO alert_observations(tenant_id,entity_key,signature)
SELECT DISTINCT tenant_id,'port:'||host_ip||':'||protocol||':'||port,'true'::JSONB FROM findings WHERE port>0 AND port_state='open';
INSERT INTO alert_observations(tenant_id,entity_key,signature)
SELECT tenant_id,'cve:'||host_ip||':'||protocol||':'||port||':'||cve_id,to_jsonb(CASE WHEN bool_or(severity='CRITICAL') THEN 'CRITICAL'::TEXT ELSE 'HIGH'::TEXT END)
FROM findings WHERE cve_id IS NOT NULL AND severity IN ('HIGH','CRITICAL') GROUP BY tenant_id,host_ip,protocol,port,cve_id;

CREATE FUNCTION alert_device_signature(device JSONB) RETURNS JSONB LANGUAGE SQL IMMUTABLE AS $$
    SELECT jsonb_build_object('mac',COALESCE(device->>'mac',''),'vendor',COALESCE(device->>'vendor',''),
      'hostname',COALESCE(device->>'hostname',''),'status',COALESCE(device->>'status',''),
      'ports',COALESCE((SELECT jsonb_agg(p ORDER BY p) FROM (SELECT DISTINCT value AS p FROM jsonb_array_elements(CASE WHEN jsonb_typeof(device->'open_ports')='array' THEN device->'open_ports' ELSE '[]'::JSONB END)) x),'[]'::JSONB));
$$;

INSERT INTO alert_observations(tenant_id,entity_key,signature)
SELECT DISTINCT ON (d.tenant_id,d.ip) d.tenant_id,'host:'||host(d.ip),alert_device_signature(d.device)
FROM discovery_devices d JOIN discovery_sessions s ON s.tenant_id=d.tenant_id AND s.id=d.session_id
ORDER BY d.tenant_id,d.ip,s.started_at DESC,s.id DESC
ON CONFLICT(tenant_id,entity_key) DO UPDATE SET signature=EXCLUDED.signature;

CREATE FUNCTION alert_record_observation(p_tenant UUID,p_key TEXT,p_signature JSONB,p_payload JSONB,p_company UUID) RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE previous JSONB; known BOOLEAN; event_id UUID; payload JSONB;
BEGIN
    -- Serialize per tenant so multi-host batches cannot take identity locks in
    -- opposing orders. Source transactions for different tenants remain separate.
    PERFORM pg_advisory_xact_lock(hashtextextended('nextgen-alert:'||p_tenant::TEXT,0));
    SELECT signature INTO previous FROM alert_observations WHERE tenant_id=p_tenant AND entity_key=p_key;
    known := FOUND;
    IF known AND p_payload->>'kind'='finding.high' AND previous='"CRITICAL"'::JSONB THEN RETURN; END IF;
    IF known AND previous=p_signature THEN RETURN; END IF;
    INSERT INTO alert_observations(tenant_id,entity_key,signature) VALUES(p_tenant,p_key,p_signature)
    ON CONFLICT(tenant_id,entity_key) DO UPDATE SET signature=EXCLUDED.signature;
    event_id:=gen_random_uuid();
    payload:=p_payload||jsonb_build_object('id',event_id,'tenant_id',p_tenant,'company_id',p_company,'occurred_at',NOW());
    IF NOT known AND payload->>'kind'='host.changed' THEN payload:=payload||jsonb_build_object('kind','host.discovered'); END IF;
    INSERT INTO alert_jobs(tenant_id,channel_id,event_id,payload)
    SELECT p_tenant,c.id,event_id,payload FROM alert_channels c
    WHERE c.tenant_id=p_tenant AND c.enabled AND c.deleted_at IS NULL
      AND (c.company_id IS NULL OR c.company_id=p_company) AND payload->>'kind'=ANY(c.events);
END;
$$;

CREATE FUNCTION alert_on_finding() RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE company UUID; payload JSONB;
BEGIN
    SELECT COALESCE(g.parent_id,g.id) INTO company FROM scans s
    JOIN targets t ON t.id=s.target_id AND t.tenant_id=s.tenant_id
    JOIN target_groups g ON g.id=t.group_id AND g.tenant_id=t.tenant_id
    WHERE s.id=NEW.scan_id AND s.tenant_id=NEW.tenant_id;
    payload:=jsonb_build_object('entity_id',NEW.id,'host_ip',NEW.host_ip,'port',NEW.port);
    -- Host discovery is represented separately from identity change snapshots.
    IF NOT EXISTS(SELECT 1 FROM alert_observations WHERE tenant_id=NEW.tenant_id AND entity_key='host:'||NEW.host_ip) THEN
      PERFORM alert_record_observation(NEW.tenant_id,'host:'||NEW.host_ip,'{}'::JSONB,payload||jsonb_build_object('kind','host.discovered'),company);
    END IF;
    IF NEW.port>0 AND NEW.port_state='open' THEN
      PERFORM alert_record_observation(NEW.tenant_id,'port:'||NEW.host_ip||':'||NEW.protocol||':'||NEW.port,'true'::JSONB,payload||jsonb_build_object('kind','port.discovered'),company);
    END IF;
    IF NEW.cve_id IS NOT NULL AND NEW.severity IN ('HIGH','CRITICAL') THEN
      PERFORM alert_record_observation(NEW.tenant_id,'cve:'||NEW.host_ip||':'||NEW.protocol||':'||NEW.port||':'||NEW.cve_id,to_jsonb(NEW.severity),payload||jsonb_build_object('kind','finding.'||lower(NEW.severity),'cve_id',NEW.cve_id,'severity',NEW.severity),company);
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER alert_finding_insert AFTER INSERT ON findings FOR EACH ROW EXECUTE FUNCTION alert_on_finding();

CREATE FUNCTION alert_on_device() RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE company UUID;
BEGIN
    -- Unknown/ambiguous company goes only to tenant channels, never to all companies.
    SELECT CASE WHEN count(DISTINCT COALESCE(g.parent_id,g.id))=1 THEN (array_agg(DISTINCT COALESCE(g.parent_id,g.id)))[1] ELSE NULL END INTO company
    FROM targets t JOIN target_groups g ON g.id=t.group_id AND g.tenant_id=t.tenant_id
    WHERE t.tenant_id=NEW.tenant_id AND t.target_type='IP' AND t.value=host(NEW.ip);
    PERFORM alert_record_observation(NEW.tenant_id,'host:'||host(NEW.ip),alert_device_signature(NEW.device),
      jsonb_build_object('kind','host.changed','entity_id',NEW.session_id,'host_ip',host(NEW.ip)),company);
    RETURN NEW;
END;
$$;
CREATE TRIGGER alert_device_change AFTER INSERT OR UPDATE ON discovery_devices FOR EACH ROW EXECUTE FUNCTION alert_on_device();

CREATE FUNCTION alert_on_scan() RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE company UUID;
BEGIN
    IF NEW.trigger_type<>'SCHEDULED' OR NEW.status NOT IN ('COMPLETED','FAILED') OR NEW.status=OLD.status THEN RETURN NEW; END IF;
    SELECT COALESCE(g.parent_id,g.id) INTO company FROM targets t JOIN target_groups g ON g.id=t.group_id AND g.tenant_id=t.tenant_id WHERE t.id=NEW.target_id AND t.tenant_id=NEW.tenant_id;
    PERFORM alert_record_observation(NEW.tenant_id,'scan:'||NEW.id||':'||NEW.status,'true'::JSONB,
      jsonb_build_object('kind',CASE WHEN NEW.status='COMPLETED' THEN 'scan.completed' ELSE 'scan.failed' END,'entity_id',NEW.id),company);
    RETURN NEW;
END;
$$;
CREATE TRIGGER alert_scan_terminal AFTER UPDATE OF status ON scans FOR EACH ROW EXECUTE FUNCTION alert_on_scan();
