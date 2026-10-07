package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/validator"
)

// CreateTargetInput racchiude i parametri per la registrazione di un bersaglio
type CreateTargetInput struct {
	TenantID      uuid.UUID  `json:"tenant_id"`
	GroupID       *uuid.UUID `json:"group_id"`
	Value         string     `json:"value"`
	Description   *string    `json:"description"`
	IsBlacklisted bool       `json:"is_blacklisted"`
}

// UpdateTargetInput racchiude i campi modificabili di un target
type UpdateTargetInput struct {
	TenantID      uuid.UUID  `json:"tenant_id"`
	ID            uuid.UUID  `json:"id"`
	GroupID       *uuid.UUID `json:"group_id"`
	Description   *string    `json:"description"`
	IsBlacklisted bool       `json:"is_blacklisted"`
}

// TargetService definisce la business logic per la gestione dei target e dei gruppi
type TargetService interface {
	CreateTarget(ctx context.Context, input CreateTargetInput) (*domain.Target, error)
	GetTarget(ctx context.Context, tenantID, id uuid.UUID) (*domain.Target, error)
	ListTargets(ctx context.Context, tenantID uuid.UUID, filter domain.TargetFilter) ([]*domain.Target, int, error)
	UpdateTarget(ctx context.Context, input UpdateTargetInput) (*domain.Target, error)
	DeleteTarget(ctx context.Context, tenantID, id uuid.UUID) error

	CreateGroup(ctx context.Context, tenantID uuid.UUID, name string, description *string) (*domain.TargetGroup, error)
	ListGroups(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroup, error)
	DeleteGroup(ctx context.Context, tenantID, id uuid.UUID) error
}

type targetService struct {
	targetRepo domain.TargetRepository
	groupRepo  domain.TargetGroupRepository
}

// NewTargetService crea una nuova istanza di TargetService
func NewTargetService(targetRepo domain.TargetRepository, groupRepo domain.TargetGroupRepository) TargetService {
	return &targetService{
		targetRepo: targetRepo,
		groupRepo:  groupRepo,
	}
}

// CreateTarget valida sintassi e blacklist, inferisce il tipo e persiste il target
func (s *targetService) CreateTarget(ctx context.Context, input CreateTargetInput) (*domain.Target, error) {
	if input.TenantID == uuid.Nil {
		return nil, domain.ErrValidation
	}

	// 1. Validazione sintattica e inferenza tipo automatica
	info, err := validator.ValidateTarget(input.Value)
	if err != nil {
		return nil, domain.ErrTargetInvalid
	}

	// 2. Verifica appartenenza del gruppo al medesimo tenant (se specificato)
	if input.GroupID != nil {
		if _, err := s.groupRepo.GetByID(ctx, input.TenantID, *input.GroupID); err != nil {
			return nil, err
		}
	}

	// 3. Controllo blacklist: un target operativo non può collidere con le blacklist del tenant
	if !input.IsBlacklisted {
		isBL, err := s.targetRepo.IsBlacklisted(ctx, input.TenantID, info.CleanValue)
		if err != nil {
			return nil, err
		}
		if isBL {
			return nil, domain.ErrTargetBlacklisted
		}
	}

	now := time.Now().UTC()
	target := &domain.Target{
		ID:            uuid.New(),
		TenantID:      input.TenantID,
		GroupID:       input.GroupID,
		Type:          info.Type,
		Value:         info.CleanValue,
		Description:   input.Description,
		IsBlacklisted: input.IsBlacklisted,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.targetRepo.Create(ctx, target); err != nil {
		return nil, err
	}

	return target, nil
}

// GetTarget recupera un singolo target garantendo l'isolamento tenant
func (s *targetService) GetTarget(ctx context.Context, tenantID, id uuid.UUID) (*domain.Target, error) {
	if tenantID == uuid.Nil || id == uuid.Nil {
		return nil, domain.ErrValidation
	}
	return s.targetRepo.GetByID(ctx, tenantID, id)
}

// ListTargets restituisce l'elenco paginato dei target per il tenant
func (s *targetService) ListTargets(ctx context.Context, tenantID uuid.UUID, filter domain.TargetFilter) ([]*domain.Target, int, error) {
	if tenantID == uuid.Nil {
		return nil, 0, domain.ErrValidation
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.targetRepo.List(ctx, tenantID, filter)
}

// UpdateTarget aggiorna i metadati del target (gruppo, descrizione, flag blacklist)
func (s *targetService) UpdateTarget(ctx context.Context, input UpdateTargetInput) (*domain.Target, error) {
	if input.TenantID == uuid.Nil || input.ID == uuid.Nil {
		return nil, domain.ErrValidation
	}

	target, err := s.targetRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return nil, err
	}

	if input.GroupID != nil {
		if _, err := s.groupRepo.GetByID(ctx, input.TenantID, *input.GroupID); err != nil {
			return nil, err
		}
	}

	target.GroupID = input.GroupID
	target.Description = input.Description
	target.IsBlacklisted = input.IsBlacklisted
	target.UpdatedAt = time.Now().UTC()

	if err := s.targetRepo.Update(ctx, target); err != nil {
		return nil, err
	}

	return target, nil
}

// DeleteTarget rimuove un target verificando l'appartenenza al tenant
func (s *targetService) DeleteTarget(ctx context.Context, tenantID, id uuid.UUID) error {
	if tenantID == uuid.Nil || id == uuid.Nil {
		return domain.ErrValidation
	}
	return s.targetRepo.Delete(ctx, tenantID, id)
}

// CreateGroup registra un nuovo gruppo logico per il tenant
func (s *targetService) CreateGroup(ctx context.Context, tenantID uuid.UUID, name string, description *string) (*domain.TargetGroup, error) {
	cleanName := strings.TrimSpace(name)
	if tenantID == uuid.Nil || cleanName == "" {
		return nil, domain.ErrValidation
	}

	now := time.Now().UTC()
	group := &domain.TargetGroup{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        cleanName,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// ListGroups elenca tutti i gruppi del tenant
func (s *targetService) ListGroups(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroup, error) {
	if tenantID == uuid.Nil {
		return nil, domain.ErrValidation
	}
	return s.groupRepo.List(ctx, tenantID)
}

// DeleteGroup elimina un gruppo appartenente al tenant
func (s *targetService) DeleteGroup(ctx context.Context, tenantID, id uuid.UUID) error {
	if tenantID == uuid.Nil || id == uuid.Nil {
		return domain.ErrValidation
	}
	return s.groupRepo.Delete(ctx, tenantID, id)
}
