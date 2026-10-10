# PROJECT STATE

> Memoria viva dell'agente. Letta a inizio sessione, aggiornata a fine di ogni turno.
> Non cancellare sezioni: se vuote, scrivi "nessuno".

**Ultimo aggiornamento:** 2026-10-10 17:49  
**Modalità progetto:** esistente (onboarded il 2026-10-09, validato dall'utente)  
**Fase:** Pulizia branch obsoleti & Validazione Epic 25 (SEC-P1) ed Epic 26 (SEC-P2)  

## Project Profile

```
Nome progetto:           NextGen-CyberSecurity Platform
Linguaggi/stack:         Golang 1.22+ (Chi Router, pgx/v5, gorilla/websocket, robfig/cron, go-pdf/fpdf v0.9.0) + React 18 / TypeScript 5.6 (Vite, Tailwind CSS, xterm.js, Recharts)
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
| 2026-10-10 17:34 | `go test ./...` (backend) | PASS (exit code 0, 14 package) | Nessuno. Tutti i test passano al 100%. |
| 2026-10-10 17:34 | `go build ./... ; go vet ./...` (backend) | PASS (exit code 0) | Nessuno. Compilazione pulita e zero lint warning. |
| 2026-10-10 17:34 | `npm test` (frontend Vitest) | PASS (exit code 0, 16 file, 99 test) | Nessuno. Tutti i test passano al 100% con **0 warning stderr**. |
| 2026-10-10 17:34 | `npm run build` (frontend tsc + Vite) | PASS (exit code 0, 7.70s) | Nessuno. Compilazione TypeScript pulita, chunk `AttackPath` 28 kB. |

## In corso

- (nessuno - Epic 25 SEC-P1 ed Epic 26 SEC-P2 completate e verificate con successo)

## Accettati recentemente dal PO (`[x]`)

- **FEAT-REPORT-PDF-UX-01 · Perfezionamento Grafico, Tipografico & Coerenza Dati dei Report PDF A4 (FR-18, UC-17)**:
  - Validato e accettato dal PO in data 2026-10-10.
- **FEAT-IP-ORG-01 .. 04 · Workstation Multi-Company Workflow**:
  - Validato e accettato dal PO in data 2026-10-10.
- **FEAT-CVE-ONLINE-01 · Ricerca Vulnerabilità Online NIST NVD 2.0 & Script Nmap (`--script=vuln`) (FR-17, UC-16)**:
  - Validato e accettato dal PO in data 2026-10-10.
- **FEAT-WORKFLOW-CLEANUP-01 .. 07 · Ottimizzazione Flusso Operativo, Riorganizzazione Sequenziale & Consolidamento per Azienda**:
  - Validato e accettato dal PO in data 2026-10-10.
- **FEAT-SCAN-PROF-01 .. 04 · Personalizzazione Avanzata Profili, Profilo Generalista 360°, Multi-Obiettivo & Guida Semplice (FR-16)**:
  - Validato e accettato dal PO in data 2026-10-10.
- **FEAT-M2-02 · Motore di Calcolo Lateral Movement & API Grafo (Backend Go)**:
  - Validato e accettato dal PO in data 2026-10-10.

## In attesa di accettazione utente

- **Epic 25 · SEC-P1: Scanner Engine Overhaul & Profiling Accuracy (SEC-P1-01 .. SEC-P1-06)**:
  - **SEC-P1-01 (Profilo UDP Specialistico)**: Nuovo profilo dedicato `UDP Specialized Services & Critical Infra` (`-sU -p U:53,69,123,161,514,623,5353 -T4`) per DNS, TFTP, NTP, SNMP, Syslog, IPMI e mDNS, isolato dalle code TCP per azzerare i colli di bottiglia. Supporto sintassi mista UDP/TCP `-p U:...,T:...`.
  - **SEC-P1-02 (Ricalibrazione Profili Nmap)**: Rimossa la porta 69 (UDP) dal profilo `Legacy Protocols` TCP per evitare timeout spuri su socket stream; rimossa flag `-n` da LAN Discovery per abilitare la reverse DNS resolution automatica; gestione raffinata dello stato porta `unfiltered` e `open|filtered` per evitare falsi positivi di porta aperta su scansioni firewall ACK.
  - **SEC-P1-03 (Tassonomia di Invasività & Gate di Sicurezza)**: Classificazione di tutti i profili in `SAFE`, `INTRUSIVE`, `DANGEROUS`. Badge semantico visivo in `NewScanModal.tsx` e blocco obbligatorio al lancio con checkbox esplicito di autorizzazione prima di avviare profili ad alto impatto di rete.
  - **SEC-P1-04 (Arricchimento Modello Finding)**: Aggiunta al modello dati e al database dei campi `confidence` (`CONFIRMED`, `PROBABLE`, `UNVERIFIED`), `detection_method` (`nse`, `banner`, `nvd_keyword`, `port_state`), `cpe` e `port_state` (migrazione SQL `000007_scanner_overhaul_findings`).
  - **SEC-P1-05 (Hardening Container Capabilities)**: Concessione minima stretta in `docker-compose.yml` (`CAP_NET_RAW` per SYN scan, rimosso `CAP_NET_ADMIN`), abilitazione env `NMAP_PRIVILEGED=1`.
  - **SEC-P1-06 (Migrazione Libreria PDF a `go-pdf/fpdf`)**: Sostituzione del package deprecato `jung-kurt/gofpdf` con il fork mantenuto `github.com/go-pdf/fpdf v0.9.0`.

- **Epic 26 · SEC-P2: Attack Path Refinement & Advanced Orchestration (SEC-P2-01 .. SEC-P2-03)**:
  - **SEC-P2-01 (Coerenza CVSS/Severity & Archi Inferiti)**: Allineamento rigoroso tra punteggio CVSS dei nodi e severità visiva (Low <4.0, Med 4.0-6.9, High 7.0-8.9, Crit >=9.0); marcatura esplicita degli archi `PIVOTS_TO` come inferiti (`inferred: true`, `confidence: 0.85`), rendering con tratto tratteggiato `strokeDasharray="6 4"`, badge `INFERRED xx%` e tooltip con base euristica.
  - **SEC-P2-02 (Formula Rischio Probabilistica & Bound DFS)**: Formula cumulativa basata sulla teoria dell'affidabilità $1 - \prod(1 - p_i)$ con fattore hop attenuato e salvaguardia $\ge 85$ per CVE critiche; limite combinatorio a 50 percorsi massimi (`maxPaths = 50`) e profondità 6 nodi.
  - **SEC-P2-03 (Diff tra Scansioni: NEW, RESOLVED, PERSISTENT)**: Endpoint `GET /api/v1/scans/{id}/diff` che calcola in tempo reale il differenziale tra la scansione corrente e la precedente per lo stesso target, identificando finding nuovi (`NEW`), persistenti (`PERSISTENT`) e confermati come risolti (`RESOLVED`).

- **Epic 24 · SEC-P0: Hardening Critico & Scanner Security Perimeter (SEC-P0-01 .. SEC-P0-06)**:
  - Deny-list scope check, stream-ticket monouso SSE/WS, delimitatore POSIX `--`, safe scripts, redazione credenziali debug, nohup watchdog rollback.

## Bloccati / Domande aperte per l'utente

- (nessuno)

## Assunzioni confermate

- Struttura documentale autonoma v6.0 approvata dal Product Owner.
- Archiviazione storica dei file legacy in `docs/_legacy/` confermata.
- Pipeline CI/CD GitHub Actions e bundle splitting implementati.
- Milestone 2 completata sia su Backend Go che su Frontend React.
- Roadmap di hardening sicurezza (Epiche SEC-P0, SEC-P1, SEC-P2) implementata al 100%.

## Decisioni recenti (ultime 10, dettagli negli ADR)

- **ADR-06 (2026-10-10):** Attack Path Probabilistic Risk Formula $1 - \prod(1 - p_i)$, bound 50 cammini, archi inferiti tratteggiati con confidenza, e Scan Diff `GET /api/v1/scans/{id}/diff` con stati NEW/RESOLVED/PERSISTENT.
- **ADR-05 (2026-10-10):** Profilo specialistico UDP, tassonomia di invasività con gate UI, modello finding arricchito con confidence/detection_method/cpe, hardening container capabilities (solo `CAP_NET_RAW`), migrazione PDF a `go-pdf/fpdf v0.9.0`.
- **ADR-04 (2026-10-10):** Stream-Ticket monouso memoria thread-safe (TTL 30s) per SSE e WebSocket; delimitatore POSIX `--` in Nmap; deny-list target interni e Docker bridge con anti-rebinding DNS; nohup watchdog remoto firewall rollback.
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

