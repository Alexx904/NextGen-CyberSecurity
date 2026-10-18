package service_test

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"strings"
	"testing"
)

func TestBuildChokepointRule(t *testing.T) {
	node := domain.AttackNode{Type: domain.NodeTypeService, IP: "192.0.2.10", Port: 443, Protocol: "tcp"}
	server := domain.SSHServer{Host: node.IP, Port: 22}
	rule, err := service.BuildChokepointRule(node, server)
	if err != nil {
		t.Fatal(err)
	}
	if got := service.BuildIptablesAddCommand(rule); got != `sudo iptables -I INPUT 1 -p tcp --dport 443 -d 192.0.2.10 -j DROP -m comment --comment 'NextGen chokepoint mitigation'` {
		t.Fatal(got)
	}
	for _, tc := range []struct {
		name   string
		change func(*domain.AttackNode, *domain.SSHServer)
	}{
		{"host mismatch", func(n *domain.AttackNode, s *domain.SSHServer) { s.Host = "192.0.2.11" }},
		{"dns", func(n *domain.AttackNode, s *domain.SSHServer) { s.Host = "example.test" }},
		{"ssh", func(n *domain.AttackNode, s *domain.SSHServer) { n.Port = 22 }},
		{"ipv6", func(n *domain.AttackNode, s *domain.SSHServer) { n.IP = "::1"; s.Host = "::1" }},
		{"injection", func(n *domain.AttackNode, s *domain.SSHServer) { n.IP = "192.0.2.10;id" }},
		{"protocol", func(n *domain.AttackNode, s *domain.SSHServer) { n.Protocol = "tcp;id" }},
		{"host node", func(n *domain.AttackNode, s *domain.SSHServer) { n.Type = domain.NodeTypeHost }},
		{"port zero", func(n *domain.AttackNode, s *domain.SSHServer) { n.Port = 0 }},
		{"port overflow", func(n *domain.AttackNode, s *domain.SSHServer) { n.Port = 65536 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n, s := node, server
			tc.change(&n, &s)
			if _, err := service.BuildChokepointRule(n, s); !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected validation: %v", err)
			}
		})
	}
	node.Protocol = "udp"
	node.Port = 22
	if _, err := service.BuildChokepointRule(node, server); err != nil {
		t.Fatal(err)
	}
}

type defenseGraph struct {
	service.AttackPathService
	graph *domain.AttackGraph
}

func (m defenseGraph) GenerateAttackGraph(context.Context, uuid.UUID, float64) (*domain.AttackGraph, error) {
	return m.graph, nil
}

type defenseServers struct{ server *domain.SSHServer }

func (m defenseServers) GetByID(_ context.Context, tenant, id uuid.UUID) (*domain.SSHServer, error) {
	return m.server, nil
}

type defenseFirewall struct {
	service.FirewallService
	called bool
	fail   error
}

func (m *defenseFirewall) AddRule(_ context.Context, tenant, id uuid.UUID, p domain.FirewallRuleParams, safe bool, seconds int) (*domain.FirewallStatus, error) {
	m.called = true
	if !safe || seconds != 60 || p.Action != domain.ActionDrop {
		panic("unsafe call")
	}
	return &domain.FirewallStatus{ServerID: id, ActiveSafetyTimer: true}, m.fail
}
func TestDefenseTenantAndChokepointValidation(t *testing.T) {
	tenant, id := uuid.New(), uuid.New()
	server := &domain.SSHServer{ID: id, TenantID: tenant, Host: "192.0.2.10", Port: 22}
	graph := &domain.AttackGraph{Nodes: []domain.AttackNode{{ID: "svc", Type: domain.NodeTypeService, IP: server.Host, Port: 443, Protocol: "tcp"}}, Chokepoints: []domain.Chokepoint{{NodeID: "svc"}}}
	fw := &defenseFirewall{}
	svc := service.NewDefenseService(defenseGraph{graph: graph}, defenseServers{server}, fw)
	req := domain.RemediationRequest{ServerID: id}
	if _, err := svc.Preview(context.Background(), tenant, "svc", req); err != nil || fw.called {
		t.Fatalf("preview: %v", err)
	}
	if _, err := svc.Remediate(context.Background(), uuid.New(), "svc", req); !errors.Is(err, domain.ErrNotFound) || fw.called {
		t.Fatalf("tenant: %v", err)
	}
	if _, err := svc.Remediate(context.Background(), tenant, "stale", req); !errors.Is(err, domain.ErrNotFound) || fw.called {
		t.Fatalf("stale: %v", err)
	}
	result, err := svc.Remediate(context.Background(), tenant, "svc", req)
	if err != nil || !fw.called || !result.Firewall.ActiveSafetyTimer || !strings.Contains(result.Command, "--dport 443") {
		t.Fatalf("apply: %+v %v", result, err)
	}
	fw.fail = domain.ErrSafetyTimerActive
	if _, err := svc.Remediate(context.Background(), tenant, "svc", req); !errors.Is(err, domain.ErrSafetyTimerActive) {
		t.Fatal(err)
	}
}
