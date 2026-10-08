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
	"nextgen-cybersec/pkg/auth"
)

type mockScanService struct {
	cancelScanFunc      func(ctx context.Context, tenantID, scanID uuid.UUID) error
	getScanFunc         func(ctx context.Context, tenantID, scanID uuid.UUID) (*domain.Scan, error)
	listScansFunc       func(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Scan, int, error)
	startScanFunc       func(ctx context.Context, tenantID, targetID, profileID uuid.UUID) (*domain.Scan, error)
	getScanFindingsFunc func(ctx context.Context, tenantID, scanID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error)
	listProfilesFunc    func(ctx context.Context) ([]*domain.ScanProfile, error)
}

func (m *mockScanService) CancelScan(ctx context.Context, tenantID, scanID uuid.UUID) error {
	if m.cancelScanFunc != nil {
		return m.cancelScanFunc(ctx, tenantID, scanID)
	}
	return nil
}

func (m *mockScanService) GetScan(ctx context.Context, tenantID, scanID uuid.UUID) (*domain.Scan, error) {
	if m.getScanFunc != nil {
		return m.getScanFunc(ctx, tenantID, scanID)
	}
	return nil, nil
}

func (m *mockScanService) ListScans(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Scan, int, error) {
	if m.listScansFunc != nil {
		return m.listScansFunc(ctx, tenantID, limit, offset)
	}
	return nil, 0, nil
}

func (m *mockScanService) StartScan(ctx context.Context, tenantID, targetID, profileID uuid.UUID) (*domain.Scan, error) {
	if m.startScanFunc != nil {
		return m.startScanFunc(ctx, tenantID, targetID, profileID)
	}
	return nil, nil
}

func (m *mockScanService) GetScanFindings(ctx context.Context, tenantID, scanID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
	if m.getScanFindingsFunc != nil {
		return m.getScanFindingsFunc(ctx, tenantID, scanID, filter)
	}
	return nil, 0, nil
}

func (m *mockScanService) ListProfiles(ctx context.Context) ([]*domain.ScanProfile, error) {
	if m.listProfilesFunc != nil {
		return m.listProfilesFunc(ctx)
	}
	return nil, nil
}

func setupScanRouter(h *handler.ScanHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/scans", h.List)
	r.Get("/scans/{id}", h.GetByID)
	r.Get("/scans/{id}/findings", h.GetFindings)
	r.Post("/scans", h.Start)
	r.Post("/scans/{id}/cancel", h.Cancel)
	r.Get("/scan-profiles", h.ListProfiles)
	return r
}

func withSecOpsContext(r *http.Request, tenantID uuid.UUID) *http.Request {
	claims := &auth.UserClaims{
		UserID:    uuid.New(),
		TenantID:  tenantID,
		Role:      domain.RoleSecOps,
		Email:     "secops@test.local",
		TokenType: auth.TokenTypeAccess,
	}
	return r.WithContext(middleware.WithUserClaims(r.Context(), claims))
}

func TestScanHandler_Cancel_Success(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	mockSvc := &mockScanService{
		cancelScanFunc: func(ctx context.Context, tID, sID uuid.UUID) error {
			if tID != tenantID || sID != scanID {
				t.Errorf("parametri errati: tID=%v, sID=%v", tID, sID)
			}
			return nil
		},
	}

	h := handler.NewScanHandler(mockSvc)
	router := setupScanRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/scans/"+scanID.String()+"/cancel", nil)
	req = withSecOpsContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d: %s", rec.Code, rec.Body.String())
	}

	var env response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal response fallito: %v", err)
	}

	if !env.Success {
		t.Errorf("atteso success true, ottenuto false")
	}
}

func TestScanHandler_Cancel_InvalidID(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockScanService{}
	h := handler.NewScanHandler(mockSvc)
	router := setupScanRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/scans/not-a-valid-uuid/cancel", nil)
	req = withSecOpsContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("atteso status 400 Bad Request, ottenuto %d", rec.Code)
	}
}

func TestScanHandler_Cancel_Unauthorized(t *testing.T) {
	scanID := uuid.New()
	mockSvc := &mockScanService{}
	h := handler.NewScanHandler(mockSvc)
	router := setupScanRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/scans/"+scanID.String()+"/cancel", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("atteso status 401 Unauthorized, ottenuto %d", rec.Code)
	}
}

func TestScanHandler_Cancel_NotFound(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	mockSvc := &mockScanService{
		cancelScanFunc: func(ctx context.Context, tID, sID uuid.UUID) error {
			return domain.ErrNotFound
		},
	}

	h := handler.NewScanHandler(mockSvc)
	router := setupScanRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/scans/"+scanID.String()+"/cancel", nil)
	req = withSecOpsContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("atteso status 404 Not Found, ottenuto %d", rec.Code)
	}
}

func TestScanHandler_Cancel_Conflict(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	mockSvc := &mockScanService{
		cancelScanFunc: func(ctx context.Context, tID, sID uuid.UUID) error {
			return domain.ErrScanCannotBeCancelled
		},
	}

	h := handler.NewScanHandler(mockSvc)
	router := setupScanRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/scans/"+scanID.String()+"/cancel", nil)
	req = withSecOpsContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("atteso status 409 Conflict, ottenuto %d", rec.Code)
	}
}

func TestScanHandler_Start_Success(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()
	profileID := uuid.New()
	scanID := uuid.New()

	mockSvc := &mockScanService{
		startScanFunc: func(ctx context.Context, tID, tgID, pID uuid.UUID) (*domain.Scan, error) {
			return &domain.Scan{
				ID:          scanID,
				TenantID:    tID,
				TargetID:    tgID,
				ProfileID:   pID,
				Status:      domain.ScanStatusQueued,
				TriggerType: domain.TriggerManual,
				CreatedAt:   time.Now(),
			}, nil
		},
	}

	h := handler.NewScanHandler(mockSvc)
	router := setupScanRouter(h)

	payload := map[string]interface{}{
		"target_id":  targetID.String(),
		"profile_id": profileID.String(),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/scans", bytes.NewReader(body))
	req = withSecOpsContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("atteso status 202 Accepted, ottenuto %d: %s", rec.Code, rec.Body.String())
	}
}

func TestScanHandler_Start_InvalidPayload(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockScanService{}
	h := handler.NewScanHandler(mockSvc)
	router := setupScanRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/scans", bytes.NewReader([]byte(`{"target_id": ""}`)))
	req = withSecOpsContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("atteso status 400 Bad Request, ottenuto %d", rec.Code)
	}
}

func TestScanHandler_GetByID_Success(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	mockSvc := &mockScanService{
		getScanFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.Scan, error) {
			return &domain.Scan{
				ID:        scanID,
				TenantID:  tenantID,
				Status:    domain.ScanStatusRunning,
				CreatedAt: time.Now(),
			}, nil
		},
	}

	h := handler.NewScanHandler(mockSvc)
	router := setupScanRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String(), nil)
	req = withSecOpsContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d", rec.Code)
	}
}

func TestScanHandler_List_Success(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockScanService{
		listScansFunc: func(ctx context.Context, tID uuid.UUID, limit, offset int) ([]*domain.Scan, int, error) {
			return []*domain.Scan{
				{ID: uuid.New(), TenantID: tID, Status: domain.ScanStatusCompleted},
			}, 1, nil
		},
	}

	h := handler.NewScanHandler(mockSvc)
	router := setupScanRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans?limit=10&offset=0", nil)
	req = withSecOpsContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d", rec.Code)
	}
}

func TestScanHandler_GetFindings_Success(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	serviceName := "http"
	mockSvc := &mockScanService{
		getScanFindingsFunc: func(ctx context.Context, tID, sID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
			return []*domain.Finding{
				{
					ID:          uuid.New(),
					ScanID:      sID,
					TenantID:    tID,
					HostIP:      "192.168.1.10",
					Port:        80,
					Protocol:    "tcp",
					ServiceName: &serviceName,
					Severity:    domain.SeverityInfo,
				},
			}, 1, nil
		},
	}

	h := handler.NewScanHandler(mockSvc)
	router := setupScanRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/findings?severity=INFO", nil)
	req = withSecOpsContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d", rec.Code)
	}
}

func TestScanHandler_ListProfiles_Success(t *testing.T) {
	mockSvc := &mockScanService{
		listProfilesFunc: func(ctx context.Context) ([]*domain.ScanProfile, error) {
			return []*domain.ScanProfile{
				{ID: uuid.New(), Name: "Fast Discovery", NmapArgs: "-T4 -F", TimeoutSeconds: 900},
			}, nil
		},
	}

	h := handler.NewScanHandler(mockSvc)
	router := setupScanRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scan-profiles", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d", rec.Code)
	}
}
