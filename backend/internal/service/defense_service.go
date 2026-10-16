package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"net/netip"
	"nextgen-cybersec/internal/domain"
	"strconv"
)

type DefenseService interface {
	Preview(context.Context, uuid.UUID, string, domain.RemediationRequest) (*domain.RemediationResult, error)
	Remediate(context.Context, uuid.UUID, string, domain.RemediationRequest) (*domain.RemediationResult, error)
}

type DefenseServerReader interface {
	GetByID(context.Context, uuid.UUID, uuid.UUID) (*domain.SSHServer, error)
}

type defenseService struct {
	graph    AttackPathService
	servers  DefenseServerReader
	firewall FirewallService
}

func NewDefenseService(graph AttackPathService, servers DefenseServerReader, firewall FirewallService) DefenseService {
	return &defenseService{graph: graph, servers: servers, firewall: firewall}
}

func (s *defenseService) Preview(ctx context.Context, tenant uuid.UUID, nodeID string, req domain.RemediationRequest) (*domain.RemediationResult, error) {
	if tenant == uuid.Nil || req.ServerID == uuid.Nil || nodeID == "" {
		return nil, domain.ErrValidation
	}
	graph, err := s.graph.GenerateAttackGraph(ctx, tenant, 0)
	if err != nil {
		return nil, err
	}
	found := false
	for _, cp := range graph.Chokepoints {
		if cp.NodeID == nodeID {
			found = true
			break
		}
	}
	if !found {
		return nil, domain.ErrNotFound
	}
	var node *domain.AttackNode
	for i := range graph.Nodes {
		if graph.Nodes[i].ID == nodeID {
			node = &graph.Nodes[i]
			break
		}
	}
	if node == nil {
		return nil, domain.ErrNotFound
	}
	server, err := s.servers.GetByID(ctx, tenant, req.ServerID)
	if err != nil {
		return nil, err
	}
	if server == nil || server.TenantID != tenant || server.ID != req.ServerID {
		return nil, domain.ErrNotFound
	}
	rule, err := BuildChokepointRule(*node, *server)
	if err != nil {
		return nil, err
	}
	return &domain.RemediationResult{NodeID: nodeID, ServerID: server.ID, Rule: rule, Command: BuildIptablesAddCommand(rule), SafetyTimeoutSeconds: 60}, nil
}

// BuildChokepointRule accepts only literal IPv4 service endpoints on the managed host.
func BuildChokepointRule(node domain.AttackNode, server domain.SSHServer) (domain.FirewallRuleParams, error) {
	invalid := func() (domain.FirewallRuleParams, error) {
		return domain.FirewallRuleParams{}, fmt.Errorf("%w: servizio o server non compatibile con la mitigazione", domain.ErrValidation)
	}
	ip, err := netip.ParseAddr(node.IP)
	host, hostErr := netip.ParseAddr(server.Host)
	if node.Type != domain.NodeTypeService || err != nil || hostErr != nil || !ip.Is4() || ip.IsUnspecified() || ip.IsMulticast() || ip != host || node.Port < 1 || node.Port > 65535 {
		return invalid()
	}
	if node.Protocol != "tcp" && node.Protocol != "udp" {
		return invalid()
	}
	if node.Protocol == "tcp" && node.Port == server.Port {
		return invalid()
	}
	rule := domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionDrop, Protocol: node.Protocol, Destination: ip.String(), Port: strconv.Itoa(node.Port), Position: 1, Comment: "NextGen chokepoint mitigation"}
	return rule, rule.Validate()
}

func (s *defenseService) Remediate(ctx context.Context, tenant uuid.UUID, nodeID string, req domain.RemediationRequest) (*domain.RemediationResult, error) {
	result, err := s.Preview(ctx, tenant, nodeID, req)
	if err != nil {
		return nil, err
	}
	result.Firewall, err = s.firewall.AddRule(ctx, tenant, result.ServerID, result.Rule, true, 60)
	if err != nil {
		return nil, err
	}
	return result, nil
}
