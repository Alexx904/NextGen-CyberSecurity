package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
)

type handlerPatchReader struct {
	tenant, id uuid.UUID
	err        error
	foreign    bool
}

func (s *handlerPatchReader) GetByID(_ context.Context, tenant, id uuid.UUID) (*domain.Finding, error) {
	s.tenant, s.id = tenant, id
	if s.foreign {
		tenant = uuid.New()
	}
	return &domain.Finding{ID: id, TenantID: tenant}, s.err
}

func TestVirtualPatchingHTTPBoundary(t *testing.T) {
	tenant, id := uuid.New(), uuid.New()
	for _, tc := range []struct {
		name, path             string
		authenticated, foreign bool
		err                    error
		status                 int
	}{
		{name: "no tenant", path: id.String(), status: 401},
		{name: "invalid uuid", path: "bad", authenticated: true, status: 400},
		{name: "nil uuid", path: uuid.Nil.String(), authenticated: true, status: 400},
		{name: "foreign tenant", path: id.String(), authenticated: true, foreign: true, status: 404},
		{name: "missing", path: id.String(), authenticated: true, err: domain.ErrNotFound, status: 404},
		{name: "internal redaction", path: id.String(), authenticated: true, err: errors.New("private diagnostic details"), status: 500},
		{name: "auditor advisory read", path: id.String(), authenticated: true, status: 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := &handlerPatchReader{err: tc.err, foreign: tc.foreign}
			h := NewVirtualPatchingHandler(service.NewVirtualPatchingService(reader))
			r := chi.NewRouter()
			r.Get("/findings/{id}/virtual-patches", h.Generate)
			req := httptest.NewRequest(http.MethodGet, "/findings/"+tc.path+"/virtual-patches", nil)
			if tc.authenticated {
				req = req.WithContext(middleware.WithUserClaims(req.Context(), &auth.UserClaims{TenantID: tenant, Role: domain.RoleAuditor}))
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status || strings.Contains(w.Body.String(), "private diagnostic") {
				t.Fatalf("status %d body %s", w.Code, w.Body)
			}
			if tc.status == 200 && (reader.tenant != tenant || reader.id != id || w.Header().Get("Cache-Control") != "no-store" || !strings.Contains(w.Body.String(), `"files":[]`)) {
				t.Fatal("invalid tenant read or empty contract", w.Body)
			}
		})
	}
}
