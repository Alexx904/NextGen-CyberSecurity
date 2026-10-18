package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
)

type FirewallService interface {
	GetStatus(ctx context.Context, tenantID, serverID uuid.UUID) (*domain.FirewallStatus, error)
	AddRule(ctx context.Context, tenantID, serverID uuid.UUID, params domain.FirewallRuleParams, withSafetyTimer bool, timeoutSec int) (*domain.FirewallStatus, error)
	DeleteRule(ctx context.Context, tenantID, serverID uuid.UUID, chain domain.ChainName, lineNum int, withSafetyTimer bool, timeoutSec int) (*domain.FirewallStatus, error)
	ApplyWithSafetyTimer(ctx context.Context, tenantID, serverID uuid.UUID, timeoutSec int) (*domain.FirewallStatus, error)
	Confirm(ctx context.Context, tenantID, serverID uuid.UUID) error
	Rollback(ctx context.Context, tenantID, serverID uuid.UUID, backupID *uuid.UUID) error
}

type firewallService struct {
	executor   SSHExecutor
	backupRepo domain.FirewallBackupRepository
}

func NewFirewallService(executor SSHExecutor, backupRepo domain.FirewallBackupRepository) FirewallService {
	return &firewallService{executor: executor, backupRepo: backupRepo}
}

func (s *firewallService) readRemoteStatus(ctx context.Context, tenantID, serverID uuid.UUID) (*remoteFirewallState, *domain.FirewallStatus, error) {
	output, err := s.remoteCall(ctx, tenantID, serverID, uuid.New(), 0, "status", "")
	if err != nil {
		return nil, nil, err
	}
	state, rules, err := parseRemoteFirewallState(output)
	if err != nil {
		return nil, nil, err
	}
	chains, err := ParseIptablesOutput(rules)
	if err != nil {
		return nil, nil, err
	}
	status := &domain.FirewallStatus{ServerID: serverID, Chains: chains}
	status.Revision, err = remoteFirewallRevision(output)
	if err != nil {
		return nil, nil, err
	}
	if state != nil {
		status.ActiveSafetyTimer = true
		expires := state.ExpiresAt
		status.SafetyTimerExpiresAt = &expires
		status.SafetyTimeoutSeconds = state.Timeout
	}
	return state, status, nil
}

func (s *firewallService) GetStatus(ctx context.Context, tenantID, serverID uuid.UUID) (*domain.FirewallStatus, error) {
	_, status, err := s.readRemoteStatus(ctx, tenantID, serverID)
	return status, err
}

func (s *firewallService) AddRule(ctx context.Context, tenantID, serverID uuid.UUID, params domain.FirewallRuleParams, _ bool, timeoutSec int) (*domain.FirewallStatus, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	verb := "-A"
	if params.Position > 0 {
		verb = "-I"
	}
	return s.mutate(ctx, tenantID, serverID, timeoutSec, "Add firewall rule", firewallRuleArguments(params, verb, params.Position)...)
}

func (s *firewallService) DeleteRule(ctx context.Context, tenantID, serverID uuid.UUID, chain domain.ChainName, lineNum int, _ bool, timeoutSec int) (*domain.FirewallStatus, error) {
	if !chain.IsValid() {
		return nil, domain.ErrInvalidChainName
	}
	if lineNum <= 0 {
		return nil, domain.ErrValidation
	}
	return s.mutate(ctx, tenantID, serverID, timeoutSec, "Delete firewall rule", "-D", string(chain), strconv.Itoa(lineNum))
}

func (s *firewallService) ApplyWithSafetyTimer(ctx context.Context, tenantID, serverID uuid.UUID, timeoutSec int) (*domain.FirewallStatus, error) {
	return s.mutate(ctx, tenantID, serverID, timeoutSec, "Preventive firewall snapshot")
}

func (s *firewallService) mutate(ctx context.Context, tenantID, serverID uuid.UUID, timeoutSec int, label string, args ...string) (*domain.FirewallStatus, error) {
	return s.mutateRevision(ctx, tenantID, serverID, timeoutSec, label, "", args...)
}

func (s *firewallService) mutateRevision(ctx context.Context, tenantID, serverID uuid.UUID, timeoutSec int, label, revision string, args ...string) (*domain.FirewallStatus, error) {
	if timeoutSec == 0 {
		timeoutSec = 60
	}
	if timeoutSec < 10 || timeoutSec > 300 {
		return nil, domain.ErrValidation
	}
	token := uuid.New()
	arguments := append([]string{revision}, args...)
	output, err := s.remoteCall(ctx, tenantID, serverID, token, timeoutSec, "apply", "", arguments...)
	if err != nil {
		return nil, err
	}
	snapshot, _, err := parseRemoteFirewallBackup(output, token)
	if err != nil {
		return nil, err
	}
	backup := &domain.FirewallBackup{ID: token, TenantID: tenantID, ServerID: serverID, RawIptablesRules: snapshot, Label: label, CreatedAt: time.Now().UTC()}
	if err := s.backupRepo.CreateBackup(ctx, backup); err != nil {
		// The detached remote watchdog remains armed. Do not acknowledge success.
		return nil, fmt.Errorf("%w: backup could not be persisted; remote rollback remains armed", domain.ErrSSHExecutionFailed)
	}
	state, status, err := s.readRemoteStatus(ctx, tenantID, serverID)
	if err != nil {
		return nil, err
	}
	if state == nil || state.ID != token || state.TenantID != tenantID || state.ServerID != serverID {
		return nil, domain.ErrNoSafetyTimerActive
	}
	return status, nil
}

func (s *firewallService) Confirm(ctx context.Context, tenantID, serverID uuid.UUID) error {
	state, _, err := s.readRemoteStatus(ctx, tenantID, serverID)
	if err != nil {
		return err
	}
	if state == nil || state.TenantID != tenantID || state.ServerID != serverID || !time.Now().Before(state.ExpiresAt) {
		return domain.ErrNoSafetyTimerActive
	}
	backup, err := s.backupRepo.GetBackupByID(ctx, tenantID, state.ID)
	if err != nil {
		return err
	}
	if backup.ServerID != serverID || backup.TenantID != tenantID {
		return domain.ErrFirewallBackupNotFound
	}
	_, err = s.remoteCall(ctx, tenantID, serverID, state.ID, 0, "confirm", "")
	return err
}

func (s *firewallService) Rollback(ctx context.Context, tenantID, serverID uuid.UUID, backupID *uuid.UUID) error {
	state, _, err := s.readRemoteStatus(ctx, tenantID, serverID)
	if err != nil {
		return err
	}
	if state != nil && (state.TenantID != tenantID || state.ServerID != serverID) {
		return domain.ErrSafetyTimerActive
	}
	var backup *domain.FirewallBackup
	if backupID != nil && *backupID != uuid.Nil {
		backup, err = s.backupRepo.GetBackupByID(ctx, tenantID, *backupID)
	} else if state != nil {
		backup, err = s.backupRepo.GetBackupByID(ctx, tenantID, state.ID)
	} else {
		backup, err = s.backupRepo.GetLatestBackup(ctx, tenantID, serverID)
	}
	if err != nil {
		return err
	}
	if backup.ServerID != serverID || backup.TenantID != tenantID {
		return domain.ErrFirewallBackupNotFound
	}
	if strings.TrimSpace(backup.RawIptablesRules) == "" {
		return domain.ErrValidation
	}
	token := backup.ID
	if state != nil {
		token = state.ID
	}
	_, err = s.remoteCall(ctx, tenantID, serverID, token, 0, "rollback", backup.RawIptablesRules)
	return err
}

func firewallRuleArguments(params domain.FirewallRuleParams, verb string, position int) []string {
	args := []string{verb, string(params.Chain)}
	if position > 0 {
		args = append(args, strconv.Itoa(position))
	}
	if params.Protocol != "" && params.Protocol != "all" {
		args = append(args, "-p", params.Protocol)
	}
	if params.Port != "" {
		args = append(args, "--dport", params.Port)
	}
	if params.Source != "" && params.Source != "0.0.0.0/0" {
		args = append(args, "-s", params.Source)
	}
	if params.Destination != "" && params.Destination != "0.0.0.0/0" {
		args = append(args, "-d", params.Destination)
	}
	if params.InInterface != "" && params.InInterface != "*" {
		args = append(args, "-i", params.InInterface)
	}
	if params.OutInterface != "" && params.OutInterface != "*" {
		args = append(args, "-o", params.OutInterface)
	}
	args = append(args, "-j", string(params.Action))
	if strings.TrimSpace(params.Comment) != "" {
		args = append(args, "-m", "comment", "--comment", params.Comment)
	}
	return args
}

// BuildIptablesAddCommand costruisce la riga di comando per aggiungere la regola
func BuildIptablesAddCommand(params domain.FirewallRuleParams) string {
	var parts []string
	if params.Position > 0 {
		parts = append(parts, fmt.Sprintf("sudo iptables -I %s %d", params.Chain, params.Position))
	} else {
		parts = append(parts, fmt.Sprintf("sudo iptables -A %s", params.Chain))
	}

	if params.Protocol != "" && params.Protocol != "all" {
		parts = append(parts, fmt.Sprintf("-p %s", params.Protocol))
	}

	if params.Port != "" {
		parts = append(parts, fmt.Sprintf("--dport %s", params.Port))
	}

	if params.Source != "" && params.Source != "0.0.0.0/0" {
		parts = append(parts, fmt.Sprintf("-s %s", params.Source))
	}

	if params.Destination != "" && params.Destination != "0.0.0.0/0" {
		parts = append(parts, fmt.Sprintf("-d %s", params.Destination))
	}

	if params.InInterface != "" && params.InInterface != "*" {
		parts = append(parts, fmt.Sprintf("-i %s", params.InInterface))
	}

	if params.OutInterface != "" && params.OutInterface != "*" {
		parts = append(parts, fmt.Sprintf("-o %s", params.OutInterface))
	}

	parts = append(parts, fmt.Sprintf("-j %s", params.Action))

	if strings.TrimSpace(params.Comment) != "" {
		parts = append(parts, "-m comment --comment "+quoteFirewallArgument(params.Comment))
	}

	return strings.Join(parts, " ")
}

func quoteFirewallArgument(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

// ParseIptablesOutput analizza l'output formattato di "iptables -L -n -v --line-numbers"
func ParseIptablesOutput(rawOutput string) (map[domain.ChainName]*domain.FirewallChain, error) {
	chains := make(map[domain.ChainName]*domain.FirewallChain)

	var currentChain *domain.FirewallChain
	lines := strings.Split(rawOutput, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Chain ") {
			// Parsing riga intestazione catena, es. "Chain INPUT (policy ACCEPT 120 packets, 8400 bytes)"
			headerParts := strings.Fields(line)
			if len(headerParts) < 2 {
				continue
			}

			chainName := domain.ChainName(headerParts[1])
			defaultPolicy := domain.ActionAccept
			var pkts, bytesCount uint64

			// Parsing policy e contatori (policy ACCEPT 120 packets, 8400 bytes)
			policyIdx := -1
			for i, p := range headerParts {
				if p == "(policy" && i+1 < len(headerParts) {
					policyIdx = i + 1
					defaultPolicy = domain.RuleAction(strings.TrimSuffix(headerParts[policyIdx], ")"))
					break
				}
			}

			for i, p := range headerParts {
				if p == "packets," && i > 0 {
					pkts = parseCounter(headerParts[i-1])
				}
				if p == "bytes)" && i > 0 {
					bytesCount = parseCounter(headerParts[i-1])
				}
			}

			chain := &domain.FirewallChain{
				Name:          chainName,
				DefaultPolicy: defaultPolicy,
				Packets:       pkts,
				Bytes:         bytesCount,
				Rules:         make([]domain.FirewallRule, 0),
			}
			chains[chainName] = chain
			currentChain = chain
			continue
		}

		// Salta intestazione delle colonne delle regole
		if strings.HasPrefix(line, "num ") || strings.HasPrefix(line, "num\t") {
			continue
		}

		if currentChain == nil {
			continue
		}

		// Parsing riga regola
		// Colonne attese: num pkts bytes target prot opt in out source destination [options...]
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		lineNum, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		pkts := parseCounter(fields[1])
		bytesCount := parseCounter(fields[2])
		target := domain.RuleAction(fields[3])
		prot := fields[4]
		// Numeric listing is used by iptables-nft with -n. Keep the editable
		// protocols canonical while retaining unknown protocol numbers verbatim.
		switch prot {
		case "0":
			prot = "all"
		case "1":
			prot = "icmp"
		case "6":
			prot = "tcp"
		case "17":
			prot = "udp"
		}
		opt := fields[5]
		inIface := fields[6]
		outIface := fields[7]
		source := fields[8]

		destination := "0.0.0.0/0"
		if len(fields) >= 10 {
			destination = fields[9]
		}

		options := ""
		if len(fields) > 10 {
			options = strings.Join(fields[10:], " ")
		}

		rule := domain.FirewallRule{
			LineNumber:   lineNum,
			Packets:      pkts,
			Bytes:        bytesCount,
			Target:       target,
			Protocol:     prot,
			Opt:          opt,
			InInterface:  inIface,
			OutInterface: outIface,
			Source:       source,
			Destination:  destination,
			Options:      options,
		}

		currentChain.Rules = append(currentChain.Rules, rule)
	}

	return chains, nil
}

func parseCounter(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	multiplier := uint64(1)
	lastChar := s[len(s)-1]
	switch lastChar {
	case 'K', 'k':
		multiplier = 1000
		s = s[:len(s)-1]
	case 'M', 'm':
		multiplier = 1000000
		s = s[:len(s)-1]
	case 'G', 'g':
		multiplier = 1000000000
		s = s[:len(s)-1]
	}

	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return val * multiplier
}
