package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/scanner"
)

type liveFindingRepo struct {
	*mockFindingRepo
	fail bool
}

func (r *liveFindingRepo) CreateBatch(ctx context.Context, findings []*domain.Finding) error {
	if r.fail {
		return errors.New("synthetic storage failure")
	}
	return r.mockFindingRepo.CreateBatch(ctx, findings)
}

func TestScanOrchestratorLivePersistence(t *testing.T) {
	for _, fail := range []bool{false, true} {
		name := "success"
		if fail {
			name = "database failure"
		}
		t.Run(name, func(t *testing.T) {
			tenant, targetID, profileID, scanID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
			scans := newMockScanRepo()
			_ = scans.Create(context.Background(), &domain.Scan{ID: scanID, TenantID: tenant, TargetID: targetID, ProfileID: profileID, Status: domain.ScanStatusQueued, TriggerType: domain.TriggerManual, CreatedAt: time.Now()})
			targets := &mockTargetRepo{targets: map[uuid.UUID]*domain.Target{targetID: {ID: targetID, TenantID: tenant, Type: domain.TargetTypeIP, Value: "192.168.1.10"}}, blacklistedMap: map[string]bool{}}
			profiles := &mockProfileRepo{profiles: map[uuid.UUID]*domain.ScanProfile{profileID: {ID: profileID, Name: "Fixture", NmapArgs: "-sV", TimeoutSeconds: 5}}}
			findings := &liveFindingRepo{mockFindingRepo: newMockFindingRepo(), fail: fail}
			xml := []byte(`<nmaprun><host><status state="up"/><address addr="192.168.1.10" addrtype="ipv4"/><ports><port protocol="tcp" portid="80"><state state="open"/></port></ports></host><runstats><finished elapsed="1"/><hosts up="1" total="1"/></runstats></nmaprun>`)
			runner := &mockRunner{runFunc: func(_ context.Context, _ string, p scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
				if p.OnHost == nil {
					t.Fatal("live persistence callback missing")
				}
				report, err := scanner.ParseNmapXMLBytes(xml)
				if err != nil {
					t.Fatal(err)
				}
				err = p.OnHost(report.Hosts[0])
				if fail {
					if err == nil {
						t.Fatal("storage failure did not stop scan")
					}
					return nil, err
				}
				if err != nil || len(findings.findings) != 1 {
					t.Fatalf("host not persisted while runner is active: %v", err)
				}
				return &scanner.ScanExecutionResult{RawXML: xml}, nil
			}}
			o := NewScanOrchestrator(OrchestratorConfig{MaxWorkers: 1, QueueSize: 1}, scans, profiles, targets, runner, silentLogger())
			o.ctx = context.Background()
			o.SetFindingRepo(findings)
			b := NewEventBroadcaster(32, nil)
			defer b.Close()
			o.SetEventBroadcaster(b)
			ch, unsubscribe := b.Subscribe(context.Background(), domain.ScanTopic(scanID))
			defer unsubscribe()
			o.processJob(&ScanJob{TenantID: tenant, ScanID: scanID}, 0)
			result, err := scans.GetByID(context.Background(), tenant, scanID)
			if err != nil {
				t.Fatal(err)
			}
			want := domain.ScanStatusCompleted
			if fail {
				want = domain.ScanStatusFailed
			}
			if result.Status != want {
				t.Fatalf("status=%s want=%s", result.Status, want)
			}
			b.Close()
			published := 0
			for event := range ch {
				if event.Event == domain.EventScanFinding {
					published++
				}
			}
			if fail && (published != 0 || len(findings.findings) != 0) {
				t.Fatal("uncommitted finding published")
			}
			if !fail && (published != 1 || len(findings.findings) != 1 || result.SummaryStats.InfoCount != 1) {
				t.Fatal("live/final parse duplicated findings")
			}
		})
	}
}
