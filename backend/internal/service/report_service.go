package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/report"
)

// ReportService definisce i casi d'uso per la generazione di report di sicurezza
type ReportService interface {
	GenerateScanPDFReport(ctx context.Context, tenantID, scanID uuid.UUID) ([]byte, string, error)
	GenerateScanCSVReport(ctx context.Context, tenantID, scanID uuid.UUID) ([]byte, string, error)
}

type reportService struct {
	scanRepo        domain.ScanRepository
	scanProfileRepo domain.ScanProfileRepository
	targetRepo      domain.TargetRepository
	findingRepo     domain.FindingRepository
	tenantRepo      domain.TenantRepository
	pdfGenerator    report.PDFGenerator
}

// NewReportService crea una nuova istanza di ReportService
func NewReportService(
	scanRepo domain.ScanRepository,
	scanProfileRepo domain.ScanProfileRepository,
	targetRepo domain.TargetRepository,
	findingRepo domain.FindingRepository,
	tenantRepo domain.TenantRepository,
	pdfGenerator report.PDFGenerator,
) ReportService {
	if pdfGenerator == nil {
		pdfGenerator = report.NewPDFGenerator()
	}
	return &reportService{
		scanRepo:        scanRepo,
		scanProfileRepo: scanProfileRepo,
		targetRepo:      targetRepo,
		findingRepo:     findingRepo,
		tenantRepo:      tenantRepo,
		pdfGenerator:    pdfGenerator,
	}
}

// GenerateScanPDFReport aggrega tutti i dettagli di una scansione e compila il report PDF white-label
func (s *reportService) GenerateScanPDFReport(ctx context.Context, tenantID, scanID uuid.UUID) ([]byte, string, error) {
	if tenantID == uuid.Nil || scanID == uuid.Nil {
		return nil, "", domain.ErrValidation
	}

	// 1. Recupero scansione con verifica rigorosa di appartenenza al tenant
	scan, err := s.scanRepo.GetByID(ctx, tenantID, scanID)
	if err != nil {
		return nil, "", err
	}

	// 2. Recupero metadati e branding del tenant
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, "", err
	}

	// 3. Recupero target analizzato
	target, err := s.targetRepo.GetByID(ctx, tenantID, scan.TargetID)
	if err != nil {
		return nil, "", err
	}

	// 4. Recupero profilo di scansione utilizzato
	profile, err := s.scanProfileRepo.GetByID(ctx, scan.ProfileID)
	if err != nil {
		return nil, "", err
	}

	// 5. Recupero findings della scansione
	filter := domain.FindingFilter{
		ScanID: &scanID,
		Limit:  1000,
		Offset: 0,
	}
	findings, _, err := s.findingRepo.List(ctx, tenantID, filter)
	if err != nil {
		return nil, "", err
	}

	// 6. Generazione del PDF tramite motore dedicato
	now := time.Now().UTC()
	reportData := report.PDFReportData{
		Tenant:      tenant,
		Scan:        scan,
		Target:      target,
		Profile:     profile,
		Findings:    findings,
		GeneratedAt: now,
	}

	pdfBytes, err := s.pdfGenerator.Generate(reportData)
	if err != nil {
		return nil, "", fmt.Errorf("generazione del documento PDF fallita: %w", err)
	}

	filename := fmt.Sprintf("report_scan_%s_%s.pdf", scan.ID.String()[:8], now.Format("20060102_150405"))
	return pdfBytes, filename, nil
}

// GenerateScanCSVReport aggrega i finding di una scansione e compila il report tabellare CSV conforme a RFC 4180
func (s *reportService) GenerateScanCSVReport(ctx context.Context, tenantID, scanID uuid.UUID) ([]byte, string, error) {
	if tenantID == uuid.Nil || scanID == uuid.Nil {
		return nil, "", domain.ErrValidation
	}

	// 1. Verifica esistenza e appartenenza al tenant della scansione
	scan, err := s.scanRepo.GetByID(ctx, tenantID, scanID)
	if err != nil {
		return nil, "", err
	}

	// 2. Recupero di tutti i finding associati alla scansione
	filter := domain.FindingFilter{
		ScanID: &scanID,
		Limit:  10000,
		Offset: 0,
	}
	findings, _, err := s.findingRepo.List(ctx, tenantID, filter)
	if err != nil {
		return nil, "", err
	}

	// 3. Generazione stream CSV conforme a RFC 4180
	csvBytes, err := report.GenerateCSV(findings)
	if err != nil {
		return nil, "", fmt.Errorf("generazione del report CSV fallita: %w", err)
	}

	now := time.Now().UTC()
	filename := fmt.Sprintf("findings_scan_%s_%s.csv", scan.ID.String()[:8], now.Format("20060102_150405"))
	return csvBytes, filename, nil
}


