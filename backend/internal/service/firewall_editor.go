package service

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
)

// FirewallRuleEditor adds revision-checked operations without breaking existing
// FirewallService consumers. Legacy API calls retain their previous contract.
type FirewallRuleEditor interface {
	SetPolicy(context.Context, uuid.UUID, uuid.UUID, domain.ChainName, domain.RuleAction, string, int) (*domain.FirewallStatus, error)
	ReplaceRule(context.Context, uuid.UUID, uuid.UUID, int, domain.FirewallRuleParams, string, int) (*domain.FirewallStatus, error)
	MoveRule(context.Context, uuid.UUID, uuid.UUID, domain.ChainName, int, int, string, int) (*domain.FirewallStatus, error)
	AddRuleWithRevision(context.Context, uuid.UUID, uuid.UUID, domain.FirewallRuleParams, string, int) (*domain.FirewallStatus, error)
	DeleteRuleWithRevision(context.Context, uuid.UUID, uuid.UUID, domain.ChainName, int, string, int) (*domain.FirewallStatus, error)
}

func (s *firewallService) SetPolicy(ctx context.Context, tenant, server uuid.UUID, chain domain.ChainName, policy domain.RuleAction, revision string, timeout int) (*domain.FirewallStatus, error) {
	if !chain.IsValid() || (policy != domain.ActionAccept && policy != domain.ActionDrop) || !validFirewallRevision(revision) {
		return nil, domain.ErrValidation
	}
	return s.mutateRevision(ctx, tenant, server, timeout, "Change default firewall policy", revision, "-P", string(chain), string(policy))
}
func (s *firewallService) ReplaceRule(ctx context.Context, tenant, server uuid.UUID, line int, params domain.FirewallRuleParams, revision string, timeout int) (*domain.FirewallStatus, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	if line < 1 || !validFirewallRevision(revision) {
		return nil, domain.ErrValidation
	}
	return s.mutateRevision(ctx, tenant, server, timeout, "Replace firewall rule", revision, firewallRuleArguments(params, "-R", line)...)
}
func (s *firewallService) MoveRule(ctx context.Context, tenant, server uuid.UUID, chain domain.ChainName, from, to int, revision string, timeout int) (*domain.FirewallStatus, error) {
	if !chain.IsValid() || from < 1 || to < 1 || from == to || !validFirewallRevision(revision) {
		return nil, domain.ErrValidation
	}
	return s.mutateRevision(ctx, tenant, server, timeout, "Move firewall rule", revision, "_MOVE", string(chain), strconv.Itoa(from), strconv.Itoa(to))
}
func (s *firewallService) AddRuleWithRevision(ctx context.Context, tenant, server uuid.UUID, params domain.FirewallRuleParams, revision string, timeout int) (*domain.FirewallStatus, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	if !validFirewallRevision(revision) {
		return nil, domain.ErrValidation
	}
	verb := "-A"
	if params.Position > 0 {
		verb = "-I"
	}
	return s.mutateRevision(ctx, tenant, server, timeout, "Add firewall rule", revision, firewallRuleArguments(params, verb, params.Position)...)
}
func (s *firewallService) DeleteRuleWithRevision(ctx context.Context, tenant, server uuid.UUID, chain domain.ChainName, line int, revision string, timeout int) (*domain.FirewallStatus, error) {
	if !chain.IsValid() || line < 1 || !validFirewallRevision(revision) {
		return nil, domain.ErrValidation
	}
	return s.mutateRevision(ctx, tenant, server, timeout, "Delete firewall rule", revision, "-D", string(chain), strconv.Itoa(line))
}
