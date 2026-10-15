# PROJECT STATE

> Memoria viva dell'agente. Letta a inizio sessione, aggiornata a fine di ogni turno.
> Non cancellare sezioni: se vuote, scrivi "nessuno".

**Ultimo aggiornamento:** 2026-10-15
**Modalità progetto:** esistente (onboarded formalmente sul branch chore/agent-onboarding)
**Fase:** Pulizia degli artefatti agentici obsoleti completata e in attesa di accettazione; governance Codex v7.1 approvata.

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
Branch principale:       main (allineamento remoto non verificato durante la migrazione documentale)
Comandi vietati:         rm -rf fuori repo, git reset --hard non confermati, git push --force, DROP database senza conferma esplicita
```

## Baseline test (prima di ogni intervento)

Verifica pre-push del 2026-10-15: `go test ./...`, `go build ./...`, `go vet ./...` PASS; frontend `npm.cmd test` PASS (18 file, 107 test), `npm.cmd run build` PASS (12,41 s). Verifica documentale PASS (24 documenti, 53 Markdown, 101 link/ancore). Controllo whitespace esteso al diff da `origin/main` PASS dopo normalizzazione. Comandi applicativi eseguiti con permessi estesi dopo blocchi locali della sandbox; nessuna modifica di logica applicativa introdotta nella preparazione.


| Data | Comando | Esito | Fallimenti preesistenti noti |
|---|---|---|---|
| 2026-10-15 17:13 | `go vet ./...` (backend) | PASS (exit code 0) | Nessuno. 0 warning, static analysis perfetta. |
| 2026-10-15 17:12 | `go test ./...` (backend) | PASS (exit code 0, 14 package) | Nessuno. Tutti i 14 package backend passano al 100%. |
| 2026-10-15 17:12 | `go build ./...` (backend) | PASS (exit code 0) | Nessuno. Compilazione binari server e cve-sync completata con successo. |
| 2026-10-15 17:13 | `npm run build` (frontend tsc + Vite) | PASS (exit code 0, 7.25s) | Nessuno. Compilazione TypeScript e bundle Vite ottimizzato, 0 errori. |
| 2026-10-15 17:12 | `npm test` (frontend Vitest) | PASS (exit code 0, 18 file, 107 test) | Nessuno. Tutti i 107 test passano al 100% con 0 fallimenti. |

## In corso

- (nessuno)

## Accettati recentemente dal PO (`[x]`)

- **AGENT-SDLC-02 · Struttura documentale e orchestrazione per feature:**
  - Validato e accettato dal PO in data 2026-10-15.
  - Documentazione canonica organizzata in `01-product`–`06-devops`, changelog in root, sette ruoli specializzati e gate proporzionati alla singola feature.
  - Confermata l'assunzione operativa: ruoli attivati quando pertinenti, validazione tecnica autonoma del Tech Lead e accettazione funzionale riservata al PO.
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

- **AGENT-CLEAN-03 · Pulizia repository:** rimossi cinque stub documentali, lo script screenshot Antigravity, la guida funzionale destinata agli utenti e 20 screenshot (3.250.486 byte). Eliminate le relative directory vuote. In `docs/` restano indice, memoria, requisiti, design, architettura, domini, sviluppo, QA e DevOps necessari agli agenti.
  - Indicazione PO: `docs/` riservata alla documentazione agentica. Evidenze visuali temporanee in `.agent/artifacts/<task-id>/`, ignorate da Git; workflow e verificatore aggiornati per rispettare il vincolo.
  - Evidenze: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify-agent-docs.ps1` PASS (exit 0: struttura agentica, 24 documenti obbligatori, 53 Markdown, 104 link/ancore, registry e fence); `git diff --check` PASS. Self-review dei riferimenti completata.
  - Recuperabilità: i 27 file rimossi durante la pulizia sono tracciati da Git e recuperabili dalla cronologia; le directory vuote non contenevano dati.

- **AGENT-CODEX-01 · Migrazione governance Antigravity → Codex:**
  - `AGENTS.md` riscritto come entry point nativo Codex con bootstrap, gerarchia delle istruzioni, tutela della working tree, permessi, fallback capability-aware e quality gate evidence-first.
  - Aggiunto `.agent/README.md` come registry; portati i workflow esistenti e creati i playbook mancanti per `init`, `fix`, `refactor`, `verify`, `sync-docs`, `audit`, `status` e `commit`.
  - Formalizzato ADR-11 e allineati contesto, task list e memoria viva. Nessuna modifica al codice applicativo.

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

- Struttura documentale SDLC v7.1 e orchestrazione per feature approvate dal Product Owner; la precedente migrazione base Codex AGENT-CODEX-01 resta formalmente in attesa.
- `docs/` contiene solo documentazione necessaria agli agenti; manuali utente, screenshot, allegati binari e archivi sono esclusi, come richiesto dal PO.
- Storico consolidato nel [CHANGELOG.md](../CHANGELOG.md); stub e cartella legacy rimossi dopo la migrazione alle fonti canoniche.
- Pipeline CI/CD GitHub Actions e bundle splitting implementati.
- Milestone 2 completata sia su Backend Go che su Frontend React.
- Roadmap di hardening sicurezza (Epiche SEC-P0, SEC-P1, SEC-P2) implementata al 100%.

## Decisioni recenti (ultime 10, dettagli negli ADR)

- **ADR-13 (2026-10-15):** Documentazione `docs/` riservata agli agenti; rimossi manuale e screenshot, evidenze temporanee fuori da `docs/`, controllo strutturale automatizzato.
- **ADR-12 (2026-10-15):** Documentazione per fasi SDLC, ruoli con input/output e handoff, gate per feature, fonti canoniche distinte e migrazione senza duplicare i documenti storici. Dettagli in `docs/03-architecture/ADR/README.md`.

- **ADR-11 (2026-10-15):** `AGENTS.md` è l'entry point nativo Codex; `.agent/workflows/` è una libreria di playbook indicizzata, non un registro di slash command. Introdotti bootstrap evidence-first, tutela delle modifiche preesistenti, fallback per browser/subagent e permessi espliciti per commit, push, merge, tag e deploy.
- **ADR-10 (2026-10-15):** Standard Operativo Software House: Context Layering su 4 livelli gerarchici per contenimento token bloat e azzeramento allucinazioni; Mappa Documenti Full-Stack estesa a frontend React (`pages`, `components`, `context`), backend Go e migrazioni SQL; protocollo Contract-First (allineamento preventivo DTO TypeScript e Go struct prima del codice di business); verifica di backward & rollback (`down.sql`) per le migrazioni DB; DoD standard a 4 punti per ogni task; istituzione del Tech Debt & Refactoring Backlog formale in `3_TASK_LIST.md`; nuovi workflow operativi `.agent/workflows/` (`review.md`, `db-migrate.md`, `release.md`).
- **ADR-09 (2026-10-11):** Roadmap Strategica Milestones 3, 4, 5: Approvazione trittico evolutivo: Milestone 3 (Active Defense & SOAR Light: Chokepoint Auto-Remediation, Virtual Patching, Alerting Hub), Milestone 4 (Continuous Compliance: NIS2, DORA, ISO 27001, Scorecard, Dossier PDF) e Milestone 5 (AI Remediation Copilot & Assistant).
- **PO-04 (2026-10-11):** Approvazione priorità strategiche e inclusione formale nella roadmap attiva di Opzione 2 (Active Defense), Opzione 3 (Compliance Frameworks) e Opzione 5 (AI Remediation Copilot).
- **ADR-08 (2026-10-11):** Architettura di Scope Aziendale Gerarchico e `CompanyContext` globale React: distinzione formale tra nodi radice (`parent_id IS NULL` = Aziende clienti) e nodi discendenti (`parent_id != NULL` = Reparti aziendali); persistenza in `localStorage`; propagazione trasparente e unificata del filtro target IP su tutte le feature (Dashboard, Targets, Scans, Findings, Reports, AttackPath) con eliminazione dei selettori ridondanti.
- **ADR-07 (2026-10-10):** Integrazione serializzazione `Invasiveness` in `ProfileResponse` API e seeding profilo `All-in-One 360° Comprehensive Audit`.
- **ADR-06 (2026-10-10):** Attack Path Probabilistic Risk Formula $1 - \prod(1 - p_i)$, bound 50 cammini, archi inferiti tratteggiati con confidenza, e Scan Diff `GET /api/v1/scans/{id}/diff` con stati NEW/RESOLVED/PERSISTENT.
- **ADR-05 (2026-10-10):** Profilo specialistico UDP, tassonomia di invasività con gate UI, modello finding arricchito con confidence/detection_method/cpe, hardening container capabilities (solo `CAP_NET_RAW`), migrazione PDF a `go-pdf/fpdf v0.9.0`.

## Contesto operativo della migrazione

- Le baseline applicative sopra sono storiche, non rieseguite per la modifica a documenti e verificatore documentale; nessun codice applicativo è stato modificato.
- Working tree già modificata dalla migrazione Codex precedente; le eliminazioni legacy preesistenti sono state mantenute e ora risultano coerenti con la pulizia delle vecchie strutture.
- Documenti canonici e indice: `docs/README.md`; task attivi: `docs/04-development/TASK_LIST.md`. Nessun commit, push o deploy eseguito.

## Prossimi passi

1. Avvio operativo della **Milestone 3 (Epic 28 · Active Defense & SOAR Light)**.
2. Predisposizione dell'Implementation Plan per il task **E28-S1-T1** (Backend Chokepoint Auto-Remediation Service & SSH Firewall Rule Applicator) e creazione del file di dominio `docs/domains/active_defense_and_soar.md`.
3. Sviluppo, test e verifica zero-assumption del primo incremento.

## Log sessioni (ultime 5, la più recente in alto)

- 2026-10-15 · Il PO richiede pubblicazione di tutte le modifiche su `main`: verificati branch/remoto e relazione fast-forward, suite applicative e documentali PASS; preparato commit di normalizzazione whitespace prima dell'integrazione e push. Nessun tag o rilascio applicativo richiesto; accettazioni funzionali pendenti mantenute.
- 2026-10-15 · AGENT-CLEAN-03 esteso su richiesta PO: rimossi anche manuale funzionale e 20 screenshot; `docs/` riservata ai documenti agentici. Aggiornati percorsi delle evidenze temporanee, task, workflow e verifica automatica della struttura. In attesa di accettazione PO.
- 2026-10-15 · AGENT-SDLC-02 accettato dal PO: task promosso a `[x]`, assunzione sui ruoli confermata, voce spostata dalle attese agli accettati e registrata nel changelog. Nessuna modifica applicativa.
- 2026-10-15 · Migrazione agentica Antigravity → Codex (AGENT-CODEX-01): riscritto `AGENTS.md` come entry point Codex; introdotto `.agent/README.md`; aggiunti otto playbook mancanti; adattati feature, feedback, onboarding, review, release e migrazioni alle capability e ai permessi reali; formalizzato ADR-11 e completato il Doc Sync della governance. Verifica documentale eseguita senza modificare codice applicativo.
- 2026-10-15 17:15 · Esecuzione Onboarding Formale (/onboard): creata snapshot sicura pre-onboard e branch `chore/agent-onboarding`; eseguita ricognizione completa inventario repository, mapping contratti API e verifica baseline con zero assunzioni: `go vet` PASS (0 warning), `go test ./...` PASS (14/14 package), `go build ./...` PASS, `vitest run` PASS (18/18 file, 107/107 test), `npm run build` PASS (0 errori, 7.25s); confermato 100% allineamento documentale full-stack e assenza di TODO/FIXME o segreti esposti nel codice
