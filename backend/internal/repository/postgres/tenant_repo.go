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

// TenantRepo implementa domain.TenantRepository per PostgreSQL
type TenantRepo struct {
	pool *pgxpool.Pool
}

// NewTenantRepo crea una nuova istanza di TenantRepo
func NewTenantRepo(pool *pgxpool.Pool) *TenantRepo {
	return &TenantRepo{pool: pool}
}

// GetByID recupera un'organizzazione tenant tramite il suo ID univoco
func (r *TenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	query := `
		SELECT id, name, slug, logo_url, brand_colors, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`

	var t domain.Tenant
	var colorsRaw []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.Name,
		&t.Slug,
		&t.LogoURL,
		&colorsRaw,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("query tenant per ID fallita: %w", err)
	}

	if len(colorsRaw) > 0 {
		if err := json.Unmarshal(colorsRaw, &t.BrandColors); err != nil {
			t.BrandColors = domain.DefaultBrandColors()
		}
	} else {
		t.BrandColors = domain.DefaultBrandColors()
	}

	return &t, nil
}

// GetBySlug recupera un'organizzazione tenant tramite il suo slug univoco
func (r *TenantRepo) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	query := `
		SELECT id, name, slug, logo_url, brand_colors, created_at, updated_at
		FROM tenants
		WHERE slug = $1
	`

	var t domain.Tenant
	var colorsRaw []byte

	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&t.ID,
		&t.Name,
		&t.Slug,
		&t.LogoURL,
		&colorsRaw,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("query tenant per slug fallita: %w", err)
	}

	if len(colorsRaw) > 0 {
		if err := json.Unmarshal(colorsRaw, &t.BrandColors); err != nil {
			t.BrandColors = domain.DefaultBrandColors()
		}
	} else {
		t.BrandColors = domain.DefaultBrandColors()
	}

	return &t, nil
}

// Create registra un nuovo tenant nel database
func (r *TenantRepo) Create(ctx context.Context, tenant *domain.Tenant) error {
	if tenant == nil {
		return errors.New("tenant nullo non consentito")
	}

	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	now := time.Now().UTC()
	if tenant.CreatedAt.IsZero() {
		tenant.CreatedAt = now
	}
	if tenant.UpdatedAt.IsZero() {
		tenant.UpdatedAt = now
	}

	colorsJSON, err := json.Marshal(tenant.BrandColors)
	if err != nil {
		return fmt.Errorf("serializzazione brand colors fallita: %w", err)
	}

	query := `
		INSERT INTO tenants (id, name, slug, logo_url, brand_colors, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.pool.Exec(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.LogoURL,
		colorsJSON,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("inserimento tenant fallito: %w", err)
	}

	return nil
}

// UpdateBranding aggiorna la ragione sociale, il logo e la palette di colori del tenant
func (r *TenantRepo) UpdateBranding(ctx context.Context, id uuid.UUID, name string, logoURL *string, colors domain.BrandColors) (*domain.Tenant, error) {
	colorsJSON, err := json.Marshal(colors)
	if err != nil {
		return nil, fmt.Errorf("serializzazione brand colors fallita: %w", err)
	}

	updatedAt := time.Now().UTC()

	query := `
		UPDATE tenants
		SET name = $1, logo_url = $2, brand_colors = $3, updated_at = $4
		WHERE id = $5
		RETURNING id, name, slug, logo_url, brand_colors, created_at, updated_at
	`

	var t domain.Tenant
	var colorsRaw []byte

	err = r.pool.QueryRow(ctx, query,
		name,
		logoURL,
		colorsJSON,
		updatedAt,
		id,
	).Scan(
		&t.ID,
		&t.Name,
		&t.Slug,
		&t.LogoURL,
		&colorsRaw,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("aggiornamento branding tenant fallito: %w", err)
	}

	if len(colorsRaw) > 0 {
		if err := json.Unmarshal(colorsRaw, &t.BrandColors); err != nil {
			t.BrandColors = colors
		}
	} else {
		t.BrandColors = colors
	}

	return &t, nil
}

// List restituisce tutti i tenant configurati
func (r *TenantRepo) List(ctx context.Context) ([]*domain.Tenant, error) {
	query := `
		SELECT id, name, slug, logo_url, brand_colors, created_at, updated_at
		FROM tenants
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query lista tenant fallita: %w", err)
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		var t domain.Tenant
		var colorsRaw []byte
		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.Slug,
			&t.LogoURL,
			&colorsRaw,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scansione tenant fallita: %w", err)
		}

		if len(colorsRaw) > 0 {
			if err := json.Unmarshal(colorsRaw, &t.BrandColors); err != nil {
				t.BrandColors = domain.DefaultBrandColors()
			}
		} else {
			t.BrandColors = domain.DefaultBrandColors()
		}

		tenants = append(tenants, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterazione tenant fallita: %w", err)
	}

	return tenants, nil
}

