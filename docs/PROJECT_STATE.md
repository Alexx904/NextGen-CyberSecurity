# PROJECT STATE

> Memoria viva dell'agente. Letta a inizio sessione, aggiornata a fine di ogni turno.
> Non cancellare sezioni: se vuote, scrivi "nessuno".

**Ultimo aggiornamento:** 2026-10-15 17:05  
**Modalità progetto:** esistente (onboarded il 2026-10-09, validato dall'utente)  
**Fase:** Epic 27.1 (Semplificazione Centro Scansioni & Scanning Studio 360°) implementata e verificata al 100%; Pacchetto di Accettazione sottoposto al Product Owner.

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
Branch principale:       main (allineato e sincronizzato)
Comandi vietati:         rm -rf fuori repo, git reset --hard non confermati, git push --force, DROP database senza conferma esplicita
```

## Baseline test (prima di ogni intervento)

| Data | Comando | Esito | Fallimenti preesistenti noti |
|---|---|---|---|
| 2026-10-15 16:45 | `go test ./...` (backend) | PASS (exit code 0, 14 package) | Nessuno. Tutti i package backend passano al 100%. |
| 2026-10-15 16:44 | `npm run build` (frontend tsc + Vite) | PASS (exit code 0, 10.64s) | Nessuno. Compilazione TypeScript pulita, 0 errori, chunk ottimizzati. |
| 2026-10-15 16:43 | `npm test` (frontend Vitest) | PASS (exit code 0, 18 file, 107 test) | Nessuno. Tutti i test passano al 100% con **0 errori** (inclusi 14 test NewScanModal). |

## In corso

- (nessuno)

## Accettati recentemente dal PO (`[x]`)

- **Epic 27 · Gestione Aziende, Reparti & Filtro Globale (E27-S1-T1 .. E27-S1-T6)**:
  - Validato e accettato dal PO in data 2026-10-11.
  - **CompanyContext & Scope Globale:** Selettore Azienda persistente in sidebar desktop/drawer mobile; calcolo automatico di tutti i reparti e IP appartenenti all'azienda attiva.
  - **Dashboard Gestione Aziende (`/companies`):** Nuova interfaccia con card metriche (reparti, IP, postura, rischio), color tag picker, creazione, modifica ed eliminazione con conferma; pulsante per impostare al volo lo scope attivo.
  - **Contestualizzazione Reparti in `/targets`:** Visualizzazione e gestione esclusiva dei soli reparti dell'azienda selezionata, con creazione rapida di sotto-reparti e assegnazione IP.
  - **Propagazione e Pulizia Ridondanze su tutte le feature:** Applicato filtro IP e banner Scope Attivo su Dashboard, Scansioni, Nuova Scansione (filtraggio gruppi e target), Vulnerabilità CVE (ispezione per reparto ed IP), Report (dossier aziendale consolidato) e Attack Path Mapping.
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

- **Epic 27.1 · Semplificazione Centro Scansioni & Scanning Studio 360° (E27.1-S1-T1 .. E27.1-S1-T3)**:
  - Riorganizzazione unificata di `NewScanModal.tsx` con architettura a 2 livelli:
    1. **Livello 1 (1-Click Launch):** 4 Master Presets intelligenti (Audit 360° Sicurezza Generale, Ricognizione Rapida Fast Sweep, Audit Vulnerabilità & CVE Exploit, Perimetro Web App & API Surface) con indicazione visiva di durata, invasività e target focus, oltre a pulsante rapido per aprire e personalizzare i parametri nello Studio.
    2. **Livello 2 (Scanning Studio 360° Fine-Tuning):** Controlli granulari ad alta precisione su scope porte (Top 100, Top 1000, Web, Critiche, Tutte 65535, Custom), slider timing T1-T5, toggle sonde attive (SYN Stealth, ACK Probe, Versioning, Deep Intensity 7, OS Fingerprinting, Vulnerability NSE Scripts, Specialized UDP probe, No Ping, Only Open) e salvataggio custom profiles.
  - **Calcolo Dinamico dell'Invasività in Tempo Reale:** Invasività valutata reattivamente tra `SAFE` e `INTRUSIVE` con Live Invasiveness Meter.
  - **Gate di Sicurezza & Consenso Obbligatorio (`SEC-P1-03`, `FR-01.3`):** Blocco tassativo dell'avvio scansione su parametri invasivi finché l'operatore non spunta la presa d'atto e autorizzazione scritta.
  - **Verifica Zero-Assumption:** 14/14 test unitari Vitest su NewScanModal PASS, suite completa frontend 18/18 file e 107/107 test PASS, compilazione TypeScript/Vite pulita in 10.64s, suite completa backend Go 14/14 package PASS.

## Bloccati / Domande aperte per l'utente

- (nessuno)

## Assunzioni confermate

- Struttura documentale autonoma v6.0 approvata dal Product Owner.
- Archiviazione storica dei file legacy in `docs/_legacy/` e [4_CHANGELOG.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/4_CHANGELOG.md) confermata.
- Pipeline CI/CD GitHub Actions e bundle splitting implementati.
- Milestone 2 completata sia su Backend Go che su Frontend React.
- Roadmap di hardening sicurezza (Epiche SEC-P0, SEC-P1, SEC-P2) implementata al 100%.

## Decisioni recenti (ultime 10, dettagli negli ADR)

- **ADR-10 (2026-10-15):** Standard Operativo Software House: Context Layering su 4 livelli gerarchici per contenimento token bloat e azzeramento allucinazioni; Mappa Documenti Full-Stack estesa a frontend React (`pages`, `components`, `context`), backend Go e migrazioni SQL; protocollo Contract-First (allineamento preventivo DTO TypeScript e Go struct prima del codice di business); verifica di backward & rollback (`down.sql`) per le migrazioni DB; DoD standard a 4 punti per ogni task; istituzione del Tech Debt & Refactoring Backlog formale in `3_TASK_LIST.md`; nuovi workflow operativi `.agent/workflows/` (`review.md`, `db-migrate.md`, `release.md`).
- **ADR-09 (2026-10-11):** Roadmap Strategica Milestones 3, 4, 5: Approvazione trittico evolutivo: Milestone 3 (Active Defense & SOAR Light: Chokepoint Auto-Remediation, Virtual Patching, Alerting Hub), Milestone 4 (Continuous Compliance: NIS2, DORA, ISO 27001, Scorecard, Dossier PDF) e Milestone 5 (AI Remediation Copilot & Assistant).
- **PO-04 (2026-10-11):** Approvazione priorità strategiche e inclusione formale nella roadmap attiva di Opzione 2 (Active Defense), Opzione 3 (Compliance Frameworks) e Opzione 5 (AI Remediation Copilot).
- **ADR-08 (2026-10-11):** Architettura di Scope Aziendale Gerarchico e `CompanyContext` globale React: distinzione formale tra nodi radice (`parent_id IS NULL` = Aziende clienti) e nodi discendenti (`parent_id != NULL` = Reparti aziendali); persistenza in `localStorage`; propagazione trasparente e unificata del filtro target IP su tutte le feature (Dashboard, Targets, Scans, Findings, Reports, AttackPath) con eliminazione dei selettori ridondanti.
- **ADR-07 (2026-10-10):** Integrazione serializzazione `Invasiveness` in `ProfileResponse` API e seeding profilo `All-in-One 360° Comprehensive Audit`.
- **ADR-06 (2026-10-10):** Attack Path Probabilistic Risk Formula $1 - \prod(1 - p_i)$, bound 50 cammini, archi inferiti tratteggiati con confidenza, e Scan Diff `GET /api/v1/scans/{id}/diff` con stati NEW/RESOLVED/PERSISTENT.
- **ADR-05 (2026-10-10):** Profilo specialistico UDP, tassonomia di invasività con gate UI, modello finding arricchito con confidence/detection_method/cpe, hardening container capabilities (solo `CAP_NET_RAW`), migrazione PDF a `go-pdf/fpdf v0.9.0`.
- **ADR-04 (2026-10-10):** Stream-Ticket monouso memoria thread-safe (TTL 30s) per SSE e WebSocket; delimitatore POSIX `--` in Nmap; deny-list target interni e Docker bridge con anti-rebinding DNS; nohup watchdog remoto firewall rollback.
- **PO-01 (2026-10-09):** Approvazione implementazione pipeline CI/CD GitHub Actions per backend e frontend.
- **PO-02 (2026-10-09):** Approvazione ottimizzazione bundle frontend tramite code-splitting e `React.lazy`.

## Prossimi passi

1. Avvio operativo della **Milestone 3 (Epic 28 · Active Defense & SOAR Light)**.
2. Predisposizione dell'Implementation Plan per il task **E28-S1-T1** (Backend Chokepoint Auto-Remediation Service & SSH Firewall Rule Applicator) e creazione del file di dominio `docs/domains/active_defense_and_soar.md`.
3. Sviluppo, test e verifica zero-assumption del primo incremento.

## Log sessioni (ultime 5, la più recente in alto)

- 2026-10-15 17:05 · Evoluzione dell'Agente e della Governance di Progetto allo Standard "Software House Enterprise" (v6.2): aggiornato [AGENTS.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/AGENTS.md) con Context Layering su 4 livelli gerarchici per abbattere i consumi di token e prevenire degradazioni dell'attenzione; estesa la Mappa Documenti in [2_ARCHITECTURE.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/2_ARCHITECTURE.md) a una copertura Full-Stack bidirezionale (frontend, backend e migrazioni DB); formalizzato ADR-10; istituito il `Tech Debt & Refactoring Backlog` in [3_TASK_LIST.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/3_TASK_LIST.md) con DoD standard a 4 punti; potenziato `feature.md` con Contract-First e hostile review; creati i workflow operativi [.agent/workflows/review.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/.agent/workflows/review.md), [.agent/workflows/db-migrate.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/.agent/workflows/db-migrate.md) e [.agent/workflows/release.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/.agent/workflows/release.md)
- 2026-10-15 16:47 · Implementazione e verifica zero-assumption completata di Epic 27.1 (Semplificazione Centro Scansioni & Scanning Studio 360°): riprogettata l'interfaccia unificata di `NewScanModal.tsx` introducendo i 4 Master Presets intelligenti a 1-clic (Audit 360° Sicurezza Generale, Ricognizione Rapida Fast Sweep, Audit Vulnerabilità & CVE Exploit, Perimetro Web App & API Surface) e il pulsante per passare con 1 clic allo Scanning Studio 360° con i parametri precaricati; integrato il pannello Fine-Tuning 360° con controlli granulari (scope porte, timing template T1-T5, sonde attive, script vuln NSE, probe UDP critici -sU e salvataggio custom profiles); implementato il calcolo dinamico e real-time dell'invasività con Live Invasiveness Meter e blocco tassativo dell'avvio finché non viene fornita la spunta sul Gate di Consenso Obbligatorio; eseguiti 14/14 test Vitest PASS su NewScanModal, 18/18 file e 107/107 test frontend PASS al 100%, compilazione di produzione TypeScript / Vite `npm run build` PASS (0 errori, 10.64s), backend Go test 14/14 package PASS (0 errori)
- 2026-10-15 16:25 · Intake & Design per Epic 27.1 (Semplificazione Centro Scansioni & Scanning Studio 360°): chiarito il concetto del Gate di Consenso Obbligatorio; eseguita analisi tecnica di cybersecurity sui 14 profili attuali evidenziando sovraccarico cognitivo e ridondanze; formulata proposta architetturale di unificazione a 2 livelli (3-4 Master Presets + Pannello Fine-Tuning 360°); creati i relativi task E27.1-S1-T1..T3 in [3_TASK_LIST.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/3_TASK_LIST.md)
- 2026-10-13 19:47 · Integrazione e approfondimento esaustivo di Feature 6 (Centro Scansioni di Sicurezza) in [GUIDA_FUNZIONALE_PIATTAFORMA.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/GUIDA_FUNZIONALE_PIATTAFORMA.md): aggiunta spiegazione dettagliata dell'architettura di esecuzione Nmap (assenza shell, delimitatore POSIX `--`, zero scritture su disco `-oX -`, deny-list e anti-rebinding, isolamento container solo `CAP_NET_RAW`), ciclo di vita dell'Orchestratore (worker pool a 2 Goroutine, 5 fasi SSE in tempo reale, filtro Alive-Only, annullamento graceful con SIGTERM/SIGKILL), vocabolario di tutti i flag Nmap ed enciclopedia completa di tutti i 14 profili operativi implementati
- 2026-10-13 19:35 · Redazione completata del manuale funzionale esaustivo per il Product Owner: creato [GUIDA_FUNZIONALE_PIATTAFORMA.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/GUIDA_FUNZIONALE_PIATTAFORMA.md) e relativo artefatto; acquisiti ed incorporati 20 screenshot visuali in alta definizione dell'applicazione attiva tramite Google Chrome headless; descritte nel dettaglio tutte le 13 macro-features implementate (Login, Dashboard, Companies, Targets/TreeExplorer, LAN Discovery, Scans, Scan Detail/Diff, Findings, Attack Path, SSH/Firewall, Reports Hub, Settings, Debug Console) con spiegazione semplice non tecnica e matrice sinottica di ogni singolo tasto e dell'operazione corrispondente nel backend
- 2026-10-11 22:33 · Brainstorming strategico concluso con il Product Owner: approvato e formalizzato il piano evolutivo per le prossime tre Milestone di prodotto: Milestone 3 (Active Defense & SOAR Light), Milestone 4 (Continuous Compliance NIS2/DORA/ISO 27001) e Milestone 5 (AI Remediation Copilot). Aggiornati `0_REQUIREMENTS.md` (FR-19, FR-20, FR-21, UC-18..20, NFR-SEC-08), `1_PROJECT_CONTEXT.md` (in-scope roadmap), `2_ARCHITECTURE.md` (Mappa documenti completa, nuovi contratti API, ADR-09) e `3_TASK_LIST.md` (Epic 28 con task E28-S1-T1..T6, e code per Epiche 29 e 30)
- 2026-10-11 15:32 · Implementazione e verifica completa di Epic 27 (Gestione Aziende, Reparti & Filtro Globale): creato `CompanyContext` React globale con supporto gerarchico (Aziende radice e Reparti figli), derivazione ricorsiva dei target IP e persistenza in `localStorage`; integrato selettore aziendale universale nella barra laterale (desktop e drawer mobile); creata dashboard di gestione aziende `/companies` con card metriche, color tag e modale CRUD; contestualizzata `/targets` con visualizzazione dei soli reparti dell'azienda attiva; propagato il filtro automatico su Dashboard, Scansioni, Nuova Scansione, Vulnerabilità CVE, Report e Attack Path; rimosse le ridondanze grafiche e i selettori duplicati; verifica zero-assumption completata: Frontend Vitest 18/18 file PASS (105/105 test), compilazione produzione TypeScript / Vite `npm run build` PASS (0 errori, built in 7.39s), Backend Go test 14/14 package PASS
- 2026-10-10 18:14 · Verifica visuale e funzionale E2E automatizzata con Google Chrome headless via Puppeteer: testate con successo Dashboard, Attack Path Graph (`/attack-path`), Scansioni (`/scans`), Modale Nuova Scansione (`NewScanModal`), Reports (`/reports`) e Live Debug Console drawer (`>_`); rilevata e corretta l'omissione del campo `invasiveness` in `ProfileResponse` (`scan_handler.go`) e inserito profilo `All-in-One 360° Comprehensive Audit` nel DB; convalidate le visualizzazioni dei badge `SAFE`/`INTRUSIVE`, del profilo `Specialized UDP Critical Services` e del gate di autorizzazione obbligatoria con checkbox; catturati screenshot visuali conformi in artifacts · Backend 14/14 pkg PASS, Frontend 16 file / 99 test PASS al 100%
- 2026-10-10 17:05 · Correzione anomalia HTTP 500 su target interni protetti: rimappati `domain.ErrTargetInternalRestricted` e `domain.ErrTargetScopeTooBroad` su HTTP 400 Bad Request espliciti (`TARGET_RESTRICTED` e `SCOPE_TOO_BROAD`) in `discovery_handler.go`, `target_handler.go` e `scan_handler.go`; aggiunto controllo preventivo `CheckTargetScope` in `target_service.go` e `scan_service.go`; introdotto bypass configurabile via env `ALLOW_INTERNAL_TARGETS=true` / `DEV_ALLOW_LOCALHOST=true` per consentire test locali di sviluppo in piena sicurezza · Test suite backend (14 pkg) e frontend (97 test) 100% VERDI
- 2026-10-10 16:50 · Implementazione e verifica completa di Epic 24 (SEC-P0 Hardening Critico): sviluppato `StreamTicketStore` monouso per SSE e WebSocket con TTL 30s ed endpoint protetto `/api/v1/auth/stream-ticket`; implementato `CheckTargetScope` con deny-list IP/CIDR/hostname (loopback, RFC1918 docker bridge, link-local, cloud metadata, .internal) e risoluzione DNS anti-rebinding; inserito delimitatore POSIX `--` prima del target in `BuildSafeArguments` Nmap; disabilitato per default script esterno `vulners`; aggiunta regex di redazione query parameters sensibili (`ticket=`, `token=`, `passphrase=`) nella Debug Console e nell'esportatore Markdown AI; aggiunto watchdog di rollback remoto distaccato via `nohup` e `sudo iptables-restore` per firewall; risolto disallineamento cleanup async in `WebTerminalModal.tsx` · Test backend 14/14 pkg PASS (0 errori, build/vet OK), Frontend 16 file / 97 test PASS, npm run build OK (0 errori)

