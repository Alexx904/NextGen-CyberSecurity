# Dominio: Remote SSH Host & iptables/nftables Firewall Manager

## SSH-TOOLS-01 · Proxy terminale e prerequisiti firewall · 2026-10-17 · working tree

Accettato dal PO il 2026-10-17 con «Ok perfetto è funzionante», dopo il feedback
sul 502. Esito operativo riferito dal PO; nessuna prova watchdog/rollback o
configurazione sudoers specifica dedotta dall'accettazione.

Nginx e Vite inoltrano `/ws/` al backend con upgrade WebSocket. La lettura
iptables usa PATH comprendente sbin, locale C e accesso diretto per root oppure
`sudo -n` per altri utenti: nessun prompt password su una sessione senza PTY.
Prerequisiti mancanti e privilegi insufficienti mantengono HTTP 502 con codici
`FIREWALL_UNAVAILABLE` / `FIREWALL_PRIVILEGES_REQUIRED` e messaggi operativi
sanificati. Il frontend mostra il messaggio API anche quando HTTP non è 2xx.
Nessuna concessione automatica di privilegi, installazione remota o modifica
alle regole/watchdog. SSH-CONNECT-02 accettato dal PO: la connessione funziona.

## SSH-HARDEN-01 · Identità host e deadline · 2026-10-17 · working tree

SSH-CONNECT-02: con più chiavi host, un mismatch esclude la famiglia appena
osservata e ritenta la negoziazione su una nuova connessione, entro lo stesso
budget handshake di 10s. Ogni tentativo richiede il pin esatto prima del login;
nessuna fiducia automatica o modifica del pin. I retry si limitano ai mismatch,
non agli errori di rete/autenticazione. Se nessuna chiave coincide, il messaggio
indica fingerprint atteso e chiavi osservate, da verificare su canale fidato.

`host_key_fingerprint` separato da `public_key_fingerprint` (chiave client).
Create accetta stringa SHA256 OpenSSH canonica o vuota; update omesso conserva,
vuoto azzera. Un cambio host/porta senza pin esplicito azzera la fiducia; il form
azzera il campo quando l'endpoint viene modificato. Cambio endpoint/pin resetta
status e last_connected_at. Pin assente o discordante blocca handshake prima
dell'autenticazione per test connessione, comandi e terminale. Nessun TOFU:
l'operatore verifica il fingerprint con l'amministratore tramite canale fidato.
La rotazione richiede sostituzione esplicita del pin tramite update autorizzato.

Dial/handshake: massimo 10s, con cancellazione e deadline del chiamante.
Comandi/test connessione: budget totale 30s, anche se il peer non completa exec
o apertura sessione. La chiusura della connessione interrompe le attese locali;
non dimostra la terminazione del processo remoto. Il terminale mantiene la sua
durata interattiva dopo l'handshake. API additive, migrazione 000010 up/down:
legacy rimane inventariato ma non connettibile; down perde solo i pin host.
Re-upgrade richiede nuova configurazione, mai riuso del fingerprint client.

TECH-DEBT-07 resta aperto: la transazione remota snapshot/watchdog/applicazione,
il coordinamento fra istanze e il collaudo Linux timeout/disconnessione/rollback
non sono risolti dalle sole deadline. E28-S1-T1 non promosso per questo incremento.
Piano e verifiche: [SSH-HARDEN-01](../04-development/plans/SSH-HARDEN-01.md).

## COMPANY-UX-02 · Clienti autonomi · 2026-10-17 · working tree

SSH/Firewall indipendente dalle aziende: route `/ssh-servers` e tool dashboard accessibili anche globalmente, senza CompanyContext nella pagina. Server con `client_name` autonomo opzionale, max 160 caratteri, trim service, default DB vuoto. Non è una FK o autorizzazione cliente; isolamento tenant/serverID invariato. Filtro/raggruppamento e ricerca interni, valore vuoto “Senza cliente”.

API create accetta stringa `client_name`; update omesso conserva, stringa vuota azzera. Risposte includono il campo, credenziali cifrate/mascherate invariate. Migrazione `000008_ssh_client_labels` up/down; down perde le sole etichette, da esportare prima del rollback operativo. Persistenza/compatibilità provate su DB dedicato, nessuna connessione SSH cliente eseguita. Gate Linux e hardening E28-S1-T1 separati e aperti.

> **Dominio:** Amministrazione Remota Sicura, Firewall Rules Engine & Terminale Web  
> **File di Riferimento:** `backend/internal/domain/ssh_server.go`, `backend/internal/domain/firewall.go`, `backend/internal/service/ssh_service.go`, `backend/internal/service/firewall_service.go`, `backend/internal/delivery/http/handler/ssh_firewall_handler.go`, `backend/internal/delivery/ws/ssh_terminal_ws.go`, `frontend/src/pages/SSHServers.tsx`, `frontend/src/components/FirewallManager.tsx`, `frontend/src/components/WebTerminalModal.tsx`

---

## 1. Gestione Server SSH Remoti & Cifratura Credenziali

### 1.1 Modello di Sicurezza (`NFR-SEC-06`)
- Le credenziali dei server remoti (password di login o passphrase/chiavi private SSH) non sono mai salvate in chiaro.
- Vengono cifrate autenticate a riposo con algoritmo **AES-256-GCM** tramite `backend/pkg/crypto/aes_gcm.go`.
- Il client SSH Go (`golang.org/x/crypto/ssh`) verifica il pin SHA256 della host key; handshake e comandi hanno deadline esplicite. Le deadline non sono timeout di inattività del terminale né garanzie di terminazione remota. Privilegi sudoers per iptables da verificare in laboratorio.

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
- 2026-10-17 · working tree · SSH-HARDEN-01: pinning host distinto dalla chiave client e deadline SSH; vedere contratto in apertura. E28-S1-T1: lock locali, validazione PID e protezione timer precedenti preservati; garanzie Linux/multiistanza restano da verificare.
- 2026-10-10 (Epic 24 SEC-P0: Hardening watchdog firewall nohup e stream-ticket WebSocket integrati e verificati al 100%)

