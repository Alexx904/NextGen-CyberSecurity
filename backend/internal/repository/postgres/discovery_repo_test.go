//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/repository/migrations"
	"nextgen-cybersec/internal/repository/postgres"
	"nextgen-cybersec/internal/scanner"
	"nextgen-cybersec/internal/service"
)

func TestDiscoveryRepositoryDurableSnapshots(t *testing.T) {
	cfg := getTestConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := postgres.RunMigrations(cfg.DB.DSN()); err != nil {
		t.Fatal(err)
	}
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatal("dedicated test database unavailable")
	}
	defer db.Close()
	tenant, other := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{tenant, other} {
		if _, err := db.Pool.Exec(ctx, `INSERT INTO tenants(id,name,slug) VALUES($1,'Discovery fixture',$2)`, id, "discovery-"+id.String()); err != nil {
			t.Fatal(err)
		}
		defer db.Pool.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, id)
	}
	repo := postgres.NewDiscoveryRepo(db.Pool)
	session := &domain.DiscoveryResult{ID: uuid.New(), TenantID: tenant, Subnet: "192.168.1.0/24", StartedAt: time.Now().UTC()}
	if err := repo.Create(ctx, session); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, session); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("duplicate session accepted")
	}
	foreign := *session
	foreign.TenantID = other
	if err := repo.Create(ctx, &foreign); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("session owner overwritten")
	}
	dev := domain.DiscoveredDevice{IP: "192.168.1.10", Hostname: "fixture", Status: "up", OpenPorts: []int{22, 443}, ScannedAt: time.Now().UTC()}
	if err := repo.SaveDevice(ctx, other, session.ID, dev); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("foreign tenant wrote device")
	}
	if err := repo.SaveDevice(ctx, tenant, session.ID, dev); err != nil {
		t.Fatal(err)
	}
	dev.Hostname = "updated fixture"
	if err := repo.SaveDevice(ctx, tenant, session.ID, dev); err != nil {
		t.Fatal(err)
	}
	// A fresh service has empty process maps: results must come from PostgreSQL.
	recovered := service.NewDiscoveryService(nil, nil, nil, nil, postgres.NewDiscoveryRepo(db.Pool))
	snapshot, err := recovered.GetSessionResults(ctx, tenant, session.ID)
	if err != nil || snapshot.TotalDiscovered != 1 || snapshot.Devices[0].Hostname != dev.Hostname || !snapshot.CompletedAt.IsZero() {
		t.Fatalf("partial recovery failed: snapshot=%+v err=%v", snapshot, err)
	}
	if _, err := recovered.GetSessionResults(ctx, other, session.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("foreign tenant read snapshot")
	}
	if _, err := recovered.GetLastResults(ctx, other); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("foreign tenant read latest snapshot")
	}
	snapshot.CompletedAt = time.Now().UTC()
	snapshot.DurationSeconds = 1.25
	if err := repo.Complete(ctx, snapshot); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveDevice(ctx, tenant, session.ID, dev); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("append after completion accepted")
	}
	if err := repo.Complete(ctx, snapshot); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("completed session rewritten")
	}
	newSession := &domain.DiscoveryResult{ID: uuid.New(), TenantID: tenant, Subnet: "192.168.2.0/24", StartedAt: time.Now().Add(time.Second)}
	if err := repo.Create(ctx, newSession); err != nil {
		t.Fatal(err)
	}
	latest, err := recovered.GetLastResults(ctx, tenant)
	if err != nil || latest.ID != newSession.ID || len(latest.Devices) != 0 {
		t.Fatalf("latest partial not returned: %v", err)
	}
	// Verify the composite FK rejects tenant/session mismatches even outside the repository.
	if _, err := db.Pool.Exec(ctx, `INSERT INTO discovery_devices(tenant_id,session_id,ip,device) VALUES($1,$2,'192.168.1.11','{}')`, other, session.ID); err == nil {
		t.Fatal("composite ownership FK missing")
	}
}

type discoveryFixtureRunner struct{}

func (discoveryFixtureRunner) RunScan(context.Context, string, scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
	return &scanner.ScanExecutionResult{RawXML: []byte(`<nmaprun><host><status state="up"/><address addr="192.168.1.10" addrtype="ipv4"/><hostnames><hostname name="fixture.local"/></hostnames></host><runstats><finished elapsed="1"/></runstats></nmaprun>`)}, nil
}

func TestDiscoveryServiceRecoveryWithPostgres(t *testing.T) {
	cfg := getTestConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := postgres.RunMigrations(cfg.DB.DSN()); err != nil {
		t.Fatal(err)
	}
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatal("dedicated test database unavailable")
	}
	defer db.Close()
	tenant := uuid.New()
	if _, err := db.Pool.Exec(ctx, `INSERT INTO tenants(id,name,slug) VALUES($1,'Discovery service fixture',$2)`, tenant, "recovery-"+tenant.String()); err != nil {
		t.Fatal(err)
	}
	defer db.Pool.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, tenant)
	svc := service.NewDiscoveryService(discoveryFixtureRunner{}, nil, nil, nil, postgres.NewDiscoveryRepo(db.Pool))
	result, err := svc.DiscoverSubnet(ctx, tenant, domain.DiscoveryRequest{Subnet: "192.168.1.0/24"})
	if err != nil {
		t.Fatal(err)
	}
	newService := service.NewDiscoveryService(nil, nil, nil, nil, postgres.NewDiscoveryRepo(db.Pool))
	restored, err := newService.GetLastResults(ctx, tenant)
	if err != nil || restored.ID != result.ID || restored.TotalDiscovered != 1 || restored.CompletedAt.IsZero() {
		t.Fatalf("completed recovery failed: result=%+v err=%v", restored, err)
	}
}

func TestDiscoveryMigrationRoundTrip(t *testing.T) {
	cfg := getTestConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := postgres.RunMigrations(cfg.DB.DSN()); err != nil {
		t.Fatal(err)
	}
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatal("dedicated test database unavailable")
	}
	defer db.Close()
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	dsn := strings.Replace(cfg.DB.DSN(), "postgres://", "pgx5://", 1)
	dsn = strings.Replace(dsn, "postgresql://", "pgx5://", 1)
	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.Migrate(10); err != nil {
		t.Fatal(err)
	}
	var removed bool
	if err := db.Pool.QueryRow(ctx, `SELECT to_regclass('public.discovery_sessions') IS NULL AND to_regclass('public.discovery_devices') IS NULL`).Scan(&removed); err != nil || !removed {
		t.Fatal("down did not remove discovery tables")
	}
	// The previous host pin schema must survive discovery rollback.
	var pins bool
	if err := db.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='ssh_servers' AND column_name='host_key_fingerprint')`).Scan(&pins); err != nil || !pins {
		t.Fatal("down changed unrelated schema")
	}
	max, err := postgres.GetMaxSourceVersion()
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Migrate(max); err != nil {
		t.Fatal(err)
	}
	var restored bool
	if err := db.Pool.QueryRow(ctx, `SELECT to_regclass('public.discovery_sessions') IS NOT NULL AND to_regclass('public.discovery_devices') IS NOT NULL`).Scan(&restored); err != nil || !restored {
		t.Fatal("up did not restore discovery tables")
	}
}
