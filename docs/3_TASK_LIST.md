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

### Sprint Archiviati (Accettati dal PO `[x]`)
> *Nota di archiviazione:* Tutti i task completati e convalidati (TASK-01..60, `FEAT-IP-ORG-01..04`, `FEAT-SCAN-PROF-01..04`, `FEAT-WORKFLOW-CLEANUP-01..07`, `FEAT-CVE-ONLINE-01`, `FEAT-REPORT-PDF-UX-01`, `FEAT-M2-01..03`, `ONBOARD-01..05`, `SEC-P0-01..06`, `SEC-P1-01..06`, `SEC-P2-01..03`) sono stati formalmente archiviati nel [4_CHANGELOG.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/4_CHANGELOG.md) in conformità alle regole di token & backlog management di `AGENTS.md`.

---

## 🔮 Prossimi Obiettivi (In attesa di indicazioni dal Product Owner)
- **Milestone 3 / Nuove Feature:** In attesa della definizione dei nuovi requisiti o delle prossime priorità strategiche da parte del Product Owner.
