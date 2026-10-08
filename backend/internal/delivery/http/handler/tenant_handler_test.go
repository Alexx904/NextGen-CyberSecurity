package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/handler"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
)

type mockTenantService struct {
	getTenantFunc      func(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error)
	updateBrandingFunc func(ctx context.Context, input service.UpdateTenantBrandingInput) (*domain.Tenant, error)
}

func (m *mockTenantService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	if m.getTenantFunc != nil {
		return m.getTenantFunc(ctx, tenantID)
	}
	return nil, domain.ErrNotFound
}

func (m *mockTenantService) UpdateBranding(ctx context.Context, input service.UpdateTenantBrandingInput) (*domain.Tenant, error) {
	if m.updateBrandingFunc != nil {
		return m.updateBrandingFunc(ctx, input)
	}
	return nil, domain.ErrNotFound
}

func setupTenantRouter(h *handler.TenantHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/tenants/current", h.GetCurrent)
	r.Put("/tenants/current/branding", h.UpdateBranding)
	return r
}

func setupTenantRouterWithRBAC(h *handler.TenantHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/tenants/current", h.GetCurrent)
	r.Group(func(adminRouter chi.Router) {
		adminRouter.Use(middleware.RequireAdmin())
		adminRouter.Put("/tenants/current/branding", h.UpdateBranding)
	})
	return r
}

func withAdminClaims(r *http.Request, tenantID uuid.UUID) *http.Request {
	claims := &auth.UserClaims{
		UserID:    uuid.New(),
		TenantID:  tenantID,
		Role:      domain.RoleTenantAdmin,
		Email:     "admin@tenant.local",
		TokenType: auth.TokenTypeAccess,
	}
	return r.WithContext(middleware.WithUserClaims(r.Context(), claims))
}

func withAuditorClaims(r *http.Request, tenantID uuid.UUID) *http.Request {
	claims := &auth.UserClaims{
		UserID:    uuid.New(),
		TenantID:  tenantID,
		Role:      domain.RoleAuditor,
		Email:     "auditor@tenant.local",
		TokenType: auth.TokenTypeAccess,
	}
	return r.WithContext(middleware.WithUserClaims(r.Context(), claims))
}

func TestTenantHandler_GetCurrent_Success(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockTenantService{
		getTenantFunc: func(ctx context.Context, tID uuid.UUID) (*domain.Tenant, error) {
			if tID != tenantID {
				t.Fatalf("tenant ID errato: %v", tID)
			}
			return &domain.Tenant{
				ID:          tenantID,
				Name:        "Cyber Corp",
				Slug:        "cyber-corp",
				BrandColors: domain.DefaultBrandColors(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	h := handler.NewTenantHandler(mockSvc)
	router := setupTenantRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/tenants/current", nil)
	req = withAdminClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d: %s", rec.Code, rec.Body.String())
	}

	var env response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal risposta fallito: %v", err)
	}
	if !env.Success {
		t.Errorf("atteso success true, ottenuto false")
	}
}

func TestTenantHandler_GetCurrent_Unauthorized(t *testing.T) {
	mockSvc := &mockTenantService{}
	h := handler.NewTenantHandler(mockSvc)
	router := setupTenantRouter(h)

	// Richiesta senza claims / context
	req := httptest.NewRequest(http.MethodGet, "/tenants/current", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("atteso status 401 Unauthorized, ottenuto %d", rec.Code)
	}
}

func TestTenantHandler_GetCurrent_NotFound(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockTenantService{
		getTenantFunc: func(ctx context.Context, tID uuid.UUID) (*domain.Tenant, error) {
			return nil, domain.ErrNotFound
		},
	}

	h := handler.NewTenantHandler(mockSvc)
	router := setupTenantRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/tenants/current", nil)
	req = withAdminClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("atteso status 404 Not Found, ottenuto %d", rec.Code)
	}
}

func TestTenantHandler_UpdateBranding_Success(t *testing.T) {
	tenantID := uuid.New()
	logo := "https://example.com/logo.png"
	reqPayload := handler.UpdateBrandingRequest{
		Name:    "Ragione Sociale Aggiornata",
		LogoURL: &logo,
		BrandColors: &domain.BrandColors{
			Primary:   "#123456",
			Secondary: "#abcdef",
			Accent:    "#654321",
		},
	}

	mockSvc := &mockTenantService{
		updateBrandingFunc: func(ctx context.Context, input service.UpdateTenantBrandingInput) (*domain.Tenant, error) {
			if input.TenantID != tenantID {
				t.Fatalf("tenant ID errato: %v", input.TenantID)
			}
			return &domain.Tenant{
				ID:          tenantID,
				Name:        input.Name,
				Slug:        "cyber-corp",
				LogoURL:     input.LogoURL,
				BrandColors: input.BrandColors,
				UpdatedAt:   time.Now(),
			}, nil
		},
	}

	h := handler.NewTenantHandler(mockSvc)
	router := setupTenantRouter(h)

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest(http.MethodPut, "/tenants/current/branding", bytes.NewReader(body))
	req = withAdminClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d: %s", rec.Code, rec.Body.String())
	}

	var env response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal risposta fallito: %v", err)
	}
	if !env.Success {
		t.Errorf("atteso success true, ottenuto false")
	}
}

func TestTenantHandler_UpdateBranding_RBAC_Forbidden(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockTenantService{}
	h := handler.NewTenantHandler(mockSvc)
	router := setupTenantRouterWithRBAC(h)

	reqPayload := handler.UpdateBrandingRequest{
		Name: "Hacked Org",
	}
	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest(http.MethodPut, "/tenants/current/branding", bytes.NewReader(body))
	// Ruolo Auditor non autorizzato a modificare il branding
	req = withAuditorClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("atteso status 403 Forbidden per ruolo non amministratore, ottenuto %d", rec.Code)
	}
}

func TestTenantHandler_UpdateBranding_InvalidJSON(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockTenantService{}
	h := handler.NewTenantHandler(mockSvc)
	router := setupTenantRouter(h)

	req := httptest.NewRequest(http.MethodPut, "/tenants/current/branding", bytes.NewReader([]byte("{invalid-json")))
	req = withAdminClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("atteso status 400 Bad Request, ottenuto %d", rec.Code)
	}
}

func TestTenantHandler_UpdateBranding_ValidationError(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockTenantService{
		updateBrandingFunc: func(ctx context.Context, input service.UpdateTenantBrandingInput) (*domain.Tenant, error) {
			return nil, domain.ErrValidation
		},
	}

	h := handler.NewTenantHandler(mockSvc)
	router := setupTenantRouter(h)

	reqPayload := handler.UpdateBrandingRequest{
		Name: "",
	}
	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest(http.MethodPut, "/tenants/current/branding", bytes.NewReader(body))
	req = withAdminClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("atteso status 400 Bad Request per ErrValidation, ottenuto %d", rec.Code)
	}
}

