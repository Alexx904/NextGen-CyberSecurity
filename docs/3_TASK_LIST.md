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

- [~] **ONBOARD-03 · Code-Splitting e Ottimizzazione Bundle Frontend (Vite manualChunks)**
  - File: `frontend/vite.config.ts`, `frontend/src/App.tsx`
  - Dipende da: Nessuna
  - Test: `npm run build` (nessun chunk > 500 kB)
  - DoD: Introduzione di `React.lazy` e `Suspense` per le route principali (`Dashboard`, `NetworkDiscovery`, `ScanDetail`, `Reports`, `SSHServers`, `Targets`, `Settings`, `Findings`) e configurazione di `build.rollupOptions.output.manualChunks` per vendor pesanti (`@xterm/xterm`, `recharts`, `lucide-react`).

### Milestone 2: Feature Cardine Scelta dal PO
- [ ] **FEAT-M2-01 · Spike & Intake: Attack Path Mapping (Grafo Visuale Vulnerabilità)**
  - File: `docs/0_REQUIREMENTS.md`, `docs/domains/attack_path.md`, `backend/internal/domain/attack_path.go`
  - Dipende da: ONBOARD-01, ONBOARD-03
  - Test: Unit test su algoritmi di correlazione percorsi e grafo adiacenze
  - DoD: Definizione formale degli Use Cases e modelli di dominio per collegamenti tra vulnerabilità scoperte, credenziali compromesse, nodi raggiungibili e calcolo del lateral movement; progettazione delle API REST per il grafo.

- [ ] **FEAT-M2-02 · Motore di Calcolo Lateral Movement & API Grafo (Backend Go)**
  - File: `backend/internal/service/attack_path_service.go`, `backend/internal/delivery/http/handler/attack_path_handler.go`
  - Dipende da: FEAT-M2-01
  - Test: Service & HTTP integration tests (happy, sad, edge)
  - DoD: Servizio Go che aggrega Scans, Findings, Subnet e Server SSH per generare nodi ed archi orientati con punteggio di rischio complessivo del percorso.

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
