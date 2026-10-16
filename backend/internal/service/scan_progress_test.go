package service

import (
	"context"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/scanner"
	"testing"
	"time"
)

func TestTotalProgressEstimate(t *testing.T) {
	pct, remaining := 50.0, 60.0
	phase, total, eta := estimateScanProgress(scanner.TaskProgress{Task: "SYN Stealth Scan", Percent: &pct, Remaining: &remaining}, 60, 0)
	if phase != "PORT_SCAN" || total != 35 || eta == nil || *eta < remaining {
		t.Fatalf("unexpected: %s %v %v", phase, total, eta)
	}
	_, next, _ := estimateScanProgress(scanner.TaskProgress{Task: "Ping Scan", Percent: &pct}, 70, total)
	if next < total {
		t.Fatal("progress regressed across tasks/batches")
	}
	_, _, eta = estimateScanProgress(scanner.TaskProgress{Task: "SYN Stealth Scan", Percent: &pct}, 1, 0)
	if eta != nil {
		t.Fatal("premature ETA")
	}
	phase, _, _ = estimateScanProgress(scanner.TaskProgress{Task: "Unknown", Percent: &pct}, 30, 0)
	if phase != "" {
		t.Fatal("unknown task misclassified")
	}
}

type progressOrchestratorClient struct {
	mockOrchestratorClient
	snapshot *domain.ScanProgressPayload
}

func (c *progressOrchestratorClient) GetProgress(id uuid.UUID) *domain.ScanProgressPayload {
	return c.snapshot
}

func TestScanProgressSnapshotTenantAndCleanup(t *testing.T) {
	tenant, id := uuid.New(), uuid.New()
	repo := newFullMockScanRepo()
	repo.Create(context.Background(), &domain.Scan{ID: id, TenantID: tenant, Status: domain.ScanStatusRunning})
	client := &progressOrchestratorClient{snapshot: &domain.ScanProgressPayload{ScanID: id, Percentage: 42, ObservedAt: time.Now()}}
	svc := NewScanService(repo, &mockProfileRepo{}, &mockTargetRepo{}, newMockFindingRepo(), newMockScanGroupRepo(), client)
	scan, err := svc.GetScan(context.Background(), tenant, id)
	if err != nil || scan.Progress == nil || scan.Progress.Percentage != 42 {
		t.Fatal("snapshot missing", err)
	}
	if _, err := svc.GetScan(context.Background(), uuid.New(), id); err == nil {
		t.Fatal("cross-tenant access allowed")
	}
	orchestrator := NewScanOrchestrator(DefaultOrchestratorConfig(), repo, &mockProfileRepo{}, &mockTargetRepo{}, nil, nil)
	orchestrator.progressByScan[id] = *client.snapshot
	if orchestrator.GetProgress(id) == nil {
		t.Fatal("snapshot not found")
	}
	orchestrator.publishEvent("", domain.SSEEvent{Data: domain.ScanPhasePayload{ScanID: id, Phase: "CVE_CORRELATION", Percentage: 92}})
	if progress := orchestrator.GetProgress(id); progress.Percentage != 92 || progress.RemainingSeconds != nil {
		t.Fatal("phase did not invalidate stale ETA")
	}
	orchestrator.unregisterActiveScan(id)
	if orchestrator.GetProgress(id) != nil {
		t.Fatal("snapshot leaked")
	}
}
