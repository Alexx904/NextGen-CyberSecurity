package handler

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
	"strings"
	"testing"
)

type stubDefense struct {
	service.DefenseService
	called bool
	err    error
}

func (s *stubDefense) Remediate(_ context.Context, tenant uuid.UUID, node string, req domain.RemediationRequest) (*domain.RemediationResult, error) {
	s.called = true
	return &domain.RemediationResult{NodeID: node, ServerID: req.ServerID, SafetyTimeoutSeconds: 60}, s.err
}
func TestDefenseHandlerGuards(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		role       domain.UserRole
		tenant     bool
		err        error
		status     int
		called     bool
	}{
		{name: "unauthenticated", body: `{}`, status: 401},
		{name: "auditor", body: `{}`, tenant: true, role: domain.RoleAuditor, status: 403},
		{name: "unknown fields", body: `{"command":"id"}`, tenant: true, role: domain.RoleSecOps, status: 400},
		{name: "trailing JSON", body: `{} {}`, tenant: true, role: domain.RoleSecOps, status: 400},
		{name: "malformed", body: `{`, tenant: true, role: domain.RoleSecOps, status: 400},
		{name: "success", body: `{"server_id":"` + uuid.NewString() + `"}`, tenant: true, role: domain.RoleSecOps, status: 200, called: true},
		{name: "not found", body: `{}`, tenant: true, role: domain.RoleSecOps, err: domain.ErrNotFound, status: 404, called: true},
		{name: "validation", body: `{}`, tenant: true, role: domain.RoleSecOps, err: domain.ErrValidation, status: 400, called: true},
		{name: "conflict", body: `{}`, tenant: true, role: domain.RoleSecOps, err: domain.ErrSafetyTimerActive, status: 409, called: true},
		{name: "remote failure", body: `{}`, tenant: true, role: domain.RoleSecOps, err: domain.ErrSSHExecutionFailed, status: 502, called: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &stubDefense{err: tc.err}
			h := NewDefenseHandler(svc)
			r := chi.NewRouter()
			r.With(middleware.RequireSecOps()).Post("/chokepoints/{id}/remediate", h.Remediate)
			req := httptest.NewRequest(http.MethodPost, "/chokepoints/svc/remediate", strings.NewReader(tc.body))
			if tc.tenant {
				req = req.WithContext(middleware.WithUserClaims(req.Context(), &auth.UserClaims{TenantID: uuid.New(), Role: tc.role}))
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status || svc.called != tc.called {
				t.Fatalf("status %d body %s called %v", w.Code, w.Body, svc.called)
			}
		})
	}
}
