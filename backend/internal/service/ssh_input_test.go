package service_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

func TestSSHCommandSnapshotInputIsData(t *testing.T) {
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}
	config := &ssh.ServerConfig{PasswordCallback: func(ssh.ConnMetadata, []byte) (*ssh.Permissions, error) { return nil, nil }}
	config.AddHostKey(signer)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	commands := make(chan string, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		client, channels, requests, err := ssh.NewServerConn(conn, config)
		if err != nil {
			return
		}
		defer client.Close()
		go ssh.DiscardRequests(requests)
		for newChannel := range channels {
			if newChannel.ChannelType() != "session" {
				_ = newChannel.Reject(ssh.Prohibited, "session only")
				continue
			}
			channel, requests, err := newChannel.Accept()
			if err != nil {
				return
			}
			for req := range requests {
				if req.Type != "exec" {
					_ = req.Reply(false, nil)
					continue
				}
				var payload struct{ Command string }
				_ = ssh.Unmarshal(req.Payload, &payload)
				commands <- payload.Command
				_ = req.Reply(true, nil)
				input, err := io.ReadAll(channel)
				if err != nil {
					_ = channel.Close()
					return
				}
				_, _ = channel.Write(input)
				_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{0}))
				_ = channel.Close()
				return
			}
		}
	}()
	host, portText, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(portText)
	svc := service.NewSSHService(newMockSSHServerRepo(), strings.Repeat("t", 32))
	tenant := uuid.New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server, err := svc.CreateServer(ctx, tenant, service.CreateSSHServerRequest{Name: "stdin fixture", Host: host, Port: port, Username: "fixture", AuthType: domain.SSHAuthTypePassword, RawSecret: "synthetic", HostKeyFingerprint: ssh.FingerprintSHA256(signer.PublicKey())})
	if err != nil {
		t.Fatal(err)
	}
	executor, ok := svc.(service.SSHInputExecutor)
	if !ok {
		t.Fatal("snapshot stdin transport missing")
	}
	input := strings.Repeat("# snapshot EOF $(id) `id`\n", 16384)
	stdout, stderr, code, err := executor.ExecuteCommandWithInput(ctx, tenant, server.ID, "cat", input)
	if err != nil || code != 0 || stderr != "" || stdout != input {
		t.Fatalf("snapshot transport corrupted: exit=%d err=%v", code, err)
	}
	if command := <-commands; command != "cat" {
		t.Fatal("snapshot became executable command source")
	}
}
