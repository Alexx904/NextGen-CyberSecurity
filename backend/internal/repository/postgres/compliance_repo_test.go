//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/repository/migrations"
	"nextgen-cybersec/internal/repository/postgres"
)

func TestCompliancePersistenceScopeLifecycleConcurrencyAndRollback(t *testing.T) {
	cfg := getTestConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	if err := postgres.RunMigrations(cfg.DB.DSN()); err != nil {
		t.Fatal(err)
	}
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := db.Pool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	tenant, other, root, dept, foreign, target, actor := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	exec(`INSERT INTO tenants(id,name,slug) VALUES($1,'M4 tenant',$2),($3,'Other tenant',$4)`, tenant, "m4-"+tenant.String(), other, "m4-"+other.String())
	exec(`INSERT INTO users(id,tenant_id,email,password_hash,full_name,role) VALUES($1,$2,$3,'test-hash','M4 Operator','SecOps')`, actor, tenant, actor.String()+"@test.invalid")
	exec(`INSERT INTO target_groups(id,tenant_id,name) VALUES($1,$2,'M4 root'),($3,$4,'Foreign root')`, root, tenant, foreign, other)
	exec(`INSERT INTO target_groups(id,tenant_id,name,parent_id) VALUES($1,$2,'M4 department',$3)`, dept, tenant, root)
	exec(`INSERT INTO targets(id,tenant_id,group_id,target_type,value) VALUES($1,$2,$3,'IP','192.0.2.10')`, target, tenant, dept)
	var profile uuid.UUID
	if err = db.Pool.QueryRow(ctx, `SELECT id FROM scan_profiles ORDER BY name LIMIT 1`).Scan(&profile); err != nil {
		t.Fatal(err)
	}
	cve := "CVE-2026-54321"
	exec(`INSERT INTO cve_entries(cve_id,severity) VALUES($1,'HIGH') ON CONFLICT DO NOTHING`, cve)
	insertFinding := func(at time.Time) uuid.UUID {
		t.Helper()
		scan, finding := uuid.New(), uuid.New()
		exec(`INSERT INTO scans(id,tenant_id,target_id,profile_id,status,created_at) VALUES($1,$2,$3,$4,'COMPLETED',$5)`, scan, tenant, target, profile, at)
		exec(`INSERT INTO findings(id,scan_id,tenant_id,host_ip,port,protocol,cve_id,severity,created_at) VALUES($1,$2,$3,'192.0.2.10',443,'tcp',$4,'HIGH',$5)`, finding, scan, tenant, cve, at)
		return finding
	}
	first := insertFinding(time.Now().Add(-48 * time.Hour).UTC())
	repo := postgres.NewComplianceRepo(db.Pool)
	asset, err := repo.CreateAsset(ctx, tenant, domain.AssessmentAsset{GroupID: dept, Name: "Verified asset identity"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = repo.CreateAsset(ctx, tenant, domain.AssessmentAsset{GroupID: foreign, Name: "cross tenant"}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("cross-tenant group accepted", err)
	}
	if _, err = repo.CreateAsset(ctx, tenant, domain.AssessmentAsset{GroupID: dept, Name: asset.Name}); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("duplicate identity label accepted", err)
	}
	id, err := repo.ObserveFinding(ctx, tenant, actor, asset.ID, first)
	if err != nil {
		t.Fatal(err)
	}
	if replay, err := repo.ObserveFinding(ctx, tenant, actor, asset.ID, first); err != nil || replay != id {
		t.Fatal("finding replay not idempotent", err)
	}
	exec(`UPDATE findings SET host_ip='192.0.2.11' WHERE id=$1`, first)
	if _, err = repo.ObserveFinding(ctx, tenant, actor, asset.ID, first); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("changed source reused immutable observation", err)
	}
	exec(`UPDATE findings SET host_ip='192.0.2.10' WHERE id=$1`, first)
	secondAsset, err := repo.CreateAsset(ctx, tenant, domain.AssessmentAsset{GroupID: dept, Name: "Different asset"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = repo.ObserveFinding(ctx, tenant, actor, secondAsset.ID, first); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("finding rebound to another identity", err)
	}
	if _, err = repo.ObserveFinding(ctx, other, actor, asset.ID, first); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("cross tenant observation", err)
	}
	if _, err = repo.Snapshot(ctx, other, domain.ComplianceScope{CompanyID: root}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("cross tenant snapshot", err)
	}
	if _, err = repo.Snapshot(ctx, tenant, domain.ComplianceScope{CompanyID: root, DepartmentID: foreign}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("foreign department", err)
	}
	command := domain.IssueCommand{Action: "configure", Actor: actor, ExpectedVersion: 1, Reason: "explicit SLA", Owner: actor, SLAHours: 24, PolicyID: "internal-high", PolicyVersion: "1"}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- repo.AppendCommand(ctx, tenant, id, command) }()
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, domain.ErrConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatal("lost-update protection", success, conflict)
	}
	scope := domain.ComplianceScope{CompanyID: root}
	snapshot, err := repo.Snapshot(ctx, tenant, scope)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Issues) != 1 || snapshot.Issues[0].Version != 2 || !snapshot.Issues[0].Breached || snapshot.Findings[0].LinkedIssueID == nil {
		t.Fatal("persistence round trip", snapshot)
	}
	if snapshot.Scans[0].ProfileID != profile || snapshot.Scans[0].Authorization != nil {
		t.Fatal("legacy scan provenance misrepresented")
	}
	if err = repo.AppendCommand(ctx, tenant, id, domain.IssueCommand{Action: "correction_declared", Actor: actor, ExpectedVersion: 2, Reason: "patch applied"}); err != nil {
		t.Fatal(err)
	}
	verify := domain.IssueCommand{Action: "verify", Actor: actor, ExpectedVersion: 3, Reason: "reviewed retest", Verification: &domain.IssueVerification{CompleteCoverage: false, NotPresent: true, Provenance: domain.IssueProvenance{Source: "manual retest", Version: "1", Reference: "test/proof"}}}
	if err = repo.AppendCommand(ctx, tenant, id, verify); !errors.Is(err, domain.ErrValidation) {
		t.Fatal("incomplete coverage accepted", err)
	}
	verify.Verification.CompleteCoverage = true
	if err = repo.AppendCommand(ctx, tenant, id, verify); err != nil {
		t.Fatal(err)
	}
	older := insertFinding(time.Now().Add(-time.Hour))
	if _, err = repo.ObserveFinding(ctx, tenant, actor, asset.ID, older); !errors.Is(err, domain.ErrValidation) {
		t.Fatal("historical evidence reopened current issue", err)
	}
	newer := insertFinding(time.Now().UTC())
	exec(`UPDATE findings SET host_ip='192.0.2.99' WHERE id=$1`, newer)
	if _, err = repo.ObserveFinding(ctx, tenant, actor, asset.ID, newer); err != nil {
		t.Fatal(err)
	}
	snapshot, err = repo.Snapshot(ctx, tenant, scope)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Issues[0].State != domain.IssueOpen || snapshot.Issues[0].Version != 5 || snapshot.CompletedCycles != 1 || snapshot.Issues[0].MTTRHours != nil {
		t.Fatal("recurrence/history/MTTR", snapshot.Issues)
	}
	if snapshot.Issues[0].Events[4].Observation.SourceHost != "192.0.2.99" {
		t.Fatal("address change lost stable identity/source")
	}
	if err = postgres.NewTargetGroupRepo(db.Pool).Delete(ctx, tenant, dept); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("referenced audit scope deleted", err)
	}
	exec(`DELETE FROM findings WHERE id=$1`, first)
	snapshot, err = repo.Snapshot(ctx, tenant, scope)
	if err != nil || snapshot.Issues[0].Events[0].Observation.SourceHost != "192.0.2.10" {
		t.Fatal("source deletion erased observation snapshot", err)
	}
	e := domain.ControlEvidence{GroupID: dept, ControlID: "ISO-A.8.8", Status: "satisfied", Actor: actor, Reason: "manual claim", Reference: "document/1", SourceVersion: "1"}
	if err = repo.AppendEvidence(ctx, tenant, e); err != nil {
		t.Fatal(err)
	}
	snapshot, err = repo.Snapshot(ctx, tenant, scope)
	if err != nil {
		t.Fatal(err)
	}
	for _, card := range snapshot.Scorecards {
		if card.GroupID == root.String() && card.Framework == "ISO27001" && card.Controls[0].Status != "not_satisfied" {
			t.Fatal("manual satisfaction hid open gap")
		}
	}
	for _, query := range []string{`UPDATE assessment_issue_events SET payload=payload`, `DELETE FROM assessment_issue_events`, `TRUNCATE assessment_issue_events`, `UPDATE assessment_control_evidence SET payload=payload`, `DELETE FROM assessment_control_evidence`, `TRUNCATE assessment_control_evidence`} {
		if _, err = db.Pool.Exec(ctx, query); err == nil {
			t.Fatal("audit mutation accepted", query)
		}
	}
	// Dedicated M4 DB only: prove the paired migration restores schema, not data.
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", source, strings.Replace(cfg.DB.DSN(), "postgres://", "pgx5://", 1))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err = m.Migrate(13); err != nil {
		t.Fatal(err)
	}
	var exists bool
	if err = db.Pool.QueryRow(ctx, `SELECT to_regclass('public.assessment_issues') IS NOT NULL`).Scan(&exists); err != nil || exists {
		t.Fatal("down failed", err)
	}
	if err = m.Migrate(14); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = db.Pool.QueryRow(ctx, `SELECT count(*) FROM assessment_issue_events`).Scan(&count); err != nil || count != 0 {
		t.Fatal("up fabricated audit", err)
	}
}
