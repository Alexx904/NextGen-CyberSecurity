package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
)

var (
	ErrInvalidToken     = errors.New("token JWT non valido")
	ErrExpiredToken     = errors.New("token JWT scaduto")
	ErrInvalidSignature = errors.New("firma del token non valida")
	ErrEmptySecret      = errors.New("il secret crittografico per JWT non può essere vuoto")
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// UserClaims definisce il payload crittografico incorporato nei token JWT
type UserClaims struct {
	UserID    uuid.UUID       `json:"user_id"`
	TenantID  uuid.UUID       `json:"tenant_id"`
	Role      domain.UserRole `json:"role"`
	Email     string          `json:"email"`
	TokenType string          `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenPair rappresenta la coppia di token emessi per una sessione autenticata
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// TokenManager definisce il contratto per la gestione dei token JWT
type TokenManager interface {
	GenerateTokenPair(user *domain.User) (*TokenPair, error)
	ValidateToken(tokenStr string) (*UserClaims, error)
}

// JWTManager implementa TokenManager basato su firma HMAC-SHA256
type JWTManager struct {
	secret          []byte
	accessDuration  time.Duration
	refreshDuration time.Duration
}

// NewJWTManager inizializza il gestore JWT verificando la presenza del secret
func NewJWTManager(secret string, accessDuration, refreshDuration time.Duration) (*JWTManager, error) {
	if len(secret) == 0 {
		return nil, ErrEmptySecret
	}
	if accessDuration <= 0 {
		accessDuration = 15 * time.Minute
	}
	if refreshDuration <= 0 {
		refreshDuration = 7 * 24 * time.Hour
	}

	return &JWTManager{
		secret:          []byte(secret),
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}, nil
}

// GenerateTokenPair crea una nuova coppia di access e refresh token per l'utente
func (m *JWTManager) GenerateTokenPair(user *domain.User) (*TokenPair, error) {
	if user == nil {
		return nil, errors.New("utente nullo non consentito per la generazione del token")
	}

	now := time.Now().UTC()
	accessExpiresAt := now.Add(m.accessDuration)
	refreshExpiresAt := now.Add(m.refreshDuration)

	// Creazione Access Token
	accessClaims := UserClaims{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Role:      user.Role,
		Email:     user.Email,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "nextgen-cybersec",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString(m.secret)
	if err != nil {
		return nil, fmt.Errorf("impossibile firmare access token: %w", err)
	}

	// Creazione Refresh Token
	refreshClaims := UserClaims{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Role:      user.Role,
		Email:     user.Email,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "nextgen-cybersec",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshString, err := refreshToken.SignedString(m.secret)
	if err != nil {
		return nil, fmt.Errorf("impossibile firmare refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessString,
		RefreshToken: refreshString,
		ExpiresAt:    accessExpiresAt,
	}, nil
}

// ValidateToken valida e decodifica la firma e i claim del token JWT
func (m *JWTManager) ValidateToken(tokenStr string) (*UserClaims, error) {
	if len(tokenStr) == 0 {
		return nil, ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validazione rigorosa dell'algoritmo di firma (previene attacchi 'alg: none' o mismatch RSA/HMAC)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, ErrInvalidSignature
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

