package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DiscoveredDevice rappresenta un singolo dispositivo attivo censito nella rete locale
type DiscoveredDevice struct {
	IP        string    `json:"ip"`
	MAC       string    `json:"mac,omitempty"`
	Vendor    string    `json:"vendor,omitempty"`
	Hostname  string    `json:"hostname,omitempty"`
	LatencyMs float64   `json:"latency_ms"`
	Status    string    `json:"status"` // "up"
	OpenPorts []int     `json:"open_ports,omitempty"`
	ScannedAt time.Time `json:"scanned_at"`
}

// DiscoveryResult racchiude il risultato complessivo di un'operazione di ricognizione LAN
type DiscoveryResult struct {
	ID              uuid.UUID          `json:"id"`
	TenantID        uuid.UUID          `json:"tenant_id"`
	Subnet          string             `json:"subnet"`
	Devices         []DiscoveredDevice `json:"devices"`
	TotalDiscovered int                `json:"total_discovered"`
	DurationSeconds float64            `json:"duration_seconds"`
	StartedAt       time.Time          `json:"started_at"`
	CompletedAt     time.Time          `json:"completed_at"`
}

// DiscoveryRequest contiene i parametri di richiesta per la scansione rapida della subnet
type DiscoveryRequest struct {
	Subnet             string     `json:"subnet"`
	IncludeCommonPorts bool       `json:"include_common_ports,omitempty"`
	SessionID          *uuid.UUID `json:"session_id,omitempty"`
}

// DiscoveryService definisce i casi d'uso per la ricognizione ad alta velocità della rete locale
type DiscoveryService interface {
	DiscoverSubnet(ctx context.Context, tenantID uuid.UUID, req DiscoveryRequest) (*DiscoveryResult, error)
	GetLastResults(ctx context.Context, tenantID uuid.UUID) (*DiscoveryResult, error)
	GetSessionResults(ctx context.Context, tenantID, sessionID uuid.UUID) (*DiscoveryResult, error)
	LookupVendor(mac string) string
}

// DiscoveryRepository stores snapshots before they are acknowledged over SSE.
type DiscoveryRepository interface {
	Create(ctx context.Context, result *DiscoveryResult) error
	SaveDevice(ctx context.Context, tenantID, sessionID uuid.UUID, device DiscoveredDevice) error
	Complete(ctx context.Context, result *DiscoveryResult) error
	GetLast(ctx context.Context, tenantID uuid.UUID) (*DiscoveryResult, error)
	GetSession(ctx context.Context, tenantID, sessionID uuid.UUID) (*DiscoveryResult, error)
}

// TenantDiscoveryTopic also isolates subscriptions made before session creation.
func TenantDiscoveryTopic(tenantID, sessionID uuid.UUID) string {
	return "discovery:" + tenantID.String() + ":" + sessionID.String()
}
