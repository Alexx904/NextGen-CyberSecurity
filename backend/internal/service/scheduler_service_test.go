package service

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
)

type mockScheduleRepo struct {
	mu        sync.RWMutex
	schedules map[uuid.UUID]*domain.Schedule
}

func newMockScheduleRepo() *mockScheduleRepo {
	return &mockScheduleRepo{
		schedules: make(map[uuid.UUID]*domain.Schedule),
	}
}

func (m *mockScheduleRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Schedule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.schedules[id]
	if !ok || s.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *mockScheduleRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Schedule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.Schedule
	for _, s := range m.schedules {
		if s.TenantID == tenantID {
			cp := *s
			list = append(list, &cp)
		}
	}
	return list, nil
}

func (m *mockScheduleRepo) ListActive(ctx context.Context) ([]*domain.Schedule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.Schedule
	for _, s := range m.schedules {
		if s.IsActive {
			cp := *s
			list = append(list, &cp)
		}
	}
	return list, nil
}

func (m *mockScheduleRepo) Create(ctx context.Context, schedule *domain.Schedule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if schedule.ID == uuid.Nil {
		schedule.ID = uuid.New()
	}
	cp := *schedule
	m.schedules[schedule.ID] = &cp
	return nil
}

func (m *mockScheduleRepo) Update(ctx context.Context, schedule *domain.Schedule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.schedules[schedule.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *schedule
	m.schedules[schedule.ID] = &cp
	return nil
}

func (m *mockScheduleRepo) UpdateRunTimes(ctx context.Context, schedule *domain.Schedule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, ok := m.schedules[schedule.ID]
	if ok && current.IsActive && current.CronExpression == schedule.CronExpression {
		copy := *current
		copy.LastRunAt = schedule.LastRunAt
		copy.NextRunAt = schedule.NextRunAt
		m.schedules[schedule.ID] = &copy
	}
	return nil
}

func (m *mockScheduleRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.schedules[id]
	if !ok || s.TenantID != tenantID {
		return domain.ErrNotFound
	}
	delete(m.schedules, id)
	return nil
}

func TestSchedulerService_Lifecycle(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	scanRepo := newMockScanRepo()
	orchClient := &mockOrchestratorClient{}

	tenantID := uuid.New()
	targetID := uuid.New()
	profileID := uuid.New()

	schedRepo.schedules[uuid.New()] = &domain.Schedule{
		ID:             uuid.New(),
		TenantID:       tenantID,
		TargetID:       targetID,
		ProfileID:      profileID,
		CronExpression: "0 3 * * *",
		IsActive:       true,
	}

	scheduler := NewSchedulerService(schedRepo, scanRepo, orchClient, nil)

	ctx := context.Background()
	if err := scheduler.Start(ctx); err != nil {
		t.Fatalf("Start scheduler fallito: %v", err)
	}

	// Arresto dello scheduler
	scheduler.Stop()
}

func TestSchedulerService_TriggerJob_Success(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	scanRepo := newMockScanRepo()
	orchClient := &mockOrchestratorClient{}

	tenantID := uuid.New()
	targetID := uuid.New()
	profileID := uuid.New()
	schedID := uuid.New()

	schedule := &domain.Schedule{
		ID:             schedID,
		TenantID:       tenantID,
		TargetID:       targetID,
		ProfileID:      profileID,
		CronExpression: "0 2 * * *",
		IsActive:       true,
	}
	_ = schedRepo.Create(context.Background(), schedule)

	scheduler := NewSchedulerService(schedRepo, scanRepo, orchClient, nil)

	// Trigger esplicito del job
	authorizeScheduleFixture(scheduler, schedule)
	_ = schedRepo.Update(context.Background(), schedule)
	scheduler.TriggerJob(schedule)

	// Verifica che la scansione sia stata creata in scanRepo
	scans, total, err := scanRepo.List(context.Background(), tenantID, 10, 0)
	if err != nil || total != 1 {
		t.Fatalf("Attesa 1 scansione creata su scanRepo, trovate %d (err: %v)", total, err)
	}

	createdScan := scans[0]
	if createdScan.TriggerType != domain.TriggerScheduled {
		t.Errorf("Atteso TriggerScheduled, ottenuto %s", createdScan.TriggerType)
	}
	if createdScan.Status != domain.ScanStatusQueued {
		t.Errorf("Atteso stato QUEUED, ottenuto %s", createdScan.Status)
	}

	// Verifica accodamento nell'orchestratore
	orchClient.mu.Lock()
	if len(orchClient.enqueuedScans) != 1 || orchClient.enqueuedScans[0] != createdScan.ID {
		t.Errorf("Scansione non accodata correttamente nell'orchestratore: %v", orchClient.enqueuedScans)
	}
	orchClient.mu.Unlock()

	// Verifica aggiornamento LastRunAt e NextRunAt su schedule
	updatedSched, _ := schedRepo.GetByID(context.Background(), tenantID, schedID)
	if updatedSched.LastRunAt == nil {
		t.Error("LastRunAt non deve essere nil dopo l'esecuzione del job")
	}
	if updatedSched.NextRunAt == nil {
		t.Error("NextRunAt non deve essere nil dopo l'aggiornamento dell'orario")
	}
}

func TestSchedulerService_TriggerJob_SkipIfRunning(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	scanRepo := newMockScanRepo()
	orchClient := &mockOrchestratorClient{}

	tenantID := uuid.New()
	targetID := uuid.New()
	schedID := uuid.New()

	// Creiamo una scansione già RUNNING sullo stesso target
	runningScan := &domain.Scan{
		ID:       uuid.New(),
		TenantID: tenantID,
		TargetID: targetID,
		Status:   domain.ScanStatusRunning,
	}
	_ = scanRepo.Create(context.Background(), runningScan)

	schedule := &domain.Schedule{
		ID:             schedID,
		TenantID:       tenantID,
		TargetID:       targetID,
		ProfileID:      uuid.New(),
		CronExpression: "0 2 * * *",
		IsActive:       true,
	}
	_ = schedRepo.Create(context.Background(), schedule)

	scheduler := NewSchedulerService(schedRepo, scanRepo, orchClient, nil)

	// Trigger: deve rilevare la scansione attiva e saltare l'esecuzione
	authorizeScheduleFixture(scheduler, schedule)
	_ = schedRepo.Update(context.Background(), schedule)
	scheduler.TriggerJob(schedule)

	// Nessuna nuova scansione deve essere stata accodata
	orchClient.mu.Lock()
	if len(orchClient.enqueuedScans) != 0 {
		t.Errorf("Attesi 0 scan accodati quando target ha scan attiva, trovati %d", len(orchClient.enqueuedScans))
	}
	orchClient.mu.Unlock()
}

func TestSchedulerService_DynamicOperations(t *testing.T) {
	schedRepo := newMockScheduleRepo()
	scanRepo := newMockScanRepo()
	orchClient := &mockOrchestratorClient{}

	scheduler := NewSchedulerService(schedRepo, scanRepo, orchClient, nil)
	ctx := context.Background()
	_ = scheduler.Start(ctx)
	defer scheduler.Stop()

	tenantID := uuid.New()
	sched := &domain.Schedule{
		ID:             uuid.New(),
		TenantID:       tenantID,
		TargetID:       uuid.New(),
		ProfileID:      uuid.New(),
		CronExpression: "0 12 * * *",
		IsActive:       true,
	}

	ctx = authorizeScheduleFixture(scheduler, sched)
	// 1. AddSchedule
	if err := scheduler.AddSchedule(ctx, sched); err != nil {
		t.Fatalf("AddSchedule fallito: %v", err)
	}

	// 2. ToggleSchedule (disattiva)
	toggled, err := scheduler.ToggleSchedule(ctx, tenantID, sched.ID)
	if err != nil {
		t.Fatalf("ToggleSchedule fallito: %v", err)
	}
	if toggled.IsActive {
		t.Error("Atteso is_active = false dopo il primo toggle")
	}

	// 3. ToggleSchedule (riattiva)
	toggledBack, err := scheduler.ToggleSchedule(ctx, tenantID, sched.ID)
	if err != nil {
		t.Fatalf("ToggleSchedule secondario fallito: %v", err)
	}
	if !toggledBack.IsActive {
		t.Error("Atteso is_active = true dopo il secondo toggle")
	}

	// 4. UpdateSchedule
	sched.CronExpression = "30 14 * * *"
	if err := scheduler.UpdateSchedule(ctx, sched); err != domain.ErrScanAuthorizationRequired {
		t.Fatalf("changed recurrence must invalidate consent, got: %v", err)
	}
	sched.CronExpression = "0 12 * * *"
	if err := scheduler.UpdateSchedule(ctx, sched); err != nil {
		t.Fatalf("UpdateSchedule fallito: %v", err)
	}

	// 5. DeleteSchedule
	if err := scheduler.DeleteSchedule(ctx, tenantID, sched.ID); err != nil {
		t.Fatalf("DeleteSchedule fallito: %v", err)
	}
	_, err = schedRepo.GetByID(ctx, tenantID, sched.ID)
	if err != domain.ErrNotFound {
		t.Errorf("Atteso ErrNotFound per schedule eliminata, ottenuto: %v", err)
	}
}
