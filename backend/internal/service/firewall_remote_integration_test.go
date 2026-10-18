//go:build integration

package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
)

// These tests mutate only an explicitly named, disposable Docker fixture.
// They do not run as part of the PostgreSQL integration package.
func runFirewallFixture(t *testing.T, action string, owner, server, token uuid.UUID, input string, args ...string) (string, int) {
	t.Helper()
	container := os.Getenv("TEST_FIREWALL_CONTAINER")
	if container != "nextgen-firewall-lab-20261018" {
		t.Fatal("TEST_FIREWALL_CONTAINER must name the dedicated disposable fixture")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := []string{"exec", "-i", container, "sh", "-c", remoteFirewallScript, "nextgen-firewall", action, owner.String(), server.String(), token.String(), "10"}
	if action == "apply" {
		command = append(command, "")
	}
	command = append(command, args...)
	cmd := exec.CommandContext(ctx, "docker", command...)
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return string(output), e.ExitCode()
		}
		t.Fatal(err)
	}
	return string(output), 0
}

type fixtureFirewallExecutor struct{}

func (fixtureFirewallExecutor) ExecuteCommand(ctx context.Context, tenant, server uuid.UUID, command string) (string, string, int, error) {
	return fixtureFirewallExecutor{}.ExecuteCommandWithInput(ctx, tenant, server, command, "")
}
func (fixtureFirewallExecutor) ExecuteCommandWithInput(ctx context.Context, _, _ uuid.UUID, command, input string) (string, string, int, error) {
	if os.Getenv("TEST_FIREWALL_CONTAINER") != "nextgen-firewall-lab-20261018" {
		return "", "", -1, errors.New("dedicated firewall fixture is required")
	}
	cmd := exec.CommandContext(ctx, "docker", "exec", "-i", "nextgen-firewall-lab-20261018", "sh", "-c", command)
	cmd.Stdin = strings.NewReader(input)
	var out, errOut strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	code := 0
	if err != nil {
		code = -1
		if e, ok := err.(*exec.ExitError); ok {
			code = e.ExitCode()
		}
	}
	return out.String(), errOut.String(), code, err
}

type fixtureFirewallBackups struct {
	mu    sync.Mutex
	items map[uuid.UUID]*domain.FirewallBackup
}

func (r *fixtureFirewallBackups) CreateBackup(_ context.Context, b *domain.FirewallBackup) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[b.ID] = b
	return nil
}
func (r *fixtureFirewallBackups) GetBackupByID(_ context.Context, tenant, id uuid.UUID) (*domain.FirewallBackup, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b := r.items[id]
	if b == nil || b.TenantID != tenant {
		return nil, domain.ErrFirewallBackupNotFound
	}
	return b, nil
}
func (r *fixtureFirewallBackups) GetLatestBackup(_ context.Context, tenant, server uuid.UUID) (*domain.FirewallBackup, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var latest *domain.FirewallBackup
	for _, b := range r.items {
		if b.TenantID == tenant && b.ServerID == server && (latest == nil || b.CreatedAt.After(latest.CreatedAt)) {
			latest = b
		}
	}
	if latest == nil {
		return nil, domain.ErrFirewallBackupNotFound
	}
	return latest, nil
}
func (r *fixtureFirewallBackups) ListBackups(context.Context, uuid.UUID, uuid.UUID) ([]*domain.FirewallBackup, error) {
	return nil, nil
}
func (r *fixtureFirewallBackups) DeleteBackup(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func TestFirewallServiceLinuxMultiInstance(t *testing.T) {
	if os.Getenv("TEST_FIREWALL_CONTAINER") != "nextgen-firewall-lab-20261018" {
		t.Fatal("dedicated firewall fixture is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	owner, server := uuid.New(), uuid.New()
	backups := &fixtureFirewallBackups{items: make(map[uuid.UUID]*domain.FirewallBackup)}
	first := NewFirewallService(fixtureFirewallExecutor{}, backups)
	second := NewFirewallService(fixtureFirewallExecutor{}, backups)
	if _, err := first.ApplyWithSafetyTimer(ctx, owner, server, 10); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = first.Confirm(context.Background(), owner, server)
		_ = second.Confirm(context.Background(), owner, server)
	})
	if _, err := second.ApplyWithSafetyTimer(ctx, owner, server, 10); !errors.Is(err, domain.ErrSafetyTimerActive) {
		t.Fatalf("second process did not reject overlap: %v", err)
	}
	if err := second.Confirm(ctx, owner, server); err != nil {
		t.Fatalf("new service could not confirm remote transaction: %v", err)
	}
	status, err := first.GetStatus(ctx, owner, server)
	if err != nil || status.ActiveSafetyTimer {
		t.Fatalf("first process retained stale local timer: %+v %v", status, err)
	}
}

func TestFirewallServiceLinuxEditor(t *testing.T) {
	if os.Getenv("TEST_FIREWALL_CONTAINER") != "nextgen-firewall-lab-20261018" {
		t.Fatal("dedicated firewall fixture is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	owner, server := uuid.New(), uuid.New()
	repo := &fixtureFirewallBackups{items: make(map[uuid.UUID]*domain.FirewallBackup)}
	svc := NewFirewallService(fixtureFirewallExecutor{}, repo)
	editor := svc.(FirewallRuleEditor)
	status, err := svc.GetStatus(ctx, owner, server)
	if err != nil || !validFirewallRevision(status.Revision) {
		t.Fatalf("revision missing: %v", err)
	}
	initialRevision := status.Revision
	first := domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionAccept, Protocol: "tcp", Port: "45670", Comment: "first fixture"}
	status, err = editor.AddRuleWithRevision(ctx, owner, server, first, status.Revision, 10)
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := repo.GetLatestBackup(ctx, owner, server)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := svc.Rollback(context.Background(), owner, server, &baseline.ID); err != nil {
			t.Error("fixture cleanup failed", err)
		}
	})
	if err := svc.Confirm(ctx, owner, server); err != nil {
		t.Fatal(err)
	}
	marker := "/tmp/nextgen-injection-" + uuid.NewString()
	second := domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionDrop, Protocol: "tcp", Port: "45671", Comment: "operator's $(touch " + marker + ")"}
	status, err = editor.AddRuleWithRevision(ctx, owner, server, second, status.Revision, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Confirm(ctx, owner, server); err != nil {
		t.Fatal(err)
	}
	if _, err := editor.MoveRule(ctx, owner, server, domain.ChainInput, 2, 1, initialRevision, 10); !errors.Is(err, domain.ErrFirewallStaleState) {
		t.Fatalf("stale revision accepted: %v", err)
	}
	status, err = svc.GetStatus(ctx, owner, server)
	if err != nil {
		t.Fatal(err)
	}
	status, err = editor.MoveRule(ctx, owner, server, domain.ChainInput, 2, 1, status.Revision, 10)
	if err != nil {
		t.Fatal(err)
	}
	rules := status.Chains[domain.ChainInput].Rules
	if len(rules) != 2 || rules[0].Target != domain.ActionDrop || rules[1].Target != domain.ActionAccept {
		t.Fatalf("rule order not preserved: %+v", rules)
	}
	if err := svc.Confirm(ctx, owner, server); err != nil {
		t.Fatal(err)
	}
	_, _, code, err := fixtureFirewallExecutor{}.ExecuteCommand(ctx, owner, server, "test ! -e "+marker)
	if err != nil || code != 0 {
		t.Fatal("comment executed as shell source")
	}
	first.Port = "45672"
	status, err = editor.ReplaceRule(ctx, owner, server, 1, first, status.Revision, 10)
	if err != nil {
		t.Fatal(err)
	}
	if status.Chains[domain.ChainInput].Rules[0].Target != domain.ActionAccept {
		t.Fatal("replacement did not change target")
	}
	if err := svc.Confirm(ctx, owner, server); err != nil {
		t.Fatal(err)
	}
	status, err = editor.SetPolicy(ctx, owner, server, domain.ChainForward, domain.ActionDrop, status.Revision, 10)
	if err != nil {
		t.Fatal(err)
	}
	if status.Chains[domain.ChainForward].DefaultPolicy != domain.ActionDrop {
		t.Fatal("policy not changed")
	}
	if err := NewFirewallService(fixtureFirewallExecutor{}, repo).Confirm(ctx, owner, server); err != nil {
		t.Fatal(err)
	}
}

func TestRemoteFirewallLinuxTransaction(t *testing.T) {
	owner, server, token := uuid.New(), uuid.New(), uuid.New()
	output, code := runFirewallFixture(t, "apply", owner, server, token, "", "-A", "INPUT", "-p", "tcp", "--dport", "45678", "-j", "DROP")
	if code != 0 {
		t.Fatalf("apply exit=%d: %s", code, output)
	}
	backup, _, err := parseRemoteFirewallBackup(output, token)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { runFirewallFixture(t, "rollback", owner, server, token, backup) })
	stateOutput, code := runFirewallFixture(t, "status", owner, server, uuid.New(), "")
	state, rules, err := parseRemoteFirewallState(stateOutput)
	if code != 0 || err != nil || state == nil || state.ID != token || !strings.Contains(rules, "45678") {
		t.Fatalf("state not durable: %s err=%v", stateOutput, err)
	}
	if _, code := runFirewallFixture(t, "apply", owner, server, uuid.New(), "", "-A", "INPUT", "-j", "ACCEPT"); code != 73 {
		t.Fatalf("overlapping instance accepted: exit=%d", code)
	}
	if _, code := runFirewallFixture(t, "confirm", uuid.New(), server, token, ""); code != 73 {
		t.Fatal("foreign owner confirmed pending transaction")
	}
	// A separate launcher can confirm without any in-memory PID/timer state.
	if output, code := runFirewallFixture(t, "confirm", owner, server, token, ""); code != 0 {
		t.Fatalf("second instance confirm failed: %d %s", code, output)
	}
	output, code = runFirewallFixture(t, "status", owner, server, uuid.New(), "")
	state, _, err = parseRemoteFirewallState(output)
	if code != 0 || err != nil || state != nil {
		t.Fatal("confirmed transaction remained pending")
	}
}

func TestRemoteFirewallLinuxExpiryAndFailedRollback(t *testing.T) {
	owner, server, token := uuid.New(), uuid.New(), uuid.New()
	output, code := runFirewallFixture(t, "apply", owner, server, token, "", "-A", "FORWARD", "-p", "udp", "--dport", "45679", "-j", "DROP")
	if code != 0 {
		t.Fatalf("apply failed: %d %s", code, output)
	}
	backup, _, err := parseRemoteFirewallBackup(output, token)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { runFirewallFixture(t, "rollback", owner, server, token, backup) })
	if _, code := runFirewallFixture(t, "rollback", owner, server, token, "invalid rules\n"); code != 74 {
		t.Fatal("failed restore reported success")
	}
	output, code = runFirewallFixture(t, "status", owner, server, uuid.New(), "")
	state, _, err := parseRemoteFirewallState(output)
	if code != 0 || err != nil || state == nil || state.ID != token {
		t.Fatal("failed restore disarmed watchdog")
	}
	deadline := time.Now().Add(13 * time.Second)
	for time.Now().Before(deadline) {
		output, code = runFirewallFixture(t, "status", owner, server, uuid.New(), "")
		state, rules, err := parseRemoteFirewallState(output)
		if code == 0 && err == nil && state == nil {
			if strings.Contains(rules, "45679") {
				t.Fatal("expired rule not restored")
			}
			if _, code := runFirewallFixture(t, "confirm", owner, server, token, ""); code != 75 {
				t.Fatal("stale confirmation accepted")
			}
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal(fmt.Sprintf("watchdog did not restore after launcher exit: %s", output))
}
