package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
)

// CreateGroupInput racchiude i parametri per la creazione di un gruppo/cartella
type CreateGroupInput struct {
	TenantID    uuid.UUID  `json:"tenant_id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	ColorTag    *string    `json:"color_tag,omitempty"`
}

// UpdateGroupInput racchiude i parametri per l'aggiornamento e riposizionamento di un gruppo
type UpdateGroupInput struct {
	TenantID    uuid.UUID  `json:"tenant_id"`
	ID          uuid.UUID  `json:"id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	ColorTag    *string    `json:"color_tag,omitempty"`
}

// TargetGroupService definisce la business logic per l'organizzazione gerarchica dei target
type TargetGroupService interface {
	CreateGroup(ctx context.Context, input CreateGroupInput) (*domain.TargetGroup, error)
	GetGroup(ctx context.Context, tenantID, id uuid.UUID) (*domain.TargetGroup, error)
	ListGroups(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroup, error)
	GetTree(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroupNode, error)
	UpdateGroup(ctx context.Context, input UpdateGroupInput) (*domain.TargetGroup, error)
	DeleteGroup(ctx context.Context, tenantID, id uuid.UUID) error
}

type targetGroupService struct {
	groupRepo domain.TargetGroupRepository
}

// NewTargetGroupService crea una nuova istanza di TargetGroupService
func NewTargetGroupService(groupRepo domain.TargetGroupRepository) TargetGroupService {
	return &targetGroupService{
		groupRepo: groupRepo,
	}
}

// CreateGroup crea un nuovo gruppo verificando l'esistenza del genitore se specificato
func (s *targetGroupService) CreateGroup(ctx context.Context, input CreateGroupInput) (*domain.TargetGroup, error) {
	cleanName := strings.TrimSpace(input.Name)
	if input.TenantID == uuid.Nil || cleanName == "" {
		return nil, domain.ErrValidation
	}

	if input.ParentID != nil {
		if *input.ParentID == uuid.Nil {
			return nil, domain.ErrValidation
		}
		// Verifica appartenenza del gruppo padre al medesimo tenant (NFR-SEC-03)
		if _, err := s.groupRepo.GetByID(ctx, input.TenantID, *input.ParentID); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	group := &domain.TargetGroup{
		ID:          uuid.New(),
		TenantID:    input.TenantID,
		ParentID:    input.ParentID,
		Name:        cleanName,
		Description: input.Description,
		ColorTag:    input.ColorTag,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// GetGroup recupera un singolo gruppo verificando l'isolamento tenant
func (s *targetGroupService) GetGroup(ctx context.Context, tenantID, id uuid.UUID) (*domain.TargetGroup, error) {
	if tenantID == uuid.Nil || id == uuid.Nil {
		return nil, domain.ErrValidation
	}
	return s.groupRepo.GetByID(ctx, tenantID, id)
}

// ListGroups restituisce l'elenco piatto di tutti i gruppi del tenant
func (s *targetGroupService) ListGroups(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroup, error) {
	if tenantID == uuid.Nil {
		return nil, domain.ErrValidation
	}
	return s.groupRepo.List(ctx, tenantID)
}

// UpdateGroup aggiorna i metadati del gruppo e controlla le dipendenze per prevenire cicli
func (s *targetGroupService) UpdateGroup(ctx context.Context, input UpdateGroupInput) (*domain.TargetGroup, error) {
	cleanName := strings.TrimSpace(input.Name)
	if input.TenantID == uuid.Nil || input.ID == uuid.Nil || cleanName == "" {
		return nil, domain.ErrValidation
	}

	group, err := s.groupRepo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		return nil, err
	}

	if input.ParentID != nil {
		// 1. Un nodo non può essere padre di sé stesso
		if *input.ParentID == input.ID {
			return nil, domain.ErrValidation
		}

		// 2. Il gruppo padre deve esistere e appartenere allo stesso tenant
		if _, err := s.groupRepo.GetByID(ctx, input.TenantID, *input.ParentID); err != nil {
			return nil, err
		}

		// 3. Prevenzione cicli nell'albero (non si può impostare come genitore un proprio discendente)
		allGroups, err := s.groupRepo.List(ctx, input.TenantID)
		if err != nil {
			return nil, err
		}

		parentMap := make(map[uuid.UUID]*uuid.UUID, len(allGroups))
		for _, g := range allGroups {
			parentMap[g.ID] = g.ParentID
		}

		curr := input.ParentID
		visited := make(map[uuid.UUID]bool)
		for curr != nil {
			if *curr == input.ID {
				return nil, domain.ErrValidation // Rilevato ciclo
			}
			if visited[*curr] {
				break
			}
			visited[*curr] = true
			curr = parentMap[*curr]
		}
	}

	group.Name = cleanName
	group.Description = input.Description
	group.ParentID = input.ParentID
	group.ColorTag = input.ColorTag
	group.UpdatedAt = time.Now().UTC()

	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

// DeleteGroup elimina un gruppo appartenente al tenant
func (s *targetGroupService) DeleteGroup(ctx context.Context, tenantID, id uuid.UUID) error {
	if tenantID == uuid.Nil || id == uuid.Nil {
		return domain.ErrValidation
	}
	return s.groupRepo.Delete(ctx, tenantID, id)
}

// GetTree costruisce l'albero gerarchico dei gruppi, calcola ricorsivamente il conteggio
// dei target e propaga il massimo livello di rischio aggregato (FR-09.1, FR-09.2, FR-09.3)
func (s *targetGroupService) GetTree(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroupNode, error) {
	if tenantID == uuid.Nil {
		return nil, domain.ErrValidation
	}

	groups, err := s.groupRepo.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return []*domain.TargetGroupNode{}, nil
	}

	targetCounts, err := s.groupRepo.GetTargetCountsByGroup(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	severityCounts, err := s.groupRepo.GetSeverityCountsByGroup(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 1. Costruzione mappa dei nodi
	nodeMap := make(map[uuid.UUID]*domain.TargetGroupNode, len(groups))
	for _, g := range groups {
		directTargets := targetCounts[g.ID]
		directSev := make(map[domain.SeverityLevel]int)
		if sc, ok := severityCounts[g.ID]; ok {
			for k, v := range sc {
				directSev[k] = v
			}
		}

		nodeMap[g.ID] = &domain.TargetGroupNode{
			ID:                g.ID,
			TenantID:          g.TenantID,
			ParentID:          g.ParentID,
			Name:              g.Name,
			Description:       g.Description,
			ColorTag:          g.ColorTag,
			CreatedAt:         g.CreatedAt,
			UpdatedAt:         g.UpdatedAt,
			Children:          make([]*domain.TargetGroupNode, 0),
			DirectTargetCount: directTargets,
			TargetCount:       directTargets,
			SeverityCounts:    directSev,
			AggregateRisk:     domain.SeverityNone,
		}
	}

	// 2. Assemblaggio della gerarchia padre-figlio
	var roots []*domain.TargetGroupNode
	for _, g := range groups {
		node := nodeMap[g.ID]
		if g.ParentID != nil {
			parent, exists := nodeMap[*g.ParentID]
			if exists {
				parent.Children = append(parent.Children, node)
			} else {
				roots = append(roots, node)
			}
		} else {
			roots = append(roots, node)
		}
	}

	// 3. Post-order traversal per aggregazione ricorsiva e calcolo rischio massimo
	var aggregateNode func(node *domain.TargetGroupNode)
	aggregateNode = func(node *domain.TargetGroupNode) {
		// Ordinamento alfabetico dei figli per output deterministico
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Name < node.Children[j].Name
		})

		for _, child := range node.Children {
			aggregateNode(child)
			node.TargetCount += child.TargetCount
			for sev, count := range child.SeverityCounts {
				node.SeverityCounts[sev] += count
			}
		}

		// Calcolo rischio aggregato con precedenza: CRITICAL > HIGH > MEDIUM > LOW > INFO > NONE
		switch {
		case node.SeverityCounts[domain.SeverityCritical] > 0:
			node.AggregateRisk = domain.SeverityCritical
		case node.SeverityCounts[domain.SeverityHigh] > 0:
			node.AggregateRisk = domain.SeverityHigh
		case node.SeverityCounts[domain.SeverityMedium] > 0:
			node.AggregateRisk = domain.SeverityMedium
		case node.SeverityCounts[domain.SeverityLow] > 0:
			node.AggregateRisk = domain.SeverityLow
		case node.SeverityCounts[domain.SeverityInfo] > 0:
			node.AggregateRisk = domain.SeverityInfo
		default:
			node.AggregateRisk = domain.SeverityNone
		}
	}

	// Ordinamento radici
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].Name < roots[j].Name
	})

	for _, root := range roots {
		aggregateNode(root)
	}

	return roots, nil
}

