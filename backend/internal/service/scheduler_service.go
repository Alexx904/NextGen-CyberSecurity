package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"nextgen-cybersec/internal/domain"
)

// OrchestratorSchedulerClient definisce l'interfaccia necessaria dello ScanOrchestrator per accodare scansioni
type OrchestratorSchedulerClient interface {
	Enqueue(tenantID, scanID uuid.UUID) error
}

// SchedulerService gestisce l'orchestrazione temporale ricorrente dei job di vulnerability scanning (NFR-PERF-01)
type SchedulerService struct {
	cronEngine   *cron.Cron
	parser       cron.Parser
	scheduleRepo domain.ScheduleRepository
	scanRepo     domain.ScanRepository
	orchestrator OrchestratorSchedulerClient
	logger       *slog.Logger

	mu       sync.RWMutex
	entryMap map[uuid.UUID]cron.EntryID
	running  bool
}

// NewSchedulerService inizializza una nuova istanza di SchedulerService
func NewSchedulerService(
	scheduleRepo domain.ScheduleRepository,
	scanRepo domain.ScanRepository,
	orchestrator OrchestratorSchedulerClient,
	logger *slog.Logger,
) *SchedulerService {
	if logger == nil {
		logger = slog.Default()
	}

	// Parser standard a 5 campi (minuto, ora, giorno del mese, mese, giorno della settimana) con supporto ai descriptor
	standardParser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)

	return &SchedulerService{
		cronEngine: cron.New(
			cron.WithParser(standardParser),
		),
		parser:       standardParser,
		scheduleRepo: scheduleRepo,
		scanRepo:     scanRepo,
		orchestrator: orchestrator,
		logger:       logger,
		entryMap:     make(map[uuid.UUID]cron.EntryID),
	}
}

// Start avvia lo scheduler in background caricando tutte le pianificazioni attive da database
func (s *SchedulerService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.logger.Info("Inizializzazione e avvio SchedulerService...")

	// 1. Caricamento pianificazioni attive da PostgreSQL
	activeSchedules, err := s.scheduleRepo.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("recupero pianificazioni attive fallito: %w", err)
	}

	// 2. Registrazione di ciascun job nell'engine cron
	for _, sched := range activeSchedules {
		if err := s.registerJobLocked(sched); err != nil {
			s.logger.Error("Registrazione pianificazione fallita all'avvio",
				"schedule_id", sched.ID,
				"cron", sched.CronExpression,
				"error", err,
			)
			continue
		}
	}

	// 3. Avvio dell'engine di background
	s.cronEngine.Start()
	s.running = true

	s.logger.Info("SchedulerService avviato con successo",
		"job_attivi_caricati", len(s.entryMap),
	)

	return nil
}

// Stop arresta graceful lo scheduler cron
func (s *SchedulerService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.logger.Info("Arresto graceful di SchedulerService...")
	stopCtx := s.cronEngine.Stop()
	<-stopCtx.Done()

	s.entryMap = make(map[uuid.UUID]cron.EntryID)
	s.running = false
	s.logger.Info("SchedulerService arrestato")
}

// AddSchedule crea e persiste una nuova pianificazione, registrandola dinamicamente nell'engine se attiva
func (s *SchedulerService) AddSchedule(ctx context.Context, schedule *domain.Schedule) error {
	// Valida la sintassi dell'espressione cron
	schedSchedule, err := s.parser.Parse(schedule.CronExpression)
	if err != nil {
		return fmt.Errorf("espressione cron non valida (%s): %w", schedule.CronExpression, err)
	}

	now := time.Now().UTC()
	next := schedSchedule.Next(now)
	schedule.NextRunAt = &next

	if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
		return fmt.Errorf("salvataggio pianificazione fallito: %w", err)
	}

	if schedule.IsActive {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.running {
			if err := s.registerJobLocked(schedule); err != nil {
				return err
			}
		}
	}

	return nil
}

// UpdateSchedule aggiorna una pianificazione e ricarica il job nel cron engine
func (s *SchedulerService) UpdateSchedule(ctx context.Context, schedule *domain.Schedule) error {
	schedSchedule, err := s.parser.Parse(schedule.CronExpression)
	if err != nil {
		return fmt.Errorf("espressione cron non valida (%s): %w", schedule.CronExpression, err)
	}

	now := time.Now().UTC()
	next := schedSchedule.Next(now)
	schedule.NextRunAt = &next

	if err := s.scheduleRepo.Update(ctx, schedule); err != nil {
		return fmt.Errorf("aggiornamento pianificazione fallito: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Rimuovi job precedente se presente
	if entryID, exists := s.entryMap[schedule.ID]; exists {
		s.cronEngine.Remove(entryID)
		delete(s.entryMap, schedule.ID)
	}

	// Ri-registra se attivo
	if schedule.IsActive && s.running {
		if err := s.registerJobLocked(schedule); err != nil {
			return err
		}
	}

	return nil
}

// DeleteSchedule rimuove la pianificazione dal database e dal cron engine
func (s *SchedulerService) DeleteSchedule(ctx context.Context, tenantID, id uuid.UUID) error {
	if err := s.scheduleRepo.Delete(ctx, tenantID, id); err != nil {
		return fmt.Errorf("eliminazione pianificazione fallita: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.entryMap[id]; exists {
		s.cronEngine.Remove(entryID)
		delete(s.entryMap, id)
	}

	return nil
}

// ToggleSchedule inverte lo stato di attivazione della pianificazione
func (s *SchedulerService) ToggleSchedule(ctx context.Context, tenantID, id uuid.UUID) (*domain.Schedule, error) {
	schedule, err := s.scheduleRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	schedule.IsActive = !schedule.IsActive

	if schedule.IsActive {
		schedSchedule, err := s.parser.Parse(schedule.CronExpression)
		if err == nil {
			next := schedSchedule.Next(time.Now().UTC())
			schedule.NextRunAt = &next
		}
	} else {
		schedule.NextRunAt = nil
	}

	if err := s.scheduleRepo.Update(ctx, schedule); err != nil {
		return nil, fmt.Errorf("aggiornamento toggle fallito: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.entryMap[schedule.ID]; exists {
		s.cronEngine.Remove(entryID)
		delete(s.entryMap, schedule.ID)
	}

	if schedule.IsActive && s.running {
		_ = s.registerJobLocked(schedule)
	}

	return schedule, nil
}

// GetByID recupera una singola pianificazione per ID e tenant
func (s *SchedulerService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Schedule, error) {
	return s.scheduleRepo.GetByID(ctx, tenantID, id)
}

// ListByTenant restituisce tutte le pianificazioni del tenant specificato
func (s *SchedulerService) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Schedule, error) {
	return s.scheduleRepo.ListByTenant(ctx, tenantID)
}

// TriggerJob esegue la logica di invocazione di una scansione programmata
func (s *SchedulerService) TriggerJob(schedule *domain.Schedule) {
	ctx := context.Background()

	s.logger.Info("Trigger cron scattato per pianificazione",
		"schedule_id", schedule.ID,
		"tenant_id", schedule.TenantID,
		"target_id", schedule.TargetID,
	)

	// 1. Controllo preventivo: scansione concorrente già in corso sul target?
	active, err := s.scanRepo.GetActiveByTarget(ctx, schedule.TenantID, schedule.TargetID)
	if err != nil {
		s.logger.Error("Controllo scansioni attive fallito nello scheduler",
			"schedule_id", schedule.ID,
			"error", err,
		)
		return
	}
	if active != nil {
		s.logger.Warn("Scansione programmata saltata: una scansione è già in corso sul target",
			"schedule_id", schedule.ID,
			"active_scan_id", active.ID,
			"target_id", schedule.TargetID,
		)
		s.updateNextRunOnly(ctx, schedule)
		return
	}

	// 2. Creazione record scansione su DB in stato QUEUED
	now := time.Now().UTC()
	scan := &domain.Scan{
		ID:          uuid.New(),
		TenantID:    schedule.TenantID,
		TargetID:    schedule.TargetID,
		ProfileID:   schedule.ProfileID,
		TriggerType: domain.TriggerScheduled,
		Status:      domain.ScanStatusQueued,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.scanRepo.Create(ctx, scan); err != nil {
		s.logger.Error("Creazione record scansione programmata fallita",
			"schedule_id", schedule.ID,
			"error", err,
		)
		return
	}

	// 3. Aggiornamento timestamp di ultima e prossima esecuzione
	schedule.LastRunAt = &now
	if parsedSchedule, err := s.parser.Parse(schedule.CronExpression); err == nil {
		next := parsedSchedule.Next(now)
		schedule.NextRunAt = &next
	}
	_ = s.scheduleRepo.Update(ctx, schedule)

	// 4. Accodamento job nello ScanOrchestrator
	if s.orchestrator != nil {
		if err := s.orchestrator.Enqueue(schedule.TenantID, scan.ID); err != nil {
			s.logger.Error("Accodamento scansione programmata nell'orchestratore fallito",
				"scan_id", scan.ID,
				"error", err,
			)
			_ = scan.MarkFailed(fmt.Sprintf("accodamento fallito: %v", err))
			_ = s.scanRepo.Update(ctx, scan)
			return
		}
	}

	s.logger.Info("Scansione programmata generata e accodata con successo",
		"schedule_id", schedule.ID,
		"scan_id", scan.ID,
	)
}

func (s *SchedulerService) registerJobLocked(schedule *domain.Schedule) error {
	schedCopy := *schedule

	entryID, err := s.cronEngine.AddFunc(schedule.CronExpression, func() {
		s.TriggerJob(&schedCopy)
	})
	if err != nil {
		return fmt.Errorf("registrazione job cron (%s) fallita: %w", schedule.CronExpression, err)
	}

	s.entryMap[schedule.ID] = entryID
	return nil
}

func (s *SchedulerService) updateNextRunOnly(ctx context.Context, schedule *domain.Schedule) {
	if parsedSchedule, err := s.parser.Parse(schedule.CronExpression); err == nil {
		next := parsedSchedule.Next(time.Now().UTC())
		schedule.NextRunAt = &next
		_ = s.scheduleRepo.Update(ctx, schedule)
	}
}
