package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/scanner"
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
	StartGroupScan(ctx context.Context, tenantID, groupID, profileID uuid.UUID) ([]*domain.Scan, error)
	GetScanFindings(ctx context.Context, tenantID, scanID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error)
	ListProfiles(ctx context.Context, tenantID *uuid.UUID, category *domain.ScanProfileCategory) ([]*domain.ScanProfile, error)
	CreateCustomProfile(ctx context.Context, tenantID uuid.UUID, name, nmapArgs string, description *string, timeoutSeconds int, category domain.ScanProfileCategory) (*domain.ScanProfile, error)
	DeleteCustomProfile(ctx context.Context, tenantID, id uuid.UUID) error
}

type scanService struct {
	scanRepo     domain.ScanRepository
	profileRepo  domain.ScanProfileRepository
	targetRepo   domain.TargetRepository
	findingRepo  domain.FindingRepository
	groupRepo    domain.TargetGroupRepository
	orchestrator OrchestratorClient
}

// NewScanService crea una nuova istanza del servizio scansioni
func NewScanService(
	scanRepo domain.ScanRepository,
	profileRepo domain.ScanProfileRepository,
	targetRepo domain.TargetRepository,
	findingRepo domain.FindingRepository,
	groupRepo domain.TargetGroupRepository,
	orchestrator OrchestratorClient,
) ScanService {
	return &scanService{
		scanRepo:     scanRepo,
		profileRepo:  profileRepo,
		targetRepo:   targetRepo,
		findingRepo:  findingRepo,
		groupRepo:    groupRepo,
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

// ListProfiles restituisce l'elenco dei profili di scansione supportati dalla piattaforma,
// con eventuale filtro per categoria specialistica e inclusione dei profili custom del tenant
func (s *scanService) ListProfiles(ctx context.Context, tenantID *uuid.UUID, category *domain.ScanProfileCategory) ([]*domain.ScanProfile, error) {
	if category != nil && *category != "" {
		return s.profileRepo.ListByCategory(ctx, tenantID, *category)
	}
	return s.profileRepo.List(ctx, tenantID)
}

// CreateCustomProfile valida gli argomenti Nmap e crea un nuovo profilo custom per il tenant
func (s *scanService) CreateCustomProfile(
	ctx context.Context,
	tenantID uuid.UUID,
	name string,
	nmapArgs string,
	description *string,
	timeoutSeconds int,
	category domain.ScanProfileCategory,
) (*domain.ScanProfile, error) {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, fmt.Errorf("%w: nome profilo obbligatorio", domain.ErrValidation)
	}

	cleanArgs := strings.TrimSpace(nmapArgs)
	if cleanArgs == "" {
		return nil, fmt.Errorf("%w: argomenti Nmap obbligatori", domain.ErrValidation)
	}

	// Validazione di sicurezza preventiva contro flag injection (NFR-SEC-01)
	parsedArgs, err := scanner.ParseProfileArgs(cleanArgs)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrDisallowedScanOption, err)
	}
	canonicalArgs := strings.Join(parsedArgs, " ")

	if timeoutSeconds <= 0 {
		timeoutSeconds = 3600
	} else if timeoutSeconds > 86400 {
		timeoutSeconds = 86400 // Max 24 ore
	}

	if category == "" {
		category = domain.ProfileCategoryCustom
	}

	profile := &domain.ScanProfile{
		ID:             uuid.New(),
		Name:           cleanName,
		NmapArgs:       canonicalArgs,
		Description:    description,
		TimeoutSeconds: timeoutSeconds,
		Category:       category,
		IsCustom:       true,
		TenantID:       &tenantID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.profileRepo.Create(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// DeleteCustomProfile elimina un profilo personalizzato del tenant
func (s *scanService) DeleteCustomProfile(ctx context.Context, tenantID, id uuid.UUID) error {
	if tenantID == uuid.Nil || id == uuid.Nil {
		return domain.ErrValidation
	}
	return s.profileRepo.Delete(ctx, tenantID, id)
}

// StartGroupScan avvia ricorsivamente una scansione su tutti i target validi appartenenti
// alla cartella specificata e a tutti i suoi sotto-alberi discendenti (FR-09.4)
func (s *scanService) StartGroupScan(ctx context.Context, tenantID, groupID, profileID uuid.UUID) ([]*domain.Scan, error) {
	if tenantID == uuid.Nil || groupID == uuid.Nil || profileID == uuid.Nil {
		return nil, domain.ErrValidation
	}

	// 1. Verifica esistenza profilo di scansione
	if _, err := s.profileRepo.GetByID(ctx, profileID); err != nil {
		return nil, fmt.Errorf("profilo di scansione non valido: %w", err)
	}

	// 2. Verifica appartenenza ed esistenza del gruppo radice
	if _, err := s.groupRepo.GetByID(ctx, tenantID, groupID); err != nil {
		return nil, err
	}

	// 3. Risoluzione ricorsiva di tutti i gruppi nel sotto-albero
	allGroups, err := s.groupRepo.List(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("recupero gerarchia cartelle fallito: %w", err)
	}

	subtreeGroupIDs := map[uuid.UUID]bool{groupID: true}
	added := true
	for added {
		added = false
		for _, g := range allGroups {
			if g.ParentID != nil && subtreeGroupIDs[*g.ParentID] && !subtreeGroupIDs[g.ID] {
				subtreeGroupIDs[g.ID] = true
				added = true
			}
		}
	}

	// 4. Recupero di tutti i target del tenant per filtrare quelli associati ai gruppi del sotto-albero
	targets, _, err := s.targetRepo.List(ctx, tenantID, domain.TargetFilter{Limit: 1000})
	if err != nil {
		return nil, fmt.Errorf("recupero target per scansione cartella fallito: %w", err)
	}

	var targetsToScan []*domain.Target
	for _, t := range targets {
		if t.GroupID != nil && subtreeGroupIDs[*t.GroupID] && !t.IsBlacklisted {
			targetsToScan = append(targetsToScan, t)
		}
	}

	if len(targetsToScan) == 0 {
		return []*domain.Scan{}, nil
	}

	// 5. Avvio sicuro dei job di scansione per ciascun target non bloccato
	var startedScans []*domain.Scan
	for _, target := range targetsToScan {
		// Controllo blacklist globale/tenant
		isBL, err := s.targetRepo.IsBlacklisted(ctx, tenantID, target.Value)
		if err != nil || isBL {
			continue
		}

		// Controllo scansioni concorrenti attive sullo stesso target
		activeScan, err := s.scanRepo.GetActiveByTarget(ctx, tenantID, target.ID)
		if err != nil || activeScan != nil {
			continue
		}

		now := time.Now()
		scan := &domain.Scan{
			ID:          uuid.New(),
			TenantID:    tenantID,
			TargetID:    target.ID,
			ProfileID:   profileID,
			TriggerType: domain.TriggerManual,
			Status:      domain.ScanStatusQueued,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := s.scanRepo.Create(ctx, scan); err != nil {
			continue
		}

		if err := s.orchestrator.Enqueue(tenantID, scan.ID); err != nil {
			_ = scan.MarkFailed("coda del worker pool satura")
			_ = s.scanRepo.Update(ctx, scan)
			continue
		}

		startedScans = append(startedScans, scan)
	}

	return startedScans, nil
}

