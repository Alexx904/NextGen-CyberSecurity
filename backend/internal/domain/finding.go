package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// FindingConfidence definisce il grado di affidabilità del rilevamento (SEC-P1-04)
type FindingConfidence string

const (
	ConfidenceConfirmed  FindingConfidence = "CONFIRMED"
	ConfidenceProbable   FindingConfidence = "PROBABLE"
	ConfidenceUnverified FindingConfidence = "UNVERIFIED"
)

// DetectionMethod definisce la tecnica euristica o deterministica di identificazione della vulnerabilità (SEC-P1-04)
type DetectionMethod string

const (
	DetectionMethodNSE        DetectionMethod = "nse"
	DetectionMethodBanner     DetectionMethod = "banner"
	DetectionMethodNVDKeyword DetectionMethod = "nvd_keyword"
	DetectionMethodPortState  DetectionMethod = "port_state"
)

// Finding rappresenta un servizio esposto o una vulnerabilità correlata rilevata durante una scansione
type Finding struct {
	ID              uuid.UUID         `json:"id" db:"id"`
	ScanID          uuid.UUID         `json:"scan_id" db:"scan_id"`
	TenantID        uuid.UUID         `json:"tenant_id" db:"tenant_id"`
	HostIP          string            `json:"host_ip" db:"host_ip"`
	Hostname        *string           `json:"hostname,omitempty" db:"hostname"`
	Port            int               `json:"port" db:"port"`
	Protocol        string            `json:"protocol" db:"protocol"`
	PortState       string            `json:"port_state,omitempty" db:"port_state"`
	ServiceName     *string           `json:"service_name,omitempty" db:"service_name"`
	ServiceVersion  *string           `json:"service_version,omitempty" db:"service_version"`
	CVEID           *string           `json:"cve_id,omitempty" db:"cve_id"`
	CVSSScore       *float64          `json:"cvss_score,omitempty" db:"cvss_score"`
	Severity        SeverityLevel     `json:"severity" db:"severity"`
	Confidence      FindingConfidence `json:"confidence,omitempty" db:"confidence"`
	DetectionMethod DetectionMethod   `json:"detection_method,omitempty" db:"detection_method"`
	CPE             *string           `json:"cpe,omitempty" db:"cpe"`
	RawBanner       *string           `json:"raw_banner,omitempty" db:"raw_banner"`
	Remediation     *string           `json:"remediation,omitempty" db:"remediation"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
}

// FindingFilter definisce i parametri di filtro per la consultazione dei finding
type FindingFilter struct {
	ScanID   *uuid.UUID
	HostIP   string
	Severity *SeverityLevel
	Port     *int
	CVEID    string
	Limit    int
	Offset   int
}

// FindingRepository definisce l'interfaccia di persistenza per le vulnerabilità e i servizi rilevati
type FindingRepository interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Finding, error)
	List(ctx context.Context, tenantID uuid.UUID, filter FindingFilter) ([]*Finding, int, error)
	CreateBatch(ctx context.Context, findings []*Finding) error
	GetSeverityCountsByScan(ctx context.Context, tenantID, scanID uuid.UUID) (map[SeverityLevel]int, error)
}

