package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
)

// FirewallService gestisce la consultazione, configurazione e resilienza anti-lockout del firewall iptables
type FirewallService interface {
	GetStatus(ctx context.Context, tenantID, serverID uuid.UUID) (*domain.FirewallStatus, error)
	AddRule(ctx context.Context, tenantID, serverID uuid.UUID, params domain.FirewallRuleParams, withSafetyTimer bool, timeoutSec int) (*domain.FirewallStatus, error)
	DeleteRule(ctx context.Context, tenantID, serverID uuid.UUID, chain domain.ChainName, lineNum int, withSafetyTimer bool, timeoutSec int) (*domain.FirewallStatus, error)
	ApplyWithSafetyTimer(ctx context.Context, tenantID, serverID uuid.UUID, timeoutSec int) (*domain.FirewallStatus, error)
	Confirm(ctx context.Context, tenantID, serverID uuid.UUID) error
	Rollback(ctx context.Context, tenantID, serverID uuid.UUID, backupID *uuid.UUID) error
}

type activeSafetyTimer struct {
	serverID   uuid.UUID
	tenantID   uuid.UUID
	backupID   uuid.UUID
	remotePID  int
	expiresAt  time.Time
	timeoutSec int
	cancelCh   chan struct{}
}

type firewallService struct {
	operationMu [64]sync.Mutex // Bounded locks serialize each server's mutation lifecycle.
	executor    SSHExecutor
	backupRepo  domain.FirewallBackupRepository
	mu          sync.Mutex
	timers      map[uuid.UUID]*activeSafetyTimer
}

// NewFirewallService crea un'istanza di FirewallService collegando l'esecutore SSH e il repository backup
func NewFirewallService(executor SSHExecutor, backupRepo domain.FirewallBackupRepository) FirewallService {
	return &firewallService{
		executor:   executor,
		backupRepo: backupRepo,
		timers:     make(map[uuid.UUID]*activeSafetyTimer),
	}
}

// GetStatus recupera le catene attive e lo stato dell'anti-lockout timer
func (s *firewallService) GetStatus(ctx context.Context, tenantID, serverID uuid.UUID) (*domain.FirewallStatus, error) {
	cmd := "sudo iptables -L -n -v --line-numbers"
	stdout, stderr, exitCode, err := s.executor.ExecuteCommand(ctx, tenantID, serverID, cmd)
	if err != nil || exitCode != 0 {
		return nil, fmt.Errorf("%w: recupero stato firewall fallito (exit %d): %s %v", domain.ErrSSHExecutionFailed, exitCode, stderr, err)
	}

	chains, err := ParseIptablesOutput(stdout)
	if err != nil {
		return nil, fmt.Errorf("parsing regole firewall fallito: %w", err)
	}

	status := &domain.FirewallStatus{
		ServerID: serverID,
		Chains:   chains,
	}

	s.mu.Lock()
	if timer, ok := s.timers[serverID]; ok {
		status.ActiveSafetyTimer = true
		exp := timer.expiresAt
		status.SafetyTimerExpiresAt = &exp
		status.SafetyTimeoutSeconds = timer.timeoutSec
	}
	s.mu.Unlock()

	return status, nil
}

// AddRule aggiunge una regola in testa (position > 0) o in coda, con opzione di Safety Rollback Timer
func (s *firewallService) AddRule(ctx context.Context, tenantID, serverID uuid.UUID, params domain.FirewallRuleParams, withSafetyTimer bool, timeoutSec int) (*domain.FirewallStatus, error) {
	lock := &s.operationMu[int(serverID[0])%len(s.operationMu)]
	lock.Lock()
	defer lock.Unlock()
	if err := params.Validate(); err != nil {
		return nil, err
	}

	if withSafetyTimer {
		if timeoutSec <= 0 {
			timeoutSec = 60
		}
		if err := s.startSafetyTimer(ctx, tenantID, serverID, timeoutSec, fmt.Sprintf("Aggiunta regola su %s (%s)", params.Chain, params.Action)); err != nil {
			return nil, err
		}
	}

	ruleCmd := BuildIptablesAddCommand(params)
	_, stderr, exitCode, err := s.executor.ExecuteCommand(ctx, tenantID, serverID, ruleCmd)
	if err != nil || exitCode != 0 {
		return nil, fmt.Errorf("%w: inserimento regola fallito (exit %d): %s %v", domain.ErrSSHExecutionFailed, exitCode, stderr, err)
	}

	return s.GetStatus(ctx, tenantID, serverID)
}

// DeleteRule rimuove una regola per numero di riga, opzionalmente protetta dal Safety Timer
func (s *firewallService) DeleteRule(ctx context.Context, tenantID, serverID uuid.UUID, chain domain.ChainName, lineNum int, withSafetyTimer bool, timeoutSec int) (*domain.FirewallStatus, error) {
	lock := &s.operationMu[int(serverID[0])%len(s.operationMu)]
	lock.Lock()
	defer lock.Unlock()
	if !chain.IsValid() {
		return nil, domain.ErrInvalidChainName
	}
	if lineNum <= 0 {
		return nil, fmt.Errorf("%w: numero riga non valido", domain.ErrValidation)
	}

	if withSafetyTimer {
		if timeoutSec <= 0 {
			timeoutSec = 60
		}
		if err := s.startSafetyTimer(ctx, tenantID, serverID, timeoutSec, fmt.Sprintf("Eliminazione regola %d da %s", lineNum, chain)); err != nil {
			return nil, err
		}
	}

	delCmd := fmt.Sprintf("sudo iptables -D %s %d", chain, lineNum)
	_, stderr, exitCode, err := s.executor.ExecuteCommand(ctx, tenantID, serverID, delCmd)
	if err != nil || exitCode != 0 {
		return nil, fmt.Errorf("%w: eliminazione regola fallita (exit %d): %s %v", domain.ErrSSHExecutionFailed, exitCode, stderr, err)
	}

	return s.GetStatus(ctx, tenantID, serverID)
}

// ApplyWithSafetyTimer attiva esplicitamente uno snapshot e avvia il timer di rollback preventivo
func (s *firewallService) ApplyWithSafetyTimer(ctx context.Context, tenantID, serverID uuid.UUID, timeoutSec int) (*domain.FirewallStatus, error) {
	lock := &s.operationMu[int(serverID[0])%len(s.operationMu)]
	lock.Lock()
	defer lock.Unlock()
	if timeoutSec <= 0 {
		timeoutSec = 60
	}

	if err := s.startSafetyTimer(ctx, tenantID, serverID, timeoutSec, "Snapshot preventivo manuale"); err != nil {
		return nil, err
	}

	return s.GetStatus(ctx, tenantID, serverID)
}

// Confirm disattiva il timer e termina il processo di revert remoto garantendo le modifiche
func (s *firewallService) Confirm(ctx context.Context, tenantID, serverID uuid.UUID) error {
	lock := &s.operationMu[int(serverID[0])%len(s.operationMu)]
	lock.Lock()
	defer lock.Unlock()
	s.mu.Lock()
	timer, exists := s.timers[serverID]
	if !exists || timer.tenantID != tenantID || !time.Now().Before(timer.expiresAt) {
		s.mu.Unlock()
		return domain.ErrNoSafetyTimerActive
	}
	s.mu.Unlock()

	// Termina il processo di revert autonomo sull'host remoto e pulisce lo snapshot temporaneo
	killCmd := fmt.Sprintf("sudo kill %d && rm -f /tmp/ng_fw_backup_%s.rules", timer.remotePID, serverID.String())
	if _, _, code, err := s.executor.ExecuteCommand(ctx, tenantID, serverID, killCmd); err != nil || code != 0 {
		return domain.ErrSSHExecutionFailed
	}
	s.mu.Lock()
	delete(s.timers, serverID)
	s.mu.Unlock()
	close(timer.cancelCh)

	return nil
}

// Rollback ripristina immediatamente l'ultimo snapshot valido via iptables-restore
func (s *firewallService) Rollback(ctx context.Context, tenantID, serverID uuid.UUID, backupID *uuid.UUID) error {
	lock := &s.operationMu[int(serverID[0])%len(s.operationMu)]
	lock.Lock()
	defer lock.Unlock()

	var backup *domain.FirewallBackup
	var err error

	if backupID != nil && *backupID != uuid.Nil {
		backup, err = s.backupRepo.GetBackupByID(ctx, tenantID, *backupID)
	} else {
		backup, err = s.backupRepo.GetLatestBackup(ctx, tenantID, serverID)
	}

	if err != nil {
		return fmt.Errorf("recupero snapshot backup fallito: %w", err)
	}
	if backup.ServerID != serverID || backup.TenantID != tenantID {
		return domain.ErrFirewallBackupNotFound
	}

	// Ripristino mediante stream di regole su sudo iptables-restore
	restoreCmd := fmt.Sprintf("sudo iptables-restore <<'EOF'\n%s\nEOF", backup.RawIptablesRules)
	_, stderr, exitCode, err := s.executor.ExecuteCommand(ctx, tenantID, serverID, restoreCmd)
	if err != nil || exitCode != 0 {
		return fmt.Errorf("%w: ripristino iptables-restore fallito (exit %d): %s %v", domain.ErrSSHExecutionFailed, exitCode, stderr, err)
	}
	// Keep the remote watchdog alive until restore has succeeded.
	s.mu.Lock()
	timer, exists := s.timers[serverID]
	s.mu.Unlock()
	if exists && timer.tenantID == tenantID {
		cmd := fmt.Sprintf("sudo kill %d && rm -f /tmp/ng_fw_backup_%s.rules", timer.remotePID, serverID.String())
		if _, _, code, err := s.executor.ExecuteCommand(ctx, tenantID, serverID, cmd); err != nil || code != 0 {
			return domain.ErrSSHExecutionFailed
		}
		s.mu.Lock()
		delete(s.timers, serverID)
		s.mu.Unlock()
		close(timer.cancelCh)
	}

	return nil
}

func (s *firewallService) startSafetyTimer(ctx context.Context, tenantID, serverID uuid.UUID, timeoutSec int, label string) error {
	s.mu.Lock()
	if _, exists := s.timers[serverID]; exists {
		s.mu.Unlock()
		return domain.ErrSafetyTimerActive
	}
	s.mu.Unlock()

	// 1. Snapshot corrente con iptables-save
	stdout, stderr, exitCode, err := s.executor.ExecuteCommand(ctx, tenantID, serverID, "sudo iptables-save")
	if err != nil || exitCode != 0 {
		return fmt.Errorf("%w: esecuzione iptables-save fallita (exit %d): %s %v", domain.ErrSSHExecutionFailed, exitCode, stderr, err)
	}

	rawRules := stdout
	backup := &domain.FirewallBackup{
		ID:               uuid.New(),
		TenantID:         tenantID,
		ServerID:         serverID,
		RawIptablesRules: rawRules,
		Label:            label,
		CreatedAt:        time.Now().UTC(),
	}

	if err := s.backupRepo.CreateBackup(ctx, backup); err != nil {
		return fmt.Errorf("salvataggio backup snapshot fallito: %w", err)
	}

	// 2. Scrittura snapshot temporaneo sull'host remoto e lancio processo watchdog distaccato
	backupFile := fmt.Sprintf("/tmp/ng_fw_backup_%s.rules", serverID.String())
	prepCmd := fmt.Sprintf("cat <<'EOF' > %s\n%s\nEOF", backupFile, rawRules)
	if _, stderr, exitCode, err := s.executor.ExecuteCommand(ctx, tenantID, serverID, prepCmd); err != nil || exitCode != 0 {
		return fmt.Errorf("scrittura file snapshot remoto fallita: %s %v", stderr, err)
	}

	// Remote self-revert watchdog distaccato con nohup (garantisce revert anche con disconnessione o restart backend)
	watchdogCmd := fmt.Sprintf("nohup sh -c 'sleep %d && sudo iptables-restore < %s && rm -f %s' >/dev/null 2>&1 & echo $!", timeoutSec, backupFile, backupFile)
	pidOut, stderr, exitCode, err := s.executor.ExecuteCommand(ctx, tenantID, serverID, watchdogCmd)
	if err != nil || exitCode != 0 {
		return fmt.Errorf("avvio watchdog remoto fallito: %s %v", stderr, err)
	}

	remotePID, pidErr := strconv.Atoi(strings.TrimSpace(pidOut))
	if pidErr != nil || remotePID <= 1 {
		return fmt.Errorf("%w: PID watchdog non valido", domain.ErrSSHExecutionFailed)
	}

	// 3. Registrazione timer in memoria
	cancelCh := make(chan struct{})
	expiresAt := time.Now().Add(time.Duration(timeoutSec) * time.Second)
	timer := &activeSafetyTimer{
		serverID:   serverID,
		tenantID:   tenantID,
		backupID:   backup.ID,
		remotePID:  remotePID,
		expiresAt:  expiresAt,
		timeoutSec: timeoutSec,
		cancelCh:   cancelCh,
	}

	s.mu.Lock()
	s.timers[serverID] = timer
	s.mu.Unlock()

	// 4. Goroutine watchdog backend per sincronizzazione dello stato
	go func() {
		select {
		case <-time.After(time.Duration(timeoutSec) * time.Second):
			s.mu.Lock()
			currentTimer, exists := s.timers[serverID]
			if exists && currentTimer == timer {
				delete(s.timers, serverID)
			}
			s.mu.Unlock()
		case <-cancelCh:
			// Timer annullato esplicitamente da Confirm o Rollback
			return
		}
	}()

	return nil
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
		sanitizedComment := strings.ReplaceAll(params.Comment, `"`, "")
		parts = append(parts, fmt.Sprintf(`-m comment --comment "%s"`, sanitizedComment))
	}

	return strings.Join(parts, " ")
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
