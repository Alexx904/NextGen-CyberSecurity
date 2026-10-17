//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/repository/postgres"
)

func TestSSHServerRepo_Integration(t *testing.T) {
	cfg := getTestConfig(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatal("Integration test database is unavailable")
		return
	}
	defer db.Close()

	if err := postgres.RunMigrations(cfg.DB.DSN()); err != nil {
		t.Fatalf("Esecuzione migrazioni fallita: %v", err)
	}

	repo := postgres.NewSSHServerRepo(db.Pool)

	// Creazione tenant di test isolato
	tenantID := uuid.New()
	_, err = db.Pool.Exec(ctx, `INSERT INTO tenants (id, name, slug) VALUES ($1, 'SSH Test Tenant', $2) ON CONFLICT (slug) DO NOTHING`, tenantID, "ssh-test-"+tenantID.String()[:8])
	if err != nil {
		t.Fatalf("Creazione tenant fallita: %v", err)
	}

	serverID := uuid.New()
	server := &domain.SSHServer{
		ID:              serverID,
		TenantID:        tenantID,
		Name:            "Gateway DMZ 01",
		ClientName:      "Cliente autonomo",
		Host:            "10.0.0.1",
		Port:            22,
		Username:        "secops",
		AuthType:        domain.SSHAuthTypeKey,
		EncryptedSecret: "ciphertext-b64-mock-data",
		Status:          domain.SSHServerStatusUnchecked,
	}

	// 1. Create Server
	t.Run("Create and GetByID", func(t *testing.T) {
		err := repo.Create(ctx, server)
		if err != nil {
			t.Fatalf("Create server fallito: %v", err)
		}

		fetched, err := repo.GetByID(ctx, tenantID, serverID)
		if err != nil {
			t.Fatalf("GetByID fallito: %v", err)
		}

		if fetched.Name != server.Name {
			t.Errorf("atteso nome %q, ottenuto %q", server.Name, fetched.Name)
		}
		if fetched.ClientName != "Cliente autonomo" {
			t.Fatal("client label not persisted")
		}
		if fetched.EncryptedSecret != server.EncryptedSecret {
			t.Errorf("secret cifrato non corrisponde")
		}
		if fetched.Status != domain.SSHServerStatusUnchecked {
			t.Errorf("atteso status UNCHECKED, ottenuto: %s", fetched.Status)
		}

		// Isolamento tenant (NFR-SEC-03): accesso con altro tenantID deve restituire ErrSSHServerNotFound
		otherTenantID := uuid.New()
		_, err = repo.GetByID(ctx, otherTenantID, serverID)
		if !errors.Is(err, domain.ErrSSHServerNotFound) {
			t.Errorf("accesso con tenant errato doveva restituire ErrSSHServerNotFound, ottenuto: %v", err)
		}
	})

	// 2. Duplicate Name Conflict
	t.Run("Duplicate name conflict", func(t *testing.T) {
		dupServer := &domain.SSHServer{
			ID:              uuid.New(),
			TenantID:        tenantID,
			Name:            "Gateway DMZ 01", // stesso nome
			Host:            "10.0.0.2",
			Port:            2222,
			Username:        "root",
			AuthType:        domain.SSHAuthTypePassword,
			EncryptedSecret: "other-secret",
			Status:          domain.SSHServerStatusUnchecked,
		}

		err := repo.Create(ctx, dupServer)
		if !errors.Is(err, domain.ErrSSHServerDuplicateName) {
			t.Errorf("atteso ErrSSHServerDuplicateName, ottenuto: %v", err)
		}
	})

	// 3. List
	t.Run("List servers", func(t *testing.T) {
		list, err := repo.List(ctx, tenantID)
		if err != nil {
			t.Fatalf("List fallito: %v", err)
		}
		if len(list) == 0 {
			t.Fatalf("atteso almeno 1 server nella lista per tenant, ottenuti 0")
		}
	})

	// 4. Update and UpdateStatus
	t.Run("Update and UpdateStatus", func(t *testing.T) {
		server.Host = "10.0.0.99"
		server.Port = 2200
		server.ClientName = "Cliente aggiornato"
		err := repo.Update(ctx, server)
		if err != nil {
			t.Fatalf("Update fallito: %v", err)
		}

		now := time.Now().UTC()
		err = repo.UpdateStatus(ctx, tenantID, serverID, domain.SSHServerStatusOnline, &now)
		if err != nil {
			t.Fatalf("UpdateStatus fallito: %v", err)
		}

		updated, err := repo.GetByID(ctx, tenantID, serverID)
		if err != nil {
			t.Fatalf("GetByID dopo update fallito: %v", err)
		}
		if updated.Host != "10.0.0.99" || updated.Port != 2200 {
			t.Errorf("campi non aggiornati correttamente: %+v", updated)
		}
		if updated.ClientName != "Cliente aggiornato" {
			t.Fatal("client label not updated")
		}
		if updated.Status != domain.SSHServerStatusOnline {
			t.Errorf("status non aggiornato a ONLINE: %s", updated.Status)
		}
	})

	// 5. Firewall Backups Operations
	t.Run("Firewall Backups CRUD", func(t *testing.T) {
		backup1 := &domain.FirewallBackup{
			ID:               uuid.New(),
			TenantID:         tenantID,
			ServerID:         serverID,
			RawIptablesRules: "*filter\n:INPUT DROP [0:0]\nCOMMIT",
			Label:            "Baseline policy snapshot",
			CreatedAt:        time.Now().Add(-1 * time.Minute).UTC(),
		}

		backup2 := &domain.FirewallBackup{
			ID:               uuid.New(),
			TenantID:         tenantID,
			ServerID:         serverID,
			RawIptablesRules: "*filter\n:INPUT ACCEPT [0:0]\nCOMMIT",
			Label:            "Rollout safety test snapshot",
			CreatedAt:        time.Now().UTC(),
		}

		if err := repo.CreateBackup(ctx, backup1); err != nil {
			t.Fatalf("CreateBackup 1 fallito: %v", err)
		}
		if err := repo.CreateBackup(ctx, backup2); err != nil {
			t.Fatalf("CreateBackup 2 fallito: %v", err)
		}

		// GetLatestBackup deve restituire backup2
		latest, err := repo.GetLatestBackup(ctx, tenantID, serverID)
		if err != nil {
			t.Fatalf("GetLatestBackup fallito: %v", err)
		}
		if latest.ID != backup2.ID {
			t.Errorf("atteso backup più recente ID %s, ottenuto %s", backup2.ID, latest.ID)
		}

		// ListBackups
		backups, err := repo.ListBackups(ctx, tenantID, serverID)
		if err != nil {
			t.Fatalf("ListBackups fallito: %v", err)
		}
		if len(backups) < 2 {
			t.Errorf("attesi almeno 2 backup, ottenuti %d", len(backups))
		}

		// DeleteBackup
		if err := repo.DeleteBackup(ctx, tenantID, backup1.ID); err != nil {
			t.Fatalf("DeleteBackup fallito: %v", err)
		}

		// Verifica eliminazione
		_, err = repo.GetBackupByID(ctx, tenantID, backup1.ID)
		if !errors.Is(err, domain.ErrFirewallBackupNotFound) {
			t.Errorf("atteso ErrFirewallBackupNotFound dopo cancellazione, ottenuto: %v", err)
		}
	})

	// 6. Delete Server (e verifica cascade)
	t.Run("Delete server", func(t *testing.T) {
		err := repo.Delete(ctx, tenantID, serverID)
		if err != nil {
			t.Fatalf("Delete server fallito: %v", err)
		}

		_, err = repo.GetByID(ctx, tenantID, serverID)
		if !errors.Is(err, domain.ErrSSHServerNotFound) {
			t.Errorf("atteso ErrSSHServerNotFound dopo eliminazione, ottenuto: %v", err)
		}
	})
}
