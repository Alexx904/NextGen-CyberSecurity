package service_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

// Models the remote protocol only. Real shell/iptables/expiry behavior is checked
// independently against the explicitly named Docker Linux fixture.
type protocolFirewallExecutor struct {
	mu                        sync.Mutex
	token, tenant, server     uuid.UUID
	expires                   time.Time
	timeout                   int
	failAction                string
	malformed, rejectSnapshot bool
	commands, inputs          []string
}

type rejectedFirewallBackup struct {
	domain.FirewallBackupRepository
}

func (rejectedFirewallBackup) CreateBackup(context.Context, *domain.FirewallBackup) error {
	return errors.New("synthetic storage failure")
}

func TestFirewallBackupFailureLeavesRemoteProtection(t *testing.T) {
	executor := &protocolFirewallExecutor{}
	repo := newMockBackupRepo()
	svc := service.NewFirewallService(executor, rejectedFirewallBackup{repo})
	tenant, server := uuid.New(), uuid.New()
	if _, err := svc.ApplyWithSafetyTimer(context.Background(), tenant, server, 60); err == nil {
		t.Fatal("backup failure acknowledged")
	}
	status, err := svc.GetStatus(context.Background(), tenant, server)
	if err != nil || !status.ActiveSafetyTimer {
		t.Fatal("backup failure disarmed remote protection")
	}
	if err := svc.Confirm(context.Background(), tenant, server); !errors.Is(err, domain.ErrFirewallBackupNotFound) {
		t.Fatalf("unpersisted transaction confirmed: %v", err)
	}
	if executor.token == uuid.Nil {
		t.Fatal("failed confirmation disarmed remote protection")
	}
}

func TestFirewallRejectsForeignBackupBeforeRestore(t *testing.T) {
	executor := &protocolFirewallExecutor{}
	repo := newMockBackupRepo()
	tenant, server := uuid.New(), uuid.New()
	backup := &domain.FirewallBackup{ID: uuid.New(), TenantID: tenant, ServerID: uuid.New(), RawIptablesRules: "*filter\nCOMMIT\n", CreatedAt: time.Now()}
	_ = repo.CreateBackup(context.Background(), backup)
	if err := service.NewFirewallService(executor, repo).Rollback(context.Background(), tenant, server, &backup.ID); !errors.Is(err, domain.ErrFirewallBackupNotFound) {
		t.Fatal("foreign server snapshot restored")
	}
	if len(executor.commands) != 1 || executor.inputs[0] != "" {
		t.Fatal("foreign snapshot reached restore")
	}
}

func TestFirewallEditorRejectsInvalidOrStaleRequests(t *testing.T) {
	executor := &protocolFirewallExecutor{}
	svc := service.NewFirewallService(executor, newMockBackupRepo())
	editor := svc.(service.FirewallRuleEditor)
	tenant, server := uuid.New(), uuid.New()
	revision := strings.Repeat("a", 64)
	checks := []func() error{
		func() error {
			_, err := editor.SetPolicy(context.Background(), tenant, server, domain.ChainInput, domain.ActionReject, revision, 60)
			return err
		},
		func() error {
			_, err := editor.MoveRule(context.Background(), tenant, server, domain.ChainInput, 1, 1, revision, 60)
			return err
		},
		func() error {
			_, err := editor.MoveRule(context.Background(), tenant, server, domain.ChainInput, 1, 2, "", 60)
			return err
		},
		func() error {
			_, err := editor.ReplaceRule(context.Background(), tenant, server, 0, domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionDrop}, revision, 60)
			return err
		},
	}
	for _, check := range checks {
		if err := check(); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("invalid editor request accepted: %v", err)
		}
	}
	if len(executor.commands) != 0 {
		t.Fatal("invalid editor request reached SSH")
	}
	if _, err := editor.SetPolicy(context.Background(), tenant, server, domain.ChainInput, domain.ActionDrop, strings.Repeat("b", 64), 60); !errors.Is(err, domain.ErrFirewallStaleState) {
		t.Fatalf("stale mutation not rejected: %v", err)
	}
	if executor.token != uuid.Nil {
		t.Fatal("stale request armed a transaction")
	}
}

func (m *protocolFirewallExecutor) ExecuteCommand(c context.Context, t, s uuid.UUID, command string) (string, string, int, error) {
	return m.ExecuteCommandWithInput(c, t, s, command, "")
}
func (m *protocolFirewallExecutor) ExecuteCommandWithInput(_ context.Context, _, _ uuid.UUID, command, input string) (string, string, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commands = append(m.commands, command)
	m.inputs = append(m.inputs, input)
	index := strings.LastIndex(command, " nextgen-firewall ")
	if index < 0 {
		return "", "missing protocol", 64, nil
	}
	fields := strings.Fields(command[index+len(" nextgen-firewall "):])
	if len(fields) < 5 {
		return "", "missing arguments", 64, nil
	}
	action := strings.Trim(fields[0], "'")
	tenant, e1 := uuid.Parse(fields[1])
	server, e2 := uuid.Parse(fields[2])
	token, e3 := uuid.Parse(fields[3])
	timeout, e4 := strconv.Atoi(fields[4])
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		return "", "invalid arguments", 64, nil
	}
	if action == m.failAction {
		return "", "synthetic remote failure", 74, nil
	}
	switch action {
	case "status":
		if m.token == uuid.Nil {
			return "NONE\nREVISION " + strings.Repeat("a", 64) + "\nRULES\n" + sampleIptablesOutput, "", 0, nil
		}
		return fmt.Sprintf("PENDING %s %s %s %d %d\nREVISION %s\nRULES\n%s", m.token, m.tenant, m.server, m.expires.Unix(), m.timeout, strings.Repeat("a", 64), sampleIptablesOutput), "", 0, nil
	case "apply":
		if m.token != uuid.Nil {
			return "", "", 73, nil
		}
		if len(fields) > 5 {
			expected := strings.Trim(fields[5], "'")
			if expected != "" && expected != strings.Repeat("a", 64) {
				return "", "", 76, nil
			}
		}
		if m.rejectSnapshot {
			return "", "snapshot refused", 126, nil
		}
		if m.malformed {
			return "BACKUP not-a-transaction 1 60\n*filter\nCOMMIT\n", "", 0, nil
		}
		m.token, m.tenant, m.server, m.timeout, m.expires = token, tenant, server, timeout, time.Now().Add(time.Duration(timeout)*time.Second)
		return fmt.Sprintf("BACKUP %s %d %d\n*filter\n:INPUT ACCEPT [0:0]\nCOMMIT\n", token, m.expires.Unix(), timeout), "", 0, nil
	case "confirm", "rollback":
		if m.token != uuid.Nil && (m.tenant != tenant || m.server != server) {
			return "", "", 73, nil
		}
		if action == "confirm" && (m.token == uuid.Nil || m.token != token || !time.Now().Before(m.expires)) {
			return "", "", 75, nil
		}
		if action == "rollback" && input == "" {
			return "", "missing snapshot input", 74, nil
		}
		m.token = uuid.Nil
		return "OK\n", "", 0, nil
	}
	return "", "unsupported action", 64, nil
}
