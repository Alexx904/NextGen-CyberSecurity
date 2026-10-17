package service_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

func TestSSHMultipleHostKeys(t *testing.T) {
	_, edKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	var authentications atomic.Int32
	var denyAuthentication atomic.Bool
	config := &ssh.ServerConfig{PasswordCallback: func(_ ssh.ConnMetadata, _ []byte) (*ssh.Permissions, error) {
		authentications.Add(1)
		if denyAuthentication.Load() {
			return nil, errors.New("authentication denied")
		}
		return nil, nil
	}}
	var signers []ssh.Signer
	for _, key := range []any{edKey, rsaKey, ecKey} {
		signer, err := ssh.NewSignerFromKey(key)
		if err != nil {
			t.Fatal(err)
		}
		signers = append(signers, signer)
		config.AddHostKey(signer)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				client, channels, requests, err := ssh.NewServerConn(conn, config)
				if err != nil {
					return
				}
				defer client.Close()
				go ssh.DiscardRequests(requests)
				for channel := range channels {
					_ = channel.Reject(ssh.Prohibited, "no sessions")
				}
			}()
		}
	}()
	host, portText, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(portText)
	repo := newMockSSHServerRepo()
	svc := service.NewSSHService(repo, strings.Repeat("t", 32))
	tenant := uuid.New()
	for _, signer := range signers {
		t.Run(signer.PublicKey().Type(), func(t *testing.T) {
			server, err := svc.CreateServer(context.Background(), tenant, service.CreateSSHServerRequest{
				Name: signer.PublicKey().Type(), Host: host, Port: port, Username: "user",
				AuthType: domain.SSHAuthTypePassword, RawSecret: "synthetic", HostKeyFingerprint: ssh.FingerprintSHA256(signer.PublicKey()),
			})
			if err != nil {
				t.Fatal(err)
			}
			before := authentications.Load()
			client, err := svc.GetRawSSHClient(context.Background(), tenant, server.ID)
			if err != nil {
				t.Fatalf("authorized host key rejected: %v", err)
			}
			client.Close()
			if authentications.Load() != before+1 {
				t.Fatal("credentials sent on mismatched host key")
			}
		})
	}
	t.Run("authentication failure is not retried", func(t *testing.T) {
		denyAuthentication.Store(true)
		defer denyAuthentication.Store(false)
		server, err := svc.CreateServer(context.Background(), tenant, service.CreateSSHServerRequest{
			Name: "denied", Host: host, Port: port, Username: "user", AuthType: domain.SSHAuthTypePassword,
			RawSecret: "synthetic", HostKeyFingerprint: ssh.FingerprintSHA256(signers[0].PublicKey()),
		})
		if err != nil {
			t.Fatal(err)
		}
		before := authentications.Load()
		client, err := svc.GetRawSSHClient(context.Background(), tenant, server.ID)
		if client != nil {
			client.Close()
			t.Fatal("authentication failure accepted")
		}
		if err == nil || !strings.Contains(err.Error(), "unable to authenticate") {
			t.Fatalf("authentication error lost: %v", err)
		}
		if authentications.Load() != before+1 {
			t.Fatal("authentication failure retried")
		}
	})
	t.Run("unknown pin", func(t *testing.T) {
		pin := "SHA256:" + strings.Repeat("A", 43)
		server, err := svc.CreateServer(context.Background(), tenant, service.CreateSSHServerRequest{
			Name: "unknown", Host: host, Port: port, Username: "user", AuthType: domain.SSHAuthTypePassword,
			RawSecret: "synthetic", HostKeyFingerprint: pin,
		})
		if err != nil {
			t.Fatal(err)
		}
		before := authentications.Load()
		client, err := svc.GetRawSSHClient(context.Background(), tenant, server.ID)
		if client != nil {
			client.Close()
			t.Fatal("unknown host accepted")
		}
		if err == nil || !strings.Contains(err.Error(), pin) {
			t.Fatalf("expected pin missing from diagnostic: %v", err)
		}
		for _, signer := range signers {
			if !strings.Contains(err.Error(), ssh.FingerprintSHA256(signer.PublicKey())) {
				t.Fatalf("observed key missing: %v", err)
			}
		}
		if authentications.Load() != before {
			t.Fatal("credentials sent to untrusted host")
		}
	})
}

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
