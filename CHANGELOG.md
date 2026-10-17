# 4. Archivio Task Completati & Changelog (Changelog & Archive)

> **Progetto:** NextGen CyberSecurity Platform  
> **Autore / Ruolo:** TechLead-Engineer-Agent (System Architect & Tech Lead)  
> **Archiviazione Eseguita:** 2026-10-08  
> **Riferimento Regole:** [AGENTS.md](AGENTS.md) · il precedente `_legacy/Agent.md` è stato rimosso dalla working tree durante l'onboarding.

---

## Accettazione e pubblicazione su main · 2026-10-17

Il PO approva esplicitamente tutte le modifiche presenti e richiede commit/push su `origin/main`, senza nuovi branch. Accettati TEST-OPT-01, COMPANY-UX-02, COMPANY-PDF-03 e AP-REVIEW-01, incluse le assunzioni documentate. Test automatici, build/vet e integrazioni PostgreSQL di test PASS come dettagliato nei rispettivi report QA. Verifiche visuali browser/PDF non eseguite dall'agente: limite dichiarato e accettato dal PO, senza convertirlo in un PASS tecnico. E28-S1-T1 rimane incompleto per prova Linux e hardening TECH-DEBT-06/07; approvazione della pubblicazione non equivale al completamento del lavoro pianificato. Nessun deploy autorizzato o eseguito.

## Archivio task accettati e proposta superata · 2026-10-16

AGENT-CLEAN-03 e AGENT-CODEX-01 confermati e approvati esplicitamente dal PO il 2026-10-16. Conservati integralmente ID, dipendenze, DoD ed evidenze dei 12 task accettati. I tre task Epic 27.1 sono archiviati come proposta superata dalla versione scanner già accettata, senza mantenere false richieste di approvazione.

### Revisione scanner Nmap (SCAN-UX-01)

- [x] SCAN-COPY-01 · Descrizioni dei profili focalizzate sullo scopo
  - File: `frontend/src/components/scans/scanConfiguration.ts`, `frontend/src/components/NewScanModal.tsx`, test modale
  - Dipende da: SCAN-UX-01 (implementazione presente)
  - Accettazione PO: 2026-10-16, OK riferito alla revisione dei testi.
  - Test: regressione frontend e build; verifica documentale e whitespace
  - Docs: dominio scanner, specifiche UI, PROJECT_STATE
  - DoD: descrizioni tecniche dirette, nessuna variazione ai comandi; controlli PASS e self-review; limite visuale dichiarato.
  - Evidenze: `npm.cmd test` PASS (21 file, 145 test); `npm.cmd run build` PASS (17,06 s); verificatore documentale PASS, exit 0. Copy verificato nel DOM e self-review; nessuna nuova verifica visuale di layout per questa modifica testuale. Restano separati i gate visuali SCAN-UX-01/SCAN-TIME-01.

- [x] SCAN-TIME-01 · Stime per profilo, barra totale ed ETA
  - File: runner/progress scanner e orchestratore Go, payload SSE Go/TypeScript; preset, modale, dettaglio/elenco scansioni e test
  - Dipende da: SCAN-UX-01 (implementazione presente)
  - Test: telemetria XML/SSE, stime/monotonicità/stale, gruppi e stati terminali; suite complete, build/vet
  - Docs: `docs/domains/scanning_and_discovery.md`, `docs/04-development/SCAN_PROGRESS_PLAN.md`
  - DoD: test mirati PASS; integrazione/UI PASS o limite dichiarato; build/typecheck/vet PASS; Doc Sync e self-review.
  - Evidenze: frontend 21 file / 145 test e build PASS; backend suite/build/vet PASS, exit 0; doc verifier PASS. Il PO conferma test delle scansioni riusciti e accetta la funzionalità il 2026-10-16. Prova funzionale eseguita dal PO; nessuna prova visuale autonoma o benchmark rete dichiarato dall'agente.

- [x] SCAN-UX-01 · Semplificazione scanner, profili veritieri ed editor Nmap
  - File: `frontend/src/components/NewScanModal.tsx`, `frontend/src/components/scans/scanConfiguration.ts`, relativi test; `backend/internal/scanner/profiles.go`, `backend/internal/service/scan_service.go`, relativi test
  - Dipende da: nessuno; sostituisce la UX proposta in E27.1 su richiesta del PO
  - Test: parser happy/error/edge, payload coerente con preview, scope/blacklist, gate, scheduler; suite Go/frontend, build, vet, verifica documentale e UI locale
  - Docs: `docs/domains/scanning_and_discovery.md`, `docs/04-development/NMAP_SCAN_REVIEW.md`
  - DoD: (1) test mirati PASS; (2) integrazione/UI PASS o limite dichiarato; (3) suite/build/vet PASS; (4) Doc Sync e self-review completati.
  - Accettazione PO: 2026-10-16, test delle scansioni dichiarati correttamente funzionanti. Il PO fornisce la verifica funzionale mancante; nessuno screenshot o controllo visuale autonomo inventato. Il debito TECH-DEBT-05 resta aperto.

### Pulizia repository · Artefatti obsoleti (AGENT-CLEAN-03)

- [x] AGENT-CLEAN-03 · Pulizia artefatti obsoleti e documentazione riservata agli agenti
  - File: stub rimossi, vecchio script Antigravity, directory vuote, manuale utente e screenshot rimossi da `docs/`; regole agentiche e verificatore aggiornati
  - Dipende da: AGENT-SDLC-02
  - Test: PASS struttura `docs/` agentica, 24 documenti obbligatori, 53 Markdown, 104 link/ancore e registry; `git diff --check` exit 0.
  - Docs: `docs/README.md`, `docs/PROJECT_STATE.md`
  - DoD: fonti canoniche intatte; rimozione manuale e screenshot come richiesto dal PO; riferimenti operativi aggiornati; verifica documentale PASS; Doc Sync e self-review completati.

### Migrazione SDLC · Orchestrazione per feature (AGENT-SDLC-02)

- [x] AGENT-SDLC-02 · Organizzazione documentale per fasi e responsabilità specializzate
  - File: `AGENTS.md`, `.agent/roles/`, `.agent/workflows/`, `docs/01-product/`–`docs/06-devops/`, `README.md`, `CHANGELOG.md`, `scripts/verify-agent-docs.ps1`
  - Dipende da: AGENT-CODEX-01 (integrazione della governance, accettato dal PO il 2026-10-16)
  - Test: PASS 24 documenti obbligatori, 59 Markdown e 153 link/ancore; 7 ruoli e 15 playbook indicizzati; preservate 40 sezioni FR/UC, ERD/API/ADR, task e changelog; diff check exit 0.
  - Docs: `docs/README.md`, `docs/PROJECT_STATE.md`, `docs/03-architecture/ADR/README.md`
  - DoD: (1) contenuti migrati preservati; (2) gate e handoff coerenti; (3) controlli Markdown PASS, suite applicativa N/A; (4) Doc Sync e self-review completati.

### Chore Agentico · Migrazione Antigravity → Codex (AGENT-CODEX-01)

- [x] AGENT-CODEX-01 · Adattamento completo della governance agentica al runtime Codex
  - File: `AGENTS.md`, `.agent/README.md`, `.agent/workflows/*.md`, `docs/PROJECT_STATE.md`, `docs/03-architecture/HLD.md`
  - Dipende da: ADR-10
  - Test: inventario playbook, ricerca riferimenti Antigravity/tool obsoleti, validazione link locali, `git diff --check`
  - Docs: `docs/01-product/PRD.md`, `docs/03-architecture/HLD.md`, `docs/04-development/TASK_LIST.md`, `docs/PROJECT_STATE.md`
  - DoD: entry point Codex nativo; tutti gli alias documentati hanno un playbook; permessi e fallback espliciti; verifiche Markdown PASS; Doc Sync e hostile review completati.

### Epic 27: Gestione Aziende, Reparti & Filtro Globale (FEAT-COMPANY-SCOPE) [ACCETTATO DAL PO]

- [x] E27-S1-T1 · Creazione CompanyContext globale con gestione stato aziendale, derivazione gerarchica reparti e filtri automatici
  - File: frontend/src/context/CompanyContext.tsx, frontend/src/context/CompanyContext.test.tsx
  - Dipende da: nessuno
  - Test: test unitari Vitest su CompanyContext (selezione azienda, reparti discendenti, filtro target/scans/findings, localStorage)
  - Docs: docs/domains/reporting_and_assets.md
  - DoD: Context operativo con memorizzazione persistente dello scope aziendale e helper di filtraggio per tutte le entità.

- [x] E27-S1-T2 · Integrazione Selettore Aziendale Globale nel Menu Laterale (Sidebar / AppLayout)
  - File: frontend/src/components/layout/AppLayout.tsx, frontend/src/App.tsx
  - Dipende da: E27-S1-T1
  - Test: test rendering AppLayout e interazione selettore azienda
  - Docs: docs/domains/reporting_and_assets.md
  - DoD: Dropdown azienda visibile in alto nella sidebar, link a gestione aziende, voce di menu "Gestione Aziende" (/companies).

- [x] E27-S1-T3 · Dashboard Dedicata per la Gestione delle Aziende (/companies)
  - File: frontend/src/pages/Companies.tsx, frontend/src/pages/Companies.test.tsx, frontend/src/components/CompanyModal.tsx
  - Dipende da: E27-S1-T1
  - Test: test Vitest per Companies.tsx (creazione azienda, elenco card con metriche reparti/IP/rischio, modifica, cancellazione, selezione scope)
  - Docs: docs/domains/reporting_and_assets.md
  - DoD: Vista completa di gestione aziende con KPI aggregati, creazione con color tag, attivazione rapida scope.

- [x] E27-S1-T4 · Contestualizzazione Pagina Target & Reparti (/targets e TreeExplorer)
  - File: frontend/src/pages/Targets.tsx, frontend/src/components/TreeExplorer.tsx
  - Dipende da: E27-S1-T1, E27-S1-T3
  - Test: test rendering Targets con azienda attiva e creazione reparti
  - Docs: docs/domains/scanning_and_discovery.md
  - DoD: /targets mostra i reparti dell'azienda selezionata, creazione di nuovi reparti associati all'azienda e gestione IP per reparto.

- [x] E27-S1-T5 · Propagazione Filtro Globale ed Eliminazione Ridondanze su tutte le Feature
  - File: frontend/src/pages/Dashboard.tsx, frontend/src/pages/Scans.tsx, frontend/src/pages/Findings.tsx, frontend/src/pages/Reports.tsx, frontend/src/pages/AttackPath.tsx, frontend/src/components/NewScanModal.tsx
  - Dipende da: E27-S1-T1, E27-S1-T4
  - Test: test regressione su tutte le viste con scope aziendale attivo
  - Docs: docs/domains/scanning_and_discovery.md, docs/domains/vulnerability_intelligence.md
  - DoD: Tutte le viste riflettono automaticamente lo scope dell'azienda selezionata senza dropdown duplicati; NewScanModal propone solo target dell'azienda.

- [x] E27-S1-T6 · Verifica Zero-Assumption e Suite Completa di Test
  - File: frontend/src/tests/**, backend/internal/**
  - Dipende da: E27-S1-T1..E27-S1-T5
  - Test: npm test (Vitest: 18/18 PASS, 105/105 PASS), npm run build (0 errori), go test ./... (tutti i package PASS)
  - Docs: docs/PROJECT_STATE.md, docs/03-architecture/HLD.md
  - DoD: Build pulita (0 errori), tutti i test frontend e backend 100% VERDI.

### Proposta storica superata · Epic 27.1: Semplificazione Centro Scansioni & Scanning Studio 360° (FEAT-SCAN-STUDIO)

- **Superato:** E27.1-S1-T1 · Riprogettazione UI di NewScanModal: 3-4 Master Presets Intelligenti e Pannello "Fine-Tuning a 360°"
  - File: `frontend/src/components/NewScanModal.tsx`, `frontend/src/components/NewScanModal.test.tsx`
  - Dipende da: E27-S1-T5
  - Test: test unitari Vitest su NewScanModal (selezione preset standard, switch a personalizzazione avanzata porte/sonde/timing, live preview reattiva, calcolo dinamico invasività e trigger gate) (14/14 PASS)
  - Docs: `docs/domains/scanning_and_discovery.md`, `docs/02-design/UI_SPECIFICATIONS.md`
  - DoD: Interfaccia unificata senza 14 opzioni disperse: card a 1-clic per i casi d'uso standard + controlli granulari (porte, OS, versioni, vuln script, UDP, rate/timing) per libertà completa a 360°.

- **Superato:** E27.1-S1-T2 · Allineamento Contratti e Documentazione Funzionale
  - File: `docs/02-design/UI_SPECIFICATIONS.md`, `docs/domains/scanning_and_discovery.md`, `docs/03-architecture/HLD.md`
  - Dipende da: E27.1-S1-T1
  - Test: verifica consistenza documentale e mappa dei profili
  - Docs: `docs/PROJECT_STATE.md`, `docs/02-design/UI_SPECIFICATIONS.md`
  - DoD: Documentazione agentica allineata con spiegazione della nuova UX unificata; eventuali screenshot di verifica fuori da `docs/`, secondo la richiesta PO di pulizia.

- **Superato:** E27.1-S1-T3 · Verifica Zero-Assumption e Suite Completa di Regressione
  - File: `frontend/src/**`, `backend/**`
  - Dipende da: E27.1-S1-T1, E27.1-S1-T2
  - Test: `npm test` (18 file, 107 test PASS), `npm run build` (tsc + vite PASS in 10.64s), `go test ./...` (14 pkg PASS)
  - Docs: `docs/PROJECT_STATE.md`
  - DoD: 100% test verdi, build TypeScript pulita con 0 errori.

## Testi profili scanner · 2026-10-16

### Scanner Nmap e monitor del gruppo · Accettati dal PO

- [x] **SCAN-UX-01:** interfaccia Profili/Comando Nmap, otto preset con scopo tecnico, matching degli argomenti, validazione semantica e classificazione custom; corretto include_common_ports di Discovery.
- [x] **SCAN-TIME-01:** durate indicative per profilo, telemetria XML Nmap, snapshot API tenant e monitor lotto con avanzamento/ETA, coda ed esiti distinti, ripristino sessione.
- **Accettazione:** 2026-10-16; il PO dichiara di aver eseguito i test e conferma il funzionamento delle scansioni, autorizzando il push di tutte le modifiche. Debito TECH-DEBT-05 sul consenso API ancora aperto; stime non certificate da benchmark.

### Descrizioni dei profili

- [x] **SCAN-COPY-01:** accettata dal PO la revisione delle otto descrizioni, focalizzate su scopo e attività della scansione. Rimosso il suggerimento Discovery sotto il selettore. Parametri, tempi e comportamento invariati; nessuna accettazione implicita degli altri task scanner.
- Verifiche precedenti: frontend 145 test e build PASS, controlli documentali PASS (exit 0).

## Governance agentica v7.1 - 2026-10-15

### AGENT-SDLC-02 · Struttura documentale e orchestrazione per feature

- [x] Documentazione organizzata nelle fasi `01-product`–`06-devops`, con indice canonico e rimandi compatibili dai vecchi percorsi.
- [x] Definiti sette ruoli specializzati con input, output, limiti e handoff; il Tech Lead orchestra e il Product Owner mantiene le decisioni di prodotto.
- [x] Introdotti gate proporzionati alla singola feature, contract-first, fonti di verità distinte e diagrammi Mermaid quando utili.
- [x] Verificati 24 documenti richiesti, 59 file Markdown e 153 link/ancore; preservate 40 sezioni FR/UC, ERD, API, ADR, task e changelog storico.
- **Accettazione PO:** 2026-10-15.

## 🏆 MILESTONE 1 (MVP) - Task Completati con Successo

### EPIC 1: Project Scaffolding & Container Infrastructure
- [x] **TASK-01: Inizializzazione Struttura Repository e Modulo Go**
  - **File coinvolti:** `go.mod`, `go.sum`, `cmd/server/main.go`, `.gitignore`
  - **DoD:** Modulo Go `nextgen-cybersec` inizializzato con Go 1.22+, `main.go` minimo funzionante che risponde a un health check HTTP e `.gitignore` configurato.
- [x] **TASK-02: Setup Docker Compose Multi-Container (PostgreSQL + Backend + Nmap)**
  - **File coinvolti:** `docker-compose.yml`, `backend/Dockerfile`, `.env.example`
  - **DoD:** `docker-compose up` avvia container PostgreSQL 16 con healthcheck e container backend Go basato su immagine Linux con `nmap` preinstallato e permessi di rete configurati.
- [x] **TASK-03: Inizializzazione Progetto Frontend React SPA**
  - **File coinvolti:** `frontend/package.json`, `frontend/vite.config.ts`, `frontend/tailwind.config.js`, `frontend/Dockerfile`
  - **DoD:** Progetto React 18 + TypeScript + Vite + Tailwind CSS inizializzato, compilabile con `npm run build` e containerizzato per servire asset via Nginx.

### EPIC 2: Database Schema & Multi-Tenant Core
- [x] **TASK-04: Implementazione Migrazioni SQL e Script DDL**
  - **File coinvolti:** `backend/internal/repository/migrations/000001_init_schema.up.sql`, `000001_init_schema.down.sql`
  - **DoD:** Script SQL completo contenente tutte le tabelle (`tenants`, `users`, `targets`, `scan_profiles`, `scans`, `findings`, `cve_entries`, `schedules`), indici compositi e foreign key vincolate.
- [x] **TASK-05: Configurazione Connection Pool PostgreSQL con pgx/v5 e Migrator**
  - **File coinvolti:** `backend/internal/repository/postgres/db.go`, `backend/internal/config/config.go`
  - **DoD:** Singleton / Pool manager per PostgreSQL con retry all'avvio, lettura variabili d'ambiente (`.env`) ed esecuzione automatica delle migrazioni pending.
- [x] **TASK-06: Modelli di Dominio Core (Clean Architecture)**
  - **File coinvolti:** `backend/internal/domain/tenant.go`, `backend/internal/domain/user.go`, `backend/internal/domain/target.go`, `backend/internal/domain/scan.go`, `backend/internal/domain/finding.go`
  - **DoD:** Struct di dominio Go con tag JSON/DB, costanti di stato (`ScanStatus`, `SeverityLevel`), ed errori di business tipizzati.

### EPIC 3: Authentication & Multi-Tenant RBAC
- [x] **TASK-07: Servizio Autenticazione (Password Hashing & JWT Token)**
  - **File coinvolti:** `backend/pkg/auth/jwt.go`, `backend/pkg/auth/hasher.go`, `backend/internal/service/auth_service.go`, `backend/internal/repository/postgres/user_repo.go`
  - **DoD:** Hashing sicuro delle password (Bcrypt/Argon2id), generazione e validazione token JWT contenenti `user_id`, `tenant_id` e `role`. Unit test con copertura 100%.
- [x] **TASK-08: HTTP Middleware per Tenant Isolation e RBAC**
  - **File coinvolti:** `backend/internal/delivery/http/middleware/auth_middleware.go`, `backend/internal/delivery/http/middleware/tenant_middleware.go`
  - **DoD:** Middleware Chi che valida il Bearer token, inietta nel `context.Context` l'identificativo tenant e verifica i permessi di ruolo (es. solo `SecOps`/`Admin` possono avviare scansioni).
- [x] **TASK-09: Handler HTTP Auth e Seed Admin/Tenant di Default**
  - **File coinvolti:** `backend/internal/delivery/http/handler/auth_handler.go`, `backend/internal/repository/postgres/seed.go`
  - **DoD:** Endpoint `/api/v1/auth/login` e `/api/v1/auth/me` attivi. Script di seeding all'avvio che crea il tenant primario e l'utente SuperAdmin iniziale in assenza di record.

### EPIC 4: Target & Asset Scope Management
- [x] **TASK-10: Validatore e Parser Target (IP, CIDR, FQDN, Blacklist)**
  - **File coinvolti:** `backend/pkg/validator/target_validator.go`, `backend/pkg/validator/target_validator_test.go`
  - **DoD:** Funzioni pure in Go per verificare la validità sintattica di IPv4, IPv6, blocchi CIDR e domini, controllando che non rientrino nella lista nera di esclusione. Unit test completi.
- [x] **TASK-11: Repository e Service Target (Multi-Tenant)**
  - **File coinvolti:** `backend/internal/repository/postgres/target_repo.go`, `backend/internal/service/target_service.go`
  - **DoD:** CRUD dei target isolato rigorosamente per `tenant_id`.
- [x] **TASK-12: Handler HTTP Target**
  - **File coinvolti:** `backend/internal/delivery/http/handler/target_handler.go`
  - **DoD:** Endpoint REST `GET /api/v1/targets`, `POST /api/v1/targets`, `DELETE /api/v1/targets/{id}` protetti da autenticazione.

### EPIC 5: Nmap Scanner Engine & XML Streaming Parser
- [x] **TASK-13: Nmap Command Runner Sicuro con Context Cancellation**
  - **File coinvolti:** `backend/internal/scanner/nmap_runner.go`, `backend/internal/scanner/profiles.go`, `backend/internal/scanner/nmap_runner_test.go`
  - **DoD:** Esecuzione di `nmap` via `exec.CommandContext` (NO shell interpretation), supporto profili (Fast, Standard, Full), propagazione immediata del `SIGTERM` in caso di annullamento.
- [x] **TASK-14: Parser XML Nmap ad Alte Prestazioni**
  - **File coinvolti:** `backend/internal/scanner/nmap_parser.go`, `backend/internal/scanner/nmap_parser_test.go`
  - **DoD:** Parser basato su streaming `xml.Decoder` capace di convertire l'XML di Nmap in struct Go di host, porte, protocolli, service name, versioni e banner. Test con fixture XML reali.

### EPIC 6: Scan Orchestration & Background Worker Pool
- [x] **TASK-15: Worker Pool Asincrono e Gestore delle Code**
  - **File coinvolti:** `backend/internal/service/scan_orchestrator.go`, `backend/internal/repository/postgres/scan_repo.go`
  - **DoD:** Coda in-memory / concorrente con Goroutine worker pool, controllo del limite massimo di scansioni parallele e aggiornamento transazionale degli stati (`QUEUED` -> `RUNNING` -> `COMPLETED`/`FAILED`).
- [x] **TASK-16: Cancellazione Graceful delle Scansioni**
  - **File coinvolti:** `backend/internal/service/scan_service.go`, `backend/internal/delivery/http/handler/scan_handler.go`
  - **DoD:** Endpoint `POST /api/v1/scans/{id}/cancel` che individua il context della scansione in corso, lo cancella, termina il processo Nmap e imposta lo stato a `CANCELLED`.
- [x] **TASK-17: API REST Gestione Scansioni e Risultati**
  - **File coinvolti:** `backend/internal/delivery/http/handler/scan_handler.go`, `backend/internal/repository/postgres/finding_repo.go`, `backend/internal/service/scan_service.go`
  - **DoD:** Endpoint `POST /api/v1/scans`, `GET /api/v1/scans`, `GET /api/v1/scans/{id}`, `GET /api/v1/scans/{id}/findings`.

### EPIC 7: Offline CVE Database & Correlation Engine
- [x] **TASK-18: Schema e Repository Catalogo CVE Locale**
  - **File coinvolti:** `backend/internal/repository/postgres/cve_repo.go`, `backend/internal/domain/cve.go`
  - **DoD:** Query ottimizzate e indici per la ricerca rapida per CPE, vendor, prodotto e versione.
- [x] **TASK-19: Motore di Correlazione Banner/CPE -> CVE con CVSS Enricher**
  - **File coinvolti:** `backend/internal/service/cve_correlator.go`, `backend/internal/service/cve_correlator_test.go`
  - **DoD:** Logica di correlazione che associa i servizi estratti dal parser Nmap ai record CVE locali, calcolando severità e raccomandazioni di remediation. Unit test con casi noti (es. Apache 2.4.49 -> CVE-2021-41773).
- [x] **TASK-20: Tool di Sincronizzazione / Import Dati CVE**
  - **File coinvolti:** `backend/cmd/cve-sync/main.go`, `backend/internal/service/cve_sync_service.go`
  - **DoD:** Utility per importare feed CVE JSON in PostgreSQL, con dataset ridotto per test rapido e meccanismo di inserimento a batch.
- [x] **TASK-33: Endpoint REST per Statistiche e Sincronizzazione CVE**
  - **File coinvolti:** `backend/internal/delivery/http/handler/cve_handler.go`, `backend/internal/delivery/http/handler/cve_handler_test.go`, `backend/cmd/server/main.go`
  - **DoD:** Implementazione di `GET /api/v1/cve/stats` e `POST /api/v1/cve/sync` (protetto da RBAC Admin). Suite di unit test con copertura ed esito positivo.
- [x] **TASK-34: Bonifica Codice Orfano ReportTemplate e Allineamento Architettura**
  - **File coinvolti:** `backend/internal/domain/tenant.go`, `backend/internal/repository/migrations/000003_drop_report_templates.up.sql`, `backend/internal/repository/migrations/000003_drop_report_templates.down.sql`, `docs/2_ARCHITECTURE.md`
  - **DoD:** Rimozione della struct orfana `ReportTemplate`, migrazione DDL versionata di eliminazione tabella/indici inutilizzati `report_templates`, aggiornamento 2_ARCHITECTURE.md (ERD e stack PDF gofpdf).

### EPIC 8: Scheduler Service (Pianificazione Scansioni)
- [x] **TASK-21: Scheduler Cron in Background con robfig/cron**
  - **File coinvolti:** `backend/internal/service/scheduler_service.go`, `backend/internal/repository/postgres/schedule_repo.go`
  - **DoD:** Scheduler in background che carica le schedulazioni attive da DB all'avvio, valuta le espressioni cron e accoda automaticamente nuove scansioni allo scoccare del timer.
- [x] **TASK-22: API REST per la Gestione Schedulazioni**
  - **File coinvolti:** `backend/internal/delivery/http/handler/schedule_handler.go`
  - **DoD:** Endpoint `GET /api/v1/schedules`, `POST /api/v1/schedules`, `PUT /api/v1/schedules/{id}/toggle`, `DELETE /api/v1/schedules/{id}`.

### EPIC 9: White-Label Reporting Engine
- [x] **TASK-23: Gestione Branding Tenant (Logo & Palette Colori)**
  - **File coinvolti:** `backend/internal/repository/postgres/tenant_repo.go`, `backend/internal/delivery/http/handler/tenant_handler.go`
  - **DoD:** Endpoint `GET /api/v1/tenants/current` e `PUT /api/v1/tenants/current/branding` per configurare logo, colori primari e ragione sociale.
- [x] **TASK-24: Generatore Report PDF White-Label**
  - **File coinvolti:** `backend/pkg/report/pdf_generator.go`, `backend/internal/service/report_service.go`
  - **DoD:** Generazione di file PDF professionale contenente copertina con logo del tenant, Executive Summary con grafici delle severità, tabella dei target e schede tecniche dei findings.
- [x] **TASK-25: Esportazione Dati Tabellare CSV**
  - **File coinvolti:** `backend/pkg/report/csv_generator.go`
  - **DoD:** Endpoint `GET /api/v1/scans/{id}/reports/csv` per il download istantaneo dei findings in formato CSV conforme RFC 4180.

### EPIC 10: Frontend Single Page Application (React + Vite + Tailwind)
- [x] **TASK-26: Setup Architettura Frontend (Router, Auth Context, Axios Client)**
  - **File coinvolti:** `frontend/src/App.tsx`, `frontend/src/context/AuthContext.tsx`, `frontend/src/api/client.ts`
  - **DoD:** Routing applicativo con rotte protette, memorizzazione JWT, gestione automatica dei refresh token e layout responsive con sidebar di navigazione.
- [x] **TASK-27: Schermata Dashboard con Metriche e Grafici di Severità**
  - **File coinvolti:** `frontend/src/pages/Dashboard.tsx`, `frontend/src/components/SeverityChart.tsx`
  - **DoD:** Cruscotto con conteggio totale scansioni, host attivi, vulnerabilità divise per severità (Critical, High, Medium, Low) con grafici interattivi Recharts.
- [x] **TASK-28: Schermata Gestione Target & Modal di Creazione**
  - **File coinvolti:** `frontend/src/pages/Targets.tsx`, `frontend/src/components/TargetModal.tsx`
  - **DoD:** Tabella dei target con filtri, validazione client-side degli IP/CIDR e pulsante per eliminazione o scansione rapida.
- [x] **TASK-29: Schermata Scansioni (Lancio On-Demand, Monitoraggio Live, Schedulazioni)**
  - **File coinvolti:** `frontend/src/pages/Scans.tsx`, `frontend/src/components/NewScanModal.tsx`, `frontend/src/components/ScanProgressBadge.tsx`
  - **DoD:** Modal per selezionare target e profilo di scansione, tabella scansioni con polling/badge di stato in tempo reale e pulsante di cancellazione manuale.
- [x] **TASK-30: Schermata Dettaglio Scansione & Tabella Findings**
  - **File coinvolti:** `frontend/src/pages/ScanDetail.tsx`, `frontend/src/components/FindingsTable.tsx`, `frontend/src/components/FindingDetailModal.tsx`
  - **DoD:** Vista completa con breakdown vulnerabilità, filtri per CVSS/porta/IP, drawer/modal con dettagli CVE e remediation, e pulsanti di download PDF/CSV.
- [x] **TASK-31: Schermata Impostazioni Branding Tenant (White-Label)**
  - **File coinvolti:** `frontend/src/pages/Settings.tsx`
  - **DoD:** Interfaccia per caricare il logo aziendale, selezionare i colori del brand e visualizzare un'anteprima del report.

### EPIC 11: End-to-End Verification & PoC Validation
- [x] **TASK-32: Collaudo Integrato Full Stack in Docker Compose**
  - **File coinvolti:** `docker-compose.yml`, test scripts
  - **DoD:** Esecuzione di `docker-compose up`, login da interfaccia web, inserimento target locale di test, esecuzione scansione Nmap, correlazione con CVE locali, visualizzazione findings nella dashboard ed esportazione del report PDF con logo personalizzato.

---

## 🏆 MILESTONE 1.1 - Task Completati con Successo (Rilascio 2026-10-09)

### EPIC 12: Alive-Only Host Filtering & Range Data Sanitization (Feature 1)
- [x] **TASK-35: Filtraggio Host Inattivi nel Parser XML e Scan Orchestrator (Backend Go)**
  - **File coinvolti:** `backend/internal/scanner/nmap_parser.go`, `backend/internal/service/scan_orchestrator.go`, `backend/internal/scanner/nmap_parser_test.go`
  - **DoD:** Il parser XML e lo scan orchestrator escludono a monte dalla generazione dei findings tutti gli host aventi stato `down`. Solo gli host attivi (`up`) vengono elaborati e persistiti; le metriche di riepilogo mantengono il conteggio aggregato `up_hosts` e `down_hosts`. Unit test con fixture di range con host offline.
- [x] **TASK-36: Ottimizzazione UI e Reportistica per Visualizzazione Esclusiva Host Attivi**
  - **File coinvolti:** `frontend/src/pages/ScanDetail.tsx`, `frontend/src/components/FindingsTable.tsx`, `backend/pkg/report/pdf_generator.go`, `backend/pkg/report/csv_generator.go`
  - **DoD:** La schermata di dettaglio scansione e la tabella findings mostrano solo gli asset operativi; l'export PDF e CSV esclude categoricamente righe vuote per host non responsivi. Toggle client-side per escludere facoltativamente anche gli host privi di porte aperte.

### EPIC 13: Centralized Reports Hub & Dual-Template White-Label (Feature 2)
- [x] **TASK-37: Backend Endpoint REST e Motore Multi-Template Reports Hub**
  - **File coinvolti:** `backend/internal/delivery/http/handler/report_handler.go`, `backend/internal/service/report_service.go`, `backend/pkg/report/pdf_generator.go`, `backend/cmd/server/main.go`
  - **DoD:** Endpoint `GET /api/v1/reports` per consultare lo storico globale delle scansioni/report con filtri per target, date, severità e cartelle. Supporto al parametro query `template=executive` (report manageriale sintetico con grafici e posture index) e `template=technical` (dettaglio tecnico esteso).
- [x] **TASK-38: Frontend Reports Hub Dashboard (`frontend/src/pages/Reports.tsx`)**
  - **File coinvolti:** `frontend/src/pages/Reports.tsx`, `frontend/src/api/client.ts`, `frontend/src/App.tsx`
  - **DoD:** Schermata Reports completa che sostituisce il segnaposto con una dashboard interattiva: elenco globale dei report, filtri istantanei, selettore template (Executive o Tecnico), download immediato PDF/CSV e visualizzazione dello stato.

### EPIC 14: Hierarchical Tree Asset & Scan Folders (Feature 3)
- [x] **TASK-39: Database Migration e Backend Repository/Service Gruppi ad Albero (`parent_id`)**
  - **File coinvolti:** `backend/internal/repository/migrations/000004_tree_target_groups.up.sql`, `backend/internal/repository/migrations/000004_tree_target_groups.down.sql`, `backend/internal/domain/target.go`, `backend/internal/repository/postgres/target_group_repo.go`, `backend/internal/service/target_group_service.go`
  - **DoD:** Script di migrazione SQL che introduce `parent_id UUID REFERENCES target_groups(id)` e `color_tag`. Service con metodi per recuperare la gerarchia ad albero, calcolo ricorsivo del conteggio target e determinazione del livello di rischio aggregato. Unit test completi.
- [x] **TASK-40: Endpoint REST Gruppi ad Albero e Supporto Scansioni su Intera Cartella**
  - **File coinvolti:** `backend/internal/delivery/http/handler/target_group_handler.go`, `backend/internal/service/scan_service.go`, `backend/cmd/server/main.go`
  - **DoD:** Endpoint REST `GET /api/v1/target-groups` (albero nidificato), `POST /api/v1/target-groups`, `PUT /api/v1/target-groups/{id}`, `DELETE /api/v1/target-groups/{id}`. Supporto all'avvio di una scansione sull'intero `group_id` ricorsivo.
- [x] **TASK-41: Frontend Gestore Cartelle ad Albero (Sidebar Tree View & Drag/Drop)**
  - **File coinvolti:** `frontend/src/pages/Targets.tsx`, `frontend/src/components/FolderTree.tsx`, `frontend/src/components/FolderModal.tsx`, `frontend/src/types/target.ts`
  - **DoD:** Componente ad albero espandibile/collassabile nella sezione Targets, gestione cartelle (crea/rinomina/elimina/color-tag), filtro bersagli per cartella selezionata con badge contatore asset e rischio.
- [x] **TASK-42: Supporto Frontend Scansione su Intera Cartella (Scope Recursivo)**
  - **File coinvolti:** `frontend/src/components/NewScanModal.tsx`, `frontend/src/pages/Targets.tsx`
  - **DoD:** Possibilità di avviare una scansione selezionando come bersaglio un'intera cartella o sottocartella dall'albero; avviso di conferma con riepilogo del numero di host inclusi.

### EPIC 15: High-Speed LAN Network Discovery (Feature 4 - Advanced IP Scanner)
- [x] **TASK-43: Scanner ARP/Ping Sweep Rapido & Endpoint REST Discovery**
  - **File coinvolti:** `backend/internal/scanner/discovery_runner.go`, `backend/internal/service/discovery_service.go`, `backend/internal/delivery/http/handler/discovery_handler.go`, `backend/cmd/server/main.go`
  - **DoD:** Wrapper Go per scansione rapida locale (`nmap -sn -PR`), risoluzione vendor hardware OUI su MAC address, endpoint `POST /api/v1/network-discovery/scan`, `GET /api/v1/network-discovery/results` e `POST /api/v1/network-discovery/import-targets` con transazione atomica.
- [x] **TASK-44: Frontend Network Discovery Dashboard (`frontend/src/pages/NetworkDiscovery.tsx`)**
  - **File coinvolti:** `frontend/src/pages/NetworkDiscovery.tsx`, `frontend/src/App.tsx`, `frontend/src/components/layout/AppLayout.tsx`, `frontend/src/types/discovery.ts`, `frontend/src/api/discovery.ts`
  - **DoD:** Nuova vista dedicata nel menu laterale stile "Advanced IP Scanner": input subnet/range, barra di progresso, tabella in tempo reale con MAC, Produttore, IP, Hostname, Latenza, selezione multipla e pulsanti "Aggiungi a Cartella Target" e "Scansiona Subito".
- [x] **FIX-504-DISCOVERY: Risoluzione 504 Gateway Timeout in Network Discovery (Nmap -n, Worker Pool DNS e Proxy Timeout)**
  - **File coinvolti:** `backend/internal/scanner/profiles.go`, `backend/internal/service/discovery_service.go`, `backend/cmd/server/main.go`, `frontend/nginx.conf`, `frontend/src/api/client.ts`, `frontend/vite.config.ts`, `backend/internal/service/discovery_service_test.go`
  - **DoD:** Aggiunta flag `-n` ai profili LAN Discovery; risoluzione parallela con worker pool degli hostname in Go; elevazione dei timeout di Nginx (`proxy_read_timeout 300s`), del server HTTP Go (`WriteTimeout 300s`) e dei client Axios/Vite (`120s`). Tempo di scansione su `/24` abbattuto da >60s a 5.2s senza errori 504. Unit test dedicati.
- [x] **FIX-DISCOVERY-PRECISION: Risoluzione Falsi Positivi Host Offline e Arricchimento Multi-Source Porte/Vendor/MAC/HTTP**
  - **File coinvolti:** `backend/internal/scanner/profiles.go`, `backend/internal/scanner/arp_cache.go`, `backend/internal/scanner/oui_database.go`, `backend/internal/scanner/vendor_infer.go`, `backend/internal/scanner/http_grabber.go`, `backend/internal/service/discovery_service.go`, `frontend/src/pages/NetworkDiscovery.tsx`, `docker-compose.yml`, `backend/Dockerfile`
  - **DoD:** 
    1. Sostituzione dei probe TCP ACK (`-PA80`) che causavano RST gateway fallaci (rilevando 256/256 IP come attivi da container Docker NAT) con probe mirati ICMP Echo + TCP SYN (`-sn -PE -PS22,80,443,445 -T4 -n`) e filtraggio degli IP di rete (`.0`) e broadcast (`.255`).
    2. Implementazione Stage 2 Enrichment concorrente: scansione porte veloci (`-T4 -F -n`), estrazione banner/titoli HTTP con timeout 500ms, risoluzione fallback vendor basata su porte/titoli Web/OS.
    3. Risoluzione MAC Address multi-sorgente per superare il routing NAT di Docker (`/proc/net/arp`, Windows host `arp -a`, file di cache `/app/data/arp_cache.txt`).
    4. Database OUI arricchito con gestione MAC casuali/privati (IEEE 802 Locally Administered `(b & 0x03) == 0x02`).
    5. Default abilitato di `includePorts: true` nell'interfaccia utente. Risultato empirico su subnet reale: identificati esattamente 7/7 host attivi con dettagli completi (porte, MAC, Vendor, titolazione HTTP).

### EPIC 16: Remote SSH Host Management, iptables Firewall GUI & Web Terminal (Feature 5)
- [x] **TASK-45: Database Migration e Modulo Crittografico Credenziali SSH (AES-256-GCM)**
  - **File coinvolti:** `backend/internal/repository/migrations/000005_ssh_servers_firewall.up.sql`, `backend/internal/repository/migrations/000005_ssh_servers_firewall.down.sql`, `backend/pkg/crypto/aes_gcm.go`, `backend/internal/domain/ssh_server.go`, `backend/internal/repository/postgres/ssh_server_repo.go`
  - **DoD:** Tabella `ssh_servers` e `firewall_backups`. Modulo `pkg/crypto` con crittografia autenticata AES-256-GCM per proteggere password e chiavi private a riposo. Unit test crittografici al 100%.
- [x] **TASK-46: Client SSH Go Nativo e Servizio Gestione Firewall con Anti-Lockout Timer**
  - **File coinvolti:** `backend/internal/service/ssh_service.go`, `backend/internal/service/firewall_service.go`, `backend/internal/domain/firewall.go`
  - **DoD:** Client SSH basato su `golang.org/x/crypto/ssh` per password e chiavi private. Servizio firewall che esegue `iptables -L -n -v --line-numbers`, parsa le regole delle catene `INPUT`, `OUTPUT`, `FORWARD`, e applica modifiche attivando un goroutine timer di safety rollback (es. 60s) con restore automatico da snapshot se non confermato.
- [x] **TASK-47: WebSocket Gateway per Web Terminal SSH PTY Interattivo**
  - **File coinvolti:** `backend/internal/delivery/ws/ssh_terminal_ws.go`, `backend/cmd/server/main.go`
  - **DoD:** Endpoint `/ws/v1/ssh/servers/{id}/terminal` con upgrade WebSocket e validazione JWT. Allocazione PTY remoto e bridging bidirezionale dello stream stdin/stdout/stderr con supporto al resize terminale.
- [x] **TASK-48: Endpoint REST Gestione Server SSH e Operazioni Firewall**
  - **File coinvolti:** `backend/internal/delivery/http/handler/ssh_firewall_handler.go`, `backend/cmd/server/main.go`
  - **DoD:** Endpoint REST `GET /api/v1/ssh/servers`, `POST /api/v1/ssh/servers`, `POST /api/v1/ssh/servers/{id}/test-connection`, `DELETE /api/v1/ssh/servers/{id}`, `GET /api/v1/ssh/servers/{id}/firewall`, `POST /api/v1/ssh/servers/{id}/firewall/rules`, `DELETE /api/v1/ssh/servers/{id}/firewall/rules/{line}`, `POST /api/v1/ssh/servers/{id}/firewall/apply`, `POST /api/v1/ssh/servers/{id}/firewall/confirm`, `POST /api/v1/ssh/servers/{id}/firewall/rollback`.
- [x] **TASK-49: Frontend Interfaccia Grafica Firewall & Web Terminal SSH**
  - **File coinvolti:** `frontend/src/pages/SSHServers.tsx`, `frontend/src/components/FirewallManager.tsx`, `frontend/src/components/WebTerminalModal.tsx`, `frontend/src/App.tsx`, `frontend/src/components/layout/AppLayout.tsx`, `frontend/package.json`
  - **DoD:** Schermata per censimento server SSH; vista grafica a schede delle catene `INPUT`/`OUTPUT`/`FORWARD` con regole, azioni, protocolli e porte; modal guidato per nuova regola; banner countdown del Safety Rollback Timer con pulsante di conferma; modal terminale interattivo a schermo intero con `@xterm/xterm`.

### EPIC 17: Full-Stack Verification & Integration Test (Milestone 1.1)
- [x] **TASK-50: Test di Integrazione e Validazione E2E Full Stack**
  - **File coinvolti:** `docker-compose.yml`, `scripts/e2e_milestone_1_1.ps1`, `scripts/e2e_test.ps1`
  - **DoD:** Collaudo end-to-end automatizzato a 5 scenari con 28/28 asserzioni superate in ambiente Docker containerizzato.

---

## ⚡ MILESTONE 1.2 - Real-Time Streaming & Live Debug Console

> **Data di Rilascio:** 2026-10-09  
> **Obiettivo:** Streaming push in tempo reale tramite Server-Sent Events (SSE) senza polling per LAN Discovery e Nmap Scans, persistenza incrementale continua dei finding, resilienza alla riconnessione dei client (NFR-REL-05) e console di debug integrata per diagnosi rapida con esportazione AI Markdown.

### EPIC 18: Real-Time Live Streaming Engine via Server-Sent Events (Feature 6)
- [x] **TASK-51: Modello Eventi SSE e Servizio In-Memory Event Broadcaster (Backend Go)**
  - **File coinvolti:** `backend/internal/domain/event.go`, `backend/internal/service/event_broadcaster.go`, `backend/internal/service/event_broadcaster_test.go`
  - **DoD:** Creazione del modello dati SSE con formattazione a standard W3C (`text/event-stream`, `id`, `event`, `data`). Sviluppo del broker in-memory concorrente e thread-safe `EventBroadcaster` con multiplexing per topic/sessione/scan, policy di non-blocking drop con buffer regolabile e disiscrizione pulita alla cancellazione del context. 10/10 unit test passati.
- [x] **TASK-52: Live LAN Discovery Engine e Streaming Endpoint SSE (Backend Go)**
  - **File coinvolti:** `backend/internal/domain/discovery.go`, `backend/internal/service/discovery_service.go`, `backend/internal/delivery/http/handler/discovery_handler.go`, `backend/internal/service/discovery_service_test.go`, `backend/cmd/server/main.go`
  - **DoD:** Streaming degli host scoperti via probe ARP/ICMP/TCP con risoluzione OUI immediata e avanzamento percentuale progressivo. Endpoint `GET /api/v1/network-discovery/stream` con `http.Flusher`, header anti-buffering (`X-Accel-Buffering: no`), autenticazione JWT e replay immediato dei nodi già scoperti per client riconnessi. Persistenza incrementale continua in memoria/DB.
- [x] **TASK-53: Streaming Progressivo Scansioni Nmap e Scrittura Incrementale Findings (Backend Go)**
  - **File coinvolti:** `backend/internal/service/scan_orchestrator.go`, `backend/internal/scanner/nmap_runner.go`, `backend/internal/delivery/http/handler/scan_handler.go`, `backend/internal/service/scan_orchestrator_test.go`, `backend/cmd/server/main.go`
  - **DoD:** Monitoraggio dell'esecuzione Nmap in tempo reale tramite pipeline multi-stage (`HOST_DISCOVERY`, `PORT_SCAN`, `SERVICE_DETECTION`, `CVE_CORRELATION`), salvataggio progressivo su DB PostgreSQL dei finding via `UpsertFindingBatch` senza attendere il completamento della scansione. Emissione eventi `scan_phase_change`, `scan_progress`, `finding_discovered` e `scan_status_change`. Endpoint `GET /api/v1/scans/{id}/stream` con replay iniziale dei finding persistiti (NFR-REL-05).
- [x] **TASK-54: Client SSE e Aggiornamento Tabella Live in Network Discovery (Frontend React)**
  - **File coinvolti:** `frontend/src/hooks/useSSE.ts`, `frontend/src/hooks/useSSE.test.ts`, `frontend/src/types/discovery.ts`, `frontend/src/api/discovery.ts`, `frontend/src/pages/NetworkDiscovery.tsx`, `frontend/src/pages/NetworkDiscovery.test.tsx`
  - **DoD:** Custom hook generico `useSSE` per connessioni `EventSource` con gestione token Bearer, backoff esponenziale e riaggancio trasparente. Componente `NetworkDiscovery` con visualizzazione badge di avanzamento in tempo reale, animazione pulse verde su ciascun host scoperto e azioni rapide interattive disponibili istantaneamente su ciascuna riga. 58/58 test passati.
- [x] **TASK-55: Monitoraggio Live di Scansione e Findings Progressivi in ScanDetail (Frontend React)**
  - **File coinvolti:** `frontend/src/pages/ScanDetail.tsx`, `frontend/src/pages/ScanDetail.test.tsx`, `frontend/src/components/ScanProgressBadge.tsx`, `frontend/src/components/ScanProgressBadge.test.tsx`, `frontend/src/types/scan.ts`, `frontend/src/api/scan.ts`
  - **DoD:** Sostituzione integrale del polling timer in `ScanDetail` con abbonamento SSE reattivo. Indicatore di fase animato (`ScanProgressBadge`) con stepper dinamico a 4 fasi, percentuale avanzamento, e inserimento immediato dei finding nella tabella Findings non appena scoperti, con animazione glow e contatori severità aggiornati al volo. 65/65 test passati.
- [x] **TASK-56: Test di Integrazione e Validazione E2E per Funzionalità Real-Time (Milestone 1.2)**
  - **File coinvolti:** `backend/internal/delivery/http/handler/realtime_e2e_test.go`, `scripts/e2e_milestone_1_2.ps1`
  - **DoD:** Test suite di integrazione E2E Go e script PowerShell dedicati che collaudano lo streaming LAN Discovery con latenza per evento < 200ms (DoD 1), verifica persistenza incrementale su endpoint GET /results (DoD 2), ricezione ordinata di fasi, progresso e finding Nmap (DoD 3), e disconnessione/riconnessione client senza interruzione del job con replay resiliente di tutti i finding (DoD 4, NFR-REL-05). 100% test superati con exit code 0.

### EPIC 19: Debug & Testing Live Console Drawer (Feature 7)
- [x] **TASK-57: Ring Buffer Logger, Sensitive Data Sanitizer & Event Bus (Frontend TypeScript)**
  - **File coinvolti:** `frontend/src/types/debug.ts`, `frontend/src/utils/debugLogger.ts`, `frontend/src/utils/debugLogger.test.ts`
  - **DoD:** Ring buffer circolare in-memory da 500 eventi con sanitizzazione ricorsiva delle credenziali sensibili (`[REDACTED]`), calcolo duration request e categorizzazione log (`API`, `SSE`, `WS`, `RUNTIME`). 20/20 unit test passati.
- [x] **TASK-58: Intercettore Trasparente Axios, Handlers Globali e DebugContext (Frontend React)**
  - **File coinvolti:** `frontend/src/api/client.ts`, `frontend/src/context/DebugContext.tsx`, `frontend/src/context/DebugContext.test.tsx`, `frontend/src/App.tsx`
  - **DoD:** Intercettore globale Axios trasparente con misurazione latenze di rete, intercettazione automatica eccezioni non gestite `window.error` e `unhandledrejection`, e React Context Provider reattivo con contatore errori real-time.
- [x] **TASK-59: Componente UI Slide-Over Drawer, Floating Toggle & Markdown Exporter (Frontend React)**
  - **File coinvolti:** `frontend/src/components/debug/DebugConsoleDrawer.tsx`, `frontend/src/components/debug/DebugFloatingToggle.tsx`, `frontend/src/utils/markdownExporter.ts`, `frontend/src/components/layout/AppLayout.tsx`
  - **DoD:** Drawer laterale destro a comparsa con shortcut `Ctrl + Shift + D` e floating toggle button, filtri per livello/sorgente e ricerca testuale istantanea, viewer JSON espandibile, pulsante "Copia per AI (Markdown)" formattato per LLM prompt e download JSON.
- [x] **TASK-60: Verifica Type-Safety e Collaudo E2E Console di Debug (Frontend React)**
  - **File coinvolti:** `frontend/src/tests/debugConsoleE2E.test.tsx`, `frontend/src/App.tsx`, `frontend/package.json`
  - **DoD:** Suite E2E completa del ciclo di vita della console di debug con simulazione errori, filtri ed esportazione Markdown. Compilazione `npm run build` con 0 errori TypeScript.

---

## 🏢 WORKSTATION MULTI-COMPANY & WORKFLOW CONSOLIDATION - Sprint 2026-10-10

> **Data di Rilascio:** 2026-10-10  
> **Obiettivo:** Ottimizzazione completa dell'esperienza utente su workstation multi-azienda (audit su PC unico presso clienti diversi), flusso operativo sequenziale Passo 1 -> Passo 5, dossier report consolidati per azienda, rimozione ridondanze, personalizzazione profili Nmap con guida semplice interattiva, correlazione CVE cloud su NIST NVD 2.0 e perfezionamento grafico/tipografico dei report PDF A4.

### EPIC 20: Workstation Multi-Company Workflow (Categorizzazione & Isolamento Visivo)
- [x] **FEAT-IP-ORG-01 · Sidebar Navigation Workflow Categorization (Desktop & Mobile)**
  - **File:** `frontend/src/components/layout/AppLayout.tsx`
  - **DoD:** Riorganizzazione della barra laterale in macro-categorie orientate all'azione ("cosa si deve fare"): Overview, Scansione di Rete & Asset, Scansione Vulnerabilità & Audit, Infrastruttura & Accesso Remoto, Reportistica & Amministrazione.
- [x] **FEAT-IP-ORG-02 · Vista Raggruppata per Cartella/Azienda & Dettaglio IP in Scansioni (`/scans`)**
  - **File:** `frontend/src/pages/Scans.tsx`, `frontend/src/pages/Scans.test.tsx`
  - **DoD:** Scansioni aggregate per Cartella/Azienda Scope con card/accordion, statistiche nodi e vulnerabilità per cliente, espansione per IP ed espandi/comprimi globale.
- [x] **FEAT-IP-ORG-03 · Vista Raggruppata per Cartella/Azienda & Dettaglio IP in Vulnerabilità (`/findings`)**
  - **File:** `frontend/src/pages/Findings.tsx`, `frontend/src/pages/Findings.test.tsx`
  - **DoD:** Visualizzazione vulnerabilità raggruppata per Cartella/Azienda con navigazione gerarchica Cliente -> IP -> Findings e filtri per severità.
- [x] **FEAT-IP-ORG-04 · Vista Raggruppata per Cartella/Azienda & Dettaglio IP in Report White-Label (`/reports`)**
  - **File:** `frontend/src/pages/Reports.tsx`, `frontend/src/pages/Reports.test.tsx`
  - **DoD:** Hub report con card per azienda, Security Posture Score & Grade per cliente e download diretto a 1-click.

### EPIC 21: Personalizzazione Avanzata Profili di Scansione Nmap (FR-16)
- [x] **FEAT-SCAN-PROF-01 · Migration Database & Domain Model per Categorie e Profili Custom**
  - **File:** `backend/internal/repository/migrations/000006_scan_profile_categories.up.sql`, `backend/internal/domain/scan.go`, `backend/internal/repository/postgres/scan_repo.go`
  - **DoD:** Estensione tabella `scan_profiles` con categorie specialistiche (`PEN_TEST`, `SYSTEM_FLAWS`, `WEB_APP`, `DISCOVERY`, `COMPLIANCE`, `CUSTOM`) e isolamento tenant per profili custom.
- [x] **FEAT-SCAN-PROF-02 · Catalogo Profili Specialistici & API REST Profili Custom (Backend Go)**
  - **File:** `backend/internal/scanner/profiles.go`, `backend/internal/service/scan_service.go`, `backend/internal/delivery/http/handler/scan_handler.go`
  - **DoD:** Catalogo profili Nmap specialistici predefiniti, API REST per profili custom e validazione anti-injection dei flag.
- [x] **FEAT-SCAN-PROF-03 · Frontend UI Semplice, Multi-Obiettivo & Professionale per Avvio Scansione**
  - **File:** `frontend/src/components/NewScanModal.tsx`, `frontend/src/types/scan.ts`
  - **DoD:** Modale Nuova Scansione con profilo consigliato "Tutto in Uno (360°)", modalità "Componi Multi-Obiettivo" via checkbox, Guida Semplice con spiegazioni umane per PO e modalità Professionale con live command preview.
- [x] **FEAT-SCAN-PROF-04 · Suite di Test E2E, Unitari e Verifica Regressione Zero**
  - **File:** `backend/internal/scanner/profiles_test.go`, `frontend/src/components/NewScanModal.test.tsx`
  - **DoD:** Test completi su profili, parsing flag Nmap e rendering modale con 0 errori.

### EPIC 22: Ottimizzazione Flusso Operativo, Consolidamento Report per Azienda & Cleanup
- [x] **FEAT-WORKFLOW-CLEANUP-01 · Report Consolidati per Azienda (Executive PDF, Tecnico PDF, CSV RFC 4180)**
  - **File:** `backend/internal/service/report_service.go`, `backend/internal/delivery/http/handler/report_handler.go`, `frontend/src/api/reports.ts`, `frontend/src/pages/Reports.tsx`
  - **DoD:** Eliminazione report per singolo IP. In `/reports` ciascuna card rappresenta un'Azienda con Security Posture Score, Grade e download 1-click di Executive PDF (direzionale), Tecnico PDF (per host) e CSV RFC 4180 consolidato con deduplica.
- [x] **FEAT-WORKFLOW-CLEANUP-02 · Rimozione Opzione Ridondante Network Discovery da `/scans`**
  - **File:** `frontend/src/components/NewScanModal.tsx`, `frontend/src/components/NewScanModal.test.tsx`
  - **DoD:** Eliminazione categoria Discovery da `/scans`, centralizzandola in `/network-discovery`.
- [x] **FEAT-WORKFLOW-CLEANUP-03 · Rimozione Avvio Diretto Scansione da `/targets` e Indirizzamento a `/scans`**
  - **File:** `frontend/src/pages/Targets.tsx`
  - **DoD:** Eliminazione pulsante rapido Play da `/targets` e inserimento banner guida orientato a `/scans`.
- [x] **FEAT-WORKFLOW-CLEANUP-04 · Dashboard Panoramica Focalizzata sulle Singole Aziende & Flusso Lineare**
  - **File:** `frontend/src/pages/Dashboard.tsx`, `frontend/src/pages/Dashboard.test.tsx`
  - **DoD:** Dashboard a schede per ciascun cliente con Security Posture Score, metriche vulnerabilità e stepper operativo a 5 passaggi.
- [x] **FEAT-WORKFLOW-CLEANUP-05 · Targets: Rimozione Profili/Lancio, Eliminazione Massiva con Checkbox & Badge Dinamico**
  - **File:** `frontend/src/components/TreeExplorer.tsx`, `frontend/src/pages/Targets.tsx`
  - **DoD:** Selezione multipla di target con checkbox di riga e seleziona tutto, barra flottante, eliminazione massiva protetta e conteggio badge dinamico nell'albero.
- [x] **FEAT-WORKFLOW-CLEANUP-06 · Network Discovery: Creazione Cartella all'Import & Cache Descrizioni Host**
  - **File:** `frontend/src/pages/NetworkDiscovery.tsx`, `backend/internal/scanner/profiles.go`, `backend/internal/service/discovery_service.go`, `backend/internal/delivery/http/handler/discovery_handler.go`
  - **DoD:** Creazione rapida di una nuova cartella/azienda all'importazione degli IP, sonde Discovery potenziate (ICMP/TCP/UDP) e cache persistente delle descrizioni IP per tenant.
- [x] **FEAT-WORKFLOW-CLEANUP-07 · Sidebar Sequenziale (Passo 1 a 5) & Scansioni Esclusive per Cartella/Azienda**
  - **File:** `frontend/src/components/layout/AppLayout.tsx`, `frontend/src/components/NewScanModal.tsx`, `frontend/src/components/NewScanModal.test.tsx`
  - **DoD:** Barra laterale riordinata per sequenza di lavoro operativa (Passo 1 Discovery -> Passo 2 Target -> Passo 3 Scansioni -> Passo 4 Vulnerabilità -> Passo 5 Dossier & Report). Selezione vincolata alla Cartella/Azienda con anteprima dei chip IP contenuti.

### EPIC 23: Vulnerability Intelligence Online & Reportistica PDF A4 Avanzata
- [x] **FEAT-CVE-ONLINE-01 · Ricerca Vulnerabilità Online NIST NVD 2.0 & Script Nmap (`--script=vuln`) (FR-17, UC-16)**
  - **File:** `backend/internal/scanner/profiles.go`, `backend/internal/scanner/nmap_parser.go`, `backend/internal/service/online_cve_client.go`, `backend/internal/service/cve_correlator.go`, `backend/internal/service/scan_orchestrator.go`, `frontend/src/components/NewScanModal.tsx`
  - **DoD:** Whitelist sicura di `--script=vuln` e `--script=vulners`; parsing XML Nmap per script a livello porte e host con estrazione automatica CVE (es. SMB EternalBlue MS17-010, SSL, RCE); client REST NIST NVD API 2.0 con timeout resiliente a 4s e auto-caching persistente in PostgreSQL locale (`cveRepo.UpsertBatch`); supporto nei profili 360°, Multi-Obiettivo e Professionale.
- [x] **FEAT-REPORT-PDF-UX-01 · Perfezionamento Grafico, Tipografico & Coerenza Dati dei Report PDF A4 (FR-18, UC-17)**
  - **File:** `backend/internal/domain/report.go`, `backend/internal/service/cve_correlator.go`, `backend/internal/service/report_service.go`, `backend/pkg/report/pdf_generator.go`
  - **DoD:** Risoluzione di tutti i difetti segnalati: Executive compresso esattamente a 2 pagine dense senza buchi; testo remediation multiriga dinamico senza troncamenti né font miniaturizzati; separazione riga CVE e riga Azione; colori semantici coerenti (MAX CVSS arancione per High 7.5, host attivi in slate); normalizzazione scoring multi-host con floor 25/100 se 0 critici; nomi software reali dai banner (`dnsmasq`, `vsftpd`, `Dropbear`); protocolli insicuri (Telnet 23, FTP 21, SMB 445) elevati a finding prescrittivi; flusso tecnico continuo fino a 270mm senza vuoti da 150px; date e durate sempre valorizzate.

### EPIC 24: Attack Path Mapping Engine (Backend Go)
- [x] **FEAT-M2-02 · Motore di Calcolo Lateral Movement & API Grafo (Backend Go)**
  - **File:** `backend/internal/service/attack_path_service.go`, `backend/internal/delivery/http/handler/attack_path_handler.go`
  - **DoD:** Servizio di aggregazione topologica per calcolo dei cammini di movimento laterale, analisi dei pivot tra reti interne, identificazione chokepoint di remediation ad alto impatto e registrazione endpoint REST `/api/v1/attack-paths/graph` e `/api/v1/attack-paths/chokepoints`.

### EPIC 25: Milestone 2 · Visual Graph Explorer & Robustezza Suite
- [x] **FEAT-M2-03 · Componente UI Visual Graph Explorer (Frontend React)**
  - **File:** `frontend/src/types/attackPath.ts`, `frontend/src/api/attackPath.ts`, `frontend/src/components/graph/AttackPathGraph.tsx`, `frontend/src/pages/AttackPath.tsx`, `frontend/src/pages/AttackPath.test.tsx`
  - **DoD:** Pagina `/attack-path` e canvas SVG interattivo con controlli zoom/pan, layout topologico per livelli di penetrazione, evidenziazione dinamica dei percorsi di attacco, identificazione visiva dei chokepoint con raccomandazioni di remediation, filtri CVSS e navigazione integrata nella sidebar (Milestone 2 completata al 100%).
- [x] **ONBOARD-01 · Pipeline CI/CD GitHub Actions per Test & Build Automatici**
  - **File:** `.github/workflows/ci.yml`
  - **DoD:** Pipeline automatica per push e PR su main: job paralleli per Go test + lint e frontend npm test + vite build.
- [x] **ONBOARD-02 · Automated Container E2E Test Runner**
  - **File:** `scripts/run_e2e_docker.sh`, `scripts/run_e2e_docker.ps1`
  - **DoD:** Runner unificato multipiattaforma per lanciare lo stack containerizzato ed eseguire la sequenza completa di test E2E.
- [x] **ONBOARD-03 · Code-Splitting e Ottimizzazione Bundle Frontend (Vite manualChunks)**
  - **File:** `frontend/vite.config.ts`, `frontend/src/App.tsx`
  - **DoD:** Introduzione di `React.lazy` per le pagine e splitting dei vendor pesanti (`@xterm/xterm`, `recharts`, `lucide-react`), nessun chunk oltre i 500 kB.
- [x] **ONBOARD-04 · Risoluzione Warning nei Test Vitest (act & Recharts container)**
  - **File:** `frontend/src/tests/setup.ts`, `frontend/src/components/debug/DebugConsoleDrawer.test.tsx`, `frontend/src/pages/Reports.test.tsx`
  - **DoD:** Pulizia di tutti i warning console dei test Vitest: 100% test verdi con 0 warning a terminale.
- [x] **ONBOARD-05 · Test di Caratterizzazione per Scenari Limite Nmap & Rete**
  - **File:** `backend/internal/scanner/nmap_runner_test.go`, `backend/internal/service/scan_orchestrator_test.go`
  - **DoD:** Test isolati per simulare interruzione brusca (SIGKILL 137), payload troncato, payload binario corrotto e buffer stderr senza deadlock.

### EPIC 26: Hardening Critico & Scanner Security Perimeter (SEC-P0)
- [x] **SEC-P0-01 · Deny-List Target Interni & Risoluzione DNS Preventiva (FR-01.4)**
  - **File:** `backend/pkg/validator/target_validator.go`, `backend/internal/service/scan_orchestrator.go`, `backend/internal/service/discovery_service.go`
  - **DoD:** Blocco preventivo delle scansioni verso target locali, cloud metadata (169.254.169.254) o hostname interni con protezione anti-rebinding DNS.
- [x] **SEC-P0-02 · Stream-Ticket Monouso per Connessioni Real-Time SSE e WebSocket (NFR-SEC-04)**
  - **File:** `backend/pkg/auth/stream_ticket.go`, `backend/internal/delivery/http/handler/auth_handler.go`, `frontend/src/hooks/useSSE.ts`, `frontend/src/components/WebTerminalModal.tsx`
  - **DoD:** Sostituzione dei JWT a lungo termine nelle query string SSE/WS con stream-ticket usa-e-getta (TTL 30s) revocabili.
- [x] **SEC-P0-03 · Sanitizzazione e Argomenti Tipizzati Nmap con Delimitatore `--` (NFR-SEC-01)**
  - **File:** `backend/internal/scanner/profiles.go`, `backend/internal/scanner/nmap_runner.go`
  - **DoD:** Inserimento del delimitatore `--` prima del target per impedire flag injection via nomi ostili; allowlist tipizzata di flag, porte e script autorizzati.
- [x] **SEC-P0-04 · Disabilitazione `vulners` Esterno di Default & Modalità Offline Conforme (NFR-REL-03)**
  - **File:** `backend/internal/scanner/profiles.go`, `backend/internal/service/cve_correlator.go`
  - **DoD:** Disabilitazione dello script esterno `vulners` per evitare leak di metadati di rete; utilizzo di `--script "vuln and safe"` locale e database CVE offline.
- [x] **SEC-P0-05 · Redazione Completa Credenziali & Token nella Debug Console (NFR-DEV-01)**
  - **File:** `frontend/src/utils/debugLogger.ts`, `frontend/src/utils/markdownExporter.ts`
  - **DoD:** Mascheramento preventivo `[REDACTED]` per qualsiasi token, bearer header o credenziale sia a schermo che nell'esportatore "Copia per AI".
- [x] **SEC-P0-06 · Remote Server-Side Rollback Timer per Firewall iptables (NFR-REL-04)**
  - **File:** `backend/internal/service/firewall_service.go`
  - **DoD:** Timer di ripristino automatico anti-lockout gestito direttamente sull'host remoto tramite `nohup` e `iptables-restore`.

### EPIC 27: Scanner Engine Overhaul & Profiling Accuracy (SEC-P1)
- [x] **SEC-P1-01 · Profilo Scansione UDP Specialistico & Tuning Porte Critiche**
  - **File:** `backend/internal/scanner/profiles.go`, `frontend/src/components/NewScanModal.tsx`
  - **DoD:** Nuovo profilo specialistico `Specialized UDP Critical Services` (`-sU -p U:53,69,123,161,514,623,5353 -T4`) separato dalle code TCP.
- [x] **SEC-P1-02 · Correzione & Ricalibrazione Profili Nmap**
  - **File:** `backend/internal/scanner/profiles.go`, `backend/internal/scanner/nmap_parser.go`
  - **DoD:** Rimossa porta 69 (UDP) da Legacy Protocols TCP; reverse DNS resolution abilitata per LAN Discovery; gestione accurata di porte `filtered/unfiltered`.
- [x] **SEC-P1-03 · Tassonomia di Invasività Profili (`SAFE` vs `INTRUSIVE`)**
  - **File:** `backend/internal/domain/scan.go`, `backend/internal/delivery/http/handler/scan_handler.go`, `frontend/src/components/NewScanModal.tsx`
  - **DoD:** Tassonomia semantica (`SAFE`, `INTRUSIVE`, `DANGEROUS`) con badge visivo nella UI e blocco con checkbox di consenso obbligatorio prima del lancio.
- [x] **SEC-P1-04 · Arricchimento Modello Dati Finding (`confidence`, `detection_method`, `cpe`)**
  - **File:** `backend/internal/domain/finding.go`, `backend/internal/repository/migrations/000007_scanner_overhaul_findings.up.sql`, `frontend/src/types/scan.ts`
  - **DoD:** Aggiunta campi `confidence` (`CONFIRMED`, `PROBABLE`, `UNVERIFIED`), `detection_method` (`nse`, `banner`, `nvd_keyword`) e CPE normalizzato.
- [x] **SEC-P1-05 · Container Capabilities Hardening & `NMAP_PRIVILEGED`**
  - **File:** `docker-compose.yml`, `backend/Dockerfile`
  - **DoD:** Concessione minima `CAP_NET_RAW` per SYN scan (rimosso `CAP_NET_ADMIN`), abilitazione env `NMAP_PRIVILEGED=1`.
- [x] **SEC-P1-06 · Aggiornamento Libreria PDF (`go-pdf/fpdf`) e Manutenzione Dipendenze**
  - **File:** `backend/go.mod`, `backend/pkg/report/pdf_generator.go`
  - **DoD:** Migrazione dal package deprecato `jung-kurt/gofpdf` al fork supportato `github.com/go-pdf/fpdf v0.9.0`.

### EPIC 28: Attack Path Refinement & Advanced Orchestration (SEC-P2)
- [x] **SEC-P2-01 · Coerenza CVSS/Severity & Archi Inferiti con Confidence nel Grafo Attack Path**
  - **File:** `backend/internal/service/attack_path_service.go`, `frontend/src/components/graph/AttackPathGraph.tsx`
  - **DoD:** Punteggi CVSS rigorosamente allineati ai livelli di severità; archi `PIVOTS_TO` marcati come inferiti con badge percentuale, linea tratteggiata e tooltip euristico.
- [x] **SEC-P2-02 · Formula Rischio Cumulativo Probabilistica & Bound su Cammini**
  - **File:** `backend/internal/service/attack_path_service.go`
  - **DoD:** Formula combinata probabilistica $1 - \prod(1 - p_i)$, bound combinatorio a 50 percorsi e profondità 6 nodi.
- [x] **SEC-P2-03 · Diff tra Scansioni (Vulnerabilità Nuove, Risolte, Persistenti)**
  - **File:** `backend/internal/service/scan_orchestrator.go`, `backend/internal/delivery/http/handler/scan_handler.go`, `backend/internal/domain/finding.go`
  - **DoD:** Endpoint `GET /api/v1/scans/{id}/diff` che calcola in tempo reale il delta findings tra la scansione attuale e la precedente per il medesimo target (`NEW`, `RESOLVED`, `PERSISTENT`).
