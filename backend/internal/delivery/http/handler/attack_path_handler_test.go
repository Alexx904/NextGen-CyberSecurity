package handler_test

import (
	"context"
	"encoding/json"
	"errors"
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
	"nextgen-cybersec/pkg/auth"
)

type mockAttackPathService struct {
	generateAttackGraphFunc func(ctx context.Context, tenantID uuid.UUID, minCVSS float64) (*domain.AttackGraph, error)
	getChokepointsFunc      func(ctx context.Context, tenantID uuid.UUID) ([]domain.Chokepoint, error)
}

func (m *mockAttackPathService) GenerateAttackGraph(ctx context.Context, tenantID uuid.UUID, minCVSS float64) (*domain.AttackGraph, error) {
	if m.generateAttackGraphFunc != nil {
		return m.generateAttackGraphFunc(ctx, tenantID, minCVSS)
	}
	return nil, nil
}

func (m *mockAttackPathService) GetChokepoints(ctx context.Context, tenantID uuid.UUID) ([]domain.Chokepoint, error) {
	if m.getChokepointsFunc != nil {
		return m.getChokepointsFunc(ctx, tenantID)
	}
	return nil, nil
}

func setupAttackPathRouter(h *handler.AttackPathHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/attack-paths/graph", h.GetGraph)
	r.Get("/attack-paths/chokepoints", h.GetChokepoints)
	return r
}

func withAttackPathTenantContext(r *http.Request, tenantID uuid.UUID) *http.Request {
	claims := &auth.UserClaims{
		UserID:    uuid.New(),
		TenantID:  tenantID,
		Role:      domain.RoleSecOps,
		Email:     "operator@test.local",
		TokenType: auth.TokenTypeAccess,
	}
	return r.WithContext(middleware.WithUserClaims(r.Context(), claims))
}

func TestAttackPathHandler_GetGraph_Success(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockAttackPathService{
		generateAttackGraphFunc: func(ctx context.Context, tid uuid.UUID, minCVSS float64) (*domain.AttackGraph, error) {
			if tid != tenantID {
				t.Errorf("expected tenantID %s, got %s", tenantID, tid)
			}
			if minCVSS != 6.5 {
				t.Errorf("expected minCVSS 6.5, got %f", minCVSS)
			}
			return &domain.AttackGraph{
				TenantID:    tid,
				GeneratedAt: time.Now().UTC(),
				Nodes: []domain.AttackNode{
					{ID: "entry-internet", Type: domain.NodeTypeEntrypoint, Label: "Internet"},
				},
				Summary: domain.AttackGraphSummary{
					TotalNodes: 1,
				},
			}, nil
		},
	}

	h := handler.NewAttackPathHandler(mockSvc)
	router := setupAttackPathRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/attack-paths/graph?min_cvss=6.5", nil)
	req = withAttackPathTenantContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Success bool                `json:"success"`
		Data    *domain.AttackGraph `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success || resp.Data == nil {
		t.Fatal("expected success response with non-nil data")
	}
	if resp.Data.TenantID != tenantID {
		t.Errorf("expected tenantID %s in graph, got %s", tenantID, resp.Data.TenantID)
	}
	if len(resp.Data.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(resp.Data.Nodes))
	}
}

func TestAttackPathHandler_GetGraph_Unauthorized(t *testing.T) {
	mockSvc := &mockAttackPathService{}
	h := handler.NewAttackPathHandler(mockSvc)
	router := setupAttackPathRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/attack-paths/graph", nil)
	// Nessun tenant context iniettato
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	var resp response.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Success {
		t.Error("expected success=false for unauthorized request")
	}
}

func TestAttackPathHandler_InvalidCVSS(t *testing.T) {
	for _, value := range []string{"NaN", "Inf", "11", "-1", "abc"} {
		t.Run(value, func(t *testing.T) {
			svc := &mockAttackPathService{generateAttackGraphFunc: func(context.Context, uuid.UUID, float64) (*domain.AttackGraph, error) {
				t.Fatal("invalid filter reached service")
				return nil, nil
			}}
			router := setupAttackPathRouter(handler.NewAttackPathHandler(svc))
			req := withAttackPathTenantContext(httptest.NewRequest(http.MethodGet, "/attack-paths/graph?min_cvss="+value, nil), uuid.New())
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", rec.Code)
			}
		})
	}
}

func TestAttackPathHandler_GetGraph_ServiceError(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockAttackPathService{
		generateAttackGraphFunc: func(ctx context.Context, tid uuid.UUID, minCVSS float64) (*domain.AttackGraph, error) {
			return nil, errors.New("database connection failed")
		},
	}

	h := handler.NewAttackPathHandler(mockSvc)
	router := setupAttackPathRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/attack-paths/graph", nil)
	req = withAttackPathTenantContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	var resp response.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Success {
		t.Error("expected success=false for service error")
	}
	if resp.Error == nil || resp.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR code, got %+v", resp.Error)
	}
}

func TestAttackPathHandler_GetChokepoints_Success(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockAttackPathService{
		getChokepointsFunc: func(ctx context.Context, tid uuid.UUID) ([]domain.Chokepoint, error) {
			return []domain.Chokepoint{
				{
					NodeID:         "host-192-168-1-10",
					Label:          "srv-web-dmz",
					MitigatedPaths: 3,
					MitigatedRisk:  85.0,
					Remediation:    "Isolare host DMZ",
				},
			}, nil
		},
	}

	h := handler.NewAttackPathHandler(mockSvc)
	router := setupAttackPathRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/attack-paths/chokepoints", nil)
	req = withAttackPathTenantContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Success bool                `json:"success"`
		Data    []domain.Chokepoint `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Fatal("expected success=true")
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 chokepoint, got %d", len(resp.Data))
	}
	if resp.Data[0].NodeID != "host-192-168-1-10" {
		t.Errorf("expected node ID host-192-168-1-10, got %s", resp.Data[0].NodeID)
	}
}

func TestAttackPathHandler_GetChokepoints_Empty(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockAttackPathService{
		getChokepointsFunc: func(ctx context.Context, tid uuid.UUID) ([]domain.Chokepoint, error) {
			return nil, nil // nil slice -> handler serializza come []
		},
	}

	h := handler.NewAttackPathHandler(mockSvc)
	router := setupAttackPathRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/attack-paths/chokepoints", nil)
	req = withAttackPathTenantContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Success bool                `json:"success"`
		Data    []domain.Chokepoint `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Fatal("expected success=true")
	}
	if resp.Data == nil || len(resp.Data) != 0 {
		t.Errorf("expected empty non-nil slice, got %v", resp.Data)
	}
}

func TestAttackPathHandler_GetChokepoints_Unauthorized(t *testing.T) {
	mockSvc := &mockAttackPathService{}
	h := handler.NewAttackPathHandler(mockSvc)
	router := setupAttackPathRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/attack-paths/chokepoints", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
