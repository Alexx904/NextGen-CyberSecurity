package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

type mockScanService struct {
	cancelScanFunc          func(ctx context.Context, tenantID, scanID uuid.UUID) error
	getScanFunc             func(ctx context.Context, tenantID, scanID uuid.UUID) (*domain.Scan, error)
	listScansFunc           func(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Scan, int, error)
	startScanFunc           func(ctx context.Context, tenantID, targetID, profileID uuid.UUID) (*domain.Scan, error)
	startGroupScanFunc      func(ctx context.Context, tenantID, groupID, profileID uuid.UUID) ([]*domain.Scan, error)
	getScanFindingsFunc     func(ctx context.Context, tenantID, scanID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error)
	listProfilesFunc        func(ctx context.Context, tenantID *uuid.UUID, category *domain.ScanProfileCategory) ([]*domain.ScanProfile, error)
	createCustomProfileFunc func(ctx context.Context, tenantID uuid.UUID, name, nmapArgs string, description *string, timeoutSeconds int, category domain.ScanProfileCategory) (*domain.ScanProfile, error)
	deleteCustomProfileFunc func(ctx context.Context, tenantID, id uuid.UUID) error
	getScanDiffFunc         func(ctx context.Context, tenantID, scanID uuid.UUID) (*domain.ScanDiffResult, error)
}

func TestScanAuthorizationHTTPIdentityAndConflict(t *testing.T) {
	tenant, user, target, profile := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, grouped := range []bool{false, true} {
		check := func(ctx context.Context) {
			a := domain.ScanAuthorizationFromContext(ctx)
			if a == nil || a.UserID != user || a.TenantID != tenant || a.AcceptedAt.Year() < 2026 {
				t.Fatal("request identity not derived from authenticated context")
			}
		}
		svc := &mockScanService{
			startScanFunc: func(ctx context.Context, _, _, _ uuid.UUID) (*domain.Scan, error) {
				check(ctx)
				return nil, domain.ErrScanAuthorizationRequired
			},
			startGroupScanFunc: func(ctx context.Context, _, _, _ uuid.UUID) ([]*domain.Scan, error) {
				check(ctx)
				return nil, domain.ErrScanAuthorizationRequired
			},
		}
		selector := "target_id"
		if grouped {
			selector = "group_id"
		}
		body := map[string]interface{}{selector: target, "profile_id": profile, "authorization": map[string]interface{}{"accepted": true, "statement": domain.ScanAuthorizationStatement, "user_id": uuid.New(), "tenant_id": uuid.New(), "accepted_at": "2000-01-01T00:00:00Z"}}
		data, _ := json.Marshal(body)
		r := httptest.NewRequest(http.MethodPost, "/scans", bytes.NewReader(data))
		r = r.WithContext(middleware.WithUserClaims(r.Context(), &auth.UserClaims{UserID: user, TenantID: tenant, Role: domain.RoleSecOps}))
		w := httptest.NewRecorder()
		handler.NewScanHandler(svc, nil).Start(w, r)
		if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "SCAN_AUTHORIZATION_REQUIRED") {
			t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
		}
	}
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

func (m *mockScanService) StartGroupScan(ctx context.Context, tenantID, groupID, profileID uuid.UUID) ([]*domain.Scan, error) {
	if m.startGroupScanFunc != nil {
		return m.startGroupScanFunc(ctx, tenantID, groupID, profileID)
	}
	return nil, nil
}

func (m *mockScanService) GetScanFindings(ctx context.Context, tenantID, scanID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
	if m.getScanFindingsFunc != nil {
		return m.getScanFindingsFunc(ctx, tenantID, scanID, filter)
	}
	return nil, 0, nil
}

func (m *mockScanService) ListProfiles(ctx context.Context, tenantID *uuid.UUID, category *domain.ScanProfileCategory) ([]*domain.ScanProfile, error) {
	if m.listProfilesFunc != nil {
		return m.listProfilesFunc(ctx, tenantID, category)
	}
	return nil, nil
}

func (m *mockScanService) CreateCustomProfile(ctx context.Context, tenantID uuid.UUID, name, nmapArgs string, description *string, timeoutSeconds int, category domain.ScanProfileCategory) (*domain.ScanProfile, error) {
	if m.createCustomProfileFunc != nil {
		return m.createCustomProfileFunc(ctx, tenantID, name, nmapArgs, description, timeoutSeconds, category)
	}
	return nil, nil
}

func (m *mockScanService) DeleteCustomProfile(ctx context.Context, tenantID, id uuid.UUID) error {
	if m.deleteCustomProfileFunc != nil {
		return m.deleteCustomProfileFunc(ctx, tenantID, id)
	}
	return nil
}

func (m *mockScanService) GetScanDiff(ctx context.Context, tenantID, scanID uuid.UUID) (*domain.ScanDiffResult, error) {
	if m.getScanDiffFunc != nil {
		return m.getScanDiffFunc(ctx, tenantID, scanID)
	}
	return nil, nil
}

func setupScanRouter(h *handler.ScanHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/scans", h.List)
	r.Get("/scans/{id}", h.GetByID)
	r.Get("/scans/{id}/stream", h.Stream)
	r.Get("/scans/{id}/findings", h.GetFindings)
	r.Get("/scans/{id}/diff", h.GetDiff)
	r.Post("/scans", h.Start)
	r.Post("/scans/{id}/cancel", h.Cancel)
	r.Get("/scan-profiles", h.ListProfiles)
	r.Post("/scan-profiles", h.CreateProfile)
	r.Delete("/scan-profiles/{id}", h.DeleteProfile)
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

	h := handler.NewScanHandler(mockSvc, nil)
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
	h := handler.NewScanHandler(mockSvc, nil)
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
	h := handler.NewScanHandler(mockSvc, nil)
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

	h := handler.NewScanHandler(mockSvc, nil)
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

	h := handler.NewScanHandler(mockSvc, nil)
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

	h := handler.NewScanHandler(mockSvc, nil)
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
	h := handler.NewScanHandler(mockSvc, nil)
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

	h := handler.NewScanHandler(mockSvc, nil)
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

	h := handler.NewScanHandler(mockSvc, nil)
	router := setupScanRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans?limit=10&offset=0", nil)
	req = withSecOpsContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d", rec.Code)
	}
}

func TestScanHandler_ListBatchProgress(t *testing.T) {
	tenantID, scanID := uuid.New(), uuid.New()
	for _, tc := range []struct {
		name, query string
		denied      bool
		status      int
	}{
		{"snapshot", scanID.String() + "," + scanID.String(), false, 200},
		{"invalid", "not-a-uuid", false, 400},
		{"empty", "", false, 400},
		{"too-many", strings.Repeat(scanID.String()+",", 100) + scanID.String(), false, 400},
		{"foreign-tenant", scanID.String(), true, 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			svc := &mockScanService{getScanFunc: func(ctx context.Context, tenant, id uuid.UUID) (*domain.Scan, error) {
				calls++
				if tenant != tenantID || id != scanID {
					t.Fatal("incorrect tenant or id")
				}
				if tc.denied {
					return nil, domain.ErrNotFound
				}
				return &domain.Scan{ID: id, TenantID: tenant, Status: domain.ScanStatusRunning, Progress: &domain.ScanProgressPayload{Percentage: 35}}, nil
			}}
			req := withSecOpsContext(httptest.NewRequest(http.MethodGet, "/scans?ids="+tc.query, nil), tenantID)
			rec := httptest.NewRecorder()
			setupScanRouter(handler.NewScanHandler(svc, nil)).ServeHTTP(rec, req)
			if rec.Code != tc.status {
				t.Fatalf("status=%d, want %d: %s", rec.Code, tc.status, rec.Body.String())
			}
			if tc.status == 200 && (calls != 1 || !strings.Contains(rec.Body.String(), `"percentage":35`)) {
				t.Fatal("missing snapshot or duplicate lookup", rec.Body.String())
			}
			if tc.status == 400 && calls != 0 {
				t.Fatal("looked up invalid ids")
			}
		})
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

	h := handler.NewScanHandler(mockSvc, nil)
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
	tenantID := uuid.New()
	mockSvc := &mockScanService{
		listProfilesFunc: func(ctx context.Context, tID *uuid.UUID, cat *domain.ScanProfileCategory) ([]*domain.ScanProfile, error) {
			catVal := domain.ProfileCategoryPenTest
			return []*domain.ScanProfile{
				{ID: uuid.New(), Name: "Stealth SYN", NmapArgs: "-sS -Pn -T2 --open", TimeoutSeconds: 1800, Category: catVal},
			}, nil
		},
	}

	h := handler.NewScanHandler(mockSvc, nil)
	router := setupScanRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scan-profiles?category=PEN_TEST", nil)
	req = withSecOpsContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d: %s", rec.Code, rec.Body.String())
	}
}

func TestScanHandler_CreateProfile_Success(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockScanService{
		createCustomProfileFunc: func(ctx context.Context, tID uuid.UUID, name, nmapArgs string, description *string, timeoutSeconds int, category domain.ScanProfileCategory) (*domain.ScanProfile, error) {
			return &domain.ScanProfile{
				ID:             uuid.New(),
				TenantID:       &tID,
				Name:           name,
				NmapArgs:       nmapArgs,
				Category:       category,
				IsCustom:       true,
				TimeoutSeconds: timeoutSeconds,
			}, nil
		},
	}

	h := handler.NewScanHandler(mockSvc, nil)
	router := setupScanRouter(h)

	body := `{"name":"Custom Pentest","nmap_args":"-sS -p 22,80,443 -T3","category":"PEN_TEST","timeout_seconds":600}`
	req := httptest.NewRequest(http.MethodPost, "/scan-profiles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withSecOpsContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("atteso status 201 Created, ottenuto %d: %s", rec.Code, rec.Body.String())
	}
}

func TestScanHandler_DeleteProfile_Success(t *testing.T) {
	tenantID := uuid.New()
	profileID := uuid.New()

	mockSvc := &mockScanService{
		deleteCustomProfileFunc: func(ctx context.Context, tID, pID uuid.UUID) error {
			if tID != tenantID || pID != profileID {
				t.Errorf("parametri errati: tID=%v, pID=%v", tID, pID)
			}
			return nil
		},
	}

	h := handler.NewScanHandler(mockSvc, nil)
	router := setupScanRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/scan-profiles/"+profileID.String(), nil)
	req = withSecOpsContext(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d: %s", rec.Code, rec.Body.String())
	}
}

func TestScanHandler_Stream(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	t.Run("NotFound_Error", func(t *testing.T) {
		mockSvc := &mockScanService{
			getScanFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.Scan, error) {
				return nil, domain.ErrNotFound
			},
		}

		h := handler.NewScanHandler(mockSvc, nil)
		router := setupScanRouter(h)

		req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/stream", nil)
		req = withSecOpsContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("atteso status 404 NOT_FOUND, ottenuto %d", rec.Code)
		}
	})

	t.Run("InvalidID_BadRequest", func(t *testing.T) {
		h := handler.NewScanHandler(&mockScanService{}, nil)
		router := setupScanRouter(h)

		req := httptest.NewRequest(http.MethodGet, "/scans/invalid-uuid/stream", nil)
		req = withSecOpsContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("atteso status 400 INVALID_ID, ottenuto %d", rec.Code)
		}
	})

	t.Run("CompletedScanReplay_Success", func(t *testing.T) {
		completedAt := time.Now()
		mockSvc := &mockScanService{
			getScanFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.Scan, error) {
				return &domain.Scan{
					ID:          scanID,
					TenantID:    tenantID,
					Status:      domain.ScanStatusCompleted,
					CompletedAt: &completedAt,
					SummaryStats: domain.ScanSummaryStats{
						TotalHosts: 1,
						UpHosts:    1,
					},
				}, nil
			},
			getScanFindingsFunc: func(ctx context.Context, tID, sID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
				return []*domain.Finding{
					{
						ID:       uuid.New(),
						ScanID:   scanID,
						TenantID: tenantID,
						HostIP:   "10.0.0.1",
						Port:     80,
						Protocol: "tcp",
						Severity: domain.SeverityMedium,
					},
				}, 1, nil
			},
		}

		h := handler.NewScanHandler(mockSvc, nil)
		router := setupScanRouter(h)

		req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/stream", nil)
		req = withSecOpsContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "event: status_change") {
			t.Errorf("atteso 'event: status_change' nel body:\n%s", body)
		}
		if !strings.Contains(body, "event: finding_discovered") {
			t.Errorf("atteso 'event: finding_discovered' nel body:\n%s", body)
		}
		if !strings.Contains(body, "10.0.0.1") {
			t.Errorf("atteso IP finding nel replay:\n%s", body)
		}
	})

	t.Run("LiveStreaming_Success", func(t *testing.T) {
		startedAt := time.Now()
		mockSvc := &mockScanService{
			getScanFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.Scan, error) {
				return &domain.Scan{
					ID:        scanID,
					TenantID:  tenantID,
					Status:    domain.ScanStatusRunning,
					StartedAt: &startedAt,
				}, nil
			},
			getScanFindingsFunc: func(ctx context.Context, tID, sID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
				return nil, 0, nil
			},
		}

		broadcaster := service.NewEventBroadcaster(10, nil)
		defer broadcaster.Close()

		h := handler.NewScanHandler(mockSvc, broadcaster)
		router := setupScanRouter(h)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/stream", nil)
		req = req.WithContext(ctx)
		req = withSecOpsContext(req, tenantID)
		rec := httptest.NewRecorder()

		doneCh := make(chan struct{})
		go func() {
			router.ServeHTTP(rec, req)
			close(doneCh)
		}()

		time.Sleep(30 * time.Millisecond)

		// Pubblica evento live finding_discovered
		topic := domain.ScanTopic(scanID)
		broadcaster.Publish(topic, domain.SSEEvent{
			Event: domain.EventScanFinding,
			Data: domain.ScanFindingPayload{
				ScanID: scanID,
				Finding: domain.Finding{
					ID:       uuid.New(),
					HostIP:   "192.168.1.100",
					Port:     443,
					Protocol: "tcp",
					Severity: domain.SeverityHigh,
				},
			},
		})

		// Pubblica evento finale status_change COMPLETED (causerà la chiusura dello stream)
		broadcaster.Publish(topic, domain.SSEEvent{
			Event: domain.EventScanStatusChange,
			Data: domain.ScanStatusPayload{
				ScanID: scanID,
				Status: domain.ScanStatusCompleted,
				SummaryStats: domain.ScanSummaryStats{
					UpHosts: 1,
				},
			},
		})

		select {
		case <-doneCh:
			// Chiuso correttamente su status COMPLETED
		case <-time.After(1 * time.Second):
			cancel()
			t.Fatal("timeout in attesa del completamento dello stream")
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "event: finding_discovered") {
			t.Errorf("atteso 'event: finding_discovered' in streaming:\n%s", body)
		}
		if !strings.Contains(body, "192.168.1.100") {
			t.Errorf("atteso IP scoperto nello stream:\n%s", body)
		}
		if !strings.Contains(body, "COMPLETED") {
			t.Errorf("atteso evento status COMPLETED nello stream:\n%s", body)
		}
	})
}

func TestScanHandler_GetDiff(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()
	prevScanID := uuid.New()

	t.Run("successo - restituisce 200 OK con payload ScanDiffResult", func(t *testing.T) {
		svc := &mockScanService{
			getScanDiffFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.ScanDiffResult, error) {
				return &domain.ScanDiffResult{
					CurrentScanID:   scanID,
					PreviousScanID:  &prevScanID,
					TargetValue:     "192.168.1.10",
					TotalNew:        1,
					TotalPersistent: 1,
					TotalResolved:   0,
					Items: []domain.FindingDiffItem{
						{
							FindingKey:    "192.168.1.10:80:tcp:CVE-2023-1234",
							Status:        domain.DiffStatusNew,
							HostIP:        "192.168.1.10",
							Port:          80,
							Protocol:      "tcp",
							Severity:      domain.SeverityHigh,
							CurrentScanID: scanID,
						},
					},
				}, nil
			},
		}

		h := handler.NewScanHandler(svc, nil)
		r := setupScanRouter(h)

		req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/diff", nil)
		req = withSecOpsContext(req, tenantID)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rec.Code)
		}

		var env response.Envelope
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Fatalf("deserializzazione response fallita: %v", err)
		}

		dataBytes, _ := json.Marshal(env.Data)
		var res domain.ScanDiffResult
		if err := json.Unmarshal(dataBytes, &res); err != nil {
			t.Fatalf("deserializzazione data fallita: %v", err)
		}

		if res.CurrentScanID != scanID {
			t.Errorf("CurrentScanID atteso %v, ottenuto %v", scanID, res.CurrentScanID)
		}
		if res.TotalNew != 1 {
			t.Errorf("TotalNew atteso 1, ottenuto %d", res.TotalNew)
		}
	})

	t.Run("errore - scansione non trovata restituisce 404", func(t *testing.T) {
		svc := &mockScanService{
			getScanDiffFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.ScanDiffResult, error) {
				return nil, domain.ErrNotFound
			},
		}

		h := handler.NewScanHandler(svc, nil)
		r := setupScanRouter(h)

		req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/diff", nil)
		req = withSecOpsContext(req, tenantID)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("atteso status 404, ottenuto %d", rec.Code)
		}
	})

	t.Run("errore - non autorizzato senza tenant restituisce 401", func(t *testing.T) {
		svc := &mockScanService{}
		h := handler.NewScanHandler(svc, nil)
		r := setupScanRouter(h)

		req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/diff", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("atteso status 401, ottenuto %d", rec.Code)
		}
	})
}
