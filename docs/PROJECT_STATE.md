# PROJECT STATE

> Memoria viva dell'agente. Letta a inizio sessione, aggiornata a fine di ogni turno.
> Non cancellare sezioni: se vuote, scrivi "nessuno".

**Ultimo aggiornamento:** 2026-10-10 17:05  
**Modalità progetto:** esistente (onboarded il 2026-10-09, validato dall'utente)  
**Fase:** Validazione Epic 24 (SEC-P0 Hardening Critico) & Pianificazione Epic 25 (SEC-P1)  

## Project Profile

```
Nome progetto:           NextGen-CyberSecurity Platform
Linguaggi/stack:         Golang 1.22+ (Chi Router, pgx/v5, gorilla/websocket, robfig/cron) + React 18 / TypeScript 5.6 (Vite, Tailwind CSS, xterm.js, Recharts)
Gestore pacchetti:       Go Modules (backend), npm (frontend)
Comando test mirati:     go test -v ./internal/<package>/... (backend), npx vitest run src/<path>/<file>.test.ts (frontend)
Comando suite completa:  go test ./... (backend), npm test (frontend)
Comando build/typecheck: go build ./... (backend), npm run build (frontend)
Comando lint/format:     go vet ./... (backend)
Percorsi per livello:    unit=backend/pkg, backend/internal/domain, backend/internal/scanner; service=backend/internal/service; integration=backend/internal/delivery, backend/internal/repository/postgres; frontend=frontend/src/tests, frontend/src/**/*.test.tsx
Prerequisiti ambiente:   Docker & Docker Compose (PostgreSQL 16, Nmap su Linux container)
Come avviare l'app:      docker compose up --build (oppure 'go run ./cmd/server' + 'cd frontend && npm run dev')
Branch principale:       main (sprint attivo su feat/FEAT-SCAN-PROFILES-custom-categories)
Comandi vietati:         rm -rf fuori repo, git reset --hard non confermati, git push --force, DROP database senza conferma esplicita
```

## Baseline test (prima di ogni intervento)

| Data | Comando | Esito | Fallimenti preesistenti noti |
|---|---|---|---|
| 2026-10-10 17:00 | `go test ./...` (backend) | PASS (exit code 0, 14 package) | Nessuno. Tutti i test passano al 100%. |
| 2026-10-10 17:00 | `go build ./... ; go vet ./...` (backend) | PASS (exit code 0) | Nessuno. Compilazione pulita e zero lint warning. |
| 2026-10-10 17:01 | `npm test` (frontend Vitest) | PASS (exit code 0, 16 file, 97 test) | Nessuno. Tutti i test passano al 100% con **0 warning stderr**. |
| 2026-10-10 16:45 | `npm run build` (frontend tsc + Vite) | PASS (exit code 0, 7.13s) | Nessuno. Compilazione TypeScript pulita, chunk `AttackPath` 28 kB. |

## In corso

- (nessuno - Epic 24 SEC-P0 implementata e verificata con successo)

## Accettati recentemente dal PO (`[x]`)

- **FEAT-REPORT-PDF-UX-01 · Perfezionamento Grafico, Tipografico & Coerenza Dati dei Report PDF A4 (FR-18, UC-17)**:
  - Validato e accettato dal PO in data 2026-10-10.
  - Risolti tutti i difetti grafici dei PDF A4 (Executive compresso a 2 pagine dense senza buchi, wrapping multiriga dinamico senza troncamenti né font miniaturizzati, separazione riga CVE e Azione, colori semantici coerenti, scoring multi-host normalizzato con salvaguardia 0-critical, software reali dai banner, flusso continuo tecnico fino a 270mm, date e durate valorizzate).
- **FEAT-IP-ORG-01 .. 04 · Workstation Multi-Company Workflow**:
  - Validato e accettato dal PO in data 2026-10-10.
  - Sidebar organizzata per azioni operative, raggruppamento per cartella/cliente in Scansioni, Vulnerabilità e Report con posture score e download 1-click.
- **FEAT-CVE-ONLINE-01 · Ricerca Vulnerabilità Online NIST NVD 2.0 & Script Nmap (`--script=vuln`) (FR-17, UC-16)**:
  - Validato e accettato dal PO in data 2026-10-10.
  - Script NSE vuln/vulners, estrazione automatica CVE, client REST NIST NVD API 2.0 con auto-caching in PostgreSQL locale.
- **FEAT-WORKFLOW-CLEANUP-01 .. 07 · Ottimizzazione Flusso Operativo, Riorganizzazione Sequenziale & Consolidamento per Azienda**:
  - Validato e accettato dal PO in data 2026-10-10.
  - Dossier report consolidati per azienda (eliminati report singoli IP), sidebar sequenziale Passo 1 -> 5, scansioni esclusive per azienda in `/scans`, eliminazione massiva checkbox in `/targets`, Network Discovery potenziata con creazione azienda e cache tenant.
- **FEAT-SCAN-PROF-01 .. 04 · Personalizzazione Avanzata Profili, Profilo Generalista 360°, Multi-Obiettivo & Guida Semplice (FR-16)**:
  - Validato e accettato dal PO in data 2026-10-10.
  - Profilo 360° "Tutto in Uno", modalità "Componi Multi-Obiettivo" via checkbox, Guida Semplice interattiva con spiegazioni per il PO, modalità Professionale con live command preview.
- **FEAT-M2-02 · Motore di Calcolo Lateral Movement & API Grafo (Backend Go)**:
  - Validato e accettato dal PO in data 2026-10-10.
  - Servizio di aggregazione topologica per calcolo lateral movement, chokepoint di sicurezza ed endpoint `/api/v1/attack-paths/graph` e `/api/v1/attack-paths/chokepoints`.

## In attesa di accettazione utente

- **Epic 24 · SEC-P0: Hardening Critico & Scanner Security Perimeter (SEC-P0-01 .. SEC-P0-06)**:
  - **SEC-P0-01 (Deny-List & Scope Check)**: Blocco categorico all'inserimento e al lancio di scansioni/discovery verso loopback (`127.0.0.0/8`, `::1`), link-local (`169.254.0.0/16`, `fe80::/10`), bridge Docker (`172.17.0.0/16` .. `172.20.0.0/16`), hostname interni (`localhost`, `postgres`, `backend`, `.internal`, `.local`), CIDR `/8` e risoluzione DNS anti-rebinding con `net.LookupIP`.
  - **SEC-P0-02 (Stream-Ticket Monouso per SSE & WebSocket)**: Sostituzione dei JWT raw a lungo termine nei query param con ticket usa-e-getta monouso (TTL 30s) emessi tramite `POST /api/v1/auth/stream-ticket` protetto. Handshake per `/api/v1/scans/{id}/stream`, `/api/v1/network-discovery/stream` e terminale WebSocket `/ws/v1/ssh/servers/{id}/terminal` migrati a `?ticket=...` con fallback trasparente. Zero cred leakage nei log Nginx e browser.
  - **SEC-P0-03 (Delimitatore `--` POSIX Nmap)**: Inserimento sistematico di `--` prima del target in `BuildSafeArguments` per rendere impossibile la flag injection via nomi ostili; argomenti tipizzati e validati.
  - **SEC-P0-04 (Default Offline & Safe Scripts)**: Disabilitazione di default per lo script esterno `vulners` dai profili di scansione; esecuzione rigorosa di script sicuri e locali (`--script "vuln and safe"`) conforme agli standard di privacy aziendale.
  - **SEC-P0-05 (Redazione Credenziali Debug Console)**: Mascheramento automatico e ricorsivo di token, ticket, passphrases, chiavi e header Bearer nei log del drawer e nell'esportatore Markdown 1-click per l'AI.
  - **SEC-P0-06 (Firewall Rollback Watchdog Remoto)**: Esecuzione del timer anti-lockout direttamente sull'host remoto via `nohup` distaccato con supporto sudo (`nohup sh -c 'sleep 60 && (iptables-restore < backup || sudo iptables-restore < backup)' &`), garantendo il rollback anche a fronte di disconnessione o caduta del backend.

- **FEAT-M2-03 · Componente UI Visual Graph Explorer (Frontend React) (FR-14, UC-13)**:
  - Canvas SVG interattivo con controlli zoom/pan e layout topologico per livelli di penetrazione.
  - Highlighting dinamico dei percorsi di attacco con attenuazione nodi esterni e selezione nodi/chokepoints.
  - Filtri CVSS interattivi, chokepoint cards con rimedio 1-click e code-split a 28 kB.

- **ONBOARD-04 · Risoluzione Warning nei Test Vitest (act & Recharts container)**:
  - Eliminazione totale dei warning stderr nei test frontend tramite `setup.ts` e cleanup globale (16 file, 97 test, 0 warning).

- **ONBOARD-05 · Test di Caratterizzazione per Scenari Limite Nmap & Rete**:
  - Test isolati per simulare interruzione brusca (SIGKILL/exit 137), XML troncato, XML corrotto, overflow stderr e fallback a FAILED senza panic.

- **ONBOARD-02 · Automated Container E2E Test Runner**:
  - Runner unificato multipiattaforma `scripts/run_e2e_docker.sh` e `scripts/run_e2e_docker.ps1` per avvio Compose, validazione REST, SSE, frontend e Attack Path.

## Bloccati / Domande aperte per l'utente

- (nessuno)

## Assunzioni confermate

- Struttura documentale autonoma v6.0 approvata dal Product Owner.
- Archiviazione storica dei file legacy in `docs/_legacy/` confermata.
- Pipeline CI/CD GitHub Actions e bundle splitting implementati.
- Milestone 2 completata sia su Backend Go che su Frontend React.
- Roadmap di hardening sicurezza (Epiche SEC-P0, SEC-P1, SEC-P2) approvata operativamente.

## Decisioni recenti (ultime 10, dettagli negli ADR)

- **ADR-04 (2026-10-10):** Stream-Ticket monouso memoria thread-safe (TTL 30s) per SSE e WebSocket; delimitatore POSIX `--` in Nmap; deny-list target interni e Docker bridge con anti-rebinding DNS; nohup watchdog remoto firewall rollback.
- **PO-04 (2026-10-10):** Approvazione cumulativa globale di tutti i task di flusso, profili, CVE online e motore lateral movement (`FEAT-SCAN-PROF`, `FEAT-WORKFLOW-CLEANUP`, `FEAT-CVE-ONLINE`, `FEAT-REPORT-PDF-UX-01`, `FEAT-M2-02`).
- **PO-01 (2026-10-09):** Approvazione implementazione pipeline CI/CD GitHub Actions per backend e frontend.
- **PO-02 (2026-10-09):** Approvazione ottimizzazione bundle frontend tramite code-splitting e `React.lazy`.
- **PO-03 (2026-10-09):** Priorità strategica Milestone 2 assegnata a Attack Path Mapping.
- **ADR-01:** Modularizzazione di `2_ARCHITECTURE.md` con scomposizione per domini in `docs/domains/`.
- **ADR-02:** Preservazione della cronologia Git via `git mv` per file storici e tracciamento in `docs/_legacy/MAPPING.md`.
- **ADR-03:** Creazione del branch dedicato `chore/agent-onboarding` per modifiche strutturali.

## Prossimi passi

1. Accettazione da parte del PO del Pacchetto di Accettazione per Epic 24 (SEC-P0).
2. Avvio esecutivo dell'Epic 25 (SEC-P1 Scanner Engine Overhaul & Profiling Accuracy):
   - SEC-P1-01: Profilo Scansione UDP Specialistico (`-sU -p U:53,69,123,161,514,623,5353`)
   - SEC-P1-02: Correzione & Ricalibrazione profili Nmap (porta 69 UDP, stati ACK filtered/unfiltered, alive-only con -Pn)
   - SEC-P1-03: Tassonomia di invasività (`SAFE` vs `INTRUSIVE`)
   - SEC-P1-04: Modello dati Finding (`confidence`, `detection_method`, `cpe`)
   - SEC-P1-05: Container capabilities hardening (`NET_ADMIN` rimosso, `NMAP_PRIVILEGED=1`)
   - SEC-P1-06: Manutenzione dipendenza PDF (`github.com/go-pdf/fpdf`)

## Log sessioni (ultime 5, la più recente in alto)

- 2026-10-10 17:05 · Correzione anomalia HTTP 500 su target interni protetti: rimappati `domain.ErrTargetInternalRestricted` e `domain.ErrTargetScopeTooBroad` su HTTP 400 Bad Request espliciti (`TARGET_RESTRICTED` e `SCOPE_TOO_BROAD`) in `discovery_handler.go`, `target_handler.go` e `scan_handler.go`; aggiunto controllo preventivo `CheckTargetScope` in `target_service.go` e `scan_service.go`; introdotto bypass configurabile via env `ALLOW_INTERNAL_TARGETS=true` / `DEV_ALLOW_LOCALHOST=true` per consentire test locali di sviluppo in piena sicurezza · Test suite backend (14 pkg) e frontend (97 test) 100% VERDI
- 2026-10-10 16:50 · Implementazione e verifica completa di Epic 24 (SEC-P0 Hardening Critico): sviluppato `StreamTicketStore` monouso per SSE e WebSocket con TTL 30s ed endpoint protetto `/api/v1/auth/stream-ticket`; implementato `CheckTargetScope` con deny-list IP/CIDR/hostname (loopback, RFC1918 docker bridge, link-local, cloud metadata, .internal) e risoluzione DNS anti-rebinding; inserito delimitatore POSIX `--` prima del target in `BuildSafeArguments` Nmap; disabilitato per default script esterno `vulners`; aggiunta regex di redazione query parameters sensibili (`ticket=`, `token=`, `passphrase=`) nella Debug Console e nell'esportatore Markdown AI; aggiunto watchdog di rollback remoto distaccato via `nohup` e `sudo iptables-restore` per firewall; risolto disallineamento cleanup async in `WebTerminalModal.tsx` · Test backend 14/14 pkg PASS (0 errori, build/vet OK), Frontend 16 file / 97 test PASS, npm run build OK (0 errori)
- 2026-10-10 16:25 · Valutazione approfondita dell'Audit di Sicurezza & Architettura (10 Ottobre 2026): allineamento immediato della documentazione (inserito requisito legale FR-01.3 di autorizzazione scansione e FR-01.4 deny-list in `0_REQUIREMENTS.md`, aggiornati i perimetri in/out-of-scope in `1_PROJECT_CONTEXT.md`, completati gli attributi mancanti nell'ERD di `2_ARCHITECTURE.md`), strutturazione del backlog di remediation in tre Epiche formalizzate in `docs/3_TASK_LIST.md` (Epic 24 SEC-P0, Epic 25 SEC-P1, Epic 26 SEC-P2) · Test suite backend (14 pkg) e frontend (95 test) 100% VERDI
- 2026-10-10 16:05 · Audit e sincronizzazione totale della cartella `docs/domains/`: aggiornati `scanning_and_discovery.md`, `reporting_and_assets.md`, `attack_path.md`, `vulnerability_intelligence.md`; allineate le rotte e i riferimenti modulari in `docs/2_ARCHITECTURE.md` · Suite backend (14 pkg) e frontend (95 test) 100% VERDI
- 2026-10-10 15:30 · Completamento totale del backlog attivo e di robustezza (FEAT-M2-03, ONBOARD-04, ONBOARD-05, ONBOARD-02): sviluppato Visual Graph Explorer interattivo SVG per Milestone 2; azzerati tutti i warning console dei test Vitest; aggiunti test di caratterizzazione per scenari limite Nmap & rete; creati runner E2E containerizzati unificati · Backend Go 14/14 pkg PASS, Frontend 95/95 test PASS
- 2026-10-10 14:00 · Esecuzione FEAT-M2-03 (Milestone 2 Attack Path Mapping completata al 100%): sviluppato Visual Graph Explorer SVG interattivo con controlli zoom/pan e layout topologico per livelli di penetrazione; evidenziazione dinamica dei percorsi di attacco con attenuazione nodi esterni; identificazione visiva dei chokepoints con aura bullseye ed evidenza remediation a massimo ROI · Frontend Vitest 16/16 file (95/95 test PASS), npm run build OK, Backend Go 14/14 pkg PASS, build & vet 100% VERDI

