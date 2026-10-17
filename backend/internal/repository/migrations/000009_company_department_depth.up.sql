-- New writes enforce Company -> Department. Existing historical trees are not rewritten.
CREATE FUNCTION enforce_company_department_depth() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'UPDATE' AND NEW.parent_id IS NOT DISTINCT FROM OLD.parent_id THEN
        RETURN NEW;
    END IF;
    IF NEW.parent_id IS NULL THEN RETURN NEW; END IF;
    -- Serialize parent changes and child creation within the tenant.
    PERFORM pg_advisory_xact_lock(hashtextextended(NEW.tenant_id::text, 0));
    IF NOT EXISTS (
        SELECT 1 FROM target_groups p
        WHERE p.id = NEW.parent_id AND p.tenant_id = NEW.tenant_id AND p.parent_id IS NULL
    ) OR EXISTS (
        SELECT 1 FROM target_groups c WHERE c.parent_id = NEW.id AND c.tenant_id = NEW.tenant_id
    ) OR NEW.id = NEW.parent_id THEN
        RAISE EXCEPTION 'Only Company -> Department hierarchy is allowed'
            USING ERRCODE = '23514', CONSTRAINT = 'company_department_depth';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER company_department_depth
BEFORE INSERT OR UPDATE OF parent_id ON target_groups
FOR EACH ROW EXECUTE FUNCTION enforce_company_department_depth();
