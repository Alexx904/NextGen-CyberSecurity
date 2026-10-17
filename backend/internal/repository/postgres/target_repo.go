package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/validator"
)

// TargetRepo implementa domain.TargetRepository per PostgreSQL
type TargetRepo struct {
	pool *pgxpool.Pool
}

// NewTargetRepo crea una nuova istanza di TargetRepo
func NewTargetRepo(pool *pgxpool.Pool) *TargetRepo {
	return &TargetRepo{pool: pool}
}

// GetByID recupera un target verificando rigorosamente l'appartenenza al tenant
func (r *TargetRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Target, error) {
	query := `
		SELECT id, tenant_id, group_id, target_type, value, description, is_blacklisted, created_at, updated_at
		FROM targets
		WHERE tenant_id = $1 AND id = $2
	`

	var t domain.Target
	var typeStr string

	err := r.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&t.ID,
		&t.TenantID,
		&t.GroupID,
		&typeStr,
		&t.Value,
		&t.Description,
		&t.IsBlacklisted,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("query target per ID fallita: %w", err)
	}

	t.Type = domain.TargetType(typeStr)
	return &t, nil
}

// List recupera l'elenco paginato dei target con filtri di ricerca per il tenant
func (r *TargetRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.TargetFilter) ([]*domain.Target, int, error) {
	whereClauses := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.GroupID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("group_id = $%d", argIdx))
		args = append(args, *filter.GroupID)
		argIdx++
	}

	if filter.Type != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("target_type = $%d", argIdx))
		args = append(args, string(*filter.Type))
		argIdx++
	}

	if filter.IsBlacklisted != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("is_blacklisted = $%d", argIdx))
		args = append(args, *filter.IsBlacklisted)
		argIdx++
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(value ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		args = append(args, searchPattern)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// 1. Calcolo del totale record
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM targets WHERE %s", whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("conteggio target fallito: %w", err)
	}

	// 2. Query paginata
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	listQuery := fmt.Sprintf(`
		SELECT id, tenant_id, group_id, target_type, value, description, is_blacklisted, created_at, updated_at
		FROM targets
		WHERE %s
		ORDER BY created_at DESC, id ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	queryArgs := append(args, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query elenco target fallita: %w", err)
	}
	defer rows.Close()

	var targets []*domain.Target
	for rows.Next() {
		var t domain.Target
		var typeStr string
		if err := rows.Scan(
			&t.ID,
			&t.TenantID,
			&t.GroupID,
			&typeStr,
			&t.Value,
			&t.Description,
			&t.IsBlacklisted,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scansione riga target fallita: %w", err)
		}
		t.Type = domain.TargetType(typeStr)
		targets = append(targets, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterazione target fallita: %w", err)
	}

	return targets, total, nil
}

// Create inserisce un nuovo target nel database
func (r *TargetRepo) Create(ctx context.Context, target *domain.Target) error {
	if target == nil {
		return errors.New("target nullo non consentito per il salvataggio")
	}

	if target.ID == uuid.Nil {
		target.ID = uuid.New()
	}
	now := time.Now().UTC()
	if target.CreatedAt.IsZero() {
		target.CreatedAt = now
	}
	if target.UpdatedAt.IsZero() {
		target.UpdatedAt = now
	}

	query := `
		INSERT INTO targets (id, tenant_id, group_id, target_type, value, description, is_blacklisted, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		target.ID,
		target.TenantID,
		target.GroupID,
		string(target.Type),
		target.Value,
		target.Description,
		target.IsBlacklisted,
		target.CreatedAt,
		target.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("inserimento target fallito: %w", err)
	}

	return nil
}

// Update aggiorna le informazioni del target
func (r *TargetRepo) Update(ctx context.Context, target *domain.Target) error {
	if target == nil {
		return errors.New("target nullo non consentito per l'aggiornamento")
	}

	target.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE targets
		SET group_id = $1, description = $2, is_blacklisted = $3, updated_at = $4
		WHERE tenant_id = $5 AND id = $6
	`

	cmdTag, err := r.pool.Exec(ctx, query,
		target.GroupID,
		target.Description,
		target.IsBlacklisted,
		target.UpdatedAt,
		target.TenantID,
		target.ID,
	)
	if err != nil {
		return fmt.Errorf("aggiornamento target fallito: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete rimuove un target verificando l'appartenenza al tenant
func (r *TargetRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	cmdTag, err := r.pool.Exec(ctx, "DELETE FROM targets WHERE tenant_id = $1 AND id = $2", tenantID, id)
	if err != nil {
		return fmt.Errorf("eliminazione target fallita: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// IsBlacklisted verifica se un valore collide con una delle blacklist registrate per il tenant
func (r *TargetRepo) IsBlacklisted(ctx context.Context, tenantID uuid.UUID, value string) (bool, error) {
	rows, err := r.pool.Query(ctx, "SELECT value FROM targets WHERE tenant_id = $1 AND is_blacklisted = true", tenantID)
	if err != nil {
		return false, fmt.Errorf("query blacklist tenant fallita: %w", err)
	}
	defer rows.Close()

	var blacklisted []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return false, fmt.Errorf("lettura entry blacklist fallita: %w", err)
		}
		blacklisted = append(blacklisted, v)
	}

	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterazione blacklist fallita: %w", err)
	}

	if len(blacklisted) == 0 {
		return false, nil
	}

	return validator.IsTargetBlacklisted(value, blacklisted)
}

