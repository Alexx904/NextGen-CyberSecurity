# 3. Interactive Task List & Implementation Backlog

> **Stato:** ATTIVO (Struttura Autonoma v6.0 - Onboarding Concluso)  
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

## 📌 Backlog Iniziale Post-Onboarding

### Priorità Alta (P1 - Robustezza Infrastrutturale & CI)
- [ ] **ONBOARD-01 · Pipeline CI/CD GitHub Actions per Test & Build Automatici**
  - File: `.github/workflows/ci.yml`
  - Dipende da: Nessuna
  - Test: Workflow GitHub Actions (lint, backend `go test ./...`, frontend `npm test` e `npm run build`)
  - DoD: Pipeline eseguita a ogni push/PR con verifica completa di compilazione backend Go, suite test backend, type-safety TypeScript, test Vitest e build Vite.

- [ ] **ONBOARD-02 · Automated Container E2E Test Runner**
  - File: `scripts/run_e2e_docker.sh`, `scripts/e2e_milestone_1_1.ps1`, `scripts/e2e_milestone_1_2.ps1`
  - Dipende da: ONBOARD-01
  - Test: Esecuzione script E2E in container Linux Docker headless
  - DoD: Runner unificato multipiattaforma (Linux/Windows) per avviare lo stack `docker compose` ed eseguire gli scenari E2E validando API REST, stream SSE e terminale WebSocket.

### Priorità Media (P2 - Debito Tecnico & Ottimizzazioni Frontend)
- [ ] **ONBOARD-03 · Code-Splitting e Ottimizzazione Bundle Frontend (Vite manualChunks)**
  - File: `frontend/vite.config.ts`, `frontend/src/App.tsx`
  - Dipende da: Nessuna
  - Test: `npm run build` (nessun chunk > 500 kB)
  - DoD: Introduzione di `React.lazy` e `Suspense` per le route principali (`Dashboard`, `NetworkDiscovery`, `ScanDetail`, `Reports`, `SSHServers`) e configurazione di `build.rollupOptions.output.manualChunks` per vendor pesanti (`@xterm/xterm`, `recharts`, `lucide-react`).

- [ ] **ONBOARD-04 · Risoluzione Warning nei Test Vitest (act & Recharts container)**
  - File: `frontend/src/components/debug/DebugConsoleDrawer.test.tsx`, `frontend/src/tests/debugConsoleE2E.test.tsx`, `frontend/src/pages/ScanDetail.test.tsx`
  - Dipende da: Nessuna
  - Test: `npm test` (0 warning a console)
  - DoD: Incapsulamento corretto degli aggiornamenti di stato asincroni in `act(...)` per il DebugProvider e mock con dimensioni definite (`width: 500, height: 300`) per il ResponsiveContainer di Recharts durante i test.

- [ ] **ONBOARD-05 · Test di Caratterizzazione per Scenari Limite Nmap & Rete**
  - File: `backend/internal/scanner/nmap_runner_test.go`, `backend/internal/service/scan_orchestrator_test.go`
  - Dipende da: Nessuna
  - Test: `go test -v ./internal/scanner/... ./internal/service/...`
  - DoD: Test isolati per simulare interruzione brusca del processo Nmap, segnali di kill anticipati, output XML corrotto o incompleto, e fallback elegante senza panic né memory leak.

### Priorità Bassa (P3 - Milestone 2 Roadmap & Discovery)
- [ ] **ONBOARD-06 · Spike Architetturale: Attack Path Mapping (Grafo Vulnerabilità)**
  - File: `docs/domains/attack_path.md`, `backend/internal/domain/`
  - Dipende da: Nessuna
  - Test: Documentazione di analisi tecnica e prototipo grafo
  - DoD: Modello dati per collegamenti tra vulnerabilità, calcolo del percorso di lateral movement e specifiche UI per grafo visuale interattivo.

- [ ] **ONBOARD-07 · Spike Architetturale: Continuous Compliance Score (NIS2 / DORA / ISO 27001)**
  - File: `docs/domains/compliance.md`, `backend/internal/domain/`
  - Dipende da: Nessuna
  - Test: Algoritmo di mapping controlli e scoring di conformità
  - DoD: Mappatura requisiti normativi sui finding di sicurezza e definizione delle formule di allineamento percentuale per ciascun framework.
