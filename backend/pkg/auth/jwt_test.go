package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/auth"
)

func TestJWTManager_Success(t *testing.T) {
	secret := "test-secret-super-safe-key-12345"
	manager, err := auth.NewJWTManager(secret, 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("Errore durante NewJWTManager: %v", err)
	}

	user := &domain.User{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Email:    "analyst@security.local",
		Role:     domain.RoleSecOps,
		FullName: "SecOps Analyst",
		IsActive: true,
	}

	tokens, err := manager.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("Errore durante GenerateTokenPair: %v", err)
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("Token generati vuoti: %+v", tokens)
	}

	// Validazione access token
	claims, err := manager.ValidateToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("Validazione access token fallita: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Atteso UserID %s, ottenuto %s", user.ID, claims.UserID)
	}
	if claims.TenantID != user.TenantID {
		t.Errorf("Atteso TenantID %s, ottenuto %s", user.TenantID, claims.TenantID)
	}
	if claims.Role != domain.RoleSecOps {
		t.Errorf("Atteso Role %s, ottenuto %s", domain.RoleSecOps, claims.Role)
	}
	if claims.Email != user.Email {
		t.Errorf("Atteso Email %s, ottenuto %s", user.Email, claims.Email)
	}
	if claims.TokenType != auth.TokenTypeAccess {
		t.Errorf("Atteso TokenType %s, ottenuto %s", auth.TokenTypeAccess, claims.TokenType)
	}

	// Validazione refresh token
	refreshClaims, err := manager.ValidateToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("Validazione refresh token fallita: %v", err)
	}
	if refreshClaims.TokenType != auth.TokenTypeRefresh {
		t.Errorf("Atteso TokenType %s, ottenuto %s", auth.TokenTypeRefresh, refreshClaims.TokenType)
	}
}

func TestJWTManager_EmptySecret(t *testing.T) {
	_, err := auth.NewJWTManager("", 15*time.Minute, 24*time.Hour)
	if !errors.Is(err, auth.ErrEmptySecret) {
		t.Fatalf("Atteso ErrEmptySecret, ottenuto: %v", err)
	}
}

func TestJWTManager_NilUser(t *testing.T) {
	manager, err := auth.NewJWTManager("some-secret-key", 0, 0)
	if err != nil {
		t.Fatalf("NewJWTManager fallito: %v", err)
	}

	_, err = manager.GenerateTokenPair(nil)
	if err == nil {
		t.Fatal("Atteso errore per utente nullo, ottenuto nil")
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	// Durata negativa per produrre un token immediatamente scaduto
	manager, _ := auth.NewJWTManager("test-secret-key", -1*time.Minute, 1*time.Hour)

	user := &domain.User{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Email:    "test@expired.com",
		Role:     domain.RoleAuditor,
	}

	// Creiamo manualmente un token scaduto per testare ValidateToken
	now := time.Now().Add(-2 * time.Hour)
	claims := auth.UserClaims{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Role:      user.Role,
		Email:     user.Email,
		TokenType: auth.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Minute)), // scaduto 1h59m fa
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret-key"))

	_, err := manager.ValidateToken(tokenString)
	if !errors.Is(err, auth.ErrExpiredToken) {
		t.Fatalf("Atteso ErrExpiredToken, ottenuto: %v", err)
	}
}

func TestJWTManager_InvalidSignature(t *testing.T) {
	manager1, _ := auth.NewJWTManager("secret-key-1", 15*time.Minute, 1*time.Hour)
	manager2, _ := auth.NewJWTManager("different-secret-key-2", 15*time.Minute, 1*time.Hour)

	user := &domain.User{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Email:    "victim@test.local",
		Role:     domain.RoleTenantAdmin,
	}

	tokens, err := manager1.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair fallito: %v", err)
	}

	// Validazione con il secondo gestore (chiave diversa)
	_, err = manager2.ValidateToken(tokens.AccessToken)
	if !errors.Is(err, auth.ErrInvalidSignature) {
		t.Fatalf("Atteso ErrInvalidSignature, ottenuto: %v", err)
	}
}

func TestJWTManager_MalformedToken(t *testing.T) {
	manager, _ := auth.NewJWTManager("test-secret-key", 15*time.Minute, 1*time.Hour)

	_, err := manager.ValidateToken("")
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("Atteso ErrInvalidToken per token vuoto, ottenuto: %v", err)
	}

	_, err = manager.ValidateToken("this.is.not.a.valid.jwt")
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("Atteso ErrInvalidToken per token malformato, ottenuto: %v", err)
	}
}

func TestJWTManager_AlgorithmNoneAttack(t *testing.T) {
	manager, _ := auth.NewJWTManager("test-secret-key", 15*time.Minute, 1*time.Hour)

	// Creiamo un token non firmato con SigningMethodNone
	claims := auth.UserClaims{
		UserID:   uuid.New(),
		TenantID: uuid.New(),
		Role:     domain.RoleSuperAdmin,
		Email:    "hacker@fake.local",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := manager.ValidateToken(tokenString)
	if err == nil {
		t.Fatal("Atteso fallimento validazione per algoritmo None, token invece accettato!")
	}
}

