package domain

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
)

// ReportTemplate definisce la tipologia di template grafico per l'esportazione del report PDF
type ReportTemplate string

const (
	ReportTemplateExecutive ReportTemplate = "executive"
	ReportTemplateTechnical ReportTemplate = "technical"
)

// IsValid verifica se la stringa corrisponde a un template supportato
func (t ReportTemplate) IsValid() bool {
	switch t {
	case ReportTemplateExecutive, ReportTemplateTechnical:
		return true
	default:
		return false
	}
}

// ReportFormat modella il formato di esportazione supportato (PDF o CSV)
type ReportFormat string

const (
	ReportFormatPDF ReportFormat = "pdf"
	ReportFormatCSV ReportFormat = "csv"
)

// ReportFilter definisce i criteri di ricerca e paginazione per l'elenco dei report
type ReportFilter struct {
	TargetID   *uuid.UUID
	GroupID    *uuid.UUID
	Severity   *SeverityLevel
	TargetType *TargetType
	DateFrom   *time.Time
	DateTo     *time.Time
	Search     string
	Status     *ScanStatus
	Limit      int
	Offset     int
}

// ReportItem rappresenta una voce dello storico report nell'Hub centralizzato
type ReportItem struct {
	ScanID       uuid.UUID        `json:"scan_id"`
	TenantID     uuid.UUID        `json:"tenant_id"`
	TargetID     uuid.UUID        `json:"target_id"`
	TargetValue  string           `json:"target_value"`
	TargetType   TargetType       `json:"target_type"`
	GroupID      *uuid.UUID       `json:"group_id,omitempty"`
	GroupName    *string          `json:"group_name,omitempty"`
	ProfileID    uuid.UUID        `json:"profile_id"`
	ProfileName  string           `json:"profile_name"`
	Status       ScanStatus       `json:"status"`
	TriggerType  ScanTriggerType  `json:"trigger_type"`
	StartedAt    *time.Time       `json:"started_at,omitempty"`
	CompletedAt  *time.Time       `json:"completed_at,omitempty"`
	SummaryStats ScanSummaryStats `json:"summary_stats"`
	MaxSeverity  SeverityLevel    `json:"max_severity"`
	PostureScore int              `json:"posture_score"` // Punteggio 0 - 100
	PostureGrade string           `json:"posture_grade"` // Rating: A, B, C, D, F
	CreatedAt    time.Time        `json:"created_at"`
}

// ReportRepository definisce l'interfaccia di persistenza per l'aggregazione dei report
type ReportRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, filter ReportFilter) ([]*ReportItem, int, error)
}

// CalculatePostureScore calcola il punteggio di postura (0-100) e il rating a lettere (A-F)
// partendo da 100 con penalità ponderate per livello di severità delle vulnerabilità
func CalculatePostureScore(stats ScanSummaryStats) (int, string) {
	deduction := float64(stats.CriticalCount*25 + stats.HighCount*15 + stats.MediumCount*5 + stats.LowCount*1)

	// Normalizzazione per perimetri multi-host consolidati (densità vulnerabilità per host)
	hosts := stats.UpHosts
	if hosts < 1 {
		hosts = stats.TotalHosts
	}
	if hosts > 1 {
		deduction = deduction / math.Sqrt(float64(hosts))
	}

	score := int(math.Round(100.0 - deduction))

	// Coerenza dei dati: se non vi è alcuna vulnerabilità critica (CVSS >= 9.0),
	// la postura non collassa a 0/100 (riservato alla compromissione critica confermata).
	// Senza criticità, il punteggio minimo con sole vulnerabilità High/Medium è mantenuto in fascia D (min 25).
	if stats.CriticalCount == 0 && score < 25 && (stats.HighCount > 0 || stats.MediumCount > 0) {
		score = 25
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	var grade string
	switch {
	case score >= 90:
		grade = "A"
	case score >= 75:
		grade = "B"
	case score >= 50:
		grade = "C"
	case score >= 25:
		grade = "D"
	default:
		grade = "F"
	}
	return score, grade
}

// CalculateMaxSeverity determina la severità massima rilevata in base alle statistiche sintetiche
func CalculateMaxSeverity(stats ScanSummaryStats) SeverityLevel {
	if stats.CriticalCount > 0 {
		return SeverityCritical
	}
	if stats.HighCount > 0 {
		return SeverityHigh
	}
	if stats.MediumCount > 0 {
		return SeverityMedium
	}
	if stats.LowCount > 0 {
		return SeverityLow
	}
	return SeverityInfo
}
