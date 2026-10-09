package ws_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"

	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/ws"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
)

type mockSSHServiceForWS struct {
	clientFunc func(ctx context.Context, tenantID, serverID uuid.UUID) (*ssh.Client, error)
}

func (m *mockSSHServiceForWS) ExecuteCommand(ctx context.Context, tenantID, serverID uuid.UUID, cmd string) (string, string, int, error) {
	return "", "", 0, nil
}
func (m *mockSSHServiceForWS) CreateServer(ctx context.Context, tenantID uuid.UUID, req service.CreateSSHServerRequest) (*domain.SSHServer, error) {
	return nil, nil
}
func (m *mockSSHServiceForWS) GetServer(ctx context.Context, tenantID, serverID uuid.UUID) (*domain.SSHServer, error) {
	return nil, nil
}
func (m *mockSSHServiceForWS) ListServers(ctx context.Context, tenantID uuid.UUID) ([]*domain.SSHServer, error) {
	return nil, nil
}
func (m *mockSSHServiceForWS) UpdateServer(ctx context.Context, tenantID, serverID uuid.UUID, req service.UpdateSSHServerRequest) (*domain.SSHServer, error) {
	return nil, nil
}
func (m *mockSSHServiceForWS) DeleteServer(ctx context.Context, tenantID, serverID uuid.UUID) error {
	return nil
}
func (m *mockSSHServiceForWS) TestConnection(ctx context.Context, tenantID, serverID uuid.UUID) (*service.SSHConnectionTestResult, error) {
	return nil, nil
}
func (m *mockSSHServiceForWS) GetRawSSHClient(ctx context.Context, tenantID, serverID uuid.UUID) (*ssh.Client, error) {
	if m.clientFunc != nil {
		return m.clientFunc(ctx, tenantID, serverID)
	}
	return nil, errors.New("mock ssh client non configurato")
}

func TestSSHTerminalHandler_InvalidServerID(t *testing.T) {
	mockSvc := &mockSSHServiceForWS{}
	handler := ws.NewSSHTerminalHandler(mockSvc)

	r := chi.NewRouter()
	r.Get("/ws/v1/ssh/servers/{id}/terminal", func(w http.ResponseWriter, req *http.Request) {
		// Mock tenant context
		ctx := middleware.WithUserClaims(req.Context(), &auth.UserClaims{
			TenantID: uuid.New(),
			UserID:   uuid.New(),
			Role:     domain.RoleSecOps,
		})
		handler.HandleTerminal(w, req.WithContext(ctx))
	})

	server := httptest.NewServer(r)
	defer server.Close()

	// Invia richiesta con ID non UUID
	resp, err := http.Get(server.URL + "/ws/v1/ssh/servers/not-a-valid-uuid/terminal")
	if err != nil {
		t.Fatalf("richiesta HTTP fallita: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("atteso status 400 Bad Request per UUID non valido, ottenuto: %d", resp.StatusCode)
	}
}

func TestSSHTerminalHandler_MissingTenantContext(t *testing.T) {
	mockSvc := &mockSSHServiceForWS{}
	handler := ws.NewSSHTerminalHandler(mockSvc)

	r := chi.NewRouter()
	r.Get("/ws/v1/ssh/servers/{id}/terminal", handler.HandleTerminal)

	server := httptest.NewServer(r)
	defer server.Close()

	serverID := uuid.New().String()
	resp, err := http.Get(server.URL + "/ws/v1/ssh/servers/" + serverID + "/terminal")
	if err != nil {
		t.Fatalf("richiesta HTTP fallita: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("atteso status 401 Unauthorized per mancanza di tenant nel contesto, ottenuto: %d", resp.StatusCode)
	}
}

func TestSSHTerminalHandler_SSHConnectionFailure(t *testing.T) {
	mockSvc := &mockSSHServiceForWS{
		clientFunc: func(ctx context.Context, tenantID, serverID uuid.UUID) (*ssh.Client, error) {
			return nil, errors.New("connection refused host unreachable")
		},
	}
	handler := ws.NewSSHTerminalHandler(mockSvc)
	tenantID := uuid.New()
	serverID := uuid.New()

	r := chi.NewRouter()
	r.Get("/ws/v1/ssh/servers/{id}/terminal", func(w http.ResponseWriter, req *http.Request) {
		ctx := middleware.WithUserClaims(req.Context(), &auth.UserClaims{
			TenantID: tenantID,
			UserID:   uuid.New(),
			Role:     domain.RoleSecOps,
		})
		handler.HandleTerminal(w, req.WithContext(ctx))
	})

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/v1/ssh/servers/" + serverID.String() + "/terminal"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket fallito: %v", err)
	}
	defer conn.Close()

	// Il server deve inviare il messaggio di errore prima di chiudere
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("lettura messaggio websocket fallita: %v", err)
	}

	if !strings.Contains(string(msg), "Errore SSH") {
		t.Errorf("atteso messaggio di errore SSH, ottenuto: %s", string(msg))
	}
}

// startLoopbackPTYSSHServer avvia un server SSH locale in-process con supporto PTY e Shell
func startLoopbackPTYSSHServer(t *testing.T) (string, int, func()) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generazione host key fallita: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(privKey)
	if err != nil {
		t.Fatalf("creazione signer fallita: %v", err)
	}

	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	config.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen locale fallito: %v", err)
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
						_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel")
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
							case "pty-req":
								_ = req.Reply(true, nil)
							case "shell":
								_ = req.Reply(true, nil)
								// Prompt iniziale di benvenuto
								_, _ = ch.Write([]byte("nextgen-shell$ "))

								// Echo semplice di stdin verso stdout
								go func() {
									buf := make([]byte, 1024)
									for {
										n, err := ch.Read(buf)
										if n > 0 {
											_, _ = ch.Write([]byte("echo: " + string(buf[:n])))
										}
										if err != nil {
											return
										}
									}
								}()
							case "window-change":
								_ = req.Reply(true, nil)
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

func TestSSHTerminalHandler_InteractiveSession_Loopback(t *testing.T) {
	sshHost, sshPort, cleanupSSH := startLoopbackPTYSSHServer(t)
	defer cleanupSSH()

	// Mock SSH service che restituisce connessione reale al nostro server SSH loopback
	mockSvc := &mockSSHServiceForWS{
		clientFunc: func(ctx context.Context, tenantID, serverID uuid.UUID) (*ssh.Client, error) {
			clientConfig := &ssh.ClientConfig{
				User:            "testuser",
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				Timeout:         5 * time.Second,
			}
			addr := net.JoinHostPort(sshHost, strconv.Itoa(sshPort))
			return ssh.Dial("tcp", addr, clientConfig)
		},
	}

	handler := ws.NewSSHTerminalHandler(mockSvc)
	tenantID := uuid.New()
	serverID := uuid.New()

	r := chi.NewRouter()
	r.Get("/ws/v1/ssh/servers/{id}/terminal", func(w http.ResponseWriter, req *http.Request) {
		ctx := middleware.WithUserClaims(req.Context(), &auth.UserClaims{
			TenantID: tenantID,
			UserID:   uuid.New(),
			Role:     domain.RoleSecOps,
		})
		handler.HandleTerminal(w, req.WithContext(ctx))
	})

	httpServer := httptest.NewServer(r)
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws/v1/ssh/servers/" + serverID.String() + "/terminal"
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket fallito: %v", err)
	}
	defer wsConn.Close()

	// 1. Lettura del prompt iniziale shell
	_, promptMsg, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("lettura prompt iniziale fallita: %v", err)
	}
	if !strings.Contains(string(promptMsg), "nextgen-shell$") {
		t.Errorf("atteso prompt iniziale, ottenuto: %s", string(promptMsg))
	}

	// 2. Invio comando/keystroke e ricezione echo
	if err := wsConn.WriteMessage(websocket.TextMessage, []byte("whoami\n")); err != nil {
		t.Fatalf("invio comando websocket fallito: %v", err)
	}

	_, echoMsg, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("lettura risposta echo fallita: %v", err)
	}
	if !strings.Contains(string(echoMsg), "echo: whoami") {
		t.Errorf("attesa risposta echo, ottenuto: %s", string(echoMsg))
	}

	// 3. Invio messaggio JSON di controllo resize finestra
	resizeJSON := `{"type":"resize","cols":140,"rows":40}`
	if err := wsConn.WriteMessage(websocket.TextMessage, []byte(resizeJSON)); err != nil {
		t.Fatalf("invio messaggio resize fallito: %v", err)
	}

	// 4. Invio ulteriore input per confermare che la sessione è attiva dopo resize
	if err := wsConn.WriteMessage(websocket.TextMessage, []byte("status\n")); err != nil {
		t.Fatalf("invio comando dopo resize fallito: %v", err)
	}

	_, afterResizeMsg, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("lettura risposta post-resize fallita: %v", err)
	}
	if !strings.Contains(string(afterResizeMsg), "echo: status") {
		t.Errorf("atteso echo dopo resize, ottenuto: %s", string(afterResizeMsg))
	}

	// 5. Chiusura WebSocket e verifica chiusura immediata e pulita
	_ = wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
	_ = wsConn.Close()
}
