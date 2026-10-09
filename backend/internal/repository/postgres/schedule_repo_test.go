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

func TestScheduleRepo_Integration(t *testing.T) {
	cfg := getTestConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Skipf("Database non raggiungibile per i test di integrazione (%v): salto il test", err)
		return
	}
	defer db.Close()

	if err := postgres.RunMigrations(cfg.DB.DSN()); err != nil {
		t.Fatalf("Esecuzione migrazioni fallita: %v", err)
	}

	scheduleRepo := postgres.NewScheduleRepo(db.Pool)
	targetRepo := postgres.NewTargetRepo(db.Pool)
	profileRepo := postgres.NewScanProfileRepo(db.Pool)

	// Recupero di un profilo esistente
	profiles, err := profileRepo.List(ctx, nil)
	if err != nil || len(profiles) == 0 {
		t.Fatalf("Nessun profilo di scansione disponibile per il test")
	}
	profile := profiles[0]

	// Creazione di un tenant fittizio o recupero di quello esistente
	tenantID := uuid.New()
	_, err = db.Pool.Exec(ctx, `INSERT INTO tenants (id, name, slug) VALUES ($1, 'Schedule Test Tenant', $2) ON CONFLICT (slug) DO NOTHING`, tenantID, "sched-test-"+tenantID.String()[:8])
	if err != nil {
		t.Fatalf("Creazione tenant di test fallita: %v", err)
	}

	// Creazione di un target di prova
	target := &domain.Target{
		ID:       uuid.New(),
		TenantID: tenantID,
		Type:     domain.TargetTypeIP,
		Value:    "192.168.10.50",
	}
	if err := targetRepo.Create(ctx, target); err != nil {
		t.Fatalf("Creazione target per test schedule fallita: %v", err)
	}

	scheduleID := uuid.New()
	cronExpr := "0 2 * * *"
	now := time.Now().UTC()
	nextRun := now.Add(24 * time.Hour)

	schedule := &domain.Schedule{
		ID:             scheduleID,
		TenantID:       tenantID,
		TargetID:       target.ID,
		ProfileID:      profile.ID,
		CronExpression: cronExpr,
		IsActive:       true,
		NextRunAt:      &nextRun,
	}

	// 1. Create
	t.Run("Create", func(t *testing.T) {
		err := scheduleRepo.Create(ctx, schedule)
		if err != nil {
			t.Fatalf("Create schedule fallito: %v", err)
		}
	})

	// 2. GetByID
	t.Run("GetByID", func(t *testing.T) {
		fetched, err := scheduleRepo.GetByID(ctx, tenantID, scheduleID)
		if err != nil {
			t.Fatalf("GetByID fallito: %v", err)
		}
		if fetched.CronExpression != cronExpr {
			t.Errorf("CronExpression attesa %s, ottenuta %s", cronExpr, fetched.CronExpression)
		}
		if !fetched.IsActive {
			t.Error("Atteso is_active = true")
		}

		// Not found
		_, err = scheduleRepo.GetByID(ctx, tenantID, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("Atteso ErrNotFound per schedule inesistente, ottenuto: %v", err)
		}
	})

	// 3. ListByTenant & ListActive
	t.Run("ListByTenant_And_ListActive", func(t *testing.T) {
		list, err := scheduleRepo.ListByTenant(ctx, tenantID)
		if err != nil {
			t.Fatalf("ListByTenant fallito: %v", err)
		}
		if len(list) == 0 {
			t.Fatal("Attesa almeno 1 pianificazione per il tenant")
		}

		activeList, err := scheduleRepo.ListActive(ctx)
		if err != nil {
			t.Fatalf("ListActive fallito: %v", err)
		}
		found := false
		for _, s := range activeList {
			if s.ID == scheduleID {
				found = true
				break
			}
		}
		if !found {
			t.Error("La pianificazione creata non è presente in ListActive")
		}
	})

	// 4. Update
	t.Run("Update", func(t *testing.T) {
		schedule.IsActive = false
		newCron := "0 4 * * *"
		schedule.CronExpression = newCron
		lastRun := time.Now().UTC()
		schedule.LastRunAt = &lastRun

		err := scheduleRepo.Update(ctx, schedule)
		if err != nil {
			t.Fatalf("Update fallito: %v", err)
		}

		updated, err := scheduleRepo.GetByID(ctx, tenantID, scheduleID)
		if err != nil {
			t.Fatalf("GetByID post-update fallito: %v", err)
		}
		if updated.IsActive {
			t.Error("Atteso is_active = false dopo update")
		}
		if updated.CronExpression != newCron {
			t.Errorf("CronExpression attesa %s, ottenuta %s", newCron, updated.CronExpression)
		}
	})

	// 5. Delete
	t.Run("Delete", func(t *testing.T) {
		err := scheduleRepo.Delete(ctx, tenantID, scheduleID)
		if err != nil {
			t.Fatalf("Delete fallito: %v", err)
		}

		_, err = scheduleRepo.GetByID(ctx, tenantID, scheduleID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("Atteso ErrNotFound dopo cancellazione, ottenuto: %v", err)
		}
	})
}
