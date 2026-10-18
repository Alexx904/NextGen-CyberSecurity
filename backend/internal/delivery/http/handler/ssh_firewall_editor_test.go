package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/delivery/http/handler"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
)

type handlerFirewallEditor struct {
	service.FirewallService
	service.FirewallRuleEditor
	hit func(string, uuid.UUID, uuid.UUID, string, map[string]any) (*domain.FirewallStatus, error)
}

func (m *handlerFirewallEditor) SetPolicy(_ context.Context, t, s uuid.UUID, c domain.ChainName, p domain.RuleAction, r string, timeout int) (*domain.FirewallStatus, error) {
	return m.hit("policy", t, s, r, map[string]any{"chain": c, "policy": p, "timeout": timeout})
}
func (m *handlerFirewallEditor) ReplaceRule(_ context.Context, t, s uuid.UUID, line int, p domain.FirewallRuleParams, r string, timeout int) (*domain.FirewallStatus, error) {
	return m.hit("replace", t, s, r, map[string]any{"line": line, "params": p, "timeout": timeout})
}
func (m *handlerFirewallEditor) MoveRule(_ context.Context, t, s uuid.UUID, c domain.ChainName, from, to int, r string, timeout int) (*domain.FirewallStatus, error) {
	return m.hit("move", t, s, r, map[string]any{"chain": c, "from": from, "to": to, "timeout": timeout})
}
func (m *handlerFirewallEditor) AddRuleWithRevision(_ context.Context, t, s uuid.UUID, p domain.FirewallRuleParams, r string, timeout int) (*domain.FirewallStatus, error) {
	return m.hit("add", t, s, r, map[string]any{"params": p, "timeout": timeout})
}
func (m *handlerFirewallEditor) DeleteRuleWithRevision(_ context.Context, t, s uuid.UUID, c domain.ChainName, line int, r string, timeout int) (*domain.FirewallStatus, error) {
	return m.hit("delete", t, s, r, map[string]any{"chain": c, "line": line, "timeout": timeout})
}

func editorRouter(editor *handlerFirewallEditor) http.Handler {
	h := handler.NewSSHFirewallHandler(nil, editor)
	r := chi.NewRouter()
	r.Use(middleware.RequireSecOps())
	r.Put("/{id}/firewall/policy", h.SetFirewallPolicy)
	r.Put("/{id}/firewall/rules/{line}", h.ReplaceFirewallRule)
	r.Post("/{id}/firewall/rules/move", h.MoveFirewallRule)
	r.Post("/{id}/firewall/rules", h.AddFirewallRule)
	r.Delete("/{id}/firewall/rules/{line}", h.DeleteFirewallRule)
	return r
}

func TestSSHFirewallEditorContracts(t *testing.T) {
	tenant, server := uuid.New(), uuid.New()
	revision := strings.Repeat("a", 64)
	for _, tc := range []struct{ name, method, path, body string }{
		{"policy", "PUT", "/firewall/policy", `{"chain":"INPUT","policy":"DROP","expected_revision":"` + revision + `","safety_timeout_seconds":30}`},
		{"replace", "PUT", "/firewall/rules/2", `{"chain":"INPUT","action":"DROP","protocol":"tcp","port":"443","expected_revision":"` + revision + `","safety_timeout_seconds":30}`},
		{"move", "POST", "/firewall/rules/move", `{"chain":"INPUT","from":2,"to":1,"expected_revision":"` + revision + `","safety_timeout_seconds":30}`},
		{"add", "POST", "/firewall/rules", `{"chain":"INPUT","action":"DROP","protocol":"tcp","port":"443","expected_revision":"` + revision + `","safety_timeout_seconds":30}`},
		{"delete", "DELETE", "/firewall/rules/2?chain=INPUT&timeout=30&expected_revision=" + revision, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			editor := &handlerFirewallEditor{hit: func(action string, owner, id uuid.UUID, rev string, data map[string]any) (*domain.FirewallStatus, error) {
				called = true
				if action != tc.name || owner != tenant || id != server || rev != revision || data["timeout"] != 30 {
					t.Fatalf("wrong contract dispatch: %s %v", action, data)
				}
				if tc.name == "move" && (data["from"] != 2 || data["to"] != 1) {
					t.Fatal("move endpoints reversed")
				}
				if tc.name == "replace" || tc.name == "delete" {
					if data["line"] != 2 {
						t.Fatal("wrong replacement/deletion line")
					}
				}
				if params, ok := data["params"].(domain.FirewallRuleParams); ok && (params.Chain != domain.ChainInput || params.Action != domain.ActionDrop || params.Port != "443") {
					t.Fatal("rule fields lost")
				}
				return &domain.FirewallStatus{ServerID: server, Revision: revision}, nil
			}}
			req := httptest.NewRequest(tc.method, "/"+server.String()+tc.path, strings.NewReader(tc.body)).WithContext(setupSSHContext(tenant))
			w := httptest.NewRecorder()
			editorRouter(editor).ServeHTTP(w, req)
			if w.Code != http.StatusOK || !called {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestSSHFirewallEditorErrorsAndRoles(t *testing.T) {
	tenant, server := uuid.New(), uuid.New()
	calls := 0
	editor := &handlerFirewallEditor{hit: func(string, uuid.UUID, uuid.UUID, string, map[string]any) (*domain.FirewallStatus, error) {
		calls++
		return nil, fmt.Errorf("%w: internal-detail", domain.ErrFirewallStaleState)
	}}
	router := editorRouter(editor)
	for _, tc := range []struct {
		name, body string
		ctx        context.Context
		code       int
	}{
		{"unauthenticated", `{}`, context.Background(), 401},
		{"auditor", `{}`, middleware.WithUserClaims(context.Background(), &auth.UserClaims{TenantID: tenant, Role: domain.RoleAuditor}), 403},
		{"malformed", `{`, setupSSHContext(tenant), 400},
		{"trailing JSON", `{} {}`, setupSSHContext(tenant), 400},
		{"stale", `{"chain":"INPUT","policy":"DROP"}`, setupSSHContext(tenant), 409},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := calls
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", "/"+server.String()+"/firewall/policy", strings.NewReader(tc.body)).WithContext(tc.ctx)
			router.ServeHTTP(w, req)
			if w.Code != tc.code || strings.Contains(w.Body.String(), "internal-detail") {
				t.Fatalf("unsafe response: %d %s", w.Code, w.Body.String())
			}
			if tc.code != 409 && calls != before {
				t.Fatal("rejected request reached service")
			}
		})
	}
}
