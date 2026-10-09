package ws

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"

	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/service"
)

// TerminalResizeMessage modella il messaggio JSON di controllo dimensione finestra
type TerminalResizeMessage struct {
	Type string `json:"type"`
	Cols uint32 `json:"cols"`
	Rows uint32 `json:"rows"`
}

type windowChangeReq struct {
	Cols   uint32
	Rows   uint32
	Width  uint32
	Height uint32
}

// SSHTerminalHandler gestisce il WebSocket Gateway per terminali Web SSH (PTY)
type SSHTerminalHandler struct {
	sshService service.SSHService
	upgrader   websocket.Upgrader
}

// NewSSHTerminalHandler crea un'istanza dell'handler WebSocket configurato con buffer e CORS
func NewSSHTerminalHandler(sshService service.SSHService) *SSHTerminalHandler {
	return &SSHTerminalHandler{
		sshService: sshService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// HandleTerminal gestisce la sessione WebSocket PTY interattiva conforme a NFR-SEC-07 e NFR-SEC-03
func (h *SSHTerminalHandler) HandleTerminal(w http.ResponseWriter, r *http.Request) {
	// 1. Estrazione tenantID dal contesto di autenticazione JWT (NFR-SEC-03)
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant non valido o non presente")
		return
	}

	// 2. Estrazione e validazione parametro server ID
	serverIDStr := chi.URLParam(r, "id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Identificativo server non valido (UUID atteso)")
		return
	}

	// 3. Upgrade connessione HTTP a WebSocket
	wsConn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Upgrade WebSocket per terminale SSH fallito", "error", err, "tenant_id", tenantID, "server_id", serverID)
		return
	}
	defer wsConn.Close()

	// Mutex per garantire la sicurezza in scrittura concorrente sul socket WebSocket
	var wsWriteMu sync.Mutex
	safeWrite := func(msgType int, data []byte) error {
		wsWriteMu.Lock()
		defer wsWriteMu.Unlock()
		return wsConn.WriteMessage(msgType, data)
	}

	// 4. Connessione SSH remota tramite credenziali cifrate
	sshClient, err := h.sshService.GetRawSSHClient(r.Context(), tenantID, serverID)
	if err != nil {
		errMsg := "\r\n\x1b[31m[Errore SSH]: Impossibile stabilire la connessione con l'host remoto: " + err.Error() + "\x1b[0m\r\n"
		_ = safeWrite(websocket.TextMessage, []byte(errMsg))
		_ = wsConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "SSH connection failed"), time.Now().Add(time.Second))
		return
	}
	defer sshClient.Close()

	// 5. Creazione sessione SSH e allocazione PTY pseudo-terminale
	sshSession, err := sshClient.NewSession()
	if err != nil {
		errMsg := "\r\n\x1b[31m[Errore SSH]: Creazione sessione terminale fallita: " + err.Error() + "\x1b[0m\r\n"
		_ = safeWrite(websocket.TextMessage, []byte(errMsg))
		return
	}
	defer sshSession.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// Dimensioni di default xterm (120x30)
	if err := sshSession.RequestPty("xterm-256color", 30, 120, modes); err != nil {
		errMsg := "\r\n\x1b[31m[Errore SSH]: Allocazione PTY fallita: " + err.Error() + "\x1b[0m\r\n"
		_ = safeWrite(websocket.TextMessage, []byte(errMsg))
		return
	}

	stdinPipe, err := sshSession.StdinPipe()
	if err != nil {
		slog.Error("Acquisizione stdin pipe fallita", "error", err)
		return
	}
	defer stdinPipe.Close()

	stdoutPipe, err := sshSession.StdoutPipe()
	if err != nil {
		slog.Error("Acquisizione stdout pipe fallita", "error", err)
		return
	}

	// 6. Avvio shell interattiva sul PTY remoto
	if err := sshSession.Shell(); err != nil {
		errMsg := "\r\n\x1b[31m[Errore SSH]: Avvio shell fallito: " + err.Error() + "\x1b[0m\r\n"
		_ = safeWrite(websocket.TextMessage, []byte(errMsg))
		return
	}

	// Canale di coordinamento per terminazione coordinata (NFR-SEC-07)
	done := make(chan struct{})

	// 7. Goroutine Streaming Output: SSH PTY -> WebSocket
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				if writeErr := safeWrite(websocket.TextMessage, buf[:n]); writeErr != nil {
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					slog.Debug("Terminazione lettura stdout pipe SSH", "error", err)
				}
				return
			}
		}
	}()

	// 8. Loop Ricezione Input: WebSocket -> SSH PTY (keystrokes e resize)
	for {
		select {
		case <-done:
			return
		default:
		}

		msgType, msgBytes, err := wsConn.ReadMessage()
		if err != nil {
			// Disconnessione del client browser -> teardown immediato (NFR-SEC-07)
			return
		}

		if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
			// Verifica se si tratta di un messaggio JSON di controllo resize
			if len(msgBytes) > 0 && msgBytes[0] == '{' {
				var ctrlMsg TerminalResizeMessage
				if jsonErr := json.Unmarshal(msgBytes, &ctrlMsg); jsonErr == nil && ctrlMsg.Type == "resize" && ctrlMsg.Cols > 0 && ctrlMsg.Rows > 0 {
					resizeReq := windowChangeReq{
						Cols:   ctrlMsg.Cols,
						Rows:   ctrlMsg.Rows,
						Width:  ctrlMsg.Cols * 8,
						Height: ctrlMsg.Rows * 16,
					}
					_, _ = sshSession.SendRequest("window-change", false, ssh.Marshal(resizeReq))
					continue
				}
			}

			// Inoltro input tastiera/comandi diretti alla stdin pipe
			if _, err := stdinPipe.Write(msgBytes); err != nil {
				return
			}
		}
	}
}
