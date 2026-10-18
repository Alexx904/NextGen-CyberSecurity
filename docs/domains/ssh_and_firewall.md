# Dominio: Remote SSH Host & iptables/nftables Firewall Manager

Decisione PO 2026-10-18: supporto iptables e iptables-nft con modello corrente
INPUT/OUTPUT/FORWARD. Nessun gestore nativo di tabelle/catene nftables richiesto.
Compatibilità iptables-nft e iptables-legacy verificata nel laboratorio descritto
sotto; nessuna prova su host cliente.

## M11-FIREWALL-01 · Remote transaction · 2026-10-18 · working tree

Supersedes the local timer/PID lifecycle described by earlier increments.
Snapshot, watchdog readiness and mutation run under one remote `flock` shared
by all application instances using the same managed SSH UID. Private state in
`/tmp/nextgen-fw-<uid>` must belong to that UID with mode 0700 and no symlink.
Random transaction IDs bind pending state to tenant/server. Other accounts,
administrators, host reboot and removal of temporary state are outside this
coordination/persistence boundary.

The watchdog starts before mutation, survives SSH closure, and restores under
the same lock at expiry. Failed restores retain pending state and retry. Confirm
checks the tenant/server DB backup and removes the matching pending marker;
it never kills a PID. Remote application precedes DB backup persistence; DB
failure returns an error and leaves rollback armed. Rollback streams the raw
snapshot on SSH stdin, without evaluating it as shell source. A local timeout
does not imply that remote execution stopped.

Remote prerequisites: sh, flock, stat, nohup, date, sha256sum, awk, core file/sleep
utilities and iptables/save/restore. Root executes directly; other SSH users need
noninteractive sudo for those individual iptables tools, including restore test.
No remote installation or privileged shell is granted automatically.

Revision is SHA256 of canonical `iptables -S`, computed under the remote lock.
New policy/replacement/move operations require the revision; new UI supplies it
for add/delete too. Stale revisions return 409 before snapshot/mutation. Move
preserves complete saved filter rules and other tables; replacement explicitly
uses only the visible form fields and warns that advanced options are replaced.
Policy supports ACCEPT/DROP. Existing SecOps, TenantAdmin and SuperAdmin access
and tenant-owned SSH inventory checks apply.

UI retains expired-but-pending state and polls it, disables confirmation after
expiry, and opens the docked terminal only on explicit action. Closing a terminal
while its ticket is pending cannot open a late WebSocket. Standalone mode remains.
[Plan and measured Linux/OpenSSH checks](../04-development/plans/M11-FIREWALL-01.md).
Browser proves expiry, confirmation, moving, policy and docked SSH I/O on the
authorized isolated fixture. Numeric iptables-nft protocol IDs 0/1/6/17 normalize
to all/icmp/tcp/udp; unknown numbers remain read-only. Replacement on the corrected
source and company-switch regression pass in browser. Task `[~]`, awaits PO.
FR-13 follow-up records WS state and frame-size metadata, without terminal content
or ticket query; [diagnostic contract](debug_console.md). Real WS filter/export PASS.

## M11-CLOSE-01 · Validazione mutazioni · 2026-10-18 · working tree

Contratto del primo incremento: IPv4/CIDR IPv4 letterali per source/destination;
interfacce ASCII massimo 15 caratteri (`_`, `.`, `-`, `:`, wildcard `+` finale),
vuoto/`*` senza filtro; commenti UTF-8 massimo 255 byte senza controlli e quotati
come singolo argomento shell. Il timer è obbligatorio per add/delete anche con
flag legacy false; timeout 0 = 60s, intervallo accettato 10..300s. Input invalidi
rifiutati prima di ogni comando SSH. Nessuna nuova garanzia watchdog Linux:
The later remote transaction above addresses the lifecycle; its browser gate
remains open. [Historical first-increment plan](../04-development/plans/M11-CLOSE-01.md).

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

The deadline increment alone did not resolve TECH-DEBT-07; see the later remote
transaction and its separate Linux/OpenSSH evidence above. E28-S1-T1 requires
its own complete remediation gates.
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
  1. Remote lock protects snapshot, watchdog readiness and mutation; see the
     current transaction contract above.
  2. The successful remote response is persisted in `FIREWALL_BACKUPS` before
     acknowledging the pending mutation to the caller.
  3. The detached watchdog restores after expiry independently of backend memory
     or SSH connection, while host/tools/state remain available.
  4. Explicit confirmation before expiry removes the matching pending marker;
     failed rollback retains protection for retry. No PID cancellation is used.

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

