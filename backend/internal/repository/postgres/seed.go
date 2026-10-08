package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nextgen-cybersec/pkg/auth"
)

// SeedConfig contiene i parametri configurabili per l'inizializzazione del tenant e dell'amministratore
type SeedConfig struct {
	AdminEmail    string
	AdminPassword string
	AdminFullName string
	TenantName    string
	TenantSlug    string
}

// DefaultSeedConfig restituisce la configurazione predefinita per l'avvio del sistema
func DefaultSeedConfig() SeedConfig {
	return SeedConfig{
		AdminEmail:    "admin@nextgen.local",
		AdminPassword: "AdminPass123!",
		AdminFullName: "Super Administrator",
		TenantName:    "Default Security Organization",
		TenantSlug:    "default-org",
	}
}

// SeedDefaultAdminAndTenant crea in modo idempotente il tenant principale e il primo utente SuperAdmin
func SeedDefaultAdminAndTenant(
	ctx context.Context,
	pool *pgxpool.Pool,
	hasher auth.PasswordHasher,
	cfg SeedConfig,
) error {
	if cfg.AdminEmail == "" || cfg.AdminPassword == "" || cfg.TenantSlug == "" {
		return errors.New("parametri di seed non validi o incompleti")
	}

	// 1. Verifica o creazione del Tenant principale
	var tenantID uuid.UUID
	err := pool.QueryRow(ctx, "SELECT id FROM tenants WHERE slug = $1", cfg.TenantSlug).Scan(&tenantID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Info("Inizializzazione tenant primario di default...", "name", cfg.TenantName, "slug", cfg.TenantSlug)
			insertTenantQuery := `
				INSERT INTO tenants (name, slug, brand_colors)
				VALUES ($1, $2, '{"primary": "#0f172a", "secondary": "#2563eb", "accent": "#38bdf8"}'::jsonb)
				RETURNING id
			`
			err = pool.QueryRow(ctx, insertTenantQuery, cfg.TenantName, cfg.TenantSlug).Scan(&tenantID)
			if err != nil {
				return fmt.Errorf("creazione tenant primario fallita: %w", err)
			}
			slog.Info("Tenant primario creato con successo", "tenant_id", tenantID)
		} else {
			return fmt.Errorf("verifica esistenza tenant fallita: %w", err)
		}
	}

	// 2. Verifica o creazione dell'utente SuperAdmin
	var existingUserID uuid.UUID
	err = pool.QueryRow(ctx, "SELECT id FROM users WHERE LOWER(email) = LOWER($1)", cfg.AdminEmail).Scan(&existingUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Info("Creazione utente SuperAdmin di default...", "email", cfg.AdminEmail)
			passwordHash, err := hasher.Hash(cfg.AdminPassword)
			if err != nil {
				return fmt.Errorf("hashing password per superadmin fallito: %w", err)
			}

			insertUserQuery := `
				INSERT INTO users (tenant_id, email, password_hash, full_name, role, is_active)
				VALUES ($1, $2, $3, $4, 'SuperAdmin', true)
				RETURNING id
			`
			var newUserID uuid.UUID
			err = pool.QueryRow(ctx, insertUserQuery, tenantID, cfg.AdminEmail, passwordHash, cfg.AdminFullName).Scan(&newUserID)
			if err != nil {
				return fmt.Errorf("creazione utente SuperAdmin fallita: %w", err)
			}
			slog.Info("Utente SuperAdmin creato con successo", "user_id", newUserID, "email", cfg.AdminEmail)
		} else {
			return fmt.Errorf("verifica esistenza utente superadmin fallita: %w", err)
		}
	} else {
		// Sincronizzazione password hash con la configurazione seed corrente
		passwordHash, err := hasher.Hash(cfg.AdminPassword)
		if err == nil {
			_, _ = pool.Exec(ctx, "UPDATE users SET password_hash = $1 WHERE id = $2", passwordHash, existingUserID)
			slog.Info("Credenziali utente SuperAdmin sincronizzate con la configurazione", "email", cfg.AdminEmail)
		}
	}

	return nil
}

