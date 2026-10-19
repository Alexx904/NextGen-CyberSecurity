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
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
)

type mockTargetGroupService struct {
	createGroupFunc func(ctx context.Context, input service.CreateGroupInput) (*domain.TargetGroup, error)
	getGroupFunc    func(ctx context.Context, tenantID, id uuid.UUID) (*domain.TargetGroup, error)
	listGroupsFunc  func(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroup, error)
	getTreeFunc     func(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroupNode, error)
	updateGroupFunc func(ctx context.Context, input service.UpdateGroupInput) (*domain.TargetGroup, error)
	deleteGroupFunc func(ctx context.Context, tenantID, id uuid.UUID) error
}

func (m *mockTargetGroupService) CreateGroup(ctx context.Context, input service.CreateGroupInput) (*domain.TargetGroup, error) {
	if m.createGroupFunc != nil {
		return m.createGroupFunc(ctx, input)
	}
	return nil, nil
}

func (m *mockTargetGroupService) GetGroup(ctx context.Context, tenantID, id uuid.UUID) (*domain.TargetGroup, error) {
	if m.getGroupFunc != nil {
		return m.getGroupFunc(ctx, tenantID, id)
	}
	return nil, nil
}

func (m *mockTargetGroupService) ListGroups(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroup, error) {
	if m.listGroupsFunc != nil {
		return m.listGroupsFunc(ctx, tenantID)
	}
	return nil, nil
}

func (m *mockTargetGroupService) GetTree(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroupNode, error) {
	if m.getTreeFunc != nil {
		return m.getTreeFunc(ctx, tenantID)
	}
	return nil, nil
}

func (m *mockTargetGroupService) UpdateGroup(ctx context.Context, input service.UpdateGroupInput) (*domain.TargetGroup, error) {
	if m.updateGroupFunc != nil {
		return m.updateGroupFunc(ctx, input)
	}
	return nil, nil
}

func (m *mockTargetGroupService) DeleteGroup(ctx context.Context, tenantID, id uuid.UUID) error {
	if m.deleteGroupFunc != nil {
		return m.deleteGroupFunc(ctx, tenantID, id)
	}
	return nil
}

func setupGroupRouter(h *handler.TargetGroupHandler, tenantID uuid.UUID, role domain.UserRole) *chi.Mux {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			claims := &auth.UserClaims{
				UserID:    uuid.New(),
				TenantID:  tenantID,
				Role:      role,
				Email:     "user@test.local",
				TokenType: auth.TokenTypeAccess,
			}
			ctx := middleware.WithUserClaims(req.Context(), claims)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	r.Get("/target-groups", h.GetTree)
	r.Get("/target-groups/{id}", h.GetByID)
	r.Group(func(secOps chi.Router) {
		secOps.Use(middleware.RequireSecOps())
		secOps.Post("/target-groups", h.Create)
		secOps.Put("/target-groups/{id}", h.Update)
		secOps.Delete("/target-groups/{id}", h.Delete)
		secOps.Post("/target-groups/{id}/scan", h.StartScan)
	})
	return r
}

func TestTargetGroupDeleteAlertConflict(t *testing.T) {
	svc := &mockTargetGroupService{deleteGroupFunc: func(context.Context, uuid.UUID, uuid.UUID) error { return domain.ErrConflict }}
	r := setupGroupRouter(handler.NewTargetGroupHandler(svc, nil), uuid.New(), domain.RoleTenantAdmin)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/target-groups/"+uuid.NewString(), nil))
	if w.Code != http.StatusConflict {
		t.Fatalf("expected actionable scope conflict, got %d: %s", w.Code, w.Body)
	}
}

func TestTargetGroupHandler_GetTree_Success(t *testing.T) {
	tenantID := uuid.New()
	treeData := []*domain.TargetGroupNode{
		{
			ID:                uuid.New(),
			TenantID:          tenantID,
			Name:              "DMZ Group",
			TargetCount:       5,
			DirectTargetCount: 2,
			AggregateRisk:     domain.SeverityHigh,
			Children:          []*domain.TargetGroupNode{},
		},
	}

	groupSvc := &mockTargetGroupService{
		getTreeFunc: func(ctx context.Context, tid uuid.UUID) ([]*domain.TargetGroupNode, error) {
			if tid != tenantID {
				return nil, domain.ErrNotFound
			}
			return treeData, nil
		},
	}
	scanSvc := &mockScanService{}
	h := handler.NewTargetGroupHandler(groupSvc, scanSvc)
	router := setupGroupRouter(h, tenantID, domain.RoleAuditor)

	req := httptest.NewRequest(http.MethodGet, "/target-groups", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Atteso 200 OK, ottenuto %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Data []*domain.TargetGroupNode `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal risposta fallito: %v", err)
	}

	if len(resp.Data) != 1 || resp.Data[0].Name != "DMZ Group" {
		t.Errorf("Dati albero non conformi: %+v", resp.Data)
	}
}

func TestTargetGroupHandler_GetByID_SuccessAndNotFound(t *testing.T) {
	tenantID := uuid.New()
	groupID := uuid.New()

	groupSvc := &mockTargetGroupService{
		getGroupFunc: func(ctx context.Context, tid, gid uuid.UUID) (*domain.TargetGroup, error) {
			if gid == groupID {
				return &domain.TargetGroup{
					ID:        groupID,
					TenantID:  tenantID,
					Name:      "Production Servers",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	scanSvc := &mockScanService{}
	h := handler.NewTargetGroupHandler(groupSvc, scanSvc)
	router := setupGroupRouter(h, tenantID, domain.RoleSecOps)

	// 1. Successo
	req := httptest.NewRequest(http.MethodGet, "/target-groups/"+groupID.String(), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Atteso 200, ottenuto %d", rr.Code)
	}

	// 2. Not Found
	missingID := uuid.New()
	req404 := httptest.NewRequest(http.MethodGet, "/target-groups/"+missingID.String(), nil)
	rr404 := httptest.NewRecorder()
	router.ServeHTTP(rr404, req404)

	if rr404.Code != http.StatusNotFound {
		t.Fatalf("Atteso 404, ottenuto %d", rr404.Code)
	}
}

func TestTargetGroupHandler_Create_SuccessAndValidation(t *testing.T) {
	tenantID := uuid.New()
	color := "#3b82f6"

	groupSvc := &mockTargetGroupService{
		createGroupFunc: func(ctx context.Context, input service.CreateGroupInput) (*domain.TargetGroup, error) {
			if input.Name == "" {
				return nil, domain.ErrValidation
			}
			return &domain.TargetGroup{
				ID:        uuid.New(),
				TenantID:  input.TenantID,
				Name:      input.Name,
				ColorTag:  input.ColorTag,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}
	scanSvc := &mockScanService{}
	h := handler.NewTargetGroupHandler(groupSvc, scanSvc)
	router := setupGroupRouter(h, tenantID, domain.RoleSecOps)

	// 1. Successo
	body := handler.CreateTargetGroupRequest{
		Name:     "Cloud VPC",
		ColorTag: &color,
	}
	jsonBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/target-groups", bytes.NewReader(jsonBytes))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Atteso 201 Created, ottenuto %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Validazione nome vuoto
	bodyInvalid := handler.CreateTargetGroupRequest{Name: ""}
	invalidBytes, _ := json.Marshal(bodyInvalid)
	reqBad := httptest.NewRequest(http.MethodPost, "/target-groups", bytes.NewReader(invalidBytes))
	rrBad := httptest.NewRecorder()
	router.ServeHTTP(rrBad, reqBad)

	if rrBad.Code != http.StatusBadRequest {
		t.Fatalf("Atteso 400 Bad Request, ottenuto %d", rrBad.Code)
	}
}

func TestTargetGroupHandler_Update_SuccessAndConflict(t *testing.T) {
	tenantID := uuid.New()
	groupID := uuid.New()

	groupSvc := &mockTargetGroupService{
		updateGroupFunc: func(ctx context.Context, input service.UpdateGroupInput) (*domain.TargetGroup, error) {
			if input.ParentID != nil && *input.ParentID == input.ID {
				return nil, domain.ErrValidation
			}
			return &domain.TargetGroup{
				ID:        input.ID,
				TenantID:  input.TenantID,
				Name:      input.Name,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}
	scanSvc := &mockScanService{}
	h := handler.NewTargetGroupHandler(groupSvc, scanSvc)
	router := setupGroupRouter(h, tenantID, domain.RoleSecOps)

	// 1. Aggiornamento valido
	validBody := handler.UpdateTargetGroupRequest{Name: "Renamed Group"}
	jsonBytes, _ := json.Marshal(validBody)
	req := httptest.NewRequest(http.MethodPut, "/target-groups/"+groupID.String(), bytes.NewReader(jsonBytes))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Atteso 200 OK, ottenuto %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Errore validazione (auto-referenza ciclo)
	badBody := handler.UpdateTargetGroupRequest{Name: "Cyclic", ParentID: &groupID}
	badBytes, _ := json.Marshal(badBody)
	reqBad := httptest.NewRequest(http.MethodPut, "/target-groups/"+groupID.String(), bytes.NewReader(badBytes))
	rrBad := httptest.NewRecorder()
	router.ServeHTTP(rrBad, reqBad)

	if rrBad.Code != http.StatusBadRequest {
		t.Fatalf("Atteso 400 Bad Request, ottenuto %d", rrBad.Code)
	}
}

func TestTargetGroupHandler_Delete_SuccessAndNotFound(t *testing.T) {
	tenantID := uuid.New()
	groupID := uuid.New()

	groupSvc := &mockTargetGroupService{
		deleteGroupFunc: func(ctx context.Context, tid, gid uuid.UUID) error {
			if gid == groupID {
				return nil
			}
			return domain.ErrNotFound
		},
	}
	scanSvc := &mockScanService{}
	h := handler.NewTargetGroupHandler(groupSvc, scanSvc)
	router := setupGroupRouter(h, tenantID, domain.RoleSecOps)

	// 1. Successo
	req := httptest.NewRequest(http.MethodDelete, "/target-groups/"+groupID.String(), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Atteso 200 OK, ottenuto %d", rr.Code)
	}

	// 2. Not found
	missingID := uuid.New()
	reqBad := httptest.NewRequest(http.MethodDelete, "/target-groups/"+missingID.String(), nil)
	rrBad := httptest.NewRecorder()
	router.ServeHTTP(rrBad, reqBad)

	if rrBad.Code != http.StatusNotFound {
		t.Fatalf("Atteso 404, ottenuto %d", rrBad.Code)
	}
}

func TestTargetGroupHandler_StartScan_Success(t *testing.T) {
	tenantID := uuid.New()
	groupID := uuid.New()
	profileID := uuid.New()

	groupSvc := &mockTargetGroupService{}
	scanSvc := &mockScanService{
		startGroupScanFunc: func(ctx context.Context, tid, gid, pid uuid.UUID) ([]*domain.Scan, error) {
			return []*domain.Scan{
				{
					ID:          uuid.New(),
					TenantID:    tid,
					TargetID:    uuid.New(),
					ProfileID:   pid,
					Status:      domain.ScanStatusQueued,
					TriggerType: domain.TriggerManual,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			}, nil
		},
	}
	h := handler.NewTargetGroupHandler(groupSvc, scanSvc)
	router := setupGroupRouter(h, tenantID, domain.RoleSecOps)

	reqBody := handler.StartGroupScanRequest{ProfileID: profileID}
	jsonBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/target-groups/"+groupID.String()+"/scan", bytes.NewReader(jsonBytes))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("Atteso 202 Accepted, ottenuto %d: %s", rr.Code, rr.Body.String())
	}
}

func TestTargetGroupHandler_RBAC_AuditorForbiddenOnWrite(t *testing.T) {
	tenantID := uuid.New()
	groupID := uuid.New()

	groupSvc := &mockTargetGroupService{}
	scanSvc := &mockScanService{}
	h := handler.NewTargetGroupHandler(groupSvc, scanSvc)

	// Ruolo Auditor (sola lettura)
	router := setupGroupRouter(h, tenantID, domain.RoleAuditor)

	// 1. POST /target-groups -> 403 Forbidden
	reqPost := httptest.NewRequest(http.MethodPost, "/target-groups", bytes.NewReader([]byte(`{"name":"test"}`)))
	rrPost := httptest.NewRecorder()
	router.ServeHTTP(rrPost, reqPost)
	if rrPost.Code != http.StatusForbidden {
		t.Errorf("POST atteso 403 Forbidden, ottenuto %d", rrPost.Code)
	}

	// 2. PUT /target-groups/{id} -> 403 Forbidden
	reqPut := httptest.NewRequest(http.MethodPut, "/target-groups/"+groupID.String(), bytes.NewReader([]byte(`{"name":"test"}`)))
	rrPut := httptest.NewRecorder()
	router.ServeHTTP(rrPut, reqPut)
	if rrPut.Code != http.StatusForbidden {
		t.Errorf("PUT atteso 403 Forbidden, ottenuto %d", rrPut.Code)
	}

	// 3. DELETE /target-groups/{id} -> 403 Forbidden
	reqDel := httptest.NewRequest(http.MethodDelete, "/target-groups/"+groupID.String(), nil)
	rrDel := httptest.NewRecorder()
	router.ServeHTTP(rrDel, reqDel)
	if rrDel.Code != http.StatusForbidden {
		t.Errorf("DELETE atteso 403 Forbidden, ottenuto %d", rrDel.Code)
	}

	// 4. POST /target-groups/{id}/scan -> 403 Forbidden
	reqScan := httptest.NewRequest(http.MethodPost, "/target-groups/"+groupID.String()+"/scan", bytes.NewReader([]byte(`{"profile_id":"`+uuid.New().String()+`"}`)))
	rrScan := httptest.NewRecorder()
	router.ServeHTTP(rrScan, reqScan)
	if rrScan.Code != http.StatusForbidden {
		t.Errorf("POST scan atteso 403 Forbidden, ottenuto %d", rrScan.Code)
	}
}
