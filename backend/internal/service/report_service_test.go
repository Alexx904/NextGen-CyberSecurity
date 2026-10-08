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
	pdfGen := &reportMockPDFGenerator{
		generateFunc: func(data report.PDFReportData) ([]byte, error) {
			generatorCalled = true
			if data.Tenant.Name != "Cyber Shield Inc" || data.Target.Value != "192.168.1.1" {
				t.Errorf("dati non corretti passati al generatore PDF: %+v", data)
			}
			return []byte("%PDF-1.4 test stream"), nil
		},
	}

	svc := service.NewReportService(scanRepo, profileRepo, targetRepo, findingRepo, tenantRepo, pdfGen)

	pdfBytes, filename, err := svc.GenerateScanPDFReport(context.Background(), tenantID, scanID)
	if err != nil {
		t.Fatalf("atteso successo, ottenuto errore: %v", err)
	}

	if !generatorCalled {
		t.Errorf("il generatore PDF doveva essere invocato")
	}

	if len(pdfBytes) == 0 {
		t.Errorf("byte PDF vuoti")
	}

	if filename == "" {
		t.Errorf("nome file non deve essere vuoto")
	}
}

func TestReportService_GenerateScanPDFReport_Validation(t *testing.T) {
	svc := service.NewReportService(&reportMockScanRepo{}, &reportMockProfileRepo{}, &reportMockTargetRepo{}, &reportMockFindingRepo{}, &reportMockTenantRepo{}, &reportMockPDFGenerator{})

	// Nil tenant ID
	_, _, err := svc.GenerateScanPDFReport(context.Background(), uuid.Nil, uuid.New())
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("atteso ErrValidation per Nil tenantID, ottenuto: %v", err)
	}

	// Nil scan ID
	_, _, err = svc.GenerateScanPDFReport(context.Background(), uuid.New(), uuid.Nil)
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

	svc := service.NewReportService(scanRepo, &reportMockProfileRepo{}, &reportMockTargetRepo{}, &reportMockFindingRepo{}, &reportMockTenantRepo{}, &reportMockPDFGenerator{})

	_, _, err := svc.GenerateScanPDFReport(context.Background(), uuid.New(), uuid.New())
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

	svc := service.NewReportService(scanRepo, &reportMockProfileRepo{}, &reportMockTargetRepo{}, &reportMockFindingRepo{}, tenantRepo, &reportMockPDFGenerator{})

	_, _, err := svc.GenerateScanPDFReport(context.Background(), uuid.New(), uuid.New())
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

	svc := service.NewReportService(scanRepo, &reportMockProfileRepo{}, &reportMockTargetRepo{}, findingRepo, &reportMockTenantRepo{}, &reportMockPDFGenerator{})

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
	svc := service.NewReportService(&reportMockScanRepo{}, &reportMockProfileRepo{}, &reportMockTargetRepo{}, &reportMockFindingRepo{}, &reportMockTenantRepo{}, &reportMockPDFGenerator{})

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

	svc := service.NewReportService(scanRepo, &reportMockProfileRepo{}, &reportMockTargetRepo{}, &reportMockFindingRepo{}, &reportMockTenantRepo{}, &reportMockPDFGenerator{})

	_, _, err := svc.GenerateScanCSVReport(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("atteso ErrNotFound, ottenuto: %v", err)
	}
}
