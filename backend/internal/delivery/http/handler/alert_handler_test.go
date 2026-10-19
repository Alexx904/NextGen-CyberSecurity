package handler

import (
	"context"
	"encoding/json"
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

type handlerAlertRepo struct {
	domain.AlertRepository
	channel *domain.AlertChannel
}

func (r *handlerAlertRepo) SaveChannel(_ context.Context, c *domain.AlertChannel, _ bool) error {
	copy := *c
	r.channel = &copy
	return nil
}
func (r *handlerAlertRepo) ListChannels(context.Context, uuid.UUID) ([]domain.AlertChannel, error) {
	if r.channel == nil {
		return []domain.AlertChannel{}, nil
	}
	return []domain.AlertChannel{*r.channel}, nil
}

type handlerAlertDispatcher struct{ calls int }

func (d *handlerAlertDispatcher) Send(context.Context, domain.AlertChannelKind, domain.AlertChannelConfig, domain.AlertEvent) service.AlertSendResult {
	d.calls++
	return service.AlertSendResult{Code: "ACCEPTED"}
}

func TestAlertHandlerRoleAndSecretBoundaries(t *testing.T) {
	secret := strings.Repeat("synthetic-test-key", 3)
	config := `{"endpoint":"https://example.com/qa-only?synthetic-secret","hmac_secret":"` + secret + `"}`
	for _, tc := range []struct {
		name, path, body string
		role             domain.UserRole
		status           int
		calls            int
	}{
		{name: "unauthenticated", path: "/", body: `{}`, status: 401},
		{name: "auditor", path: "/", body: `{}`, role: domain.RoleAuditor, status: 403},
		{name: "secops destination forbidden", path: "/", body: `{}`, role: domain.RoleSecOps, status: 403},
		{name: "unknown secret field", path: "/", body: `{"secret":"private"}`, role: domain.RoleTenantAdmin, status: 400},
		{name: "trailing JSON", path: "/", body: `{} {}`, role: domain.RoleTenantAdmin, status: 400},
		{name: "save not send", path: "/", body: `{"name":"QA","kind":"webhook","enabled":true,"events":["finding.critical"],"config":` + config + `}`, role: domain.RoleTenantAdmin, status: 201},
		{name: "unacknowledged test", path: "/test", body: `{"kind":"webhook","config":` + config + `}`, role: domain.RoleTenantAdmin, status: 400},
		{name: "explicit synthetic test", path: "/test", body: `{"kind":"webhook","config":` + config + `,"confirmed_external_send":true}`, role: domain.RoleTenantAdmin, status: 200, calls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &handlerAlertRepo{}
			dispatcher := &handlerAlertDispatcher{}
			h := NewAlertHandler(service.NewAlertService(repo, dispatcher, []byte("test-only-master-key-32-bytes-or-longer"), nil))
			r := chi.NewRouter()
			r.Use(middleware.RequireTenant(), middleware.RequireAdmin())
			r.Post("/", h.Create)
			r.Post("/test", h.Test)
			req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
			if tc.role != "" {
				req = req.WithContext(middleware.WithUserClaims(req.Context(), &auth.UserClaims{TenantID: uuid.New(), Role: tc.role}))
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status || dispatcher.calls != tc.calls {
				t.Fatalf("status %d calls %d body %s", w.Code, dispatcher.calls, w.Body)
			}
			if strings.Contains(w.Body.String(), secret) || strings.Contains(w.Body.String(), "synthetic-secret") || strings.Contains(w.Body.String(), "encrypted_config") {
				t.Fatal("secret leaked")
			}
			if tc.status == 201 {
				if repo.channel == nil || repo.channel.EncryptedConfig == "" {
					t.Fatal("config not encrypted")
				}
				var body map[string]any
				if json.Unmarshal(w.Body.Bytes(), &body) != nil {
					t.Fatal("invalid public JSON")
				}
			}
		})
	}
}
