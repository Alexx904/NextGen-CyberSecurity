package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
)

type mockOrchestratorClient struct {
	mu            sync.Mutex
	enqueuedScans []uuid.UUID
	canceledScans []uuid.UUID
	enqueueErr    error
	cancelErr     error
}

func (m *mockOrchestratorClient) Enqueue(tenantID, scanID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.enqueueErr != nil {
		return m.enqueueErr
	}
	m.enqueuedScans = append(m.enqueuedScans, scanID)
	return nil
}

func (m *mockOrchestratorClient) CancelScan(tenantID, scanID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancelErr != nil {
		return m.cancelErr
	}
	m.canceledScans = append(m.canceledScans, scanID)
	return nil
}

type fullMockScanRepo struct {
	*mockScanRepo
	activeScans map[uuid.UUID]*domain.Scan
}

func newFullMockScanRepo() *fullMockScanRepo {
	return &fullMockScanRepo{
		mockScanRepo: newMockScanRepo(),
		activeScans:  make(map[uuid.UUID]*domain.Scan),
	}
}

func (m *fullMockScanRepo) GetActiveByTarget(ctx context.Context, tenantID, targetID uuid.UUID) (*domain.Scan, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, s := range m.scans {
		if s.TenantID == tenantID && s.TargetID == targetID && (s.Status == domain.ScanStatusQueued || s.Status == domain.ScanStatusRunning) {
			cp := *s
			return &cp, nil
		}
	}
	return nil, nil
}

type mockFindingRepo struct {
	findings map[uuid.UUID]*domain.Finding
}

func newMockFindingRepo() *mockFindingRepo {
	return &mockFindingRepo{findings: make(map[uuid.UUID]*domain.Finding)}
}

func (m *mockFindingRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Finding, error) {
	f, ok := m.findings[id]
	if !ok || f.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	return f, nil
}

func (m *mockFindingRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.FindingFilter) ([]*domain.Finding, int, error) {
	var list []*domain.Finding
	for _, f := range m.findings {
		if f.TenantID == tenantID {
			if filter.ScanID != nil && f.ScanID != *filter.ScanID {
				continue
			}
			list = append(list, f)
		}
	}
	return list, len(list), nil
}

func (m *mockFindingRepo) CreateBatch(ctx context.Context, findings []*domain.Finding) error {
	for _, f := range findings {
		m.findings[f.ID] = f
	}
	return nil
}

func (m *mockFindingRepo) GetSeverityCountsByScan(ctx context.Context, tenantID, scanID uuid.UUID) (map[domain.SeverityLevel]int, error) {
	counts := make(map[domain.SeverityLevel]int)
	for _, f := range m.findings {
		if f.TenantID == tenantID && f.ScanID == scanID {
			counts[f.Severity]++
		}
	}
	return counts, nil
}

func TestScanService_CancelScan_Success(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	scanRepo := newFullMockScanRepo()
	scan := &domain.Scan{
		ID:       scanID,
		TenantID: tenantID,
		Status:   domain.ScanStatusRunning,
	}
	_ = scanRepo.Create(context.Background(), scan)

	orchClient := &mockOrchestratorClient{}
	svc := NewScanService(scanRepo, &mockProfileRepo{}, &mockTargetRepo{}, newMockFindingRepo(), orchClient)

	err := svc.CancelScan(context.Background(), tenantID, scanID)
	if err != nil {
		t.Fatalf("CancelScan ha restituito un errore inatteso: %v", err)
	}

	if len(orchClient.canceledScans) != 1 || orchClient.canceledScans[0] != scanID {
		t.Errorf("orchestrator non è stato invocato con scanID corretto: %v", orchClient.canceledScans)
	}
}

func TestScanService_CancelScan_NotFound(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	scanRepo := newFullMockScanRepo()
	orchClient := &mockOrchestratorClient{}
	svc := NewScanService(scanRepo, &mockProfileRepo{}, &mockTargetRepo{}, newMockFindingRepo(), orchClient)

	err := svc.CancelScan(context.Background(), tenantID, scanID)
	if err == nil || !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("atteso ErrNotFound, ottenuto: %v", err)
	}
}

func TestScanService_CancelScan_CannotCancelTerminal(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	scanRepo := newFullMockScanRepo()
	scan := &domain.Scan{
		ID:       scanID,
		TenantID: tenantID,
		Status:   domain.ScanStatusCompleted,
	}
	_ = scanRepo.Create(context.Background(), scan)

	orchClient := &mockOrchestratorClient{}
	svc := NewScanService(scanRepo, &mockProfileRepo{}, &mockTargetRepo{}, newMockFindingRepo(), orchClient)

	err := svc.CancelScan(context.Background(), tenantID, scanID)
	if err == nil || !errors.Is(err, domain.ErrScanCannotBeCancelled) {
		t.Errorf("atteso ErrScanCannotBeCancelled, ottenuto: %v", err)
	}
}

func TestScanService_StartScan_Success(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()
	profileID := uuid.New()

	scanRepo := newFullMockScanRepo()
	targetRepo := &mockTargetRepo{
		targets: map[uuid.UUID]*domain.Target{
			targetID: {
				ID:            targetID,
				TenantID:      tenantID,
				Value:         "192.168.1.1",
				IsBlacklisted: false,
			},
		},
		blacklistedMap: make(map[string]bool),
	}
	profileRepo := &mockProfileRepo{
		profiles: map[uuid.UUID]*domain.ScanProfile{
			profileID: {ID: profileID, Name: "Standard", NmapArgs: "-sV -T4"},
		},
	}
	orchClient := &mockOrchestratorClient{}

	svc := NewScanService(scanRepo, profileRepo, targetRepo, newMockFindingRepo(), orchClient)

	scan, err := svc.StartScan(context.Background(), tenantID, targetID, profileID)
	if err != nil {
		t.Fatalf("StartScan ha fallito inaspettatamente: %v", err)
	}

	if scan == nil || scan.Status != domain.ScanStatusQueued {
		t.Fatalf("scansione non valida o non in stato QUEUED: %+v", scan)
	}

	if len(orchClient.enqueuedScans) != 1 || orchClient.enqueuedScans[0] != scan.ID {
		t.Errorf("scansione non accodata nell'orchestratore: %v", orchClient.enqueuedScans)
	}
}

func TestScanService_StartScan_BlacklistedTarget(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()
	profileID := uuid.New()

	scanRepo := newFullMockScanRepo()
	targetRepo := &mockTargetRepo{
		targets: map[uuid.UUID]*domain.Target{
			targetID: {
				ID:            targetID,
				TenantID:      tenantID,
				Value:         "10.0.0.1",
				IsBlacklisted: true,
			},
		},
		blacklistedMap: map[string]bool{"10.0.0.1": true},
	}
	profileRepo := &mockProfileRepo{
		profiles: map[uuid.UUID]*domain.ScanProfile{
			profileID: {ID: profileID, Name: "Standard"},
		},
	}
	orchClient := &mockOrchestratorClient{}

	svc := NewScanService(scanRepo, profileRepo, targetRepo, newMockFindingRepo(), orchClient)

	_, err := svc.StartScan(context.Background(), tenantID, targetID, profileID)
	if err == nil || !errors.Is(err, domain.ErrTargetBlacklisted) {
		t.Errorf("atteso ErrTargetBlacklisted, ottenuto: %v", err)
	}
}

func TestScanService_StartScan_AlreadyRunning(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()
	profileID := uuid.New()

	scanRepo := newFullMockScanRepo()
	existingScan := &domain.Scan{
		ID:        uuid.New(),
		TenantID:  tenantID,
		TargetID:  targetID,
		ProfileID: profileID,
		Status:    domain.ScanStatusRunning,
	}
	_ = scanRepo.Create(context.Background(), existingScan)

	targetRepo := &mockTargetRepo{
		targets: map[uuid.UUID]*domain.Target{
			targetID: {
				ID:            targetID,
				TenantID:      tenantID,
				Value:         "192.168.1.1",
				IsBlacklisted: false,
			},
		},
		blacklistedMap: make(map[string]bool),
	}
	profileRepo := &mockProfileRepo{
		profiles: map[uuid.UUID]*domain.ScanProfile{
			profileID: {ID: profileID, Name: "Standard"},
		},
	}
	orchClient := &mockOrchestratorClient{}

	svc := NewScanService(scanRepo, profileRepo, targetRepo, newMockFindingRepo(), orchClient)

	_, err := svc.StartScan(context.Background(), tenantID, targetID, profileID)
	if err == nil || !errors.Is(err, domain.ErrScanAlreadyRunning) {
		t.Errorf("atteso ErrScanAlreadyRunning, ottenuto: %v", err)
	}
}

func TestScanService_GetScanFindings_Success(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	scanRepo := newFullMockScanRepo()
	scan := &domain.Scan{ID: scanID, TenantID: tenantID, Status: domain.ScanStatusCompleted}
	_ = scanRepo.Create(context.Background(), scan)

	findingRepo := newMockFindingRepo()
	fID := uuid.New()
	_ = findingRepo.CreateBatch(context.Background(), []*domain.Finding{
		{ID: fID, ScanID: scanID, TenantID: tenantID, HostIP: "192.168.1.50", Port: 80, Severity: domain.SeverityInfo},
	})

	svc := NewScanService(scanRepo, &mockProfileRepo{}, &mockTargetRepo{}, findingRepo, &mockOrchestratorClient{})

	findings, total, err := svc.GetScanFindings(context.Background(), tenantID, scanID, domain.FindingFilter{})
	if err != nil {
		t.Fatalf("GetScanFindings ha fallito: %v", err)
	}

	if total != 1 || len(findings) != 1 || findings[0].Port != 80 {
		t.Errorf("findings non corrispondenti: total=%d, items=%+v", total, findings)
	}
}

func TestScanService_GetScanFindings_ScanNotFound(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	scanRepo := newFullMockScanRepo()
	svc := NewScanService(scanRepo, &mockProfileRepo{}, &mockTargetRepo{}, newMockFindingRepo(), &mockOrchestratorClient{})

	_, _, err := svc.GetScanFindings(context.Background(), tenantID, scanID, domain.FindingFilter{})
	if err == nil || !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("atteso ErrNotFound per scan inesistente, ottenuto: %v", err)
	}
}

func TestScanService_ListProfiles(t *testing.T) {
	profileRepo := &mockProfileRepo{
		profiles: map[uuid.UUID]*domain.ScanProfile{
			uuid.New(): {Name: "Fast Discovery"},
			uuid.New(): {Name: "Standard Service Scan"},
		},
	}

	svc := NewScanService(newFullMockScanRepo(), profileRepo, &mockTargetRepo{}, newMockFindingRepo(), &mockOrchestratorClient{})

	profiles, err := svc.ListProfiles(context.Background())
	if err != nil {
		t.Fatalf("ListProfiles ha fallito: %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("attesi 2 profili, trovati %d", len(profiles))
	}
}
