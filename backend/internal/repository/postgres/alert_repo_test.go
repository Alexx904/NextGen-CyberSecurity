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

func TestAlertRepositoryAtomicScopeAndLeases(t *testing.T) {
	cfg := getTestConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	if err := postgres.RunMigrations(cfg.DB.DSN()); err != nil {
		t.Fatal(err)
	}
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatal("dedicated test DB unavailable")
	}
	defer db.Close()
	tenant, other, company, otherCompany, profile, target, scan := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	cve := "CVE-2099-" + uuid.NewString()[:8]
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := db.Pool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO tenants(id,name,slug) VALUES($1::UUID,'Alert fixture',$1::UUID::TEXT),($2::UUID,'Other fixture',$2::UUID::TEXT)`, tenant, other)
	defer func() {
		cleanupCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if _, err := db.Pool.Exec(cleanupCtx, `DELETE FROM tenants WHERE id=$1 OR id=$2`, tenant, other); err != nil {
			t.Error("fixture cleanup", err)
		}
		db.Pool.Exec(cleanupCtx, `DELETE FROM scan_profiles WHERE id=$1`, profile)
		db.Pool.Exec(cleanupCtx, `DELETE FROM cve_entries WHERE cve_id=$1`, cve)
	}()
	exec(`INSERT INTO target_groups(id,tenant_id,name) VALUES($1,$3,'Alert company'),($2,$3,'Other company')`, company, otherCompany, tenant)
	exec(`INSERT INTO targets(id,tenant_id,group_id,target_type,value) VALUES($1,$2,$3,'IP','192.0.2.44')`, target, tenant, company)
	exec(`INSERT INTO scan_profiles(id,name,nmap_args) VALUES($1::UUID,$1::UUID::TEXT,'-sT')`, profile)
	exec(`INSERT INTO scans(id,tenant_id,target_id,profile_id,trigger_type) VALUES($1,$2,$3,$4,'SCHEDULED')`, scan, tenant, target, profile)
	exec(`INSERT INTO cve_entries(cve_id,severity) VALUES($1,'CRITICAL')`, cve)
	repo := postgres.NewAlertRepo(db.Pool)
	events := []string{"finding.high", "finding.critical", "host.discovered", "host.changed", "port.discovered", "scan.completed", "scan.failed"}
	makeChannel := func(owner uuid.UUID, scope *uuid.UUID) *domain.AlertChannel {
		t.Helper()
		c := &domain.AlertChannel{ID: uuid.New(), TenantID: owner, CompanyID: scope, Name: "QA only", Kind: domain.AlertWebhook, Enabled: true, Events: events, EncryptedConfig: "test-only-ciphertext"}
		if err := repo.SaveChannel(ctx, c, true); err != nil {
			t.Fatal(err)
		}
		return c
	}
	global, scoped, unrelated, foreign := makeChannel(tenant, nil), makeChannel(tenant, &company), makeChannel(tenant, &otherCompany), makeChannel(other, nil)
	if _, err := db.Pool.Exec(ctx, `UPDATE target_groups SET parent_id=$2 WHERE id=$1 AND tenant_id=$3`, company, otherCompany, tenant); err == nil {
		t.Fatal("company with channels silently reparented")
	}
	if _, err := db.Pool.Exec(ctx, `DELETE FROM target_groups WHERE id=$1 AND tenant_id=$2`, company, tenant); err == nil {
		t.Fatal("company with channels silently removed")
	}
	if err := postgres.NewTargetGroupRepo(db.Pool).Delete(ctx, tenant, company); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("company delete did not report conflict", err)
	}
	if _, err := repo.GetChannel(ctx, other, global.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("cross-tenant channel read")
	}
	bad := &domain.AlertChannel{ID: uuid.New(), TenantID: other, CompanyID: &company, Name: "Foreign", Kind: domain.AlertWebhook, Events: events, EncryptedConfig: "fixture"}
	if err := repo.SaveChannel(ctx, bad, true); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("cross-tenant company accepted", err)
	}
	count := func(channel uuid.UUID) int {
		t.Helper()
		var n int
		if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM alert_jobs WHERE channel_id=$1`, channel).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	// The source row and its notification jobs must roll back together.
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO findings(scan_id,tenant_id,host_ip,port,protocol,cve_id,severity) VALUES($1,$2,'192.0.2.44',443,'tcp',$3,'CRITICAL')`, scan, tenant, cve); err != nil {
		t.Fatal(err)
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if count(global.ID) != 0 {
		t.Fatal("rolled-back source left jobs")
	}
	insert := func(severity string) {
		exec(`INSERT INTO findings(scan_id,tenant_id,host_ip,port,protocol,cve_id,severity) VALUES($1,$2,'192.0.2.44',443,'tcp',$3,$4)`, scan, tenant, cve, severity)
	}
	insert("HIGH")
	insert("HIGH")
	if count(global.ID) != 3 || count(scoped.ID) != 3 || count(unrelated.ID) != 0 || count(foreign.ID) != 0 {
		t.Fatal("dedup or company routing failed")
	}
	insert("CRITICAL")
	insert("HIGH")
	if count(global.ID) != 4 {
		t.Fatal("escalation duplicated or downgrade alerted")
	}
	// Two workers cannot own the same lease.
	var wg sync.WaitGroup
	var mu sync.Mutex
	jobs := []*domain.AlertJob{}
	errs := []error{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			job, err := repo.ClaimJob(ctx)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
			}
			if job != nil {
				jobs = append(jobs, job)
			}
		}()
	}
	wg.Wait()
	if len(errs) > 0 || len(jobs) != 8 {
		t.Fatalf("claims %d errors %v", len(jobs), errs)
	}
	seen := map[uuid.UUID]bool{}
	for i, job := range jobs {
		if seen[job.ID] || job.Event.TenantID != tenant || job.ClaimToken == uuid.Nil {
			t.Fatal("invalid duplicate claim")
		}
		seen[job.ID] = true
		stale := *job
		stale.ClaimToken = uuid.New()
		if err := repo.CompleteJob(ctx, &stale, "ACCEPTED", false, 0); !errors.Is(err, domain.ErrConflict) {
			t.Fatal("stale completion accepted")
		}
		code, retry, delay := "ACCEPTED", false, time.Duration(0)
		if i == 0 {
			code, retry, delay = "REMOTE_RETRY", true, time.Minute
		}
		if err := repo.CompleteJob(ctx, job, code, retry, delay); err != nil {
			t.Fatal(err)
		}
	}
	if job, err := repo.ClaimJob(ctx); err != nil || job != nil {
		t.Fatal("retry ignored next-attempt deadline")
	}
	exec(`UPDATE alert_jobs SET next_attempt_at=NOW()-INTERVAL '1 second' WHERE id=$1 AND tenant_id=$2`, jobs[0].ID, tenant)
	retry, err := repo.ClaimJob(ctx)
	if err != nil || retry == nil || retry.ID != jobs[0].ID || retry.Attempts != 2 {
		t.Fatal("durable retry did not recover", err)
	}
	if err := repo.CompleteJob(ctx, retry, "ACCEPTED", false, 0); err != nil {
		t.Fatal(err)
	}
	// Stable metadata updates (including latency/port ordering) are not changes.
	discovery := postgres.NewDiscoveryRepo(db.Pool)
	session := &domain.DiscoveryResult{ID: uuid.New(), TenantID: tenant, Subnet: "192.0.2.0/24", StartedAt: time.Now()}
	if err := discovery.Create(ctx, session); err != nil {
		t.Fatal(err)
	}
	device := domain.DiscoveredDevice{IP: "192.0.2.44", Hostname: "QA", Status: "up", OpenPorts: []int{443, 22}, ScannedAt: time.Now()}
	if err := discovery.SaveDevice(ctx, tenant, session.ID, device); err != nil {
		t.Fatal(err)
	}
	before := count(global.ID)
	device.LatencyMs = 99
	device.ScannedAt = time.Now()
	device.OpenPorts = []int{22, 443}
	if err := discovery.SaveDevice(ctx, tenant, session.ID, device); err != nil {
		t.Fatal(err)
	}
	if count(global.ID) != before {
		t.Fatal("latency/time/order created change alert")
	}
	device.Hostname = "QA changed"
	if err := discovery.SaveDevice(ctx, tenant, session.ID, device); err != nil {
		t.Fatal(err)
	}
	if count(global.ID) != before+1 {
		t.Fatal("identity change lost")
	}
	exec(`UPDATE scans SET status='COMPLETED' WHERE id=$1`, scan)
	if count(global.ID) != before+2 {
		t.Fatal("scheduled completion not queued")
	}
	old := *global
	global.Name = "Updated QA"
	global.Enabled = false
	if err := repo.SaveChannel(ctx, global, false); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveChannel(ctx, &old, false); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("stale channel overwrite accepted", err)
	}
	var pending int
	if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM alert_jobs WHERE channel_id=$1 AND state='pending'`, global.ID).Scan(&pending); err != nil || pending != 0 {
		t.Fatal("disable left old pending messages")
	}
	if err := repo.DeleteChannel(ctx, tenant, global.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetChannel(ctx, tenant, global.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("deleted channel readable")
	}
}

func TestAlertMigrationRoundTrip(t *testing.T) {
	cfg := getTestConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := postgres.RunMigrations(cfg.DB.DSN()); err != nil {
		t.Fatal(err)
	}
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatal("dedicated test DB unavailable")
	}
	defer db.Close()
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	dsn := strings.Replace(strings.Replace(cfg.DB.DSN(), "postgres://", "pgx5://", 1), "postgresql://", "pgx5://", 1)
	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err = m.Migrate(11); err != nil {
		t.Fatal(err)
	}
	var okay bool
	if err = db.Pool.QueryRow(ctx, `SELECT to_regclass('public.alert_channels') IS NULL AND to_regclass('public.alert_jobs') IS NULL AND to_regclass('public.discovery_devices') IS NOT NULL`).Scan(&okay); err != nil || !okay {
		t.Fatal("rollback changed unrelated schema", err)
	}
	if err = m.Migrate(12); err != nil {
		t.Fatal(err)
	}
	if err = db.Pool.QueryRow(ctx, `SELECT to_regclass('public.alert_channels') IS NOT NULL AND to_regclass('public.alert_jobs') IS NOT NULL AND EXISTS(SELECT 1 FROM pg_trigger WHERE tgname='alert_finding_insert')`).Scan(&okay); err != nil || !okay {
		t.Fatal("up did not restore alerting", err)
	}
}
