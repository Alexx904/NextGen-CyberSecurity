//go:build integration

package postgres_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/repository/migrations"
	"nextgen-cybersec/internal/repository/postgres"
)

func TestScanAuthorizationAtomicAuditAndMigration(t *testing.T) {
	cfg := getTestConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := postgres.RunMigrations(cfg.DB.DSN()); err != nil {
		t.Fatal(err)
	}
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tenant, targetID := uuid.New(), uuid.New()
	if _, err := db.Pool.Exec(ctx, `INSERT INTO tenants(id,name,slug) VALUES($1,'Consent fixture',$2)`, tenant, "consent-"+tenant.String()); err != nil {
		t.Fatal(err)
	}
	defer db.Pool.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, tenant)
	target := &domain.Target{ID: targetID, TenantID: tenant, Type: domain.TargetTypeIP, Value: "192.168.1.20"}
	if err := postgres.NewTargetRepo(db.Pool).Create(ctx, target); err != nil {
		t.Fatal(err)
	}
	profiles, err := postgres.NewScanProfileRepo(db.Pool).List(ctx, nil)
	if err != nil || len(profiles) == 0 {
		t.Fatal("profile fixture unavailable", err)
	}
	profile := profiles[0]
	request := &domain.ScanAuthorizationRequest{Accepted: true, Statement: domain.ScanAuthorizationStatement, Targets: []domain.AuthorizedTarget{{ID: targetID, Value: target.Value}}, NmapArgs: profile.NmapArgs, TimeoutSeconds: profile.TimeoutSeconds}
	auth := domain.ScanAuthorizationFromContext(domain.WithScanAuthorization(ctx, tenant, uuid.New(), request))
	repo := postgres.NewScanRepo(db.Pool)
	scan := &domain.Scan{ID: uuid.New(), TenantID: tenant, TargetID: targetID, ProfileID: profile.ID, Authorization: auth}
	if err := repo.Create(ctx, scan); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.GetByID(ctx, tenant, scan.ID)
	if err != nil || loaded.Authorization == nil || loaded.Authorization.UserID != auth.UserID {
		t.Fatal("consent_snapshot round trip", err)
	}
	if _, err := repo.GetByID(ctx, uuid.New(), scan.ID); err != domain.ErrNotFound {
		t.Fatal("cross-tenant read", err)
	}
	var count int
	if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM scan_authorization_audit WHERE object_id=$1 AND tenant_id=$2 AND user_id=$3`, scan.ID, tenant, auth.UserID).Scan(&count); err != nil || count != 1 {
		t.Fatal("audit not atomic", count, err)
	}
	for _, query := range []string{`UPDATE scan_authorization_audit SET user_id=user_id WHERE object_id=$1`, `DELETE FROM scan_authorization_audit WHERE object_id=$1`, `UPDATE scans SET consent_snapshot='{}' WHERE id=$1`} {
		if _, err := db.Pool.Exec(ctx, query, scan.ID); err == nil {
			t.Fatal("evidence mutation allowed", query)
		}
	}
	if _, err := db.Pool.Exec(ctx, `TRUNCATE scan_authorization_audit`); err == nil {
		t.Fatal("audit truncation allowed")
	}
	// Invalid evidence must roll back both source and audit rows.
	bad := *auth
	bad.TenantID = uuid.New()
	failed := &domain.Scan{ID: uuid.New(), TenantID: tenant, TargetID: targetID, ProfileID: profile.ID, Authorization: &bad}
	if err := repo.Create(ctx, failed); err == nil {
		t.Fatal("invalid evidence accepted")
	}
	if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM scans WHERE id=$1`, failed.ID).Scan(&count); err != nil || count != 0 {
		t.Fatal("source write survived failed audit", err)
	}
	// Removing mutable inventory does not remove the consent_snapshot history.
	if _, err := db.Pool.Exec(ctx, `DELETE FROM scans WHERE id=$1`, scan.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM scan_authorization_audit WHERE object_id=$1`, scan.ID).Scan(&count); err != nil || count != 1 {
		t.Fatal("history lost", err)
	}
	scheduleID := uuid.New()
	recurring := *auth
	recurring.CronExpression = "0 2 * * *"
	recurring.ScheduleID = &scheduleID
	schedule := &domain.Schedule{ID: scheduleID, TenantID: tenant, TargetID: targetID, ProfileID: profile.ID, CronExpression: recurring.CronExpression, IsActive: true, Authorization: &recurring}
	schedules := postgres.NewScheduleRepo(db.Pool)
	if err := schedules.Create(ctx, schedule); err != nil {
		t.Fatal(err)
	}
	loadedSchedule, err := schedules.GetByID(ctx, tenant, scheduleID)
	if err != nil || loadedSchedule.Authorization == nil || loadedSchedule.Authorization.CronExpression != recurring.CronExpression {
		t.Fatal("schedule evidence round trip", err)
	}
	loadedSchedule.IsActive = false
	if err := schedules.Update(ctx, loadedSchedule); err != nil {
		t.Fatal(err)
	}
	// Simulate metadata from a job holding an older enabled copy.
	now := time.Now()
	schedule.LastRunAt = &now
	if err := schedules.UpdateRunTimes(ctx, schedule); err != nil {
		t.Fatal(err)
	}
	loadedSchedule, err = schedules.GetByID(ctx, tenant, scheduleID)
	if err != nil || loadedSchedule.IsActive || loadedSchedule.LastRunAt != nil {
		t.Fatal("runtime write resurrected disabled schedule", err)
	}
	if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM scan_authorization_audit WHERE object_id=$1`, scheduleID).Scan(&count); err != nil || count != 1 {
		t.Fatal("schedule evidence duplicated or absent", err)
	}
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	dsn := strings.Replace(strings.Replace(cfg.DB.DSN(), "postgres://", "pgx5://", 1), "postgresql://", "pgx5://", 1)
	migrator, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer migrator.Close()
	if err := migrator.Migrate(12); err != nil {
		t.Fatal(err)
	}
	var exists bool
	if err := db.Pool.QueryRow(ctx, `SELECT to_regclass('public.scan_authorization_audit') IS NOT NULL`).Scan(&exists); err != nil || exists {
		t.Fatal("down failed", err)
	}
	if err := migrator.Migrate(13); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM scan_authorization_audit`).Scan(&count); err != nil || count != 0 {
		t.Fatal("up fabricated consent", err)
	}
}
