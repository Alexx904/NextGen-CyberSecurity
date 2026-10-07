package middleware

import (
	"context"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/auth"
)

type contextKey string

const (
	claimsKey   contextKey = "auth_claims"
	tenantIDKey contextKey = "auth_tenant_id"
	userIDKey   contextKey = "auth_user_id"
	userRoleKey contextKey = "auth_user_role"
)

// WithUserClaims inietta i claim dell'utente e i relativi campi nel contesto della richiesta
func WithUserClaims(ctx context.Context, claims *auth.UserClaims) context.Context {
	if claims == nil {
		return ctx
	}
	ctx = context.WithValue(ctx, claimsKey, claims)
	ctx = context.WithValue(ctx, tenantIDKey, claims.TenantID)
	ctx = context.WithValue(ctx, userIDKey, claims.UserID)
	ctx = context.WithValue(ctx, userRoleKey, claims.Role)
	return ctx
}

// GetUserClaims estrae i claim completi dal contesto
func GetUserClaims(ctx context.Context) (*auth.UserClaims, bool) {
	claims, ok := ctx.Value(claimsKey).(*auth.UserClaims)
	return claims, ok && claims != nil
}

// GetTenantID estrae l'identificativo tenant dal contesto garantendo il tipo uuid.UUID
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(tenantIDKey).(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

// GetUserID estrae l'ID dell'utente autenticato dal contesto
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

// GetUserRole estrae il ruolo utente RBAC dal contesto
func GetUserRole(ctx context.Context) (domain.UserRole, bool) {
	role, ok := ctx.Value(userRoleKey).(domain.UserRole)
	if !ok || !role.IsValid() {
		return "", false
	}
	return role, true
}

