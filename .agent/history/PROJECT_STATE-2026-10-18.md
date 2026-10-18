# Historical state snapshot · 2026-10-18

Historical evidence only. Read when reconstructing prior work, never during ordinary bootstrap. Authorizations below apply only to their original task/action and do not authorize new operations. Captured before AGENT-HARNESS-01 compaction; content and acceptance history preserved, relative links rebased from docs/ to this location. Current truth: [PROJECT_STATE](../../docs/PROJECT_STATE.md).

---

# PROJECT STATE

**Ultimo aggiornamento:** 2026-10-18
**Modalità progetto:** esistente (onboarded formalmente).

> Memoria viva dell'agente. Letta a inizio sessione, aggiornata a fine di ogni turno.
> Non cancellare sezioni: se vuote, scrivi "nessuno".

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

La baseline è consultata prima degli interventi, non rieseguita automaticamente: politica proporzionata `AGENTS.md` sez. 8 e [Test Plan](../../docs/05-qa/TEST_PLAN.md). Una sola suite finale per lato coinvolto; test mirati durante lo sviluppo e riuso solo con input/ambiente invariati. Il percorso rapido Go ora esclude PostgreSQL, che richiede tag `integration`; i PASS storici precedenti non vengono reinterpretati.

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

**Fase:** rimozione della feature verificata e pubblicata su origin/main con le modifiche SSH/firewall, commit `4f9ea78` (25 file), push PASS exit 0. Riprogettazione da definire; nessun deploy.

## In corso

**M11-CLOSE-01 · 2026-10-18:** PO richiede completamento M1.1, senza riordino
delle milestone. Applicati feedback/feature e orchestration; working tree iniziale
pulita su main, nessun branch/commit/push/deploy. Ricognizione FR-07..FR-13 nella
[matrice QA](../../docs/05-qa/M11_CLOSURE_QA.md), piano in
[M11-CLOSE-01](../../docs/04-development/plans/M11-CLOSE-01.md). Individuati gap funzionali
oltre ai collaudi: filtri report, operazioni GUI firewall, persistenza discovery.
Primo incremento: validazione IPv4/interfacce/commenti, quoting shell dei commenti,
timer add/delete non disattivabile e timeout 10..300s con errori HTTP 400.
Riproduzione rossa exit 1, test mirati backend e 2 test UI PASS exit 0; verifica
finale PASS per il primo incremento: backend suite 20,816s/vet 1,189s/build 1,432s;
frontend 168 test 5,032s/build 15,301s, docs 0,760s/diff 0,051s, tutti exit 0.
Build iniziale FAIL exit 2 per fixture TypeScript incompleta, corretta; rieseguito
solo il gate frontend, backend invariato. Dettaglio e report nella matrice QA.
Nessun dato o host remoto modificato. Browser runtime apps/browser
vuoti: gate visuale aperto. Docker accessibile con escalation, nessun laboratorio
ancora avviato. TECH-DEBT-07 e milestone restano aperti. Assunzione reversibile:
compatibilità campo legacy safety_timer preservata, false non rimuove protezione.
Decisione PO 2026-10-18: iptables e iptables-nft con modello attuale; nessun
motore nftables nativo. Prosegue M11-REPORT-01 (solo frontend): archivio completo
tramite paginazione API esistente, filtri locali e scope anche cronologico.
M11-REPORT-01 implementato: filtri inclusivi UTC/profilo/target/cartella/severità,
cronologico paginato, errori su archivio parziale e risposte superate ignorate.
Gate frontend PASS exit 0: 175 test 4,995s, build 10,805s, docs 0,699s/diff 0,053s;
report `20261018-113105-c03e6e33.json`. Baseline backend invariata riutilizzata.
Self-review e Doc Sync; nessuna API/SQL/migrazione/dipendenza modificata. Limite
caricamento storico lineare registrato TECH-DEBT-10/ADR-22. Gate visuale/PDF
rimane aperto, task non promosso; nessun commit/push/deploy o accesso target.

**CI-CONTEXT-01 · 2026-10-18:** PO segnala workflow non valido al run 39113975631: `job` non ammesso nell'env del job. Riproduzione locale actionlint v1.7.7 exit 1 conferma contesto e indice numerico; spostata TEST_DB_PORT nell'env dello step integration con chiave stringa '5432'. Porta dinamica, TEST_DB_* restanti, healthcheck e dipendenze job invariati. Solo workflow e documenti; nessun codice/schema/DB/servizio modificato. Tool temporaneo in cache artifacts, nessuna dipendenza aggiunta ai manifest. Primo download bloccato dalla sandbox, ripetuto con escalation autorizzata. Validazione finale in corso; nessun PASS remoto dichiarato.

**README-GH-01 · correzione porta 2026-10-18:** PO indica l'interfaccia su `http://localhost:3080`. Applicato feedback/modifica: URL e troubleshooting README corretti; quickstart e procedura artifact impostano esplicitamente `FRONTEND_PORT=3080` in PowerShell, dato il fallback 3000 del manifest. Vite nativo resta 3000. Nessuna lettura di `.env`, modifica runtime o configurazione applicativa; destinazione operativa fornita dal PO, raggiungibilità non verificata. Correzione compresa nella richiesta precedente di pubblicazione README su main; accettazione editoriale ancora pendente.

Verifiche correzione dalla root: verificatore documentale PASS exit 0 (1,035 s, 224 link/ancore); `git diff --check` PASS exit 0 (con stat: 0,113 s). Self-review dei riferimenti e dei comandi PowerShell completata; solo Markdown, suite applicative non necessarie. Pubblicazione prevista: `docs(readme): correct Docker web interface port to 3080`.

**README-GH-01 · 2026-10-18:** PO richiede README completo per GitHub e push su main, autorizzando commit/pubblicazione di questo incremento. Working tree iniziale pulita su main, baseline `e692e76`; fetch origin PASS exit 0, divergenza 0/0. Applicati feedback/modifica, orchestration e commit. README ampliato con funzionalità, architettura, quickstart, configurazione effettiva, sviluppo, test, artifact CI, troubleshooting, limiti, roadmap e licenza non dichiarata. Fonti: manifest, config/server, route e documentazione canonica. Nessun codice, governance, dipendenza o servizio modificato. Assunzione reversibile: README in italiano, coerente con repository e richiesta. Solo verifiche documentali previste; nessun avvio stack, scansione, deploy o suite applicativa dichiarati. Pubblicazione autorizzata distinta dall'accettazione editoriale del PO.

**DOCKER-CI-01 · pubblicazione richiesta 2026-10-17:** PO «effettua il push» autorizza commit e push delle nove modifiche Docker/Actions accettate su origin/main. Applicato commit.md, fetch origin PASS e divergenza iniziale 0/0; nessuna modifica estranea. Evidenze build/smoke/archivio/Compose precedenti riutilizzate perché sorgenti, dipendenze e configurazione invariati; soli Markdown aggiornati dopo i controlli. Messaggio commit: `ci(docker): package verified images as workflow artifacts`. La richiesta autorizza l'esecuzione CI conseguente al push, nessun deploy o pubblicazione registry. Esito push da verificare al termine; questa nota registra l'autorizzazione e la preparazione.

**DOCKER-CI-01 · 2026-10-17:** PO «ok procedi» sulla proposta Docker/Actions. Applicati feature/orchestration, working tree iniziale pulita su main; mantenuto main secondo preferenza PO già registrata, nessun branch/commit/push. Estesa CI con avvio manuale e job immagini Linux amd64 subordinato ai tre job esistenti; pacchetto artifact 7 giorni con archivio, checksum/revisione e Compose. Aggiunti override immagini e dockerignore; frontend npm ci. Build backend/frontend, smoke Nmap/Nginx senza rete, save/load GZip e Compose config PASS exit 0; dettagli e durate in [CI](../../docs/06-devops/CI_CD_PIPELINE.md#docker-ci-01--pacchetto-scaricabile). Doc Sync e self-review completati. Nessun deploy, DB o target toccato. Prima esecuzione Actions/upload non effettuata; suite applicative restano gate remoto, non rieseguite localmente per sorgenti invariati. Avvisi npm registrati come TECH-DEBT-09.

**SSH-TOOLS-01 · accettazione autorevole 2026-10-17:** PO «Ok perfetto è funzionante» conferma il flusso SSH/terminale/iptables dopo la diagnosi del 502. Task `[x]`, rimosso dalle attese di accettazione. Le richieste di output Debian/refresh sotto sono storiche e non più pendenti; non si deduce quale intervento abbia risolto l'errore né si inventano prove autonome sul target. Assunzione operativa sui prerequisiti predisposti dal PO accettata nel flusso funzionante. SSH-HARDEN-01/TECH-DEBT-07 e prove watchdog/rollback restano separati. Solo registrazione documentale, nessun codice o comando remoto, commit/push/deploy; test applicativi non rieseguiti.

**SSH-TOOLS-01 · feedback successivo 2026-10-17:** PO chiede se iptables debba essere sul server e riporta ancora «Request failed with status code 502svg». Confermato requisito del pacchetto sul Debian remoto e privilegi root/sudo. Ricognizione read-only locale: bundle attivo SSHServers-BeM5cCm3.js coincide con build verificata; stringa FIREWALL_PRIVILEGES_REQUIRED presente nel binario /app/server; richiesta senza credenziali al proxy `/api/v1/ssh/servers` restituisce 401 JSON atteso, exit 0 curl. Immagini container diverse dagli ultimi tag disponibili, ma questo NON prova assenza del fix (presenza verificata direttamente). Nessuna diagnosi definitiva di pacchetto assente, sudo o cache: proposti Ctrl+F5 e, nella sessione Debian del PO, `dpkg-query -W iptables` / `sudo -n /usr/sbin/iptables -L -n -v --line-numbers`. Nessuna lettura credenziali, connessione remota, installazione, modifica codice o deploy; task riaperto `[ ]`. Test applicativi invariati, non rieseguiti; verifiche documentali a parte.

**SSH-TOOLS-01 · 2026-10-17 (aggiornamento autorevole):** PO «funziona la connessione» accetta SSH-CONNECT-02; terminale KO con errore WebSocket e iptables KO con HTTP 502. Corretti proxy `/ws/` mancanti in Nginx/Vite e lettura firewall root/PATH/sudo non interattivo; nuovi errori API sanificati e frontend mostra messaggio API. [QA](../../docs/05-qa/SSH_TOOLS_QA.md): riproduzioni rosse→verdi, proxy echo bidirezionale Vite/Nginx PASS, gate Final All PASS exit 0 con 175 test UI. Nessuna nuova dipendenza, modifica DB, accesso target, lettura credenziali, concessione privilegi o deploy; modifiche preesistenti preservate. Assunzione reversibile: eventuali privilegi sudo sul Debian sono predisposti dall'amministratore, non concessi dall'app; stato effettivo ancora ignoto.

Evidenze correnti dalla root con `scripts/verify.ps1 -Profile Final -Scope All`: backend `go test ./...` PASS exit 0 (23,983s), `go vet ./...` PASS exit 0 (3,918s), `go build ./...` PASS exit 0 (1,493s); frontend inizialmente FAIL exit 1 per esbuild sandbox. Report `.agent/artifacts/testing/20261017-153007-31f2fb13.json`. Da backend `go test -tags=integration -count=1 ./internal/repository/postgres` FAIL exit 1 (4,984s wall), configurazione TEST_DB_HOST assente; nessuno skip o PASS DB dedotto.

**Aggiornamento autorevole 2026-10-17:** “Ok puoi pushare tutto sul main (non voglio vedere branch) Approvo tutto.” Tutte le modifiche presenti, incluse quelle preesistenti, sono autorizzate per commit e push su `https://github.com/Alexx904/NextGen-CyberSecurity.git`, destinazione `main`. Nessun nuovo branch o force push. I resoconti sotto sono evidenze storiche degli incrementi; gli stati precedenti `[!]`/`[~]` sono superati dall'accettazione PO per i quattro task consegnati. Rischio comunicato una volta: prove visuali non effettuate e laboratorio SSH assente. L'accettazione con questi limiti non dichiara PASS visuale né chiude lavoro non implementato. Test applicativi invariati rispetto alle ultime evidenze, riutilizzati secondo AGENTS sez. 8; controlli documentali/staging rieseguiti.

- **COMPANY-PDF-03 (2026-10-17) `[!]`:** richiesta PO di separare anche le CVE per reparto nei PDF dell'intera azienda. Implementati schede tecniche complete per reparto e top cinque Executive locali con totale, KPI aziendali mantenuti, intestazioni di continuazione, target diretti e reparti vuoti distinti. Aziende senza reparti/singoli download conservano il flusso precedente. Raggruppamento per identità gruppo e scan/target, non IP/nome; test collisione etichetta “Senza reparto”. Gate finale backend runner PASS: suite 16,296 s, vet 1,034 s, build 1,340 s; documenti/diff PASS exit 0. Evidenza e Doc Sync nel [QA](../../docs/05-qa/COMPANY_WORKFLOW_QA.md#incremento-company-pdf-03--cve-per-reparto). Assunzione: Executive resta sintesi; tutti i finding nel Tecnico. Renderer assente: contenuto/paginazione verificati, visuale aperto. Nessuna dipendenza/API/DB/frontend modificata, commit/push/deploy o mutazione dati. Modifiche TEST-OPT/COMPANY/AP preesistenti preservate.

- **TEST-OPT-01 (2026-10-17) `[~]`:** PO autorizza l'ottimizzazione proposta. Regole proporzionate, runner con durate/report, integrazioni PostgreSQL esplicite e CI con DB effimero. Backend suite/build/vet, DB reale e 12 casi configurazione PASS; frontend runner smoke 30 test PASS; FAIL attesi senza DB/config e fail-fast runner verificati. Doc Sync e self-review completati; nessuna modifica applicativa o nuova dipendenza. Le modifiche preesistenti COMPANY/AP sono preservate. Nessun commit/push/deploy; CI remota non eseguita.

- **COMPANY-UX-02 (2026-10-17):** PO richiede dashboard/aziende unica, reparti solo diretti, report per reparto/azienda e IP distinti nel documento, AP senza CVSS UI con selettore reparto, obbligo azienda per navigare. Successiva precisazione confermata: SSH/Firewall indipendente, sempre accessibile, con clienti autonomi per server.
- Implementati guard route e sidebar, CRUD dashboard e aggregazioni reparti, gerarchia UI/service/trigger DB, refresh/paginazione CompanyContext, scope reparto AP e guida, PDF/CSV per reparto, campo/filtro/gruppi client_name SSH. Migrazioni 000008/000009 up/down. [Piano](../../docs/04-development/plans/COMPANY-UX-02.md), [QA](../../docs/05-qa/COMPANY_WORKFLOW_QA.md).
- Verifiche: frontend 26 file/164 test e build PASS; backend suite/build/vet PASS con DB_NAME esplicito di test. Integrazione repository SSH/etichette, migrazioni/idempotenza, down/up e trigger PostgreSQL PASS exit 0. Database dedicato creato e mantenuto: nextgen_company_ux02_test_20261017. Database applicativo solo letto (0 sotto-reparti), nessuna modifica dati/schema o connessione SSH remota.
- Limiti: computer-use senza app/browser e renderer PDF assente; PDF verificato in memoria per contenuto/paginazione, non visualmente. Task [!] per gate UI/PDF. Assunzioni dichiarate: clienti SSH non sono nuove ACL; dossier aziendali multipagina ammessi per inventario; reparto AP mostra solo catene interne, vista azienda include ipotesi tra reparti. Nessuna dipendenza, push, commit o deploy.

- **AP-REVIEW-01:** richiesta PO di validazione cyber, funzionamento, miglioramento grafico e guida per `/attack-path`. Confermati dati presenti/scenari sintetici e 3D solo se utile. Corretti cammini disconnessi, falsa esposizione Internet, pivot/trust arbitrari, collisione TCP/UDP, scope globale di fallback, layout non terminante su cicli e fetch su selezione. UI 2D percorso/completa, inferenze esplicite, indice euristico, provenienza e guida integrata. [Report QA](../../docs/05-qa/ATTACK_PATH_REVIEW.md), [piano](../../docs/04-development/plans/AP-REVIEW-01.md).
- **Verifica AP 2026-10-17:** test backend mirati rosso→verde, `go test ./...`, `go build ./...`, `go vet ./...` PASS exit 0; frontend `npm.cmd test` PASS (23 file, 156 test) e build PASS exit 0. Snapshot locale in sola lettura: 11 tenant, 17 target, 20 finding, nessuna CVE. Servizio aggiornato validato su snapshot per tenant a soglie 0/7/9. Nessun exploit, scansione o cambiamento DB. Browser assenti nel runtime (iab/Chrome indisponibili): niente screenshot né PASS visuale, task `[!]`.
- **Assunzioni AP:** strumento di triage euristico, nessuna certificazione di reachability/compromissione; /24 convenzionale non maschera dimostrata. 2D scelto per sequenza senza occlusioni, non sulla base di user testing. Dati storici e cap dichiarati; approfondimenti in TECH-DEBT-08. Guida nel prodotto, `docs/` solo agenti. Nessuna dipendenza o migrazione.

- **E28-S1-T1:** primo incremento backend implementato in working tree: preview e remediation SERVICE IPv4 TCP/UDP sul server del chokepoint, scelta confermata dal PO. Timer 60s obbligatorio, porta SSH protetta, server e grafo validati nel tenant, RBAC SecOps/admin. Correzioni watchdog su PID/concorrenza/conferma/rollback. Stato `[ ]`: gate operativo aperto per prova Linux e hardening preesistente SSH/watchdog (TECH-DEBT-06/07), nessuna consegna per accettazione ancora. Piano in `docs/04-development/plans/E28-S1-T1-chokepoint-remediation.md`.
- Verifiche 2026-10-17 in `backend`: `go test ./...`, `go build ./...`, `go vet ./...` PASS exit 0; test mirati service/handler PASS exit 0. In `frontend`: `npm.cmd test` (145 test), `npm.cmd run build` PASS exit 0, eseguiti fuori sandbox dopo blocco esbuild. Cache Go locale ignorata dopo errore della cache predefinita. Race detector non eseguibile (CGO disabilitato, exit 1); nessuna prova remota/visuale.
- Assunzioni tecniche esplicite: IPv4 letterale coincidente fra servizio e host SSH; niente DNS/IPv6/gateway o isolamento; riuso lifecycle firewall esistente, singola istanza backend. Nessuna nuova dipendenza o migrazione. Modifiche preesistenti a CHANGELOG, TASK_LIST e PROJECT_STATE preservate; nessun commit/push/deploy.

- Checkpoint `b16264b`: 23 file remediation/firewall, test/DTO e documenti/archivio task pubblicati prima della revisione AP. Dopo rifiuto iniziale dell'auto-review, PO conferma contenuto e repository/branch esplicitamente. Gate Linux di E28-S1-T1 invariato; commit non equivale ad accettazione funzionale.

## Accettati recentemente dal PO (`[x]`)

- **DOCKER-CI-01 · 2026-10-17:** PO «OK» accetta il pacchetto Docker/Actions e le assunzioni consegnate: Linux amd64, artifact per 7 giorni, PostgreSQL separato, nessun registry/deploy. Task `[x]`, attesa rimossa. Build e controlli locali PASS già documentati; prima esecuzione Actions/upload ancora da verificare. Applicato feedback.md ramo accettazione; solo aggiornamenti Markdown, sorgenti e configurazioni invariati, nessuna suite applicativa ripetuta. Nessun commit/push/deploy autorizzato o eseguito in questo turno.

- **REMOVE-01 · 2026-10-17:** il PO richiede il push totale su main dopo la conferma della rimozione. Verificati assenza dei file e dei collegamenti operativi della feature e allineamento HEAD/origin/main (0/0 dopo fetch). Inclusione delle modifiche SSH/firewall presenti esplicitamente autorizzata da «push totale». Evidenze applicative del turno precedente riutilizzate: sorgenti, test e configurazioni invariati; nessuna suite ripetuta dopo soli Markdown. Nessun deploy o pulizia dati applicativi.

- **SSH-TOOLS-01 · 2026-10-17:** PO «Ok perfetto è funzionante». Accettati proxy terminale e lettura/diagnostica iptables. Evidenze tecniche nel [QA SSH](../../docs/05-qa/SSH_TOOLS_QA.md): gate All PASS, 175 test frontend, proxy Vite/Nginx bidirezionale. Esito reale confermato dal PO, non da una sessione agente. Nessun nuovo dettaglio sullo stato pacchetto/sudoers o test watchdog dedotto; nessun commit/push/deploy autorizzato implicitamente.

- **SSH-CONNECT-02 · 2026-10-17:** PO conferma connessione funzionante. Accettato il fix di negoziazione multi-key, senza estendere l'accettazione a terminale/iptables o SSH-HARDEN-01. Evidenze backend e limiti storici nel [piano SSH](../../docs/04-development/plans/SSH-HARDEN-01.md). Nessun commit/push/deploy autorizzato dal feedback.

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

## Evidenze della rimozione autorizzata alla pubblicazione

**REMOVE-01 · Ritiro Penetration Test · 2026-10-17 · working tree.** Su richiesta esplicita del PO rimossi pagina, componenti, API/tipi, handler/service/domain/repository, esecutore e 53 file dedicati, incluse migrazioni sorgente 000011/000012, documenti, piani e ADR della feature. Eliminati task e vecchie attese/decisioni della feature; riprogettazione da definire, nessun requisito precedente assunto valido. Sidebar e azione Findings rimosse. Preservate modifiche SSH/firewall e roadmap restante. Self-review in prima persona, nessun commit/push/deploy.

- Prova PO: avviare la versione ricompilata; sidebar senza voce dedicata e Findings senza «Esegui test». Verifica visuale browser non eseguita.
- Backend dalla root: `scripts/verify.ps1 -Profile Final -Scope All`, cache artifacts dedicata; `go test ./...` 29,737s, `go vet ./...` 3,654s, `go build ./...` 1,790s PASS exit 0. Report `.agent/artifacts/testing/20261017-185904-cf2452f2.json`; runner complessivo exit 1 per esbuild bloccato dalla sandbox, non per test backend falliti.
- Frontend dalla root: `scripts/verify.ps1 -Profile Final -Scope Frontend` con escalation: 167 test/28 file 4,980s, build/typecheck 10,724s, docs 0,626s e diff 0,063s PASS exit 0. Report `.agent/artifacts/testing/20261017-185958-07928d7e.json`.
- PostgreSQL da `backend`: `go test -tags=integration -count=1 ./internal/repository/postgres`, PASS exit 0, 5,551s wall / 4,657s test. Include migrazioni e round-trip SSH down/up. Configurazione TEST_DB_* esplicita su container dedicato `nextgen-removal-test-20261017`, loopback 55441, DB `nextgen_removal_test`; nessun database applicativo toccato. Prima invocazione non partita per path GOCACHE errato, corretta prima della verifica. Preparazione Docker separata dalle durate dei test.
- Doc Sync: ripuliti indice docs; PRD/requisiti/use case/user story; UX/UI; HLD, DOMAIN_MAP, indice ADR e contratti API/schema; TASK_LIST, DEPLOYMENT e questo stato. Aggiornati domini reporting_and_assets e vulnerability_intelligence per navigazione e azione Findings. Nessun link a documenti eliminati.
- Assunzione: ritiro dal repository e dalla build corrente; nessuna cancellazione dei dati persistiti nei database esistenti, nessuna modifica a credenziali/runtime o container applicativi. Il migratore preesistente riallinea il marker se superiore alla versione sorgente 10, senza eliminare tabelle residue. Non introdotto cleanup distruttivo. Categorie storiche Nmap e note metodologiche generali preservate.
- Limiti: nessun collaudo visuale, nessun deploy; un servizio già avviato continua a usare la build precedente fino al riavvio/aggiornamento gestito dal PO. Il PO ha autorizzato esplicitamente commit e push nel turno successivo alla consegna.

## In attesa di accettazione utente

M11-CLOSE-01: checkpoint tecnico, non consegna globale della milestone. Primo
incremento firewall verificato automaticamente (vedere QA); prova su laboratorio
e browser ancora da eseguire. Nessuna promozione a `[~]`/`[x]`. Chiarimento
acquisito dal PO: iptables/iptables-nft, niente nft nativo. Report incrementato
e verificato con 175 test frontend/build PASS, browser/PDF aperti; gap funzionali
residui espliciti nella matrice. Nessun commit,
push o deploy. Verifica manuale futura: nel form aggiunta regola il timer è
obbligatorio, timeout default 60s e intervallo 10..300s; eseguire modifiche solo
nel laboratorio identificato, non su server reali per la prova.

CI-CONTEXT-01: fix env della porta dinamica PostgreSQL verificato con actionlint v1.7.7, rosso exit 1 e verde exit 0 (0,670 s). Nessuna suite applicativa o DB locale rieseguita: nessun sorgente, schema o comando test cambiato. Self-review conferma invariati healthcheck, porta dinamica e gate job Docker. CI remota da verificare dopo pubblicazione. Cache moduli spostata in `.gocache/actionlint-mod` perché il verificatore documentale includeva i Markdown scaricati sotto `.agent/artifacts`; primo controllo documentale FAIL per file terzi, da non confondere con errore dei documenti del progetto.

Gate documentale finale dalla root PASS exit 0 (0,864 s, 224 link/ancore) e `git diff --check` PASS exit 0. Correzione della CI che accompagna la pubblicazione su main già richiesta dal PO; commit previsto `fix(ci): resolve PostgreSQL service port in step context`, nessun deploy. Nessun esito remoto anticipato.

README-GH-01: README GitHub completo in italiano, consegna editoriale richiesta il 2026-10-18. Verifica tramite lettura del README su GitHub; nessuna modifica del comportamento applicativo. Pubblicato con commit `19e84ca`: `git push origin main` PASS exit 0; `git ls-remote origin refs/heads/main` conferma `19e84ca1a2b8f4aacfb6f72f5c3ee929d572c2de`, working tree pulita dopo il push. Questa nota successiva registra l'esito reale; nessun deploy o esito CI remota dichiarato.

Evidenze README-GH-01 dalla root: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify-agent-docs.ps1` PASS exit 0, 1,143 s, 64 Markdown e 224 link/ancore; `git diff --check` PASS exit 0 (con inventario diff/status: 0,148 s). Self-review in prima persona: comandi confrontati con manifest/config, roadmap distinta dall'implementato, nessuna credenziale reale o nuovo termine di licenza. Task `[~]`, accettazione editoriale pendente. Fetch inizialmente negato su FETCH_HEAD dalla sandbox, poi PASS con escalation autorizzata; nessun rifiuto auto-review. Commit `docs(readme): add comprehensive GitHub project guide` pubblicato su main su richiesta del PO.

DOCKER-CI-01 accettato dal PO il 2026-10-17, «OK»; nessuna attesa di accettazione relativa a questo incremento. Prima esecuzione remota ancora da verificare dopo pubblicazione autorizzata; istruzioni in [Deployment](../../docs/06-devops/DEPLOYMENT.md#docker-ci-01--caricamento-artifact).

Nessuna attesa relativa alla feature rimossa; restano validi i gate SSH/Linux descritti nelle rispettive sezioni.

SSH-HARDEN-01: primo incremento TECH-DEBT-06/07 implementato e verificato automaticamente; task `[ ]` per collaudo visuale non eseguito, nessuna dichiarazione di gate Verified. Pin SHA256 host separato dal fingerprint client, connessioni prive di fiducia bloccate, deadline handshake 10s e comandi/test 30s con cancellazione. PostgreSQL isolato e migrazione up/down/up PASS. Browser runtime vuoto, iab indisponibile; nessun collaudo Linux remoto. Piano/evidenze in [SSH-HARDEN-01](../../docs/04-development/plans/SSH-HARDEN-01.md). Nessun commit/push/deploy; modifiche Discovery preesistenti preservate. TECH-DEBT-07 ed E28-S1-T1 restano aperti.

DISCOVERY-01 accettato dal PO il 2026-10-17 con “OK”; evidenze e limiti nell'[analisi](../../docs/04-development/NETWORK_DISCOVERY_REVIEW.md). Nessuna prova LAN reale o benchmark aggiuntivo dedotto dall'accettazione.

## Bloccati / Domande aperte per l'utente

SSH-CONNECT-01/02: il PO conferma ora che la connessione funziona; la richiesta
di confronto TCP precedente sotto è storica e non va ripetuta. Nuova prova
richiesta solo per terminale e firewall dopo attivazione SSH-TOOLS-01.

Aggiornamento SSH 2026-10-17: il nuovo errore handshake riferito dal PO supera
il sintomo timeout di SSH-CONNECT-01 descritto sotto. Nessun nuovo probe TCP
eseguito: SSH-CONNECT-02 corregge il difetto multi-key riprodotto localmente,
senza attribuire automaticamente questa causa al server reale.

- **SSH-CONNECT-01 · 2026-10-17:** check `/ssh-servers` su richiesta PO, errore `ssh dial 192.168.1.18:22 ... i/o timeout`. Service usa TCP dal backend prima del controllo host key; nessuna evidenza che il fingerprint fornito sia la causa. Fuori sandbox: TcpClient Windows timeout a 5s e 12s, `nc -z -v -w 5` nel container backend timeout (route via 172.18.0.1, sorgente 172.18.0.3); Windows 192.168.1.8/24 Ethernet, ping restituisce host irraggiungibile e neighbor .18 Unreachable/MAC nullo. `ssh -G alex@192.168.1.18` conferma hostname e porta 22, nessun proxy riportato. PO riferisce invece accesso via CMD con quel comando: richiesti nuova sessione e Test-NetConnection contemporanei sul PC Docker. Diagnosi aperta, nessuna modifica codice/rete/firewall, credenziale letta o autenticazione remota. Prove iniziali in sandbox negate, non confuse con l'esito dei probe estesi. Nessuna suite applicativa necessaria per questa diagnosi senza cambi codice; nessun PASS connessione o causa definitiva.

- SSH-HARDEN-01: inventario browser vuoto e apertura iab fallita; nessuna prova visuale/screenshot inventata. Assunzione esplicita: pin fornito dall'operatore tramite un canale indipendentemente fidato, inventario senza pin conservato ma non connettibile. La transazione watchdog sull'host e il relativo collaudo Linux sono ancora lavoro necessario per TECH-DEBT-07.

- **Limite visuale COMPANY-UX-02 / COMPANY-PDF-03:** collaudo autonomo desktop/mobile e PDF non eseguito, browser e renderer non disponibili. Prove automatiche/DB concluse; il PO accetta gli incrementi con questo limite esplicito e autorizza pubblicazione. Nessun PDF autonomo consegnato come verificato visualmente, nessuna domanda funzionale pendente.

- **Limite visuale AP-REVIEW-01 accettato dal PO:** inventario browser vuoto, verifica grafica autonoma e screenshot non eseguiti; DOM e interazioni coperte da test. I dati locali privi di CVE non permettono validazione reale di catene di exploit; scenari sintetici coprono il modello. Nessun blocco alla pubblicazione richiesta.

- Il PO conferma il 2026-10-17 che non è ancora disponibile un host Linux di laboratorio e chiede di mantenere aperta la verifica remota E28-S1-T1. Prima dell'uso operativo risolvere anche TECH-DEBT-06/07: host-key non verificata, chiamate SSH senza deadline e coordinamento watchdog. Nessuna azione remota eseguita.

- Nessun blocco sulla consegna scanner: il PO ha eseguito i test delle scansioni e confermato il corretto funzionamento il 2026-10-16. Verifica browser autonoma e benchmark ETA non effettuati dall'agente; debito TECH-DEBT-05 sul consenso API mantenuto.

## Assunzioni confermate

- DOCKER-CI-01 (2026-10-17): artifact per 7 giorni, Linux amd64 su Docker Desktop Windows in modalità Linux, PostgreSQL scaricato separatamente, nessun registry/deploy; accettate con «OK». Nessun eseguibile Windows incluso.

- **DISCOVERY-01 (2026-10-17):** accettata la scelta conservativa di non recuperare identità storiche senza MAC corrente verificabile. Limiti OUI, hostname di presentazione e assenza di benchmark/LAN reale dichiarati nella consegna preservati. Suite backend/vet/build PASS precedenti restano valide: solo documentazione modificata dopo il gate.

- Struttura documentale SDLC v7.1 e orchestrazione per feature approvate dal Product Owner; anche la migrazione base Codex AGENT-CODEX-01 e la pulizia AGENT-CLEAN-03 sono approvate.
- `docs/` contiene solo documentazione necessaria agli agenti; manuali utente, screenshot, allegati binari e archivi sono esclusi, come richiesto dal PO.
- Storico consolidato nel [CHANGELOG.md](../../CHANGELOG.md); stub e cartella legacy rimossi dopo la migrazione alle fonti canoniche.
- Pipeline CI/CD GitHub Actions e bundle splitting implementati.
- Milestone 2 completata sia su Backend Go che su Frontend React.
- Roadmap di hardening sicurezza (Epiche SEC-P0, SEC-P1, SEC-P2) implementata al 100%.

## Decisioni recenti (ultime 10, dettagli negli ADR)

- **ADR-20 (2026-10-17):** pin host SHA256 esplicito e distinto dalla chiave client, fail closed per legacy senza pin/mismatch, rotazione e reset endpoint; deadline SSH con cancellazione. Migrazione 000010 up/down preserva inventario e perde solo i pin. Nessuna garanzia di terminazione remota dedotta dal timeout locale.

- **ADR-19 (2026-10-17):** test mirati per iterazione, un gate finale condiviso per lato coinvolto, riuso di sole evidenze valide, misure per comando. PostgreSQL esplicito con TEST_DB_* senza fallback applicativo, host locale e DB dedicato; indisponibilità = FAIL. CI con job DB isolato. Decisione richiesta dal PO, non deroga estemporanea.

- **ADR-18 (2026-10-17):** dashboard unica, guard aziendale con eccezione tool SSH, clienti SSH autonomi, gerarchia due livelli nelle nuove scritture, report/IP per reparto e selettore reparto AP senza soglia UI.

- **ADR-17 (2026-10-17):** Attack Path come analisi euristica verificabile con evidenze/ipotesi distinte, percorsi su archi reali del modello, score non probabilistico, scope aziendale completo e guida UI 2D. Nessuna inferenza di trust dall'inventario SSH. Gate visuale separato.

- **ADR-15 (2026-10-16):** telemetria XML Nmap, snapshot attivi e lettura batch per monitor lotto; API additive, nessuna migrazione DB, ETA esplicitamente euristica. Durate indicative richieste dal PO dopo la precedente rimozione delle stime promozionali.

- **ADR-13 (2026-10-15):** Documentazione `docs/` riservata agli agenti; rimossi manuale e screenshot, evidenze temporanee fuori da `docs/`, controllo strutturale automatizzato.
- **ADR-14 (2026-10-15):** profili espliciti + editor Nmap entro allowlist, API/DB compatibili, matching argomenti/timeout e classificazione server dei custom. Discovery rispetta l'opzione porte; limiti dei controlli documentati nel dominio scanner.
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

0. SSH-HARDEN-01 implementato con verifiche automatiche PASS; completare collaudo visuale, quindi accettazione distinta. Proseguire TECH-DEBT-07 con transazione sull'host e test Linux/multiistanza; successivamente TECH-DEBT-05. Nessuna chiusura globale del registro debiti.

1. Modifiche approvate pubblicate su main con fast-forward e push riuscito (`26cf93d`); registro documentale di consegna aggiornato. Accettazione PO acquisita, limiti visuali residui registrati. Nessun nuovo branch o deploy.
2. E28-S1-T1 resta separato: hardening SSH/watchdog TECH-DEBT-06/07 e prova Linux quando disponibile.
3. Proseguire Epic 28 secondo dipendenze e priorità PO; nessun push/merge/deploy implicito della revisione AP.

## Log sessioni (ultime 5, la più recente in alto)

- 2026-10-17 · Pubblicazione richiesta dal PO: feedback/accettazione e commit.md, ricontrollo rimozione e fetch origin PASS, main inizialmente allineato. Push totale autorizzato include SSH/firewall preesistenti; gate applicativi precedenti validi perché input invariati, documenti/staging ricontrollati. Commit `4f9ea78`, 25 file; `git push origin main` PASS exit 0, remoto aggiornato da `3af97df` a `4f9ea78`. Staging senza file sensibili o firme credenziali. Nessun deploy, force push o modifica ai database applicativi. Questa nota successiva registra l'esito reale della pubblicazione.

- 2026-10-17 · REMOVE-01: feedback/modifica → feature/orchestration; rimozione integrale richiesta dal PO, 53 file dedicati eliminati, documenti/task/riferimenti condivisi ripuliti. Working tree preesistente SSH preservata. Backend suite/vet/build, frontend 167 test/build, PostgreSQL dedicato e docs/diff PASS exit 0 per i rispettivi comandi; blocco esbuild sandbox risolto con escalation. Self-review e Doc Sync, nessun target/DB applicativo/deploy/commit/push. Riprogettazione da definire; consegna in attesa del PO.

- 2026-10-17 · SSH-TOOLS-01: feedback/accettazione «Ok perfetto è funzionante». Task `[x]`, stato/QA/dominio allineati e attesa rimossa. Risoluzione confermata dal PO, causa operativa precisa non dedotta. Solo Markdown, suite applicative non rieseguite; gate watchdog/Linux separati preservati, nessun commit/push/deploy.

- 2026-10-17 · SSH-TOOLS-01: feedback/correzione e accettazione connessione. Proxy WS mancanti trovati, lettura firewall non interattiva e messaggi API corretti. Red/green service/UI, proxy reale echo Vite/Nginx, gate All PASS exit 0 con 175 test UI. Self-review e Doc Sync; nessun DB/target modificato, privilegio concesso, nuova dipendenza o deploy. Prova Debian al PO; blocchi sandbox Vite/Docker risolti con escalation, nessun rifiuto auto-review. Primo riferimento immagine smoke obsoleto e argomento runner Quick non valido corretti, senza attribuirli ai test applicativi.

- 2026-10-17 · PO “Procedi” sul rientro debiti: feedback→feature/orchestration, incremento SSH-HARDEN-01 (TECH-DEBT-06 e deadline di 07). Discovery preesistente preservata, nessun branch/commit/push/deploy. Test pin assente rosso→verde; pinning, rotazione/reset, deadline/cancellazione, migrazione 000010 e modale. Gate finale All+Integration PASS exit 0: backend suite 21,260s/vet 0,796s/build 1,353s; frontend 166 test 10,865s/build 14,767s; DB 4,905s incluso up/down/up. Docker solo test dedicato; nessuna migrazione applicativa. Vite inizialmente bloccato dalla sandbox, riuscito con escalation; evento submit esplicito nel test DOM dopo controllo validità. Self-review e Doc Sync; browser vuoto/iab assente: task `[ ]`, nessun PASS visuale o Linux. TECH-DEBT-07 richiede ancora transazione remota/coordinamento e collaudo; altri debiti non modificati.

- 2026-10-17 · DISCOVERY-01: feedback/modifica e orchestration, working tree iniziale pulita. Riprodotti MAC invalidi e falsi vendor software, corretti e verificati con gate finale backend PASS; test cache cambio MAC, /31-/32 e HTTP locali PASS. GOCACHE locale negli artifacts per limite cache di sistema. Doc Sync e self-review; limiti OUI/hostname/ARP e pool HTTP futuro documentati, nessuna scansione cliente o modifica remota.
