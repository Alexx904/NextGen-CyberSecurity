# 3. Interactive Task List & Implementation Backlog

> **Stato:** ATTIVO (Priorità validate dal Product Owner)
> **Ultimo Aggiornamento:** 2026-10-17
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

- [x] DOCKER-CI-01 · Immagini Docker scaricabili da GitHub Actions
  - Accettazione PO: 2026-10-17, «OK». Prima esecuzione GitHub ancora da verificare; nessuna autorizzazione implicita a commit/push/deploy.
  - File: .github/workflows/ci.yml, compose.images.yml, backend/.dockerignore, frontend/.dockerignore, frontend/Dockerfile.
  - Dipende da: nessuno; riusa i tre job CI esistenti come gate.
  - Test: build immagini, Nmap/Nginx isolati, archivio load e Compose config; docs/diff. Prima esecuzione remota non eseguita.
  - Evidenze: entrambe le build, Nmap --version, nginx -t, docker save/load compresso, Compose config e docs/diff PASS exit 0; dettaglio in CI_CD_PIPELINE. Nessuna suite applicativa locale rieseguita: sorgenti invariati, controlli applicativi conservati nel workflow.
  - Docs: CI_CD_PIPELINE, DEPLOYMENT e PROJECT_STATE; nessuna modifica ai domini/API/DB.
  - DoD: (1) verifiche locali PASS; (2) limiti CI remota dichiarati; (3) build immagini PASS; (4) Doc Sync e self-review.

- [x] SSH-TOOLS-01 · Terminale WebSocket e lettura iptables
  - Accettazione PO: 2026-10-17, «Ok perfetto è funzionante» dopo la diagnosi del 502. Nessuna chiusura implicita del gate watchdog/Linux.
  - File: frontend/nginx.conf, frontend/vite.config.ts, FirewallManager e test; backend domain/service/handler firewall e test.
  - Dipende da: SSH-CONNECT-02, accettato dal PO per la connessione.
  - Test: proxy WebSocket isolato, sudo non interattivo/root, errori prerequisiti/privilegi, visualizzazione API; gate finale All senza DB (nessun repository/schema modificato).
  - Docs: dominio SSH, API_SPEC, DOMAIN_MAP, DEPLOYMENT, [QA](../05-qa/SSH_TOOLS_QA.md) e PROJECT_STATE.
  - Evidenze: proxy Vite/Nginx bidirezionale PASS; gate Final All suite backend/vet/build, 175 test frontend/build e documenti/diff PASS exit 0. Nessun DB o target modificato.
  - Feedback successivo: PO riferisce ancora 502 generico. Check runtime 2026-10-17: bundle SSHServers-BeM5cCm3.js servito, diagnostica FIREWALL_PRIVILEGES_REQUIRED presente nel binario, proxy API → 401 JSON senza autenticazione (atteso). Causa target e cache browser non provate; richiesti verifica pacchetto/lettura sudo sul Debian e refresh forzato. Nessun nuovo fix o PASS reale dichiarato.
  - DoD: riproduzione rossa→verde, gate finale, Doc Sync e self-review; prova Debian reale al PO.

- [x] SSH-CONNECT-02 · Negoziazione host con più chiavi e pin SHA256
  - File: backend/internal/service/ssh_service.go, ssh_hardening_test.go.
  - Dipende da: nessuno.
  - Test: loopback multi-key Ed25519/RSA rosso→verde, ECDSA e pin sconosciuto senza autenticazione PASS; errore login non ritentato e cancellazione PASS. Gate Final Backend: suite 17,840s/vet 0,871s/build 1,470s, documenti/diff PASS exit 0.
  - Docs: ssh_and_firewall, piano SSH-HARDEN-01, PROJECT_STATE.
  - DoD: riproduzione rossa→verde, regressione backend, Doc Sync e self-review; connessione Debian reale da confermare dal PO.

- [!] SSH-CONNECT-01 · Diagnosi timeout TCP server 192.168.1.18:22
  - Scope: `/ssh-servers` non collega; PO riferisce accesso da CMD con `ssh alex@192.168.1.18` e fingerprint già fornito.
  - File: nessuna modifica applicativa; analizzati service SSH, handler, client frontend e rete Docker.
  - Dipende da: confronto contemporaneo dal CMD del PO sul PC Docker.
  - Test: TCP Windows timeout dopo 5s e 12s; `docker exec nextgen-backend sh -c 'nc -z -v -w 5 192.168.1.18 22'` timeout; ping host irraggiungibile, neighbor Ethernet Unreachable/MAC nullo. `ssh -G` conferma hostname/porta, nessun proxy riportato.
  - Docs: PROJECT_STATE; contratto SSH esistente invariato.
  - DoD: raggiungibilità dal backend e test connessione applicativo da verificare dopo riconciliazione delle evidenze; nessun fix o PASS funzionale dichiarato.

- [ ] SSH-HARDEN-01 · Pinning host SSH e deadline connessioni/comandi
  - Scope PO: avvio TECH-DEBT-06/07, 2026-10-17 “Procedi”; primo incremento implementato, nessuna chiusura implicita del watchdog o degli altri debiti.
  - File: domain/service SSH, repository/handler, migrazione 000010 up/down, tipi e modale SSH e test.
  - Dipende da: nessuno; modifiche DISCOVERY-01 preesistenti preservate.
  - Test: pin assente rosso→verde; mismatch/pin fidato, rotazione/reset endpoint, cancellazione handshake/comando bloccato, persistenza/tenant e up/down/up PostgreSQL PASS. Gate finale backend suite/vet/build, frontend 166 test/build e DB PASS exit 0.
  - Docs: [piano e QA](plans/SSH-HARDEN-01.md), ssh_and_firewall, REQUIREMENTS, UI_SPECIFICATIONS, DATABASE_SCHEMA, ADR-20, DEPLOYMENT, PROJECT_STATE.
  - DoD: (1) mirati PASS; (2) PostgreSQL e rollback PASS, visuale NON ESEGUITO; (3) suite/build/vet PASS; (4) Doc Sync e self-review completati.
  - Gate aperto: browser runtime vuoto e iab indisponibile, task non promosso a `[~]`. Nessun collaudo Linux/azione remota; TECH-DEBT-07 ed E28-S1-T1 restano aperti.

- [x] DISCOVERY-01 · Precisione identità e probe HTTP Network Discovery
  - Accettazione PO: 2026-10-17, “OK”; limiti dichiarati preservati, nessuna autorizzazione implicita a commit/push/deploy.
  - File: backend/internal/scanner/{arp_cache,http_grabber,vendor_infer} e test, backend/internal/service/discovery_service.go e test.
  - Dipende da: nessuno.
  - Test: MAC invalidi/software rosso→verde; cache cambio MAC, /31-/32, HTTP redirect/titolo, suite backend/vet/build.
  - Evidenze: backend `go test ./...` 29,632s, `go vet ./...` 3,237s, `go build ./...` 1,376s PASS exit 0; documenti/diff PASS. Nessun collaudo LAN reale; API/DB/UI invariati.
  - Docs: [analisi](NETWORK_DISCOVERY_REVIEW.md), dominio scanning_and_discovery, PROJECT_STATE.
  - DoD: test mirati e regressione PASS; limite LAN reale dichiarato; suite/build/vet PASS; Doc Sync e self-review.

Le attività concluse e la proposta Epic 27.1 superata sono archiviate nel [CHANGELOG](../../CHANGELOG.md), con ID, evidenze e stato di accettazione preservati. SSH-HARDEN-01 resta al gate visuale, senza promozione a `[~]`.

### Sprint Archiviati (Accettati dal PO `[x]`)
> *Nota di archiviazione:* Tutti i task completati e convalidati (TASK-01..60, `FEAT-IP-ORG-01..04`, `FEAT-SCAN-PROF-01..04`, `FEAT-WORKFLOW-CLEANUP-01..07`, `FEAT-CVE-ONLINE-01`, `FEAT-REPORT-PDF-UX-01`, `FEAT-M2-01..03`, `ONBOARD-01..05`, `SEC-P0-01..06`, `SEC-P1-01..06`, `SEC-P2-01..03`) sono stati formalmente archiviati nel [CHANGELOG.md](../../CHANGELOG.md) in conformità alle regole di token & backlog management di `AGENTS.md`.

---

## 🔮 Prossimi Obiettivi & Roadmap Approvata dal PO

### TEST-OPT-01 · Ottimizzazione della verifica software

- [x] TEST-OPT-01 · Test proporzionati, integrazioni esplicite e durate misurate
  - Accettazione PO: 2026-10-17, “Approvo tutto”, commit e push su main esplicitamente autorizzati.
  - File: AGENTS.md, playbook feature/feedback/fix/refactor/verify, scripts/verify.ps1, test PostgreSQL, .github/workflows/ci.yml.
  - Dipende da: nessuno; modifiche applicative preesistenti preservate.
  - Test: percorso rapido, backend suite/build/vet, integrazione reale e configurazione (12 casi) PASS exit 0; DB/config assenti FAIL atteso exit 1 senza skip; runner happy/rejection/fail-fast/report PASS; frontend smoke 30 test PASS. Documenti/diff PASS; CI GitHub non eseguita. Container effimero arrestato/rimosso, container applicativi preservati.
  - Docs: Test Plan, CI/CD, schema DB (contratto test), ADR-19, registry e PROJECT_STATE.
  - DoD: (1) verifiche mirate PASS; (2) integration reale PASS, CI remota non eseguita dichiarata; (3) backend suite/build/vet e runner PASS; (4) Doc Sync e self-review completati.

### COMPANY-UX-02 · Aziende/reparti e tool indipendenti

- [x] COMPANY-PDF-03 · CVE nei PDF aziendali separate per reparto
  - Accettazione PO: 2026-10-17, “Approvo tutto”. Limite visuale dichiarato accettato, nessun PASS render inventato.
  - File: service report/departments e test, pkg/report/pdf_generator.go e test.
  - Dipende da: COMPANY-UX-02; preserva modifiche AP/TEST-OPT preesistenti.
  - Test: mirati service/generatore PASS; gate finale backend 2026-10-17 suite (16,296 s), vet (1,034 s), build (1,340 s), documenti/diff PASS exit 0. PDF contenuto/paginazione PASS, visuale da verificare.
  - Docs: reporting_and_assets, REQUIREMENTS, UI_SPECIFICATIONS, piano COMPANY-UX-02, COMPANY_WORKFLOW_QA, PROJECT_STATE.
  - DoD: attribuzione target/scan corretta; schede per reparto nei due modelli, riepilogo consolidato; suite/build/vet e Doc Sync; PDF renderizzato.
  - Limite residuo: renderer PDF indisponibile, nessun PASS visivo né PDF autonomo consegnato. Self-review e Doc Sync completati. PO autorizza pubblicazione su main con questo limite.

- [x] COMPANY-UX-02 · Dashboard unica, gerarchia due livelli, report per reparto e clienti SSH
  - Accettazione PO: 2026-10-17, “Approvo tutto”; limiti visuali dichiarati accettati per pubblicazione.
  - File: App/layout/CompanyRoute, Dashboard/Targets/GroupModal/TreeExplorer, CompanyContext, AttackPath e guida, SSH modal/page/types; service gruppi/report/SSH, generatori PDF/CSV, repository SSH/gruppi, migrazioni 000008/000009 e test.
  - Dipende da: nessuno; integra AP-REVIEW-01 senza accettazione implicita.
  - Test: frontend 164 test/build PASS; backend suite/build/vet su DB dedicato PASS; up/down e gerarchia reale PostgreSQL PASS, tutti exit 0. Test PDF contenuto/paginazione PASS, render visivo non eseguito.
  - Docs: [piano](plans/COMPANY-UX-02.md), [dominio](../domains/reporting_and_assets.md), [QA](../05-qa/COMPANY_WORKFLOW_QA.md).
  - DoD: (1) mirati PASS; (2) collaudo browser desktop/mobile e PDF renderizzato; (3) suite/build/vet e migrazioni up/down PASS; (4) Doc Sync e self-review.
  - Limite residuo: browser e renderer PDF indisponibili; nessun PASS visuale dichiarato. Schema/dati applicativi invariati; migrazioni provate solo su test DB. Accettazione PO distinta da prova visuale autonoma.

### AP-REVIEW-01 · Revisione cyber e usabilità Attack Path

- [x] AP-REVIEW-01 · Correttezza del modello, scope, grafo e guida integrata
  - Accettazione PO: 2026-10-17, “Approvo tutto”; pubblicazione esplicita su main, limite browser accettato.
  - File: domain/service/handler Attack Path e test, repository target/finding (ordine paginazione), types/page e `frontend/src/components/graph/`.
  - Dipende da: nessuno; checkpoint `b16264b` pubblicato prima dell'intervento su richiesta PO.
  - Test: regressione backend suite/build/vet e frontend 156 test/build PASS exit 0; test rosso→verde su cammini/esposizione/protocolli; snapshot locale 17 target/20 finding senza CVE, servizio validato a soglie 0/7/9. Nessuna mutazione DB.
  - Docs: [piano](plans/AP-REVIEW-01.md), [dominio](../domains/attack_path.md), [report QA](../05-qa/ATTACK_PATH_REVIEW.md).
  - DoD: (1) test mirati PASS; (2) prova browser desktop/mobile con screenshot; (3) suite/build/vet PASS; (4) Doc Sync e self-review.
  - Limite residuo: browser runtime indisponibile (inventario vuoto, iab/Chrome non disponibili); nessuna validazione visuale autonoma o exploit/reachability reale dichiarata. Accettazione PO con limiti espliciti, non certificazione di sicurezza.

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
  - File: `backend/internal/domain/alert.go`, `backend/internal/service/alert_service.go`, `backend/internal/delivery/alert_handler.go`, prossima migrazione SQL libera da `000010`
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

DOCKER-CI-01: `npm ci` nella build segnala 16 vulnerabilità (7 moderate, 6 high, 3 critical). Il conteggio è output npm, non una verifica di sfruttabilità; nessun aggiornamento automatico applicato.

- [ ] TECH-DEBT-09 · Alta · Dipendenze frontend: analizzare advisory e distinguere dipendenze runtime/build, aggiornare in incremento dedicato con test/build e verifica compatibilità. Impatto: vulnerabilità segnalate dal resolver npm durante la build Docker; dettaglio advisory ancora da valutare.

Registro formale del debito tecnico e delle ottimizzazioni architetturali (regola ADR-10):

| ID | Priorità | Componente / Percorso | Descrizione & Impatto | Piano di Rientro |
|---|---|---|---|---|
| `TECH-DEBT-08` | Media | Attack Path service, scope UI | Analisi euristica senza routing/ACL/identità/prerequisiti; cap globale prima dello scope, osservazioni storiche e letture non transazionali possono sottorappresentare un'azienda. | Mantenere avvisi; futuro contratto di evidenze di raggiungibilità e scope backend prima del cap, snapshot coerente e ciclo di vita finding. Validare in laboratorio autorizzato prima di dichiarare lateral movement reale. |
| `TECH-DEBT-06` | Alta | `backend/internal/service/ssh_service.go` | SSH-HARDEN-01 implementa pin host SHA256 separato dalla chiave client: assente/mismatch bloccano connessioni. Migrazione 000010 e test rotazione/up/down/up PASS; gate visuale aperto. | Collaudare il form e verificare i pin legacy tramite canale fidato; chiusura formale dopo gate residuo e accettazione. |
| `TECH-DEBT-07` | Alta | `backend/internal/service/firewall_service.go`, `ssh_service.go` | SSH-HARDEN-01 limita handshake a 10s e comandi/test a 30s con cancellazione. Restano possibile esecuzione remota tardiva, transazione snapshot/watchdog/applicazione separata e timer locali fra istanze; nessuna prova Linux reale. | Rendere la transazione protetta sull'host, verificare terminazione e ripristino sotto disconnessione/timeout; laboratorio Linux e test multiistanza prima di chiudere E28-S1-T1. Le deadline locali non chiudono il debito. |
| `TECH-DEBT-05` | Alta | `backend/internal/delivery/http/handler/scan_handler.go`, scheduler, audit | FR-01.3 incompleto: consenso solo UI, non attestato immutabilmente né imposto dall'API; classificazione impatto euristica. | Definire contratto di attestazione per scope/configurazione, validazione server e registrazione audit; coprire API diretta e pianificazioni con test. Nessuna attestazione di conformità finché il gap resta. |
| `TECH-DEBT-01` | Media | `frontend/src/` | Ottimizzazione manual chunks in `vite.config.ts` per separare librerie pesanti (`@xterm`, `recharts`, `lucide-react`) e ridurre il bundle iniziale sotto 500kB. | Refactoring bundle-splitting in Milestone 3 |
| `TECH-DEBT-02` | Media | `backend/internal/repository/migrations` | Verifica sistematica e test automatico di rollback (`down.sql`) per tutte le migrazioni storiche da `000001` a `000007`. | Script CI dedicato `migrate -path ... down` in Docker test |
| `TECH-DEBT-03` | Bassa | `frontend/src/components/` | Allineamento e standardizzazione componenti modali con componente comune `ModalWrapper` per ridurre duplicazione del backdrop e focus trap. | Da accorpare durante lo sviluppo delle modali di Milestone 3 |
| `TECH-DEBT-04` | Bassa | `backend/internal/scanner/nmap.go` | Centralizzazione delle stringhe di flag Nmap in costanti tipizzate enum per evitare dispersioni di string literals. | Pulizia preliminare prima di Milestone 3 |
