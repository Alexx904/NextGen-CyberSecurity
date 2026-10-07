package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
)

// mockUserRepo implementa domain.UserRepository per i test unitari
type mockUserRepo struct {
	usersByID    map[string]*domain.User
	usersByEmail map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByID:    make(map[string]*domain.User),
		usersByEmail: make(map[string]*domain.User),
	}
}

func (m *mockUserRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.User, error) {
	key := tenantID.String() + ":" + id.String()
	user, ok := m.usersByID[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return user, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, ok := m.usersByEmail[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return user, nil
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if _, exists := m.usersByEmail[user.Email]; exists {
		return domain.ErrConflict
	}
	key := user.TenantID.String() + ":" + user.ID.String()
	m.usersByID[key] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	key := user.TenantID.String() + ":" + user.ID.String()
	if _, exists := m.usersByID[key]; !exists {
		return domain.ErrNotFound
	}
	m.usersByID[key] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockUserRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.User, error) {
	var list []*domain.User
	for _, u := range m.usersByID {
		if u.TenantID == tenantID {
			list = append(list, u)
		}
	}
	return list, nil
}

func setupAuthService(t *testing.T) (service.AuthService, *mockUserRepo, *domain.User, string) {
	repo := newMockUserRepo()
	hasher := auth.NewBcryptHasher(10)
	tokenManager, err := auth.NewJWTManager("super-secure-unit-test-secret-12345", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("Errore configurazione tokenManager: %v", err)
	}

	rawPassword := "SecurePassword#2026!"
	hash, err := hasher.Hash(rawPassword)
	if err != nil {
		t.Fatalf("Errore generazione hash: %v", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		TenantID:     uuid.New(),
		Email:        "admin@nextgen.local",
		PasswordHash: hash,
		FullName:     "System Admin",
		Role:         domain.RoleSuperAdmin,
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("Errore inserimento utente mock: %v", err)
	}

	authSvc := service.NewAuthService(repo, hasher, tokenManager)
	return authSvc, repo, user, rawPassword
}

func TestAuthService_Login_Success(t *testing.T) {
	authSvc, _, user, password := setupAuthService(t)
	ctx := context.Background()

	out, err := authSvc.Login(ctx, service.LoginInput{
		Email:    "ADMIN@nextgen.local ", // Test normalizzazione maiuscole e spazi
		Password: password,
	})
	if err != nil {
		t.Fatalf("Login fallito: %v", err)
	}

	if out.User.ID != user.ID {
		t.Errorf("Atteso UserID %s, ottenuto %s", user.ID, out.User.ID)
	}
	if out.Tokens.AccessToken == "" || out.Tokens.RefreshToken == "" {
		t.Errorf("Token non generati correttamente: %+v", out.Tokens)
	}
}

func TestAuthService_Login_ValidationErrors(t *testing.T) {
	authSvc, _, _, password := setupAuthService(t)
	ctx := context.Background()

	// Email vuota
	_, err := authSvc.Login(ctx, service.LoginInput{Email: "", Password: password})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Atteso ErrValidation per email vuota, ottenuto: %v", err)
	}

	// Password vuota
	_, err = authSvc.Login(ctx, service.LoginInput{Email: "admin@nextgen.local", Password: ""})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Atteso ErrValidation per password vuota, ottenuto: %v", err)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	authSvc, _, _, password := setupAuthService(t)
	ctx := context.Background()

	_, err := authSvc.Login(ctx, service.LoginInput{
		Email:    "unknown@user.com",
		Password: password,
	})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Atteso ErrUnauthorized per utente sconosciuto, ottenuto: %v", err)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	authSvc, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	_, err := authSvc.Login(ctx, service.LoginInput{
		Email:    "admin@nextgen.local",
		Password: "IncorrectPassword#123",
	})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Atteso ErrUnauthorized per password errata, ottenuto: %v", err)
	}
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	authSvc, repo, user, password := setupAuthService(t)
	ctx := context.Background()

	user.IsActive = false
	_ = repo.Update(ctx, user)

	_, err := authSvc.Login(ctx, service.LoginInput{
		Email:    "admin@nextgen.local",
		Password: password,
	})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Atteso ErrUnauthorized per utente disattivato, ottenuto: %v", err)
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	authSvc, _, _, password := setupAuthService(t)
	ctx := context.Background()

	out, err := authSvc.Login(ctx, service.LoginInput{
		Email:    "admin@nextgen.local",
		Password: password,
	})
	if err != nil {
		t.Fatalf("Login fallito: %v", err)
	}

	claims, err := authSvc.ValidateToken(ctx, out.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("Validazione token fallita: %v", err)
	}

	if claims.Email != "admin@nextgen.local" {
		t.Errorf("Attesa email admin@nextgen.local, ottenuta %s", claims.Email)
	}

	// Token vuoto
	_, err = authSvc.ValidateToken(ctx, "")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Atteso ErrUnauthorized per token vuoto, ottenuto: %v", err)
	}

	// Token malformato
	_, err = authSvc.ValidateToken(ctx, "malformed-token")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Atteso ErrUnauthorized per token malformato, ottenuto: %v", err)
	}
}

func TestAuthService_GetCurrentUser(t *testing.T) {
	authSvc, _, user, _ := setupAuthService(t)
	ctx := context.Background()

	found, err := authSvc.GetCurrentUser(ctx, user.TenantID, user.ID)
	if err != nil {
		t.Fatalf("GetCurrentUser fallito: %v", err)
	}

	if found.Email != user.Email {
		t.Errorf("Atteso utente con email %s, ottenuto %s", user.Email, found.Email)
	}

	// Tenant errato (Cross-Tenant prevention)
	differentTenantID := uuid.New()
	_, err = authSvc.GetCurrentUser(ctx, differentTenantID, user.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Atteso ErrNotFound per tenant disallineato, ottenuto: %v", err)
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	authSvc, repo, user, password := setupAuthService(t)
	ctx := context.Background()

	out, err := authSvc.Login(ctx, service.LoginInput{
		Email:    user.Email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Login fallito: %v", err)
	}

	// 1. Successo con refresh token valido
	newTokens, err := authSvc.RefreshToken(ctx, out.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshToken fallito: %v", err)
	}
	if newTokens.AccessToken == "" || newTokens.RefreshToken == "" {
		t.Errorf("Nuovi token non generati: %+v", newTokens)
	}

	// 2. Token vuoto
	_, err = authSvc.RefreshToken(ctx, "")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Atteso ErrUnauthorized per token vuoto, ottenuto: %v", err)
	}

	// 3. Passaggio di un Access Token invece di un Refresh Token
	_, err = authSvc.RefreshToken(ctx, out.Tokens.AccessToken)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Atteso ErrUnauthorized per token di tipo non refresh, ottenuto: %v", err)
	}

	// 4. Token malformato
	_, err = authSvc.RefreshToken(ctx, "invalid-token-string")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Atteso ErrUnauthorized per token malformato, ottenuto: %v", err)
	}

	// 5. Utente disattivato nel frattempo
	user.IsActive = false
	_ = repo.Update(ctx, user)

	_, err = authSvc.RefreshToken(ctx, out.Tokens.RefreshToken)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Atteso ErrUnauthorized per utente disattivato, ottenuto: %v", err)
	}
}


