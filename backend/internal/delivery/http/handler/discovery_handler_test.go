package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/handler"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
)

type mockDiscoveryService struct {
	discoverSubnetFunc    func(ctx context.Context, tenantID uuid.UUID, req domain.DiscoveryRequest) (*domain.DiscoveryResult, error)
	getLastResultsFunc    func(ctx context.Context, tenantID uuid.UUID) (*domain.DiscoveryResult, error)
	getSessionResultsFunc func(ctx context.Context, tenantID, sessionID uuid.UUID) (*domain.DiscoveryResult, error)
	lookupVendorFunc      func(mac string) string
}

func (m *mockDiscoveryService) DiscoverSubnet(ctx context.Context, tenantID uuid.UUID, req domain.DiscoveryRequest) (*domain.DiscoveryResult, error) {
	if m.discoverSubnetFunc != nil {
		return m.discoverSubnetFunc(ctx, tenantID, req)
	}
	return nil, nil
}

func (m *mockDiscoveryService) GetLastResults(ctx context.Context, tenantID uuid.UUID) (*domain.DiscoveryResult, error) {
	if m.getLastResultsFunc != nil {
		return m.getLastResultsFunc(ctx, tenantID)
	}
	return nil, nil
}

func (m *mockDiscoveryService) GetSessionResults(ctx context.Context, tenantID, sessionID uuid.UUID) (*domain.DiscoveryResult, error) {
	if m.getSessionResultsFunc != nil {
		return m.getSessionResultsFunc(ctx, tenantID, sessionID)
	}
	return nil, nil
}

func (m *mockDiscoveryService) LookupVendor(mac string) string {
	if m.lookupVendorFunc != nil {
		return m.lookupVendorFunc(mac)
	}
	return ""
}

func setupDiscoveryContext(tenantID uuid.UUID) context.Context {
	claims := &auth.UserClaims{
		UserID:    uuid.New(),
		TenantID:  tenantID,
		Role:      domain.RoleSecOps,
		Email:     "secops@test.local",
		TokenType: auth.TokenTypeAccess,
	}
	return middleware.WithUserClaims(context.Background(), claims)
}

func TestDiscoveryHandler_Scan(t *testing.T) {
	tenantID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockDisc := &mockDiscoveryService{
			discoverSubnetFunc: func(ctx context.Context, tID uuid.UUID, req domain.DiscoveryRequest) (*domain.DiscoveryResult, error) {
				if tID != tenantID {
					t.Fatalf("atteso tenantID %v, ottenuto %v", tenantID, tID)
				}
				if req.Subnet != "192.168.1.0/24" {
					t.Fatalf("attesa subnet '192.168.1.0/24', ottenuta '%s'", req.Subnet)
				}
				return &domain.DiscoveryResult{
					ID:              uuid.New(),
					TenantID:        tenantID,
					Subnet:          req.Subnet,
					TotalDiscovered: 1,
					DurationSeconds: 1.5,
					Devices: []domain.DiscoveredDevice{
						{
							IP:        "192.168.1.1",
							MAC:       "00:11:22:33:44:55",
							Vendor:    "Cisco Systems",
							Hostname:  "router.lan",
							LatencyMs: 0.8,
							Status:    "up",
						},
					},
				}, nil
			},
		}

		h := handler.NewDiscoveryHandler(mockDisc, &mockTargetService{}, &mockTargetGroupService{}, nil)

		reqBody := `{"subnet": "192.168.1.0/24", "include_common_ports": false}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/network-discovery/scan", bytes.NewBufferString(reqBody))
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.Scan(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d: %s", rr.Code, rr.Body.String())
		}

		var env response.Envelope
		if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
			t.Fatalf("unmarshal envelope fallito: %v", err)
		}
		if !env.Success {
			t.Fatalf("atteso success=true, errore: %+v", env.Error)
		}
	})

	t.Run("InvalidSubnet_Error", func(t *testing.T) {
		mockDisc := &mockDiscoveryService{
			discoverSubnetFunc: func(ctx context.Context, tID uuid.UUID, req domain.DiscoveryRequest) (*domain.DiscoveryResult, error) {
				return nil, domain.ErrTargetInvalid
			},
		}

		h := handler.NewDiscoveryHandler(mockDisc, &mockTargetService{}, &mockTargetGroupService{}, nil)

		reqBody := `{"subnet": "invalid-subnet"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/network-discovery/scan", bytes.NewBufferString(reqBody))
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.Scan(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("atteso status 400, ottenuto %d", rr.Code)
		}
	})

	t.Run("EmptySubnet_Error", func(t *testing.T) {
		h := handler.NewDiscoveryHandler(&mockDiscoveryService{}, &mockTargetService{}, &mockTargetGroupService{}, nil)

		reqBody := `{"subnet": "   "}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/network-discovery/scan", bytes.NewBufferString(reqBody))
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.Scan(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("atteso status 400, ottenuto %d", rr.Code)
		}
	})

	t.Run("ScannerExecutionFailure", func(t *testing.T) {
		mockDisc := &mockDiscoveryService{
			discoverSubnetFunc: func(ctx context.Context, tID uuid.UUID, req domain.DiscoveryRequest) (*domain.DiscoveryResult, error) {
				return nil, domain.ErrScannerExecFailed
			},
		}

		h := handler.NewDiscoveryHandler(mockDisc, &mockTargetService{}, &mockTargetGroupService{}, nil)

		reqBody := `{"subnet": "192.168.1.0/24"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/network-discovery/scan", bytes.NewBufferString(reqBody))
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.Scan(rr, req)

		if rr.Code != http.StatusBadGateway {
			t.Fatalf("atteso status 502, ottenuto %d", rr.Code)
		}
	})

	t.Run("MissingTenant_Unauthorized", func(t *testing.T) {
		h := handler.NewDiscoveryHandler(&mockDiscoveryService{}, &mockTargetService{}, &mockTargetGroupService{}, nil)

		reqBody := `{"subnet": "192.168.1.0/24"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/network-discovery/scan", bytes.NewBufferString(reqBody))
		rr := httptest.NewRecorder()

		h.Scan(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("atteso status 401, ottenuto %d", rr.Code)
		}
	})
}

func TestDiscoveryHandler_GetResults(t *testing.T) {
	tenantID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockDisc := &mockDiscoveryService{
			getLastResultsFunc: func(ctx context.Context, tID uuid.UUID) (*domain.DiscoveryResult, error) {
				return &domain.DiscoveryResult{
					ID:              uuid.New(),
					TenantID:        tenantID,
					Subnet:          "10.0.0.0/24",
					TotalDiscovered: 2,
					Devices: []domain.DiscoveredDevice{
						{IP: "10.0.0.1", Status: "up"},
						{IP: "10.0.0.2", Status: "up"},
					},
				}, nil
			},
		}

		h := handler.NewDiscoveryHandler(mockDisc, &mockTargetService{}, &mockTargetGroupService{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/network-discovery/results", nil)
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.GetResults(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mockDisc := &mockDiscoveryService{
			getLastResultsFunc: func(ctx context.Context, tID uuid.UUID) (*domain.DiscoveryResult, error) {
				return nil, domain.ErrNotFound
			},
		}

		h := handler.NewDiscoveryHandler(mockDisc, &mockTargetService{}, &mockTargetGroupService{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/network-discovery/results", nil)
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.GetResults(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("atteso status 404, ottenuto %d", rr.Code)
		}
	})

	t.Run("Unauthorized_NoTenant", func(t *testing.T) {
		h := handler.NewDiscoveryHandler(&mockDiscoveryService{}, &mockTargetService{}, &mockTargetGroupService{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/network-discovery/results", nil)
		rr := httptest.NewRecorder()

		h.GetResults(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("atteso status 401, ottenuto %d", rr.Code)
		}
	})
}

func TestDiscoveryHandler_ImportTargets(t *testing.T) {
	tenantID := uuid.New()
	groupID := uuid.New()

	t.Run("Success_WithGroupAndAutoDescription", func(t *testing.T) {
		mockGroup := &mockTargetGroupService{
			getGroupFunc: func(ctx context.Context, tID, gID uuid.UUID) (*domain.TargetGroup, error) {
				if tID != tenantID || gID != groupID {
					return nil, domain.ErrNotFound
				}
				return &domain.TargetGroup{ID: groupID, TenantID: tenantID, Name: "Office LAN"}, nil
			},
		}

		var createdInputs []service.CreateTargetInput
		mockTarget := &mockTargetService{
			createTargetFunc: func(ctx context.Context, input service.CreateTargetInput) (*domain.Target, error) {
				createdInputs = append(createdInputs, input)
				return &domain.Target{
					ID:          uuid.New(),
					TenantID:    input.TenantID,
					GroupID:     input.GroupID,
					Type:        domain.TargetTypeIP,
					Value:       input.Value,
					Description: input.Description,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, nil
			},
		}

		h := handler.NewDiscoveryHandler(&mockDiscoveryService{}, mockTarget, mockGroup, nil)

		payload := handler.ImportTargetsRequest{
			GroupID: &groupID,
			Devices: []handler.ImportDiscoveredDevice{
				{
					IP:       "192.168.1.15",
					Vendor:   "Apple Inc.",
					Hostname: "macbook-alessandro",
				},
				{
					IP:       "192.168.1.20",
					Vendor:   "Dell Inc.",
					Hostname: "",
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/network-discovery/import-targets", bytes.NewBuffer(bodyBytes))
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.ImportTargets(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d: %s", rr.Code, rr.Body.String())
		}

		var env response.Envelope
		if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
			t.Fatalf("unmarshal envelope fallito: %v", err)
		}

		dataBytes, _ := json.Marshal(env.Data)
		var resp handler.ImportTargetsResponse
		_ = json.Unmarshal(dataBytes, &resp)

		if resp.TotalImported != 2 {
			t.Errorf("attesi 2 target importati, ottenuti %d", resp.TotalImported)
		}
		if resp.TotalSkipped != 0 {
			t.Errorf("attesi 0 target scartati, ottenuti %d", resp.TotalSkipped)
		}
		if len(createdInputs) != 2 {
			t.Fatalf("attesi 2 CreateTarget, eseguiti %d", len(createdInputs))
		}
		if *createdInputs[0].Description != "Discovered LAN device (Apple Inc. - macbook-alessandro)" {
			t.Errorf("descrizione non formattata come atteso: %s", *createdInputs[0].Description)
		}
		if *createdInputs[1].Description != "Discovered LAN device (Dell Inc.)" {
			t.Errorf("descrizione non formattata come atteso: %s", *createdInputs[1].Description)
		}
	})

	t.Run("Success_NoGroup_CustomDescription", func(t *testing.T) {
		customDesc := "Custom Server Desc"
		mockTarget := &mockTargetService{
			createTargetFunc: func(ctx context.Context, input service.CreateTargetInput) (*domain.Target, error) {
				return &domain.Target{
					ID:          uuid.New(),
					TenantID:    input.TenantID,
					GroupID:     nil,
					Type:        domain.TargetTypeIP,
					Value:       input.Value,
					Description: input.Description,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, nil
			},
		}

		h := handler.NewDiscoveryHandler(&mockDiscoveryService{}, mockTarget, &mockTargetGroupService{}, nil)

		payload := handler.ImportTargetsRequest{
			GroupID: nil,
			Devices: []handler.ImportDiscoveredDevice{
				{
					IP:          "10.0.0.5",
					Description: &customDesc,
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/network-discovery/import-targets", bytes.NewBuffer(bodyBytes))
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.ImportTargets(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}
	})

	t.Run("GroupNotFound_Error", func(t *testing.T) {
		mockGroup := &mockTargetGroupService{
			getGroupFunc: func(ctx context.Context, tID, gID uuid.UUID) (*domain.TargetGroup, error) {
				return nil, domain.ErrNotFound
			},
		}

		h := handler.NewDiscoveryHandler(&mockDiscoveryService{}, &mockTargetService{}, mockGroup, nil)

		payload := handler.ImportTargetsRequest{
			GroupID: &groupID,
			Devices: []handler.ImportDiscoveredDevice{
				{IP: "192.168.1.10"},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/network-discovery/import-targets", bytes.NewBuffer(bodyBytes))
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.ImportTargets(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("atteso status 404 GROUP_NOT_FOUND, ottenuto %d", rr.Code)
		}
	})

	t.Run("EmptyDeviceList_Error", func(t *testing.T) {
		h := handler.NewDiscoveryHandler(&mockDiscoveryService{}, &mockTargetService{}, &mockTargetGroupService{}, nil)

		payload := handler.ImportTargetsRequest{
			Devices: []handler.ImportDiscoveredDevice{},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/network-discovery/import-targets", bytes.NewBuffer(bodyBytes))
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.ImportTargets(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("atteso status 400 EMPTY_DEVICE_LIST, ottenuto %d", rr.Code)
		}
	})

	t.Run("PartialSuccess_Resilience_ConflictAndBlacklist", func(t *testing.T) {
		mockTarget := &mockTargetService{
			createTargetFunc: func(ctx context.Context, input service.CreateTargetInput) (*domain.Target, error) {
				switch input.Value {
				case "192.168.1.100":
					return &domain.Target{
						ID:        uuid.New(),
						TenantID:  input.TenantID,
						Value:     input.Value,
						CreatedAt: time.Now(),
					}, nil
				case "192.168.1.101":
					return nil, domain.ErrConflict
				case "192.168.1.102":
					return nil, domain.ErrTargetBlacklisted
				case "192.168.1.103":
					return nil, domain.ErrTargetInvalid
				default:
					return nil, errors.New("database disk full")
				}
			},
		}

		h := handler.NewDiscoveryHandler(&mockDiscoveryService{}, mockTarget, &mockTargetGroupService{}, nil)

		payload := handler.ImportTargetsRequest{
			Devices: []handler.ImportDiscoveredDevice{
				{IP: "192.168.1.100"}, // Ok
				{IP: "192.168.1.101"}, // Conflict
				{IP: "192.168.1.102"}, // Blacklisted
				{IP: "192.168.1.103"}, // Invalid IP
				{IP: "192.168.1.100"}, // Duplicate intra-batch
				{IP: "   "},           // Empty IP
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/network-discovery/import-targets", bytes.NewBuffer(bodyBytes))
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.ImportTargets(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d: %s", rr.Code, rr.Body.String())
		}

		var env response.Envelope
		if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
			t.Fatalf("unmarshal fallito: %v", err)
		}

		dataBytes, _ := json.Marshal(env.Data)
		var resp handler.ImportTargetsResponse
		_ = json.Unmarshal(dataBytes, &resp)

		if resp.TotalImported != 1 {
			t.Errorf("atteso 1 target importato, ottenuti %d", resp.TotalImported)
		}
		if resp.TotalSkipped != 5 {
			t.Errorf("attesi 5 target scartati, ottenuti %d", resp.TotalSkipped)
		}
		if len(resp.Errors) != 5 {
			t.Errorf("attesi 5 errori registrati, ottenuti %d", len(resp.Errors))
		}
	})
}

func TestDiscoveryHandler_Stream(t *testing.T) {
	tenantID := uuid.New()
	sessionID := uuid.New()

	t.Run("MissingSessionID_BadRequest", func(t *testing.T) {
		h := handler.NewDiscoveryHandler(&mockDiscoveryService{}, &mockTargetService{}, &mockTargetGroupService{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/network-discovery/stream", nil)
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.Stream(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("atteso status 400 per session_id mancante, ottenuto %d", rr.Code)
		}
	})

	t.Run("InvalidSessionID_BadRequest", func(t *testing.T) {
		h := handler.NewDiscoveryHandler(&mockDiscoveryService{}, &mockTargetService{}, &mockTargetGroupService{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/network-discovery/stream?session_id=not-a-uuid", nil)
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.Stream(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("atteso status 400 per session_id non valido, ottenuto %d", rr.Code)
		}
	})

	t.Run("Unauthorized_NoTenant", func(t *testing.T) {
		h := handler.NewDiscoveryHandler(&mockDiscoveryService{}, &mockTargetService{}, &mockTargetGroupService{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/network-discovery/stream?session_id="+sessionID.String(), nil)
		rr := httptest.NewRecorder()

		h.Stream(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("atteso status 401 per richiesta non autenticata, ottenuto %d", rr.Code)
		}
	})

	t.Run("Success_LiveStreamingAndComplete", func(t *testing.T) {
		broadcaster := service.NewEventBroadcaster(10, nil)
		defer broadcaster.Close()

		mockDisc := &mockDiscoveryService{}
		h := handler.NewDiscoveryHandler(mockDisc, &mockTargetService{}, &mockTargetGroupService{}, broadcaster)

		ctx, cancel := context.WithCancel(setupDiscoveryContext(tenantID))
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/network-discovery/stream?session_id="+sessionID.String(), nil)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		// Esegui Stream in una goroutine per consentire l'emissione di eventi
		doneCh := make(chan struct{})
		go func() {
			h.Stream(rr, req)
			close(doneCh)
		}()

		// Breve attesa per consentire la subscription
		time.Sleep(30 * time.Millisecond)

		// Pubblica un evento device_discovered
		topic := domain.TenantDiscoveryTopic(tenantID, sessionID)
		broadcaster.Publish(topic, domain.SSEEvent{
			Event: domain.EventDeviceDiscovered,
			Data: domain.DeviceDiscoveredPayload{
				SessionID: sessionID,
				Device: domain.DiscoveredDevice{
					IP:     "192.168.1.10",
					Vendor: "Netgear",
					Status: "up",
				},
			},
		})

		// Pubblica evento finale complete (che causerà il return di Stream)
		broadcaster.Publish(topic, domain.SSEEvent{
			Event: domain.EventDiscoveryComplete,
			Data: domain.DiscoveryCompletePayload{
				SessionID:       sessionID,
				TotalDiscovered: 1,
				DurationSeconds: 1.5,
			},
		})

		select {
		case <-doneCh:
			// Stream terminato con successo dopo complete
		case <-time.After(1 * time.Second):
			cancel()
			t.Fatal("timeout in attesa del completamento dello stream")
		}

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}

		body := rr.Body.String()
		if !strings.Contains(body, "event: device_discovered") {
			t.Errorf("il body SSE non contiene 'event: device_discovered':\n%s", body)
		}
		if !strings.Contains(body, "192.168.1.10") {
			t.Errorf("il body SSE non contiene IP scoperto:\n%s", body)
		}
		if !strings.Contains(body, "event: complete") {
			t.Errorf("il body SSE non contiene 'event: complete':\n%s", body)
		}
	})

	t.Run("Success_ReplayExistingDevices", func(t *testing.T) {
		completedAt := time.Now()
		mockDisc := &mockDiscoveryService{
			getSessionResultsFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.DiscoveryResult, error) {
				return &domain.DiscoveryResult{
					ID:              sessionID,
					TenantID:        tID,
					TotalDiscovered: 1,
					CompletedAt:     completedAt,
					Devices: []domain.DiscoveredDevice{
						{
							IP:     "192.168.1.5",
							Vendor: "Synology",
							Status: "up",
						},
					},
				}, nil
			},
		}

		broadcaster := service.NewEventBroadcaster(10, nil)
		defer broadcaster.Close()

		h := handler.NewDiscoveryHandler(mockDisc, &mockTargetService{}, &mockTargetGroupService{}, broadcaster)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/network-discovery/stream?session_id="+sessionID.String(), nil)
		req = req.WithContext(setupDiscoveryContext(tenantID))
		rr := httptest.NewRecorder()

		h.Stream(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}

		body := rr.Body.String()
		if !strings.Contains(body, "192.168.1.5") {
			t.Errorf("atteso dispositivo in replay 192.168.1.5, ottenuto:\n%s", body)
		}
		if !strings.Contains(body, "event: complete") {
			t.Errorf("atteso evento complete per sessione già conclusa, ottenuto:\n%s", body)
		}
	})
}
