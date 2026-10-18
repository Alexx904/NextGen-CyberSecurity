package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/scanner"
)

type discoveryStoreStub struct {
	create   func(context.Context, *domain.DiscoveryResult) error
	save     func(context.Context, uuid.UUID, uuid.UUID, domain.DiscoveredDevice) error
	complete func(context.Context, *domain.DiscoveryResult) error
	last     func(context.Context, uuid.UUID) (*domain.DiscoveryResult, error)
	session  func(context.Context, uuid.UUID, uuid.UUID) (*domain.DiscoveryResult, error)
}

func (r *discoveryStoreStub) Create(c context.Context, d *domain.DiscoveryResult) error {
	return r.create(c, d)
}
func (r *discoveryStoreStub) SaveDevice(c context.Context, t, s uuid.UUID, d domain.DiscoveredDevice) error {
	return r.save(c, t, s, d)
}
func (r *discoveryStoreStub) Complete(c context.Context, d *domain.DiscoveryResult) error {
	return r.complete(c, d)
}
func (r *discoveryStoreStub) GetLast(c context.Context, t uuid.UUID) (*domain.DiscoveryResult, error) {
	return r.last(c, t)
}
func (r *discoveryStoreStub) GetSession(c context.Context, t, s uuid.UUID) (*domain.DiscoveryResult, error) {
	return r.session(c, t, s)
}

func TestDiscoveryPersistenceFailureOrdering(t *testing.T) {
	for _, failAt := range []string{"create", "device", "complete", "cancel"} {
		t.Run(failAt, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			tenant, session := uuid.New(), uuid.New()
			b := NewEventBroadcaster(32, nil)
			defer b.Close()
			ch, unsubscribe := b.Subscribe(ctx, domain.TenantDiscoveryTopic(tenant, session))
			defer unsubscribe()
			runs, saves, completes := 0, 0, 0
			dbError := errors.New("test storage unavailable")
			r := &discoveryStoreStub{
				create: func(context.Context, *domain.DiscoveryResult) error {
					if failAt == "create" {
						return domain.ErrConflict
					}
					return nil
				},
				save: func(c context.Context, tenantID, sessionID uuid.UUID, device domain.DiscoveredDevice) error {
					if tenantID != tenant || sessionID != session {
						t.Fatal("wrong persistence owner")
					}
					saves++
					if failAt == "device" && saves == 2 {
						return dbError
					}
					return c.Err()
				},
				complete: func(context.Context, *domain.DiscoveryResult) error { completes++; return dbError },
			}
			runner := &mockScanRunner{runScanFn: func(context.Context, string, scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
				runs++
				if failAt == "cancel" {
					cancel()
				}
				return &scanner.ScanExecutionResult{RawXML: sampleDiscoveryXML}, nil
			}}
			svc := NewDiscoveryService(runner, nil, nil, b, r)
			_, err := svc.DiscoverSubnet(ctx, tenant, domain.DiscoveryRequest{Subnet: "192.168.1.0/24", SessionID: &session})
			if err == nil {
				t.Fatal("storage/cancellation failure reported success")
			}
			if failAt == "create" && (runs != 0 || !errors.Is(err, domain.ErrConflict)) {
				t.Fatal("conflict did not stop scanner")
			}
			if failAt != "complete" && completes != 0 {
				t.Fatal("completed after a failed write")
			}
			if failAt == "cancel" && !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation lost: %v", err)
			}
			// Closing drains already queued events without a timing-dependent wait.
			b.Close()
			devices := 0
			for e := range ch {
				if e.Event == domain.EventDiscoveryComplete {
					t.Fatal("published uncommitted completion")
				}
				if e.Event == domain.EventDeviceDiscovered {
					devices++
				}
			}
			if failAt == "device" && devices != 1 {
				t.Fatalf("published %d devices, want only first committed device", devices)
			}
		})
	}
}

func TestDiscoveryPersistenceReadsNeverFallbackToMemory(t *testing.T) {
	dbError := errors.New("test storage unavailable")
	tenant, session := uuid.New(), uuid.New()
	r := &discoveryStoreStub{
		last: func(_ context.Context, owner uuid.UUID) (*domain.DiscoveryResult, error) {
			if owner != tenant {
				t.Fatal("wrong tenant")
			}
			return nil, dbError
		},
		session: func(_ context.Context, owner, id uuid.UUID) (*domain.DiscoveryResult, error) {
			if owner != tenant || id != session {
				t.Fatal("wrong session owner")
			}
			return nil, dbError
		},
	}
	svc := NewDiscoveryService(nil, nil, nil, nil, r)
	if _, err := svc.GetLastResults(context.Background(), tenant); !errors.Is(err, dbError) {
		t.Fatal("database failure hidden")
	}
	if _, err := svc.GetSessionResults(context.Background(), tenant, session); !errors.Is(err, dbError) {
		t.Fatal("database failure hidden")
	}
}

func TestDiscoveryRejectsReusedSessionBeforeScanner(t *testing.T) {
	runs := 0
	runner := &mockScanRunner{runScanFn: func(context.Context, string, scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
		runs++
		return nil, errors.New("scanner unavailable")
	}}
	svc := NewDiscoveryService(runner, nil, nil, nil)
	tenant, session := uuid.New(), uuid.New()
	req := domain.DiscoveryRequest{Subnet: "192.168.1.0/24", SessionID: &session}
	_, _ = svc.DiscoverSubnet(context.Background(), tenant, req)
	if _, err := svc.DiscoverSubnet(context.Background(), uuid.New(), req); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("session overwritten by different tenant")
	}
	if runs != 1 {
		t.Fatal("duplicate session launched scanner")
	}
}

func TestDiscoveryPersistsLiveHostBeforeRunnerReturns(t *testing.T) {
	tenant, session := uuid.New(), uuid.New()
	var saved []domain.DiscoveredDevice
	r := &discoveryStoreStub{
		create: func(context.Context, *domain.DiscoveryResult) error { return nil },
		save: func(_ context.Context, owner, id uuid.UUID, device domain.DiscoveredDevice) error {
			if owner != tenant || id != session {
				t.Fatal("incorrect owner")
			}
			saved = append(saved, device)
			return nil
		},
		complete: func(context.Context, *domain.DiscoveryResult) error { return nil },
	}
	runner := &mockScanRunner{runScanFn: func(_ context.Context, _ string, profile scanner.ProfileConfig) (*scanner.ScanExecutionResult, error) {
		if profile.OnHost == nil {
			t.Fatal("live host callback missing")
		}
		if err := profile.OnHost(scanner.ParsedHost{IP: "192.168.1.10", Status: "up"}); err != nil {
			t.Fatal(err)
		}
		if len(saved) != 1 || saved[0].IP != "192.168.1.10" {
			t.Fatal("host not durable while scanner still running")
		}
		if err := profile.OnHost(scanner.ParsedHost{IP: "192.168.1.50", Status: "down"}); err != nil {
			t.Fatal(err)
		}
		if len(saved) != 1 {
			t.Fatal("down host persisted")
		}
		return &scanner.ScanExecutionResult{RawXML: sampleDiscoveryXML}, nil
	}}
	svc := NewDiscoveryService(runner, nil, nil, nil, r)
	result, err := svc.DiscoverSubnet(context.Background(), tenant, domain.DiscoveryRequest{Subnet: "192.168.1.0/24", SessionID: &session})
	if err != nil || result.TotalDiscovered != 2 || len(saved) != 3 {
		t.Fatalf("enrichment/upsert failed: saves=%d err=%v", len(saved), err)
	}
}
