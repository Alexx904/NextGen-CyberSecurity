# 3. Interactive Task List & Implementation Backlog

> **Stato:** ATTIVO (Priorità validate dal Product Owner)
> **Ultimo Aggiornamento:** 2026-10-16
> **Autore / Ruolo:** Codex Tech Lead (System Architect & QA)
> **Documenti Correlati:** [REQUIREMENTS.md](../01-product/REQUIREMENTS.md) | [PRD.md](../01-product/PRD.md) | [HLD.md](../03-architecture/HLD.md) | [CHANGELOG.md](../../CHANGELOG.md)

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
5. **Archiviazione:** I task completati e accettati nelle Milestone 1, 1.1 e 1.2 (TASK-01 fino a TASK-60) sono archiviati in [CHANGELOG.md](../../CHANGELOG.md). Oltre 10 task `[x]`, spostare in archivio.

---

## 🚀 Coda di Sviluppo Attiva (Priorità validate dal PO)

Le attività concluse e la proposta Epic 27.1 superata sono archiviate nel [CHANGELOG](../../CHANGELOG.md), con ID, evidenze e stato di accettazione preservati. Nessuna attività in attesa di approvazione.

### Sprint Archiviati (Accettati dal PO `[x]`)
> *Nota di archiviazione:* Tutti i task completati e convalidati (TASK-01..60, `FEAT-IP-ORG-01..04`, `FEAT-SCAN-PROF-01..04`, `FEAT-WORKFLOW-CLEANUP-01..07`, `FEAT-CVE-ONLINE-01`, `FEAT-REPORT-PDF-UX-01`, `FEAT-M2-01..03`, `ONBOARD-01..05`, `SEC-P0-01..06`, `SEC-P1-01..06`, `SEC-P2-01..03`) sono stati formalmente archiviati nel [CHANGELOG.md](../../CHANGELOG.md) in conformità alle regole di token & backlog management di `AGENTS.md`.

---

## 🔮 Prossimi Obiettivi & Roadmap Approvata dal PO

### Milestone 3 · Active Defense & SOAR Light (Epic 28 - In Partenza)

- [ ] E28-S1-T1 · Backend Chokepoint Auto-Remediation Service & SSH Firewall Rule Applicator
  - Avanzamento 2026-10-17: implementati preview e blocco porta sul server del chokepoint, scelta confermata dal PO; isolamento host escluso. Suite backend/build/vet e test frontend PASS. Gate operativo aperto: prova Linux reale e limiti watchdog/SSH (TECH-DEBT-06/07). Non promosso a `[~]`.
  - Piano: [E28-S1-T1](plans/E28-S1-T1-chokepoint-remediation.md).
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
  - Docs: `docs/PROJECT_STATE.md`, `docs/03-architecture/HLD.md`
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
| `TECH-DEBT-06` | Alta | `backend/internal/service/ssh_service.go` | Host-key SSH non verificata (`InsecureIgnoreHostKey`); il fingerprint esistente riguarda la chiave client. Possibile impersonificazione del server, anche sul nuovo percorso remediation. | Progettare registrazione e pinning host-key con contratto dedicato, migrazione up/down e test mismatch/rotazione prima dell'uso operativo. |
| `TECH-DEBT-07` | Alta | `backend/internal/service/firewall_service.go`, `ssh_service.go` | Watchdog preesistente: chiamate remote senza deadline di esecuzione, possibile mutazione tardiva dopo scadenza e timer solo locali fra istanze; nessuna prova Linux reale in questa sessione. | Rendere la transazione snapshot/watchdog/applicazione protetta sull'host, verificare kill e ripristino sotto disconnessione/timeout; laboratorio Linux e test concorrenti multiistanza prima di chiudere E28-S1-T1. |
| `TECH-DEBT-05` | Alta | `backend/internal/delivery/http/handler/scan_handler.go`, scheduler, audit | FR-01.3 incompleto: consenso solo UI, non attestato immutabilmente né imposto dall'API; classificazione impatto euristica. | Definire contratto di attestazione per scope/configurazione, validazione server e registrazione audit; coprire API diretta e pianificazioni con test. Nessuna attestazione di conformità finché il gap resta. |
| `TECH-DEBT-01` | Media | `frontend/src/` | Ottimizzazione manual chunks in `vite.config.ts` per separare librerie pesanti (`@xterm`, `recharts`, `lucide-react`) e ridurre il bundle iniziale sotto 500kB. | Refactoring bundle-splitting in Milestone 3 |
| `TECH-DEBT-02` | Media | `backend/internal/repository/migrations` | Verifica sistematica e test automatico di rollback (`down.sql`) per tutte le migrazioni storiche da `000001` a `000007`. | Script CI dedicato `migrate -path ... down` in Docker test |
| `TECH-DEBT-03` | Bassa | `frontend/src/components/` | Allineamento e standardizzazione componenti modali con componente comune `ModalWrapper` per ridurre duplicazione del backdrop e focus trap. | Da accorpare durante lo sviluppo delle modali di Milestone 3 |
| `TECH-DEBT-04` | Bassa | `backend/internal/scanner/nmap.go` | Centralizzazione delle stringhe di flag Nmap in costanti tipizzate enum per evitare dispersioni di string literals. | Pulizia preliminare prima di Milestone 3 |
