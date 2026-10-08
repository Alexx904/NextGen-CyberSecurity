package service

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/scanner"
)

// ============================================================================
// MOCKS PER TEST UNITARI
// ============================================================================

type mockScanRepo struct {
	mu    sync.RWMutex
	scans map[uuid.UUID]*domain.Scan
}

func newMockScanRepo() *mockScanRepo {
	return &mockScanRepo{scans: make(map[uuid.UUID]*domain.Scan)}
}

func (m *mockScanRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Scan, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.scans[id]
	if !ok || s.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *mockScanRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Scan, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []*domain.Scan
	for _, s := range m.scans {
		if s.TenantID == tenantID {
			cp := *s
			res = append(res, &cp)
		}
	}
	return res, len(res), nil
}

func (m *mockScanRepo) Create(ctx context.Context, scan *domain.Scan) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *scan
	m.scans[scan.ID] = &cp
	return nil
}

func (m *mockScanRepo) GetActiveByTarget(ctx context.Context, tenantID, targetID uuid.UUID) (*domain.Scan, error) {
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

func (m *mockScanRepo) Update(ctx context.Context, scan *domain.Scan) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.scans[scan.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *scan
	m.scans[scan.ID] = &cp
	return nil
}

type mockProfileRepo struct {
	profiles map[uuid.UUID]*domain.ScanProfile
}

func (m *mockProfileRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ScanProfile, error) {
	p, ok := m.profiles[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (m *mockProfileRepo) GetByName(ctx context.Context, name string) (*domain.ScanProfile, error) {
	for _, p := range m.profiles {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockProfileRepo) List(ctx context.Context) ([]*domain.ScanProfile, error) {
	var list []*domain.ScanProfile
	for _, p := range m.profiles {
		list = append(list, p)
	}
	return list, nil
}

type mockTargetRepo struct {
	targets        map[uuid.UUID]*domain.Target
	blacklistedMap map[string]bool
}

func (m *mockTargetRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Target, error) {
	t, ok := m.targets[id]
	if !ok || t.TenantID != tenantID {
		return nil, domain.ErrNotFound
	}
	return t, nil
}

func (m *mockTargetRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.TargetFilter) ([]*domain.Target, int, error) {
	return nil, 0, nil
}

func (m *mockTargetRepo) Create(ctx context.Context, target *domain.Target) error {
	m.targets[target.ID] = target
	return nil
}

func (m *mockTargetRepo) Update(ctx context.Context, target *domain.Target) error {
	m.targets[target.ID] = target
	return nil
}

func (m *mockTargetRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	delete(m.targets, id)
	return nil
}

func (m *mockTargetRepo) IsBlacklisted(ctx context.Context, tenantID uuid.UUID, value string) (bool, error) {
	return m.blacklistedMap[value], nil
}

type mockRunner struct {
	runFunc func(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error)
}

func (m *mockRunner) RunScan(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, target, profile)
	}
	return &scanner.ScanExecutionResult{
		RawXML:   []byte(`<?xml version="1.0"?><nmaprun scanner="nmap" version="7.94"><host><status state="up"/><address addr="` + target + `" addrtype="ipv4"/><ports><port protocol="tcp" portid="80"><state state="open"/><service name="http"/></port></ports></host><runstats><finished elapsed="2.5"/><hosts up="1" down="0" total="1"/></runstats></nmaprun>`),
		ExitCode: 0,
		Duration: 2 * time.Second,
	}, nil
}

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// ============================================================================
// SUITE DI TEST
// ============================================================================

func TestScanOrchestrator_Lifecycle_Success(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()
	profileID := uuid.New()
	scanID := uuid.New()

	scanRepo := newMockScanRepo()
	targetRepo := &mockTargetRepo{
		targets: map[uuid.UUID]*domain.Target{
			targetID: {
				ID:            targetID,
				TenantID:      tenantID,
				Type:          domain.TargetTypeIP,
				Value:         "192.168.1.50",
				IsBlacklisted: false,
			},
		},
		blacklistedMap: make(map[string]bool),
	}

	profileRepo := &mockProfileRepo{
		profiles: map[uuid.UUID]*domain.ScanProfile{
			profileID: {
				ID:             profileID,
				Name:           "Fast Discovery",
				NmapArgs:       "-T4 -F",
				TimeoutSeconds: 10,
			},
		},
	}

	runner := &mockRunner{}

	initialScan := &domain.Scan{
		ID:          scanID,
		TenantID:    tenantID,
		TargetID:    targetID,
		ProfileID:   profileID,
		TriggerType: domain.TriggerManual,
		Status:      domain.ScanStatusQueued,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_ = scanRepo.Create(context.Background(), initialScan)

	orchestrator := NewScanOrchestrator(
		OrchestratorConfig{MaxWorkers: 1, QueueSize: 10},
		scanRepo,
		profileRepo,
		targetRepo,
		runner,
		silentLogger(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orchestrator.Start(ctx)
	defer orchestrator.Stop()

	// Accodamento del job
	if err := orchestrator.Enqueue(tenantID, scanID); err != nil {
		t.Fatalf("Enqueue ha restituito errore: %v", err)
	}

	// Polling attesa completamento scansione
	var finalScan *domain.Scan
	for i := 0; i < 30; i++ {
		s, err := scanRepo.GetByID(context.Background(), tenantID, scanID)
		if err == nil && s.Status.IsTerminal() {
			finalScan = s
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if finalScan == nil {
		t.Fatal("timeout in attesa del completamento della scansione")
	}

	if finalScan.Status != domain.ScanStatusCompleted {
		t.Errorf("atteso stato COMPLETED, trovato %s (errore: %v)", finalScan.Status, finalScan.ErrorMessage)
	}

	if finalScan.StartedAt == nil || finalScan.CompletedAt == nil {
		t.Error("timestamp StartedAt o CompletedAt mancanti")
	}

	if finalScan.SummaryStats.UpHosts != 1 || finalScan.SummaryStats.TotalOpenPorts != 1 {
		t.Errorf("summary stats errate: %+v", finalScan.SummaryStats)
	}
}

func TestScanOrchestrator_WorkerPoolConcurrency(t *testing.T) {
	tenantID := uuid.New()
	profileID := uuid.New()

	scanRepo := newMockScanRepo()
	targetRepo := &mockTargetRepo{
		targets:        make(map[uuid.UUID]*domain.Target),
		blacklistedMap: make(map[string]bool),
	}
	profileRepo := &mockProfileRepo{
		profiles: map[uuid.UUID]*domain.ScanProfile{
			profileID: {
				ID:             profileID,
				Name:           "Fast",
				NmapArgs:       "-T4 -F",
				TimeoutSeconds: 5,
			},
		},
	}

	var activeScans atomic.Int32
	var maxObservedActive atomic.Int32

	runner := &mockRunner{
		runFunc: func(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
			curr := activeScans.Add(1)
			for {
				max := maxObservedActive.Load()
				if curr <= max || maxObservedActive.CompareAndSwap(max, curr) {
					break
				}
			}

			// Simula durata operazione
			time.Sleep(100 * time.Millisecond)
			activeScans.Add(-1)

			return &scanner.ScanExecutionResult{
				RawXML:   []byte(`<?xml version="1.0"?><nmaprun scanner="nmap"><host><status state="up"/><address addr="` + target + `" addrtype="ipv4"/></host><runstats><finished elapsed="0.1"/><hosts up="1" down="0" total="1"/></runstats></nmaprun>`),
				ExitCode: 0,
			}, nil
		},
	}

	maxWorkers := 2
	orchestrator := NewScanOrchestrator(
		OrchestratorConfig{MaxWorkers: maxWorkers, QueueSize: 20},
		scanRepo,
		profileRepo,
		targetRepo,
		runner,
		silentLogger(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orchestrator.Start(ctx)
	defer orchestrator.Stop()

	// Creazione e accodamento di 6 scansioni
	numScans := 6
	scanIDs := make([]uuid.UUID, numScans)

	for i := 0; i < numScans; i++ {
		targetID := uuid.New()
		sID := uuid.New()
		scanIDs[i] = sID

		targetRepo.targets[targetID] = &domain.Target{
			ID:       targetID,
			TenantID: tenantID,
			Type:     domain.TargetTypeIP,
			Value:    "192.168.1.1",
		}

		s := &domain.Scan{
			ID:          sID,
			TenantID:    tenantID,
			TargetID:    targetID,
			ProfileID:   profileID,
			TriggerType: domain.TriggerManual,
			Status:      domain.ScanStatusQueued,
		}
		_ = scanRepo.Create(context.Background(), s)
		_ = orchestrator.Enqueue(tenantID, sID)
	}

	// Attesa completamento di tutti i job
	for _, sID := range scanIDs {
		for j := 0; j < 40; j++ {
			s, err := scanRepo.GetByID(context.Background(), tenantID, sID)
			if err == nil && s.Status.IsTerminal() {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Verifica che il picco di scansioni concorrenti non abbia mai superato maxWorkers
	if maxObservedActive.Load() > int32(maxWorkers) {
		t.Errorf("limite di concorrenza violato: atteso max %d, osservati %d", maxWorkers, maxObservedActive.Load())
	}
}

func TestScanOrchestrator_TargetBlacklisted(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()
	profileID := uuid.New()
	scanID := uuid.New()

	scanRepo := newMockScanRepo()
	targetRepo := &mockTargetRepo{
		targets: map[uuid.UUID]*domain.Target{
			targetID: {
				ID:            targetID,
				TenantID:      tenantID,
				Type:          domain.TargetTypeIP,
				Value:         "10.0.0.99",
				IsBlacklisted: true, // Target blacklisted
			},
		},
		blacklistedMap: map[string]bool{"10.0.0.99": true},
	}

	profileRepo := &mockProfileRepo{
		profiles: map[uuid.UUID]*domain.ScanProfile{
			profileID: {ID: profileID, Name: "Fast", NmapArgs: "-T4 -F", TimeoutSeconds: 5},
		},
	}

	runnerCalled := false
	runner := &mockRunner{
		runFunc: func(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
			runnerCalled = true
			return nil, nil
		},
	}

	scan := &domain.Scan{
		ID:          scanID,
		TenantID:    tenantID,
		TargetID:    targetID,
		ProfileID:   profileID,
		TriggerType: domain.TriggerManual,
		Status:      domain.ScanStatusQueued,
	}
	_ = scanRepo.Create(context.Background(), scan)

	orchestrator := NewScanOrchestrator(
		OrchestratorConfig{MaxWorkers: 1, QueueSize: 5},
		scanRepo,
		profileRepo,
		targetRepo,
		runner,
		silentLogger(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orchestrator.Start(ctx)
	defer orchestrator.Stop()

	_ = orchestrator.Enqueue(tenantID, scanID)

	// Attesa completamento
	var finalScan *domain.Scan
	for i := 0; i < 20; i++ {
		s, err := scanRepo.GetByID(context.Background(), tenantID, scanID)
		if err == nil && s.Status.IsTerminal() {
			finalScan = s
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if finalScan == nil || finalScan.Status != domain.ScanStatusFailed {
		t.Fatalf("atteso stato FAILED per target in blacklist, ottenuto: %+v", finalScan)
	}

	if runnerCalled {
		t.Error("il runner Nmap non doveva essere invocato per un target in blacklist")
	}
}

func TestScanOrchestrator_CancelScan(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()
	profileID := uuid.New()
	scanID := uuid.New()

	scanRepo := newMockScanRepo()
	targetRepo := &mockTargetRepo{
		targets: map[uuid.UUID]*domain.Target{
			targetID: {ID: targetID, TenantID: tenantID, Type: domain.TargetTypeIP, Value: "192.168.1.1"},
		},
		blacklistedMap: make(map[string]bool),
	}
	profileRepo := &mockProfileRepo{
		profiles: map[uuid.UUID]*domain.ScanProfile{
			profileID: {ID: profileID, Name: "Slow Scan", NmapArgs: "-T4", TimeoutSeconds: 60},
		},
	}

	scanStartedChan := make(chan struct{})

	runner := &mockRunner{
		runFunc: func(ctx context.Context, target string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
			close(scanStartedChan)
			// Attende l'annullamento del context
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}

	scan := &domain.Scan{
		ID:          scanID,
		TenantID:    tenantID,
		TargetID:    targetID,
		ProfileID:   profileID,
		TriggerType: domain.TriggerManual,
		Status:      domain.ScanStatusQueued,
	}
	_ = scanRepo.Create(context.Background(), scan)

	orchestrator := NewScanOrchestrator(
		OrchestratorConfig{MaxWorkers: 1, QueueSize: 5},
		scanRepo,
		profileRepo,
		targetRepo,
		runner,
		silentLogger(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orchestrator.Start(ctx)
	defer orchestrator.Stop()

	_ = orchestrator.Enqueue(tenantID, scanID)

	// Attende che il runner sia effettivamente in esecuzione
	select {
	case <-scanStartedChan:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout in attesa dell'avvio della scansione")
	}

	// Invocazione cancellazione
	if err := orchestrator.CancelScan(tenantID, scanID); err != nil {
		t.Fatalf("CancelScan ha restituito errore: %v", err)
	}

	// Attesa transizione a CANCELLED
	var finalScan *domain.Scan
	for i := 0; i < 20; i++ {
		s, err := scanRepo.GetByID(context.Background(), tenantID, scanID)
		if err == nil && s.Status == domain.ScanStatusCancelled {
			finalScan = s
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if finalScan == nil || finalScan.Status != domain.ScanStatusCancelled {
		t.Errorf("atteso stato CANCELLED, ottenuto: %+v", finalScan)
	}
}

func TestScanOrchestrator_GracefulStop(t *testing.T) {
	orchestrator := NewScanOrchestrator(
		OrchestratorConfig{MaxWorkers: 2, QueueSize: 5},
		newMockScanRepo(),
		&mockProfileRepo{},
		&mockTargetRepo{},
		&mockRunner{},
		silentLogger(),
	)

	ctx := context.Background()
	orchestrator.Start(ctx)

	// Verifica che lo stop avvenga senza deadlock o panic
	stopDone := make(chan struct{})
	go func() {
		orchestrator.Stop()
		close(stopDone)
	}()

	select {
	case <-stopDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout durante Graceful Stop dell'orchestratore")
	}
}
