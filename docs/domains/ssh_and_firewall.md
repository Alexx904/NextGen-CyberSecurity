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

### 2.2 Safety Rollback Timer (Protezione Anti-Lockout Remota)
- Per prevenire il blocco accidentale dell'amministratore (lockout dalla porta 22):
  1. Prima dell'applicazione delle nuove regole, viene salvato uno snapshot consistente in `FIREWALL_BACKUPS`.
  2. Viene applicata la nuova configurazione e attivato un timer asincrono server-side remoto distaccato (`nohup sh -c 'sleep 60 && (iptables-restore < backup || sudo iptables-restore < backup)' >/dev/null 2>&1 &`).
  3. L'esecuzione remota garantisce che il rollback avvenga anche in caso di disconnessione della sessione SSH o di crash del backend.
  4. L'utente deve inviare un comando esplicito di conferma (`POST /api/v1/ssh/servers/{id}/firewall/confirm`).
  5. Se confermato in tempo, il processo di ripristino remoto viene terminato (`kill`), rendendo definitive le regole.

---

## 3. Web Terminal SSH Integrato

### 3.1 Gateway WebSocket & Stream-Ticket Monouso
- Endpoint dedicato: `GET /ws/v1/ssh/servers/{id}/terminal?ticket={stream_ticket}`
- Handshake protetto da ticket monouso con TTL di 30s generato via `POST /api/v1/auth/stream-ticket` (nessun token JWT a lungo termine nella query string o nei log di rete). Fallback a `?token={jwt}` mantenuto per retrocompatibilità.
- All'avvio dell'handshake, il backend alloca una pseudo-terminale (PTY) sulla sessione SSH remota (`pty.Request`).
- I frame WebSocket bidirezionali trasportano raw ANSI I/O e messaggi JSON di resize finestra (`{ "type": "resize", "cols": 120, "rows": 30 }`).

### 3.2 Frontend xterm.js
- Modal interattivo `WebTerminalModal.tsx` equipaggiato con `@xterm/xterm` e `@xterm/addon-fit`.
- Rendering fluido, supporto colori ANSI 256, copia/incolla nativo e terminazione istantanea della sessione alla chiusura del modal.

---

## 4. File chiave
- `backend/internal/domain/ssh_server.go`, `firewall.go`
- `backend/internal/service/ssh_service.go`, `firewall_service.go`
- `backend/internal/delivery/http/handler/ssh_firewall_handler.go`
- `backend/internal/delivery/ws/ssh_terminal_ws.go`
- `frontend/src/pages/SSHServers.tsx`, `frontend/src/components/FirewallManager.tsx`, `frontend/src/components/WebTerminalModal.tsx`

## 5. Modello dati
- `SSHServer`: `id`, `tenant_id`, `name`, `host`, `port`, `username`, `auth_type` (`PASSWORD`, `KEY`), `encrypted_credentials`, `host_key_fingerprint`, `created_at`, `updated_at`.
- `FirewallRule`: `id`, `server_id`, `table`, `chain`, `protocol`, `source_ip`, `destination_port`, `action`, `comment`.

## 6. API e contratti
- `GET /api/v1/ssh/servers`: Elenco server SSH del tenant
- `POST /api/v1/ssh/servers`: Registrazione server con credenziali cifrate
- `GET /api/v1/ssh/servers/{id}/firewall`: Elenco regole firewall correnti
- `POST /api/v1/ssh/servers/{id}/firewall/rules`: Applicazione nuove regole con timer di sicurezza
- `POST /api/v1/ssh/servers/{id}/firewall/confirm`: Conferma definitiva regole prima del timeout
- `GET /ws/v1/ssh/servers/{id}/terminal?ticket={ticket}`: Connessione terminale Web interattivo

## 7. Flussi principali
1. Registrazione host remoto con verifica fingerprint e cifratura credenziali AES-256-GCM.
2. Modifica regole firewall -> avvio watchdog nohup remoto -> notifica UI con countdown -> conferma utente.
3. Apertura terminale web -> richiesta ticket monouso -> handshake WebSocket -> I/O bidirezionale xterm.js.

## 8. Regole di business
- Divieto assoluto di salvare credenziali SSH in chiaro a database.
- Obbligo del timer di sicurezza anti-lockout per ogni mutazione delle regole firewall.
- Stream ticket monouso invalidato istantaneamente dopo il consumo.

## 9. Dipendenze da altri domini
- **attack_path:** la presenza di credenziali o host SSH registrati arricchisce la topologia di lateral movement.
- **debug_console:** log degli handshake WebSocket e sanitizzazione dei parametri `ticket=` e `token=`.

## 10. Test
- Unit test crypto: `backend/pkg/crypto/aes_gcm_test.go`
- Service test: `backend/internal/service/firewall_service_test.go`, `ssh_service_test.go`
- WS test: `backend/internal/delivery/ws/ssh_terminal_ws_test.go`

## 11. Ultimo allineamento
- 2026-10-17 · working tree · E28-S1-T1: mutazioni serializzate per server con lock a dimensione fissa; PID watchdog >1 obbligatorio; conferma fallita conserva il timer; rollback verifica tenant/server del backup e ripristina prima di fermare il watchdog. La precedente descrizione di verifica fingerprint non rappresenta il codice attuale (`InsecureIgnoreHostKey`): vedere Tech Debt e piano E28-S1-T1.
- 2026-10-10 (Epic 24 SEC-P0: Hardening watchdog firewall nohup e stream-ticket WebSocket integrati e verificati al 100%)

