package postgres

import (
	"context"
	"encoding/json"
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

type DiscoveryRepo struct{ pool *pgxpool.Pool }

var _ domain.DiscoveryRepository = (*DiscoveryRepo)(nil)

func NewDiscoveryRepo(pool *pgxpool.Pool) *DiscoveryRepo { return &DiscoveryRepo{pool: pool} }

func (r *DiscoveryRepo) Create(ctx context.Context, result *domain.DiscoveryResult) error {
	if result.ID == uuid.Nil || result.TenantID == uuid.Nil {
		return domain.ErrValidation
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO discovery_sessions (id,tenant_id,subnet,started_at) VALUES ($1,$2,$3,$4)`, result.ID, result.TenantID, result.Subnet, result.StartedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return domain.ErrConflict
	}
	if err != nil {
		return fmt.Errorf("create discovery session: %w", err)
	}
	return nil
}

func (r *DiscoveryRepo) SaveDevice(ctx context.Context, tenantID, sessionID uuid.UUID, device domain.DiscoveredDevice) error {
	data, err := json.Marshal(device)
	if err != nil {
		return fmt.Errorf("encode discovery device: %w", err)
	}
	// Lock the session row: completion and appends cannot race across instances.
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var completed *time.Time
	err = tx.QueryRow(ctx, `SELECT completed_at FROM discovery_sessions WHERE tenant_id=$1 AND id=$2 FOR UPDATE`, tenantID, sessionID).Scan(&completed)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	if completed != nil {
		return domain.ErrConflict
	}
	_, err = tx.Exec(ctx, `INSERT INTO discovery_devices (tenant_id,session_id,ip,device) VALUES ($1,$2,$3,$4)
		ON CONFLICT (tenant_id,session_id,ip) DO UPDATE SET device=EXCLUDED.device`, tenantID, sessionID, device.IP, data)
	if err != nil {
		return fmt.Errorf("save discovery device: %w", err)
	}
	return tx.Commit(ctx)
}

func (r *DiscoveryRepo) Complete(ctx context.Context, result *domain.DiscoveryResult) error {
	if result.CompletedAt.IsZero() {
		return domain.ErrValidation
	}
	tag, err := r.pool.Exec(ctx, `UPDATE discovery_sessions SET completed_at=$3,duration_seconds=$4 WHERE tenant_id=$1 AND id=$2 AND completed_at IS NULL`, result.TenantID, result.ID, result.CompletedAt, result.DurationSeconds)
	if err != nil {
		return fmt.Errorf("complete discovery session: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}

func (r *DiscoveryRepo) GetLast(ctx context.Context, tenantID uuid.UUID) (*domain.DiscoveryResult, error) {
	return r.get(ctx, tenantID, nil)
}

func (r *DiscoveryRepo) GetSession(ctx context.Context, tenantID, sessionID uuid.UUID) (*domain.DiscoveryResult, error) {
	return r.get(ctx, tenantID, &sessionID)
}

func (r *DiscoveryRepo) get(ctx context.Context, tenantID uuid.UUID, sessionID *uuid.UUID) (*domain.DiscoveryResult, error) {
	var result domain.DiscoveryResult
	var completed *time.Time
	var data []byte
	// A single statement gives metadata and devices the same MVCC snapshot.
	err := r.pool.QueryRow(ctx, `SELECT s.id,s.tenant_id,s.subnet,s.started_at,s.completed_at,s.duration_seconds,
		COALESCE((SELECT jsonb_agg(d.device ORDER BY d.ip) FROM discovery_devices d WHERE d.tenant_id=s.tenant_id AND d.session_id=s.id),'[]'::jsonb)
		FROM discovery_sessions s WHERE s.tenant_id=$1 AND ($2::uuid IS NULL OR s.id=$2)
		ORDER BY s.started_at DESC,s.id DESC LIMIT 1`, tenantID, sessionID).Scan(&result.ID, &result.TenantID, &result.Subnet, &result.StartedAt, &completed, &result.DurationSeconds, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read discovery snapshot: %w", err)
	}
	if err := json.Unmarshal(data, &result.Devices); err != nil {
		return nil, err
	}
	result.TotalDiscovered = len(result.Devices)
	if completed != nil {
		result.CompletedAt = *completed
	}
	return &result, nil
}
