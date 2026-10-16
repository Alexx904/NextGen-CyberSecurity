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

type mockScanGroupRepo struct {
	groups map[uuid.UUID]*domain.TargetGroup
}

func newMockScanGroupRepo() *mockScanGroupRepo {
	return &mockScanGroupRepo{groups: make(map[uuid.UUID]*domain.TargetGroup)}
}

func (m *mockScanGroupRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.TargetGroup, error) {
	g, ok := m.groups[id]
	if !ok || g.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	return g, nil
}

func (m *mockScanGroupRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroup, error) {
	var list []*domain.TargetGroup
	for _, g := range m.groups {
		if g.TenantID == tenantID {
			list = append(list, g)
		}
	}
	return list, nil
}

func (m *mockScanGroupRepo) Create(ctx context.Context, group *domain.TargetGroup) error {
	m.groups[group.ID] = group
	return nil
}

func (m *mockScanGroupRepo) Update(ctx context.Context, group *domain.TargetGroup) error {
	m.groups[group.ID] = group
	return nil
}

func (m *mockScanGroupRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	delete(m.groups, id)
	return nil
}

func (m *mockScanGroupRepo) GetTargetCountsByGroup(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]int, error) {
	return make(map[uuid.UUID]int), nil
}

func (m *mockScanGroupRepo) GetSeverityCountsByGroup(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]map[domain.SeverityLevel]int, error) {
	return make(map[uuid.UUID]map[domain.SeverityLevel]int), nil
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
	svc := NewScanService(scanRepo, &mockProfileRepo{}, &mockTargetRepo{}, newMockFindingRepo(), newMockScanGroupRepo(), orchClient)

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
	svc := NewScanService(scanRepo, &mockProfileRepo{}, &mockTargetRepo{}, newMockFindingRepo(), newMockScanGroupRepo(), orchClient)

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
	svc := NewScanService(scanRepo, &mockProfileRepo{}, &mockTargetRepo{}, newMockFindingRepo(), newMockScanGroupRepo(), orchClient)

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

	svc := NewScanService(scanRepo, profileRepo, targetRepo, newMockFindingRepo(), newMockScanGroupRepo(), orchClient)

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

	svc := NewScanService(scanRepo, profileRepo, targetRepo, newMockFindingRepo(), newMockScanGroupRepo(), orchClient)

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

	svc := NewScanService(scanRepo, profileRepo, targetRepo, newMockFindingRepo(), newMockScanGroupRepo(), orchClient)

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

	svc := NewScanService(scanRepo, &mockProfileRepo{}, &mockTargetRepo{}, findingRepo, newMockScanGroupRepo(), &mockOrchestratorClient{})

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
	svc := NewScanService(scanRepo, &mockProfileRepo{}, &mockTargetRepo{}, newMockFindingRepo(), newMockScanGroupRepo(), &mockOrchestratorClient{})

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

	svc := NewScanService(newFullMockScanRepo(), profileRepo, &mockTargetRepo{}, newMockFindingRepo(), newMockScanGroupRepo(), &mockOrchestratorClient{})

	profiles, err := svc.ListProfiles(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("ListProfiles ha fallito: %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("attesi 2 profili, trovati %d", len(profiles))
	}
}

func TestScanService_CustomProfileInvasiveness(t *testing.T) {
	for _, tc := range []struct {
		args string
		want domain.InvasivenessLevel
	}{
		{"-sS -sV -T3", domain.InvasivenessSafe},
		{"-sV --script vuln -T3", domain.InvasivenessIntrusive},
		{"-sA -Pn -T3", domain.InvasivenessIntrusive},
		{"-sS -p- -T3", domain.InvasivenessIntrusive},
		{"-sV -T5", domain.InvasivenessIntrusive},
	} {
		t.Run(tc.args, func(t *testing.T) {
			repo := &mockProfileRepo{}
			svc := NewScanService(newFullMockScanRepo(), repo, &mockTargetRepo{}, newMockFindingRepo(), newMockScanGroupRepo(), &mockOrchestratorClient{})
			profile, err := svc.CreateCustomProfile(context.Background(), uuid.New(), "Custom Nmap", tc.args, nil, 1800, domain.ProfileCategoryCustom)
			if err != nil {
				t.Fatal(err)
			}
			if profile.Invasiveness != tc.want || repo.profiles[profile.ID].Invasiveness != tc.want {
				t.Fatalf("invasiveness not correctly persisted: %+v", profile)
			}
		})
	}
}

func TestScanService_StartGroupScan_RecursiveTargets(t *testing.T) {
	tenantID := uuid.New()
	profileID := uuid.New()

	// Crea albero cartelle: RootFolder -> SubFolder
	rootGroupID := uuid.New()
	subGroupID := uuid.New()

	groupRepo := newMockScanGroupRepo()
	_ = groupRepo.Create(context.Background(), &domain.TargetGroup{
		ID:       rootGroupID,
		TenantID: tenantID,
		Name:     "Root Folder",
	})
	_ = groupRepo.Create(context.Background(), &domain.TargetGroup{
		ID:       subGroupID,
		TenantID: tenantID,
		ParentID: &rootGroupID,
		Name:     "Sub Folder",
	})

	// Target 1 in root, Target 2 in subfolder, Target 3 in other folder
	t1ID := uuid.New()
	t2ID := uuid.New()
	otherGroupID := uuid.New()
	t3ID := uuid.New()

	targetRepo := &mockTargetRepo{
		targets: map[uuid.UUID]*domain.Target{
			t1ID: {ID: t1ID, TenantID: tenantID, GroupID: &rootGroupID, Value: "10.0.1.1"},
			t2ID: {ID: t2ID, TenantID: tenantID, GroupID: &subGroupID, Value: "10.0.2.1"},
			t3ID: {ID: t3ID, TenantID: tenantID, GroupID: &otherGroupID, Value: "10.0.3.1"},
		},
		blacklistedMap: make(map[string]bool),
	}

	profileRepo := &mockProfileRepo{
		profiles: map[uuid.UUID]*domain.ScanProfile{
			profileID: {ID: profileID, Name: "Standard"},
		},
	}

	scanRepo := newFullMockScanRepo()
	orchClient := &mockOrchestratorClient{}

	svc := NewScanService(scanRepo, profileRepo, targetRepo, newMockFindingRepo(), groupRepo, orchClient)

	scans, err := svc.StartGroupScan(context.Background(), tenantID, rootGroupID, profileID)
	if err != nil {
		t.Fatalf("StartGroupScan ha fallito inaspettatamente: %v", err)
	}

	// Devono essere avviate 2 scansioni (t1 e t2, escluso t3)
	if len(scans) != 2 {
		t.Fatalf("Attese 2 scansioni avviate, ottenute %d", len(scans))
	}
	if len(orchClient.enqueuedScans) != 2 {
		t.Errorf("Attese 2 scansioni accodate all'orchestrator, ottenute %d", len(orchClient.enqueuedScans))
	}
}

func TestScanService_StartGroupScan_SkipsBlacklistedAndActive(t *testing.T) {
	tenantID := uuid.New()
	profileID := uuid.New()
	groupID := uuid.New()

	groupRepo := newMockScanGroupRepo()
	_ = groupRepo.Create(context.Background(), &domain.TargetGroup{
		ID:       groupID,
		TenantID: tenantID,
		Name:     "Folder",
	})

	t1ID := uuid.New() // blacklisted
	t2ID := uuid.New() // already active
	t3ID := uuid.New() // valid

	targetRepo := &mockTargetRepo{
		targets: map[uuid.UUID]*domain.Target{
			t1ID: {ID: t1ID, TenantID: tenantID, GroupID: &groupID, Value: "10.0.1.1", IsBlacklisted: true},
			t2ID: {ID: t2ID, TenantID: tenantID, GroupID: &groupID, Value: "10.0.1.2"},
			t3ID: {ID: t3ID, TenantID: tenantID, GroupID: &groupID, Value: "10.0.1.3"},
		},
		blacklistedMap: map[string]bool{"10.0.1.1": true},
	}

	scanRepo := newFullMockScanRepo()
	// Crea scansione già attiva per t2
	_ = scanRepo.Create(context.Background(), &domain.Scan{
		ID:        uuid.New(),
		TenantID:  tenantID,
		TargetID:  t2ID,
		ProfileID: profileID,
		Status:    domain.ScanStatusRunning,
	})

	profileRepo := &mockProfileRepo{
		profiles: map[uuid.UUID]*domain.ScanProfile{
			profileID: {ID: profileID, Name: "Standard"},
		},
	}
	orchClient := &mockOrchestratorClient{}

	svc := NewScanService(scanRepo, profileRepo, targetRepo, newMockFindingRepo(), groupRepo, orchClient)

	scans, err := svc.StartGroupScan(context.Background(), tenantID, groupID, profileID)
	if err != nil {
		t.Fatalf("StartGroupScan ha fallito: %v", err)
	}

	// Solo t3 deve essere scansionato
	if len(scans) != 1 {
		t.Fatalf("Attesa 1 sola scansione per il target valido, ottenute %d", len(scans))
	}
	if scans[0].TargetID != t3ID {
		t.Errorf("Target ID atteso %v, ottenuto %v", t3ID, scans[0].TargetID)
	}
}

func TestScanService_GetScanDiff_NoPreviousScan(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()
	scanID := uuid.New()

	scanRepo := newFullMockScanRepo()
	_ = scanRepo.Create(context.Background(), &domain.Scan{
		ID:       scanID,
		TenantID: tenantID,
		TargetID: targetID,
		Status:   domain.ScanStatusCompleted,
	})

	findingRepo := newMockFindingRepo()
	fID := uuid.New()
	svcName := "http"
	cve := "CVE-2023-1234"
	findingRepo.findings[fID] = &domain.Finding{
		ID:          fID,
		ScanID:      scanID,
		TenantID:    tenantID,
		HostIP:      "192.168.1.10",
		Port:        80,
		Protocol:    "tcp",
		ServiceName: &svcName,
		CVEID:       &cve,
		Severity:    domain.SeverityHigh,
	}

	targetRepo := &mockTargetRepo{
		targets: map[uuid.UUID]*domain.Target{
			targetID: {ID: targetID, TenantID: tenantID, Value: "192.168.1.10"},
		},
	}

	svc := NewScanService(scanRepo, &mockProfileRepo{}, targetRepo, findingRepo, newMockScanGroupRepo(), &mockOrchestratorClient{})

	diff, err := svc.GetScanDiff(context.Background(), tenantID, scanID)
	if err != nil {
		t.Fatalf("GetScanDiff ha restituito un errore inatteso: %v", err)
	}

	if diff.PreviousScanID != nil {
		t.Errorf("Atteso PreviousScanID nil per prima scansione, ottenuto %v", diff.PreviousScanID)
	}
	if diff.TotalNew != 1 {
		t.Errorf("Atteso TotalNew = 1, ottenuto %d", diff.TotalNew)
	}
	if diff.TotalResolved != 0 || diff.TotalPersistent != 0 {
		t.Errorf("Attesi 0 resolved e 0 persistent, ottenuti res=%d, pers=%d", diff.TotalResolved, diff.TotalPersistent)
	}
	if len(diff.Items) != 1 || diff.Items[0].Status != domain.DiffStatusNew {
		t.Errorf("Item atteso con stato NEW, ottenuto: %+v", diff.Items)
	}
}

func TestScanService_GetScanDiff_WithPreviousScan_NewPersistentResolved(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()
	prevScanID := uuid.New()
	currScanID := uuid.New()

	scanRepo := newFullMockScanRepo()
	now := time.Now()
	prevTime := now.Add(-2 * time.Hour)
	_ = scanRepo.Create(context.Background(), &domain.Scan{
		ID:          prevScanID,
		TenantID:    tenantID,
		TargetID:    targetID,
		Status:      domain.ScanStatusCompleted,
		CompletedAt: &prevTime,
	})
	_ = scanRepo.Create(context.Background(), &domain.Scan{
		ID:          currScanID,
		TenantID:    tenantID,
		TargetID:    targetID,
		Status:      domain.ScanStatusCompleted,
		CompletedAt: &now,
	})

	findingRepo := newMockFindingRepo()
	sshSvc := "ssh"
	cveResolved := "CVE-2021-0001"
	cvePersistent := "CVE-2022-0002"
	cveNew := "CVE-2023-0003"

	// Finding in prevScan:
	// 1. CVE-2021-0001 (che sarà RISOLTA)
	f1ID := uuid.New()
	findingRepo.findings[f1ID] = &domain.Finding{
		ID:          f1ID,
		ScanID:      prevScanID,
		TenantID:    tenantID,
		HostIP:      "192.168.1.10",
		Port:        22,
		Protocol:    "tcp",
		ServiceName: &sshSvc,
		CVEID:       &cveResolved,
		Severity:    domain.SeverityMedium,
	}
	// 2. CVE-2022-0002 (che sarà PERSISTENTE)
	f2ID := uuid.New()
	findingRepo.findings[f2ID] = &domain.Finding{
		ID:       f2ID,
		ScanID:   prevScanID,
		TenantID: tenantID,
		HostIP:   "192.168.1.10",
		Port:     80,
		Protocol: "tcp",
		CVEID:    &cvePersistent,
		Severity: domain.SeverityHigh,
	}

	// Finding in currScan:
	// 1. CVE-2022-0002 (PERSISTENTE)
	f3ID := uuid.New()
	findingRepo.findings[f3ID] = &domain.Finding{
		ID:       f3ID,
		ScanID:   currScanID,
		TenantID: tenantID,
		HostIP:   "192.168.1.10",
		Port:     80,
		Protocol: "tcp",
		CVEID:    &cvePersistent,
		Severity: domain.SeverityHigh,
	}
	// 2. CVE-2023-0003 (NUOVA)
	f4ID := uuid.New()
	findingRepo.findings[f4ID] = &domain.Finding{
		ID:       f4ID,
		ScanID:   currScanID,
		TenantID: tenantID,
		HostIP:   "192.168.1.10",
		Port:     443,
		Protocol: "tcp",
		CVEID:    &cveNew,
		Severity: domain.SeverityCritical,
	}

	targetRepo := &mockTargetRepo{
		targets: map[uuid.UUID]*domain.Target{
			targetID: {ID: targetID, TenantID: tenantID, Value: "192.168.1.10"},
		},
	}

	svc := NewScanService(scanRepo, &mockProfileRepo{}, targetRepo, findingRepo, newMockScanGroupRepo(), &mockOrchestratorClient{})

	diff, err := svc.GetScanDiff(context.Background(), tenantID, currScanID)
	if err != nil {
		t.Fatalf("GetScanDiff ha restituito un errore inatteso: %v", err)
	}

	if diff.PreviousScanID == nil || *diff.PreviousScanID != prevScanID {
		t.Errorf("Atteso PreviousScanID %v, ottenuto %v", prevScanID, diff.PreviousScanID)
	}

	if diff.TotalNew != 1 {
		t.Errorf("Atteso TotalNew = 1, ottenuto %d", diff.TotalNew)
	}
	if diff.TotalPersistent != 1 {
		t.Errorf("Atteso TotalPersistent = 1, ottenuto %d", diff.TotalPersistent)
	}
	if diff.TotalResolved != 1 {
		t.Errorf("Atteso TotalResolved = 1, ottenuto %d", diff.TotalResolved)
	}

	statusCount := make(map[domain.ScanDiffStatus]int)
	for _, it := range diff.Items {
		statusCount[it.Status]++
	}
	if statusCount[domain.DiffStatusNew] != 1 || statusCount[domain.DiffStatusPersistent] != 1 || statusCount[domain.DiffStatusResolved] != 1 {
		t.Errorf("Conteggio item per stato errato: %+v", statusCount)
	}
}
