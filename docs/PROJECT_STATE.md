# PROJECT STATE

> Memoria viva dell'agente. Letta a inizio sessione, aggiornata a fine di ogni turno.
> Non cancellare sezioni: se vuote, scrivi "nessuno".

**Ultimo aggiornamento:** 2026-10-10 18:24  
**Modalità progetto:** esistente (onboarded il 2026-10-09, validato dall'utente)  
**Fase:** Tutte le Epiche (Milestone 2, SEC-P0, SEC-P1, SEC-P2) convalidate e accettate dal PO. Sistema pronto per nuove feature.  

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
| 2026-10-10 18:14 | `go test ./...` (backend) | PASS (exit code 0, 14 package) | Nessuno. Tutti i test passano al 100%. |
| 2026-10-10 18:14 | `go build ./... ; go vet ./...` (backend) | PASS (exit code 0) | Nessuno. Compilazione pulita e zero lint warning. |
| 2026-10-10 18:14 | `npm test` (frontend Vitest) | PASS (exit code 0, 16 file, 99 test) | Nessuno. Tutti i test passano al 100% con **0 warning stderr**. |
| 2026-10-10 18:14 | `npm run build` (frontend tsc + Vite) | PASS (exit code 0, 7.53s) | Nessuno. Compilazione TypeScript pulita, chunk `AttackPath` 28.88 kB. |

## In corso

- (nessuno - Tutti i task pianificati sono stati completati, convalidati e accettati dal PO)

## Accettati recentemente dal PO (`[x]`)

- **Epic 26 · SEC-P2: Attack Path Refinement & Advanced Orchestration (SEC-P2-01 .. SEC-P2-03)**:
  - Validato e accettato dal PO in data 2026-10-10.
- **Epic 25 · SEC-P1: Scanner Engine Overhaul & Profiling Accuracy (SEC-P1-01 .. SEC-P1-06)**:
  - Validato e accettato dal PO in data 2026-10-10.
- **Epic 24 · SEC-P0: Hardening Critico & Scanner Security Perimeter (SEC-P0-01 .. SEC-P0-06)**:
  - Validato e accettato dal PO in data 2026-10-10.
- **Milestone 2 · Visual Graph Explorer & Robustezza Suite (FEAT-M2-03, ONBOARD-01 .. ONBOARD-05)**:
  - Validato e accettato dal PO in data 2026-10-10.
- **FEAT-REPORT-PDF-UX-01 · Perfezionamento Grafico, Tipografico & Coerenza Dati dei Report PDF A4 (FR-18, UC-17)**:
  - Validato e accettato dal PO in data 2026-10-10.

## In attesa di accettazione utente

- (nessuno)

## Bloccati / Domande aperte per l'utente

- (nessuno)

## Assunzioni confermate

- Struttura documentale autonoma v6.0 approvata dal Product Owner.
- Archiviazione storica dei file legacy in `docs/_legacy/` e [4_CHANGELOG.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/4_CHANGELOG.md) confermata.
- Pipeline CI/CD GitHub Actions e bundle splitting implementati.
- Milestone 2 completata sia su Backend Go che su Frontend React.
- Roadmap di hardening sicurezza (Epiche SEC-P0, SEC-P1, SEC-P2) implementata al 100%.

## Decisioni recenti (ultime 10, dettagli negli ADR)

- **ADR-07 (2026-10-10):** Integrazione serializzazione `Invasiveness` in `ProfileResponse` API e seeding profilo `All-in-One 360° Comprehensive Audit`.
- **ADR-06 (2026-10-10):** Attack Path Probabilistic Risk Formula $1 - \prod(1 - p_i)$, bound 50 cammini, archi inferiti tratteggiati con confidenza, e Scan Diff `GET /api/v1/scans/{id}/diff` con stati NEW/RESOLVED/PERSISTENT.
- **ADR-05 (2026-10-10):** Profilo specialistico UDP, tassonomia di invasività con gate UI, modello finding arricchito con confidence/detection_method/cpe, hardening container capabilities (solo `CAP_NET_RAW`), migrazione PDF a `go-pdf/fpdf v0.9.0`.
- **ADR-04 (2026-10-10):** Stream-Ticket monouso memoria thread-safe (TTL 30s) per SSE e WebSocket; delimitatore POSIX `--` in Nmap; deny-list target interni e Docker bridge con anti-rebinding DNS; nohup watchdog remoto firewall rollback.
- **PO-01 (2026-10-09):** Approvazione implementazione pipeline CI/CD GitHub Actions per backend e frontend.
- **PO-02 (2026-10-09):** Approvazione ottimizzazione bundle frontend tramite code-splitting e `React.lazy`.
- **PO-03 (2026-10-09):** Priorità strategica Milestone 2 assegnata a Attack Path Mapping.

## Prossimi passi

1. In attesa delle nuove direttive, priorità o specifiche per la Milestone 3 da parte del Product Owner.

## Log sessioni (ultime 5, la più recente in alto)

- 2026-10-10 18:14 · Verifica visuale e funzionale E2E automatizzata con Google Chrome headless via Puppeteer: testate con successo Dashboard, Attack Path Graph (`/attack-path`), Scansioni (`/scans`), Modale Nuova Scansione (`NewScanModal`), Reports (`/reports`) e Live Debug Console drawer (`>_`); rilevata e corretta l'omissione del campo `invasiveness` in `ProfileResponse` (`scan_handler.go`) e inserito profilo `All-in-One 360° Comprehensive Audit` nel DB; convalidate le visualizzazioni dei badge `SAFE`/`INTRUSIVE`, del profilo `Specialized UDP Critical Services` e del gate di autorizzazione obbligatoria con checkbox; catturati screenshot visuali conformi in artifacts · Backend 14/14 pkg PASS, Frontend 16 file / 99 test PASS al 100%
- 2026-10-10 17:05 · Correzione anomalia HTTP 500 su target interni protetti: rimappati `domain.ErrTargetInternalRestricted` e `domain.ErrTargetScopeTooBroad` su HTTP 400 Bad Request espliciti (`TARGET_RESTRICTED` e `SCOPE_TOO_BROAD`) in `discovery_handler.go`, `target_handler.go` e `scan_handler.go`; aggiunto controllo preventivo `CheckTargetScope` in `target_service.go` e `scan_service.go`; introdotto bypass configurabile via env `ALLOW_INTERNAL_TARGETS=true` / `DEV_ALLOW_LOCALHOST=true` per consentire test locali di sviluppo in piena sicurezza · Test suite backend (14 pkg) e frontend (97 test) 100% VERDI
- 2026-10-10 16:50 · Implementazione e verifica completa di Epic 24 (SEC-P0 Hardening Critico): sviluppato `StreamTicketStore` monouso per SSE e WebSocket con TTL 30s ed endpoint protetto `/api/v1/auth/stream-ticket`; implementato `CheckTargetScope` con deny-list IP/CIDR/hostname (loopback, RFC1918 docker bridge, link-local, cloud metadata, .internal) e risoluzione DNS anti-rebinding; inserito delimitatore POSIX `--` prima del target in `BuildSafeArguments` Nmap; disabilitato per default script esterno `vulners`; aggiunta regex di redazione query parameters sensibili (`ticket=`, `token=`, `passphrase=`) nella Debug Console e nell'esportatore Markdown AI; aggiunto watchdog di rollback remoto distaccato via `nohup` e `sudo iptables-restore` per firewall; risolto disallineamento cleanup async in `WebTerminalModal.tsx` · Test backend 14/14 pkg PASS (0 errori, build/vet OK), Frontend 16 file / 97 test PASS, npm run build OK (0 errori)
- 2026-10-10 16:25 · Valutazione approfondita dell'Audit di Sicurezza & Architettura (10 Ottobre 2026): allineamento immediato della documentazione (inserito requisito legale FR-01.3 di autorizzazione scansione e FR-01.4 deny-list in `0_REQUIREMENTS.md`, aggiornati i perimetri in/out-of-scope in `1_PROJECT_CONTEXT.md`, completati gli attributi mancanti nell'ERD di `2_ARCHITECTURE.md`), strutturazione del backlog di remediation in tre Epiche formalizzate in `docs/3_TASK_LIST.md` (Epic 24 SEC-P0, Epic 25 SEC-P1, Epic 26 SEC-P2) · Test suite backend (14 pkg) e frontend (95 test) 100% VERDI
- 2026-10-10 16:05 · Audit e sincronizzazione totale della cartella `docs/domains/`: aggiornati `scanning_and_discovery.md`, `reporting_and_assets.md`, `attack_path.md`, `vulnerability_intelligence.md`; allineate le rotte e i riferimenti modulari in `docs/2_ARCHITECTURE.md` · Suite backend (14 pkg) e frontend (95 test) 100% VERDI
- 2026-10-10 14:00 · Esecuzione FEAT-M2-03 (Milestone 2 Attack Path Mapping completata al 100%): sviluppato Visual Graph Explorer SVG interattivo con controlli zoom/pan e layout topologico per livelli di penetrazione; evidenziazione dinamica dei percorsi di attacco con attenuazione nodi esterni; identificazione visiva dei chokepoints con aura bullseye ed evidenza remediation a massimo ROI · Frontend Vitest 16/16 file (95/95 test PASS), npm run build OK, Backend Go 14/14 pkg PASS, build & vet 100% VERDI

