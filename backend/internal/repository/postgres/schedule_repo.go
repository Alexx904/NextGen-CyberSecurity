package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nextgen-cybersec/internal/domain"
)

// ScheduleRepo implementa domain.ScheduleRepository per PostgreSQL
type ScheduleRepo struct {
	pool *pgxpool.Pool
}

// NewScheduleRepo crea una nuova istanza di ScheduleRepo
func NewScheduleRepo(pool *pgxpool.Pool) *ScheduleRepo {
	return &ScheduleRepo{pool: pool}
}

// GetByID recupera una singola pianificazione verificando l'isolamento tenant
func (r *ScheduleRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Schedule, error) {
	query := `
		SELECT id, tenant_id, target_id, profile_id, cron_expression,
		       is_active, next_run_at, last_run_at, created_at, updated_at, consent_snapshot
		FROM schedules
		WHERE tenant_id = $1 AND id = $2
	`

	var s domain.Schedule
	err := r.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&s.ID,
		&s.TenantID,
		&s.TargetID,
		&s.ProfileID,
		&s.CronExpression,
		&s.IsActive,
		&s.NextRunAt,
		&s.LastRunAt,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.Authorization,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("recupero pianificazione fallito: %w", err)
	}

	return &s, nil
}

// ListByTenant restituisce tutte le pianificazioni associate a un'organizzazione tenant
func (r *ScheduleRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Schedule, error) {
	query := `
		SELECT id, tenant_id, target_id, profile_id, cron_expression,
		       is_active, next_run_at, last_run_at, created_at, updated_at, consent_snapshot
		FROM schedules
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query pianificazioni tenant fallita: %w", err)
	}
	defer rows.Close()

	return r.scanSchedules(rows)
}

// ListActive carica tutte le pianificazioni attive a livello globale di sistema (usato all'avvio dello scheduler)
func (r *ScheduleRepo) ListActive(ctx context.Context) ([]*domain.Schedule, error) {
	query := `
		SELECT id, tenant_id, target_id, profile_id, cron_expression,
		       is_active, next_run_at, last_run_at, created_at, updated_at, consent_snapshot
		FROM schedules
		WHERE is_active = true
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query pianificazioni attive fallita: %w", err)
	}
	defer rows.Close()

	return r.scanSchedules(rows)
}

// Create inserisce una nuova pianificazione
func (r *ScheduleRepo) Create(ctx context.Context, schedule *domain.Schedule) error {
	if schedule.ID == uuid.Nil {
		schedule.ID = uuid.New()
	}
	now := time.Now().UTC()
	if schedule.CreatedAt.IsZero() {
		schedule.CreatedAt = now
	}
	if schedule.UpdatedAt.IsZero() {
		schedule.UpdatedAt = now
	}

	query := `
		INSERT INTO schedules (
			id, tenant_id, target_id, profile_id, cron_expression,
			is_active, next_run_at, last_run_at, created_at, updated_at, consent_snapshot
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		schedule.ID,
		schedule.TenantID,
		schedule.TargetID,
		schedule.ProfileID,
		schedule.CronExpression,
		schedule.IsActive,
		schedule.NextRunAt,
		schedule.LastRunAt,
		schedule.CreatedAt,
		schedule.UpdatedAt,
		schedule.Authorization,
	)
	if err != nil {
		return fmt.Errorf("creazione pianificazione fallita: %w", err)
	}

	return nil
}

// Update aggiorna una pianificazione esistente
func (r *ScheduleRepo) Update(ctx context.Context, schedule *domain.Schedule) error {
	schedule.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE schedules
		SET target_id = $1, profile_id = $2, cron_expression = $3,
		    is_active = $4, next_run_at = $5, last_run_at = $6, updated_at = $7
		WHERE tenant_id = $8 AND id = $9
	`

	tag, err := r.pool.Exec(
		ctx,
		query,
		schedule.TargetID,
		schedule.ProfileID,
		schedule.CronExpression,
		schedule.IsActive,
		schedule.NextRunAt,
		schedule.LastRunAt,
		schedule.UpdatedAt,
		schedule.TenantID,
		schedule.ID,
	)
	if err != nil {
		return fmt.Errorf("aggiornamento pianificazione fallito: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete elimina una pianificazione per tenant e ID
func (r *ScheduleRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM schedules WHERE tenant_id = $1 AND id = $2`

	tag, err := r.pool.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("eliminazione pianificazione fallita: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *ScheduleRepo) UpdateRunTimes(ctx context.Context, schedule *domain.Schedule) error {
	_, err := r.pool.Exec(ctx, `UPDATE schedules SET last_run_at=$1, next_run_at=$2
		WHERE tenant_id=$3 AND id=$4 AND is_active=true AND target_id=$5
		AND profile_id=$6 AND cron_expression=$7 AND consent_snapshot=$8`,
		schedule.LastRunAt, schedule.NextRunAt, schedule.TenantID, schedule.ID,
		schedule.TargetID, schedule.ProfileID, schedule.CronExpression, schedule.Authorization)
	return err
}

func (r *ScheduleRepo) scanSchedules(rows pgx.Rows) ([]*domain.Schedule, error) {
	var list []*domain.Schedule

	for rows.Next() {
		var s domain.Schedule
		if err := rows.Scan(
			&s.ID,
			&s.TenantID,
			&s.TargetID,
			&s.ProfileID,
			&s.CronExpression,
			&s.IsActive,
			&s.NextRunAt,
			&s.LastRunAt,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.Authorization,
		); err != nil {
			return nil, fmt.Errorf("scansione riga pianificazione fallita: %w", err)
		}
		list = append(list, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterazione righe pianificazioni fallita: %w", err)
	}

	return list, nil
}
