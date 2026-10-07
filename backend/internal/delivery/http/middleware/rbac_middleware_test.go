package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/auth"
)

func createAuthenticatedContext(role domain.UserRole) context.Context {
	claims := &auth.UserClaims{
		UserID:    uuid.New(),
		TenantID:  uuid.New(),
		Role:      role,
		Email:     "user@test.local",
		TokenType: auth.TokenTypeAccess,
	}
	return middleware.WithUserClaims(context.Background(), claims)
}

func TestRBAC_RequireSuperAdmin(t *testing.T) {
	r := chi.NewRouter()
	r.Use(middleware.RequireSuperAdmin())
	r.Get("/superadmin-only", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, "welcome superadmin")
	})

	// SuperAdmin: ammissibile
	req := httptest.NewRequest(http.MethodGet, "/superadmin-only", nil)
	req = req.WithContext(createAuthenticatedContext(domain.RoleSuperAdmin))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("SuperAdmin respinto: ottenuto %d", rec.Code)
	}

	// TenantAdmin: respinto con 403
	req2 := httptest.NewRequest(http.MethodGet, "/superadmin-only", nil)
	req2 = req2.WithContext(createAuthenticatedContext(domain.RoleTenantAdmin))
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Errorf("TenantAdmin non respinto con 403: ottenuto %d", rec2.Code)
	}
}

func TestRBAC_RequireAdmin(t *testing.T) {
	r := chi.NewRouter()
	r.Use(middleware.RequireAdmin())
	r.Get("/admin-only", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, "welcome admin")
	})

	roles := []struct {
		role           domain.UserRole
		expectedStatus int
	}{
		{domain.RoleSuperAdmin, http.StatusOK},
		{domain.RoleTenantAdmin, http.StatusOK},
		{domain.RoleSecOps, http.StatusForbidden},
		{domain.RoleAuditor, http.StatusForbidden},
	}

	for _, tc := range roles {
		req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
		req = req.WithContext(createAuthenticatedContext(tc.role))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != tc.expectedStatus {
			t.Errorf("Ruolo %s: atteso status %d, ottenuto %d", tc.role, tc.expectedStatus, rec.Code)
		}
	}
}

func TestRBAC_RequireSecOps(t *testing.T) {
	r := chi.NewRouter()
	r.Use(middleware.RequireSecOps())
	r.Get("/secops-scans", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, "scan operations permitted")
	})

	roles := []struct {
		role           domain.UserRole
		expectedStatus int
	}{
		{domain.RoleSuperAdmin, http.StatusOK},
		{domain.RoleTenantAdmin, http.StatusOK},
		{domain.RoleSecOps, http.StatusOK},
		{domain.RoleAuditor, http.StatusForbidden},
	}

	for _, tc := range roles {
		req := httptest.NewRequest(http.MethodGet, "/secops-scans", nil)
		req = req.WithContext(createAuthenticatedContext(tc.role))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != tc.expectedStatus {
			t.Errorf("Ruolo %s: atteso status %d, ottenuto %d", tc.role, tc.expectedStatus, rec.Code)
		}
	}
}

func TestRBAC_MissingAuthContext(t *testing.T) {
	r := chi.NewRouter()
	r.Use(middleware.RequireSecOps())
	r.Get("/unauthenticated", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/unauthenticated", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Atteso 401 per richiesta senza contesto di autenticazione, ottenuto %d", rec.Code)
	}
}

