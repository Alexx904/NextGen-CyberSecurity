//go:build integration

package service

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
)

type fixtureDefenseGraph struct {
	AttackPathService
	tenant uuid.UUID
	graph  *domain.AttackGraph
}

func (g fixtureDefenseGraph) GenerateAttackGraph(_ context.Context, tenant uuid.UUID, _ float64) (*domain.AttackGraph, error) {
	if tenant != g.tenant {
		return nil, domain.ErrNotFound
	}
	return g.graph, nil
}

func TestDefenseSSHBlocksServiceAndRestoresAfterRollback(t *testing.T) {
	// The fixture helper rejects local commands outside the named disposable
	// container. This test must run there so HTTP and iptables share a namespace.
	if hostname, _ := os.Hostname(); os.Getenv("TEST_FIREWALL_SSH_IN_CONTAINER") != "1" || hostname != "nextgen-ssh-fixture" {
		t.Fatal("dedicated Linux SSH fixture is required")
	}
	executor, tenant, serverID := firewallSSHFixture(t)
	listener, err := net.Listen("tcp4", "127.0.0.1:45695")
	if err != nil {
		t.Fatal(err)
	}
	endpoint := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	_ = endpoint.Listener.Close()
	endpoint.Listener = listener
	endpoint.Start()
	defer endpoint.Close()
	client := &http.Client{Timeout: 750 * time.Millisecond, Transport: &http.Transport{DisableKeepAlives: true}}
	defer client.CloseIdleConnections()
	request := func() error {
		response, err := client.Get(endpoint.URL)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return errors.New("unexpected fixture HTTP status")
		}
		return nil
	}
	if err := request(); err != nil {
		t.Fatal("fixture baseline unreachable", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	server := &domain.SSHServer{ID: serverID, TenantID: tenant, Host: "127.0.0.1", Port: 22}
	inventory := &fixtureSSHInventory{server: server}
	backups := &fixtureFirewallBackups{items: make(map[uuid.UUID]*domain.FirewallBackup)}
	firewall := NewFirewallService(executor, backups)
	graph := &domain.AttackGraph{Nodes: []domain.AttackNode{{ID: "fixture-service", Type: domain.NodeTypeService, IP: server.Host, Port: 45695, Protocol: "tcp"}}, Chokepoints: []domain.Chokepoint{{NodeID: "fixture-service"}}}
	defense := NewDefenseService(fixtureDefenseGraph{tenant: tenant, graph: graph}, inventory, firewall)
	before, err := firewall.GetStatus(ctx, tenant, serverID)
	if err != nil || before.ActiveSafetyTimer {
		t.Fatalf("fixture baseline not available: %+v %v", before, err)
	}
	req := domain.RemediationRequest{ServerID: serverID}
	if _, err := defense.Preview(ctx, tenant, "fixture-service", req); err != nil {
		t.Fatal(err)
	}
	if _, err := defense.Remediate(ctx, uuid.New(), "fixture-service", req); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("foreign tenant accepted", err)
	}
	if _, err := defense.Remediate(ctx, tenant, "stale-node", req); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("stale chokepoint accepted", err)
	}
	afterPreview, err := firewall.GetStatus(ctx, tenant, serverID)
	if err != nil || afterPreview.Revision != before.Revision || afterPreview.ActiveSafetyTimer {
		t.Fatal("preview/rejected requests mutated firewall", err)
	}
	result, err := defense.Remediate(ctx, tenant, "fixture-service", req)
	if err != nil || result == nil || result.Firewall == nil || !result.Firewall.ActiveSafetyTimer {
		t.Fatalf("protected remediation failed: %+v %v", result, err)
	}
	backup, err := backups.GetLatestBackup(ctx, tenant, serverID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := firewall.Rollback(context.Background(), tenant, serverID, &backup.ID); err != nil {
			t.Error("fixture rollback cleanup", err)
		}
	})
	if err := request(); err == nil {
		t.Fatal("remediation did not block actual HTTP connection")
	}
	second := NewFirewallService(executor, backups)
	if err := second.Confirm(ctx, tenant, serverID); err != nil {
		t.Fatal("fresh instance confirmation failed", err)
	}
	if err := request(); err == nil {
		t.Fatal("confirmation did not retain the block")
	}
	if err := second.Rollback(ctx, tenant, serverID, &backup.ID); err != nil {
		t.Fatal(err)
	}
	if err := request(); err != nil {
		t.Fatal("explicit rollback did not restore HTTP reachability", err)
	}
	final, err := second.GetStatus(ctx, tenant, serverID)
	if err != nil || final.ActiveSafetyTimer || final.Revision != before.Revision {
		t.Fatal("fixture firewall baseline not restored", err)
	}
}
