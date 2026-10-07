package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"nextgen-cybersec/internal/domain"
)

// UserRepo implementa domain.UserRepository per PostgreSQL con isolamento multi-tenant
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo istanzia il repository utenti
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// GetByID recupera un utente per tenant e ID (garantisce l'isolamento NFR-SEC-03)
func (r *UserRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE tenant_id = $1 AND id = $2
	`

	var user domain.User
	var roleStr string

	err := r.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&user.ID,
		&user.TenantID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&roleStr,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("query utente per ID fallita: %w", err)
	}

	user.Role = domain.UserRole(roleStr)
	return &user, nil
}

// GetByEmail recupera un utente univocamente per indirizzo email (usato durante il login)
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`

	var user domain.User
	var roleStr string

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.TenantID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&roleStr,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("query utente per email fallita: %w", err)
	}

	user.Role = domain.UserRole(roleStr)
	return &user, nil
}

// Create inserisce un nuovo utente nel database
func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	if user == nil {
		return errors.New("utente nullo non consentito per il salvataggio")
	}

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now().UTC()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	query := `
		INSERT INTO users (id, tenant_id, email, password_hash, full_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.TenantID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		string(user.Role),
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("inserimento utente fallito: %w", err)
	}

	return nil
}

// Update aggiorna le informazioni dell'utente rispettando il vincolo di tenant
func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	if user == nil {
		return errors.New("utente nullo non consentito per l'aggiornamento")
	}

	user.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE users
		SET full_name = $1, role = $2, is_active = $3, updated_at = $4
		WHERE tenant_id = $5 AND id = $6
	`

	cmdTag, err := r.pool.Exec(ctx, query,
		user.FullName,
		string(user.Role),
		user.IsActive,
		user.UpdatedAt,
		user.TenantID,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("aggiornamento utente fallito: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// ListByTenant restituisce la lista di utenti appartenenti al tenant
func (r *UserRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query lista utenti fallita: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		var roleStr string
		if err := rows.Scan(
			&u.ID,
			&u.TenantID,
			&u.Email,
			&u.PasswordHash,
			&u.FullName,
			&roleStr,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scansione riga utente fallita: %w", err)
		}
		u.Role = domain.UserRole(roleStr)
		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterazione righe utenti fallita: %w", err)
	}

	return users, nil
}

