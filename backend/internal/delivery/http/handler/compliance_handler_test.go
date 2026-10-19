package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/delivery/http/handler"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
)

type complianceHandlerRepo struct {
	writes        int
	snapshotCalls int
	snapshotErr   error
	command       domain.IssueCommand
	scope         domain.ComplianceScope
	tenant        uuid.UUID
}

func (m *complianceHandlerRepo) Snapshot(_ context.Context, tenant uuid.UUID, scope domain.ComplianceScope) (*domain.ComplianceSnapshot, error) {
	m.snapshotCalls++
	m.scope = scope
	m.tenant = tenant
	return &domain.ComplianceSnapshot{Cutoff: time.Now().UTC(), Scope: scope}, m.snapshotErr
}
func (m *complianceHandlerRepo) CreateAsset(context.Context, uuid.UUID, domain.AssessmentAsset) (domain.AssessmentAsset, error) {
	m.writes++
	return domain.AssessmentAsset{}, nil
}
func (m *complianceHandlerRepo) ObserveFinding(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) (uuid.UUID, error) {
	m.writes++
	return uuid.New(), nil
}
func (m *complianceHandlerRepo) AppendCommand(_ context.Context, _ uuid.UUID, _ uuid.UUID, c domain.IssueCommand) error {
	m.writes++
	m.command = c
	return nil
}
func (m *complianceHandlerRepo) AppendEvidence(context.Context, uuid.UUID, domain.ControlEvidence) error {
	m.writes++
	return nil
}
func TestComplianceHTTPAuthScopeStrictDTOAndAudit(t *testing.T) {
	for _, role := range []domain.UserRole{domain.RoleAuditor, domain.RoleSecOps} {
		t.Run(string(role), func(t *testing.T) {
			repo := &complianceHandlerRepo{}
			h := handler.NewComplianceHandler(service.NewComplianceService(repo))
			tenant, actor, company := uuid.New(), uuid.New(), uuid.New()
			router := chi.NewRouter()
			router.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("X-Test-Auth") == "yes" {
						r = r.WithContext(middleware.WithUserClaims(r.Context(), &auth.UserClaims{TenantID: tenant, UserID: actor, Role: role}))
					}
					next.ServeHTTP(w, r)
				})
			})
			router.Use(middleware.RequireTenant())
			router.Get("/compliance", h.Snapshot)
			router.Post("/reports/audit", h.Audit)
			router.Group(func(r chi.Router) {
				r.Use(middleware.RequireSecOps())
				r.Post("/compliance/assets", h.CreateAsset)
				r.Post("/compliance/issues/{id}/events", h.Command)
			})
			request := func(method, path, body string, authenticated bool) *httptest.ResponseRecorder {
				t.Helper()
				r := httptest.NewRequest(method, path, strings.NewReader(body))
				if authenticated {
					r.Header.Set("X-Test-Auth", "yes")
				}
				w := httptest.NewRecorder()
				router.ServeHTTP(w, r)
				return w
			}
			if w := request("GET", "/compliance?company_id="+company.String(), "", false); w.Code != 401 {
				t.Fatal("unauthenticated read", w.Code)
			}
			if w := request("GET", "/compliance?company_id="+company.String(), "", true); w.Code != 200 || repo.scope.CompanyID != company || repo.tenant != tenant {
				t.Fatal("scope lost", w.Code)
			}
			if w := request("GET", "/compliance?company_id=bad", "", true); w.Code != 400 {
				t.Fatal("invalid UUID not rejected", w.Code)
			}
			repo.snapshotErr = domain.ErrNotFound
			if w := request("GET", "/compliance?company_id="+company.String(), "", true); w.Code != 404 {
				t.Fatal("scope error not mapped", w.Code)
			}
			repo.snapshotErr = nil
			body := `{"group_id":"` + company.String() + `","name":"test"}`
			if role == domain.RoleAuditor {
				if w := request("POST", "/compliance/assets", body, true); w.Code != 403 || repo.writes != 0 {
					t.Fatal("auditor wrote data", w.Code)
				}
			} else {
				if w := request("POST", "/compliance/assets", body+` {}`, true); w.Code != 400 || repo.writes != 0 {
					t.Fatal("trailing JSON accepted", w.Code)
				}
				if w := request("POST", "/compliance/assets", `{"tenant_id":"`+uuid.New().String()+`"}`, true); w.Code != 400 {
					t.Fatal("tenant override accepted", w.Code)
				}
				path := "/compliance/issues/" + uuid.New().String() + "/events"
				if w := request("POST", path, `{"action":"configure","expected_version":1,"reason":"assign","actor":"`+uuid.New().String()+`"}`, true); w.Code != 400 {
					t.Fatal("actor override accepted", w.Code)
				}
				if w := request("POST", path, `{"action":"configure","expected_version":1,"reason":"assign"}`, true); w.Code != 200 || repo.command.Actor != actor || !repo.command.At.IsZero() {
					t.Fatal("trusted actor or server timestamp boundary", w.Code, repo.command)
				}
			}
			before := repo.snapshotCalls
			if w := request("POST", "/reports/audit", `{"scope":{"company_id":"`+company.String()+`"},"metadata":{"legal_entity":""}}`, true); w.Code != 400 || repo.snapshotCalls != before {
				t.Fatal("invalid audit inputs fetched data", w.Code)
			}
			if w := request("POST", "/reports/audit", `{"scope":{"company_id":"`+company.String()+`"},"metadata":{"legal_entity":"Test legal entity","period":"October 2026","audit_date":"2026-10-19T00:00:00Z"}}`, true); w.Code != 200 || w.Header().Get("Content-Type") != "application/pdf" || !strings.HasPrefix(w.Body.String(), "%PDF-") || repo.snapshotCalls != before+1 {
				t.Fatal("audit must use exactly one authorized snapshot", w.Code)
			}
		})
	}
}
