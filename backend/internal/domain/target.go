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

// TargetGroup raggruppa logicamente un insieme di target (es. DMZ, Cloud, Uffici)
type TargetGroup struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TenantID    uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
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

// TargetGroupRepository definisce l'interfaccia di persistenza per i gruppi di target
type TargetGroupRepository interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*TargetGroup, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*TargetGroup, error)
	Create(ctx context.Context, group *TargetGroup) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}

