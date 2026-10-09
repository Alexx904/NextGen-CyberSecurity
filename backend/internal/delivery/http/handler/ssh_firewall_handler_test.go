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
	"golang.org/x/crypto/ssh"

	"nextgen-cybersec/internal/delivery/http/handler"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
)

type mockSSHServiceForHandler struct {
	executeCommandFunc func(ctx context.Context, tenantID, serverID uuid.UUID, cmd string) (string, string, int, error)
	createServerFunc   func(ctx context.Context, tenantID uuid.UUID, req service.CreateSSHServerRequest) (*domain.SSHServer, error)
	getServerFunc      func(ctx context.Context, tenantID, serverID uuid.UUID) (*domain.SSHServer, error)
	listServersFunc    func(ctx context.Context, tenantID uuid.UUID) ([]*domain.SSHServer, error)
	updateServerFunc   func(ctx context.Context, tenantID, serverID uuid.UUID, req service.UpdateSSHServerRequest) (*domain.SSHServer, error)
	deleteServerFunc   func(ctx context.Context, tenantID, serverID uuid.UUID) error
	testConnectionFunc func(ctx context.Context, tenantID, serverID uuid.UUID) (*service.SSHConnectionTestResult, error)
	getRawSSHClientFunc func(ctx context.Context, tenantID, serverID uuid.UUID) (*ssh.Client, error)
}

func (m *mockSSHServiceForHandler) ExecuteCommand(ctx context.Context, tenantID, serverID uuid.UUID, cmd string) (string, string, int, error) {
	if m.executeCommandFunc != nil {
		return m.executeCommandFunc(ctx, tenantID, serverID, cmd)
	}
	return "", "", 0, nil
}
func (m *mockSSHServiceForHandler) CreateServer(ctx context.Context, tenantID uuid.UUID, req service.CreateSSHServerRequest) (*domain.SSHServer, error) {
	if m.createServerFunc != nil {
		return m.createServerFunc(ctx, tenantID, req)
	}
	return nil, nil
}
func (m *mockSSHServiceForHandler) GetServer(ctx context.Context, tenantID, serverID uuid.UUID) (*domain.SSHServer, error) {
	if m.getServerFunc != nil {
		return m.getServerFunc(ctx, tenantID, serverID)
	}
	return nil, nil
}
func (m *mockSSHServiceForHandler) ListServers(ctx context.Context, tenantID uuid.UUID) ([]*domain.SSHServer, error) {
	if m.listServersFunc != nil {
		return m.listServersFunc(ctx, tenantID)
	}
	return nil, nil
}
func (m *mockSSHServiceForHandler) UpdateServer(ctx context.Context, tenantID, serverID uuid.UUID, req service.UpdateSSHServerRequest) (*domain.SSHServer, error) {
	if m.updateServerFunc != nil {
		return m.updateServerFunc(ctx, tenantID, serverID, req)
	}
	return nil, nil
}
func (m *mockSSHServiceForHandler) DeleteServer(ctx context.Context, tenantID, serverID uuid.UUID) error {
	if m.deleteServerFunc != nil {
		return m.deleteServerFunc(ctx, tenantID, serverID)
	}
	return nil
}
func (m *mockSSHServiceForHandler) TestConnection(ctx context.Context, tenantID, serverID uuid.UUID) (*service.SSHConnectionTestResult, error) {
	if m.testConnectionFunc != nil {
		return m.testConnectionFunc(ctx, tenantID, serverID)
	}
	return nil, nil
}
func (m *mockSSHServiceForHandler) GetRawSSHClient(ctx context.Context, tenantID, serverID uuid.UUID) (*ssh.Client, error) {
	if m.getRawSSHClientFunc != nil {
		return m.getRawSSHClientFunc(ctx, tenantID, serverID)
	}
	return nil, nil
}

type mockFirewallServiceForHandler struct {
	getStatusFunc            func(ctx context.Context, tenantID, serverID uuid.UUID) (*domain.FirewallStatus, error)
	addRuleFunc              func(ctx context.Context, tenantID, serverID uuid.UUID, params domain.FirewallRuleParams, withSafetyTimer bool, timeoutSec int) (*domain.FirewallStatus, error)
	deleteRuleFunc           func(ctx context.Context, tenantID, serverID uuid.UUID, chain domain.ChainName, lineNum int, withSafetyTimer bool, timeoutSec int) (*domain.FirewallStatus, error)
	applyWithSafetyTimerFunc func(ctx context.Context, tenantID, serverID uuid.UUID, timeoutSec int) (*domain.FirewallStatus, error)
	confirmFunc              func(ctx context.Context, tenantID, serverID uuid.UUID) error
	rollbackFunc             func(ctx context.Context, tenantID, serverID uuid.UUID, backupID *uuid.UUID) error
}

func (m *mockFirewallServiceForHandler) GetStatus(ctx context.Context, tenantID, serverID uuid.UUID) (*domain.FirewallStatus, error) {
	if m.getStatusFunc != nil {
		return m.getStatusFunc(ctx, tenantID, serverID)
	}
	return nil, nil
}
func (m *mockFirewallServiceForHandler) AddRule(ctx context.Context, tenantID, serverID uuid.UUID, params domain.FirewallRuleParams, withSafetyTimer bool, timeoutSec int) (*domain.FirewallStatus, error) {
	if m.addRuleFunc != nil {
		return m.addRuleFunc(ctx, tenantID, serverID, params, withSafetyTimer, timeoutSec)
	}
	return nil, nil
}
func (m *mockFirewallServiceForHandler) DeleteRule(ctx context.Context, tenantID, serverID uuid.UUID, chain domain.ChainName, lineNum int, withSafetyTimer bool, timeoutSec int) (*domain.FirewallStatus, error) {
	if m.deleteRuleFunc != nil {
		return m.deleteRuleFunc(ctx, tenantID, serverID, chain, lineNum, withSafetyTimer, timeoutSec)
	}
	return nil, nil
}
func (m *mockFirewallServiceForHandler) ApplyWithSafetyTimer(ctx context.Context, tenantID, serverID uuid.UUID, timeoutSec int) (*domain.FirewallStatus, error) {
	if m.applyWithSafetyTimerFunc != nil {
		return m.applyWithSafetyTimerFunc(ctx, tenantID, serverID, timeoutSec)
	}
	return nil, nil
}
func (m *mockFirewallServiceForHandler) Confirm(ctx context.Context, tenantID, serverID uuid.UUID) error {
	if m.confirmFunc != nil {
		return m.confirmFunc(ctx, tenantID, serverID)
	}
	return nil
}
func (m *mockFirewallServiceForHandler) Rollback(ctx context.Context, tenantID, serverID uuid.UUID, backupID *uuid.UUID) error {
	if m.rollbackFunc != nil {
		return m.rollbackFunc(ctx, tenantID, serverID, backupID)
	}
	return nil
}

func setupSSHContext(tenantID uuid.UUID) context.Context {
	claims := &auth.UserClaims{
		UserID:    uuid.New(),
		TenantID:  tenantID,
		Role:      domain.RoleSecOps,
		Email:     "secops@test.local",
		TokenType: auth.TokenTypeAccess,
	}
	return middleware.WithUserClaims(context.Background(), claims)
}

func TestSSHFirewallHandler_ServerCRUD(t *testing.T) {
	tenantID := uuid.New()
	serverID := uuid.New()

	t.Run("ListServers_Success", func(t *testing.T) {
		mockSSH := &mockSSHServiceForHandler{
			listServersFunc: func(ctx context.Context, tID uuid.UUID) ([]*domain.SSHServer, error) {
				return []*domain.SSHServer{
					{
						ID:        serverID,
						TenantID:  tenantID,
						Name:      "Web Server",
						Host:      "192.168.1.10",
						Port:      22,
						Username:  "root",
						AuthType:  domain.SSHAuthTypeKey,
						Status:    domain.SSHServerStatusOnline,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}, nil
			},
		}

		h := handler.NewSSHFirewallHandler(mockSSH, &mockFirewallServiceForHandler{})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ssh/servers", nil)
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		h.ListServers(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}

		var env response.Envelope
		_ = json.Unmarshal(rr.Body.Bytes(), &env)
		if !env.Success {
			t.Fatalf("atteso success=true, errore: %+v", env.Error)
		}
	})

	t.Run("GetServer_Success", func(t *testing.T) {
		mockSSH := &mockSSHServiceForHandler{
			getServerFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.SSHServer, error) {
				return &domain.SSHServer{
					ID:        serverID,
					TenantID:  tenantID,
					Name:      "Web Server",
					Host:      "192.168.1.10",
					Port:      22,
					Username:  "root",
					AuthType:  domain.SSHAuthTypePassword,
					Status:    domain.SSHServerStatusOnline,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			},
		}

		h := handler.NewSSHFirewallHandler(mockSSH, &mockFirewallServiceForHandler{})
		r := chi.NewRouter()
		r.Get("/api/v1/ssh/servers/{id}", h.GetServer)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/ssh/servers/"+serverID.String(), nil)
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}
	})

	t.Run("GetServer_NotFound", func(t *testing.T) {
		mockSSH := &mockSSHServiceForHandler{
			getServerFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.SSHServer, error) {
				return nil, domain.ErrSSHServerNotFound
			},
		}

		h := handler.NewSSHFirewallHandler(mockSSH, &mockFirewallServiceForHandler{})
		r := chi.NewRouter()
		r.Get("/api/v1/ssh/servers/{id}", h.GetServer)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/ssh/servers/"+serverID.String(), nil)
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("atteso status 404, ottenuto %d", rr.Code)
		}
	})

	t.Run("CreateServer_Success", func(t *testing.T) {
		mockSSH := &mockSSHServiceForHandler{
			createServerFunc: func(ctx context.Context, tID uuid.UUID, req service.CreateSSHServerRequest) (*domain.SSHServer, error) {
				return &domain.SSHServer{
					ID:        uuid.New(),
					TenantID:  tID,
					Name:      req.Name,
					Host:      req.Host,
					Port:      req.Port,
					Username:  req.Username,
					AuthType:  req.AuthType,
					Status:    domain.SSHServerStatusUnchecked,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			},
		}

		h := handler.NewSSHFirewallHandler(mockSSH, &mockFirewallServiceForHandler{})
		payload := `{"name":"Database Server","host":"10.0.0.5","port":22,"username":"admin","auth_type":"PASSWORD","raw_secret":"secretpass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ssh/servers", bytes.NewBufferString(payload))
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		h.CreateServer(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("atteso status 201 Created, ottenuto %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("CreateServer_DuplicateName", func(t *testing.T) {
		mockSSH := &mockSSHServiceForHandler{
			createServerFunc: func(ctx context.Context, tID uuid.UUID, req service.CreateSSHServerRequest) (*domain.SSHServer, error) {
				return nil, domain.ErrSSHServerDuplicateName
			},
		}

		h := handler.NewSSHFirewallHandler(mockSSH, &mockFirewallServiceForHandler{})
		payload := `{"name":"Database Server","host":"10.0.0.5","port":22,"username":"admin","auth_type":"PASSWORD","raw_secret":"secretpass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ssh/servers", bytes.NewBufferString(payload))
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		h.CreateServer(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("atteso status 409 Conflict, ottenuto %d", rr.Code)
		}
	})

	t.Run("UpdateServer_Success", func(t *testing.T) {
		mockSSH := &mockSSHServiceForHandler{
			updateServerFunc: func(ctx context.Context, tID, sID uuid.UUID, req service.UpdateSSHServerRequest) (*domain.SSHServer, error) {
				return &domain.SSHServer{
					ID:        sID,
					TenantID:  tID,
					Name:      req.Name,
					Host:      req.Host,
					Port:      req.Port,
					Username:  req.Username,
					AuthType:  req.AuthType,
					Status:    domain.SSHServerStatusOnline,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			},
		}

		h := handler.NewSSHFirewallHandler(mockSSH, &mockFirewallServiceForHandler{})
		r := chi.NewRouter()
		r.Put("/api/v1/ssh/servers/{id}", h.UpdateServer)

		payload := `{"name":"Updated Name","host":"10.0.0.5","port":2222,"username":"admin","auth_type":"PASSWORD"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/ssh/servers/"+serverID.String(), bytes.NewBufferString(payload))
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}
	})

	t.Run("DeleteServer_Success", func(t *testing.T) {
		mockSSH := &mockSSHServiceForHandler{
			deleteServerFunc: func(ctx context.Context, tID, sID uuid.UUID) error {
				return nil
			},
		}

		h := handler.NewSSHFirewallHandler(mockSSH, &mockFirewallServiceForHandler{})
		r := chi.NewRouter()
		r.Delete("/api/v1/ssh/servers/{id}", h.DeleteServer)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/ssh/servers/"+serverID.String(), nil)
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}
	})

	t.Run("TestConnection_Success", func(t *testing.T) {
		mockSSH := &mockSSHServiceForHandler{
			testConnectionFunc: func(ctx context.Context, tID, sID uuid.UUID) (*service.SSHConnectionTestResult, error) {
				return &service.SSHConnectionTestResult{
					Online:    true,
					LatencyMs: 15,
					OSInfo:    "Linux 6.8.0",
					IsRoot:    true,
				}, nil
			},
		}

		h := handler.NewSSHFirewallHandler(mockSSH, &mockFirewallServiceForHandler{})
		r := chi.NewRouter()
		r.Post("/api/v1/ssh/servers/{id}/test-connection", h.TestConnection)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/ssh/servers/"+serverID.String()+"/test-connection", nil)
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}
	})
}

func TestSSHFirewallHandler_FirewallOperations(t *testing.T) {
	tenantID := uuid.New()
	serverID := uuid.New()

	t.Run("GetFirewallStatus_Success", func(t *testing.T) {
		mockFW := &mockFirewallServiceForHandler{
			getStatusFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.FirewallStatus, error) {
				return &domain.FirewallStatus{
					ServerID:             serverID,
					ActiveSafetyTimer:    false,
					SafetyTimeoutSeconds: 60,
					Chains: map[domain.ChainName]*domain.FirewallChain{
						domain.ChainInput: {
							Name:          domain.ChainInput,
							DefaultPolicy: domain.ActionAccept,
							Rules: []domain.FirewallRule{
								{LineNumber: 1, Target: domain.ActionAccept, Protocol: "tcp", Source: "0.0.0.0/0"},
							},
						},
					},
				}, nil
			},
		}

		h := handler.NewSSHFirewallHandler(&mockSSHServiceForHandler{}, mockFW)
		r := chi.NewRouter()
		r.Get("/api/v1/ssh/servers/{id}/firewall", h.GetFirewallStatus)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/ssh/servers/"+serverID.String()+"/firewall", nil)
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}
	})

	t.Run("AddFirewallRule_Success", func(t *testing.T) {
		mockFW := &mockFirewallServiceForHandler{
			addRuleFunc: func(ctx context.Context, tID, sID uuid.UUID, params domain.FirewallRuleParams, withSafetyTimer bool, timeoutSec int) (*domain.FirewallStatus, error) {
				return &domain.FirewallStatus{
					ServerID:             serverID,
					ActiveSafetyTimer:    withSafetyTimer,
					SafetyTimeoutSeconds: timeoutSec,
				}, nil
			},
		}

		h := handler.NewSSHFirewallHandler(&mockSSHServiceForHandler{}, mockFW)
		r := chi.NewRouter()
		r.Post("/api/v1/ssh/servers/{id}/firewall/rules", h.AddFirewallRule)

		payload := `{"chain":"INPUT","action":"ACCEPT","protocol":"tcp","port":"80","source":"0.0.0.0/0","with_safety_timer":true,"safety_timeout_seconds":60}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ssh/servers/"+serverID.String()+"/firewall/rules", bytes.NewBufferString(payload))
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}
	})

	t.Run("DeleteFirewallRule_Success", func(t *testing.T) {
		mockFW := &mockFirewallServiceForHandler{
			deleteRuleFunc: func(ctx context.Context, tID, sID uuid.UUID, chain domain.ChainName, lineNum int, withSafetyTimer bool, timeoutSec int) (*domain.FirewallStatus, error) {
				if lineNum != 3 || chain != domain.ChainInput {
					return nil, errors.New("parametri errati")
				}
				return &domain.FirewallStatus{
					ServerID:          serverID,
					ActiveSafetyTimer: withSafetyTimer,
				}, nil
			},
		}

		h := handler.NewSSHFirewallHandler(&mockSSHServiceForHandler{}, mockFW)
		r := chi.NewRouter()
		r.Delete("/api/v1/ssh/servers/{id}/firewall/rules/{line}", h.DeleteFirewallRule)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/ssh/servers/"+serverID.String()+"/firewall/rules/3?chain=INPUT", nil)
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("ApplyWithSafetyTimer_Success", func(t *testing.T) {
		mockFW := &mockFirewallServiceForHandler{
			applyWithSafetyTimerFunc: func(ctx context.Context, tID, sID uuid.UUID, timeoutSec int) (*domain.FirewallStatus, error) {
				return &domain.FirewallStatus{
					ServerID:             serverID,
					ActiveSafetyTimer:    true,
					SafetyTimeoutSeconds: timeoutSec,
				}, nil
			},
		}

		h := handler.NewSSHFirewallHandler(&mockSSHServiceForHandler{}, mockFW)
		r := chi.NewRouter()
		r.Post("/api/v1/ssh/servers/{id}/firewall/apply", h.ApplyWithSafetyTimer)

		payload := `{"safety_timeout_seconds":90}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ssh/servers/"+serverID.String()+"/firewall/apply", bytes.NewBufferString(payload))
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}
	})

	t.Run("ConfirmFirewall_Success", func(t *testing.T) {
		mockFW := &mockFirewallServiceForHandler{
			confirmFunc: func(ctx context.Context, tID, sID uuid.UUID) error {
				return nil
			},
		}

		h := handler.NewSSHFirewallHandler(&mockSSHServiceForHandler{}, mockFW)
		r := chi.NewRouter()
		r.Post("/api/v1/ssh/servers/{id}/firewall/confirm", h.ConfirmFirewall)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/ssh/servers/"+serverID.String()+"/firewall/confirm", nil)
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}
	})

	t.Run("RollbackFirewall_Success", func(t *testing.T) {
		mockFW := &mockFirewallServiceForHandler{
			rollbackFunc: func(ctx context.Context, tID, sID uuid.UUID, backupID *uuid.UUID) error {
				return nil
			},
		}

		h := handler.NewSSHFirewallHandler(&mockSSHServiceForHandler{}, mockFW)
		r := chi.NewRouter()
		r.Post("/api/v1/ssh/servers/{id}/firewall/rollback", h.RollbackFirewall)

		payload := `{}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ssh/servers/"+serverID.String()+"/firewall/rollback", bytes.NewBufferString(payload))
		req = req.WithContext(setupSSHContext(tenantID))
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("atteso status 200, ottenuto %d", rr.Code)
		}
	})
}

