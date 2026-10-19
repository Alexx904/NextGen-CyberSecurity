package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"nextgen-cybersec/internal/domain"
)

type ComplianceRepo struct{ pool *pgxpool.Pool }

func NewComplianceRepo(pool *pgxpool.Pool) *ComplianceRepo { return &ComplianceRepo{pool} }

func assessmentError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var sqlErr *pgconn.PgError
	if errors.As(err, &sqlErr) && (sqlErr.Code == "23505" || sqlErr.Code == "40001") {
		return domain.ErrConflict
	}
	return err
}

func (r *ComplianceRepo) CreateAsset(ctx context.Context, tenant uuid.UUID, asset domain.AssessmentAsset) (domain.AssessmentAsset, error) {
	asset.Name = strings.TrimSpace(asset.Name)
	if tenant == uuid.Nil || asset.GroupID == uuid.Nil || asset.Name == "" || len(asset.Name) > 200 {
		return asset, domain.ErrValidation
	}
	asset.ID = uuid.New()
	result, err := r.pool.Exec(ctx, `INSERT INTO assessment_assets(id,tenant_id,group_id,name) SELECT $1,$2,g.id,$4 FROM target_groups g WHERE g.id=$3 AND g.tenant_id=$2`, asset.ID, tenant, asset.GroupID, asset.Name)
	if err != nil {
		return asset, assessmentError(err)
	}
	if result.RowsAffected() == 0 {
		return asset, domain.ErrNotFound
	}
	return asset, nil
}

func loadComplianceIssue(ctx context.Context, tx pgx.Tx, tenant, id uuid.UUID, lock bool) (*domain.ComplianceIssue, error) {
	query := `SELECT i.id,i.asset_id,a.group_id,a.name,i.issue_key,i.port,i.protocol FROM assessment_issues i JOIN assessment_assets a ON (a.tenant_id,a.id)=(i.tenant_id,i.asset_id) WHERE i.tenant_id=$1 AND i.id=$2`
	if lock {
		query += " FOR UPDATE OF i"
	}
	var issue domain.ComplianceIssue
	if err := tx.QueryRow(ctx, query, tenant, id).Scan(&issue.ID, &issue.AssetID, &issue.GroupID, &issue.AssetName, &issue.Key, &issue.Port, &issue.Protocol); err != nil {
		return nil, assessmentError(err)
	}
	rows, err := tx.Query(ctx, `SELECT payload FROM assessment_issue_events WHERE tenant_id=$1 AND issue_id=$2 ORDER BY version`, tenant, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	issue.Events = []domain.IssueCommand{}
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var command domain.IssueCommand
		if err := json.Unmarshal(payload, &command); err != nil {
			return nil, err
		}
		issue.Events = append(issue.Events, command)
	}
	return &issue, rows.Err()
}

func appendIssueEvent(ctx context.Context, tx pgx.Tx, tenant uuid.UUID, issue *domain.ComplianceIssue, command domain.IssueCommand) error {
	if command.ExpectedVersion != len(issue.Events) {
		return domain.ErrConflict
	}
	issue.Events = append(issue.Events, command)
	if _, err := domain.ReplayIssue(tenant, issue, command.At); err != nil {
		return err
	}
	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}
	var findingID *uuid.UUID
	if command.Observation != nil {
		findingID = &command.Observation.FindingID
	}
	_, err = tx.Exec(ctx, `INSERT INTO assessment_issue_events(tenant_id,issue_id,version,payload,finding_id) VALUES($1,$2,$3,$4,$5)`, tenant, issue.ID, len(issue.Events), payload, findingID)
	return assessmentError(err)
}

func (r *ComplianceRepo) ObserveFinding(ctx context.Context, tenant, actor, assetID, findingID uuid.UUID) (uuid.UUID, error) {
	if tenant == uuid.Nil || actor == uuid.Nil || assetID == uuid.Nil || findingID == uuid.Nil {
		return uuid.Nil, domain.ErrValidation
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)
	// Serialize registration/observation for this asset before choosing identity.
	var groupID uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT group_id FROM assessment_assets WHERE tenant_id=$1 AND id=$2 FOR UPDATE`, tenant, assetID).Scan(&groupID); err != nil {
		return uuid.Nil, assessmentError(err)
	}
	var scanID uuid.UUID
	var at time.Time
	var port int
	var protocol, key string
	var host string
	var targetID uuid.UUID
	var confidence domain.FindingConfidence
	var method domain.DetectionMethod
	err = tx.QueryRow(ctx, `SELECT f.scan_id,f.created_at,f.port,f.protocol,COALESCE(NULLIF(f.cve_id,''),'service:'||COALESCE(NULLIF(f.service_name,''),'unknown')),f.host_ip,s.target_id,f.confidence,f.detection_method FROM findings f JOIN scans s ON s.id=f.scan_id AND s.tenant_id=f.tenant_id JOIN targets t ON t.id=s.target_id AND t.tenant_id=s.tenant_id WHERE f.id=$1 AND f.tenant_id=$2 AND t.group_id=$3 FOR SHARE OF f,s,t`, findingID, tenant, groupID).Scan(&scanID, &at, &port, &protocol, &key, &host, &targetID, &confidence, &method)
	if err != nil {
		return uuid.Nil, assessmentError(err)
	}
	identity, err := domain.NewIssueIdentity(tenant, assetID, protocol, port, key)
	if err != nil {
		return uuid.Nil, err
	}
	var existingIssue uuid.UUID
	var oldPayload []byte
	err = tx.QueryRow(ctx, `SELECT issue_id,payload FROM assessment_issue_events WHERE tenant_id=$1 AND finding_id=$2`, tenant, findingID).Scan(&existingIssue, &oldPayload)
	if err == nil {
		var existingAsset uuid.UUID
		if err = tx.QueryRow(ctx, `SELECT asset_id FROM assessment_issues WHERE tenant_id=$1 AND id=$2`, tenant, existingIssue).Scan(&existingAsset); err != nil {
			return uuid.Nil, err
		}
		if existingAsset != assetID {
			return uuid.Nil, domain.ErrConflict
		}
		var old domain.IssueCommand
		if err = json.Unmarshal(oldPayload, &old); err != nil {
			return uuid.Nil, err
		}
		if old.Observation == nil || old.Observation.Identity != identity || old.Observation.ScanID != scanID || !old.Observation.At.Equal(at) || old.Observation.SourceHost != host || old.Observation.SourceTargetID != targetID || old.Observation.SourceGroupID != groupID || old.Observation.Confidence != confidence || old.Observation.DetectionMethod != method {
			return uuid.Nil, domain.ErrConflict
		}
		return existingIssue, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, err
	}
	id := uuid.New()
	_, err = tx.Exec(ctx, `INSERT INTO assessment_issues(id,tenant_id,asset_id,protocol,port,issue_key) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(tenant_id,asset_id,protocol,port,issue_key) DO NOTHING`, id, tenant, assetID, identity.Protocol, identity.Port, identity.Key)
	if err != nil {
		return uuid.Nil, assessmentError(err)
	}
	if err = tx.QueryRow(ctx, `SELECT id FROM assessment_issues WHERE tenant_id=$1 AND asset_id=$2 AND protocol=$3 AND port=$4 AND issue_key=$5`, tenant, assetID, identity.Protocol, identity.Port, identity.Key).Scan(&id); err != nil {
		return uuid.Nil, err
	}
	issue, err := loadComplianceIssue(ctx, tx, tenant, id, true)
	if err != nil {
		return uuid.Nil, err
	}
	command := domain.IssueCommand{Action: "observe", Actor: actor, At: time.Now().UTC(), ExpectedVersion: len(issue.Events), Observation: &domain.IssueObservation{FindingID: findingID, ScanID: scanID, Identity: identity, At: at, SourceHost: host, SourceTargetID: targetID, SourceGroupID: groupID, Confidence: confidence, DetectionMethod: method, Provenance: domain.IssueProvenance{Source: "persisted_finding", Version: "finding-v2", Reference: findingID.String()}}}
	if err = appendIssueEvent(ctx, tx, tenant, issue, command); err != nil {
		return uuid.Nil, err
	}
	return id, tx.Commit(ctx)
}

func (r *ComplianceRepo) AppendCommand(ctx context.Context, tenant, id uuid.UUID, command domain.IssueCommand) error {
	if tenant == uuid.Nil || id == uuid.Nil || command.Actor == uuid.Nil || command.Action == "observe" || command.Observation != nil || len(command.Reason) > 4000 || len(command.PolicyID) > 100 || len(command.PolicyVersion) > 100 {
		return domain.ErrValidation
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	issue, err := loadComplianceIssue(ctx, tx, tenant, id, true)
	if err != nil {
		return err
	}
	if command.Owner != uuid.Nil {
		var exists bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE tenant_id=$1 AND id=$2 AND is_active)`, tenant, command.Owner).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return domain.ErrNotFound
		}
	}
	command.At = time.Now().UTC()
	if command.Verification != nil {
		command.Verification.Identity = domain.IssueIdentity{TenantID: tenant, AssetID: issue.AssetID, Protocol: issue.Protocol, Port: issue.Port, Key: issue.Key}
		command.Verification.At = command.At
		p := command.Verification.Provenance
		if len(p.Source) > 200 || len(p.Version) > 100 || len(p.Reference) > 2000 {
			return domain.ErrValidation
		}
	}
	if err = appendIssueEvent(ctx, tx, tenant, issue, command); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *ComplianceRepo) AppendEvidence(ctx context.Context, tenant uuid.UUID, evidence domain.ControlEvidence) error {
	if tenant == uuid.Nil || !evidence.Valid() {
		return domain.ErrValidation
	}
	evidence.ID, evidence.At = uuid.New(), time.Now().UTC()
	payload, err := json.Marshal(evidence)
	if err != nil {
		return err
	}
	result, err := r.pool.Exec(ctx, `INSERT INTO assessment_control_evidence(id,tenant_id,group_id,control_id,payload) SELECT $1,$2,g.id,$4,$5 FROM target_groups g WHERE g.tenant_id=$2 AND g.id=$3`, evidence.ID, tenant, evidence.GroupID, evidence.ControlID, payload)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ComplianceRepo) Snapshot(ctx context.Context, tenant uuid.UUID, scope domain.ComplianceScope) (*domain.ComplianceSnapshot, error) {
	if tenant == uuid.Nil || scope.CompanyID == uuid.Nil {
		return nil, domain.ErrValidation
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	s := &domain.ComplianceSnapshot{Scope: scope, Groups: []domain.ComplianceGroup{}, Assets: []domain.AssessmentAsset{}, Findings: []domain.ComplianceFinding{}, Scans: []domain.ComplianceScan{}, Issues: []domain.ComplianceIssue{}, Evidence: []domain.ControlEvidence{}}
	var rootExists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM target_groups WHERE tenant_id=$1 AND id=$2 AND parent_id IS NULL), clock_timestamp(), pg_current_snapshot()::text`, tenant, scope.CompanyID).Scan(&rootExists, &s.Cutoff, &s.SnapshotID); err != nil {
		return nil, err
	}
	if !rootExists {
		return nil, domain.ErrNotFound
	}
	s.Owners = []domain.ComplianceOwner{}
	ownerRows, ownerErr := tx.Query(ctx, `SELECT id,full_name FROM users WHERE tenant_id=$1 AND is_active ORDER BY full_name,id`, tenant)
	if ownerErr != nil {
		return nil, ownerErr
	}
	for ownerRows.Next() {
		var owner domain.ComplianceOwner
		if err = ownerRows.Scan(&owner.ID, &owner.Name); err != nil {
			ownerRows.Close()
			return nil, err
		}
		s.Owners = append(s.Owners, owner)
	}
	err = ownerRows.Err()
	ownerRows.Close()
	if err != nil {
		return nil, err
	}
	if scope.DepartmentID != uuid.Nil {
		var valid bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM target_groups WHERE tenant_id=$1 AND id=$2 AND parent_id=$3)`, tenant, scope.DepartmentID, scope.CompanyID).Scan(&valid); err != nil {
			return nil, err
		}
		if !valid {
			return nil, domain.ErrNotFound
		}
	}
	rows, err := tx.Query(ctx, `SELECT id,name,parent_id FROM target_groups WHERE tenant_id=$1 AND (id=$2 OR parent_id=$2) AND ($3::uuid='00000000-0000-0000-0000-000000000000' OR id=$3) ORDER BY parent_id NULLS FIRST,name,id`, tenant, scope.CompanyID, scope.DepartmentID)
	if err != nil {
		return nil, err
	}
	var groups []uuid.UUID
	for rows.Next() {
		var g domain.ComplianceGroup
		if err = rows.Scan(&g.ID, &g.Name, &g.ParentID); err != nil {
			rows.Close()
			return nil, err
		}
		s.Groups = append(s.Groups, g)
		groups = append(groups, g.ID)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	rows, err = tx.Query(ctx, `SELECT id,group_id,name FROM assessment_assets WHERE tenant_id=$1 AND group_id=ANY($2) ORDER BY name,id`, tenant, groups)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var a domain.AssessmentAsset
		if err = rows.Scan(&a.ID, &a.GroupID, &a.Name); err != nil {
			rows.Close()
			return nil, err
		}
		s.Assets = append(s.Assets, a)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	rows, err = tx.Query(ctx, `SELECT i.id FROM assessment_issues i JOIN assessment_assets a ON (a.tenant_id,a.id)=(i.tenant_id,i.asset_id) WHERE i.tenant_id=$1 AND a.group_id=ANY($2) ORDER BY i.created_at,i.id`, tenant, groups)
	if err != nil {
		return nil, err
	}
	var issueIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		issueIDs = append(issueIDs, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	linked := map[uuid.UUID]uuid.UUID{}
	for _, id := range issueIDs {
		issue, loadErr := loadComplianceIssue(ctx, tx, tenant, id, false)
		if loadErr != nil {
			return nil, loadErr
		}
		if _, err = domain.ReplayIssue(tenant, issue, s.Cutoff); err != nil {
			return nil, fmt.Errorf("issue replay: %w", err)
		}
		s.Issues = append(s.Issues, *issue)
		for _, event := range issue.Events {
			if event.Observation != nil {
				linked[event.Observation.FindingID] = id
			}
		}
	}
	rows, err = tx.Query(ctx, `SELECT e.payload FROM assessment_control_evidence e WHERE tenant_id=$1 AND group_id=ANY($2) ORDER BY recorded_at,id`, tenant, groups)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var payload []byte
		var e domain.ControlEvidence
		if err = rows.Scan(&payload); err != nil {
			rows.Close()
			return nil, err
		}
		if err = json.Unmarshal(payload, &e); err != nil {
			rows.Close()
			return nil, err
		}
		s.Evidence = append(s.Evidence, e)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	s.TargetCounts = map[string]int{}
	rows, err = tx.Query(ctx, `SELECT group_id,count(*) FROM targets WHERE tenant_id=$1 AND group_id=ANY($2) GROUP BY group_id`, tenant, groups)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id uuid.UUID
		var count int
		if err = rows.Scan(&id, &count); err != nil {
			rows.Close()
			return nil, err
		}
		s.TargetCounts[id.String()] = count
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	rows, err = tx.Query(ctx, `SELECT DISTINCT ON(t.id) s.id,t.id,t.group_id,s.status,s.created_at,s.profile_id,s.consent_snapshot FROM targets t JOIN scans s ON s.target_id=t.id AND s.tenant_id=t.tenant_id WHERE t.tenant_id=$1 AND t.group_id=ANY($2) ORDER BY t.id,s.created_at DESC,s.id DESC`, tenant, groups)
	if err != nil {
		return nil, err
	}
	var scanIDs []uuid.UUID
	for rows.Next() {
		var scan domain.ComplianceScan
		var authorization []byte
		if err = rows.Scan(&scan.ID, &scan.TargetID, &scan.GroupID, &scan.Status, &scan.At, &scan.ProfileID, &authorization); err != nil {
			rows.Close()
			return nil, err
		}
		if len(authorization) > 0 && string(authorization) != "null" {
			if err = json.Unmarshal(authorization, &scan.Authorization); err != nil {
				rows.Close()
				return nil, err
			}
		}
		s.Scans = append(s.Scans, scan)
		scanIDs = append(scanIDs, scan.ID)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	rows, err = tx.Query(ctx, `SELECT f.id,f.scan_id,f.tenant_id,f.host_ip,f.hostname,f.port,f.protocol,COALESCE(f.service_name,''),f.cve_id,f.severity,f.confidence,f.detection_method,f.created_at,s.target_id,t.group_id,s.status,f.port_state FROM findings f JOIN scans s ON s.id=f.scan_id AND s.tenant_id=f.tenant_id JOIN targets t ON t.id=s.target_id AND t.tenant_id=s.tenant_id WHERE f.tenant_id=$1 AND f.scan_id=ANY($2) ORDER BY f.created_at,f.id`, tenant, scanIDs)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var f domain.ComplianceFinding
		var service string
		if err = rows.Scan(&f.ID, &f.ScanID, &f.TenantID, &f.HostIP, &f.Hostname, &f.Port, &f.Protocol, &service, &f.CVEID, &f.Severity, &f.Confidence, &f.DetectionMethod, &f.CreatedAt, &f.TargetID, &f.GroupID, &f.ScanStatus, &f.PortState); err != nil {
			rows.Close()
			return nil, err
		}
		f.ServiceName = &service
		if id, ok := linked[f.ID]; ok {
			copy := id
			f.LinkedIssueID = &copy
		}
		s.Findings = append(s.Findings, f)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	domain.BuildComplianceScorecards(s)
	return s, nil
}
