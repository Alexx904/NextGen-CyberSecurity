package handler_test

import (
	"bytes"
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
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
)

// mockTargetService implementa service.TargetService per i test del controller
type mockTargetService struct {
	createTargetFunc func(ctx context.Context, input service.CreateTargetInput) (*domain.Target, error)
	getTargetFunc    func(ctx context.Context, tenantID, id uuid.UUID) (*domain.Target, error)
	listTargetsFunc  func(ctx context.Context, tenantID uuid.UUID, filter domain.TargetFilter) ([]*domain.Target, int, error)
	updateTargetFunc func(ctx context.Context, input service.UpdateTargetInput) (*domain.Target, error)
	deleteTargetFunc func(ctx context.Context, tenantID, id uuid.UUID) error

	createGroupFunc func(ctx context.Context, tenantID uuid.UUID, name string, description *string) (*domain.TargetGroup, error)
	listGroupsFunc  func(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroup, error)
	deleteGroupFunc func(ctx context.Context, tenantID, id uuid.UUID) error
}

func (m *mockTargetService) CreateTarget(ctx context.Context, input service.CreateTargetInput) (*domain.Target, error) {
	if m.createTargetFunc != nil {
		return m.createTargetFunc(ctx, input)
	}
	return nil, nil
}

func (m *mockTargetService) GetTarget(ctx context.Context, tenantID, id uuid.UUID) (*domain.Target, error) {
	if m.getTargetFunc != nil {
		return m.getTargetFunc(ctx, tenantID, id)
	}
	return nil, nil
}

func (m *mockTargetService) ListTargets(ctx context.Context, tenantID uuid.UUID, filter domain.TargetFilter) ([]*domain.Target, int, error) {
	if m.listTargetsFunc != nil {
		return m.listTargetsFunc(ctx, tenantID, filter)
	}
	return nil, 0, nil
}

func (m *mockTargetService) UpdateTarget(ctx context.Context, input service.UpdateTargetInput) (*domain.Target, error) {
	if m.updateTargetFunc != nil {
		return m.updateTargetFunc(ctx, input)
	}
	return nil, nil
}

func (m *mockTargetService) DeleteTarget(ctx context.Context, tenantID, id uuid.UUID) error {
	if m.deleteTargetFunc != nil {
		return m.deleteTargetFunc(ctx, tenantID, id)
	}
	return nil
}

func (m *mockTargetService) CreateGroup(ctx context.Context, tenantID uuid.UUID, name string, description *string) (*domain.TargetGroup, error) {
	if m.createGroupFunc != nil {
		return m.createGroupFunc(ctx, tenantID, name, description)
	}
	return nil, nil
}

func (m *mockTargetService) ListGroups(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroup, error) {
	if m.listGroupsFunc != nil {
		return m.listGroupsFunc(ctx, tenantID)
	}
	return nil, nil
}

func (m *mockTargetService) DeleteGroup(ctx context.Context, tenantID, id uuid.UUID) error {
	if m.deleteGroupFunc != nil {
		return m.deleteGroupFunc(ctx, tenantID, id)
	}
	return nil
}

func setupTargetRouter(h *handler.TargetHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/targets", h.List)
	r.Get("/targets/{id}", h.GetByID)
	r.Post("/targets", h.Create)
	r.Delete("/targets/{id}", h.Delete)
	return r
}

func withTenantContext(r *http.Request, tenantID uuid.UUID) *http.Request {
	claims := &auth.UserClaims{
		UserID:    uuid.New(),
		TenantID:  tenantID,
		Role:      domain.RoleSecOps,
		Email:     "operator@test.local",
		TokenType: auth.TokenTypeAccess,
	}
	return r.WithContext(middleware.WithUserClaims(r.Context(), claims))
}

func TestTargetHandler_List(t *testing.T) {
	tenantID := uuid.New()
	targets := []*domain.Target{
		{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Type:      domain.TargetTypeIP,
			Value:     "192.168.1.100",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Type:      domain.TargetTypeCIDR,
			Value:     "10.0.0.0/24",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}

	t.Run("success with results", func(t *testing.T) {
		mockSvc := &mockTargetService{
			listTargetsFunc: func(ctx context.Context, tid uuid.UUID, filter domain.TargetFilter) ([]*domain.Target, int, error) {
				if tid != tenantID {
					return nil, 0, errors.New("mismatched tenant")
				}
				return targets, 2, nil
			},
		}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		req := httptest.NewRequest(http.MethodGet, "/targets?limit=10&offset=0&is_blacklisted=false", nil)
		req = withTenantContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Atteso 200, ottenuto %d", rec.Code)
		}

		var env response.Envelope
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Fatalf("Errore parsing envelope: %v", err)
		}
		if !env.Success {
			t.Fatalf("Atteso success true")
		}
	})

	t.Run("missing tenant in context returns 401", func(t *testing.T) {
		mockSvc := &mockTargetService{}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		req := httptest.NewRequest(http.MethodGet, "/targets", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("Atteso 401, ottenuto %d", rec.Code)
		}
	})

	t.Run("service internal error returns 500", func(t *testing.T) {
		mockSvc := &mockTargetService{
			listTargetsFunc: func(ctx context.Context, tid uuid.UUID, filter domain.TargetFilter) ([]*domain.Target, int, error) {
				return nil, 0, errors.New("db error")
			},
		}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		req := httptest.NewRequest(http.MethodGet, "/targets", nil)
		req = withTenantContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("Atteso 500, ottenuto %d", rec.Code)
		}
	})
}

func TestTargetHandler_GetByID(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()
	sampleTarget := &domain.Target{
		ID:        targetID,
		TenantID:  tenantID,
		Type:      domain.TargetTypeFQDN,
		Value:     "api.target.local",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	t.Run("success found", func(t *testing.T) {
		mockSvc := &mockTargetService{
			getTargetFunc: func(ctx context.Context, tid, id uuid.UUID) (*domain.Target, error) {
				if tid == tenantID && id == targetID {
					return sampleTarget, nil
				}
				return nil, domain.ErrNotFound
			},
		}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		req := httptest.NewRequest(http.MethodGet, "/targets/"+targetID.String(), nil)
		req = withTenantContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Atteso 200, ottenuto %d", rec.Code)
		}
	})

	t.Run("invalid UUID returns 400", func(t *testing.T) {
		mockSvc := &mockTargetService{}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		req := httptest.NewRequest(http.MethodGet, "/targets/not-a-uuid", nil)
		req = withTenantContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Atteso 400, ottenuto %d", rec.Code)
		}
	})

	t.Run("target not found returns 404", func(t *testing.T) {
		mockSvc := &mockTargetService{
			getTargetFunc: func(ctx context.Context, tid, id uuid.UUID) (*domain.Target, error) {
				return nil, domain.ErrNotFound
			},
		}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		req := httptest.NewRequest(http.MethodGet, "/targets/"+uuid.New().String(), nil)
		req = withTenantContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("Atteso 404, ottenuto %d", rec.Code)
		}
	})
}

func TestTargetHandler_Create(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()

	t.Run("success creation", func(t *testing.T) {
		mockSvc := &mockTargetService{
			createTargetFunc: func(ctx context.Context, input service.CreateTargetInput) (*domain.Target, error) {
				return &domain.Target{
					ID:            targetID,
					TenantID:      input.TenantID,
					Type:          domain.TargetTypeIP,
					Value:         input.Value,
					IsBlacklisted: input.IsBlacklisted,
					CreatedAt:     time.Now().UTC(),
					UpdatedAt:     time.Now().UTC(),
				}, nil
			},
		}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		body, _ := json.Marshal(map[string]any{
			"value":          "192.168.1.50",
			"is_blacklisted": false,
		})
		req := httptest.NewRequest(http.MethodPost, "/targets", bytes.NewReader(body))
		req = withTenantContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Atteso 201 Created, ottenuto %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("malformed json returns 400", func(t *testing.T) {
		mockSvc := &mockTargetService{}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		req := httptest.NewRequest(http.MethodPost, "/targets", bytes.NewReader([]byte("{invalid-json")))
		req = withTenantContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Atteso 400, ottenuto %d", rec.Code)
		}
	})

	t.Run("invalid target syntax returns 400 INVALID_TARGET_FORMAT", func(t *testing.T) {
		mockSvc := &mockTargetService{
			createTargetFunc: func(ctx context.Context, input service.CreateTargetInput) (*domain.Target, error) {
				return nil, domain.ErrTargetInvalid
			},
		}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		body, _ := json.Marshal(map[string]any{"value": "--script=vuln"})
		req := httptest.NewRequest(http.MethodPost, "/targets", bytes.NewReader(body))
		req = withTenantContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Atteso 400, ottenuto %d", rec.Code)
		}

		var env response.Envelope
		_ = json.Unmarshal(rec.Body.Bytes(), &env)
		if env.Error == nil || env.Error.Code != "INVALID_TARGET_FORMAT" {
			t.Fatalf("Atteso error code INVALID_TARGET_FORMAT, ottenuto %v", env.Error)
		}
	})

	t.Run("blacklist collision returns 409 TARGET_BLACKLISTED", func(t *testing.T) {
		mockSvc := &mockTargetService{
			createTargetFunc: func(ctx context.Context, input service.CreateTargetInput) (*domain.Target, error) {
				return nil, domain.ErrTargetBlacklisted
			},
		}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		body, _ := json.Marshal(map[string]any{"value": "127.0.0.1"})
		req := httptest.NewRequest(http.MethodPost, "/targets", bytes.NewReader(body))
		req = withTenantContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("Atteso 409 Conflict, ottenuto %d", rec.Code)
		}

		var env response.Envelope
		_ = json.Unmarshal(rec.Body.Bytes(), &env)
		if env.Error == nil || env.Error.Code != "TARGET_BLACKLISTED" {
			t.Fatalf("Atteso error code TARGET_BLACKLISTED, ottenuto %v", env.Error)
		}
	})

	t.Run("target already exists returns 409 TARGET_CONFLICT", func(t *testing.T) {
		mockSvc := &mockTargetService{
			createTargetFunc: func(ctx context.Context, input service.CreateTargetInput) (*domain.Target, error) {
				return nil, domain.ErrConflict
			},
		}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		body, _ := json.Marshal(map[string]any{"value": "192.168.1.1"})
		req := httptest.NewRequest(http.MethodPost, "/targets", bytes.NewReader(body))
		req = withTenantContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("Atteso 409, ottenuto %d", rec.Code)
		}
	})
}

func TestTargetHandler_Delete(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()

	t.Run("success delete", func(t *testing.T) {
		mockSvc := &mockTargetService{
			deleteTargetFunc: func(ctx context.Context, tid, id uuid.UUID) error {
				if tid == tenantID && id == targetID {
					return nil
				}
				return domain.ErrNotFound
			},
		}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		req := httptest.NewRequest(http.MethodDelete, "/targets/"+targetID.String(), nil)
		req = withTenantContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Atteso 200 OK, ottenuto %d", rec.Code)
		}
	})

	t.Run("target not found returns 404", func(t *testing.T) {
		mockSvc := &mockTargetService{
			deleteTargetFunc: func(ctx context.Context, tid, id uuid.UUID) error {
				return domain.ErrNotFound
			},
		}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		req := httptest.NewRequest(http.MethodDelete, "/targets/"+targetID.String(), nil)
		req = withTenantContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("Atteso 404, ottenuto %d", rec.Code)
		}
	})

	t.Run("invalid UUID returns 400", func(t *testing.T) {
		mockSvc := &mockTargetService{}
		router := setupTargetRouter(handler.NewTargetHandler(mockSvc))

		req := httptest.NewRequest(http.MethodDelete, "/targets/invalid-id", nil)
		req = withTenantContext(req, tenantID)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Atteso 400, ottenuto %d", rec.Code)
		}
	})
}
