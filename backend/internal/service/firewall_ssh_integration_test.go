//go:build integration

package service

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/crypto"
)

type fixtureSSHInventory struct {
	domain.SSHServerRepository
	server *domain.SSHServer
}

func fixtureSSHCommand(t *testing.T, ctx context.Context, args ...string) *exec.Cmd {
	t.Helper()
	if os.Getenv("TEST_FIREWALL_SSH_IN_CONTAINER") == "1" {
		hostname, err := os.Hostname()
		if err != nil || hostname != "nextgen-ssh-fixture" {
			t.Fatal("local fixture commands require the named disposable Linux container")
		}
		return exec.CommandContext(ctx, args[0], args[1:]...)
	}
	command := append([]string{"exec", "-i", "nextgen-ssh-lab-20261018"}, args...)
	return exec.CommandContext(ctx, "docker", command...)
}

func (r *fixtureSSHInventory) GetByID(_ context.Context, tenant, id uuid.UUID) (*domain.SSHServer, error) {
	if r.server.TenantID != tenant || r.server.ID != id {
		return nil, domain.ErrSSHServerNotFound
	}
	return r.server, nil
}

func firewallSSHFixture(t *testing.T) (SSHService, uuid.UUID, uuid.UUID) {
	t.Helper()
	expectedPort := "55222"
	if os.Getenv("TEST_FIREWALL_SSH_IN_CONTAINER") == "1" {
		expectedPort = "22"
	}
	if os.Getenv("TEST_FIREWALL_SSH_CONTAINER") != "nextgen-ssh-lab-20261018" || os.Getenv("TEST_FIREWALL_SSH_PORT") != expectedPort {
		t.Fatal("explicit dedicated loopback SSH fixture is required")
	}
	// Obtain only a public host-key fingerprint over the trusted Docker control
	// channel. No key scanning, TOFU or user SSH credentials are involved.
	identityCtx, identityCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer identityCancel()
	output, err := fixtureSSHCommand(t, identityCtx, "ssh-keygen", "-lf", "/etc/ssh/ssh_host_ed25519_key.pub", "-E", "sha256").Output()
	if err != nil {
		t.Fatal("fixture public host identity unavailable")
	}
	fields := strings.Fields(string(output))
	if len(fields) < 2 || !domain.ValidHostKeyFingerprint(fields[1]) {
		t.Fatal("invalid fixture public host fingerprint")
	}
	tenant, server := uuid.New(), uuid.New()
	master := strings.Repeat("t", 32)
	key, err := crypto.DeriveTenantKey([]byte(master), tenant)
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := crypto.EncryptAES256GCM([]byte("synthetic-lab-only"), key, []byte(tenant.String()))
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(expectedPort)
	inventory := &fixtureSSHInventory{server: &domain.SSHServer{ID: server, TenantID: tenant, Name: "Disposable SSH fixture", Host: "127.0.0.1", Port: port, Username: "root", AuthType: domain.SSHAuthTypePassword, EncryptedSecret: encrypted, HostKeyFingerprint: fields[1]}}
	svc := NewSSHService(inventory, master)
	probeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := svc.GetRawSSHClient(probeCtx, tenant, server)
	if err != nil {
		t.Fatalf("dedicated SSH fixture unavailable: %v", err)
	}
	_ = client.Close()
	return svc, tenant, server
}

func TestFirewallSSHDisconnectAndRemoteExpiry(t *testing.T) {
	executor, tenant, server := firewallSSHFixture(t)
	repo := &fixtureFirewallBackups{items: make(map[uuid.UUID]*domain.FirewallBackup)}
	svc := NewFirewallService(executor, repo)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	initial, err := svc.GetStatus(ctx, tenant, server)
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range initial.Chains[domain.ChainInput].Rules {
		if strings.Contains(rule.Options, "45680") {
			t.Fatal("fixture baseline already contains test rule 45680")
		}
	}
	status, err := svc.AddRule(ctx, tenant, server, domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionDrop, Protocol: "tcp", Port: "45680"}, false, 10)
	if err != nil || !status.ActiveSafetyTimer {
		t.Fatalf("SSH protected mutation failed: %+v %v", status, err)
	}
	baseline, err := repo.GetLatestBackup(ctx, tenant, server)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := svc.Rollback(context.Background(), tenant, server, &baseline.ID); err != nil {
			t.Error("SSH fixture cleanup", err)
		}
	})
	// ExecuteCommand has already closed its SSH connection. A fresh service has
	// no process timer/PID and still sees the host-owned pending transaction.
	second := NewFirewallService(executor, repo)
	status, err = second.GetStatus(ctx, tenant, server)
	if err != nil || !status.ActiveSafetyTimer {
		t.Fatal("SSH disconnect lost remote protection")
	}
	deadline := time.Now().Add(13 * time.Second)
	for time.Now().Before(deadline) {
		status, err = second.GetStatus(ctx, tenant, server)
		if err != nil {
			t.Fatal(err)
		}
		if !status.ActiveSafetyTimer {
			for _, rule := range status.Chains[domain.ChainInput].Rules {
				if strings.Contains(rule.Options, "45680") {
					t.Fatal("SSH watchdog did not restore")
				}
			}
			if err := second.Confirm(ctx, tenant, server); !errors.Is(err, domain.ErrNoSafetyTimerActive) {
				t.Fatal("late SSH confirmation accepted")
			}
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("SSH watchdog did not finish after disconnect")
}

func TestFirewallSSHTimeoutCannotLeaveLateMutation(t *testing.T) {
	executor, tenant, server := firewallSSHFixture(t)
	repo := &fixtureFirewallBackups{items: make(map[uuid.UUID]*domain.FirewallBackup)}
	svc := NewFirewallService(executor, repo).(*firewallService)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	baseline, _, code, err := executor.ExecuteCommand(ctx, tenant, server, "iptables-save")
	if err != nil || code != 0 {
		t.Fatal("fixture baseline unavailable")
	}
	// Only this disposable fixture is instrumented: delay the actual -A command
	// after the watchdog is ready, allowing the real SSH caller to time out first.
	wrapper := `#!/bin/sh
if [ "${3:-}" = "-A" ]; then
    trap '' HUP TERM
    sleep 4
    date +%s > /tmp/nextgen-late-mutation-observed
fi
exec /usr/sbin/iptables-nft "$@"
`
	cmd := fixtureSSHCommand(t, ctx, "sh", "-c", "rm -f /tmp/nextgen-late-mutation-observed && rm /usr/sbin/iptables && cat > /usr/sbin/iptables && chmod 755 /usr/sbin/iptables")
	cmd.Stdin = strings.NewReader(wrapper)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture instrumentation failed: %s %v", output, err)
	}
	t.Cleanup(func() {
		cmd := fixtureSSHCommand(t, context.Background(), "sh", "-c", "rm /usr/sbin/iptables && ln -s /etc/alternatives/iptables /usr/sbin/iptables")
		if err := cmd.Run(); err != nil {
			t.Error("fixture executable cleanup", err)
		}
		state, _, err := svc.readRemoteStatus(context.Background(), tenant, server)
		if err != nil {
			t.Error("fixture state cleanup", err)
			return
		}
		token := uuid.New()
		if state != nil {
			token = state.ID
		}
		if _, err := svc.remoteCall(context.Background(), tenant, server, token, 0, "rollback", baseline); err != nil {
			t.Error("fixture rules cleanup", err)
		}
	})
	callCtx, callCancel := context.WithTimeout(ctx, time.Second)
	defer callCancel()
	_, err = svc.AddRule(callCtx, tenant, server, domain.FirewallRuleParams{Chain: domain.ChainInput, Action: domain.ActionDrop, Protocol: "tcp", Port: "45681"}, true, 10)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("local timeout not preserved: %v", err)
	}
	if len(repo.items) != 0 {
		t.Fatal("timed-out mutation acknowledged a database backup")
	}
	deadline := time.Now().Add(13 * time.Second)
	for time.Now().Before(deadline) {
		status, err := svc.GetStatus(ctx, tenant, server)
		if err != nil {
			t.Fatal(err)
		}
		if !status.ActiveSafetyTimer {
			for _, rule := range status.Chains[domain.ChainInput].Rules {
				if strings.Contains(rule.Options, "45681") {
					t.Fatal("late mutation remained after remote expiry")
				}
			}
			_, err := fixtureSSHCommand(t, ctx, "test", "-f", "/tmp/nextgen-late-mutation-observed").Output()
			t.Logf("fixture observed delayed command after caller deadline: %v", err == nil)
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("remote expiry did not resolve a timed-out SSH mutation")
}
