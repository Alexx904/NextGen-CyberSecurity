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

### Sprint Archiviato: Workstation Multi-Company, Workflow Consolidation & Vulnerability Intelligence (Accettati dal PO)
> *Nota di archiviazione:* Tutti i task completati e convalidati (EPIC 20, 21, 22, 23: `FEAT-IP-ORG-01..04`, `FEAT-SCAN-PROF-01..04`, `FEAT-WORKFLOW-CLEANUP-01..07`, `FEAT-CVE-ONLINE-01`, `FEAT-REPORT-PDF-UX-01`) sono stati archiviati nel [4_CHANGELOG.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/4_CHANGELOG.md) in conformità alle regole di token & backlog management di `AGENTS.md`.

### Sprint Attivo: Milestone 2 · Attack Path Mapping (Grafo Visuale Vulnerabilità & Lateral Movement)
- [x] **FEAT-M2-01 · Spike & Intake: Attack Path Mapping (Grafo Visuale Vulnerabilità)**
  - File: `docs/0_REQUIREMENTS.md`, `docs/domains/attack_path.md`, `backend/internal/domain/attack_path.go`
  - Dipende da: ONBOARD-01, ONBOARD-03
  - Test: Unit test su algoritmi di correlazione percorsi e grafo adiacenze
  - DoD: Definizione formale degli Use Cases e modelli di dominio per collegamenti tra vulnerabilità scoperte, credenziali compromesse, nodi raggiungibili e calcolo del lateral movement; progettazione delle API REST per il grafo.

- [x] **FEAT-M2-02 · Motore di Calcolo Lateral Movement & API Grafo (Backend Go)**
  - File: `backend/internal/service/attack_path_service.go`, `backend/internal/delivery/http/handler/attack_path_handler.go`
  - Dipende da: FEAT-M2-01
  - Test: Service & HTTP integration tests (happy, sad, edge) (PASS)
  - DoD: Servizio Go che aggrega Scans, Findings, Subnet e Server SSH per generare nodi ed archi orientati con punteggio di rischio complessivo del percorso; endpoint REST `/api/v1/attack-paths/graph` e `/api/v1/attack-paths/chokepoints`.

- [~] **FEAT-M2-03 · Componente UI Visual Graph Explorer (Frontend React)**
  - File: `frontend/src/types/attackPath.ts`, `frontend/src/api/attackPath.ts`, `frontend/src/components/graph/AttackPathGraph.tsx`, `frontend/src/pages/AttackPath.tsx`, `frontend/src/pages/AttackPath.test.tsx`, `frontend/src/App.tsx`, `frontend/src/components/layout/AppLayout.tsx`
  - Dipende da: FEAT-M2-02
  - Test: Vitest UI rendering test `AttackPath.test.tsx` (6/6 PASS), suite completa frontend `npm test` (95/95 PASS, exit code 0)
  - DoD: Nuova pagina `/attack-path` e componente grafico `AttackPathGraph` interattivo: canvas SVG con controlli zoom/pan, distribuzione topologica per livelli di rischio/hop, evidenziazione dinamica dei percorsi di attacco, visualizzazione chokepoint ad alto ROI con prescrizione di remediation, filtri CVSS e navigazione integrata nella sidebar (Milestone 2 completata al 100%).

### Backlog Secondario (Manutenzione & Robustezza)
- [~] **ONBOARD-02 · Automated Container E2E Test Runner**
  - File: `scripts/run_e2e_docker.sh`, `scripts/run_e2e_docker.ps1`, `scripts/e2e_milestone_1_1.ps1`, `scripts/e2e_milestone_1_2.ps1`
  - Dipende da: ONBOARD-01
  - Test: Esecuzione script E2E in container Linux Docker headless e sintassi PowerShell validata al 100%
  - DoD: Runner unificato multipiattaforma (`run_e2e_docker.sh` per Linux/CI e `run_e2e_docker.ps1` per Windows) per avviare lo stack `docker compose` ed eseguire l'intera sequenza di 10 test E2E validando API REST, stream SSE, SPA frontend e calcolo Attack Paths.

- [~] **ONBOARD-04 · Risoluzione Warning nei Test Vitest (act & Recharts container)**
  - File: `frontend/src/tests/setup.ts`, `frontend/vite.config.ts`, `frontend/src/components/debug/DebugConsoleDrawer.test.tsx`, `frontend/src/tests/debugConsoleE2E.test.tsx`, `frontend/src/context/DebugContext.test.tsx`, `frontend/src/pages/Reports.test.tsx`, `frontend/src/pages/ScanDetail.test.tsx`
  - Dipende da: Nessuna
  - Test: `npm test` (16 file, 95 test PASS al 100%, **0 warning a console**)
  - DoD: Eliminazione totale di tutti i warning stderr dei test frontend: introdotto `setup.ts` con cleanup globale automatico del DOM e soppressione warning informativi v7, incapsulamento rigoroso in `cleanup()` per il DebugProvider e mock di dimensione per il ResponsiveContainer di Recharts.

- [~] **ONBOARD-05 · Test di Caratterizzazione per Scenari Limite Nmap & Rete**
  - File: `backend/internal/scanner/nmap_runner_test.go`, `backend/internal/service/scan_orchestrator_test.go`
  - Dipende da: Nessuna
  - Test: `go test -v ./internal/scanner/... ./internal/service/...` (100% PASS, exit code 0)
  - DoD: Test isolati per simulare interruzione brusca del processo Nmap (SIGKILL / exit 137), payload XML troncato o incompleto, payload binario corrotto non-XML, buffer stderr di grandi dimensioni senza deadlock, e transizione atomica elegante a FAILED dell'orchestrator con registrazione dell'errore senza panic né memory leak.

---

## 🛡️ Roadmap di Remediation Sicurezza & Scanner Engine (Audit 10 Ottobre 2026)

### Epic 24 · SEC-P0: Hardening Critico & Scanner Security Perimeter (Priorità Immediata)
- [~] **SEC-P0-01 · Deny-List Target Interni & Risoluzione DNS Preventiva (FR-01.4)**
  - File: `backend/pkg/validator/target_validator.go`, `backend/internal/service/scan_orchestrator.go`, `backend/internal/service/discovery_service.go`
  - Dipende da: Nessuna
  - Test: Unit test su validazione IP/CIDR/FQDN (`TestCheckTargetScope` 100% PASS)
  - DoD: Blocco categorico all'inserimento target e al lancio scansione/discovery per destinazioni verso la piattaforma stessa, con risoluzione DNS preventiva e verifica dell'IP risolto.

- [~] **SEC-P0-02 · Stream-Ticket Monouso per Connessioni Real-Time SSE e WebSocket (NFR-SEC-04)**
  - File: `backend/pkg/auth/stream_ticket.go`, `backend/internal/delivery/http/handler/auth_handler.go`, `backend/internal/delivery/http/middleware/auth_middleware.go`, `frontend/src/api/client.ts`, `frontend/src/hooks/useSSE.ts`, `frontend/src/components/WebTerminalModal.tsx`
  - Dipende da: Nessuna
  - Test: Integration test HTTP e auth test (`TestStreamTicketStore`, `TestAuthMiddleware_StreamTicket` 100% PASS), handshake SSE e WS con ticket query string monouso
  - DoD: Eliminazione dei token JWT a lungo termine dalle query string degli stream SSE e WebSocket; passaggio a ticket usa-e-getta revocabile che previene leakage nei log Nginx, nella cronologia browser e nei proxy.

- [~] **SEC-P0-03 · Sanitizzazione e Argomenti Tipizzati Nmap con Delimitatore `--` (NFR-SEC-01)**
  - File: `backend/internal/scanner/profiles.go`, `backend/internal/scanner/nmap_runner.go`
  - Dipende da: Nessuna
  - Test: Unit test di regressione su `BuildSafeArguments` e `IsSafeArgument` (100% PASS)
  - DoD: Inserimento del delimitatore `--` prima del target per impedire flag injection via nomi ostili; passaggio da blacklist a struttura di allowlist rigorosa e tipizzata su porte, timing, rate e categorie script autorizzate; divieto assoluto di percorsi file locali in `--script`.

- [~] **SEC-P0-04 · Disabilitazione `vulners` Esterno di Default & Modalità Offline Conforme (NFR-REL-03)**
  - File: `backend/internal/scanner/profiles.go`, `backend/internal/service/cve_correlator.go`
  - Dipende da: Nessuna
  - Test: Verifica profili e flag `--script` per default offline (`internal/scanner` 100% PASS)
  - DoD: Rimozione di `vulners` dai profili attivi predefiniti (invio di dati tecnici di rete a terze parti vietato senza consenso esplicito); utilizzo rigoroso di `--script "vuln and safe"` locale e gestione offline/air-gapped.

- [~] **SEC-P0-05 · Redazione Completa Credenziali & Token nella Debug Console (NFR-DEV-01)**
  - File: `frontend/src/utils/debugLogger.ts`, `frontend/src/utils/markdownExporter.ts`
  - Dipende da: Nessuna
  - Test: Vitest su debugLogger e markdownExporter con payload contenenti token query param, bearer header e credenziali (36/36 test PASS)
  - DoD: Mascheramento automatico preventivo di qualsiasi occorrenza di `token=`, `bearer`, `passphrase`, password sia nella visualizzazione a schermo che nell'esportatore "Copia per AI".

- [~] **SEC-P0-06 · Remote Server-Side Rollback Timer per Firewall iptables (NFR-REL-04)**
  - File: `backend/internal/service/firewall_service.go`
  - Dipende da: Nessuna
  - Test: Test di orchestrazione comandi rollback con mock SSH (`internal/service` 100% PASS)
  - DoD: Timer di rollback anti-lockout gestito direttamente sull'host remoto tramite watchdog distaccato `nohup` (ripristino garantito anche in caso di caduta connessione o crash del backend).

### Epic 25 · SEC-P1: Scanner Engine Overhaul & Profiling Accuracy (Priorità Alta)
- [ ] **SEC-P1-01 · Profilo Scansione UDP Specialistico & Tuning Porte Critiche**
  - File: `backend/internal/scanner/profiles.go`, `frontend/src/components/NewScanModal.tsx`
  - Dipende da: SEC-P0-03
  - Test: Scanner test su argomenti UDP (`-sU -p U:53,69,123,161,514,623,5353`)
  - DoD: Profilo UDP dedicato separato dai profili TCP per tempi di scansione ottimali, con rilevamento mirato di SNMP, TFTP, Syslog, IPMI.

- [ ] **SEC-P1-02 · Correzione & Ricalibrazione Profili Nmap**
  - File: `backend/internal/scanner/profiles.go`, `backend/internal/scanner/nmap_parser.go`
  - Dipende da: SEC-P0-03
  - Test: Test unitari su combinazioni profili e parsing stati porta
  - DoD: Correzione porta 69 (UDP) in Legacy Protocols, profilazione ACK scan con tipi di risultato `filtered/unfiltered` dedicati (non trasformati in falso positivo porta aperta), gestione Alive-Only consapevole quando presente `-Pn`, rimozione `-n` o arricchimento reverse DNS per LAN Discovery.

- [ ] **SEC-P1-03 · Tassonomia di Invasività Profili (`SAFE` vs `INTRUSIVE`)**
  - File: `backend/internal/domain/scan.go`, `frontend/src/components/NewScanModal.tsx`
  - Dipende da: SEC-P1-02
  - Test: Validazione livello di rischio profilo ed esclusioni NSE
  - DoD: Classificazione operativa di ogni profilo (`SAFE`, `INTRUSIVE`, `DANGEROUS`), esclusione di categorie NSE intrusive dai default e gate con conferma esplicita per profili intrusivi.

- [ ] **SEC-P1-04 · Arricchimento Modello Dati Finding (`confidence`, `detection_method`, `cpe`)**
  - File: `backend/internal/domain/finding.go`, `backend/internal/repository/migrations/`, `frontend/src/types/scan.ts`
  - Dipende da: Nessuna
  - Test: Migration e test repository
  - DoD: Introduzione nei findings di `confidence` (`CONFIRMED`, `PROBABLE`, `UNVERIFIED`), `detection_method` (`nse`, `banner`, `nvd_keyword`) e tracciamento CPE normalizzato.

- [ ] **SEC-P1-05 · Container Capabilities Hardening & `NMAP_PRIVILEGED`**
  - File: `docker-compose.yml`, `backend/Dockerfile`
  - Dipende da: Nessuna
  - Test: Verifica container startup e capabilities
  - DoD: Rimozione di `NET_ADMIN` e configurazione esplicita di `NMAP_PRIVILEGED=1` nell'ambiente container; documentazione agent/probe locale per LAN Layer 2/ARP discovery in scenari non-Linux.

- [ ] **SEC-P1-06 · Aggiornamento Libreria PDF (`go-pdf/fpdf`) e Manutenzione Dipendenze**
  - File: `backend/go.mod`, `backend/pkg/report/pdf_generator.go`
  - Dipende da: Nessuna
  - Test: `go test ./pkg/report/...` (100% PASS)
  - DoD: Migrazione da `jung-kurt/gofpdf` (abbandonata) al fork supportato `github.com/go-pdf/fpdf`.

### Epic 26 · SEC-P2: Attack Path Refinement & Advanced Orchestration (Miglioramento Continuo)
- [ ] **SEC-P2-01 · Coerenza CVSS/Severity & Archi Inferiti con Confidence nel Grafo Attack Path**
  - File: `backend/internal/service/attack_path_service.go`, `frontend/src/components/graph/AttackPathGraph.tsx`
  - Dipende da: Nessuna
  - Test: Unit test su assegnazione severità nodi e proprietà archi
  - DoD: Allineamento rigoroso tra score CVSS e severità dei nodi (es. CVSS 7.5 = HIGH, non CRITICAL); marcatura esplicita degli archi `PIVOTS_TO` e `TRUSTED_RELATION` con flag `inferred: true`, base dell'inferenza e score di confidenza.

- [ ] **SEC-P2-02 · Formula Rischio Cumulativo Probabilistica & Bound su Cammini**
  - File: `backend/internal/service/attack_path_service.go`
  - Dipende da: SEC-P2-01
  - Test: Unit test su calcolo probabilistico $1 - \prod(1 - p_i)$
  - DoD: Sostituzione della somma algebrica pura con formulazione di rischio probabilistico combinato per evitare distorsioni su molte vulnerabilità minori; vincoli di profondità massima e tetto al numero di percorsi enumerati.

- [ ] **SEC-P2-03 · Diff tra Scansioni (Vulnerabilità Nuove, Risolte, Persistenti)**
  - File: `backend/internal/service/scan_orchestrator.go`, `backend/internal/domain/finding.go`
  - Dipende da: SEC-P1-04
  - Test: Service integration test su scansioni consecutive dello stesso target
  - DoD: Rilevamento automatico del differenziale findings tra scansione corrente e precedente per evidenziare remediations riuscite o nuove esposizioni.
