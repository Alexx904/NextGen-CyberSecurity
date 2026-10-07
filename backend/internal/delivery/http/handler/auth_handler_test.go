package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// mockAuthService implementa service.AuthService per i test del controller
type mockAuthService struct {
	loginFunc        func(ctx context.Context, input service.LoginInput) (*domain.User, *auth.TokenPair, error)
	refreshTokenFunc func(ctx context.Context, refreshToken string) (*auth.TokenPair, error)
	getCurrentFunc   func(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error)
}

func (m *mockAuthService) Login(ctx context.Context, input service.LoginInput) (*service.LoginOutput, error) {
	u, t, err := m.loginFunc(ctx, input)
	if err != nil {
		return nil, err
	}
	return &service.LoginOutput{
		User:   u,
		Tokens: t,
	}, nil
}

func (m *mockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	return m.refreshTokenFunc(ctx, refreshToken)
}

func (m *mockAuthService) ValidateToken(ctx context.Context, token string) (*auth.UserClaims, error) {
	return nil, nil
}

func (m *mockAuthService) GetCurrentUser(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error) {
	return m.getCurrentFunc(ctx, tenantID, userID)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	expectedUser := &domain.User{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Email:     "user@test.local",
		FullName:  "Test User",
		Role:      domain.RoleSecOps,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	expectedTokens := &auth.TokenPair{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
	}

	mockSvc := &mockAuthService{
		loginFunc: func(ctx context.Context, input service.LoginInput) (*domain.User, *auth.TokenPair, error) {
			if input.Email == "user@test.local" && input.Password == "Password123!" {
				return expectedUser, expectedTokens, nil
			}
			return nil, nil, domain.ErrUnauthorized
		},
	}

	h := handler.NewAuthHandler(mockSvc)

	body, _ := json.Marshal(map[string]string{
		"email":    "user@test.local",
		"password": "Password123!",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Atteso status 200, ottenuto %d", rec.Code)
	}

	var env response.Envelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	if !env.Success {
		t.Fatalf("Atteso success = true")
	}
}

func TestAuthHandler_Login_ValidationAndAuthErrors(t *testing.T) {
	mockSvc := &mockAuthService{
		loginFunc: func(ctx context.Context, input service.LoginInput) (*domain.User, *auth.TokenPair, error) {
			if input.Email == "" || input.Password == "" {
				return nil, nil, domain.ErrValidation
			}
			return nil, nil, domain.ErrUnauthorized
		},
	}

	h := handler.NewAuthHandler(mockSvc)

	// 1. JSON non valido
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("{invalid-json")))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Atteso 400 per JSON non valido, ottenuto %d", rec.Code)
	}

	// 2. Campi vuoti (ErrValidation)
	body, _ := json.Marshal(map[string]string{"email": "", "password": ""})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	rec2 := httptest.NewRecorder()
	h.Login(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("Atteso 400 per validazione fallita, ottenuto %d", rec2.Code)
	}

	// 3. Credenziali errate (ErrUnauthorized)
	body3, _ := json.Marshal(map[string]string{"email": "wrong@test.local", "password": "wrong"})
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body3))
	rec3 := httptest.NewRecorder()
	h.Login(rec3, req3)
	if rec3.Code != http.StatusUnauthorized {
		t.Errorf("Atteso 401 per credenziali errate, ottenuto %d", rec3.Code)
	}
}

func TestAuthHandler_Me(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	expectedUser := &domain.User{
		ID:        userID,
		TenantID:  tenantID,
		Email:     "admin@security.local",
		FullName:  "Admin User",
		Role:      domain.RoleTenantAdmin,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	mockSvc := &mockAuthService{
		getCurrentFunc: func(ctx context.Context, tID, uID uuid.UUID) (*domain.User, error) {
			if tID == tenantID && uID == userID {
				return expectedUser, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	h := handler.NewAuthHandler(mockSvc)

	// 1. Successo con contesto autenticato
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	claims := &auth.UserClaims{
		UserID:    userID,
		TenantID:  tenantID,
		Role:      domain.RoleTenantAdmin,
		Email:     "admin@security.local",
		TokenType: auth.TokenTypeAccess,
	}
	req = req.WithContext(middleware.WithUserClaims(req.Context(), claims))
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Atteso 200 OK per /me, ottenuto %d", rec.Code)
	}

	// 2. Senza contesto di autenticazione
	reqUnauth := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	recUnauth := httptest.NewRecorder()
	h.Me(recUnauth, reqUnauth)

	if recUnauth.Code != http.StatusUnauthorized {
		t.Errorf("Atteso 401 per richiesta non autenticata a /me, ottenuto %d", recUnauth.Code)
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	expectedTokens := &auth.TokenPair{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
	}

	mockSvc := &mockAuthService{
		refreshTokenFunc: func(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
			if refreshToken == "valid-refresh-token" {
				return expectedTokens, nil
			}
			return nil, domain.ErrUnauthorized
		},
	}

	h := handler.NewAuthHandler(mockSvc)

	// 1. Successo
	body, _ := json.Marshal(map[string]string{"refresh_token": "valid-refresh-token"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.RefreshToken(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Atteso 200 per refresh token valido, ottenuto %d", rec.Code)
	}

	// 2. Token mancante
	bodyEmpty, _ := json.Marshal(map[string]string{"refresh_token": ""})
	reqEmpty := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(bodyEmpty))
	recEmpty := httptest.NewRecorder()
	h.RefreshToken(recEmpty, reqEmpty)

	if recEmpty.Code != http.StatusBadRequest {
		t.Errorf("Atteso 400 per token vuoto, ottenuto %d", recEmpty.Code)
	}

	// 3. Token non valido
	bodyInvalid, _ := json.Marshal(map[string]string{"refresh_token": "invalid-token"})
	reqInvalid := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(bodyInvalid))
	recInvalid := httptest.NewRecorder()
	h.RefreshToken(recInvalid, reqInvalid)

	if recInvalid.Code != http.StatusUnauthorized {
		t.Errorf("Atteso 401 per token non valido, ottenuto %d", recInvalid.Code)
	}
}
