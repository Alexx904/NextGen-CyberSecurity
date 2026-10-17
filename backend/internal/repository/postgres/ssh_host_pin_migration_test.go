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
	"nextgen-cybersec/internal/repository/migrations"
	"nextgen-cybersec/internal/repository/postgres"
)

func TestSSHHostPinMigrationRoundTrip(t *testing.T) {
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
	tenant, id := uuid.New(), uuid.New()
	if _, err := db.Pool.Exec(ctx, `INSERT INTO tenants (id,name,slug) VALUES ($1,'Host pin migration',$2)`, tenant, "pin-"+tenant.String()); err != nil {
		t.Fatal(err)
	}
	defer db.Pool.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, tenant)
	pin := "SHA256:" + strings.Repeat("A", 43)
	if _, err := db.Pool.Exec(ctx, `INSERT INTO ssh_servers (id,tenant_id,name,host,port,username,auth_type,encrypted_secret,client_name,public_key_fingerprint,host_key_fingerprint) VALUES ($1,$2,'Migration fixture','192.0.2.1',22,'test','PASSWORD','synthetic-ciphertext','Client label','client-fingerprint',$3)`, id, tenant, pin); err != nil {
		t.Fatal(err)
	}
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	dsn := strings.Replace(cfg.DB.DSN(), "postgres://", "pgx5://", 1)
	dsn = strings.Replace(dsn, "postgresql://", "pgx5://", 1)
	migrator, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer migrator.Close()
	if err := migrator.Migrate(9); err != nil {
		t.Fatal(err)
	}
	var exists bool
	if err := db.Pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='ssh_servers' AND column_name='host_key_fingerprint')`).Scan(&exists); err != nil || exists {
		t.Fatalf("down failed: exists=%v err=%v", exists, err)
	}
	var secret, client, clientFP string
	if err := db.Pool.QueryRow(ctx, `SELECT encrypted_secret,client_name,public_key_fingerprint FROM ssh_servers WHERE id=$1`, id).Scan(&secret, &client, &clientFP); err != nil {
		t.Fatal(err)
	}
	if secret != "synthetic-ciphertext" || client != "Client label" || clientFP != "client-fingerprint" {
		t.Fatal("down damaged inventory/client identity")
	}
	maxVersion, err := postgres.GetMaxSourceVersion()
	if err != nil {
		t.Fatal(err)
	}
	if err := migrator.Migrate(maxVersion); err != nil {
		t.Fatal(err)
	}
	var restored string
	if err := db.Pool.QueryRow(ctx, `SELECT host_key_fingerprint FROM ssh_servers WHERE id=$1`, id).Scan(&restored); err != nil || restored != "" {
		t.Fatalf("up inferred trust: pin=%q err=%v", restored, err)
	}
}
