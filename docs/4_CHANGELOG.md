# 4. Archivio Task Completati & Changelog (Changelog & Archive)

> **Progetto:** NextGen CyberSecurity Platform  
> **Autore / Ruolo:** TechLead-Engineer-Agent (System Architect & Tech Lead)  
> **Archiviazione Eseguita:** 2026-10-08  
> **Riferimento Regole:** [Agent.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/Agent.md) (Token Management & Modular Archiving)

---

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

