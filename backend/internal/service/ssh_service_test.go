package service_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/crypto"
)

type mockSSHServerRepo struct {
	mu      sync.RWMutex
	servers map[string]*domain.SSHServer
}

func newMockSSHServerRepo() *mockSSHServerRepo {
	return &mockSSHServerRepo{
		servers: make(map[string]*domain.SSHServer),
	}
}

func (m *mockSSHServerRepo) makeKey(tenantID, id uuid.UUID) string {
	return tenantID.String() + ":" + id.String()
}

func (m *mockSSHServerRepo) Create(ctx context.Context, server *domain.SSHServer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, s := range m.servers {
		if s.TenantID == server.TenantID && s.Name == server.Name {
			return domain.ErrSSHServerDuplicateName
		}
	}
	m.servers[m.makeKey(server.TenantID, server.ID)] = server
	return nil
}

func (m *mockSSHServerRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.SSHServer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.servers[m.makeKey(tenantID, id)]
	if !ok {
		return nil, domain.ErrSSHServerNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *mockSSHServerRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.SSHServer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*domain.SSHServer
	for _, s := range m.servers {
		if s.TenantID == tenantID {
			cp := *s
			res = append(res, &cp)
		}
	}
	return res, nil
}

func (m *mockSSHServerRepo) Update(ctx context.Context, server *domain.SSHServer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.makeKey(server.TenantID, server.ID)
	if _, ok := m.servers[key]; !ok {
		return domain.ErrSSHServerNotFound
	}
	m.servers[key] = server
	return nil
}

func (m *mockSSHServerRepo) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status domain.SSHServerStatus, lastConnectedAt *time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.makeKey(tenantID, id)
	s, ok := m.servers[key]
	if !ok {
		return domain.ErrSSHServerNotFound
	}
	s.Status = status
	s.LastConnectedAt = lastConnectedAt
	return nil
}

func (m *mockSSHServerRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.makeKey(tenantID, id)
	if _, ok := m.servers[key]; !ok {
		return domain.ErrSSHServerNotFound
	}
	delete(m.servers, key)
	return nil
}

func TestSSHService_CreateServer_Encryption(t *testing.T) {
	repo := newMockSSHServerRepo()
	masterKey := "super-secure-32-byte-master-key!"
	svc := service.NewSSHService(repo, masterKey)

	tenantID := uuid.New()
	rawPassword := "VerySecretRootPassword!"

	req := service.CreateSSHServerRequest{
		Name:      "Web Server Prod",
		Host:      "192.168.1.50",
		Port:      22,
		Username:  "admin",
		AuthType:  domain.SSHAuthTypePassword,
		RawSecret: rawPassword,
	}

	created, err := svc.CreateServer(context.Background(), tenantID, req)
	if err != nil {
		t.Fatalf("CreateServer fallito: %v", err)
	}

	if created.Name != req.Name {
		t.Errorf("atteso nome %q, ottenuto %q", req.Name, created.Name)
	}

	// Il secret salvato nel DB non deve MAI essere in chiaro
	if created.EncryptedSecret == rawPassword {
		t.Fatal("il secret nel server non deve essere in chiaro!")
	}

	// Verifica decifratura con chiave derivata del tenant
	tenantKey, err := crypto.DeriveTenantKey([]byte(masterKey), tenantID)
	if err != nil {
		t.Fatalf("DeriveTenantKey fallito: %v", err)
	}

	decrypted, err := crypto.DecryptAES256GCM(created.EncryptedSecret, tenantKey, []byte(tenantID.String()))
	if err != nil {
		t.Fatalf("decifratura del secret fallita: %v", err)
	}

	if string(decrypted) != rawPassword {
		t.Errorf("secret decifrato non corrisponde: atteso %q, ottenuto %q", rawPassword, string(decrypted))
	}
}

func TestSSHService_UpdateServer_ReEncryption(t *testing.T) {
	repo := newMockSSHServerRepo()
	masterKey := "super-secure-32-byte-master-key!"
	svc := service.NewSSHService(repo, masterKey)

	tenantID := uuid.New()
	req := service.CreateSSHServerRequest{
		Name:      "Database Prod",
		Host:      "10.0.0.5",
		Port:      2222,
		Username:  "dba",
		AuthType:  domain.SSHAuthTypePassword,
		RawSecret: "oldPassword",
	}

	created, err := svc.CreateServer(context.Background(), tenantID, req)
	if err != nil {
		t.Fatalf("CreateServer fallito: %v", err)
	}

	newSecret := "NewSuperSecretPassword2026!"
	updateReq := service.UpdateSSHServerRequest{
		Name:      "Database Prod Renamed",
		Host:      "10.0.0.6",
		Port:      2222,
		Username:  "dba",
		AuthType:  domain.SSHAuthTypePassword,
		RawSecret: &newSecret,
	}

	updated, err := svc.UpdateServer(context.Background(), tenantID, created.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdateServer fallito: %v", err)
	}

	if updated.Name != "Database Prod Renamed" || updated.Host != "10.0.0.6" {
		t.Errorf("dati non aggiornati: %+v", updated)
	}

	// Verifica decifratura del nuovo secret
	tenantKey, _ := crypto.DeriveTenantKey([]byte(masterKey), tenantID)
	decrypted, err := crypto.DecryptAES256GCM(updated.EncryptedSecret, tenantKey, []byte(tenantID.String()))
	if err != nil {
		t.Fatalf("decifratura nuovo secret fallita: %v", err)
	}
	if string(decrypted) != newSecret {
		t.Errorf("nuovo secret decifrato errato: atteso %q, ottenuto %q", newSecret, string(decrypted))
	}
}

// startTestSSHServer avvia un server SSH di test in-process su porta casuale locale
func startTestSSHServer(t *testing.T, expectedUser, expectedPassword string) (string, int, func()) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generazione host key RSA fallita: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(privKey)
	if err != nil {
		t.Fatalf("creazione signer fallita: %v", err)
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == expectedUser && string(pass) == expectedPassword {
				return nil, nil
			}
			return nil, errors.New("autenticazione rifiutata")
		},
	}
	config.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen su porta locale fallito: %v", err)
	}

	host, portStr, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(portStr)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go func(c net.Conn) {
				defer c.Close()
				sshConn, chans, reqs, err := ssh.NewServerConn(c, config)
				if err != nil {
					return
				}
				defer sshConn.Close()
				go ssh.DiscardRequests(reqs)

				for newChannel := range chans {
					if newChannel.ChannelType() != "session" {
						_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
						continue
					}
					channel, requests, err := newChannel.Accept()
					if err != nil {
						return
					}

					go func(ch ssh.Channel, in <-chan *ssh.Request) {
						defer ch.Close()
						for req := range in {
							switch req.Type {
							case "exec":
								// payload ha 4 byte di lunghezza stringa seguiti dal comando
								if len(req.Payload) >= 4 {
									cmd := string(req.Payload[4:])
									if strings.Contains(cmd, "uname") {
										_, _ = ch.Write([]byte("Linux 6.6.0-custom-arch x86_64\n0\n"))
										_ = req.Reply(true, nil)
										sendExitStatus(ch, 0)
										return
									} else if strings.Contains(cmd, "echo") {
										_, _ = ch.Write([]byte("hello from mock ssh\n"))
										_ = req.Reply(true, nil)
										sendExitStatus(ch, 0)
										return
									} else if strings.Contains(cmd, "fail_cmd") {
										_, _ = ch.Write([]byte("command failed\n"))
										_ = req.Reply(true, nil)
										sendExitStatus(ch, 1)
										return
									}
								}
								_ = req.Reply(true, nil)
								sendExitStatus(ch, 0)
								return
							default:
								_ = req.Reply(false, nil)
							}
						}
					}(channel, requests)
				}
			}(conn)
		}
	}()

	cleanup := func() {
		_ = listener.Close()
	}

	return host, port, cleanup
}

func sendExitStatus(ch io.Writer, status uint32) {
	type exitStatusMsg struct {
		Status uint32
	}
	msg := exitStatusMsg{Status: status}
	b := ssh.Marshal(msg)
	// Invio del payload RFC 4254 exit-status
	if channel, ok := ch.(ssh.Channel); ok {
		_, _ = channel.SendRequest("exit-status", false, b)
	}
}

func TestSSHService_TestConnectionAndExecute_Loopback(t *testing.T) {
	username := "testadmin"
	password := "TestPass123!"

	host, port, cleanup := startTestSSHServer(t, username, password)
	defer cleanup()

	repo := newMockSSHServerRepo()
	masterKey := "super-secure-32-byte-master-key!"
	svc := service.NewSSHService(repo, masterKey)
	tenantID := uuid.New()

	// 1. Registrazione server di test
	server, err := svc.CreateServer(context.Background(), tenantID, service.CreateSSHServerRequest{
		Name:      "Loopback Server",
		Host:      host,
		Port:      port,
		Username:  username,
		AuthType:  domain.SSHAuthTypePassword,
		RawSecret: password,
	})
	if err != nil {
		t.Fatalf("CreateServer fallito: %v", err)
	}

	// 2. Test Connection riuscito
	res, err := svc.TestConnection(context.Background(), tenantID, server.ID)
	if err != nil {
		t.Fatalf("TestConnection ha restituito errore: %v", err)
	}
	if !res.Online {
		t.Fatalf("atteso server Online=true, ottenuto errore: %s", res.Error)
	}
	if !res.IsRoot {
		t.Errorf("atteso IsRoot=true (id 0), ottenuto false")
	}
	if !strings.Contains(res.OSInfo, "Linux") {
		t.Errorf("atteso OSInfo contenente Linux, ottenuto: %s", res.OSInfo)
	}

	// Verifica che lo status nel repo sia stato aggiornato a ONLINE
	stored, _ := repo.GetByID(context.Background(), tenantID, server.ID)
	if stored.Status != domain.SSHServerStatusOnline {
		t.Errorf("atteso status nel repository ONLINE, ottenuto: %s", stored.Status)
	}

	// 3. Esecuzione comando remoto riuscito
	stdout, stderr, exitCode, err := svc.ExecuteCommand(context.Background(), tenantID, server.ID, "echo 'hello'")
	if err != nil {
		t.Fatalf("ExecuteCommand fallito: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("atteso exitCode 0, ottenuto %d", exitCode)
	}
	if !strings.Contains(stdout, "hello from mock ssh") {
		t.Errorf("output comando inatteso: stdout=%q, stderr=%q", stdout, stderr)
	}

	// 4. Test con credenziali errate
	wrongServer, err := svc.CreateServer(context.Background(), tenantID, service.CreateSSHServerRequest{
		Name:      "Wrong Creds Server",
		Host:      host,
		Port:      port,
		Username:  username,
		AuthType:  domain.SSHAuthTypePassword,
		RawSecret: "WrongPassword999!",
	})
	if err != nil {
		t.Fatalf("CreateServer wrong fallito: %v", err)
	}

	resWrong, err := svc.TestConnection(context.Background(), tenantID, wrongServer.ID)
	if err != nil {
		t.Fatalf("TestConnection ha restituito errore inatteso: %v", err)
	}
	if resWrong.Online {
		t.Errorf("il server con password errata doveva risultare Online=false")
	}
	if resWrong.Error == "" {
		t.Errorf("atteso messaggio di errore nel risultato")
	}

	// Verifica che lo status sia stato marcato OFFLINE
	storedWrong, _ := repo.GetByID(context.Background(), tenantID, wrongServer.ID)
	if storedWrong.Status != domain.SSHServerStatusOffline {
		t.Errorf("atteso status OFFLINE, ottenuto: %s", storedWrong.Status)
	}
}
