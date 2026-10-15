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

## 🚀 Coda di Sviluppo Attiva (Priorità validate dal PO)

### Sprint Attivo · Epic 27: Gestione Aziende, Reparti & Filtro Globale (FEAT-COMPANY-SCOPE) [ACCETTATO DAL PO]

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
  - Docs: docs/PROJECT_STATE.md, docs/2_ARCHITECTURE.md
  - DoD: Build pulita (0 errori), tutti i test frontend e backend 100% VERDI.

### Sprint Attivo · Epic 27.1: Semplificazione Centro Scansioni & Scanning Studio 360° (FEAT-SCAN-STUDIO)

- [~] E27.1-S1-T1 · Riprogettazione UI di NewScanModal: 3-4 Master Presets Intelligenti e Pannello "Fine-Tuning a 360°"
  - File: `frontend/src/components/NewScanModal.tsx`, `frontend/src/components/NewScanModal.test.tsx`
  - Dipende da: E27-S1-T5
  - Test: test unitari Vitest su NewScanModal (selezione preset standard, switch a personalizzazione avanzata porte/sonde/timing, live preview reattiva, calcolo dinamico invasività e trigger gate) (14/14 PASS)
  - Docs: `docs/domains/scanning_and_discovery.md`, `docs/GUIDA_FUNZIONALE_PIATTAFORMA.md`
  - DoD: Interfaccia unificata senza 14 opzioni disperse: card a 1-clic per i casi d'uso standard + controlli granulari (porte, OS, versioni, vuln script, UDP, rate/timing) per libertà completa a 360°.

- [~] E27.1-S1-T2 · Allineamento Contratti e Documentazione Funzionale
  - File: `docs/GUIDA_FUNZIONALE_PIATTAFORMA.md`, `docs/domains/scanning_and_discovery.md`, `docs/2_ARCHITECTURE.md`
  - Dipende da: E27.1-S1-T1
  - Test: verifica consistenza documentale e mappa dei profili
  - Docs: `docs/PROJECT_STATE.md`, `docs/GUIDA_FUNZIONALE_PIATTAFORMA.md`
  - DoD: Documentazione allineata con spiegazione della nuova UX unificata e aggiornamento screenshot di NewScanModal.

- [~] E27.1-S1-T3 · Verifica Zero-Assumption e Suite Completa di Regressione
  - File: `frontend/src/**`, `backend/**`
  - Dipende da: E27.1-S1-T1, E27.1-S1-T2
  - Test: `npm test` (18 file, 107 test PASS), `npm run build` (tsc + vite PASS in 10.64s), `go test ./...` (14 pkg PASS)
  - Docs: `docs/PROJECT_STATE.md`
  - DoD: 100% test verdi, build TypeScript pulita con 0 errori.

### Sprint Archiviati (Accettati dal PO `[x]`)
> *Nota di archiviazione:* Tutti i task completati e convalidati (TASK-01..60, `FEAT-IP-ORG-01..04`, `FEAT-SCAN-PROF-01..04`, `FEAT-WORKFLOW-CLEANUP-01..07`, `FEAT-CVE-ONLINE-01`, `FEAT-REPORT-PDF-UX-01`, `FEAT-M2-01..03`, `ONBOARD-01..05`, `SEC-P0-01..06`, `SEC-P1-01..06`, `SEC-P2-01..03`) sono stati formalmente archiviati nel [4_CHANGELOG.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/4_CHANGELOG.md) in conformità alle regole di token & backlog management di `AGENTS.md`.

---

## 🔮 Prossimi Obiettivi & Roadmap Approvata dal PO

### Milestone 3 · Active Defense & SOAR Light (Epic 28 - In Partenza)

- [ ] E28-S1-T1 · Backend Chokepoint Auto-Remediation Service & SSH Firewall Rule Applicator
  - File: `backend/internal/domain/defense.go`, `backend/internal/service/defense_service.go`, `backend/internal/delivery/defense_handler.go`
  - Dipende da: Epic 26 (Attack Path), Epic 22 (SSH Manager)
  - Test: test unitari Go su generazione regole firewall da chokepoint e invocazione sicura SSH con rollback
  - Docs: `docs/domains/active_defense_and_soar.md`
  - DoD: Endpoint `POST /api/v1/attack-paths/chokepoints/{id}/remediate` operativo con isolamento o blocco porta e fallback sicuro.

- [ ] E28-S1-T2 · Frontend Attack Path UI: Azione "Neutralizza Chokepoint" con Safety Timer
  - File: `frontend/src/pages/AttackPath.tsx`, `frontend/src/components/ChokepointRemediationModal.tsx`
  - Dipende da: E28-S1-T1
  - Test: test Vitest su apertura modale di mitigazione, preview della regola e countdown di rollback
  - Docs: `docs/domains/active_defense_and_soar.md`
  - DoD: Pulsante interattivo nel drawer del chokepoint per applicare la regola firewall con 1-click e monitorare il safety timer.

- [ ] E28-S1-T3 · Virtual Patching & WAF Rules Generator (Fail2ban, CrowdSec, ModSecurity/Nginx)
  - File: `backend/internal/service/virtual_patching.go`, `frontend/src/components/VirtualPatchModal.tsx`
  - Dipende da: Epic 25 (CVE Finding models)
  - Test: test unitari Go su generazione regole regex/WAF e test Vitest su visualizzazione/copia
  - Docs: `docs/domains/active_defense_and_soar.md`
  - DoD: Generazione accurata di regole virtual patch per le principali CVE di rete e web con download/copia immediata.

- [ ] E28-S1-T4 · Multi-Channel Alerting Hub Backend (Slack, Teams, Telegram, Webhook REST, SMTP)
  - File: `backend/internal/domain/alert.go`, `backend/internal/service/alert_service.go`, `backend/internal/delivery/alert_handler.go`, migrazione SQL `000008`
  - Dipende da: nessuno
  - Test: test unitari Go su dispatcher notifiche e HMAC signing per webhooks
  - Docs: `docs/domains/active_defense_and_soar.md`
  - DoD: Notifiche asincrone scatenate alla scoperta di CVE critiche o variazioni host, con retry esponenziale.

- [ ] E28-S1-T5 · Gestione Canali Notifica Frontend in Impostazioni & Test Live
  - File: `frontend/src/pages/Settings.tsx`, `frontend/src/components/AlertChannelsCard.tsx`
  - Dipende da: E28-S1-T4
  - Test: test Vitest su configurazione canali e pulsante test invio
  - Docs: `docs/domains/active_defense_and_soar.md`
  - DoD: Interfaccia utente per aggiungere/modificare webhook e testare l'invio in tempo reale.

- [ ] E28-S1-T6 · Verifica Zero-Assumption e Test Suite Completa E2E
  - File: `backend/internal/**`, `frontend/src/**`
  - Dipende da: E28-S1-T1..E28-S1-T5
  - Test: `go test ./...`, `npm test`, `npm run build`
  - Docs: `docs/PROJECT_STATE.md`, `docs/2_ARCHITECTURE.md`
  - DoD: Tutti i test backend e frontend verdi al 100%, 0 warning di linting o typecheck.

---

### Milestone 4 · Continuous Compliance & Regulatory Frameworks (Epic 29 - Pianificata)
- [ ] **E29-S1-T1..T6**: Automated Compliance Mapping (NIS2 Art. 21, DORA, ISO/IEC 27001:2022), Compliance Scorecard per Reparto/Azienda, SLA MTTR Tracking, Generatore Dossier Formale Audit-Ready PDF A4.

---

### Milestone 5 · AI Remediation Copilot & Conversational Assistant (Epic 30 - Pianificata)
- [ ] **E30-S1-T1..T6**: AI Remediation Playbooks generator (script Bash/Ansible contestualizzati), AI Executive Summaries discorsivi nei Report, "Ask Infrastructure" assistente conversazionale con Zero Data Leakage.

---

## 🛠️ Tech Debt & Refactoring Backlog

Registro formale del debito tecnico e delle ottimizzazioni architetturali (regola ADR-10):

| ID | Priorità | Componente / Percorso | Descrizione & Impatto | Piano di Rientro |
|---|---|---|---|---|
| `TECH-DEBT-01` | Media | `frontend/src/` | Ottimizzazione manual chunks in `vite.config.ts` per separare librerie pesanti (`@xterm`, `recharts`, `lucide-react`) e ridurre il bundle iniziale sotto 500kB. | Refactoring bundle-splitting in Milestone 3 |
| `TECH-DEBT-02` | Media | `backend/internal/repository/migrations` | Verifica sistematica e test automatico di rollback (`down.sql`) per tutte le migrazioni storiche da `000001` a `000007`. | Script CI dedicato `migrate -path ... down` in Docker test |
| `TECH-DEBT-03` | Bassa | `frontend/src/components/` | Allineamento e standardizzazione componenti modali con componente comune `ModalWrapper` per ridurre duplicazione del backdrop e focus trap. | Da accorpare durante lo sviluppo delle modali di Milestone 3 |
| `TECH-DEBT-04` | Bassa | `backend/internal/scanner/nmap.go` | Centralizzazione delle stringhe di flag Nmap in costanti tipizzate enum per evitare dispersioni di string literals. | Pulizia preliminare prima di Milestone 3 |

