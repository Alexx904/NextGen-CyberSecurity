package domain

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SSHAuthType specifica la modalità di autenticazione per la connessione SSH
type SSHAuthType string

const (
	SSHAuthTypeKey      SSHAuthType = "KEY"
	SSHAuthTypePassword SSHAuthType = "PASSWORD"
)

// IsValid verifica che il tipo di autenticazione sia supportato
func (t SSHAuthType) IsValid() bool {
	switch t {
	case SSHAuthTypeKey, SSHAuthTypePassword:
		return true
	default:
		return false
	}
}

// SSHServerStatus rappresenta lo stato di connettività dell'host remoto
type SSHServerStatus string

const (
	SSHServerStatusOnline    SSHServerStatus = "ONLINE"
	SSHServerStatusOffline   SSHServerStatus = "OFFLINE"
	SSHServerStatusUnchecked SSHServerStatus = "UNCHECKED"
)

// IsValid verifica se lo stato del server SSH è valido
func (s SSHServerStatus) IsValid() bool {
	switch s {
	case SSHServerStatusOnline, SSHServerStatusOffline, SSHServerStatusUnchecked:
		return true
	default:
		return false
	}
}

// SSHServer modella un host remoto Linux gestito per audit e configurazione firewall
type SSHServer struct {
	ID                   uuid.UUID       `json:"id" db:"id"`
	TenantID             uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Name                 string          `json:"name" db:"name"`
	Host                 string          `json:"host" db:"host"`
	Port                 int             `json:"port" db:"port"`
	Username             string          `json:"username" db:"username"`
	AuthType             SSHAuthType     `json:"auth_type" db:"auth_type"`
	EncryptedSecret      string          `json:"-" db:"encrypted_secret"` // Mascherato da qualsiasi output JSON per sicurezza
	PublicKeyFingerprint *string         `json:"public_key_fingerprint,omitempty" db:"public_key_fingerprint"`
	Status               SSHServerStatus `json:"status" db:"status"`
	LastConnectedAt      *time.Time      `json:"last_connected_at,omitempty" db:"last_connected_at"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

// Validate esegue la validazione dei campi obbligatori e dei vincoli logici dell'entità
func (s *SSHServer) Validate() error {
	if s.TenantID == uuid.Nil {
		return ErrValidation
	}
	if strings.TrimSpace(s.Name) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(s.Host) == "" {
		return ErrEmptyHost
	}
	if s.Port <= 0 || s.Port > 65535 {
		return ErrInvalidSSHPort
	}
	if strings.TrimSpace(s.Username) == "" {
		return ErrEmptyUsername
	}
	if !s.AuthType.IsValid() {
		return ErrInvalidSSHAuthType
	}
	if strings.TrimSpace(s.EncryptedSecret) == "" {
		return ErrEmptySecret
	}
	if s.Status == "" {
		s.Status = SSHServerStatusUnchecked
	} else if !s.Status.IsValid() {
		return ErrValidation
	}
	return nil
}

// FirewallBackup rappresenta uno snapshot persistito delle regole iptables di un server
type FirewallBackup struct {
	ID               uuid.UUID `json:"id" db:"id"`
	TenantID         uuid.UUID `json:"tenant_id" db:"tenant_id"`
	ServerID         uuid.UUID `json:"server_id" db:"server_id"`
	RawIptablesRules string    `json:"raw_iptables_rules" db:"raw_iptables_rules"`
	Label            string    `json:"label" db:"label"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// Validate verifica l'integrità dei dati dello snapshot firewall
func (b *FirewallBackup) Validate() error {
	if b.TenantID == uuid.Nil || b.ServerID == uuid.Nil {
		return ErrValidation
	}
	if strings.TrimSpace(b.RawIptablesRules) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(b.Label) == "" {
		return ErrValidation
	}
	return nil
}

// SSHServerRepository definisce i contratti di persistenza per i server remoti SSH
type SSHServerRepository interface {
	Create(ctx context.Context, server *SSHServer) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*SSHServer, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*SSHServer, error)
	Update(ctx context.Context, server *SSHServer) error
	UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status SSHServerStatus, lastConnectedAt *time.Time) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}

// FirewallBackupRepository definisce i contratti di persistenza per gli snapshot delle regole iptables
type FirewallBackupRepository interface {
	CreateBackup(ctx context.Context, backup *FirewallBackup) error
	GetBackupByID(ctx context.Context, tenantID, backupID uuid.UUID) (*FirewallBackup, error)
	GetLatestBackup(ctx context.Context, tenantID, serverID uuid.UUID) (*FirewallBackup, error)
	ListBackups(ctx context.Context, tenantID, serverID uuid.UUID) ([]*FirewallBackup, error)
	DeleteBackup(ctx context.Context, tenantID, backupID uuid.UUID) error
}
