package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/auth"
)

// LoginInput racchiude i parametri per la richiesta di autenticazione
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginOutput restituisce l'utente autenticato e i relativi token
type LoginOutput struct {
	User   *domain.User    `json:"user"`
	Tokens *auth.TokenPair `json:"tokens"`
}

// AuthService definisce le operazioni di business sull'autenticazione
type AuthService interface {
	Login(ctx context.Context, input LoginInput) (*LoginOutput, error)
	RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error)
	ValidateToken(ctx context.Context, token string) (*auth.UserClaims, error)
	GetCurrentUser(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error)
}

// authService implementa AuthService coordinando repository, hasher e token manager
type authService struct {
	userRepo     domain.UserRepository
	hasher       auth.PasswordHasher
	tokenManager auth.TokenManager
}

// NewAuthService crea una nuova istanza del servizio di autenticazione
func NewAuthService(
	userRepo domain.UserRepository,
	hasher auth.PasswordHasher,
	tokenManager auth.TokenManager,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		hasher:       hasher,
		tokenManager: tokenManager,
	}
}

// Login esegue la validazione delle credenziali e l'emissione dei token JWT
func (s *authService) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	password := input.Password

	if email == "" || password == "" {
		return nil, domain.ErrValidation
	}

	// Recupero utente dal database (senza divulgare se l'email esiste o meno)
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	// Verifica stato attivo dell'account
	if !user.IsActive {
		return nil, domain.ErrUnauthorized
	}

	// Verifica della password con l'hasher
	if err := s.hasher.Compare(user.PasswordHash, password); err != nil {
		return nil, domain.ErrUnauthorized
	}

	// Emissione della coppia di token JWT
	tokens, err := s.tokenManager.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("generazione token fallita: %w", err)
	}

	return &LoginOutput{
		User:   user,
		Tokens: tokens,
	}, nil
}

// ValidateToken decodifica e verifica un token JWT
func (s *authService) ValidateToken(ctx context.Context, token string) (*auth.UserClaims, error) {
	if strings.TrimSpace(token) == "" {
		return nil, domain.ErrUnauthorized
	}

	claims, err := s.tokenManager.ValidateToken(token)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return claims, nil
}

// RefreshToken valida un refresh token ed emette una nuova coppia di access e refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, domain.ErrUnauthorized
	}

	claims, err := s.tokenManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	if claims.TokenType != auth.TokenTypeRefresh {
		return nil, domain.ErrUnauthorized
	}

	user, err := s.userRepo.GetByID(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	if !user.IsActive {
		return nil, domain.ErrUnauthorized
	}

	newTokens, err := s.tokenManager.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("generazione token durante refresh fallita: %w", err)
	}

	return newTokens, nil
}

// GetCurrentUser restituisce il profilo utente verificando rigorosamente il tenant
func (s *authService) GetCurrentUser(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}


