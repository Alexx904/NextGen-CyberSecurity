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

// SSHServerRepo implementa domain.SSHServerRepository e domain.FirewallBackupRepository per PostgreSQL
type SSHServerRepo struct {
	pool *pgxpool.Pool
}

// NewSSHServerRepo crea una nuova istanza di SSHServerRepo collegata al pool PostgreSQL
func NewSSHServerRepo(pool *pgxpool.Pool) *SSHServerRepo {
	return &SSHServerRepo{pool: pool}
}

// ============================================================================
// 1. Gestione Server SSH Remoti (domain.SSHServerRepository)
// ============================================================================

// Create inserisce un nuovo server SSH garantendo l'isolamento multi-tenant
func (r *SSHServerRepo) Create(ctx context.Context, server *domain.SSHServer) error {
	if err := server.Validate(); err != nil {
		return err
	}

	if server.ID == uuid.Nil {
		server.ID = uuid.New()
	}
	now := time.Now().UTC()
	server.CreatedAt = now
	server.UpdatedAt = now

	query := `
		INSERT INTO ssh_servers (
			id, tenant_id, name, host, port, username, auth_type,
			encrypted_secret, public_key_fingerprint, status, last_connected_at,
			created_at, updated_at, client_name, host_key_fingerprint
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		server.ID,
		server.TenantID,
		server.Name,
		server.Host,
		server.Port,
		server.Username,
		string(server.AuthType),
		server.EncryptedSecret,
		server.PublicKeyFingerprint,
		string(server.Status),
		server.LastConnectedAt,
		server.CreatedAt,
		server.UpdatedAt,
		server.ClientName,
		server.HostKeyFingerprint,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrSSHServerDuplicateName
		}
		return fmt.Errorf("inserimento server SSH fallito: %w", err)
	}

	return nil
}

// GetByID recupera un server SSH filtrando rigorosamente per tenant_id (NFR-SEC-03)
func (r *SSHServerRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.SSHServer, error) {
	query := `
		SELECT
			id, tenant_id, name, host, port, username, auth_type,
			encrypted_secret, public_key_fingerprint, status, last_connected_at,
			created_at, updated_at, client_name, host_key_fingerprint
		FROM ssh_servers
		WHERE tenant_id = $1 AND id = $2
	`

	var s domain.SSHServer
	var authTypeStr string
	var statusStr string

	err := r.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&s.ID,
		&s.TenantID,
		&s.Name,
		&s.Host,
		&s.Port,
		&s.Username,
		&authTypeStr,
		&s.EncryptedSecret,
		&s.PublicKeyFingerprint,
		&statusStr,
		&s.LastConnectedAt,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.ClientName,
		&s.HostKeyFingerprint,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSSHServerNotFound
		}
		return nil, fmt.Errorf("recupero server SSH per ID fallito: %w", err)
	}

	s.AuthType = domain.SSHAuthType(authTypeStr)
	s.Status = domain.SSHServerStatus(statusStr)

	return &s, nil
}

// List recupera tutti i server SSH registrati per un determinato tenant
func (r *SSHServerRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.SSHServer, error) {
	query := `
		SELECT
			id, tenant_id, name, host, port, username, auth_type,
			encrypted_secret, public_key_fingerprint, status, last_connected_at,
			created_at, updated_at, client_name, host_key_fingerprint
		FROM ssh_servers
		WHERE tenant_id = $1
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query lista server SSH fallita: %w", err)
	}
	defer rows.Close()

	servers := make([]*domain.SSHServer, 0)
	for rows.Next() {
		var s domain.SSHServer
		var authTypeStr string
		var statusStr string

		err := rows.Scan(
			&s.ID,
			&s.TenantID,
			&s.Name,
			&s.Host,
			&s.Port,
			&s.Username,
			&authTypeStr,
			&s.EncryptedSecret,
			&s.PublicKeyFingerprint,
			&statusStr,
			&s.LastConnectedAt,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.ClientName,
			&s.HostKeyFingerprint,
		)
		if err != nil {
			return nil, fmt.Errorf("scan record server SSH fallito: %w", err)
		}

		s.AuthType = domain.SSHAuthType(authTypeStr)
		s.Status = domain.SSHServerStatus(statusStr)
		servers = append(servers, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterazione righe server SSH fallita: %w", err)
	}

	return servers, nil
}

// Update aggiorna le impostazioni del server SSH preservando l'isolamento tenant
func (r *SSHServerRepo) Update(ctx context.Context, server *domain.SSHServer) error {
	if err := server.Validate(); err != nil {
		return err
	}

	server.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE ssh_servers
		SET
			name = $3,
			host = $4,
			port = $5,
			username = $6,
			auth_type = $7,
			encrypted_secret = $8,
			public_key_fingerprint = $9,
			status = $10,
			updated_at = $11,
			client_name = $12,
			host_key_fingerprint = $13,
			last_connected_at = $14
		WHERE tenant_id = $1 AND id = $2
	`

	cmdTag, err := r.pool.Exec(
		ctx,
		query,
		server.TenantID,
		server.ID,
		server.Name,
		server.Host,
		server.Port,
		server.Username,
		string(server.AuthType),
		server.EncryptedSecret,
		server.PublicKeyFingerprint,
		string(server.Status),
		server.UpdatedAt,
		server.ClientName,
		server.HostKeyFingerprint,
		server.LastConnectedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrSSHServerDuplicateName
		}
		return fmt.Errorf("aggiornamento server SSH fallito: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrSSHServerNotFound
	}

	return nil
}

// UpdateStatus aggiorna rapidamente stato di connessione e timestamp di ultima connessione
func (r *SSHServerRepo) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status domain.SSHServerStatus, lastConnectedAt *time.Time) error {
	if !status.IsValid() {
		return domain.ErrValidation
	}

	query := `
		UPDATE ssh_servers
		SET
			status = $3,
			last_connected_at = $4,
			updated_at = $5
		WHERE tenant_id = $1 AND id = $2
	`

	now := time.Now().UTC()
	cmdTag, err := r.pool.Exec(ctx, query, tenantID, id, string(status), lastConnectedAt, now)
	if err != nil {
		return fmt.Errorf("aggiornamento status server SSH fallito: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrSSHServerNotFound
	}

	return nil
}

// Delete rimuove un server SSH e a cascata i relativi snapshot firewall
func (r *SSHServerRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `
		DELETE FROM ssh_servers
		WHERE tenant_id = $1 AND id = $2
	`

	cmdTag, err := r.pool.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("eliminazione server SSH fallita: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrSSHServerNotFound
	}

	return nil
}

// ============================================================================
// 2. Gestione Snapshot Firewall (domain.FirewallBackupRepository)
// ============================================================================

// CreateBackup inserisce un nuovo snapshot di regole iptables per un server
func (r *SSHServerRepo) CreateBackup(ctx context.Context, backup *domain.FirewallBackup) error {
	if err := backup.Validate(); err != nil {
		return err
	}

	if backup.ID == uuid.Nil {
		backup.ID = uuid.New()
	}
	if backup.CreatedAt.IsZero() {
		backup.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO firewall_backups (
			id, tenant_id, server_id, raw_iptables_rules, label, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		backup.ID,
		backup.TenantID,
		backup.ServerID,
		backup.RawIptablesRules,
		backup.Label,
		backup.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("salvataggio snapshot firewall fallito: %w", err)
	}

	return nil
}

// GetBackupByID recupera uno snapshot firewall specifico verificando il tenant
func (r *SSHServerRepo) GetBackupByID(ctx context.Context, tenantID, backupID uuid.UUID) (*domain.FirewallBackup, error) {
	query := `
		SELECT id, tenant_id, server_id, raw_iptables_rules, label, created_at
		FROM firewall_backups
		WHERE tenant_id = $1 AND id = $2
	`

	var b domain.FirewallBackup
	err := r.pool.QueryRow(ctx, query, tenantID, backupID).Scan(
		&b.ID,
		&b.TenantID,
		&b.ServerID,
		&b.RawIptablesRules,
		&b.Label,
		&b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFirewallBackupNotFound
		}
		return nil, fmt.Errorf("recupero backup firewall fallito: %w", err)
	}

	return &b, nil
}

// GetLatestBackup recupera l'ultimo snapshot valido salvato per un determinato server SSH
func (r *SSHServerRepo) GetLatestBackup(ctx context.Context, tenantID, serverID uuid.UUID) (*domain.FirewallBackup, error) {
	query := `
		SELECT id, tenant_id, server_id, raw_iptables_rules, label, created_at
		FROM firewall_backups
		WHERE tenant_id = $1 AND server_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var b domain.FirewallBackup
	err := r.pool.QueryRow(ctx, query, tenantID, serverID).Scan(
		&b.ID,
		&b.TenantID,
		&b.ServerID,
		&b.RawIptablesRules,
		&b.Label,
		&b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFirewallBackupNotFound
		}
		return nil, fmt.Errorf("recupero ultimo backup firewall fallito: %w", err)
	}

	return &b, nil
}

// ListBackups recupera la cronologia ordinata degli snapshot firewall per un dato server
func (r *SSHServerRepo) ListBackups(ctx context.Context, tenantID, serverID uuid.UUID) ([]*domain.FirewallBackup, error) {
	query := `
		SELECT id, tenant_id, server_id, raw_iptables_rules, label, created_at
		FROM firewall_backups
		WHERE tenant_id = $1 AND server_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, tenantID, serverID)
	if err != nil {
		return nil, fmt.Errorf("query cronologia backup firewall fallita: %w", err)
	}
	defer rows.Close()

	backups := make([]*domain.FirewallBackup, 0)
	for rows.Next() {
		var b domain.FirewallBackup
		err := rows.Scan(
			&b.ID,
			&b.TenantID,
			&b.ServerID,
			&b.RawIptablesRules,
			&b.Label,
			&b.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan riga backup firewall fallito: %w", err)
		}
		backups = append(backups, &b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterazione righe backup firewall fallita: %w", err)
	}

	return backups, nil
}

// DeleteBackup elimina un singolo snapshot firewall
func (r *SSHServerRepo) DeleteBackup(ctx context.Context, tenantID, backupID uuid.UUID) error {
	query := `
		DELETE FROM firewall_backups
		WHERE tenant_id = $1 AND id = $2
	`

	cmdTag, err := r.pool.Exec(ctx, query, tenantID, backupID)
	if err != nil {
		return fmt.Errorf("eliminazione backup firewall fallita: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrFirewallBackupNotFound
	}

	return nil
}
