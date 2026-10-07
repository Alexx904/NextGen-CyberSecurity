package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserRole definisce i ruoli autorizzativi del sistema (RBAC)
type UserRole string

const (
	RoleSuperAdmin  UserRole = "SuperAdmin"
	RoleTenantAdmin UserRole = "TenantAdmin"
	RoleSecOps      UserRole = "SecOps"
	RoleAuditor     UserRole = "Auditor"
)

// IsValid verifica se il ruolo appartiene a quelli ammessi
func (r UserRole) IsValid() bool {
	switch r {
	case RoleSuperAdmin, RoleTenantAdmin, RoleSecOps, RoleAuditor:
		return true
	default:
		return false
	}
}

// CanTriggerScans determina se il ruolo è autorizzato ad avviare o cancellare scansioni
func (r UserRole) CanTriggerScans() bool {
	return r == RoleSuperAdmin || r == RoleTenantAdmin || r == RoleSecOps
}

// CanManageTargets determina se il ruolo può creare o modificare target e blacklist
func (r UserRole) CanManageTargets() bool {
	return r == RoleSuperAdmin || r == RoleTenantAdmin || r == RoleSecOps
}

// CanManageUsers determina se il ruolo può gestire gli account utente
func (r UserRole) CanManageUsers() bool {
	return r == RoleSuperAdmin || r == RoleTenantAdmin
}

// CanManageBranding determina se il ruolo può modificare loghi e colori del tenant
func (r UserRole) CanManageBranding() bool {
	return r == RoleSuperAdmin || r == RoleTenantAdmin
}

// IsReadOnly verifica se il ruolo ha accesso in sola visualizzazione
func (r UserRole) IsReadOnly() bool {
	return r == RoleAuditor
}

// User rappresenta un utente all'interno della piattaforma NextGen
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	TenantID     uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FullName     string    `json:"full_name" db:"full_name"`
	Role         UserRole  `json:"role" db:"role"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserRepository definisce i metodi di accesso ai dati per gli utenti
type UserRepository interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*User, error)
}

