package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// BrandColors rappresenta la palette colori personalizzata per i report white-label
type BrandColors struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
	Accent    string `json:"accent"`
}

// DefaultBrandColors restituisce la palette colori predefinita del sistema
func DefaultBrandColors() BrandColors {
	return BrandColors{
		Primary:   "#0f172a",
		Secondary: "#2563eb",
		Accent:    "#38bdf8",
	}
}

// Tenant rappresenta un'organizzazione o cliente isolato (Multi-Tenancy)
type Tenant struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Slug        string      `json:"slug" db:"slug"`
	LogoURL     *string     `json:"logo_url,omitempty" db:"logo_url"`
	BrandColors BrandColors `json:"brand_colors" db:"brand_colors"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// ReportTemplate rappresenta la configurazione personalizzata dei report per un tenant
type ReportTemplate struct {
	ID        uuid.UUID      `json:"id" db:"id"`
	TenantID  uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	Name      string         `json:"name" db:"name"`
	IsDefault bool           `json:"is_default" db:"is_default"`
	Config    map[string]any `json:"config" db:"config"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

// TenantRepository definisce il contratto di persistenza per le entità Tenant
type TenantRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*Tenant, error)
	Create(ctx context.Context, tenant *Tenant) error
	UpdateBranding(ctx context.Context, id uuid.UUID, logoURL *string, colors BrandColors) error
	List(ctx context.Context) ([]*Tenant, error)
}

