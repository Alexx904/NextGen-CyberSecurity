# 3. Interactive Task List & Implementation Backlog

> **Updated:** 2026-10-19 - working tree
> **Related:** [Requirements](../01-product/REQUIREMENTS.md), [PRD](../01-product/PRD.md), [HLD](../03-architecture/HLD.md), [accepted tasks](../../CHANGELOG.md#accepted-tasks-and-priority-reorganization--2026-10-18).

## Execution priorities

PO request: order by risk and dependencies. Previous pending acceptance tasks are now accepted and archived in CHANGELOG. States and gates follow [AGENTS sections 5-10](../../AGENTS.md). IDs remain stable when order changes.

| Order | Scope | Reason / dependency |
|---|---|---|
| P0 | TECH-DEBT-05 and TECH-DEBT-09 | Enforce server-side consent and assess dependency advisories before new features. |
| P1 | E29-S1-T1..T6 (M4) | Resolve consent/audit gaps before claiming compliance. |
| P2 | E30-S1-T1..T6 (M5) | Copilot builds on inventory, findings and reporting. |
| P3 | E31-S1-T1 -> E32-S1-T1 -> E33-S1-T1 (M6-M8) | Catalog, controlled validation, then authorized testing; approvals remain mandatory. |
| P4 | Remaining medium/low debt | Order by measured impact; resolve feature prerequisites before that feature. |
| Final | E28-S1-T4 -> E28-S1-T5 -> E28-S1-T6 | Notifications deferred by PO until project end; full Epic 28 gate depends on both. |

## P0 - Security prerequisites

- [x] TECH-DEBT-05 · Mandatory server-side inventory scan consent, scope/configuration-bound recurring authorization and immutable audit. Backend/frontend/PostgreSQL/up-down/browser PASS; accepted by PO on 2026-10-19. [Contract and evidence](plans/TECH-DEBT-05.md).

TECH-DEBT-05 retains its detailed contract in the debt register below.

2026-10-19: PO selected consent for every inventory scan, recurring consent bound
to authorized scope/configuration. TECH-DEBT-05 implementation covers services,
API, scheduler, worker, immutable transactional audit and UI; Final All/Integration and final backend gates PASS; accepted by PO on 2026-10-19.
[Design and evidence](plans/TECH-DEBT-05.md). Browser is now connected; real QA
checks run on updated sources at 3000/8081 with disposable PostgreSQL on 55432,
separate from the existing deployment at 3080. No scans or external sends performed.

- [x] TECH-DEBT-09 · Alta · Dipendenze frontend: analizzare advisory e distinguere dipendenze runtime/build, aggiornare in incremento dedicato con test/build e verifica compatibilità. Impatto: vulnerabilità segnalate dal resolver npm durante la build Docker; dettaglio advisory ancora da valutare.
  - 2026-10-18: Router/Vite/Vitest/Tailwind 4 updated; official PostCSS plugin authorized by PO. Full and runtime npm audits: zero vulnerabilities; frontend Final PASS (229 tests + typecheck/build/docs). Real-browser visual gate NOT EXECUTED: no browser connected, task remains `[ ]`. Assessment, checks and limits: [dependency report](../05-qa/TECH-DEBT-09.md).

  - 2026-10-19: connected Chromium desktop/390px compatibility and keyboard checks PASS on updated QA sources; Final integrated frontend 233 tests/build PASS. Original browser blocker resolved; accepted by PO on 2026-10-19. [Completion](../05-qa/TECH-DEBT-09.md#browser-completion---2026-10-19).

## P1 - M4 - Compliance

- [ ] **E29-S1-T1..T6**: Automated Compliance Mapping (NIS2 Art. 21, DORA, ISO/IEC 27001:2022), Compliance Scorecard per Reparto/Azienda, SLA MTTR Tracking, Generatore Dossier Formale Audit-Ready PDF A4.

## P2 - M5 - Copilot

- [ ] **E30-S1-T1..T6**: AI Remediation Playbooks generator (script Bash/Ansible contestualizzati), AI Executive Summaries discorsivi nei Report, "Ask Infrastructure" assistente conversazionale con Zero Data Leakage.

## P3 - M6 - Exploit intelligence

- [ ] E31-S1-T1 · Catalogo exploit/moduli e valutazione applicabilità
  - File previsti: backend catalogo/correlazione, dettaglio asset/finding e reporting; percorsi definitivi dopo design e aggiornamento DOMAIN_MAP.
  - Dipende da: inventario/scansioni/CVE; revisione Metasploit Framework con versione, licenze componenti, supply-chain e aggiornamenti prima di approvare dipendenze.
  - Scope: riferimenti pubblici, moduli, prerequisiti, provenienza/data, KEV e motivazione compatibilità; nessuna esecuzione o download automatico di PoC. Dati mancanti e cataloghi obsoleti espliciti.
  - Test: matching/falsi positivi, versioni/backport sconosciuti, fonti assenti/obsolete, isolamento tenant, report e browser; gate Final per lati coinvolti, PostgreSQL e migrazioni up/down se pertinenti.
  - Docs: PRD, FR-22, UC-21, futuro dominio/DOMAIN_MAP, API/schema dopo design, Test Plan.
  - DoD: correlazione distinta da validazione, fonti verificabili, contratti prima del codice, gate PASS, self-review/security-review e Doc Sync; accettazione PO separata.

## P3 - M7 - Controlled validation

- [ ] E32-S1-T1 · Worker RPC e validazione controllata
  - File previsti: backend integrazione/worker, autorizzazioni/job/evidenze, UI validazione, deployment e migrazioni se necessarie; percorsi dopo design.
  - Dipende da: E31-S1-T1, approvazione dipendenze/architettura e laboratorio dedicato. API/DTO/errori/auth/tenant/schema definiti prima del codice.
  - Scope: worker separato con RPC protetta accessibile dal backend, controlli selezionati/revisionati, RBAC, perimetro/tecniche/finestra, timeout/cancellazione, audit ed evidenze redatte. Workspace non sufficienti per isolamento. Disponibilità e impatto di `check` valutati per modulo.
  - Test: accessi cross-tenant, destinazioni fuori scope anche dopo risoluzione, autorizzazioni scadute, moduli non ammessi, timeout/disconnessioni/cancellazione, segreti nei log, concorrenza e laboratorio; E2E browser, Final All e PostgreSQL/up-down se pertinenti.
  - Docs: PRD, FR-23, UC-22, dominio/DOMAIN_MAP, ADR architetturale prima del codice, API/schema, deployment e Test Plan.
  - DoD: accessi fuori scope negati, esiti inconcludenti espliciti, nessuna esecuzione automatica dalla CVE; laboratorio e gate PASS, security-review/Doc Sync, accettazione PO separata.

## P3 - M8 - Authorized testing and retest

- [ ] E33-S1-T1 · Prove autorizzate, sessioni, report e retest
  - File previsti: backend orchestrazione/job/sessioni/evidenze, UI pentest/reporting e migrazioni se necessarie; percorsi dopo design.
  - Dipende da: E32-S1-T1, laboratorio isolato, moduli revisionati e regole d'ingaggio con strategia stop/cleanup.
  - Scope: prove selezionate e confermate dall'operatore nel perimetro autorizzato, limiti di concorrenza, monitoraggio/arresto, cleanup con effetti residui documentati, report e retest in finestra valida. Prima versione senza campagne autonome, persistenza, raccolta indiscriminata, lateral movement automatico o DoS.
  - Test: rifiuto scope/ruoli/finestra errati, arresto in esecuzione, sessioni residue, riavvio worker/backend, audit/redazione, cleanup e retest in laboratorio; E2E browser/PDF, Final All e PostgreSQL/up-down se pertinenti.
  - Docs: PRD, FR-24, UC-23, dominio dedicato, API/schema, procedure operative e Test Plan con matrice effetti/cleanup.
  - DoD: impatto osservato distinto da ipotesi, prova fallita distinta da sicurezza, cancellazione distinta da rollback; laboratorio/gate PASS, security-review/Doc Sync, accettazione PO separata. Autorizzazioni reali da acquisire per ogni perimetro.

## P4 - Technical debt

- [ ] UI-TOOLS-01 · Introduce a post-login tool selection hub and contextual navigation
  - PO request (2026-10-18): after login, show a tool selection screen instead of entering the current unified operational interface. Planning only; implementation and final UI design remain pending.
  - Tool 1: keep Network Discovery, Nmap scans, CVE findings, reports and Attack Path Graph together as one connected assessment workflow. Final tool name to be defined during design.
  - Tool 2: group SSH Servers and Firewall management in a separate operational tool. Define navigation within each tool and an accessible return to the tool selection hub; preserve existing capabilities and data.
  - Future feature placement: at design time for every upcoming feature/milestone, explicitly decide whether it extends an existing tool or warrants an independent tool, based on user goal, workflow and shared data. Record the rationale in the relevant task/design; do not automatically create one tool per milestone. Identify shared settings and cross-tool entry points during design.
  - Files: frontend authentication redirect/router, `frontend/src/components/layout/AppLayout.tsx`, new tool hub page and affected navigation/pages/tests; definitive paths after route inventory. No backend/schema changes assumed.
  - Dependencies: inspect current auth, routes, roles and shared company/target context; assess ADR-18 and document any supersession before implementation. Preserve existing priority queue; consider placement when designing upcoming milestones.
  - Tests: post-login hub, both tool entry flows, return/switch navigation, direct authenticated links, refresh/back/logout, unauthorized access and context preservation; keyboard/focus and responsive desktop/mobile browser checks. Targeted routing regressions, Final Frontend; extend gate scope if shared backend/auth contracts change.
  - Docs: PRD/requirements, UI_SPECIFICATIONS, relevant domains/DOMAIN_MAP, ADR for navigation architecture where warranted and PROJECT_STATE. Future milestone plans must include tool placement and rationale.
  - DoD: two clearly distinguished tools accessible after login, coherent navigation without lost features or broken deep links, shared concerns placed explicitly, required gates PASS, self-review and Doc Sync; PO validates the designed experience and implementation separately.

- [ ] E28-S1-T7 · Replace the virtual patch generator with online mitigation sources
  - PO decision (2026-10-18): plan the deprecation of the eight-CVE generator in favor of "Mitigazioni e fonti" in finding details. Planning only; the current feature remains unchanged until implementation.
  - Scope: consolidate existing NVD/MITRE references; link to vendor advisories through CVE references and offer an explicitly labeled CrowdSec Hub search using the CVE ID. A search link is not evidence that a compatible rule exists. Verify official destination URLs when implementing.
  - Constraints: no new database tables, local template catalog expansion, periodic synchronization, automatic external content retrieval, rule generation or installation. No new API dependency; links disclose only the public CVE ID, never asset or tenant information.
  - Files: `frontend/src/components/FindingDetailModal.tsx`, `VirtualPatchModal.tsx`, `frontend/src/api/virtualPatching.ts`; backend virtual-patching service/catalog, handler, route and tests after usage/compatibility inspection. Retire obsolete code coherently and document endpoint compatibility before removal.
  - Dependencies: accepted E28-S1-T3 baseline; inspect shared finding details and current API consumers. Queue with P4 maintenance; no change to the PO deferral of T4/T5.
  - Tests: valid/missing/invalid CVE IDs, safe external URLs and new-tab behavior, absence of duplicate links, accessible desktop/mobile finding flow; targeted regressions and Final Frontend, plus Final Backend if backend code changes. Real-browser verification required for the changed UI.
  - Docs: PRD, REQUIREMENTS, UI_SPECIFICATIONS, active_defense_and_soar domain, DOMAIN_MAP/API documentation where contracts change, E28-S1-T3 plan and PROJECT_STATE; preserve the accepted historical task and mark its behavior superseded only upon implementation.
  - DoD: useful external references for any valid CVE without claiming mitigation coverage or automatic correction; generator entry point replaced, obsolete paths handled explicitly, required gates PASS, self-review/security review and Doc Sync complete; PO acceptance separate.

- [ ] TECH-DEBT-10 · Media · Archivio Report: caricamento e memoria proporzionali allo storico del tenant; paginazione offset senza snapshot transazionale. Misurare sui volumi reali, poi introdurre filtro profilo/aggregazioni server e ordinamento stabile con integrazione PostgreSQL. ADR-22; nessun limite silenzioso accettato.

| ID | Priorità | Componente / Percorso | Descrizione & Impatto | Piano di Rientro |
|---|---|---|---|---|
| `TECH-DEBT-05` | High | Scan API, services, scheduler, worker, PostgreSQL audit and UI | Mandatory server consent implemented; impact classification remains heuristic. | Gates PASS; accepted by PO on 2026-10-19. Legacy schedules require suspension and explicitly authorized replacement. [Contract/limits](plans/TECH-DEBT-05.md). |
| `TECH-DEBT-06` | Alta | `backend/internal/service/ssh_service.go` | SSH-HARDEN-01 implementa pin host SHA256 separato dalla chiave client: assente/mismatch bloccano connessioni. Migrazione 000010 e test rotazione/up/down/up PASS; visual gate PASS; accepted with SSH-HARDEN-01. | Form verified; independent legacy-pin verification remains an operational prerequisite. |
| `TECH-DEBT-07` | Alta · mitigato, attende PO | `backend/internal/service/firewall_service.go`, `ssh_service.go` | M11-FIREWALL-01: transazione/lock/watchdog remoto senza kill PID, stdin e revisioni; Linux nft/legacy, multiistanza, OpenSSH con modifica tardiva e browser PASS. | Mitigation accepted with M11-FIREWALL-01; residual limits: stesso UID SSH, disponibilità host/tool/stato temporaneo; reboot, altri amministratori/account e cancellazione dello stato fuori garanzia. E28-S1-T1 richiede comunque prova del proprio flusso remediation. |
| `TECH-DEBT-08` | Media | Attack Path service, scope UI | Analisi euristica senza routing/ACL/identità/prerequisiti; cap globale prima dello scope, osservazioni storiche e letture non transazionali possono sottorappresentare un'azienda. | Mantenere avvisi; futuro contratto di evidenze di raggiungibilità e scope backend prima del cap, snapshot coerente e ciclo di vita finding. Validare in laboratorio autorizzato prima di dichiarare lateral movement reale. |
| `TECH-DEBT-02` | Media | `backend/internal/repository/migrations` | Verifica sistematica e test automatico di rollback (`down.sql`) per tutte le migrazioni storiche da `000001` a `000007`. | Script CI dedicato `migrate -path ... down` in Docker test |
| `TECH-DEBT-01` | Media | `frontend/src/` | Ottimizzazione manual chunks in `vite.config.ts` per separare librerie pesanti (`@xterm`, `recharts`, `lucide-react`) e ridurre il bundle iniziale sotto 500kB. | Refactoring bundle-splitting in Milestone 3 |
| `TECH-DEBT-03` | Bassa | `frontend/src/components/` | Allineamento e standardizzazione componenti modali con componente comune `ModalWrapper` per ridurre duplicazione del backdrop e focus trap. | Da accorpare durante lo sviluppo delle modali di Milestone 3 |
| `TECH-DEBT-04` | Bassa | `backend/internal/scanner/nmap.go` | Centralizzazione delle stringhe di flag Nmap in costanti tipizzate enum per evitare dispersioni di string literals. | Pulizia preliminare prima di Milestone 3 |

## Final - Deferred notifications and Epic 28 gate

- [ ] E28-S1-T4 · Multi-Channel Alerting Hub Backend (Slack, Teams, Telegram, Webhook REST, SMTP)
  - PO decision 2026-10-18: deferred until the end of the project. Stop implementation, deployment and live tests; preserve prepared work. Not completed or accepted.
  - Avanzamento 2026-10-18: implementati canali cifrati admin, outbox transazionale/lease/retry, hook finding/discovery/scansioni programmate e trasporti con egress/TLS. Mirati unitari e PostgreSQL/up-down PASS; gate integrato e browser T5 aperti. [Piano](plans/E28-S1-T4-alerting.md). Nessun provider esterno contattato; certificati TLS senza modello osservato non ancora supportati.
  - File: `backend/internal/domain/alert.go`, `backend/internal/service/alert*`, `backend/internal/delivery/http/handler/alert_handler.go`, migrazione `000012_alerting`
  - Dipende da: nessuno
  - Test: test unitari Go su dispatcher notifiche e HMAC signing per webhooks
  - Docs: `docs/domains/active_defense_and_soar.md`
  - DoD: Notifiche asincrone scatenate alla scoperta di CVE critiche o variazioni host, con retry esponenziale.

- [ ] E28-S1-T5 · Gestione Canali Notifica Frontend in Impostazioni & Test Live
  - PO decision 2026-10-18: deferred until the end of the project with T4. Stop implementation, deployment and live tests; preserve prepared work. Not completed or accepted.
  - Avanzamento 2026-10-18: card Settings con scope/eventi, configurazione write-only, test previo consenso e rimozione esplicita; sei mirati iniziali PASS. Fix redazione config webhook rosso→verde, 45 test diagnostici PASS. [Piano](plans/E28-S1-T5-alert-settings.md); gate integrato/browser aperti.
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

## Historical diagnosis

- [!] SSH-CONNECT-01 · Diagnosi timeout TCP server 192.168.1.18:22
  - Scope: `/ssh-servers` non collega; PO riferisce accesso da CMD con `ssh alex@192.168.1.18` e fingerprint già fornito.
  - File: nessuna modifica applicativa; analizzati service SSH, handler, client frontend e rete Docker.
  - Dipende da: confronto contemporaneo dal CMD del PO sul PC Docker.
  - Test: TCP Windows timeout dopo 5s e 12s; `docker exec nextgen-backend sh -c 'nc -z -v -w 5 192.168.1.18 22'` timeout; ping host irraggiungibile, neighbor Ethernet Unreachable/MAC nullo. `ssh -G` conferma hostname/porta, nessun proxy riportato.
  - Docs: PROJECT_STATE; contratto SSH esistente invariato.
  - DoD: raggiungibilità dal backend e test connessione applicativo da verificare dopo riconciliazione delle evidenze; nessun fix o PASS funzionale dichiarato.

Superseded by accepted SSH connection work; historical evidence retained, no current queue blocker.
