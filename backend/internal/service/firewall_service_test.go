package service_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

type mockSSHExecutor struct {
	mu          sync.Mutex
	commandLogs []string
	executeFunc func(ctx context.Context, tenantID, serverID uuid.UUID, cmd string) (string, string, int, error)
}

func (m *mockSSHExecutor) ExecuteCommand(ctx context.Context, tenantID, serverID uuid.UUID, cmd string) (string, string, int, error) {
	m.mu.Lock()
	m.commandLogs = append(m.commandLogs, cmd)
	m.mu.Unlock()

	if m.executeFunc != nil {
		return m.executeFunc(ctx, tenantID, serverID, cmd)
	}
	return "", "", 0, nil
}

type mockBackupRepo struct {
	mu      sync.Mutex
	backups map[string]*domain.FirewallBackup
}

func newMockBackupRepo() *mockBackupRepo {
	return &mockBackupRepo{
		backups: make(map[string]*domain.FirewallBackup),
	}
}

func (m *mockBackupRepo) makeKey(tenantID, backupID uuid.UUID) string {
	return tenantID.String() + ":" + backupID.String()
}

func (m *mockBackupRepo) CreateBackup(ctx context.Context, backup *domain.FirewallBackup) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.backups[m.makeKey(backup.TenantID, backup.ID)] = backup
	return nil
}

func (m *mockBackupRepo) GetBackupByID(ctx context.Context, tenantID, backupID uuid.UUID) (*domain.FirewallBackup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.backups[m.makeKey(tenantID, backupID)]
	if !ok {
		return nil, domain.ErrFirewallBackupNotFound
	}
	return b, nil
}

func (m *mockBackupRepo) GetLatestBackup(ctx context.Context, tenantID, serverID uuid.UUID) (*domain.FirewallBackup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var latest *domain.FirewallBackup
	for _, b := range m.backups {
		if b.TenantID == tenantID && b.ServerID == serverID {
			if latest == nil || b.CreatedAt.After(latest.CreatedAt) {
				latest = b
			}
		}
	}
	if latest == nil {
		return nil, domain.ErrFirewallBackupNotFound
	}
	return latest, nil
}

func (m *mockBackupRepo) ListBackups(ctx context.Context, tenantID, serverID uuid.UUID) ([]*domain.FirewallBackup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []*domain.FirewallBackup
	for _, b := range m.backups {
		if b.TenantID == tenantID && b.ServerID == serverID {
			list = append(list, b)
		}
	}
	return list, nil
}

func (m *mockBackupRepo) DeleteBackup(ctx context.Context, tenantID, backupID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.makeKey(tenantID, backupID)
	if _, ok := m.backups[key]; !ok {
		return domain.ErrFirewallBackupNotFound
	}
	delete(m.backups, key)
	return nil
}

const sampleIptablesOutput = `
Chain INPUT (policy ACCEPT 150 packets, 12K bytes)
num   pkts bytes target     prot opt in     out     source               destination         
1      100  8000 ACCEPT     all  --  lo     *       0.0.0.0/0            0.0.0.0/0           
2       40  3200 ACCEPT     tcp  --  eth0   *       0.0.0.0/0            0.0.0.0/0            tcp dpt:22
3       10   800 DROP       all  --  *      *       192.168.1.200        0.0.0.0/0           

Chain FORWARD (policy DROP 0 packets, 0 bytes)
num   pkts bytes target     prot opt in     out     source               destination         

Chain OUTPUT (policy ACCEPT 200 packets, 18000 bytes)
num   pkts bytes target     prot opt in     out     source               destination         
1      200 18000 ACCEPT     all  --  *      *       0.0.0.0/0            0.0.0.0/0           
`

func TestParseIptablesOutput(t *testing.T) {
	chains, err := service.ParseIptablesOutput(sampleIptablesOutput)
	if err != nil {
		t.Fatalf("ParseIptablesOutput fallito: %v", err)
	}

	// 1. Verifica catene presenti
	if len(chains) != 3 {
		t.Fatalf("attese 3 catene, trovate %d", len(chains))
	}

	inputChain, ok := chains[domain.ChainInput]
	if !ok {
		t.Fatal("catena INPUT non trovata")
	}
	if inputChain.DefaultPolicy != domain.ActionAccept {
		t.Errorf("attesa policy ACCEPT per INPUT, ottenuto: %s", inputChain.DefaultPolicy)
	}
	if inputChain.Packets != 150 {
		t.Errorf("attesi 150 pacchetti, ottenuti %d", inputChain.Packets)
	}
	if inputChain.Bytes != 12000 {
		t.Errorf("attesi 12000 byte (12K), ottenuti %d", inputChain.Bytes)
	}

	// 2. Verifica regole catena INPUT
	if len(inputChain.Rules) != 3 {
		t.Fatalf("attese 3 regole in INPUT, ottenute %d", len(inputChain.Rules))
	}

	// Regola 1: loopback
	r1 := inputChain.Rules[0]
	if r1.LineNumber != 1 || r1.Target != domain.ActionAccept || r1.InInterface != "lo" {
		t.Errorf("regola 1 errata: %+v", r1)
	}

	// Regola 2: SSH porta 22
	r2 := inputChain.Rules[1]
	if r2.LineNumber != 2 || r2.Protocol != "tcp" || r2.InInterface != "eth0" || !strings.Contains(r2.Options, "dpt:22") {
		t.Errorf("regola 2 errata: %+v", r2)
	}

	// Regola 3: DROP IP molesto
	r3 := inputChain.Rules[2]
	if r3.LineNumber != 3 || r3.Target != domain.ActionDrop || r3.Source != "192.168.1.200" {
		t.Errorf("regola 3 errata: %+v", r3)
	}

	// 3. Verifica catena FORWARD (vuota con policy DROP)
	forwardChain, ok := chains[domain.ChainForward]
	if !ok {
		t.Fatal("catena FORWARD non trovata")
	}
	if forwardChain.DefaultPolicy != domain.ActionDrop {
		t.Errorf("attesa policy DROP per FORWARD, ottenuto: %s", forwardChain.DefaultPolicy)
	}
	if len(forwardChain.Rules) != 0 {
		t.Errorf("FORWARD doveva avere 0 regole, ottenute %d", len(forwardChain.Rules))
	}
}

func TestFirewallRejectsInvalidTransactionBeforeAcknowledgement(t *testing.T) {
	executor := &protocolFirewallExecutor{malformed: true}
	repo := newMockBackupRepo()
	svc := service.NewFirewallService(executor, repo)
	_, err := svc.AddRule(context.Background(), uuid.New(), uuid.New(), domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionAccept}, true, 60)
	if !errors.Is(err, domain.ErrSSHExecutionFailed) {
		t.Fatalf("malformed transaction acknowledged: %v", err)
	}
	if len(repo.backups) != 0 {
		t.Fatal("malformed snapshot persisted")
	}
}

func TestFirewallFailedConfirmOrRollbackKeepsWatchdog(t *testing.T) {
	for _, operation := range []string{"confirm", "rollback"} {
		t.Run(operation, func(t *testing.T) {
			executor := &protocolFirewallExecutor{}
			repo := newMockBackupRepo()
			svc := service.NewFirewallService(executor, repo)
			tenant, id := uuid.New(), uuid.New()
			if _, err := svc.ApplyWithSafetyTimer(context.Background(), tenant, id, 60); err != nil {
				t.Fatal(err)
			}
			executor.failAction = operation
			var err error
			if operation == "confirm" {
				err = svc.Confirm(context.Background(), tenant, id)
			} else {
				err = svc.Rollback(context.Background(), tenant, id, nil)
			}
			if err == nil {
				t.Fatal("expected remote error")
			}
			status, err := svc.GetStatus(context.Background(), tenant, id)
			if err != nil || !status.ActiveSafetyTimer {
				t.Fatalf("remote protection lost: %+v %v", status, err)
			}
			executor.failAction = ""
			// A fresh service instance sees and confirms the same remote state.
			if err := service.NewFirewallService(executor, repo).Confirm(context.Background(), tenant, id); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFirewallConcurrentTimerPreparation(t *testing.T) {
	executor := &protocolFirewallExecutor{}
	repo := newMockBackupRepo()
	tenant, id := uuid.New(), uuid.New()
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.NewFirewallService(executor, repo).ApplyWithSafetyTimer(context.Background(), tenant, id, 60)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	success, conflicts := 0, 0
	for err := range errs {
		if err == nil {
			success++
		} else if errors.Is(err, domain.ErrSafetyTimerActive) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("success %d conflicts %d", success, conflicts)
	}
	if err := service.NewFirewallService(executor, repo).Confirm(context.Background(), tenant, id); err != nil {
		t.Fatal(err)
	}
}

func TestBuildIptablesAddCommand(t *testing.T) {
	t.Run("Append rule with full parameters", func(t *testing.T) {
		p := domain.FirewallRuleParams{
			Chain:        domain.ChainInput,
			Action:       domain.ActionAccept,
			Protocol:     "tcp",
			Source:       "10.0.0.0/24",
			Destination:  "192.168.1.10",
			Port:         "443",
			InInterface:  "eth0",
			OutInterface: "*",
			Comment:      "HTTPS access",
			Position:     0,
		}
		cmd := service.BuildIptablesAddCommand(p)
		expected := `sudo iptables -A INPUT -p tcp --dport 443 -s 10.0.0.0/24 -d 192.168.1.10 -i eth0 -j ACCEPT -m comment --comment 'HTTPS access'`
		if cmd != expected {
			t.Errorf("comando errato:\natteso:  %s\nottenuto: %s", expected, cmd)
		}
	})

	t.Run("Insert rule at position 1", func(t *testing.T) {
		p := domain.FirewallRuleParams{
			Chain:    domain.ChainInput,
			Action:   domain.ActionDrop,
			Protocol: "all",
			Source:   "1.2.3.4",
			Position: 1,
		}
		cmd := service.BuildIptablesAddCommand(p)
		expected := `sudo iptables -I INPUT 1 -s 1.2.3.4 -j DROP`
		if cmd != expected {
			t.Errorf("comando insert errato:\natteso:  %s\nottenuto: %s", expected, cmd)
		}
	})
}

func TestFirewallMutationCannotDisableSafetyTimer(t *testing.T) {
	for _, remove := range []bool{false, true} {
		executor := &protocolFirewallExecutor{rejectSnapshot: true}
		svc := service.NewFirewallService(executor, newMockBackupRepo())
		tenant, server := uuid.New(), uuid.New()
		var err error
		if remove {
			_, err = svc.DeleteRule(context.Background(), tenant, server, domain.ChainInput, 1, false, 60)
		} else {
			_, err = svc.AddRule(context.Background(), tenant, server, domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionAccept}, false, 60)
		}
		if err == nil {
			t.Fatal("mutation bypassed failed safety snapshot")
		}
		if len(executor.commands) != 1 || !strings.Contains(executor.commands[0], "nextgen-firewall 'apply'") {
			t.Fatal("legacy false flag bypassed the atomic protection request")
		}
	}
}

func TestFirewallRejectsInjectionBeforeSSH(t *testing.T) {
	executor := &mockSSHExecutor{}
	svc := service.NewFirewallService(executor, newMockBackupRepo())
	_, err := svc.AddRule(context.Background(), uuid.New(), uuid.New(), domain.FirewallRuleParams{
		Chain: domain.ChainInput, Action: domain.ActionAccept, Source: "192.0.2.1; id",
	}, true, 60)
	if !errors.Is(err, domain.ErrValidation) || len(executor.commandLogs) != 0 {
		t.Fatalf("invalid input reached SSH: %v %v", err, executor.commandLogs)
	}
}

func TestFirewallCommentPreservesShellMetacharactersAsData(t *testing.T) {
	params := domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionAccept, Comment: "operator's $(id) `id` $HOME \\\"quoted\\\""}
	if err := params.Validate(); err != nil {
		t.Fatal(err)
	}
	got := service.BuildIptablesAddCommand(params)
	want := `sudo iptables -A INPUT -j ACCEPT -m comment --comment 'operator'"'"'s $(id) ` + "`id`" + ` $HOME \"quoted\"'`
	if got != want {
		t.Fatalf("comment quoting:\n%s\nwant:\n%s", got, want)
	}
}

func TestFirewallRejectsInvalidTimeoutBeforeSSH(t *testing.T) {
	for _, timeout := range []int{-1, 1, 9, 301, 2147483647} {
		executor := &mockSSHExecutor{}
		svc := service.NewFirewallService(executor, newMockBackupRepo())
		_, err := svc.ApplyWithSafetyTimer(context.Background(), uuid.New(), uuid.New(), timeout)
		if !errors.Is(err, domain.ErrValidation) || len(executor.commandLogs) != 0 {
			t.Fatalf("timeout %d: err %v, commands %v", timeout, err, executor.commandLogs)
		}
	}
}

func TestFirewallService_GetStatus(t *testing.T) {
	svc := service.NewFirewallService(&protocolFirewallExecutor{}, newMockBackupRepo())
	status, err := svc.GetStatus(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if status.ActiveSafetyTimer || len(status.Chains) != 3 {
		t.Fatalf("incorrect remote status: %+v", status)
	}
}

func TestFirewallStatusUnattendedAccess(t *testing.T) {
	for _, tc := range []struct {
		name, stderr, expected string
		code                   int
	}{
		{"password required", "sudo: a password is required", "privilegi", 1},
		{"not allowed", "alex is not in the sudoers file", "privilegi", 1},
		{"missing iptables", "", "iptables", 127},
		{"missing sudo", "", "privilegi", 126},
	} {
		t.Run(tc.name, func(t *testing.T) {
			executor := &mockSSHExecutor{executeFunc: func(_ context.Context, _, _ uuid.UUID, cmd string) (string, string, int, error) {
				if !strings.Contains(cmd, "sudo -n") || !strings.Contains(cmd, "id -u") || !strings.Contains(cmd, "/usr/sbin") {
					t.Error("firewall status does not support root/PATH/noninteractive sudo")
				}
				return "", tc.stderr, tc.code, errors.New("exit status")
			}}
			_, err := service.NewFirewallService(executor, newMockBackupRepo()).GetStatus(context.Background(), uuid.New(), uuid.New())
			if err == nil || !strings.Contains(err.Error(), tc.expected) {
				t.Fatalf("missing actionable diagnostic: %v", err)
			}
		})
	}
}

func TestFirewallService_AddRule_WithSafetyTimer_And_Confirm(t *testing.T) {
	executor := &protocolFirewallExecutor{}
	repo := newMockBackupRepo()
	svc := service.NewFirewallService(executor, repo)
	tenant, server := uuid.New(), uuid.New()
	status, err := svc.AddRule(context.Background(), tenant, server, domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionAccept, Protocol: "tcp", Port: "8080"}, true, 60)
	if err != nil || !status.ActiveSafetyTimer || status.SafetyTimeoutSeconds != 60 {
		t.Fatalf("protection not active: %+v %v", status, err)
	}
	latest, err := repo.GetLatestBackup(context.Background(), tenant, server)
	if err != nil || !strings.Contains(latest.RawIptablesRules, "*filter") || latest.ID != executor.token {
		t.Fatalf("transaction backup not persisted: %+v %v", latest, err)
	}
	if _, err := svc.ApplyWithSafetyTimer(context.Background(), tenant, server, 60); !errors.Is(err, domain.ErrSafetyTimerActive) {
		t.Fatalf("overlap accepted: %v", err)
	}
	if err := svc.Confirm(context.Background(), tenant, server); err != nil {
		t.Fatal(err)
	}
	status, err = svc.GetStatus(context.Background(), tenant, server)
	if err != nil || status.ActiveSafetyTimer {
		t.Fatal("confirmed state remained pending")
	}
	if err := svc.Confirm(context.Background(), tenant, server); !errors.Is(err, domain.ErrNoSafetyTimerActive) {
		t.Fatalf("duplicate confirm accepted: %v", err)
	}
}

func TestFirewallService_Rollback(t *testing.T) {
	executor := &protocolFirewallExecutor{}
	repo := newMockBackupRepo()
	svc := service.NewFirewallService(executor, repo)
	tenant, server := uuid.New(), uuid.New()
	backup := &domain.FirewallBackup{ID: uuid.New(), TenantID: tenant, ServerID: server, RawIptablesRules: "*filter\n# snapshot data with EOF and $(id)\n:INPUT ACCEPT [0:0]\nCOMMIT\n", Label: "Manual fixture", CreatedAt: time.Now().UTC()}
	_ = repo.CreateBackup(context.Background(), backup)
	if err := svc.Rollback(context.Background(), tenant, server, &backup.ID); err != nil {
		t.Fatal(err)
	}
	if len(executor.inputs) != 2 || executor.inputs[1] != backup.RawIptablesRules {
		t.Fatal("snapshot was not sent as stdin data")
	}
	if strings.Contains(executor.commands[1], backup.RawIptablesRules) {
		t.Fatal("snapshot injected into shell source")
	}
}
