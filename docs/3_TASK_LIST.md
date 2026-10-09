# 3. Interactive Task List & Implementation Backlog

> **Stato:** ATTIVO (Priorità validate dal Product Owner)  
> **Ultimo Aggiornamento:** 2026-10-09  
> **Autore / Ruolo:** TechLead-Engineer-Agent (System Architect & Tech Lead)  
> **Documenti Correlati:** [0_REQUIREMENTS.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/0_REQUIREMENTS.md) | [1_PROJECT_CONTEXT.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/1_PROJECT_CONTEXT.md) | [2_ARCHITECTURE.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/2_ARCHITECTURE.md) | [4_CHANGELOG.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/4_CHANGELOG.md)

---

## 📋 Regole di Avanzamento (Pipeline Autonoma)
1. **Intake & Plan:** L'agente analizza la richiesta, aggiorna i requisiti se necessario e produce il piano con strategia di test.
2. **Execute & Test:** Scrittura codice e test (happy path, error path, edge cases).
3. **Zero-Assumption Verification:** I test devono essere eseguiti a terminale con esito positivo al 100% (exit code 0) prima di procedere.
4. **Stati dei task:**
   - `[ ]` da fare
   - `[~]` implementato e verificato dall'agente con evidenze, in attesa di accettazione del Product Owner
   - `[x]` accettato dall'utente e verificato
   - `[!]` bloccato (con motivo indicato)
5. **Archiviazione:** I task completati e accettati nelle Milestone 1, 1.1 e 1.2 (TASK-01 fino a TASK-60) sono archiviati in [4_CHANGELOG.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/4_CHANGELOG.md). Oltre 10 task `[x]`, spostare in archivio.

---

## 🚀 Coda di Sviluppo Attiva (Prioritàvalidate dal PO)

### Sprint Immediato: Robustezza & Ottimizzazione (Approvati da PO)
- [x] **ONBOARD-01 · Pipeline CI/CD GitHub Actions per Test & Build Automatici**
  - File: `.github/workflows/ci.yml`
  - Dipende da: Nessuna
  - Test: Validazione workflow con test backend `go test ./...`, lint `go vet ./...`, frontend `npm test` e `npm run build`
  - DoD: Pipeline eseguita a ogni push/PR sui branch principali con job paralleli per backend Go (test e build) e frontend React (vitest e typecheck/vite build).

- [x] **ONBOARD-03 · Code-Splitting e Ottimizzazione Bundle Frontend (Vite manualChunks)**
  - File: `frontend/vite.config.ts`, `frontend/src/App.tsx`
  - Dipende da: Nessuna
  - Test: `npm run build` (nessun chunk > 500 kB)
  - DoD: Introduzione di `React.lazy` e `Suspense` per le route principali (`Dashboard`, `NetworkDiscovery`, `ScanDetail`, `Reports`, `SSHServers`, `Targets`, `Settings`, `Findings`) e configurazione di `build.rollupOptions.output.manualChunks` per vendor pesanti (`@xterm/xterm`, `recharts`, `lucide-react`).

### Sprint Attivo: Workstation Multi-Company Workflow (Navigazione & Organizzazione Gruppi IP)
- [~] **FEAT-IP-ORG-01 · Sidebar Navigation Workflow Categorization (Desktop & Mobile)**
  - File: `frontend/src/components/layout/AppLayout.tsx`
  - Dipende da: Nessuna
  - Test: `npm test` & verifica interattività menu/drawer
  - DoD: Riorganizzazione della barra laterale in macro-categorie orientate all'azione ("cosa si deve fare"): Overview, Scansione di Rete & Asset (Targets, Network Discovery), Scansione Vulnerabilità & Audit (Scansioni Nmap, Vulnerabilità CVE), Infrastruttura & Accesso Remoto (SSH & Firewall), Reportistica & Amministrazione (Report White-Label, Settings). Header di categoria tematici, badge contestuali e drawer mobile aggiornato.

- [~] **FEAT-IP-ORG-02 · Vista Raggruppata per Cartella/Azienda & Dettaglio IP in Scansioni (`/scans`)**
  - File: `frontend/src/pages/Scans.tsx`, `frontend/src/pages/Scans.test.tsx`
  - Dipende da: FEAT-IP-ORG-01
  - Test: Vitest UI rendering test su vista raggruppata ed espansione cartelle (4/4 PASS)
  - DoD: Scansioni visualizzate aggregate per cartella scope/azienda e non assegnati. Card/accordion per cartella con colore, conteggi scansioni, nodi up e vulnerabilità. Espansione della cartella per mostrare ciascun IP/target con le sue scansioni, stati, durata, chip vulnerabilità e azioni 1-click. Controlli "Espandi tutti" / "Comprimi tutti" e toggle vista raggruppata / vista tabellare lineare.

- [~] **FEAT-IP-ORG-03 · Vista Raggruppata per Cartella/Azienda & Dettaglio IP in Vulnerabilità (`/findings`)**
  - File: `frontend/src/pages/Findings.tsx`, `frontend/src/pages/Findings.test.tsx`
  - Dipende da: FEAT-IP-ORG-01
  - Test: Vitest UI rendering test su navigazione Cartella -> IP -> Findings (3/3 PASS)
  - DoD: Visualizzazione delle vulnerabilità strutturata per Cartella Scope/Azienda. Possibilità di aprire/selezionare la cartella per consultare la lista degli IP censiti per quel cliente e relative vulnerabilità CVE, con ricerca e filtri dedicati per IP e severità.

- [~] **FEAT-IP-ORG-04 · Vista Raggruppata per Cartella/Azienda & Dettaglio IP in Report White-Label (`/reports`)**
  - File: `frontend/src/pages/Reports.tsx`, `frontend/src/pages/Reports.test.tsx`
  - Dipende da: FEAT-IP-ORG-01
  - Test: Vitest UI rendering test su raggruppamento report per azienda e download 1-click (4/4 PASS)
  - DoD: Visualizzazione dei report raggruppati per Cartella/Azienda Scope con card di riepilogo Security Posture Score & Grade per cliente. Espansione per ispezionare gli specifici report per IP con 1-click download (Executive PDF, Tecnico PDF, CSV RFC 4180) e toggle vista raggruppata / vista cronologica.

### Sprint Attivo: Personalizzazione Avanzata Profili di Scansione & Categorie Specialistiche (FR-16)
- [~] **FEAT-SCAN-PROF-01 · Migration Database & Domain Model per Categorie e Profili Custom**
  - File: `backend/internal/repository/migrations/000006_scan_profile_categories.up.sql`, `backend/internal/domain/scan.go`, `backend/internal/repository/postgres/scan_repo.go`
  - Dipende da: Nessuna
  - Test: Unit & migration integration test (PASS)
  - DoD: Estensione tabella `scan_profiles` con colonna `category` (PEN_TEST, SYSTEM_FLAWS, WEB_APP, DISCOVERY, COMPLIANCE, CUSTOM), `is_custom` e `tenant_id` opzionale con indici. Aggiornamento interfaccia `domain.ScanProfileRepository` (ListByCategory, CreateCustom, DeleteCustom) e mapping struct Go.

- [~] **FEAT-SCAN-PROF-02 · Catalogo Profili Specialistici & API REST Profili Custom (Backend Go)**
  - File: `backend/internal/scanner/profiles.go`, `backend/internal/service/scan_service.go`, `backend/internal/delivery/http/handler/scan_handler.go`
  - Dipende da: FEAT-SCAN-PROF-01
  - Test: Unit & integration test su validazione Nmap safe args e REST API handler (PASS)
  - DoD: Registrazione catalogo profili predefiniti per ciascuna categoria specialistica (Stealth SYN, Aggressive Pen Test, Firewall Bypass, System Flaws OS & Demoni, Critical Services RCE, Web Stack Audit, Discovery Sweep, Insecure Protocols). Endpoint `GET /api/v1/scan-profiles?category=...`, `POST /api/v1/scan-profiles` con validazione anti-injection e `DELETE /api/v1/scan-profiles/{id}`.

- [~] **FEAT-SCAN-PROF-03 · Frontend UI Semplice & Professionale per Avvio e Configurazione Scansione**
  - File: `frontend/src/components/NewScanModal.tsx`, `frontend/src/types/scan.ts`
  - Dipende da: FEAT-SCAN-PROF-02
  - Test: Vitest rendering test su navigazione categorie, configuratore porte/stealth e live command preview (7/7 PASS)
  - DoD: Modale Nuova Scansione con:
    1. Modalità Guidata / Semplice: card delle 5 categorie specialistiche (Pen Test, Falle di Sistema, Web App, Discovery, Compliance) con schede profilo complete (descrizione, cosa rileva, impatto);
    2. Modalità Professionale / Avanzata: configuratore porte (Top 100/1000, All 65535, porte custom), timing template (T1-T5), stealth options (`-sS`, `-sA`, `-Pn`, `--open`), OS detection (`-O`), live preview del comando Nmap sicuro e possibilità di salvataggio come profilo custom;
    3. Pieno supporto a scansioni on-demand e schedulazioni ricorrenti.

- [~] **FEAT-SCAN-PROF-04 · Suite di Test E2E, Unitari e Verifica Regressione Zero**
  - File: `backend/internal/scanner/profiles_test.go`, `frontend/src/components/NewScanModal.test.tsx`
  - Dipende da: FEAT-SCAN-PROF-03
  - Test: `go test ./...` (14/14 pkg PASS) e `npm test` (83/83 test PASS, exit code 0)
  - DoD: Validazione completa end-to-end con test automatici su parsing profili, avvio scansione con profilo specialistico o custom e verifica assenza di regressioni.

- [x] **FEAT-M2-01 · Spike & Intake: Attack Path Mapping (Grafo Visuale Vulnerabilità)**
  - File: `docs/0_REQUIREMENTS.md`, `docs/domains/attack_path.md`, `backend/internal/domain/attack_path.go`
  - Dipende da: ONBOARD-01, ONBOARD-03
  - Test: Unit test su algoritmi di correlazione percorsi e grafo adiacenze
  - DoD: Definizione formale degli Use Cases e modelli di dominio per collegamenti tra vulnerabilità scoperte, credenziali compromesse, nodi raggiungibili e calcolo del lateral movement; progettazione delle API REST per il grafo.

- [~] **FEAT-M2-02 · Motore di Calcolo Lateral Movement & API Grafo (Backend Go)**
  - File: `backend/internal/service/attack_path_service.go`, `backend/internal/delivery/http/handler/attack_path_handler.go`
  - Dipende da: FEAT-M2-01
  - Test: Service & HTTP integration tests (happy, sad, edge)
  - DoD: Servizio Go che aggrega Scans, Findings, Subnet e Server SSH per generare nodi ed archi orientati con punteggio di rischio complessivo del percorso; endpoint REST `/api/v1/attack-paths/graph` e `/api/v1/attack-paths/chokepoints`.

- [ ] **FEAT-M2-03 · Componente UI Visual Graph Explorer (Frontend React)**
  - File: `frontend/src/pages/AttackPath.tsx`, `frontend/src/components/graph/AttackPathGraph.tsx`
  - Dipende da: FEAT-M2-02
  - Test: Vitest UI rendering test
  - DoD: Nuova pagina `/attack-path` con rendering visuale interattivo a nodi e connessioni (es. zoom, pan, filtro per severità e clic sul nodo per dettagli vulnerabilità ed exploit path).

### Backlog Secondario (Manutenzione)
- [ ] **ONBOARD-02 · Automated Container E2E Test Runner**
  - File: `scripts/run_e2e_docker.sh`, `scripts/e2e_milestone_1_1.ps1`, `scripts/e2e_milestone_1_2.ps1`
  - Dipende da: ONBOARD-01
  - Test: Esecuzione script E2E in container Linux Docker headless
  - DoD: Runner unificato multipiattaforma (Linux/Windows) per avviare lo stack `docker compose` ed eseguire gli scenari E2E validando API REST, stream SSE e terminale WebSocket.

- [ ] **ONBOARD-04 · Risoluzione Warning nei Test Vitest (act & Recharts container)**
  - File: `frontend/src/components/debug/DebugConsoleDrawer.test.tsx`, `frontend/src/tests/debugConsoleE2E.test.tsx`, `frontend/src/pages/ScanDetail.test.tsx`
  - Dipende da: Nessuna
  - Test: `npm test` (0 warning a console)
  - DoD: Incapsulamento corretto degli aggiornamenti di stato asincroni in `act(...)` per il DebugProvider e mock con dimensioni definite per il ResponsiveContainer di Recharts durante i test.

- [ ] **ONBOARD-05 · Test di Caratterizzazione per Scenari Limite Nmap & Rete**
  - File: `backend/internal/scanner/nmap_runner_test.go`, `backend/internal/service/scan_orchestrator_test.go`
  - Dipende da: Nessuna
  - Test: `go test -v ./internal/scanner/... ./internal/service/...`
  - DoD: Test isolati per simulare interruzione brusca del processo Nmap, segnali di kill anticipati, output XML corrotto o incompleto, e fallback elegante senza panic né memory leak.
