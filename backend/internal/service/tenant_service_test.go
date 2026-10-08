package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

type mockTenantRepo struct {
	getByIDFunc        func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	getBySlugFunc      func(ctx context.Context, slug string) (*domain.Tenant, error)
	createFunc         func(ctx context.Context, tenant *domain.Tenant) error
	updateBrandingFunc func(ctx context.Context, id uuid.UUID, name string, logoURL *string, colors domain.BrandColors) (*domain.Tenant, error)
	listFunc           func(ctx context.Context) ([]*domain.Tenant, error)
}

func (m *mockTenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *mockTenantRepo) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	if m.getBySlugFunc != nil {
		return m.getBySlugFunc(ctx, slug)
	}
	return nil, domain.ErrNotFound
}

func (m *mockTenantRepo) Create(ctx context.Context, tenant *domain.Tenant) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, tenant)
	}
	return nil
}

func (m *mockTenantRepo) UpdateBranding(ctx context.Context, id uuid.UUID, name string, logoURL *string, colors domain.BrandColors) (*domain.Tenant, error) {
	if m.updateBrandingFunc != nil {
		return m.updateBrandingFunc(ctx, id, name, logoURL, colors)
	}
	return nil, domain.ErrNotFound
}

func (m *mockTenantRepo) List(ctx context.Context) ([]*domain.Tenant, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, nil
}

func TestTenantService_GetTenant_Success(t *testing.T) {
	tenantID := uuid.New()
	expectedTenant := &domain.Tenant{
		ID:          tenantID,
		Name:        "Test Org",
		Slug:        "test-org",
		BrandColors: domain.DefaultBrandColors(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	repo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			if id == tenantID {
				return expectedTenant, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	svc := service.NewTenantService(repo)
	tenant, err := svc.GetTenant(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("atteso nessun errore, ottenuto: %v", err)
	}
	if tenant.ID != tenantID || tenant.Name != "Test Org" {
		t.Errorf("dati tenant non corrispondenti: %+v", tenant)
	}
}

func TestTenantService_GetTenant_NilID(t *testing.T) {
	svc := service.NewTenantService(&mockTenantRepo{})
	_, err := svc.GetTenant(context.Background(), uuid.Nil)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("atteso ErrValidation, ottenuto: %v", err)
	}
}

func TestTenantService_GetTenant_NotFound(t *testing.T) {
	svc := service.NewTenantService(&mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return nil, domain.ErrNotFound
		},
	})
	_, err := svc.GetTenant(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("atteso ErrNotFound, ottenuto: %v", err)
	}
}

func TestTenantService_UpdateBranding_Success(t *testing.T) {
	tenantID := uuid.New()
	logo := "https://example.com/logo.png"
	colors := domain.BrandColors{
		Primary:   "#112233",
		Secondary: "#445566",
		Accent:    "#778899",
	}

	repo := &mockTenantRepo{
		updateBrandingFunc: func(ctx context.Context, id uuid.UUID, name string, logoURL *string, c domain.BrandColors) (*domain.Tenant, error) {
			return &domain.Tenant{
				ID:          id,
				Name:        name,
				Slug:        "test-slug",
				LogoURL:     logoURL,
				BrandColors: c,
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	svc := service.NewTenantService(repo)
	updated, err := svc.UpdateBranding(context.Background(), service.UpdateTenantBrandingInput{
		TenantID:    tenantID,
		Name:        "Nuova Ragione Sociale",
		LogoURL:     &logo,
		BrandColors: colors,
	})
	if err != nil {
		t.Fatalf("atteso successo, ottenuto errore: %v", err)
	}
	if updated.Name != "Nuova Ragione Sociale" || *updated.LogoURL != logo || updated.BrandColors != colors {
		t.Errorf("dati aggiornati non corretti: %+v", updated)
	}
}

func TestTenantService_UpdateBranding_DefaultColorsFallback(t *testing.T) {
	tenantID := uuid.New()
	repo := &mockTenantRepo{
		updateBrandingFunc: func(ctx context.Context, id uuid.UUID, name string, logoURL *string, c domain.BrandColors) (*domain.Tenant, error) {
			return &domain.Tenant{
				ID:          id,
				Name:        name,
				BrandColors: c,
			}, nil
		},
	}

	svc := service.NewTenantService(repo)
	updated, err := svc.UpdateBranding(context.Background(), service.UpdateTenantBrandingInput{
		TenantID:    tenantID,
		Name:        "Org Name",
		BrandColors: domain.BrandColors{}, // Vuota
	})
	if err != nil {
		t.Fatalf("atteso successo con fallback, ottenuto errore: %v", err)
	}
	if updated.BrandColors != domain.DefaultBrandColors() {
		t.Errorf("attesi colori di default, ottenuto: %+v", updated.BrandColors)
	}
}

func TestTenantService_UpdateBranding_ValidationErrors(t *testing.T) {
	svc := service.NewTenantService(&mockTenantRepo{})

	// 1. Nil TenantID
	_, err := svc.UpdateBranding(context.Background(), service.UpdateTenantBrandingInput{
		TenantID: uuid.Nil,
		Name:     "Valid Name",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("atteso ErrValidation per Nil TenantID, ottenuto: %v", err)
	}

	// 2. Empty Name
	_, err = svc.UpdateBranding(context.Background(), service.UpdateTenantBrandingInput{
		TenantID: uuid.New(),
		Name:     "   ",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("atteso ErrValidation per Name vuoto, ottenuto: %v", err)
	}

	// 3. Invalid Color
	_, err = svc.UpdateBranding(context.Background(), service.UpdateTenantBrandingInput{
		TenantID: uuid.New(),
		Name:     "Valid Name",
		BrandColors: domain.BrandColors{
			Primary:   "not-a-color",
			Secondary: "#112233",
			Accent:    "#445566",
		},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("atteso ErrValidation per colore primario non valido, ottenuto: %v", err)
	}

	// 4. Logo troppo grande (> 2MB)
	hugeLogo := "data:image/png;base64," + strings.Repeat("A", 3*1024*1024)
	_, err = svc.UpdateBranding(context.Background(), service.UpdateTenantBrandingInput{
		TenantID: uuid.New(),
		Name:     "Valid Name",
		LogoURL:  &hugeLogo,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("atteso ErrValidation per logo troppo grande, ottenuto: %v", err)
	}
}

