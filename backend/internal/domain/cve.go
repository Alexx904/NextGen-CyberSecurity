package domain

import (
	"context"
	"time"
)

// SeverityLevel definisce i livelli di criticità delle vulnerabilità
type SeverityLevel string

const (
	SeverityCritical SeverityLevel = "CRITICAL"
	SeverityHigh     SeverityLevel = "HIGH"
	SeverityMedium   SeverityLevel = "MEDIUM"
	SeverityLow      SeverityLevel = "LOW"
	SeverityInfo     SeverityLevel = "INFO"
)

// IsValid verifica se il livello di severità è contemplato
func (s SeverityLevel) IsValid() bool {
	switch s {
	case SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo:
		return true
	default:
		return false
	}
}

// SeverityFromCVSS calcola la severità qualitativa a partire dal punteggio CVSS v3.1 numerico
func SeverityFromCVSS(score float64) SeverityLevel {
	switch {
	case score >= 9.0:
		return SeverityCritical
	case score >= 7.0:
		return SeverityHigh
	case score >= 4.0:
		return SeverityMedium
	case score > 0.0:
		return SeverityLow
	default:
		return SeverityInfo
	}
}

// CPEMatchCriterion descrive i criteri di matching CPE (Common Platform Enumeration)
type CPEMatchCriterion struct {
	CPE23URI              string `json:"cpe23Uri"`
	Vulnerable            bool   `json:"vulnerable"`
	VersionStartIncluding string `json:"versionStartIncluding,omitempty"`
	VersionEndExcluding   string `json:"versionEndExcluding,omitempty"`
}

// CVEEntry rappresenta una voce del catalogo locale delle vulnerabilità note (Common Vulnerabilities and Exposures)
type CVEEntry struct {
	CVEID            string               `json:"cve_id" db:"cve_id"`
	CVSSScore        *float64             `json:"cvss_score,omitempty" db:"cvss_score"`
	Severity         SeverityLevel        `json:"severity" db:"severity"`
	Description      string               `json:"description" db:"description"`
	CPEMatchCriteria []CPEMatchCriterion `json:"cpe_match_criteria" db:"cpe_match_criteria"`
	PublishedDate    *time.Time           `json:"published_date,omitempty" db:"published_date"`
	UpdatedAt        time.Time            `json:"updated_at" db:"updated_at"`
}

// CVERepository definisce i metodi di accesso al catalogo CVE locale (Offline-Ready)
type CVERepository interface {
	GetByID(ctx context.Context, cveID string) (*CVEEntry, error)
	SearchByCPE(ctx context.Context, cpePrefix string, version string) ([]*CVEEntry, error)
	UpsertBatch(ctx context.Context, entries []*CVEEntry) error
	Count(ctx context.Context) (int64, error)
}

