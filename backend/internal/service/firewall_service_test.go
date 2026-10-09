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
	mu           sync.Mutex
	commandLogs  []string
	executeFunc  func(ctx context.Context, tenantID, serverID uuid.UUID, cmd string) (string, string, int, error)
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
		expected := `sudo iptables -A INPUT -p tcp --dport 443 -s 10.0.0.0/24 -d 192.168.1.10 -i eth0 -j ACCEPT -m comment --comment "HTTPS access"`
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

func TestFirewallService_GetStatus(t *testing.T) {
	executor := &mockSSHExecutor{
		executeFunc: func(ctx context.Context, tenantID, serverID uuid.UUID, cmd string) (string, string, int, error) {
			if strings.Contains(cmd, "iptables -L") {
				return sampleIptablesOutput, "", 0, nil
			}
			return "", "", 0, nil
		},
	}
	repo := newMockBackupRepo()
	svc := service.NewFirewallService(executor, repo)

	tenantID := uuid.New()
	serverID := uuid.New()

	status, err := svc.GetStatus(context.Background(), tenantID, serverID)
	if err != nil {
		t.Fatalf("GetStatus fallito: %v", err)
	}

	if status.ActiveSafetyTimer {
		t.Errorf("nessun safety timer doveva essere attivo")
	}
	if len(status.Chains) != 3 {
		t.Errorf("attese 3 catene, ottenute %d", len(status.Chains))
	}
}

func TestFirewallService_AddRule_WithSafetyTimer_And_Confirm(t *testing.T) {
	executor := &mockSSHExecutor{
		executeFunc: func(ctx context.Context, tenantID, serverID uuid.UUID, cmd string) (string, string, int, error) {
			if strings.Contains(cmd, "iptables-save") {
				return "# Raw rules\n*filter\n:INPUT ACCEPT [0:0]\nCOMMIT\n", "", 0, nil
			}
			if strings.Contains(cmd, "echo $!") {
				return "98765\n", "", 0, nil
			}
			if strings.Contains(cmd, "iptables -L") {
				return sampleIptablesOutput, "", 0, nil
			}
			return "", "", 0, nil
		},
	}
	repo := newMockBackupRepo()
	svc := service.NewFirewallService(executor, repo)

	tenantID := uuid.New()
	serverID := uuid.New()

	params := domain.FirewallRuleParams{
		Chain:    domain.ChainInput,
		Action:   domain.ActionAccept,
		Protocol: "tcp",
		Port:     "8080",
	}

	// 1. Add rule with safety timer
	status, err := svc.AddRule(context.Background(), tenantID, serverID, params, true, 60)
	if err != nil {
		t.Fatalf("AddRule fallito: %v", err)
	}

	if !status.ActiveSafetyTimer {
		t.Errorf("il safety timer doveva risultare attivo")
	}
	if status.SafetyTimeoutSeconds != 60 {
		t.Errorf("atteso timeout di 60 secondi, ottenuto: %d", status.SafetyTimeoutSeconds)
	}

	// Verifica che lo snapshot sia stato salvato nel repository
	latestBackup, err := repo.GetLatestBackup(context.Background(), tenantID, serverID)
	if err != nil {
		t.Fatalf("GetLatestBackup fallito: %v", err)
	}
	if !strings.Contains(latestBackup.RawIptablesRules, "*filter") {
		t.Errorf("contenuto backup non conforme: %s", latestBackup.RawIptablesRules)
	}

	// 2. Tentativo di avviare un secondo safety timer sullo stesso server deve dare ErrSafetyTimerActive
	_, err = svc.ApplyWithSafetyTimer(context.Background(), tenantID, serverID, 60)
	if !errors.Is(err, domain.ErrSafetyTimerActive) {
		t.Errorf("atteso ErrSafetyTimerActive, ottenuto: %v", err)
	}

	// 3. Confirm
	if err := svc.Confirm(context.Background(), tenantID, serverID); err != nil {
		t.Fatalf("Confirm fallito: %v", err)
	}

	// 4. Dopo confirm, GetStatus deve mostrare timer non più attivo
	statusAfter, _ := svc.GetStatus(context.Background(), tenantID, serverID)
	if statusAfter.ActiveSafetyTimer {
		t.Errorf("dopo confirm il timer non deve essere attivo")
	}

	// 5. Secondo Confirm deve restituire ErrNoSafetyTimerActive
	err = svc.Confirm(context.Background(), tenantID, serverID)
	if !errors.Is(err, domain.ErrNoSafetyTimerActive) {
		t.Errorf("atteso ErrNoSafetyTimerActive, ottenuto: %v", err)
	}
}

func TestFirewallService_Rollback(t *testing.T) {
	restoreExecuted := false
	executor := &mockSSHExecutor{
		executeFunc: func(ctx context.Context, tenantID, serverID uuid.UUID, cmd string) (string, string, int, error) {
			if strings.Contains(cmd, "iptables-restore") {
				restoreExecuted = true
				return "", "", 0, nil
			}
			if strings.Contains(cmd, "iptables-save") {
				return "*filter\n:INPUT ACCEPT [0:0]\nCOMMIT\n", "", 0, nil
			}
			if strings.Contains(cmd, "iptables -L") {
				return sampleIptablesOutput, "", 0, nil
			}
			return "", "", 0, nil
		},
	}
	repo := newMockBackupRepo()
	svc := service.NewFirewallService(executor, repo)

	tenantID := uuid.New()
	serverID := uuid.New()

	// Inserimento manuale backup di test
	backup := &domain.FirewallBackup{
		ID:               uuid.New(),
		TenantID:         tenantID,
		ServerID:         serverID,
		RawIptablesRules: "*filter\n:INPUT ACCEPT [0:0]\nCOMMIT\n",
		Label:            "Manual initial snapshot",
		CreatedAt:        time.Now().UTC(),
	}
	_ = repo.CreateBackup(context.Background(), backup)

	// Esecuzione rollback esplicito
	err := svc.Rollback(context.Background(), tenantID, serverID, &backup.ID)
	if err != nil {
		t.Fatalf("Rollback fallito: %v", err)
	}

	if !restoreExecuted {
		t.Errorf("il comando iptables-restore doveva essere eseguito")
	}
}

