package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/report"
)

// ReportService definisce i casi d'uso per l'hub dei report e la generazione di documenti di sicurezza
type ReportService interface {
	ListReports(ctx context.Context, tenantID uuid.UUID, filter domain.ReportFilter) ([]*domain.ReportItem, int, error)
	GenerateScanPDFReport(ctx context.Context, tenantID, scanID uuid.UUID, template domain.ReportTemplate) ([]byte, string, error)
	GenerateScanCSVReport(ctx context.Context, tenantID, scanID uuid.UUID) ([]byte, string, error)
	GenerateCompanyPDFReport(ctx context.Context, tenantID, groupID uuid.UUID, template domain.ReportTemplate) ([]byte, string, error)
	GenerateCompanyCSVReport(ctx context.Context, tenantID, groupID uuid.UUID) ([]byte, string, error)
}

type reportService struct {
	reportRepo      domain.ReportRepository
	scanRepo        domain.ScanRepository
	scanProfileRepo domain.ScanProfileRepository
	targetRepo      domain.TargetRepository
	targetGroupRepo domain.TargetGroupRepository
	findingRepo     domain.FindingRepository
	tenantRepo      domain.TenantRepository
	pdfGenerator    report.PDFGenerator
}

// NewReportService crea una nuova istanza di ReportService
func NewReportService(
	reportRepo domain.ReportRepository,
	scanRepo domain.ScanRepository,
	scanProfileRepo domain.ScanProfileRepository,
	targetRepo domain.TargetRepository,
	findingRepo domain.FindingRepository,
	tenantRepo domain.TenantRepository,
	pdfGenerator report.PDFGenerator,
	targetGroupRepo ...domain.TargetGroupRepository,
) ReportService {
	if pdfGenerator == nil {
		pdfGenerator = report.NewPDFGenerator()
	}
	var groupRepo domain.TargetGroupRepository
	if len(targetGroupRepo) > 0 {
		groupRepo = targetGroupRepo[0]
	}
	return &reportService{
		reportRepo:      reportRepo,
		scanRepo:        scanRepo,
		scanProfileRepo: scanProfileRepo,
		targetRepo:      targetRepo,
		targetGroupRepo: groupRepo,
		findingRepo:     findingRepo,
		tenantRepo:      tenantRepo,
		pdfGenerator:    pdfGenerator,
	}
}

// ListReports recupera l'elenco paginato dei report aggregati per il Reports Hub
func (s *reportService) ListReports(ctx context.Context, tenantID uuid.UUID, filter domain.ReportFilter) ([]*domain.ReportItem, int, error) {
	if tenantID == uuid.Nil {
		return nil, 0, domain.ErrValidation
	}
	if s.reportRepo == nil {
		return nil, 0, errors.New("report repository non configurato")
	}
	return s.reportRepo.List(ctx, tenantID, filter)
}

// GenerateScanPDFReport aggrega tutti i dettagli di una scansione e compila il report PDF white-label nel template scelto
func (s *reportService) GenerateScanPDFReport(ctx context.Context, tenantID, scanID uuid.UUID, template domain.ReportTemplate) ([]byte, string, error) {
	if tenantID == uuid.Nil || scanID == uuid.Nil {
		return nil, "", domain.ErrValidation
	}

	if !template.IsValid() {
		template = domain.ReportTemplateExecutive
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

	// 6. Generazione del PDF tramite motore dedicato con il template specificato
	now := time.Now().UTC()
	reportData := report.PDFReportData{
		Tenant:      tenant,
		Scan:        scan,
		Target:      target,
		Profile:     profile,
		Findings:    findings,
		Template:    template,
		GeneratedAt: now,
	}

	pdfBytes, err := s.pdfGenerator.Generate(reportData)
	if err != nil {
		return nil, "", fmt.Errorf("generazione del documento PDF fallita: %w", err)
	}

	filename := fmt.Sprintf("report_%s_scan_%s_%s.pdf", template, scan.ID.String()[:8], now.Format("20060102_150405"))
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

// GenerateCompanyPDFReport compila il documento PDF white-label consolidato per l'intera Azienda (tutti gli IP censiti)
func (s *reportService) GenerateCompanyPDFReport(ctx context.Context, tenantID, groupID uuid.UUID, template domain.ReportTemplate) ([]byte, string, error) {
	if tenantID == uuid.Nil || groupID == uuid.Nil {
		return nil, "", domain.ErrValidation
	}
	if s.targetGroupRepo == nil {
		return nil, "", errors.New("target group repository non configurato")
	}

	if !template.IsValid() {
		template = domain.ReportTemplateExecutive
	}

	// 1. Recupero metadati azienda / cartella scope
	group, err := s.targetGroupRepo.GetByID(ctx, tenantID, groupID)
	if err != nil {
		return nil, "", err
	}

	// 2. Recupero branding e metadati del tenant
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, "", err
	}

	targets, reportItems, err := s.loadReportScope(ctx, tenantID, groupID)
	if err != nil {
		return nil, "", err
	}

	// 5. Raccolta findings di tutti gli host dell'azienda con deduplica
	var allFindings []*domain.Finding
	seenFinding := make(map[string]bool)
	upHostsCount := 0

	for _, item := range reportItems {
		if item.SummaryStats.UpHosts > 0 {
			upHostsCount += item.SummaryStats.UpHosts
		}
		if item.Status != domain.ScanStatusCompleted {
			continue
		}

		itemFindings, err := s.loadReportFindings(ctx, tenantID, item.ScanID)
		if err != nil {
			return nil, "", err
		}
		{
			for _, f := range itemFindings {
				key := fmt.Sprintf("%s:%s:%d:%s:%s", item.TargetID, f.HostIP, f.Port, f.Protocol, getFindingDedupKey(f))
				if !seenFinding[key] {
					seenFinding[key] = true
					allFindings = append(allFindings, f)
				}
			}
		}
	}

	if upHostsCount > len(targets) {
		upHostsCount = len(targets)
	}

	// Calcolo conteggi severità aggregati
	critCount, highCount, medCount, lowCount, infoCount := 0, 0, 0, 0, 0
	for _, f := range allFindings {
		switch f.Severity {
		case domain.SeverityCritical:
			critCount++
		case domain.SeverityHigh:
			highCount++
		case domain.SeverityMedium:
			medCount++
		case domain.SeverityLow:
			lowCount++
		case domain.SeverityInfo:
			infoCount++
		}
	}

	// 6. Preparazione contesto sintetico per l'Azienda
	now := time.Now().UTC()
	companyDesc := fmt.Sprintf("Dossier di Sicurezza Aziendale Consolidato (%d IP/Target censiti)", len(targets))
	if group.Description != nil && *group.Description != "" {
		companyDesc = fmt.Sprintf("%s | %s", *group.Description, companyDesc)
	}

	companyTarget := &domain.Target{
		ID:          group.ID,
		TenantID:    tenantID,
		GroupID:     &groupID,
		Type:        domain.TargetType("COMPANY_SCOPE"),
		Value:       group.Name,
		Description: &companyDesc,
	}

	companyProfile := &domain.ScanProfile{
		ID:       uuid.Nil,
		Name:     "Audit Consolidato Perimetrale Aziendale",
		NmapArgs: fmt.Sprintf("Consolidato su %d target dell'azienda", len(targets)),
	}

	// Calcolo tempistiche consolidate per valorizzare Data Avvio e Durata Totale
	var earliestStarted *time.Time
	var latestCompleted *time.Time
	totalDurationSec := 0.0

	for _, item := range reportItems {
		if item.StartedAt != nil {
			if earliestStarted == nil || item.StartedAt.Before(*earliestStarted) {
				t := *item.StartedAt
				earliestStarted = &t
			}
		}
		if item.CompletedAt != nil {
			if latestCompleted == nil || item.CompletedAt.After(*latestCompleted) {
				t := *item.CompletedAt
				latestCompleted = &t
			}
		}
		totalDurationSec += float64(item.SummaryStats.DurationSeconds)
	}

	if earliestStarted == nil {
		t := group.CreatedAt
		earliestStarted = &t
	}
	if latestCompleted == nil {
		latestCompleted = &now
	}

	companyScan := &domain.Scan{
		ID:          group.ID,
		TenantID:    tenantID,
		TargetID:    group.ID,
		Status:      domain.ScanStatusCompleted,
		TriggerType: domain.TriggerManual,
		StartedAt:   earliestStarted,
		CompletedAt: latestCompleted,
		SummaryStats: domain.ScanSummaryStats{
			TotalHosts:      len(targets),
			UpHosts:         upHostsCount,
			DownHosts:       len(targets) - upHostsCount,
			CriticalCount:   critCount,
			HighCount:       highCount,
			MediumCount:     medCount,
			LowCount:        lowCount,
			InfoCount:       infoCount,
			DurationSeconds: totalDurationSec,
		},
		CreatedAt: group.CreatedAt,
	}

	reportData := report.PDFReportData{
		Tenant:      tenant,
		Scan:        companyScan,
		Target:      companyTarget,
		Profile:     companyProfile,
		Findings:    allFindings,
		Template:    template,
		GeneratedAt: now,
	}
	sections, _, grouped, err := s.departmentInventory(ctx, tenantID, group, targets, reportItems, allFindings)
	if err != nil {
		return nil, "", err
	}
	reportData.Departments = sections
	reportData.GroupFindingsByDepartment = grouped

	pdfBytes, err := s.pdfGenerator.Generate(reportData)
	if err != nil {
		return nil, "", fmt.Errorf("generazione del documento PDF aziendale fallita: %w", err)
	}

	safeName := sanitizeFilename(group.Name)
	filename := fmt.Sprintf("report_azienda_%s_%s_%s.pdf", template, safeName, now.Format("20060102_150405"))
	return pdfBytes, filename, nil
}

// GenerateCompanyCSVReport compila il report CSV consolidato per l'intera Azienda
func (s *reportService) GenerateCompanyCSVReport(ctx context.Context, tenantID, groupID uuid.UUID) ([]byte, string, error) {
	if tenantID == uuid.Nil || groupID == uuid.Nil {
		return nil, "", domain.ErrValidation
	}
	if s.targetGroupRepo == nil {
		return nil, "", errors.New("target group repository non configurato")
	}

	group, err := s.targetGroupRepo.GetByID(ctx, tenantID, groupID)
	if err != nil {
		return nil, "", err
	}

	targets, reportItems, err := s.loadReportScope(ctx, tenantID, groupID)
	if err != nil {
		return nil, "", err
	}

	var allFindings []*domain.Finding
	seenFinding := make(map[string]bool)

	for _, item := range reportItems {
		if item.Status != domain.ScanStatusCompleted {
			continue
		}
		itemFindings, err := s.loadReportFindings(ctx, tenantID, item.ScanID)
		if err != nil {
			return nil, "", err
		}
		{
			for _, f := range itemFindings {
				key := fmt.Sprintf("%s:%s:%d:%s:%s", item.TargetID, f.HostIP, f.Port, f.Protocol, getFindingDedupKey(f))
				if !seenFinding[key] {
					seenFinding[key] = true
					allFindings = append(allFindings, f)
				}
			}
		}
	}

	_, scanDepartments, _, err := s.departmentInventory(ctx, tenantID, group, targets, reportItems, allFindings)
	if err != nil {
		return nil, "", err
	}
	csvBytes, err := report.GenerateDepartmentCSV(allFindings, scanDepartments)
	if err != nil {
		return nil, "", fmt.Errorf("generazione del report CSV aziendale fallita: %w", err)
	}

	now := time.Now().UTC()
	safeName := sanitizeFilename(group.Name)
	filename := fmt.Sprintf("findings_azienda_%s_%s.csv", safeName, now.Format("20060102_150405"))
	return csvBytes, filename, nil
}

func sanitizeFilename(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if r == ' ' {
			b.WriteRune('_')
		}
	}
	res := b.String()
	if res == "" {
		return "company"
	}
	return res
}

func getFindingDedupKey(f *domain.Finding) string {
	if f.CVEID != nil && *f.CVEID != "" {
		return *f.CVEID
	}
	if f.ServiceName != nil && *f.ServiceName != "" {
		return *f.ServiceName
	}
	return f.Protocol
}
