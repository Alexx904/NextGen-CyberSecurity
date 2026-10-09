package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ScanTriggerType definisce l'origine dell'attivazione della scansione
type ScanTriggerType string

const (
	TriggerManual    ScanTriggerType = "MANUAL"
	TriggerScheduled ScanTriggerType = "SCHEDULED"
)

// ScanStatus definisce gli stati del ciclo di vita di una scansione
type ScanStatus string

const (
	ScanStatusQueued    ScanStatus = "QUEUED"
	ScanStatusRunning   ScanStatus = "RUNNING"
	ScanStatusCompleted ScanStatus = "COMPLETED"
	ScanStatusFailed    ScanStatus = "FAILED"
	ScanStatusCancelled ScanStatus = "CANCELLED"
)

// IsTerminal restituisce true se lo stato è finale e non può più mutare
func (s ScanStatus) IsTerminal() bool {
	return s == ScanStatusCompleted || s == ScanStatusFailed || s == ScanStatusCancelled
}

// IsValid verifica se lo stato specificato è ammesso
func (s ScanStatus) IsValid() bool {
	switch s {
	case ScanStatusQueued, ScanStatusRunning, ScanStatusCompleted, ScanStatusFailed, ScanStatusCancelled:
		return true
	default:
		return false
	}
}

// ScanSummaryStats racchiude le metriche sintetiche calcolate al termine della scansione
type ScanSummaryStats struct {
	TotalHosts      int     `json:"total_hosts"`
	UpHosts         int     `json:"up_hosts"`
	DownHosts       int     `json:"down_hosts"`
	TotalOpenPorts  int     `json:"total_open_ports"`
	CriticalCount   int     `json:"critical_count"`
	HighCount       int     `json:"high_count"`
	MediumCount     int     `json:"medium_count"`
	LowCount        int     `json:"low_count"`
	InfoCount       int     `json:"info_count"`
	DurationSeconds float64 `json:"duration_seconds"`
}

// ScanProfile definisce la configurazione dei flag di scansione Nmap preimpostati
type ScanProfile struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	NmapArgs       string    `json:"nmap_args" db:"nmap_args"`
	Description    *string   `json:"description,omitempty" db:"description"`
	TimeoutSeconds int       `json:"timeout_seconds" db:"timeout_seconds"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Scan rappresenta un'esecuzione (on-demand o schedulata) di vulnerability scanning
type Scan struct {
	ID           uuid.UUID        `json:"id" db:"id"`
	TenantID     uuid.UUID        `json:"tenant_id" db:"tenant_id"`
	TargetID     uuid.UUID        `json:"target_id" db:"target_id"`
	ProfileID    uuid.UUID        `json:"profile_id" db:"profile_id"`
	TriggerType  ScanTriggerType  `json:"trigger_type" db:"trigger_type"`
	Status       ScanStatus       `json:"status" db:"status"`
	StartedAt    *time.Time       `json:"started_at,omitempty" db:"started_at"`
	CompletedAt  *time.Time       `json:"completed_at,omitempty" db:"completed_at"`
	SummaryStats ScanSummaryStats `json:"summary_stats" db:"summary_stats"`
	ErrorMessage *string          `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at" db:"updated_at"`
}

// CanCancel determina se la scansione può essere annullata dall'utente
func (s *Scan) CanCancel() bool {
	return s.Status == ScanStatusQueued || s.Status == ScanStatusRunning
}

// MarkRunning avvia la scansione e imposta il timestamp di inizio
func (s *Scan) MarkRunning() error {
	if s.Status != ScanStatusQueued {
		return fmt.Errorf("%w: impossibile avviare una scansione con stato %s", ErrInvalidStateTransition, s.Status)
	}
	now := time.Now()
	s.Status = ScanStatusRunning
	s.StartedAt = &now
	s.UpdatedAt = now
	return nil
}

// MarkCompleted segna la scansione come conclusa con successo
func (s *Scan) MarkCompleted(stats ScanSummaryStats) error {
	if s.Status != ScanStatusRunning {
		return fmt.Errorf("%w: impossibile completare una scansione non in esecuzione (stato: %s)", ErrInvalidStateTransition, s.Status)
	}
	now := time.Now()
	s.Status = ScanStatusCompleted
	s.CompletedAt = &now
	s.SummaryStats = stats
	s.UpdatedAt = now
	return nil
}

// MarkFailed segna la scansione come fallita registrando l'errore
func (s *Scan) MarkFailed(errMsg string) error {
	if s.Status.IsTerminal() {
		return fmt.Errorf("%w: impossibile far fallire una scansione già terminata (stato: %s)", ErrInvalidStateTransition, s.Status)
	}
	now := time.Now()
	s.Status = ScanStatusFailed
	s.CompletedAt = &now
	s.ErrorMessage = &errMsg
	s.UpdatedAt = now
	return nil
}

// MarkCancelled contrassegna la scansione come annullata
func (s *Scan) MarkCancelled() error {
	if !s.CanCancel() {
		return fmt.Errorf("%w: la scansione è nello stato %s", ErrScanCannotBeCancelled, s.Status)
	}
	now := time.Now()
	s.Status = ScanStatusCancelled
	s.CompletedAt = &now
	s.UpdatedAt = now
	return nil
}

// Schedule definisce una pianificazione ricorrente per l'esecuzione di scansioni periodiche
type Schedule struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	TenantID       uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	TargetID       uuid.UUID  `json:"target_id" db:"target_id"`
	ProfileID      uuid.UUID  `json:"profile_id" db:"profile_id"`
	CronExpression string     `json:"cron_expression" db:"cron_expression"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	NextRunAt      *time.Time `json:"next_run_at,omitempty" db:"next_run_at"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty" db:"last_run_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// ScanRepository definisce l'interfaccia di persistenza per le scansioni
type ScanRepository interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Scan, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*Scan, int, error)
	GetActiveByTarget(ctx context.Context, tenantID, targetID uuid.UUID) (*Scan, error)
	Create(ctx context.Context, scan *Scan) error
	Update(ctx context.Context, scan *Scan) error
}

// ScanProfileRepository definisce l'interfaccia per i profili di scansione
type ScanProfileRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*ScanProfile, error)
	GetByName(ctx context.Context, name string) (*ScanProfile, error)
	List(ctx context.Context) ([]*ScanProfile, error)
}

// ScheduleRepository definisce l'interfaccia per la gestione delle pianificazioni ricorrenti
type ScheduleRepository interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Schedule, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*Schedule, error)
	ListActive(ctx context.Context) ([]*Schedule, error)
	Create(ctx context.Context, schedule *Schedule) error
	Update(ctx context.Context, schedule *Schedule) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}

