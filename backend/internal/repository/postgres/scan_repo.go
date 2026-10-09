package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nextgen-cybersec/internal/domain"
)

// ScanRepo implementa domain.ScanRepository per PostgreSQL
type ScanRepo struct {
	pool *pgxpool.Pool
}

// NewScanRepo istanzia un nuovo ScanRepo
func NewScanRepo(pool *pgxpool.Pool) *ScanRepo {
	return &ScanRepo{pool: pool}
}

// GetByID recupera una scansione verificando rigorosamente l'appartenenza al tenant
func (r *ScanRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Scan, error) {
	query := `
		SELECT id, tenant_id, target_id, profile_id, trigger_type, status,
		       started_at, completed_at, summary_stats, error_message, created_at, updated_at
		FROM scans
		WHERE tenant_id = $1 AND id = $2
	`

	var s domain.Scan
	var triggerStr, statusStr string
	var statsRaw []byte

	err := r.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&s.ID,
		&s.TenantID,
		&s.TargetID,
		&s.ProfileID,
		&triggerStr,
		&statusStr,
		&s.StartedAt,
		&s.CompletedAt,
		&statsRaw,
		&s.ErrorMessage,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("recupero scansione per ID fallito: %w", err)
	}

	s.TriggerType = domain.ScanTriggerType(triggerStr)
	s.Status = domain.ScanStatus(statusStr)
	if len(statsRaw) > 0 {
		if err := json.Unmarshal(statsRaw, &s.SummaryStats); err != nil {
			return nil, fmt.Errorf("deserializzazione summary_stats fallita: %w", err)
		}
	}

	return &s, nil
}

// List recupera l'elenco paginato delle scansioni per un dato tenant con conteggio totale
func (r *ScanRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Scan, int, error) {
	countQuery := `SELECT COUNT(*) FROM scans WHERE tenant_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("conteggio scansioni fallito: %w", err)
	}

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, tenant_id, target_id, profile_id, trigger_type, status,
		       started_at, completed_at, summary_stats, error_message, created_at, updated_at
		FROM scans
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("recupero lista scansioni fallito: %w", err)
	}
	defer rows.Close()

	var scans []*domain.Scan
	for rows.Next() {
		var s domain.Scan
		var triggerStr, statusStr string
		var statsRaw []byte

		if err := rows.Scan(
			&s.ID,
			&s.TenantID,
			&s.TargetID,
			&s.ProfileID,
			&triggerStr,
			&statusStr,
			&s.StartedAt,
			&s.CompletedAt,
			&statsRaw,
			&s.ErrorMessage,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("lettura riga scansione fallita: %w", err)
		}

		s.TriggerType = domain.ScanTriggerType(triggerStr)
		s.Status = domain.ScanStatus(statusStr)
		if len(statsRaw) > 0 {
			_ = json.Unmarshal(statsRaw, &s.SummaryStats)
		}

		scans = append(scans, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterazione righe scansioni fallita: %w", err)
	}

	return scans, total, nil
}

// Create persiste una nuova scansione nel database
func (r *ScanRepo) Create(ctx context.Context, scan *domain.Scan) error {
	if scan.ID == uuid.Nil {
		scan.ID = uuid.New()
	}
	now := time.Now()
	if scan.CreatedAt.IsZero() {
		scan.CreatedAt = now
	}
	if scan.UpdatedAt.IsZero() {
		scan.UpdatedAt = now
	}
	if scan.Status == "" {
		scan.Status = domain.ScanStatusQueued
	}
	if scan.TriggerType == "" {
		scan.TriggerType = domain.TriggerManual
	}

	statsJSON, err := json.Marshal(scan.SummaryStats)
	if err != nil {
		return fmt.Errorf("serializzazione summary_stats fallita: %w", err)
	}

	query := `
		INSERT INTO scans (id, tenant_id, target_id, profile_id, trigger_type, status,
		                   started_at, completed_at, summary_stats, error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = r.pool.Exec(
		ctx,
		query,
		scan.ID,
		scan.TenantID,
		scan.TargetID,
		scan.ProfileID,
		string(scan.TriggerType),
		string(scan.Status),
		scan.StartedAt,
		scan.CompletedAt,
		statsJSON,
		scan.ErrorMessage,
		scan.CreatedAt,
		scan.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserimento nuova scansione fallito: %w", err)
	}

	return nil
}

// Update aggiorna atomicamente lo stato e le statistiche della scansione
func (r *ScanRepo) Update(ctx context.Context, scan *domain.Scan) error {
	scan.UpdatedAt = time.Now()

	statsJSON, err := json.Marshal(scan.SummaryStats)
	if err != nil {
		return fmt.Errorf("serializzazione summary_stats fallita: %w", err)
	}

	query := `
		UPDATE scans
		SET status = $1, started_at = $2, completed_at = $3, summary_stats = $4,
		    error_message = $5, updated_at = $6
		WHERE tenant_id = $7 AND id = $8
	`

	tag, err := r.pool.Exec(
		ctx,
		query,
		string(scan.Status),
		scan.StartedAt,
		scan.CompletedAt,
		statsJSON,
		scan.ErrorMessage,
		scan.UpdatedAt,
		scan.TenantID,
		scan.ID,
	)
	if err != nil {
		return fmt.Errorf("aggiornamento stato scansione fallito: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetActiveByTarget controlla se vi è già una scansione in stato non terminale sul target specificato
func (r *ScanRepo) GetActiveByTarget(ctx context.Context, tenantID, targetID uuid.UUID) (*domain.Scan, error) {
	query := `
		SELECT id, tenant_id, target_id, profile_id, trigger_type, status,
		       started_at, completed_at, summary_stats, error_message, created_at, updated_at
		FROM scans
		WHERE tenant_id = $1 AND target_id = $2 AND status IN ('QUEUED', 'RUNNING')
		LIMIT 1
	`

	var s domain.Scan
	var triggerStr, statusStr string
	var statsRaw []byte

	err := r.pool.QueryRow(ctx, query, tenantID, targetID).Scan(
		&s.ID,
		&s.TenantID,
		&s.TargetID,
		&s.ProfileID,
		&triggerStr,
		&statusStr,
		&s.StartedAt,
		&s.CompletedAt,
		&statsRaw,
		&s.ErrorMessage,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Nessuna scansione attiva
		}
		return nil, fmt.Errorf("verifica scansione attiva fallita: %w", err)
	}

	s.TriggerType = domain.ScanTriggerType(triggerStr)
	s.Status = domain.ScanStatus(statusStr)
	if len(statsRaw) > 0 {
		_ = json.Unmarshal(statsRaw, &s.SummaryStats)
	}

	return &s, nil
}

// ============================================================================
// SCAN PROFILE REPOSITORY IMPLEMENTATION
// ============================================================================

// ScanProfileRepo implementa domain.ScanProfileRepository per PostgreSQL
type ScanProfileRepo struct {
	pool *pgxpool.Pool
}

// NewScanProfileRepo crea una nuova istanza di ScanProfileRepo
func NewScanProfileRepo(pool *pgxpool.Pool) *ScanProfileRepo {
	return &ScanProfileRepo{pool: pool}
}

// GetByID recupera un profilo di scansione tramite UUID
func (r *ScanProfileRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ScanProfile, error) {
	query := `
		SELECT id, name, nmap_args, description, timeout_seconds, category, is_custom, tenant_id, created_at, updated_at
		FROM scan_profiles
		WHERE id = $1
	`

	var p domain.ScanProfile
	var catStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.Name,
		&p.NmapArgs,
		&p.Description,
		&p.TimeoutSeconds,
		&catStr,
		&p.IsCustom,
		&p.TenantID,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("recupero profilo per ID fallito: %w", err)
	}
	p.Category = domain.ScanProfileCategory(catStr)

	return &p, nil
}

// GetByName recupera un profilo di scansione tramite nome identificativo
func (r *ScanProfileRepo) GetByName(ctx context.Context, name string) (*domain.ScanProfile, error) {
	query := `
		SELECT id, name, nmap_args, description, timeout_seconds, category, is_custom, tenant_id, created_at, updated_at
		FROM scan_profiles
		WHERE name = $1
	`

	var p domain.ScanProfile
	var catStr string
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&p.ID,
		&p.Name,
		&p.NmapArgs,
		&p.Description,
		&p.TimeoutSeconds,
		&catStr,
		&p.IsCustom,
		&p.TenantID,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("recupero profilo per nome fallito: %w", err)
	}
	p.Category = domain.ScanProfileCategory(catStr)

	return &p, nil
}

// List recupera i profili di sistema globali e i profili custom del tenant specificato
func (r *ScanProfileRepo) List(ctx context.Context, tenantID *uuid.UUID) ([]*domain.ScanProfile, error) {
	var query string
	var rows pgx.Rows
	var err error

	if tenantID == nil {
		query = `
			SELECT id, name, nmap_args, description, timeout_seconds, category, is_custom, tenant_id, created_at, updated_at
			FROM scan_profiles
			WHERE tenant_id IS NULL
			ORDER BY category ASC, name ASC
		`
		rows, err = r.pool.Query(ctx, query)
	} else {
		query = `
			SELECT id, name, nmap_args, description, timeout_seconds, category, is_custom, tenant_id, created_at, updated_at
			FROM scan_profiles
			WHERE tenant_id IS NULL OR tenant_id = $1
			ORDER BY category ASC, name ASC
		`
		rows, err = r.pool.Query(ctx, query, *tenantID)
	}

	if err != nil {
		return nil, fmt.Errorf("recupero lista profili fallito: %w", err)
	}
	defer rows.Close()

	var profiles []*domain.ScanProfile
	for rows.Next() {
		var p domain.ScanProfile
		var catStr string
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.NmapArgs,
			&p.Description,
			&p.TimeoutSeconds,
			&catStr,
			&p.IsCustom,
			&p.TenantID,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("lettura riga profilo fallita: %w", err)
		}
		p.Category = domain.ScanProfileCategory(catStr)
		profiles = append(profiles, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterazione profili fallita: %w", err)
	}

	return profiles, nil
}

// ListByCategory filtra i profili per specifica categoria specialistica
func (r *ScanProfileRepo) ListByCategory(ctx context.Context, tenantID *uuid.UUID, category domain.ScanProfileCategory) ([]*domain.ScanProfile, error) {
	var query string
	var rows pgx.Rows
	var err error

	if tenantID == nil {
		query = `
			SELECT id, name, nmap_args, description, timeout_seconds, category, is_custom, tenant_id, created_at, updated_at
			FROM scan_profiles
			WHERE tenant_id IS NULL AND category = $1
			ORDER BY name ASC
		`
		rows, err = r.pool.Query(ctx, query, string(category))
	} else {
		query = `
			SELECT id, name, nmap_args, description, timeout_seconds, category, is_custom, tenant_id, created_at, updated_at
			FROM scan_profiles
			WHERE (tenant_id IS NULL OR tenant_id = $1) AND category = $2
			ORDER BY name ASC
		`
		rows, err = r.pool.Query(ctx, query, *tenantID, string(category))
	}

	if err != nil {
		return nil, fmt.Errorf("recupero profili per categoria fallito: %w", err)
	}
	defer rows.Close()

	var profiles []*domain.ScanProfile
	for rows.Next() {
		var p domain.ScanProfile
		var catStr string
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.NmapArgs,
			&p.Description,
			&p.TimeoutSeconds,
			&catStr,
			&p.IsCustom,
			&p.TenantID,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("lettura riga profilo per categoria fallita: %w", err)
		}
		p.Category = domain.ScanProfileCategory(catStr)
		profiles = append(profiles, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterazione profili fallita: %w", err)
	}

	return profiles, nil
}

// Create persiste un nuovo profilo di scansione (custom per tenant o di sistema)
func (r *ScanProfileRepo) Create(ctx context.Context, profile *domain.ScanProfile) error {
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	now := time.Now()
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = now
	}
	if profile.UpdatedAt.IsZero() {
		profile.UpdatedAt = now
	}
	if profile.Category == "" {
		profile.Category = domain.ProfileCategoryCustom
	}
	if profile.TimeoutSeconds <= 0 {
		profile.TimeoutSeconds = 3600
	}

	query := `
		INSERT INTO scan_profiles (id, name, nmap_args, description, timeout_seconds, category, is_custom, tenant_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		profile.ID,
		profile.Name,
		profile.NmapArgs,
		profile.Description,
		profile.TimeoutSeconds,
		string(profile.Category),
		profile.IsCustom,
		profile.TenantID,
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserimento nuovo profilo fallito: %w", err)
	}

	return nil
}

// Delete elimina un profilo custom verificando l'appartenenza al tenant (i profili di sistema non possono essere eliminati)
func (r *ScanProfileRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `
		DELETE FROM scan_profiles
		WHERE id = $1 AND tenant_id = $2 AND is_custom = true
	`

	tag, err := r.pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("eliminazione profilo custom fallita: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
