package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
)

// UpdateTenantBrandingInput racchiude i dati per aggiornare le impostazioni grafiche e anagrafiche del tenant
type UpdateTenantBrandingInput struct {
	TenantID    uuid.UUID          `json:"tenant_id"`
	Name        string             `json:"name"`
	LogoURL     *string            `json:"logo_url"`
	BrandColors domain.BrandColors `json:"brand_colors"`
}

// TenantService definisce i casi d'uso applicativi per la gestione dei tenant e del branding
type TenantService interface {
	GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error)
	UpdateBranding(ctx context.Context, input UpdateTenantBrandingInput) (*domain.Tenant, error)
}

type tenantService struct {
	tenantRepo domain.TenantRepository
}

// NewTenantService crea una nuova istanza di TenantService
func NewTenantService(tenantRepo domain.TenantRepository) TenantService {
	return &tenantService{
		tenantRepo: tenantRepo,
	}
}

// GetTenant recupera i metadati e il branding del tenant per ID
func (s *tenantService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	if tenantID == uuid.Nil {
		return nil, domain.ErrValidation
	}

	return s.tenantRepo.GetByID(ctx, tenantID)
}

// UpdateBranding valida e applica l'aggiornamento di ragione sociale, logo e colori
func (s *tenantService) UpdateBranding(ctx context.Context, input UpdateTenantBrandingInput) (*domain.Tenant, error) {
	if input.TenantID == uuid.Nil {
		return nil, domain.ErrValidation
	}

	cleanName := strings.TrimSpace(input.Name)
	if cleanName == "" || len(cleanName) > 255 {
		return nil, domain.ErrValidation
	}

	colors := input.BrandColors
	if strings.TrimSpace(colors.Primary) == "" &&
		strings.TrimSpace(colors.Secondary) == "" &&
		strings.TrimSpace(colors.Accent) == "" {
		colors = domain.DefaultBrandColors()
	} else {
		colors.Primary = strings.TrimSpace(colors.Primary)
		colors.Secondary = strings.TrimSpace(colors.Secondary)
		colors.Accent = strings.TrimSpace(colors.Accent)

		if !colors.IsValid() {
			return nil, domain.ErrValidation
		}
	}

	var logoURL *string
	if input.LogoURL != nil {
		cleanLogo := strings.TrimSpace(*input.LogoURL)
		if cleanLogo != "" {
			// Limite di sicurezza per prevenire payload base64 sproporzionati (> 2MB)
			if len(cleanLogo) > 2*1024*1024 {
				return nil, domain.ErrValidation
			}
			logoURL = &cleanLogo
		}
	}

	return s.tenantRepo.UpdateBranding(ctx, input.TenantID, cleanName, logoURL, colors)
}

