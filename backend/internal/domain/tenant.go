package domain

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var hexColorRegex = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

// IsValidHexColor verifica che una stringa sia un codice colore esadecimale valido (#RGB o #RRGGBB)
func IsValidHexColor(hex string) bool {
	return hexColorRegex.MatchString(strings.TrimSpace(hex))
}

// BrandColors rappresenta la palette colori personalizzata per i report white-label
type BrandColors struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
	Accent    string `json:"accent"`
}

// IsValid verifica che tutti i colori della palette siano codici esadecimali validi
func (c BrandColors) IsValid() bool {
	return IsValidHexColor(c.Primary) && IsValidHexColor(c.Secondary) && IsValidHexColor(c.Accent)
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

// TenantRepository definisce il contratto di persistenza per le entità Tenant
type TenantRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*Tenant, error)
	Create(ctx context.Context, tenant *Tenant) error
	UpdateBranding(ctx context.Context, id uuid.UUID, name string, logoURL *string, colors BrandColors) (*Tenant, error)
	List(ctx context.Context) ([]*Tenant, error)
}

