# PROJECT STATE

> Memoria viva dell'agente. Letta a inizio sessione, aggiornata a fine di ogni turno.
> Non cancellare sezioni: se vuote, scrivi "nessuno".

**Ultimo aggiornamento:** 2026-10-17
**Modalità progetto:** esistente (onboarded formalmente sul branch chore/agent-onboarding)
**Fase:** Tutte le modifiche approvate pubblicate su origin/main il 2026-10-17, commit applicativo `26cf93d`. Copia locale su main, nessun nuovo branch o force push. TEST-OPT-01, COMPANY-UX-02, COMPANY-PDF-03 e AP-REVIEW-01 accettati con limiti visuali espliciti. E28-S1-T1 e debiti SSH/Linux restano aperti; nessun deploy.

## Project Profile

```
Nome progetto:           NextGen-CyberSecurity Platform
Linguaggi/stack:         Golang 1.22+ (Chi Router, pgx/v5, gorilla/websocket, robfig/cron, go-pdf/fpdf v0.9.0) + React 18 / TypeScript 5.6 (Vite, Tailwind CSS, xterm.js, Recharts)
Gestore pacchetti:       Go Modules (backend), npm (frontend)
Comando test mirati:     go test -v ./internal/<package>/... (backend), npx vitest run src/<path>/<file>.test.ts (frontend)
Comando suite completa:  go test ./... (backend), npm test (frontend)
Comando integrazione DB: go test -tags=integration -count=1 ./internal/repository/postgres (backend; TEST_DB_* esplicite, DB locale dedicato)
Runner verifica:        powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1 -Profile Quick|Final|Full|Docs -Scope Backend|Frontend|All
Comando build/typecheck: go build ./... (backend), npm run build (frontend)
Comando lint/format:     go vet ./... (backend)
Percorsi per livello:    unit=backend/pkg, backend/internal/domain, backend/internal/scanner; service=backend/internal/service; integration=backend/internal/delivery, backend/internal/repository/postgres; frontend=frontend/src/tests, frontend/src/**/*.test.tsx
Prerequisiti ambiente:   Docker & Docker Compose (PostgreSQL 16, Nmap su Linux container)
Come avviare l'app:      docker compose up --build (oppure 'go run ./cmd/server' + 'cd frontend && npm run dev')
Branch principale:       main (allineamento remoto non verificato durante la migrazione documentale)
Comandi vietati:         rm -rf fuori repo, git reset --hard non confermati, git push --force, DROP database senza conferma esplicita
```

## Baseline test (prima di ogni intervento)

La baseline è consultata prima degli interventi, non rieseguita automaticamente: politica proporzionata `AGENTS.md` sez. 8 e [Test Plan](05-qa/TEST_PLAN.md). Una sola suite finale per lato coinvolto; test mirati durante lo sviluppo e riuso solo con input/ambiente invariati. Il percorso rapido Go ora esclude PostgreSQL, che richiede tag `integration`; i PASS storici precedenti non vengono reinterpretati.

Verifica pre-push 2026-10-16: `npm.cmd test` PASS (21 file, 145 test), `npm.cmd run build` PASS (12,65 s); `go test ./...`, `go build ./...`, `go vet ./...` PASS, tutti exit 0. Prove funzionali delle scansioni eseguite e confermate dal PO. Fetch origin eseguito; branch inizialmente allineato a origin/main (0 commit di divergenza). Messaggio previsto: `feat(scanner): simplify scan workflows and add batch progress estimates`; destinazione origin/feat/nmap-scan-workflow-review.

Ultima ricognizione `/status` (2026-10-16): branch `feat/nmap-scan-workflow-review`, working tree con modifiche scanner/documenti non committate e nuovi file non tracciati. SCAN-COPY-01 accettato; SCAN-UX-01 e SCAN-TIME-01 implementati ma con gate visuale/laboratorio pendente. Ultime verifiche applicative del 2026-10-16: frontend 145 test e build PASS, backend suite/build/vet PASS. Nessuna suite applicativa rieseguita per questa richiesta di stato. La tabella del 2026-10-15 sotto è storica e non rappresenta la baseline più recente.

Verifica pre-push del 2026-10-15: `go test ./...`, `go build ./...`, `go vet ./...` PASS; frontend `npm.cmd test` PASS (18 file, 107 test), `npm.cmd run build` PASS (12,41 s). Verifica documentale PASS (24 documenti, 53 Markdown, 101 link/ancore). Controllo whitespace esteso al diff da `origin/main` PASS dopo normalizzazione. Comandi applicativi eseguiti con permessi estesi dopo blocchi locali della sandbox; nessuna modifica di logica applicativa introdotta nella preparazione.


| Data | Comando | Esito | Fallimenti preesistenti noti |
|---|---|---|---|
| 2026-10-15 17:13 | `go vet ./...` (backend) | PASS (exit code 0) | Nessuno. 0 warning, static analysis perfetta. |
| 2026-10-15 17:12 | `go test ./...` (backend) | PASS (exit code 0, 14 package) | Nessuno. Tutti i 14 package backend passano al 100%. |
| 2026-10-15 17:12 | `go build ./...` (backend) | PASS (exit code 0) | Nessuno. Compilazione binari server e cve-sync completata con successo. |
| 2026-10-15 17:13 | `npm run build` (frontend tsc + Vite) | PASS (exit code 0, 7.25s) | Nessuno. Compilazione TypeScript e bundle Vite ottimizzato, 0 errori. |
| 2026-10-15 17:12 | `npm test` (frontend Vitest) | PASS (exit code 0, 18 file, 107 test) | Nessuno. Tutti i 107 test passano al 100% con 0 fallimenti. |

## In corso

**Aggiornamento autorevole 2026-10-17:** “Ok puoi pushare tutto sul main (non voglio vedere branch) Approvo tutto.” Tutte le modifiche presenti, incluse quelle preesistenti, sono autorizzate per commit e push su `https://github.com/Alexx904/NextGen-CyberSecurity.git`, destinazione `main`. Nessun nuovo branch o force push. I resoconti sotto sono evidenze storiche degli incrementi; gli stati precedenti `[!]`/`[~]` sono superati dall'accettazione PO per i quattro task consegnati. Rischio comunicato una volta: prove visuali non effettuate e laboratorio SSH assente. L'accettazione con questi limiti non dichiara PASS visuale né chiude lavoro non implementato. Test applicativi invariati rispetto alle ultime evidenze, riutilizzati secondo AGENTS sez. 8; controlli documentali/staging rieseguiti.

- **COMPANY-PDF-03 (2026-10-17) `[!]`:** richiesta PO di separare anche le CVE per reparto nei PDF dell'intera azienda. Implementati schede tecniche complete per reparto e top cinque Executive locali con totale, KPI aziendali mantenuti, intestazioni di continuazione, target diretti e reparti vuoti distinti. Aziende senza reparti/singoli download conservano il flusso precedente. Raggruppamento per identità gruppo e scan/target, non IP/nome; test collisione etichetta “Senza reparto”. Gate finale backend runner PASS: suite 16,296 s, vet 1,034 s, build 1,340 s; documenti/diff PASS exit 0. Evidenza e Doc Sync nel [QA](05-qa/COMPANY_WORKFLOW_QA.md#incremento-company-pdf-03--cve-per-reparto). Assunzione: Executive resta sintesi; tutti i finding nel Tecnico. Renderer assente: contenuto/paginazione verificati, visuale aperto. Nessuna dipendenza/API/DB/frontend modificata, commit/push/deploy o mutazione dati. Modifiche TEST-OPT/COMPANY/AP preesistenti preservate.

- **TEST-OPT-01 (2026-10-17) `[~]`:** PO autorizza l'ottimizzazione proposta. Regole proporzionate, runner con durate/report, integrazioni PostgreSQL esplicite e CI con DB effimero. Backend suite/build/vet, DB reale e 12 casi configurazione PASS; frontend runner smoke 30 test PASS; FAIL attesi senza DB/config e fail-fast runner verificati. Doc Sync e self-review completati; nessuna modifica applicativa o nuova dipendenza. Le modifiche preesistenti COMPANY/AP sono preservate. Nessun commit/push/deploy; CI remota non eseguita.

- **COMPANY-UX-02 (2026-10-17):** PO richiede dashboard/aziende unica, reparti solo diretti, report per reparto/azienda e IP distinti nel documento, AP senza CVSS UI con selettore reparto, obbligo azienda per navigare. Successiva precisazione confermata: SSH/Firewall indipendente, sempre accessibile, con clienti autonomi per server.
- Implementati guard route e sidebar, CRUD dashboard e aggregazioni reparti, gerarchia UI/service/trigger DB, refresh/paginazione CompanyContext, scope reparto AP e guida, PDF/CSV per reparto, campo/filtro/gruppi client_name SSH. Migrazioni 000008/000009 up/down. [Piano](04-development/plans/COMPANY-UX-02.md), [QA](05-qa/COMPANY_WORKFLOW_QA.md).
- Verifiche: frontend 26 file/164 test e build PASS; backend suite/build/vet PASS con DB_NAME esplicito di test. Integrazione repository SSH/etichette, migrazioni/idempotenza, down/up e trigger PostgreSQL PASS exit 0. Database dedicato creato e mantenuto: nextgen_company_ux02_test_20261017. Database applicativo solo letto (0 sotto-reparti), nessuna modifica dati/schema o connessione SSH remota.
- Limiti: computer-use senza app/browser e renderer PDF assente; PDF verificato in memoria per contenuto/paginazione, non visualmente. Task [!] per gate UI/PDF. Assunzioni dichiarate: clienti SSH non sono nuove ACL; dossier aziendali multipagina ammessi per inventario; reparto AP mostra solo catene interne, vista azienda include ipotesi tra reparti. Nessuna dipendenza, push, commit o deploy.


- **AP-REVIEW-01:** richiesta PO di validazione cyber, funzionamento, miglioramento grafico e guida per `/attack-path`. Confermati dati presenti/scenari sintetici e 3D solo se utile. Corretti cammini disconnessi, falsa esposizione Internet, pivot/trust arbitrari, collisione TCP/UDP, scope globale di fallback, layout non terminante su cicli e fetch su selezione. UI 2D percorso/completa, inferenze esplicite, indice euristico, provenienza e guida integrata. [Report QA](05-qa/ATTACK_PATH_REVIEW.md), [piano](04-development/plans/AP-REVIEW-01.md).
- **Verifica AP 2026-10-17:** test backend mirati rosso→verde, `go test ./...`, `go build ./...`, `go vet ./...` PASS exit 0; frontend `npm.cmd test` PASS (23 file, 156 test) e build PASS exit 0. Snapshot locale in sola lettura: 11 tenant, 17 target, 20 finding, nessuna CVE. Servizio aggiornato validato su snapshot per tenant a soglie 0/7/9. Nessun exploit, scansione o cambiamento DB. Browser assenti nel runtime (iab/Chrome indisponibili): niente screenshot né PASS visuale, task `[!]`.
- **Assunzioni AP:** strumento di triage euristico, nessuna certificazione di reachability/compromissione; /24 convenzionale non maschera dimostrata. 2D scelto per sequenza senza occlusioni, non sulla base di user testing. Dati storici e cap dichiarati; approfondimenti in TECH-DEBT-08. Guida nel prodotto, `docs/` solo agenti. Nessuna dipendenza o migrazione.

- **E28-S1-T1:** primo incremento backend implementato in working tree: preview e remediation SERVICE IPv4 TCP/UDP sul server del chokepoint, scelta confermata dal PO. Timer 60s obbligatorio, porta SSH protetta, server e grafo validati nel tenant, RBAC SecOps/admin. Correzioni watchdog su PID/concorrenza/conferma/rollback. Stato `[ ]`: gate operativo aperto per prova Linux e hardening preesistente SSH/watchdog (TECH-DEBT-06/07), nessuna consegna per accettazione ancora. Piano in `docs/04-development/plans/E28-S1-T1-chokepoint-remediation.md`.
- Verifiche 2026-10-17 in `backend`: `go test ./...`, `go build ./...`, `go vet ./...` PASS exit 0; test mirati service/handler PASS exit 0. In `frontend`: `npm.cmd test` (145 test), `npm.cmd run build` PASS exit 0, eseguiti fuori sandbox dopo blocco esbuild. Cache Go locale ignorata dopo errore della cache predefinita. Race detector non eseguibile (CGO disabilitato, exit 1); nessuna prova remota/visuale.
- Assunzioni tecniche esplicite: IPv4 letterale coincidente fra servizio e host SSH; niente DNS/IPv6/gateway o isolamento; riuso lifecycle firewall esistente, singola istanza backend. Nessuna nuova dipendenza o migrazione. Modifiche preesistenti a CHANGELOG, TASK_LIST e PROJECT_STATE preservate; nessun commit/push/deploy.

- Checkpoint `b16264b`: 23 file remediation/firewall, test/DTO e documenti/archivio task pubblicati prima della revisione AP. Dopo rifiuto iniziale dell'auto-review, PO conferma contenuto e repository/branch esplicitamente. Gate Linux di E28-S1-T1 invariato; commit non equivale ad accettazione funzionale.

## Accettati recentemente dal PO (`[x]`)

- **TEST-OPT-01, COMPANY-UX-02, COMPANY-PDF-03, AP-REVIEW-01:** accettazione esplicita globale del PO il 2026-10-17 e richiesta di pubblicare tutto su main. Assunzioni confermate, inclusi Executive sintetico con cinque priorità per reparto, SSH indipendente con clienti autonomi e Attack Path euristico 2D. Limiti visuali dichiarati accettati per la consegna; nessuna prova visuale autonoma o attestazione di lateral movement reale dedotta. Rischi SSH/Linux e task futuri mantengono i loro stati tecnici.

- **AGENT-CLEAN-03 / AGENT-CODEX-01:** confermati e approvati dal PO il 2026-10-16. Accettate pulizia repository e migrazione governance Codex; assunzioni documentali confermate. Task dettagliati archiviati nel CHANGELOG insieme ai task accettati preesistenti, essendo stato superato il limite di 10. Nessuna approvazione retroattiva del vecchio design Epic 27.1: registrato come superato.

- **SCAN-UX-01 / SCAN-TIME-01:** accettati il 2026-10-16. Il PO riferisce di avere testato le scansioni e ne conferma il corretto funzionamento, autorizzando il push di tutte le modifiche. Prova funzionale del PO distinta dai test automatici dell'agente; nessuno screenshot autonomo o benchmark di accuratezza ETA dichiarato. Assunzioni operative confermate dall'accettazione: T3, stime indicative, profili/ID storici preservati e monitor dell'intero gruppo. TECH-DEBT-05 sul consenso API resta aperto.

- **SCAN-COPY-01 · Descrizioni dei profili:** accettato dal PO con OK il 2026-10-16. Otto descrizioni focalizzate su scopo e attività; rimosso suggerimento Discovery sotto il selettore. Nessuna assunzione da confermare. Evidenze precedenti: frontend 145 test, build e controlli documentali PASS, exit 0. Accettazione limitata ai testi; SCAN-UX-01 e SCAN-TIME-01 mantengono i rispettivi stati.

- **AGENT-SDLC-02 · Struttura documentale e orchestrazione per feature:**
  - Validato e accettato dal PO in data 2026-10-15.
  - Documentazione canonica organizzata in `01-product`–`06-devops`, changelog in root, sette ruoli specializzati e gate proporzionati alla singola feature.
  - Confermata l'assunzione operativa: ruoli attivati quando pertinenti, validazione tecnica autonoma del Tech Lead e accettazione funzionale riservata al PO.
- **Epic 27 · Gestione Aziende, Reparti & Filtro Globale (E27-S1-T1 .. E27-S1-T6)**:
  - Validato e accettato dal PO in data 2026-10-11.
  - **CompanyContext & Scope Globale:** Selettore Azienda persistente in sidebar desktop/drawer mobile; calcolo automatico di tutti i reparti e IP appartenenti all'azienda attiva.
  - **Dashboard Gestione Aziende (`/companies`):** Nuova interfaccia con card metriche (reparti, IP, postura, rischio), color tag picker, creazione, modifica ed eliminazione con conferma; pulsante per impostare al volo lo scope attivo.
  - **Contestualizzazione Reparti in `/targets`:** Visualizzazione e gestione esclusiva dei soli reparti dell'azienda selezionata, con creazione rapida di sotto-reparti e assegnazione IP.
  - **Propagazione e Pulizia Ridondanze su tutte le feature:** Applicato filtro IP e banner Scope Attivo su Dashboard, Scansioni, Nuova Scansione (filtraggio gruppi e target), Vulnerabilità CVE (ispezione per reparto ed IP), Report (dossier aziendale consolidato) e Attack Path Mapping.
- **Epic 26 · SEC-P2: Attack Path Refinement & Advanced Orchestration (SEC-P2-01 .. SEC-P2-03)**:
  - Validato e accettato dal PO in data 2026-10-10.
- **Epic 25 · SEC-P1: Scanner Engine Overhaul & Profiling Accuracy (SEC-P1-01 .. SEC-P1-06)**:
  - Validato e accettato dal PO in data 2026-10-10.
- **Epic 24 · SEC-P0: Hardening Critico & Scanner Security Perimeter (SEC-P0-01 .. SEC-P0-06)**:
  - Validato e accettato dal PO in data 2026-10-10.
- **Milestone 2 · Visual Graph Explorer & Robustezza Suite (FEAT-M2-03, ONBOARD-01 .. ONBOARD-05)**:
  - Validato e accettato dal PO in data 2026-10-10.
- **FEAT-REPORT-PDF-UX-01 · Perfezionamento Grafico, Tipografico & Coerenza Dati dei Report PDF A4 (FR-18, UC-17)**:
  - Validato e accettato dal PO in data 2026-10-10.

## In attesa di accettazione utente

Nessuno: tutti gli incrementi consegnati sono accettati dal PO il 2026-10-17. Evidenze e limiti nei report QA; nessun PASS tecnico aggiuntivo attribuito all'approvazione.

## Bloccati / Domande aperte per l'utente

- **Limite visuale COMPANY-UX-02 / COMPANY-PDF-03:** collaudo autonomo desktop/mobile e PDF non eseguito, browser e renderer non disponibili. Prove automatiche/DB concluse; il PO accetta gli incrementi con questo limite esplicito e autorizza pubblicazione. Nessun PDF autonomo consegnato come verificato visualmente, nessuna domanda funzionale pendente.


- **Limite visuale AP-REVIEW-01 accettato dal PO:** inventario browser vuoto, verifica grafica autonoma e screenshot non eseguiti; DOM e interazioni coperte da test. I dati locali privi di CVE non permettono validazione reale di catene di exploit; scenari sintetici coprono il modello. Nessun blocco alla pubblicazione richiesta.

- Il PO conferma il 2026-10-17 che non è ancora disponibile un host Linux di laboratorio e chiede di mantenere aperta la verifica remota E28-S1-T1. Prima dell'uso operativo risolvere anche TECH-DEBT-06/07: host-key non verificata, chiamate SSH senza deadline e coordinamento watchdog. Nessuna azione remota eseguita.

- Nessun blocco sulla consegna scanner: il PO ha eseguito i test delle scansioni e confermato il corretto funzionamento il 2026-10-16. Verifica browser autonoma e benchmark ETA non effettuati dall'agente; debito TECH-DEBT-05 sul consenso API mantenuto.

## Assunzioni confermate

- Struttura documentale SDLC v7.1 e orchestrazione per feature approvate dal Product Owner; anche la migrazione base Codex AGENT-CODEX-01 e la pulizia AGENT-CLEAN-03 sono approvate.
- `docs/` contiene solo documentazione necessaria agli agenti; manuali utente, screenshot, allegati binari e archivi sono esclusi, come richiesto dal PO.
- Storico consolidato nel [CHANGELOG.md](../CHANGELOG.md); stub e cartella legacy rimossi dopo la migrazione alle fonti canoniche.
- Pipeline CI/CD GitHub Actions e bundle splitting implementati.
- Milestone 2 completata sia su Backend Go che su Frontend React.
- Roadmap di hardening sicurezza (Epiche SEC-P0, SEC-P1, SEC-P2) implementata al 100%.

## Decisioni recenti (ultime 10, dettagli negli ADR)

- **ADR-19 (2026-10-17):** test mirati per iterazione, un gate finale condiviso per lato coinvolto, riuso di sole evidenze valide, misure per comando. PostgreSQL esplicito con TEST_DB_* senza fallback applicativo, host locale e DB dedicato; indisponibilità = FAIL. CI con job DB isolato. Decisione richiesta dal PO, non deroga estemporanea.

- **ADR-18 (2026-10-17):** dashboard unica, guard aziendale con eccezione tool SSH, clienti SSH autonomi, gerarchia due livelli nelle nuove scritture, report/IP per reparto e selettore reparto AP senza soglia UI.


- **ADR-17 (2026-10-17):** Attack Path come analisi euristica verificabile con evidenze/ipotesi distinte, percorsi su archi reali del modello, score non probabilistico, scope aziendale completo e guida UI 2D. Nessuna inferenza di trust dall'inventario SSH. Gate visuale separato.

- **ADR-15 (2026-10-16):** telemetria XML Nmap, snapshot attivi e lettura batch per monitor lotto; API additive, nessuna migrazione DB, ETA esplicitamente euristica. Durate indicative richieste dal PO dopo la precedente rimozione delle stime promozionali.

- **ADR-14 (2026-10-15):** profili espliciti + editor Nmap entro allowlist, API/DB compatibili, matching argomenti/timeout e classificazione server dei custom. Nessuna promessa di penetration test/conformità; Discovery rispetta l'opzione porte.

- **ADR-13 (2026-10-15):** Documentazione `docs/` riservata agli agenti; rimossi manuale e screenshot, evidenze temporanee fuori da `docs/`, controllo strutturale automatizzato.
- **ADR-12 (2026-10-15):** Documentazione per fasi SDLC, ruoli con input/output e handoff, gate per feature, fonti canoniche distinte e migrazione senza duplicare i documenti storici. Dettagli in `docs/03-architecture/ADR/README.md`.

- **ADR-11 (2026-10-15):** `AGENTS.md` è l'entry point nativo Codex; `.agent/workflows/` è una libreria di playbook indicizzata, non un registro di slash command. Introdotti bootstrap evidence-first, tutela delle modifiche preesistenti, fallback per browser/subagent e permessi espliciti per commit, push, merge, tag e deploy.
- **ADR-10 (2026-10-15):** Standard Operativo Software House: Context Layering su 4 livelli gerarchici per contenimento token bloat e azzeramento allucinazioni; Mappa Documenti Full-Stack estesa a frontend React (`pages`, `components`, `context`), backend Go e migrazioni SQL; protocollo Contract-First (allineamento preventivo DTO TypeScript e Go struct prima del codice di business); verifica di backward & rollback (`down.sql`) per le migrazioni DB; DoD standard a 4 punti per ogni task; istituzione del Tech Debt & Refactoring Backlog formale in `3_TASK_LIST.md`; nuovi workflow operativi `.agent/workflows/` (`review.md`, `db-migrate.md`, `release.md`).
- **ADR-09 (2026-10-11):** Roadmap Strategica Milestones 3, 4, 5: Approvazione trittico evolutivo: Milestone 3 (Active Defense & SOAR Light: Chokepoint Auto-Remediation, Virtual Patching, Alerting Hub), Milestone 4 (Continuous Compliance: NIS2, DORA, ISO 27001, Scorecard, Dossier PDF) e Milestone 5 (AI Remediation Copilot & Assistant).
- **PO-04 (2026-10-11):** Approvazione priorità strategiche e inclusione formale nella roadmap attiva di Opzione 2 (Active Defense), Opzione 3 (Compliance Frameworks) e Opzione 5 (AI Remediation Copilot).
- **ADR-08 (2026-10-11):** Architettura di Scope Aziendale Gerarchico e `CompanyContext` globale React: distinzione formale tra nodi radice (`parent_id IS NULL` = Aziende clienti) e nodi discendenti (`parent_id != NULL` = Reparti aziendali); persistenza in `localStorage`; propagazione trasparente e unificata del filtro target IP su tutte le feature (Dashboard, Targets, Scans, Findings, Reports, AttackPath) con eliminazione dei selettori ridondanti.
- **ADR-07 (2026-10-10):** Integrazione serializzazione `Invasiveness` in `ProfileResponse` API e seeding profilo `All-in-One 360° Comprehensive Audit`.
- **ADR-06 (2026-10-10):** Attack Path Probabilistic Risk Formula $1 - \prod(1 - p_i)$, bound 50 cammini, archi inferiti tratteggiati con confidenza, e Scan Diff `GET /api/v1/scans/{id}/diff` con stati NEW/RESOLVED/PERSISTENT.
- **ADR-05 (2026-10-10):** Profilo specialistico UDP, tassonomia di invasività con gate UI, modello finding arricchito con confidence/detection_method/cpe, hardening container capabilities (solo `CAP_NET_RAW`), migrazione PDF a `go-pdf/fpdf v0.9.0`.

## Contesto operativo della migrazione

- Le baseline applicative sopra sono storiche, non rieseguite per la modifica a documenti e verificatore documentale; nessun codice applicativo è stato modificato.
- Working tree già modificata dalla migrazione Codex precedente; le eliminazioni legacy preesistenti sono state mantenute e ora risultano coerenti con la pulizia delle vecchie strutture.
- Documenti canonici e indice: `docs/README.md`; task attivi: `docs/04-development/TASK_LIST.md`. Nessun commit, push o deploy eseguito.

## Prossimi passi

1. Modifiche approvate pubblicate su main con fast-forward e push riuscito (`26cf93d`); registro documentale di consegna aggiornato. Accettazione PO acquisita, limiti visuali residui registrati. Nessun nuovo branch o deploy.
2. E28-S1-T1 resta separato: hardening SSH/watchdog TECH-DEBT-06/07 e prova Linux quando disponibile.
3. Proseguire Epic 28 secondo dipendenze e priorità PO; nessun push/merge/deploy implicito della revisione AP.

## Log sessioni (ultime 5, la più recente in alto)

- 2026-10-17 · PO approva tutto e richiede push su main. Accettati TEST-OPT-01/COMPANY-UX-02/COMPANY-PDF-03/AP-REVIEW-01 con limiti visuali documentati. Fonti applicative/test/config invariati dalle evidenze PASS: riuso proporzionato, controlli documentali e staging rieseguiti PASS. Controllo file sensibili/firme: unico header chiave è placeholder UI, nessun corpo chiave o token rilevato. Commit `26cf93d` (91 file), main aggiornato fast-forward e push origin/main riuscito, nessun force push/nuovo branch/deploy. E28-S1-T1/Linux e TECH-DEBT-06/07 restano aperti. COMPANY-PDF-03 include CVE per reparto, Tecnico completo ed Executive top cinque locali.

- 2026-10-17 · TEST-OPT-01: autorizzata ottimizzazione del testing, implementati gate per lato e riuso evidenze, runner misurato e tag integration con TEST_DB_* sicure, CI PostgreSQL effimero. Backend suite/build/vet e DB reali PASS, configurazione 12 casi PASS, smoke frontend 30 test PASS; negativi DB/config e runner FAIL attesi verificati. Container solo test arrestato/rimosso. Doc Sync e self-review; `[~]` in attesa PO. Nessuna dipendenza, modifica applicativa, commit/push/deploy; CI remota non eseguita. Modifiche preesistenti preservate.

- 2026-10-17 · COMPANY-UX-02: integrata richiesta dashboard/aziende/reparti/report/AP e precisazione SSH indipendente con clienti autonomi confermata. UI/service/DB e DTO aggiornati; test gerarchia rosso→verde, frontend 164 test/build e Go suite/build/vet PASS. Up/down e trigger verificati su DB test separato, dati applicativi invariati. PDF contenuto/paginazione PASS, gate browser/render aperti [!]. Nessuna nuova dipendenza, commit/push/deploy; AP precedente preservato e non accettato.


- 2026-10-17 · Checkpoint `b16264b` pubblicato dopo conferma esplicita del payload/destinazione richiesta da auto-review. Revisione AP-REVIEW-01: difetti backend riprodotti rosso→verde, correzioni modello/scope/layout e guida integrata. Suite Go/build/vet, frontend 156 test/build PASS exit 0; snapshot locale read-only 17 target/20 finding/0 CVE validato per tenant. Nessuna prova visuale per browser indisponibile, task `[!]`; revisione non committata/pushata. Self-review e Doc Sync, nessuna dipendenza/migrazione o modifica DB.

- 2026-10-17 · Avviato E28-S1-T1 con feature.md/orchestration.md. PO sceglie blocco porta sul server chokepoint, isolamento separato. Backend/DTO/test e fix watchdog implementati; suite backend/build/vet e frontend 145 test/build PASS exit 0. Test PID rosso→verde; race non eseguibile senza CGO. Self-review e documenti allineati; task `[ ]` per gate Linux/hardening SSH residuo. Nessuna azione remota, commit o push; modifiche precedenti preservate.
