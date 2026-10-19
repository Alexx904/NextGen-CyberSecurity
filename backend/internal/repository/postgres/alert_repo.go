package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"nextgen-cybersec/internal/domain"
)

type AlertRepo struct{ pool *pgxpool.Pool }

func NewAlertRepo(pool *pgxpool.Pool) *AlertRepo { return &AlertRepo{pool: pool} }

const alertChannelColumns = `id,tenant_id,company_id,name,kind,enabled,events,created_at,updated_at,encrypted_config`

func scanAlertChannel(row pgx.Row) (*domain.AlertChannel, error) {
	var c domain.AlertChannel
	err := row.Scan(&c.ID, &c.TenantID, &c.CompanyID, &c.Name, &c.Kind, &c.Enabled, &c.Events, &c.CreatedAt, &c.UpdatedAt, &c.EncryptedConfig)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &c, err
}

func (r *AlertRepo) ListChannels(ctx context.Context, tenant uuid.UUID) ([]domain.AlertChannel, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+alertChannelColumns+` FROM alert_channels WHERE tenant_id=$1 AND deleted_at IS NULL ORDER BY created_at,id`, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	channels := []domain.AlertChannel{}
	for rows.Next() {
		c, err := scanAlertChannel(rows)
		if err != nil {
			return nil, err
		}
		channels = append(channels, *c)
	}
	return channels, rows.Err()
}
func (r *AlertRepo) GetChannel(ctx context.Context, tenant, id uuid.UUID) (*domain.AlertChannel, error) {
	return scanAlertChannel(r.pool.QueryRow(ctx, `SELECT `+alertChannelColumns+` FROM alert_channels WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`, tenant, id))
}
func (r *AlertRepo) SaveChannel(ctx context.Context, c *domain.AlertChannel, create bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var previous *domain.AlertChannel
	if !create {
		previous, err = scanAlertChannel(tx.QueryRow(ctx, `SELECT `+alertChannelColumns+` FROM alert_channels WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL FOR UPDATE`, c.TenantID, c.ID))
		if err != nil {
			return err
		}
		if !previous.UpdatedAt.Equal(c.UpdatedAt) {
			return domain.ErrConflict
		}
	}
	if c.CompanyID != nil {
		var id uuid.UUID
		if err = tx.QueryRow(ctx, `SELECT id FROM target_groups WHERE tenant_id=$1 AND id=$2 AND parent_id IS NULL FOR KEY SHARE`, c.TenantID, *c.CompanyID).Scan(&id); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrNotFound
			}
			return err
		}
	}
	if create {
		err = tx.QueryRow(ctx, `INSERT INTO alert_channels(id,tenant_id,company_id,name,kind,enabled,events,encrypted_config) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING created_at,updated_at`, c.ID, c.TenantID, c.CompanyID, c.Name, c.Kind, c.Enabled, c.Events, c.EncryptedConfig).Scan(&c.CreatedAt, &c.UpdatedAt)
	} else {
		err = tx.QueryRow(ctx, `UPDATE alert_channels SET company_id=$3,name=$4,kind=$5,enabled=$6,events=$7,encrypted_config=$8,updated_at=NOW() WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL RETURNING created_at,updated_at`, c.ID, c.TenantID, c.CompanyID, c.Name, c.Kind, c.Enabled, c.Events, c.EncryptedConfig).Scan(&c.CreatedAt, &c.UpdatedAt)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == "23503" || pgErr.Code == "23514") {
		return domain.ErrValidation
	}
	if err != nil {
		return err
	}
	if previous != nil && (!c.Enabled || previous.EncryptedConfig != c.EncryptedConfig || previous.Kind != c.Kind || !sameAlertCompany(previous.CompanyID, c.CompanyID) || !sameAlertEvents(previous.Events, c.Events)) {
		_, err = tx.Exec(ctx, `UPDATE alert_jobs SET state='cancelled',result_code='CHANNEL_CHANGED',claim_token=NULL,lease_until=NULL,updated_at=NOW() WHERE tenant_id=$1 AND channel_id=$2 AND state IN ('pending','sending') AND (lease_until IS NULL OR lease_until<=NOW())`, c.TenantID, c.ID)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func sameAlertCompany(a, b *uuid.UUID) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}
func sameAlertEvents(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func (r *AlertRepo) DeleteChannel(ctx context.Context, tenant, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `UPDATE alert_channels SET enabled=false,deleted_at=NOW(),updated_at=NOW(),company_id=NULL,encrypted_config='' WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`, tenant, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AlertRepo) ClaimJob(ctx context.Context) (*domain.AlertJob, error) {
	// Clean up exhausted leases and channels disabled/deleted between retries.
	_, err := r.pool.Exec(ctx, `UPDATE alert_jobs j SET state=CASE WHEN j.attempts>=6 THEN 'dead' ELSE 'cancelled' END,result_code=CASE WHEN j.attempts>=6 THEN 'ATTEMPTS_EXHAUSTED' ELSE 'CHANNEL_DISABLED' END,claim_token=NULL,lease_until=NULL,updated_at=NOW() FROM alert_channels c WHERE c.id=j.channel_id AND c.tenant_id=j.tenant_id AND j.state IN ('pending','sending') AND (j.lease_until IS NULL OR j.lease_until<=NOW()) AND (j.attempts>=6 OR NOT c.enabled OR c.deleted_at IS NOT NULL)`)
	if err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx, `WITH candidate AS (
SELECT j.id FROM alert_jobs j JOIN alert_channels c ON c.id=j.channel_id AND c.tenant_id=j.tenant_id
WHERE j.state IN ('pending','sending') AND j.attempts<6 AND j.next_attempt_at<=NOW() AND (j.lease_until IS NULL OR j.lease_until<=NOW()) AND c.enabled AND c.deleted_at IS NULL
ORDER BY j.next_attempt_at,j.id FOR UPDATE OF j SKIP LOCKED LIMIT 1
), claimed AS (
UPDATE alert_jobs j SET state='sending',attempts=attempts+1,claim_token=gen_random_uuid(),lease_until=NOW()+INTERVAL '30 seconds',updated_at=NOW() FROM candidate WHERE j.id=candidate.id RETURNING j.*
)
SELECT j.id,j.payload,j.attempts,j.claim_token,c.id,c.tenant_id,c.company_id,c.name,c.kind,c.enabled,c.events,c.created_at,c.updated_at,c.encrypted_config
FROM claimed j JOIN alert_channels c ON c.id=j.channel_id AND c.tenant_id=j.tenant_id`)
	var job domain.AlertJob
	var payload []byte
	c := &job.Channel
	err = row.Scan(&job.ID, &payload, &job.Attempts, &job.ClaimToken, &c.ID, &c.TenantID, &c.CompanyID, &c.Name, &c.Kind, &c.Enabled, &c.Events, &c.CreatedAt, &c.UpdatedAt, &c.EncryptedConfig)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(payload, &job.Event); err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *AlertRepo) CompleteJob(ctx context.Context, job *domain.AlertJob, code string, retry bool, delay time.Duration) error {
	state := "dead"
	if code == "ACCEPTED" {
		state = "sent"
	} else if retry && job.Attempts < 6 {
		state = "pending"
	}
	tag, err := r.pool.Exec(ctx, `UPDATE alert_jobs SET state=$4,result_code=$5,next_attempt_at=NOW()+make_interval(secs=>$6),claim_token=NULL,lease_until=NULL,updated_at=NOW() WHERE id=$1 AND tenant_id=$2 AND claim_token=$3 AND state='sending' AND lease_until>NOW()`, job.ID, job.Channel.TenantID, job.ClaimToken, state, code, delay.Seconds())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}

var _ domain.AlertRepository = (*AlertRepo)(nil)
