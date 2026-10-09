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

// TargetGroupRepo implementa domain.TargetGroupRepository per PostgreSQL
type TargetGroupRepo struct {
	pool *pgxpool.Pool
}

// NewTargetGroupRepo crea una nuova istanza di TargetGroupRepo
func NewTargetGroupRepo(pool *pgxpool.Pool) *TargetGroupRepo {
	return &TargetGroupRepo{pool: pool}
}

// GetByID recupera un gruppo verificando l'appartenenza al tenant
func (r *TargetGroupRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.TargetGroup, error) {
	query := `
		SELECT id, tenant_id, parent_id, name, description, color_tag, created_at, updated_at
		FROM target_groups
		WHERE tenant_id = $1 AND id = $2
	`

	var g domain.TargetGroup
	err := r.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&g.ID,
		&g.TenantID,
		&g.ParentID,
		&g.Name,
		&g.Description,
		&g.ColorTag,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("query gruppo per ID fallita: %w", err)
	}

	return &g, nil
}

// List recupera tutti i gruppi appartenenti al tenant ordinati per nome
func (r *TargetGroupRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroup, error) {
	query := `
		SELECT id, tenant_id, parent_id, name, description, color_tag, created_at, updated_at
		FROM target_groups
		WHERE tenant_id = $1
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query lista gruppi fallita: %w", err)
	}
	defer rows.Close()

	var groups []*domain.TargetGroup
	for rows.Next() {
		var g domain.TargetGroup
		if err := rows.Scan(
			&g.ID,
			&g.TenantID,
			&g.ParentID,
			&g.Name,
			&g.Description,
			&g.ColorTag,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scansione gruppo fallita: %w", err)
		}
		groups = append(groups, &g)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterazione gruppi fallita: %w", err)
	}

	return groups, nil
}

// Create inserisce un nuovo gruppo di target (con eventuale parent_id e color_tag)
func (r *TargetGroupRepo) Create(ctx context.Context, group *domain.TargetGroup) error {
	if group == nil {
		return errors.New("gruppo nullo non consentito per il salvataggio")
	}

	if group.ID == uuid.Nil {
		group.ID = uuid.New()
	}
	now := time.Now().UTC()
	if group.CreatedAt.IsZero() {
		group.CreatedAt = now
	}
	if group.UpdatedAt.IsZero() {
		group.UpdatedAt = now
	}

	query := `
		INSERT INTO target_groups (id, tenant_id, parent_id, name, description, color_tag, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		group.ID,
		group.TenantID,
		group.ParentID,
		group.Name,
		group.Description,
		group.ColorTag,
		group.CreatedAt,
		group.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("inserimento gruppo fallito: %w", err)
	}

	return nil
}

// Update aggiorna le informazioni del gruppo (nome, descrizione, parent_id, color_tag)
func (r *TargetGroupRepo) Update(ctx context.Context, group *domain.TargetGroup) error {
	if group == nil {
		return errors.New("gruppo nullo non consentito per l'aggiornamento")
	}

	group.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE target_groups
		SET parent_id = $1, name = $2, description = $3, color_tag = $4, updated_at = $5
		WHERE tenant_id = $6 AND id = $7
	`

	cmdTag, err := r.pool.Exec(ctx, query,
		group.ParentID,
		group.Name,
		group.Description,
		group.ColorTag,
		group.UpdatedAt,
		group.TenantID,
		group.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("aggiornamento gruppo fallito: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete elimina un gruppo di target appartenente al tenant
func (r *TargetGroupRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	cmdTag, err := r.pool.Exec(ctx, "DELETE FROM target_groups WHERE tenant_id = $1 AND id = $2", tenantID, id)
	if err != nil {
		return fmt.Errorf("eliminazione gruppo fallita: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetTargetCountsByGroup calcola il numero di target assegnati direttamente a ciascun gruppo
func (r *TargetGroupRepo) GetTargetCountsByGroup(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]int, error) {
	query := `
		SELECT group_id, COUNT(*)
		FROM targets
		WHERE tenant_id = $1 AND group_id IS NOT NULL
		GROUP BY group_id
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query conteggio target per gruppo fallita: %w", err)
	}
	defer rows.Close()

	counts := make(map[uuid.UUID]int)
	for rows.Next() {
		var groupID uuid.UUID
		var count int
		if err := rows.Scan(&groupID, &count); err != nil {
			return nil, fmt.Errorf("scansione conteggio target per gruppo fallita: %w", err)
		}
		counts[groupID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterazione conteggio target per gruppo fallita: %w", err)
	}

	return counts, nil
}

// GetSeverityCountsByGroup raggruppa le vulnerabilità attive per gruppo e livello di severità
// estraendo le scansioni completate più recenti per ciascun target
func (r *TargetGroupRepo) GetSeverityCountsByGroup(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]map[domain.SeverityLevel]int, error) {
	query := `
		WITH latest_scans AS (
			SELECT DISTINCT ON (s.target_id) s.id AS scan_id, t.group_id
			FROM scans s
			JOIN targets t ON t.id = s.target_id
			WHERE s.tenant_id = $1 AND s.status = 'COMPLETED' AND t.group_id IS NOT NULL
			ORDER BY s.target_id, s.completed_at DESC NULLS LAST
		)
		SELECT ls.group_id, f.severity, COUNT(*)
		FROM latest_scans ls
		JOIN findings f ON f.scan_id = ls.scan_id AND f.tenant_id = $1
		GROUP BY ls.group_id, f.severity
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query severità per gruppo fallita: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]map[domain.SeverityLevel]int)
	for rows.Next() {
		var groupID uuid.UUID
		var sevStr string
		var count int
		if err := rows.Scan(&groupID, &sevStr, &count); err != nil {
			return nil, fmt.Errorf("scansione severità per gruppo fallita: %w", err)
		}

		if result[groupID] == nil {
			result[groupID] = make(map[domain.SeverityLevel]int)
		}
		result[groupID][domain.SeverityLevel(sevStr)] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterazione severità per gruppo fallita: %w", err)
	}

	return result, nil
}

