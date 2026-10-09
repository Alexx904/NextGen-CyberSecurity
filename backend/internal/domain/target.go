package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TargetType definisce la tipologia sintattica del target
type TargetType string

const (
	TargetTypeIP    TargetType = "IP"
	TargetTypeCIDR  TargetType = "CIDR"
	TargetTypeRange TargetType = "RANGE"
	TargetTypeFQDN  TargetType = "FQDN"
)

// IsValid verifica se la tipologia di target è supportata
func (t TargetType) IsValid() bool {
	switch t {
	case TargetTypeIP, TargetTypeCIDR, TargetTypeRange, TargetTypeFQDN:
		return true
	default:
		return false
	}
}

// TargetGroup raggruppa logicamente un insieme di target in una gerarchia ad albero (es. DMZ, Cloud, Uffici)
type TargetGroup struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TenantID    uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description,omitempty" db:"description"`
	ColorTag    *string    `json:"color_tag,omitempty" db:"color_tag"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// TargetGroupNode rappresenta un nodo della gerarchia ad albero dei gruppi, arricchito con
// conteggi ricorsivi dei target e calcolo del livello di rischio aggregato.
type TargetGroupNode struct {
	ID                uuid.UUID             `json:"id"`
	TenantID          uuid.UUID             `json:"tenant_id"`
	ParentID          *uuid.UUID            `json:"parent_id,omitempty"`
	Name              string                `json:"name"`
	Description       *string               `json:"description,omitempty"`
	ColorTag          *string               `json:"color_tag,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	Children          []*TargetGroupNode    `json:"children"`
	DirectTargetCount int                   `json:"direct_target_count"`
	TargetCount       int                   `json:"target_count"`
	AggregateRisk     SeverityLevel         `json:"aggregate_risk"`
	SeverityCounts    map[SeverityLevel]int `json:"severity_counts"`
}

// Target definisce un singolo bersaglio o scope di rete autorizzato alla scansione
type Target struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	TenantID      uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	GroupID       *uuid.UUID `json:"group_id,omitempty" db:"group_id"`
	Type          TargetType `json:"target_type" db:"target_type"`
	Value         string     `json:"value" db:"value"`
	Description   *string    `json:"description,omitempty" db:"description"`
	IsBlacklisted bool       `json:"is_blacklisted" db:"is_blacklisted"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// TargetFilter definisce i criteri di filtraggio e paginazione per l'elenco dei target
type TargetFilter struct {
	GroupID       *uuid.UUID
	Type          *TargetType
	IsBlacklisted *bool
	Search        string
	Limit         int
	Offset        int
}

// TargetRepository definisce l'interfaccia di persistenza per i target
type TargetRepository interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Target, error)
	List(ctx context.Context, tenantID uuid.UUID, filter TargetFilter) ([]*Target, int, error)
	Create(ctx context.Context, target *Target) error
	Update(ctx context.Context, target *Target) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	IsBlacklisted(ctx context.Context, tenantID uuid.UUID, value string) (bool, error)
}

// TargetGroupRepository definisce l'interfaccia di persistenza per i gruppi di target e la gestione dell'albero
type TargetGroupRepository interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*TargetGroup, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*TargetGroup, error)
	Create(ctx context.Context, group *TargetGroup) error
	Update(ctx context.Context, group *TargetGroup) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	GetTargetCountsByGroup(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]int, error)
	GetSeverityCountsByGroup(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]map[SeverityLevel]int, error)
}

