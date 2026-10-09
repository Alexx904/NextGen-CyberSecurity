# Dominio: Remote SSH Host & iptables/nftables Firewall Manager

> **Dominio:** Amministrazione Remota Sicura, Firewall Rules Engine & Terminale Web  
> **File di Riferimento:** `backend/internal/domain/ssh_server.go`, `backend/internal/domain/firewall.go`, `backend/internal/service/ssh_service.go`, `backend/internal/service/firewall_service.go`, `backend/internal/delivery/http/handler/ssh_firewall_handler.go`, `backend/internal/delivery/ws/ssh_terminal_ws.go`, `frontend/src/pages/SSHServers.tsx`, `frontend/src/components/FirewallManager.tsx`, `frontend/src/components/WebTerminalModal.tsx`

---

## 1. Gestione Server SSH Remoti & Cifratura Credenziali

### 1.1 Modello di Sicurezza (`NFR-SEC-06`)
- Le credenziali dei server remoti (password di login o passphrase/chiavi private SSH) non sono mai salvate in chiaro.
- Vengono cifrate autenticate a riposo con algoritmo **AES-256-GCM** tramite `backend/pkg/crypto/aes_gcm.go`.
- Il client SSH Go (`golang.org/x/crypto/ssh`) stabilisce connessioni conformi RFC 4251 con timeout di inattività, controllo fingerprint di host key e gestione privilegi (sudoers per regole iptables).

---

## 2. Firewall Management & Anti-Lockout Safety Timer

### 2.1 Tabella Visiva Regole (`iptables`/`nftables`)
- Ispezione delle catene standard (`INPUT`, `OUTPUT`, `FORWARD`) con relative policy di default (`ACCEPT`, `DROP`, `REJECT`).
- Aggiunta guidata e rimozione sicura di regole di filtraggio per protocollo (TCP/UDP/ICMP), IP sorgente, porta e commento.

### 2.2 Safety Rollback Timer (Protezione Anti-Lockout)
- Per prevenire il blocco accidentale dell'amministratore (lockout dalla porta 22):
  1. Prima dell'applicazione delle nuove regole, viene salvato uno snapshot consistente in `FIREWALL_BACKUPS`.
  2. Viene applicata la nuova configurazione e attivato un timer asincrono (es. 60 secondi).
  3. L'utente deve inviare un comando esplicito di conferma (`POST /api/v1/ssh/servers/{id}/firewall/confirm`).
  4. In assenza di conferma prima della scadenza del timer, il backend Go ripristina automaticamente lo snapshot precedente tramite `iptables-restore`.

---

## 3. Web Terminal SSH Integrato

### 3.1 Gateway WebSocket
- Endpoint dedicato: `GET /ws/v1/ssh/servers/{id}/terminal?token={jwt}`
- Autenticazione rigorosa via token JWT nel query param al momento dell'upgrade HTTP -> WebSocket.
- All'avvio dell'handshake, il backend alloca una pseudo-terminale (PTY) sulla sessione SSH remota (`pty.Request`).
- I frame WebSocket bidirezionali trasportano raw ANSI I/O e messaggi JSON di resize finestra (`{ "type": "resize", "cols": 120, "rows": 30 }`).

### 3.2 Frontend xterm.js
- Modal interattivo `WebTerminalModal.tsx` equipaggiato con `@xterm/xterm` e `@xterm/addon-fit`.
- Rendering fluido, supporto colori ANSI 256, copia/incolla nativo e terminazione istantanea della sessione alla chiusura del modal.
