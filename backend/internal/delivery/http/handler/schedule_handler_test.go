package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
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

// --- Mock Repositories ---

type mockScheduleRepo struct {
	mu        sync.RWMutex
	schedules map[uuid.UUID]*domain.Schedule
}

func newMockScheduleRepo() *mockScheduleRepo {
	return &mockScheduleRepo{
		schedules: make(map[uuid.UUID]*domain.Schedule),
	}
}

func (m *mockScheduleRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Schedule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.schedules[id]
	if !ok || s.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *mockScheduleRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Schedule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.Schedule
	for _, s := range m.schedules {
		if s.TenantID == tenantID {
			cp := *s
			list = append(list, &cp)
		}
	}
	return list, nil
}

func (m *mockScheduleRepo) ListActive(ctx context.Context) ([]*domain.Schedule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.Schedule
	for _, s := range m.schedules {
		if s.IsActive {
			cp := *s
			list = append(list, &cp)
		}
	}
	return list, nil
}

func (m *mockScheduleRepo) Create(ctx context.Context, schedule *domain.Schedule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if schedule.ID == uuid.Nil {
		schedule.ID = uuid.New()
	}
	cp := *schedule
	m.schedules[schedule.ID] = &cp
	return nil
}

func (m *mockScheduleRepo) Update(ctx context.Context, schedule *domain.Schedule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.schedules[schedule.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *schedule
	m.schedules[schedule.ID] = &cp
	return nil
}

func (m *mockScheduleRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.schedules[id]
	if !ok || s.TenantID != tenantID {
		return domain.ErrNotFound
	}
	delete(m.schedules, id)
	return nil
}

type mockScanRepoForSched struct{}

func (m *mockScanRepoForSched) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Scan, error) {
	return nil, domain.ErrNotFound
}
func (m *mockScanRepoForSched) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Scan, int, error) {
	return nil, 0, nil
}
func (m *mockScanRepoForSched) GetActiveByTarget(ctx context.Context, tenantID, targetID uuid.UUID) (*domain.Scan, error) {
	return nil, nil
}
func (m *mockScanRepoForSched) Create(ctx context.Context, scan *domain.Scan) error {
	return nil
}
func (m *mockScanRepoForSched) Update(ctx context.Context, scan *domain.Scan) error {
	return nil
}

type mockOrchestratorForSched struct{}

func (m *mockOrchestratorForSched) Enqueue(tenantID, scanID uuid.UUID) error {
	return nil
}

type mockTargetRepoForSched struct {
	mu      sync.RWMutex
	targets map[uuid.UUID]*domain.Target
}

func newMockTargetRepoForSched() *mockTargetRepoForSched {
	return &mockTargetRepoForSched{
		targets: make(map[uuid.UUID]*domain.Target),
	}
}

func (m *mockTargetRepoForSched) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Target, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.targets[id]
	if !ok || t.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (m *mockTargetRepoForSched) List(ctx context.Context, tenantID uuid.UUID, filter domain.TargetFilter) ([]*domain.Target, int, error) {
	return nil, 0, nil
}
func (m *mockTargetRepoForSched) Create(ctx context.Context, target *domain.Target) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *target
	m.targets[target.ID] = &cp
	return nil
}
func (m *mockTargetRepoForSched) Update(ctx context.Context, target *domain.Target) error {
	return nil
}
func (m *mockTargetRepoForSched) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return nil
}
func (m *mockTargetRepoForSched) IsBlacklisted(ctx context.Context, tenantID uuid.UUID, value string) (bool, error) {
	return false, nil
}

type mockProfileRepoForSched struct {
	mu       sync.RWMutex
	profiles map[uuid.UUID]*domain.ScanProfile
}

func newMockProfileRepoForSched() *mockProfileRepoForSched {
	return &mockProfileRepoForSched{
		profiles: make(map[uuid.UUID]*domain.ScanProfile),
	}
}

func (m *mockProfileRepoForSched) GetByID(ctx context.Context, id uuid.UUID) (*domain.ScanProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.profiles[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *p
	return &cp, nil
}

func (m *mockProfileRepoForSched) GetByName(ctx context.Context, name string) (*domain.ScanProfile, error) {
	return nil, domain.ErrNotFound
}

func (m *mockProfileRepoForSched) List(ctx context.Context, tenantID *uuid.UUID) ([]*domain.ScanProfile, error) {
	return nil, nil
}

func (m *mockProfileRepoForSched) ListByCategory(ctx context.Context, tenantID *uuid.UUID, category domain.ScanProfileCategory) ([]*domain.ScanProfile, error) {
	return nil, nil
}

func (m *mockProfileRepoForSched) Create(ctx context.Context, profile *domain.ScanProfile) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *profile
	m.profiles[profile.ID] = &cp
	return nil
}

func (m *mockProfileRepoForSched) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.profiles, id)
	return nil
}

// --- Helpers di Test ---

func setupScheduleTestRouter(
	schedRepo domain.ScheduleRepository,
	targetRepo domain.TargetRepository,
	profileRepo domain.ScanProfileRepository,
) (*chi.Mux, *service.SchedulerService) {
	schedSvc := service.NewSchedulerService(schedRepo, &mockScanRepoForSched{}, &mockOrchestratorForSched{}, nil)
	h := handler.NewScheduleHandler(schedSvc, targetRepo, profileRepo)

	r := chi.NewRouter()
	r.Get("/schedules", h.List)
	r.Get("/schedules/{id}", h.GetByID)
	r.Post("/schedules", h.Create)
	r.Put("/schedules/{id}/toggle", h.Toggle)
	r.Delete("/schedules/{id}", h.Delete)

	return r, schedSvc
}

func withScheduleUserContext(r *http.Request, tenantID uuid.UUID, role domain.UserRole) *http.Request {
	claims := &auth.UserClaims{
		UserID:    uuid.New(),
		TenantID:  tenantID,
		Role:      role,
		Email:     "secops@test.local",
		TokenType: auth.TokenTypeAccess,
	}
	return r.WithContext(middleware.WithUserClaims(r.Context(), claims))
}

type apiResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// --- Test Suites ---

func TestScheduleHandler_List(t *testing.T) {
	tenantID := uuid.New()
	schedRepo := newMockScheduleRepo()
	targetRepo := newMockTargetRepoForSched()
	profileRepo := newMockProfileRepoForSched()

	// Popola due pianificazioni
	_ = schedRepo.Create(context.Background(), &domain.Schedule{
		ID:             uuid.New(),
		TenantID:       tenantID,
		TargetID:       uuid.New(),
		ProfileID:      uuid.New(),
		CronExpression: "0 2 * * *",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})
	_ = schedRepo.Create(context.Background(), &domain.Schedule{
		ID:             uuid.New(),
		TenantID:       tenantID,
		TargetID:       uuid.New(),
		ProfileID:      uuid.New(),
		CronExpression: "*/30 * * * *",
		IsActive:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	router, _ := setupScheduleTestRouter(schedRepo, targetRepo, profileRepo)

	t.Run("Unauthorized - no tenant in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/schedules", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("atteso status 401, ottenuto %d", rec.Code)
		}
	})

	t.Run("Success - returns schedule list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/schedules", nil)
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d: %s", rec.Code, rec.Body.String())
		}

		var resp apiResponse[struct {
			Items []handler.ScheduleResponse `json:"items"`
			Total int                        `json:"total"`
		}]
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("errore decodifica risposta: %v", err)
		}

		if !resp.Success {
			t.Fatalf("atteso success true, ottenuto false")
		}
		if resp.Data.Total != 2 || len(resp.Data.Items) != 2 {
			t.Errorf("attesi 2 elementi, ottenuti total=%d len=%d", resp.Data.Total, len(resp.Data.Items))
		}
	})
}

func TestScheduleHandler_GetByID(t *testing.T) {
	tenantID := uuid.New()
	scheduleID := uuid.New()
	schedRepo := newMockScheduleRepo()
	targetRepo := newMockTargetRepoForSched()
	profileRepo := newMockProfileRepoForSched()

	_ = schedRepo.Create(context.Background(), &domain.Schedule{
		ID:             scheduleID,
		TenantID:       tenantID,
		TargetID:       uuid.New(),
		ProfileID:      uuid.New(),
		CronExpression: "0 3 * * *",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	router, _ := setupScheduleTestRouter(schedRepo, targetRepo, profileRepo)

	t.Run("Unauthorized - no tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/schedules/"+scheduleID.String(), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("atteso status 401, ottenuto %d", rec.Code)
		}
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/schedules/not-a-valid-uuid", nil)
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("atteso status 400, ottenuto %d", rec.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		randomID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/schedules/"+randomID.String(), nil)
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("atteso status 404, ottenuto %d", rec.Code)
		}
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/schedules/"+scheduleID.String(), nil)
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rec.Code)
		}

		var resp apiResponse[handler.ScheduleResponse]
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("errore decodifica: %v", err)
		}

		if !resp.Success {
			t.Fatalf("atteso success true")
		}
		if resp.Data.ID != scheduleID.String() {
			t.Errorf("atteso ID %s, ottenuto %s", scheduleID.String(), resp.Data.ID)
		}
		if resp.Data.CronExpression != "0 3 * * *" {
			t.Errorf("attesa cron_expression '0 3 * * *', ottenuta %s", resp.Data.CronExpression)
		}
	})
}

func TestScheduleHandler_Create(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()
	blacklistedTargetID := uuid.New()
	profileID := uuid.New()

	schedRepo := newMockScheduleRepo()
	targetRepo := newMockTargetRepoForSched()
	profileRepo := newMockProfileRepoForSched()

	// Configurazione target valido
	_ = targetRepo.Create(context.Background(), &domain.Target{
		ID:            targetID,
		TenantID:      tenantID,
		Type:          domain.TargetTypeIP,
		Value:         "192.168.1.50",
		IsBlacklisted: false,
	})

	// Configurazione target in blacklist
	_ = targetRepo.Create(context.Background(), &domain.Target{
		ID:            blacklistedTargetID,
		TenantID:      tenantID,
		Type:          domain.TargetTypeIP,
		Value:         "127.0.0.1",
		IsBlacklisted: true,
	})

	// Configurazione profilo
	profileRepo.profiles[profileID] = &domain.ScanProfile{
		ID:   profileID,
		Name: "Full Vulnerability Scan",
	}

	router, _ := setupScheduleTestRouter(schedRepo, targetRepo, profileRepo)

	t.Run("Unauthorized - no tenant", func(t *testing.T) {
		body, _ := json.Marshal(handler.CreateScheduleRequest{
			TargetID:       targetID,
			ProfileID:      profileID,
			CronExpression: "0 2 * * *",
		})
		req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("atteso status 401, ottenuto %d", rec.Code)
		}
	})

	t.Run("Bad Request - invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewReader([]byte("{invalid-json")))
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("atteso status 400, ottenuto %d", rec.Code)
		}
	})

	t.Run("Validation Failed - missing fields", func(t *testing.T) {
		body, _ := json.Marshal(handler.CreateScheduleRequest{
			TargetID: uuid.Nil,
		})
		req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewReader(body))
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("atteso status 400, ottenuto %d", rec.Code)
		}
	})

	t.Run("Target Not Found", func(t *testing.T) {
		body, _ := json.Marshal(handler.CreateScheduleRequest{
			TargetID:       uuid.New(),
			ProfileID:      profileID,
			CronExpression: "0 2 * * *",
		})
		req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewReader(body))
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("atteso status 404, ottenuto %d", rec.Code)
		}
	})

	t.Run("Target Blacklisted - Forbidden", func(t *testing.T) {
		body, _ := json.Marshal(handler.CreateScheduleRequest{
			TargetID:       blacklistedTargetID,
			ProfileID:      profileID,
			CronExpression: "0 2 * * *",
		})
		req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewReader(body))
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("atteso status 403, ottenuto %d", rec.Code)
		}
	})

	t.Run("Profile Not Found", func(t *testing.T) {
		body, _ := json.Marshal(handler.CreateScheduleRequest{
			TargetID:       targetID,
			ProfileID:      uuid.New(),
			CronExpression: "0 2 * * *",
		})
		req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewReader(body))
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("atteso status 404, ottenuto %d", rec.Code)
		}
	})

	t.Run("Invalid Cron Expression", func(t *testing.T) {
		body, _ := json.Marshal(handler.CreateScheduleRequest{
			TargetID:       targetID,
			ProfileID:      profileID,
			CronExpression: "not-a-cron-expression",
		})
		req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewReader(body))
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("atteso status 400, ottenuto %d", rec.Code)
		}
	})

	t.Run("Success - 201 Created", func(t *testing.T) {
		body, _ := json.Marshal(handler.CreateScheduleRequest{
			TargetID:       targetID,
			ProfileID:      profileID,
			CronExpression: "0 5 * * *",
		})
		req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewReader(body))
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("atteso status 201, ottenuto %d: %s", rec.Code, rec.Body.String())
		}

		var resp apiResponse[handler.ScheduleResponse]
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("errore decodifica: %v", err)
		}

		if !resp.Success {
			t.Fatalf("atteso success true")
		}
		if resp.Data.TargetID != targetID.String() {
			t.Errorf("atteso TargetID %s, ottenuto %s", targetID.String(), resp.Data.TargetID)
		}
		if !resp.Data.IsActive {
			t.Errorf("atteso IsActive true di default")
		}
		if resp.Data.NextRunAt == nil {
			t.Errorf("atteso NextRunAt calcolato")
		}
	})
}

func TestScheduleHandler_Toggle(t *testing.T) {
	tenantID := uuid.New()
	scheduleID := uuid.New()

	schedRepo := newMockScheduleRepo()
	targetRepo := newMockTargetRepoForSched()
	profileRepo := newMockProfileRepoForSched()

	_ = schedRepo.Create(context.Background(), &domain.Schedule{
		ID:             scheduleID,
		TenantID:       tenantID,
		TargetID:       uuid.New(),
		ProfileID:      uuid.New(),
		CronExpression: "0 1 * * *",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	router, _ := setupScheduleTestRouter(schedRepo, targetRepo, profileRepo)

	t.Run("Unauthorized - no tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/schedules/"+scheduleID.String()+"/toggle", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("atteso status 401, ottenuto %d", rec.Code)
		}
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/schedules/bad-uuid/toggle", nil)
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("atteso status 400, ottenuto %d", rec.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		randomID := uuid.New()
		req := httptest.NewRequest(http.MethodPut, "/schedules/"+randomID.String()+"/toggle", nil)
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("atteso status 404, ottenuto %d", rec.Code)
		}
	})

	t.Run("Success - Toggle active to inactive and back", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/schedules/"+scheduleID.String()+"/toggle", nil)
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d: %s", rec.Code, rec.Body.String())
		}

		var resp apiResponse[handler.ScheduleResponse]
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("errore decodifica: %v", err)
		}
		if !resp.Success {
			t.Fatalf("atteso success true")
		}
		if resp.Data.IsActive != false {
			t.Errorf("atteso IsActive false dopo il primo toggle, ottenuto %v", resp.Data.IsActive)
		}

		// Secondo toggle: da false a true
		req2 := httptest.NewRequest(http.MethodPut, "/schedules/"+scheduleID.String()+"/toggle", nil)
		req2 = withScheduleUserContext(req2, tenantID, domain.RoleSecOps)
		rec2 := httptest.NewRecorder()

		router.ServeHTTP(rec2, req2)

		if rec2.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d: %s", rec2.Code, rec2.Body.String())
		}

		var resp2 apiResponse[handler.ScheduleResponse]
		if err := json.NewDecoder(rec2.Body).Decode(&resp2); err != nil {
			t.Fatalf("errore decodifica: %v", err)
		}
		if !resp2.Success {
			t.Fatalf("atteso success true")
		}
		if resp2.Data.IsActive != true {
			t.Errorf("atteso IsActive true dopo il secondo toggle, ottenuto %v", resp2.Data.IsActive)
		}
	})
}

func TestScheduleHandler_Delete(t *testing.T) {
	tenantID := uuid.New()
	scheduleID := uuid.New()

	schedRepo := newMockScheduleRepo()
	targetRepo := newMockTargetRepoForSched()
	profileRepo := newMockProfileRepoForSched()

	_ = schedRepo.Create(context.Background(), &domain.Schedule{
		ID:             scheduleID,
		TenantID:       tenantID,
		TargetID:       uuid.New(),
		ProfileID:      uuid.New(),
		CronExpression: "0 0 * * *",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	router, _ := setupScheduleTestRouter(schedRepo, targetRepo, profileRepo)

	t.Run("Unauthorized - no tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/schedules/"+scheduleID.String(), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("atteso status 401, ottenuto %d", rec.Code)
		}
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/schedules/bad-uuid", nil)
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("atteso status 400, ottenuto %d", rec.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		randomID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/schedules/"+randomID.String(), nil)
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("atteso status 404, ottenuto %d", rec.Code)
		}
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/schedules/"+scheduleID.String(), nil)
		req = withScheduleUserContext(req, tenantID, domain.RoleSecOps)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d: %s", rec.Code, rec.Body.String())
		}

		// Verifica che sia stato eliminato dal repo
		_, err := schedRepo.GetByID(context.Background(), tenantID, scheduleID)
		if err == nil {
			t.Errorf("la schedulazione non avrebbe dovuto essere presente nel repo dopo la cancellazione")
		}
	})
}

