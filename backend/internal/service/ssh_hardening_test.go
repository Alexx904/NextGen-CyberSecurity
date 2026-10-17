package service_test

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

func TestSSHHostIdentityBeforeAuthentication(t *testing.T) {
	host, port, pin, cleanup := startTestSSHServer(t, "user", "synthetic-password")
	defer cleanup()
	repo := newMockSSHServerRepo()
	svc := service.NewSSHService(repo, strings.Repeat("t", 32))
	tenant := uuid.New()
	server, err := svc.CreateServer(context.Background(), tenant, service.CreateSSHServerRequest{Name: "identity", Host: host, Port: port, Username: "user", AuthType: domain.SSHAuthTypePassword, RawSecret: "synthetic-password"})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range []string{"", "SHA256:" + strings.Repeat("A", 43), pin} {
		server.HostKeyFingerprint = candidate
		if err := repo.Update(context.Background(), server); err != nil {
			t.Fatal(err)
		}
		_, _, code, err := svc.ExecuteCommand(context.Background(), tenant, server.ID, "echo test")
		if candidate == pin {
			if err != nil || code != 0 {
				t.Fatalf("trusted host rejected: %v (%d)", err, code)
			}
		} else if err == nil || code != -1 {
			t.Fatalf("untrusted host accepted: pin=%q code=%d", candidate, code)
		}
	}
}

func TestSSHCommandCancellation(t *testing.T) {
	started := make(chan struct{}, 1)
	host, port, pin, cleanup := startTestSSHServer(t, "user", "synthetic-password", func(cmd string) {
		if cmd == "stall" {
			started <- struct{}{}
		}
	})
	defer cleanup()
	repo := newMockSSHServerRepo()
	svc := service.NewSSHService(repo, strings.Repeat("t", 32))
	tenant := uuid.New()
	server, err := svc.CreateServer(context.Background(), tenant, service.CreateSSHServerRequest{Name: "cancel", Host: host, Port: port, Username: "user", AuthType: domain.SSHAuthTypePassword, RawSecret: "synthetic-password"})
	if err != nil {
		t.Fatal(err)
	}
	server.HostKeyFingerprint = pin
	if err := repo.Update(context.Background(), server); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { _, _, _, err := svc.ExecuteCommand(ctx, tenant, server.ID, "stall"); done <- err }()
	select {
	case <-started:
	case err := <-done:
		t.Fatalf("command did not start: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("command never reached remote server")
	}
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("deadline not preserved: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("command ignored deadline")
	}
}

func TestSSHHandshakeCancellation(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()
	host, portText, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(portText)
	repo := newMockSSHServerRepo()
	svc := service.NewSSHService(repo, strings.Repeat("t", 32))
	tenant := uuid.New()
	server, err := svc.CreateServer(context.Background(), tenant, service.CreateSSHServerRequest{Name: "handshake", Host: host, Port: port, Username: "user", AuthType: domain.SSHAuthTypePassword, RawSecret: "synthetic"})
	if err != nil {
		t.Fatal(err)
	}
	server.HostKeyFingerprint = "SHA256:" + strings.Repeat("A", 43)
	if err := repo.Update(context.Background(), server); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		client, err := svc.GetRawSSHClient(ctx, tenant, server.ID)
		if client != nil {
			client.Close()
		}
		done <- err
	}()
	select {
	case conn := <-accepted:
		defer conn.Close()
	case <-time.After(time.Second):
		t.Fatal("no connection")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation not preserved: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("handshake ignored cancellation")
	}
}

func TestSSHHostPinEnrollmentAndRotation(t *testing.T) {
	repo := newMockSSHServerRepo()
	svc := service.NewSSHService(repo, strings.Repeat("t", 32))
	tenant := uuid.New()
	pin := "SHA256:" + strings.Repeat("A", 43)
	ctx := context.Background()
	create := service.CreateSSHServerRequest{Name: "rotation", Host: "192.0.2.1", Port: 22, Username: "user", AuthType: domain.SSHAuthTypePassword, RawSecret: "synthetic", HostKeyFingerprint: pin}
	server, err := svc.CreateServer(ctx, tenant, create)
	if err != nil {
		t.Fatal(err)
	}
	if server.HostKeyFingerprint != pin || server.PublicKeyFingerprint != nil {
		t.Fatal("host/client fingerprints confused")
	}
	req := service.UpdateSSHServerRequest{Name: server.Name, Host: server.Host, Port: 22, Username: server.Username, AuthType: server.AuthType}
	updated, err := svc.UpdateServer(ctx, tenant, server.ID, req)
	if err != nil || updated.HostKeyFingerprint != pin {
		t.Fatalf("omitted pin lost: %v", err)
	}
	rotated := "SHA256:" + strings.Repeat("B", 42) + "A"
	req.HostKeyFingerprint = &rotated
	updated, err = svc.UpdateServer(ctx, tenant, server.ID, req)
	if err != nil || updated.HostKeyFingerprint != rotated {
		t.Fatalf("rotation failed: %v", err)
	}
	req.Host = "192.0.2.2"
	req.HostKeyFingerprint = nil
	updated, err = svc.UpdateServer(ctx, tenant, server.ID, req)
	if err != nil || updated.HostKeyFingerprint != "" || updated.Status != domain.SSHServerStatusUnchecked {
		t.Fatalf("endpoint retained trust: %+v %v", updated, err)
	}
	req.HostKeyFingerprint = &pin
	if _, err = svc.UpdateServer(ctx, tenant, server.ID, req); err != nil {
		t.Fatal(err)
	}
	empty := ""
	req.HostKeyFingerprint = &empty
	updated, err = svc.UpdateServer(ctx, tenant, server.ID, req)
	if err != nil || updated.HostKeyFingerprint != "" {
		t.Fatalf("explicit clearing failed: %v", err)
	}
	for _, bad := range []string{"MD5:00:00", pin + "=", "SHA256:" + strings.Repeat("A", 42) + "B", "SHA256:bad"} {
		create.Name = bad
		create.HostKeyFingerprint = bad
		if _, err := svc.CreateServer(ctx, tenant, create); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("invalid pin accepted %q: %v", bad, err)
		}
	}
	if _, err := svc.GetRawSSHClient(ctx, uuid.New(), server.ID); !errors.Is(err, domain.ErrSSHServerNotFound) {
		t.Fatalf("tenant bypass: %v", err)
	}
}
