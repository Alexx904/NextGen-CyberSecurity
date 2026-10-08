package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
)

// OrchestratorClient astrae l'orchestratore per consentire dependency injection e mocking nei test
type OrchestratorClient interface {
	Enqueue(tenantID, scanID uuid.UUID) error
	CancelScan(tenantID, scanID uuid.UUID) error
}

// ScanService definisce l'interfaccia di alto livello per i casi d'uso relativi alle scansioni di rete
type ScanService interface {
	CancelScan(ctx context.Context, tenantID, scanID uuid.UUID) error
	GetScan(ctx context.Context, tenantID, scanID uuid.UUID) (*domain.Scan, error)
	ListScans(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Scan, int, error)
	StartScan(ctx context.Context, tenantID, targetID, profileID uuid.UUID) (*domain.Scan, error)
	GetScanFindings(ctx context.Context, tenantID, scanID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error)
	ListProfiles(ctx context.Context) ([]*domain.ScanProfile, error)
}

type scanService struct {
	scanRepo     domain.ScanRepository
	profileRepo  domain.ScanProfileRepository
	targetRepo   domain.TargetRepository
	findingRepo  domain.FindingRepository
	orchestrator OrchestratorClient
}

// NewScanService crea una nuova istanza del servizio scansioni
func NewScanService(
	scanRepo domain.ScanRepository,
	profileRepo domain.ScanProfileRepository,
	targetRepo domain.TargetRepository,
	findingRepo domain.FindingRepository,
	orchestrator OrchestratorClient,
) ScanService {
	return &scanService{
		scanRepo:     scanRepo,
		profileRepo:  profileRepo,
		targetRepo:   targetRepo,
		findingRepo:  findingRepo,
		orchestrator: orchestrator,
	}
}

// CancelScan gestisce la richiesta di cancellazione graceful verificando i vincoli di business e tenant
func (s *scanService) CancelScan(ctx context.Context, tenantID, scanID uuid.UUID) error {
	// 1. Verifica esistenza e appartenenza al tenant della scansione (NFR-SEC-03)
	scan, err := s.scanRepo.GetByID(ctx, tenantID, scanID)
	if err != nil {
		return err
	}

	// 2. Controllo dello stato: una scansione può essere annullata solo se QUEUED o RUNNING
	if !scan.CanCancel() {
		return domain.ErrScanCannotBeCancelled
	}

	// 3. Propagazione della richiesta di cancellazione all'orchestratore (NFR-REL-01)
	if err := s.orchestrator.CancelScan(tenantID, scanID); err != nil {
		return fmt.Errorf("cancellazione tramite orchestratore fallita: %w", err)
	}

	return nil
}

// GetScan recupera una scansione per ID verificando l'isolamento per tenant
func (s *scanService) GetScan(ctx context.Context, tenantID, scanID uuid.UUID) (*domain.Scan, error) {
	return s.scanRepo.GetByID(ctx, tenantID, scanID)
}

// ListScans restituisce l'elenco paginato delle scansioni del tenant
func (s *scanService) ListScans(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Scan, int, error) {
	return s.scanRepo.List(ctx, tenantID, limit, offset)
}

// StartScan convalida bersaglio e profilo, crea la scansione in stato QUEUED e la invia al worker pool
func (s *scanService) StartScan(ctx context.Context, tenantID, targetID, profileID uuid.UUID) (*domain.Scan, error) {
	// 1. Verifica esistenza del target e assenza da blacklist
	target, err := s.targetRepo.GetByID(ctx, tenantID, targetID)
	if err != nil {
		return nil, err
	}

	if target.IsBlacklisted {
		return nil, domain.ErrTargetBlacklisted
	}

	isBlacklisted, err := s.targetRepo.IsBlacklisted(ctx, tenantID, target.Value)
	if err != nil {
		return nil, fmt.Errorf("verifica blacklist fallita: %w", err)
	}
	if isBlacklisted {
		return nil, domain.ErrTargetBlacklisted
	}

	// 2. Verifica esistenza del profilo di scansione
	_, err = s.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("profilo di scansione non valido: %w", err)
	}

	// 3. Verifica preventiva di scansioni concorrenti attive sullo stesso target
	activeScan, err := s.scanRepo.GetActiveByTarget(ctx, tenantID, targetID)
	if err != nil {
		return nil, fmt.Errorf("controllo scansioni attive fallito: %w", err)
	}
	if activeScan != nil {
		return nil, domain.ErrScanAlreadyRunning
	}

	// 4. Creazione record scansione con stato QUEUED
	now := time.Now()
	scan := &domain.Scan{
		ID:          uuid.New(),
		TenantID:    tenantID,
		TargetID:    targetID,
		ProfileID:   profileID,
		TriggerType: domain.TriggerManual,
		Status:      domain.ScanStatusQueued,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.scanRepo.Create(ctx, scan); err != nil {
		return nil, fmt.Errorf("salvataggio scansione su DB fallito: %w", err)
	}

	// 5. Inoltro alla coda asincrona del worker pool
	if err := s.orchestrator.Enqueue(tenantID, scan.ID); err != nil {
		// Se la coda è satura, contrassegna come fallita
		_ = scan.MarkFailed("coda del worker pool satura")
		_ = s.scanRepo.Update(ctx, scan)
		return nil, fmt.Errorf("accodamento scansione fallito: %w", err)
	}

	return scan, nil
}

// GetScanFindings recupera i finding rilevati per una specifica scansione applicando i filtri richiesti
func (s *scanService) GetScanFindings(ctx context.Context, tenantID, scanID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
	// 1. Verifica preliminare dell'esistenza e isolamento tenant della scansione
	if _, err := s.scanRepo.GetByID(ctx, tenantID, scanID); err != nil {
		return nil, 0, err
	}

	filter.ScanID = &scanID
	return s.findingRepo.List(ctx, tenantID, filter)
}

// ListProfiles restituisce l'elenco dei profili di scansione supportati dalla piattaforma
func (s *scanService) ListProfiles(ctx context.Context) ([]*domain.ScanProfile, error) {
	return s.profileRepo.List(ctx)
}
