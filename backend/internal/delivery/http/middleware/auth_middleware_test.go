package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/auth"
)

func setupTestRouter(tokenMgr auth.TokenManager) (*chi.Mux, uuid.UUID, uuid.UUID) {
	r := chi.NewRouter()

	tenantID := uuid.New()
	userID := uuid.New()

	r.Use(middleware.Authenticate(tokenMgr))
	r.Use(middleware.RequireTenant())

	r.Get("/test-protected", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tid, _ := middleware.GetTenantID(ctx)
		uid, _ := middleware.GetUserID(ctx)
		role, _ := middleware.GetUserRole(ctx)

		response.JSON(w, http.StatusOK, map[string]interface{}{
			"tenant_id": tid.String(),
			"user_id":   uid.String(),
			"role":      string(role),
		})
	})

	return r, tenantID, userID
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	tokenMgr, _ := auth.NewJWTManager("test-secret-key-12345", 15*time.Minute, 24*time.Hour)
	r, _, _ := setupTestRouter(tokenMgr)

	req := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Atteso 401 Unauthorized, ottenuto %d", rec.Code)
	}

	var env response.Envelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	if env.Success || env.Error == nil || env.Error.Code != "UNAUTHORIZED" {
		t.Errorf("Risposta di errore non corretta: %+v", env)
	}
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	tokenMgr, _ := auth.NewJWTManager("test-secret-key-12345", 15*time.Minute, 24*time.Hour)
	r, _, _ := setupTestRouter(tokenMgr)

	testCases := []string{
		"Bearer",
		"Basic dXNlcjpwYXNz",
		"TokenWithoutScheme",
		"Bearer   ",
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
		req.Header.Set("Authorization", tc)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Per header '%s': atteso 401, ottenuto %d", tc, rec.Code)
		}
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	mgr, _ := auth.NewJWTManager("test-secret-key-12345", 15*time.Minute, 24*time.Hour)
	r, tenantID, userID := setupTestRouter(mgr)

	user := &domain.User{
		ID:       userID,
		TenantID: tenantID,
		Email:    "expired@test.com",
		Role:     domain.RoleSecOps,
	}

	// Creiamo un token scaduto nel passato
	now := time.Now().Add(-2 * time.Hour)
	claims := auth.UserClaims{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Role:      user.Role,
		Email:     user.Email,
		TokenType: auth.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Minute)), // scaduto ~2h fa
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredTokenStr, _ := token.SignedString([]byte("test-secret-key-12345"))

	req := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredTokenStr)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Atteso 401 Unauthorized, ottenuto %d", rec.Code)
	}

	var env response.Envelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	if env.Error == nil || env.Error.Code != "TOKEN_EXPIRED" {
		t.Errorf("Atteso errore TOKEN_EXPIRED, ottenuto %+v", env.Error)
	}
}

func TestAuthMiddleware_RefreshTokenRejected(t *testing.T) {
	tokenMgr, _ := auth.NewJWTManager("test-secret-key-12345", 15*time.Minute, 24*time.Hour)
	r, tenantID, userID := setupTestRouter(tokenMgr)

	user := &domain.User{
		ID:       userID,
		TenantID: tenantID,
		Email:    "user@test.com",
		Role:     domain.RoleSecOps,
	}

	tokens, _ := tokenMgr.GenerateTokenPair(user)

	// Invia il refresh token invece dell'access token
	req := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.RefreshToken)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Atteso 401 Unauthorized per refresh token, ottenuto %d", rec.Code)
	}

	var env response.Envelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	if env.Error.Code != "INVALID_TOKEN_TYPE" {
		t.Errorf("Atteso INVALID_TOKEN_TYPE, ottenuto %s", env.Error.Code)
	}
}

func TestAuthMiddleware_ValidToken_Success(t *testing.T) {
	tokenMgr, _ := auth.NewJWTManager("test-secret-key-12345", 15*time.Minute, 24*time.Hour)
	r, tenantID, userID := setupTestRouter(tokenMgr)

	user := &domain.User{
		ID:       userID,
		TenantID: tenantID,
		Email:    "valid@test.com",
		Role:     domain.RoleSecOps,
	}

	tokens, _ := tokenMgr.GenerateTokenPair(user)

	req := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Atteso 200 OK, ottenuto %d. Body: %s", rec.Code, rec.Body.String())
	}

	var env response.Envelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	data := env.Data.(map[string]interface{})

	if data["tenant_id"] != tenantID.String() {
		t.Errorf("Tenant ID errato: %v", data["tenant_id"])
	}
	if data["user_id"] != userID.String() {
		t.Errorf("User ID errato: %v", data["user_id"])
	}
	if data["role"] != string(domain.RoleSecOps) {
		t.Errorf("Ruolo errato: %v", data["role"])
	}
}

func TestTenantMiddleware_SuperAdminOverride(t *testing.T) {
	tokenMgr, _ := auth.NewJWTManager("test-secret-key-12345", 15*time.Minute, 24*time.Hour)
	r, primaryTenantID, userID := setupTestRouter(tokenMgr)

	overrideTenantID := uuid.New()

	// SuperAdmin
	superAdmin := &domain.User{
		ID:       userID,
		TenantID: primaryTenantID,
		Email:    "superadmin@nextgen.local",
		Role:     domain.RoleSuperAdmin,
	}
	tokens, _ := tokenMgr.GenerateTokenPair(superAdmin)

	req := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("X-Tenant-ID", overrideTenantID.String())
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Atteso 200 OK per override SuperAdmin, ottenuto %d", rec.Code)
	}

	var env response.Envelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	data := env.Data.(map[string]interface{})
	if data["tenant_id"] != overrideTenantID.String() {
		t.Errorf("Atteso tenant_id %s, ottenuto %v", overrideTenantID, data["tenant_id"])
	}

	// SecOps che tenta di impostare un altro tenant (IDOR prevention)
	secOps := &domain.User{
		ID:       uuid.New(),
		TenantID: primaryTenantID,
		Email:    "secops@test.com",
		Role:     domain.RoleSecOps,
	}
	secOpsTokens, _ := tokenMgr.GenerateTokenPair(secOps)

	reqIDOR := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
	reqIDOR.Header.Set("Authorization", "Bearer "+secOpsTokens.AccessToken)
	reqIDOR.Header.Set("X-Tenant-ID", overrideTenantID.String())
	recIDOR := httptest.NewRecorder()

	r.ServeHTTP(recIDOR, reqIDOR)

	if recIDOR.Code != http.StatusForbidden {
		t.Fatalf("Atteso 403 Forbidden per tentativo IDOR da parte di non-SuperAdmin, ottenuto %d", recIDOR.Code)
	}
}

func TestAuthMiddleware_InvalidSignature(t *testing.T) {
	mgr1, _ := auth.NewJWTManager("secret-one-12345", 15*time.Minute, 24*time.Hour)
	mgr2, _ := auth.NewJWTManager("secret-two-12345", 15*time.Minute, 24*time.Hour)
	r, tenantID, userID := setupTestRouter(mgr2)

	user := &domain.User{
		ID:       userID,
		TenantID: tenantID,
		Email:    "tampered@test.com",
		Role:     domain.RoleSecOps,
	}

	tokens, _ := mgr1.GenerateTokenPair(user)

	req := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Atteso 401 per firma non valida, ottenuto %d", rec.Code)
	}
}

func TestTenantMiddleware_InvalidUUIDHeader(t *testing.T) {
	tokenMgr, _ := auth.NewJWTManager("test-secret-key-12345", 15*time.Minute, 24*time.Hour)
	r, primaryTenantID, userID := setupTestRouter(tokenMgr)

	superAdmin := &domain.User{
		ID:       userID,
		TenantID: primaryTenantID,
		Email:    "admin@test.com",
		Role:     domain.RoleSuperAdmin,
	}
	tokens, _ := tokenMgr.GenerateTokenPair(superAdmin)

	req := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("X-Tenant-ID", "not-a-valid-uuid")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Atteso 400 Bad Request per UUID malformato, ottenuto %d", rec.Code)
	}
}

func TestTenantMiddleware_MissingTenantInContext(t *testing.T) {
	r := chi.NewRouter()
	r.Use(middleware.RequireTenant())
	r.Get("/tenant-only", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/tenant-only", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Atteso 401 per tenant assente nel contesto, ottenuto %d", rec.Code)
	}
}

func TestAuthMiddleware_QueryToken_Success(t *testing.T) {
	tokenMgr, _ := auth.NewJWTManager("test-secret-key-12345", 15*time.Minute, 24*time.Hour)
	r, tenantID, userID := setupTestRouter(tokenMgr)

	user := &domain.User{
		ID:       userID,
		TenantID: tenantID,
		Email:    "querytoken@test.com",
		Role:     domain.RoleSecOps,
	}

	tokens, _ := tokenMgr.GenerateTokenPair(user)

	req := httptest.NewRequest(http.MethodGet, "/test-protected?token="+tokens.AccessToken, nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Atteso 200 OK con query token, ottenuto %d. Body: %s", rec.Code, rec.Body.String())
	}

	var env response.Envelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	data := env.Data.(map[string]interface{})

	if data["tenant_id"] != tenantID.String() {
		t.Errorf("Tenant ID errato: %v", data["tenant_id"])
	}
	if data["user_id"] != userID.String() {
		t.Errorf("User ID errato: %v", data["user_id"])
	}
	if data["role"] != string(domain.RoleSecOps) {
		t.Errorf("Ruolo errato: %v", data["role"])
	}
}


