package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/report"
)

type reportMockPDFGenerator struct {
	generateFunc func(data report.PDFReportData) ([]byte, error)
}

func (m *reportMockPDFGenerator) Generate(data report.PDFReportData) ([]byte, error) {
	if m.generateFunc != nil {
		return m.generateFunc(data)
	}
	return []byte("%PDF-dummy"), nil
}

type reportMockReportRepo struct {
	listFunc func(ctx context.Context, tenantID uuid.UUID, filter domain.ReportFilter) ([]*domain.ReportItem, int, error)
}

func (m *reportMockReportRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.ReportFilter) ([]*domain.ReportItem, int, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, tenantID, filter)
	}
	return nil, 0, nil
}

type reportMockScanRepo struct {
	getByIDFunc func(ctx context.Context, tID, sID uuid.UUID) (*domain.Scan, error)
}

func (m *reportMockScanRepo) GetByID(ctx context.Context, tID, sID uuid.UUID) (*domain.Scan, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, tID, sID)
	}
	return nil, domain.ErrNotFound
}
func (m *reportMockScanRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Scan, int, error) {
	return nil, 0, nil
}
func (m *reportMockScanRepo) Create(ctx context.Context, scan *domain.Scan) error { return nil }
func (m *reportMockScanRepo) Update(ctx context.Context, scan *domain.Scan) error { return nil }
func (m *reportMockScanRepo) GetActiveByTarget(ctx context.Context, tenantID, targetID uuid.UUID) (*domain.Scan, error) {
	return nil, nil
}

type reportMockTenantRepo struct {
	getByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
}

func (m *reportMockTenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, domain.ErrNotFound
}
func (m *reportMockTenantRepo) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	return nil, domain.ErrNotFound
}
func (m *reportMockTenantRepo) Create(ctx context.Context, tenant *domain.Tenant) error { return nil }
func (m *reportMockTenantRepo) UpdateBranding(ctx context.Context, id uuid.UUID, name string, logoURL *string, colors domain.BrandColors) (*domain.Tenant, error) {
	return nil, domain.ErrNotFound
}
func (m *reportMockTenantRepo) List(ctx context.Context) ([]*domain.Tenant, error) { return nil, nil }

type reportMockTargetRepo struct {
	getByIDFunc func(ctx context.Context, tID, id uuid.UUID) (*domain.Target, error)
}

func (m *reportMockTargetRepo) GetByID(ctx context.Context, tID, id uuid.UUID) (*domain.Target, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, tID, id)
	}
	return nil, domain.ErrNotFound
}
func (m *reportMockTargetRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.TargetFilter) ([]*domain.Target, int, error) {
	return nil, 0, nil
}
func (m *reportMockTargetRepo) Create(ctx context.Context, target *domain.Target) error { return nil }
func (m *reportMockTargetRepo) Update(ctx context.Context, target *domain.Target) error { return nil }
func (m *reportMockTargetRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error { return nil }
func (m *reportMockTargetRepo) IsBlacklisted(ctx context.Context, tenantID uuid.UUID, value string) (bool, error) {
	return false, nil
}

type reportMockProfileRepo struct {
	getByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.ScanProfile, error)
}

func (m *reportMockProfileRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ScanProfile, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, domain.ErrNotFound
}
func (m *reportMockProfileRepo) GetByName(ctx context.Context, name string) (*domain.ScanProfile, error) {
	return nil, domain.ErrNotFound
}
func (m *reportMockProfileRepo) List(ctx context.Context) ([]*domain.ScanProfile, error) {
	return nil, nil
}

type reportMockFindingRepo struct {
	listFunc func(ctx context.Context, tID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error)
}

func (m *reportMockFindingRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Finding, error) {
	return nil, domain.ErrNotFound
}
func (m *reportMockFindingRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, tenantID, filter)
	}
	return nil, 0, nil
}
func (m *reportMockFindingRepo) CreateBatch(ctx context.Context, findings []*domain.Finding) error {
	return nil
}
func (m *reportMockFindingRepo) GetSeverityCountsByScan(ctx context.Context, tenantID, scanID uuid.UUID) (map[domain.SeverityLevel]int, error) {
	return nil, nil
}

func TestReportService_ListReports_Success(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()
	groupID := uuid.New()
	groupName := "Core DMZ"
	now := time.Now().UTC()

	mockItems := []*domain.ReportItem{
		{
			ScanID:       uuid.New(),
			TenantID:     tenantID,
			TargetID:     targetID,
			TargetValue:  "192.168.1.0/24",
			TargetType:   domain.TargetTypeCIDR,
			GroupID:      &groupID,
			GroupName:    &groupName,
			ProfileID:    uuid.New(),
			ProfileName:  "Standard Service Scan",
			Status:       domain.ScanStatusCompleted,
			TriggerType:  domain.TriggerManual,
			StartedAt:    &now,
			CompletedAt:  &now,
			MaxSeverity:  domain.SeverityHigh,
			PostureScore: 70,
			PostureGrade: "C",
			CreatedAt:    now,
		},
	}

	reportRepo := &reportMockReportRepo{
		listFunc: func(ctx context.Context, tID uuid.UUID, filter domain.ReportFilter) ([]*domain.ReportItem, int, error) {
			if tID != tenantID {
				t.Fatalf("tenantID non corrispondente: atteso %s, ottenuto %s", tenantID, tID)
			}
			if filter.TargetID == nil || *filter.TargetID != targetID {
				t.Errorf("TargetID errato nel filtro")
			}
			return mockItems, 1, nil
		},
	}

	svc := service.NewReportService(reportRepo, &reportMockScanRepo{}, &reportMockProfileRepo{}, &reportMockTargetRepo{}, &reportMockFindingRepo{}, &reportMockTenantRepo{}, &reportMockPDFGenerator{})

	filter := domain.ReportFilter{
		TargetID: &targetID,
		Limit:    10,
		Offset:   0,
	}

	items, total, err := svc.ListReports(context.Background(), tenantID, filter)
	if err != nil {
		t.Fatalf("ListReports ha restituito un errore inatteso: %v", err)
	}

	if total != 1 {
		t.Errorf("atteso totale 1, ottenuto %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("atteso 1 item, ottenuti %d", len(items))
	}
	if items[0].PostureScore != 70 || items[0].PostureGrade != "C" {
		t.Errorf("valori posture score errati: score=%d, grade=%s", items[0].PostureScore, items[0].PostureGrade)
	}
}

func TestReportService_ListReports_Validation(t *testing.T) {
	svc := service.NewReportService(&reportMockReportRepo{}, &reportMockScanRepo{}, &reportMockProfileRepo{}, &reportMockTargetRepo{}, &reportMockFindingRepo{}, &reportMockTenantRepo{}, &reportMockPDFGenerator{})

	_, _, err := svc.ListReports(context.Background(), uuid.Nil, domain.ReportFilter{})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("atteso ErrValidation per tenantID Nil, ottenuto %v", err)
	}
}

func TestReportService_ListReports_RepoError(t *testing.T) {
	tenantID := uuid.New()
	reportRepo := &reportMockReportRepo{
		listFunc: func(ctx context.Context, tID uuid.UUID, filter domain.ReportFilter) ([]*domain.ReportItem, int, error) {
			return nil, 0, errors.New("database connection down")
		},
	}

	svc := service.NewReportService(reportRepo, &reportMockScanRepo{}, &reportMockProfileRepo{}, &reportMockTargetRepo{}, &reportMockFindingRepo{}, &reportMockTenantRepo{}, &reportMockPDFGenerator{})

	_, _, err := svc.ListReports(context.Background(), tenantID, domain.ReportFilter{})
	if err == nil {
		t.Fatal("atteso errore dal repository, ottenuto nil")
	}
}

func TestReportService_GenerateScanPDFReport_Success(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()
	targetID := uuid.New()
	profileID := uuid.New()

	scanRepo := &reportMockScanRepo{
		getByIDFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.Scan, error) {
			return &domain.Scan{
				ID:        scanID,
				TenantID:  tenantID,
				TargetID:  targetID,
				ProfileID: profileID,
				Status:    domain.ScanStatusCompleted,
				CreatedAt: time.Now(),
			}, nil
		},
	}

	tenantRepo := &reportMockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return &domain.Tenant{
				ID:          tenantID,
				Name:        "Cyber Shield Inc",
				BrandColors: domain.DefaultBrandColors(),
			}, nil
		},
	}

	targetRepo := &reportMockTargetRepo{
		getByIDFunc: func(ctx context.Context, tID, id uuid.UUID) (*domain.Target, error) {
			return &domain.Target{
				ID:       targetID,
				TenantID: tenantID,
				Value:    "192.168.1.1",
				Type:     domain.TargetTypeIP,
			}, nil
		},
	}

	profileRepo := &reportMockProfileRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.ScanProfile, error) {
			return &domain.ScanProfile{
				ID:       profileID,
				Name:     "Fast Scan",
				NmapArgs: "-F",
			}, nil
		},
	}

	findingRepo := &reportMockFindingRepo{
		listFunc: func(ctx context.Context, tID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
			return []*domain.Finding{
				{
					ID:       uuid.New(),
					ScanID:   scanID,
					TenantID: tenantID,
					HostIP:   "192.168.1.1",
					Port:     80,
					Protocol: "tcp",
					Severity: domain.SeverityMedium,
				},
			}, 1, nil
		},
	}

	generatorCalled := false
	var passedTemplate domain.ReportTemplate
	pdfGen := &reportMockPDFGenerator{
		generateFunc: func(data report.PDFReportData) ([]byte, error) {
			generatorCalled = true
			passedTemplate = data.Template
			if data.Tenant.Name != "Cyber Shield Inc" || data.Target.Value != "192.168.1.1" {
				t.Errorf("dati non corretti passati al generatore PDF: %+v", data)
			}
			return []byte("%PDF-1.4 test stream"), nil
		},
	}

	svc := service.NewReportService(&reportMockReportRepo{}, scanRepo, profileRepo, targetRepo, findingRepo, tenantRepo, pdfGen)

	// Test con template Executive
	pdfBytes, filename, err := svc.GenerateScanPDFReport(context.Background(), tenantID, scanID, domain.ReportTemplateExecutive)
	if err != nil {
		t.Fatalf("atteso successo, ottenuto errore: %v", err)
	}

	if !generatorCalled {
		t.Errorf("il generatore PDF doveva essere invocato")
	}
	if passedTemplate != domain.ReportTemplateExecutive {
		t.Errorf("atteso template %s, ottenuto %s", domain.ReportTemplateExecutive, passedTemplate)
	}
	if len(pdfBytes) == 0 {
		t.Errorf("byte PDF vuoti")
	}
	if filename == "" {
		t.Errorf("nome file non deve essere vuoto")
	}

	// Test con template Technical
	generatorCalled = false
	_, _, err = svc.GenerateScanPDFReport(context.Background(), tenantID, scanID, domain.ReportTemplateTechnical)
	if err != nil {
		t.Fatalf("atteso successo per template Technical, ottenuto: %v", err)
	}
	if passedTemplate != domain.ReportTemplateTechnical {
		t.Errorf("atteso template Technical, ottenuto %s", passedTemplate)
	}

	// Test fallback su stringa vuota (deve fallbackare su Executive)
	_, _, err = svc.GenerateScanPDFReport(context.Background(), tenantID, scanID, "")
	if err != nil {
		t.Fatalf("atteso successo per fallback template, ottenuto: %v", err)
	}
	if passedTemplate != domain.ReportTemplateExecutive {
		t.Errorf("atteso fallback su Executive, ottenuto %s", passedTemplate)
	}
}

func TestReportService_GenerateScanPDFReport_Validation(t *testing.T) {
	svc := service.NewReportService(&reportMockReportRepo{}, &reportMockScanRepo{}, &reportMockProfileRepo{}, &reportMockTargetRepo{}, &reportMockFindingRepo{}, &reportMockTenantRepo{}, &reportMockPDFGenerator{})

	// Nil tenant ID
	_, _, err := svc.GenerateScanPDFReport(context.Background(), uuid.Nil, uuid.New(), domain.ReportTemplateExecutive)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("atteso ErrValidation per Nil tenantID, ottenuto: %v", err)
	}

	// Nil scan ID
	_, _, err = svc.GenerateScanPDFReport(context.Background(), uuid.New(), uuid.Nil, domain.ReportTemplateExecutive)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("atteso ErrValidation per Nil scanID, ottenuto: %v", err)
	}
}

func TestReportService_GenerateScanPDFReport_ScanNotFound(t *testing.T) {
	scanRepo := &reportMockScanRepo{
		getByIDFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.Scan, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := service.NewReportService(&reportMockReportRepo{}, scanRepo, &reportMockProfileRepo{}, &reportMockTargetRepo{}, &reportMockFindingRepo{}, &reportMockTenantRepo{}, &reportMockPDFGenerator{})

	_, _, err := svc.GenerateScanPDFReport(context.Background(), uuid.New(), uuid.New(), domain.ReportTemplateExecutive)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("atteso ErrNotFound per scan inesistente, ottenuto: %v", err)
	}
}

func TestReportService_GenerateScanPDFReport_TenantNotFound(t *testing.T) {
	scanRepo := &reportMockScanRepo{
		getByIDFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.Scan, error) {
			return &domain.Scan{
				ID:        sID,
				TenantID:  tID,
				TargetID:  uuid.New(),
				ProfileID: uuid.New(),
			}, nil
		},
	}
	tenantRepo := &reportMockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := service.NewReportService(&reportMockReportRepo{}, scanRepo, &reportMockProfileRepo{}, &reportMockTargetRepo{}, &reportMockFindingRepo{}, tenantRepo, &reportMockPDFGenerator{})

	_, _, err := svc.GenerateScanPDFReport(context.Background(), uuid.New(), uuid.New(), domain.ReportTemplateExecutive)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("atteso ErrNotFound per tenant inesistente, ottenuto: %v", err)
	}
}

func TestReportService_GenerateScanCSVReport_Success(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	scanRepo := &reportMockScanRepo{
		getByIDFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.Scan, error) {
			return &domain.Scan{
				ID:        scanID,
				TenantID:  tenantID,
				Status:    domain.ScanStatusCompleted,
				CreatedAt: time.Now(),
			}, nil
		},
	}

	findingRepo := &reportMockFindingRepo{
		listFunc: func(ctx context.Context, tID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
			return []*domain.Finding{
				{
					ID:        uuid.New(),
					ScanID:    scanID,
					TenantID:  tenantID,
					HostIP:    "10.0.0.5",
					Port:      443,
					Protocol:  "tcp",
					Severity:  domain.SeverityHigh,
					CreatedAt: time.Now(),
				},
			}, 1, nil
		},
	}

	svc := service.NewReportService(&reportMockReportRepo{}, scanRepo, &reportMockProfileRepo{}, &reportMockTargetRepo{}, findingRepo, &reportMockTenantRepo{}, &reportMockPDFGenerator{})

	csvBytes, filename, err := svc.GenerateScanCSVReport(context.Background(), tenantID, scanID)
	if err != nil {
		t.Fatalf("generazione CSV report fallita: %v", err)
	}

	if len(csvBytes) == 0 {
		t.Errorf("byte CSV vuoti")
	}
	if filename == "" {
		t.Errorf("nome file non deve essere vuoto")
	}
}

func TestReportService_GenerateScanCSVReport_Validation(t *testing.T) {
	svc := service.NewReportService(&reportMockReportRepo{}, &reportMockScanRepo{}, &reportMockProfileRepo{}, &reportMockTargetRepo{}, &reportMockFindingRepo{}, &reportMockTenantRepo{}, &reportMockPDFGenerator{})

	_, _, err := svc.GenerateScanCSVReport(context.Background(), uuid.Nil, uuid.New())
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("atteso ErrValidation per Nil tenantID, ottenuto: %v", err)
	}

	_, _, err = svc.GenerateScanCSVReport(context.Background(), uuid.New(), uuid.Nil)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("atteso ErrValidation per Nil scanID, ottenuto: %v", err)
	}
}

func TestReportService_GenerateScanCSVReport_ScanNotFound(t *testing.T) {
	scanRepo := &reportMockScanRepo{
		getByIDFunc: func(ctx context.Context, tID, sID uuid.UUID) (*domain.Scan, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := service.NewReportService(&reportMockReportRepo{}, scanRepo, &reportMockProfileRepo{}, &reportMockTargetRepo{}, &reportMockFindingRepo{}, &reportMockTenantRepo{}, &reportMockPDFGenerator{})

	_, _, err := svc.GenerateScanCSVReport(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("atteso ErrNotFound, ottenuto: %v", err)
	}
}
