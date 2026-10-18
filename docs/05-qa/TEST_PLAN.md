# Test Plan

Ultimo allineamento governance: 2026-10-18 · working tree · AGENT-HARNESS-01. Evidenze applicative datate sotto non rieseguite da questa revisione.

COMPANY-UX-02 · 2026-10-17: [evidenze flussi aziendali e clienti SSH](COMPANY_WORKFLOW_QA.md); 164 test frontend, suite backend e integrazioni su DB test separato, migrazioni up/down e trigger. Gate browser e rendering PDF aperti, nessuna validazione visuale presunta.

AP-REVIEW-01 · 2026-10-17: [report di revisione Attack Path](ATTACK_PATH_REVIEW.md), con difetti riprodotti, correzioni, test sintetici e snapshot locale in sola lettura. Gate browser aperto; non dichiarare verificato il layout senza screenshot reali.

## E28-S1-T1 · 2026-10-17 · working tree

Test aggiunti in `defense_service_test.go`, `defense_handler_test.go` e
`firewall_service_test.go`: regola deterministica, TCP/UDP, input non supportati,
protezione porta SSH, mismatch IP e tenant, chokepoint obsoleto, preview senza
mutazione, inoltro timer obbligatorio, JSON/RBAC/errori HTTP, PID watchdog invalido,
concorrenza preparazione timer, conservazione watchdog su conferma/ripristino fallito.
Riproduzione rossa PID exit 1, verde exit 0. Test di servizio/handler exit 0.
Suite backend, build e vet exit 0; frontend 145 test e build exit 0.
Race detector non eseguito: toolchain senza CGO (comando exit 1).
Prova SSH/iptables reale non eseguita: nessun host Linux di laboratorio identificato.
Limiti e piano nel [piano E28-S1-T1](../04-development/plans/E28-S1-T1-chokepoint-remediation.md).

> Responsabile: QA Engineer. Strategia preventiva; risultati eseguiti e baseline sono in [PROJECT_STATE](../PROJECT_STATE.md).

## Strategia

Ultimo allineamento: 2026-10-17 · working tree · TEST-OPT-01. Politica vincolante: `AGENTS.md` sez. 8; decisione [ADR-19](../03-architecture/ADR/README.md#adr-19).

Policy di scope, gate finale e riuso: [AGENTS sez. 8](../../AGENTS.md). La [skill repo-verification](../../.agents/skills/repo-verification/SKILL.md) seleziona il profilo; qui risiedono comandi, ambiente e strategia dei test. Report del runner misurano esiti e durate, non certificano automaticamente la freschezza.

Per ogni feature collegare FR/UC/storia → criterio osservabile → test → esito. Coprire happy/error/edge, isolamento tenant, autorizzazioni, input ostili e lifecycle risorse nei percorsi pertinenti. Per bug dimostrare riproduzione rossa e test verde; per refactor fissare prima gli invarianti.

| Livello | Comandi e scope | Evidenza richiesta |
|---|---|---|
| Backend | Durante sviluppo: `go test ./internal/<package> -run <pattern>`; finale backend: `go test ./...`, `go vet ./...`, `go build ./...` | Durata, exit code e package; nessuna connessione DB nel percorso rapido |
| Frontend | Durante sviluppo: `npm test -- src/<file>.test.tsx`; finale frontend: `npm test`, `npm run build` | Durata, exit code, file/test e typecheck; la build include già `tsc` |
| PostgreSQL | Da `backend`: `go test -tags=integration -count=1 ./internal/repository/postgres` | DB dedicato locale, parametri `TEST_DB_*` espliciti, zero skip; DB assente = FAIL |
| Integrazione/E2E | API, DB di test, migrazioni up/down, flussi browser interessati | Ambiente, dati, output e screenshot reali per UI |
| Documentazione | `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify-agent-docs.ps1`, confronto dei contenuti migrati, `git diff --check` | Esito e perimetro; nessuna suite applicativa per soli Markdown e verificatore |

Lo script `scripts/e2e_test.ps1` esiste ma esegue operazioni sulla piattaforma: usarlo solo dopo avere identificato destinazione e credenziali di test. Un test saltato internamente non prova l'integrazione con un servizio reale.

### Runner e misurazioni

Dalla root, `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1` supporta:

- `-Profile Quick -Scope Backend -BackendPackages ./internal/service`: solo test mirati, nessuna build finale.
- `-Profile Quick -Scope Frontend -FrontendFiles src/pages/Dashboard.test.tsx`: solo file selezionati esistenti.
- `-Profile Final -Scope Backend` oppure `-Scope Frontend`: suite completa del lato, build/vet pertinenti e controlli documentali. Per DB aggiungere `-Integration` con scope Backend/All.
- `-Profile Full`: entrambi i lati, integrazioni DB e documenti; corrisponde a `/verify` senza scope.
- `-Profile Docs`: verificatore Markdown e `git diff --check`.

Ogni esecuzione genera `.agent/artifacts/testing/<run-id>.json` con scope, comando, durata ed exit code, anche in caso di fallimento; il runner si ferma al primo errore. Nessun log di variabili d'ambiente o credenziali nel report. I comandi non raggiunti non sono PASS. Avvio servizi/installazione dipendenze vanno misurati separatamente. Per confronti prestazionali distinguere cache calda/fredda, stesso hardware, ambiente e scope; non stimare percentuali da una sola misura. La selezione dei test segue anche gli utilizzatori del codice, non soltanto i file cambiati.

Le integrazioni sono escluse dal percorso rapido mediante build tag `integration` (non skip). Richiedono `TEST_DB_HOST` (`127.0.0.1` o `localhost`), `TEST_DB_PORT`, `TEST_DB_NAME` (suffisso `_test` oppure `_test_<identificatore>`), `TEST_DB_USER`, `TEST_DB_PASSWORD`. Non leggono `.env` né ereditano `DB_*`; usare solo un database dedicato sacrificabile. Non parallelizzare migrazioni o fixture sul medesimo database. La CI fornisce un PostgreSQL 16 effimero con health check; flussi browser, rendering PDF e rollback non sono coperti automaticamente da questo runner.

Test nuovi solo per rischi/comportamenti distinti; preferire sincronizzazione esplicita alle attese fisse quando il profiling ne giustifica il costo, conservando test reali di timeout/cancellazione. Nessun limite di tempo va abbassato per nascondere errori. Cache Go preservata; integrazioni con `-count=1` perché dipendono dallo stato esterno.

Salvare le evidenze visuali temporanee in `.agent/artifacts/<task-id>/` (ignorato da Git). In `docs/` restano strategia, criteri e risultati testuali, senza screenshot o allegati binari.

## Governance e skill

AGENT-HARNESS-01: `scripts/test-agent-harness.ps1` esegue fixture isolate per glob, copertura/ambiguità, manifest, limiti dello stato e inclusione/esclusione dei documenti. `scripts/verify-agent-docs.ps1` controlla il repository corrente. [Piano e limiti](../04-development/plans/AGENT-HARNESS-01.md); nessuna attestazione di qualità/routing Sol da un PASS strutturale.

## Test case per feature

### TEST-OPT-01 · 2026-10-17 · working tree

Comandi applicativi eseguiti con permessi per cache Go/esbuild; il primo tentativo in sandbox è fallito prima dei test per accesso cache, non è una regressione. Toolchain locale Go 1.27.2 / Windows; workflow remoto Go 1.22 / Ubuntu non eseguito e non dichiarato PASS. Nessuna dipendenza aggiunta.

| Directory / comando | Esito reale | Durata osservata |
|---|---|---|
| Root: `scripts/verify.ps1 -Profile Final -Scope Backend -Integration` con TEST_DB_* su container dedicato | PASS exit 0; suite/vet/build/DB/docs/diff | Ultimo report: suite cache 0,856 s; vet 0,381 s; build 1,110 s; DB senza cache 4,635 s |
| Backend: `go test ./...` (prima esecuzione del runner) | PASS exit 0; PostgreSQL escluso, cache parziale | 21,577 s; package service 16,844 s |
| Backend: `go test -tags=integration -count=1 -v ./internal/repository/postgres` | PASS exit 0; repository/migrazioni idempotenti/configurazione (12 casi), nessuno skip | Package 5,292 s nell'esecuzione dopo ampliamento casi |
| Root: `scripts/verify.ps1 -Profile Quick -Scope Frontend -FrontendFiles src/utils/debugLogger.test.ts` | PASS exit 0; 1 file / 30 test | Runner 16,818 s; asserzioni 13 ms; startup Vitest/ambiente/setup prevalenti |
| Root: `scripts/verify.ps1 -Profile Quick -Scope Backend -BackendPackages ./internal/repository/postgres` | PASS exit 0; nessun DB | 0,272 s, cache calda |
| Backend: `go test -tags=integration -count=1 -run '^TestPostgresConnectionAndPing$' ./internal/repository/postgres` senza TEST_DB_* | FAIL atteso exit 1; configurazione obbligatoria, nessuna connessione | Test fallisce immediatamente |
| Stesso comando dopo arresto del solo container di test e con TEST_DB_* esplicite | FAIL atteso exit 1; nessuno skip | Test 8,03 s per retry reali |

Runner: quattro selezioni/scope invalidi rifiutati prima dei comandi (exit 1, report con zero check); Full senza TEST_DB_* rifiutato nel preflight (exit 1, zero check); package inesistente propaga exit 1 e produce report con un solo check fallito. Harness di verifica di questi risultati: PASS exit 0. Profilo Docs e `git diff --check`: PASS exit 0.

Evidenze temporanee: `.agent/artifacts/testing/20261017-090423-52365c16.json` (finale backend+DB), `20261017-090216-b2b42f59.json` (frontend smoke) e altri report nella medesima directory per i casi negativi. Il gate finale è stato ripetuto solo dopo modifica del runner/preflight e ampliamento dei test di configurazione; nessuna suite rieseguita per il successivo Doc Sync. Confronti a cache diversa non dimostrano un guadagno dell'ottimizzazione.

Self-review su separazione build tag, assenza fallback DB applicativo, fail-fast/exit code, validazione percorsi, report senza variabili sensibili e isolamento CI. Nessuna review indipendente dichiarata. Il container `nextgen-testing-opt-20261017` è stato creato da immagine PostgreSQL 16 già disponibile, verificato e arrestato/rimosso; non sono stati toccati servizi/dati applicativi. Nessuno schema o migrazione modificati dal task: rollback N/A. Frontend applicativo invariato: suite/build complete non rieseguite, smoke limitato al percorso del runner. Browser/PDF non pertinenti a questa modifica. Le misure non giustificano ancora refactor delle attese né una percentuale di risparmio; la policy riduce le esecuzioni complete ripetute durante le prossime implementazioni.

Accettazione scanner 2026-10-16: il PO conferma di avere testato le scansioni e ne attesta il corretto funzionamento, autorizzando il push. SCAN-UX-01/SCAN-TIME-01 sono accettati; i NON ESEGUITI sotto riguardano le prove autonome dell'agente e non vengono trasformati in PASS inventati. Verifica pre-push ripetuta: frontend 145 test, backend suite/build/vet e build frontend PASS, exit 0. Nessun benchmark di precisione ETA dichiarato.

### SCAN-TIME-01 · 2026-10-16

Preset: copertura di otto intervalli positivi, ordinati e distinti dal timeout, ipotesi visibili. Runner: XML frammentato, attributi mancanti/invalidi, buffer limitato, flag interno prima del target, RawXML preservato. Orchestratore: percentuali pesate non regressive, niente ETA prematura, snapshot tenant e cleanup/fase senza ETA. Handler: ID validi/duplicati/invalidi/vuoti/eccessivi e 404 tenant estraneo.

Frontend: progresso solo membri del lotto, coda, errori/annullamenti distinti, telemetria stale/esaurita, gruppi oltre 50 job, retry non sovrapposto, stop dei terminali/teardown, ripristino sessione senza storico estraneo. Dettaglio: barra accessibile, coda senza ETA, terminali. Test automatici con fixture e processo helper, nessuna scansione reale. Browser inventory del 2026-10-16 vuoto, gate visuale non eseguibile. Risultati finali in [SCAN_PROGRESS_PLAN](../04-development/SCAN_PROGRESS_PLAN.md).

ID, FR/UC, precondizioni/dati, passi, risultato atteso, livello, comando, esito (PASS/FAIL/NON ESEGUITO) ed evidenza. L'assenza di un ambiente obbligatorio impedisce il gate verified.

## Bug report

### SCAN-UX-01 · 2026-10-15 · working tree

- FR-16/FR-17: selezione preset, argomenti identici a preview/payload, passaggio all'editor, opzioni ostili/porte/intervalli/tecniche incompatibili, consenso invalidato all'edit, focus/Escape, scope aziendale assente, blacklist, paginazione, errori remoti e retry scheduler parziale.
- Backend: parser inizialmente rosso per 15 combinazioni incoerenti, poi verde; Discovery inizialmente rosso con include_common_ports=false (2 port scan invece di 0), poi verde in entrambe le modalità. Classificazione custom testata sulla persistenza nel repository fake.
- Frontend finale: `npm.cmd test` PASS, 19 file / 136 test, exit 0; `npm.cmd run build` PASS, exit 0, TypeScript + Vite (7,30 s).
- Backend finale: `go test ./...`, `go build ./...`, `go vet ./...` PASS, exit 0. Eseguiti da backend, con permessi estesi per cache Go; frontend da frontend, permessi estesi necessari a esbuild. Nessun test indebolito o skip aggiunto.
- UI reale e Nmap/container end-to-end: NON ESEGUITI. Browser inventory vuoto e iab indisponibile; nessun laboratorio target autorizzato identificato. Test automatici usano mock/fixture, non provano accuratezza CVE su rete reale.
- Self-review (stesso agente): input/output gestiti dal server, niente shell/target inline, equivalenza profili, scope, metadati invasività, consenso e retry. Rischio residuo FR-01.3 registrato in TECH-DEBT-05; nessuna review indipendente dichiarata.
- Esiti documentali e riepilogo in [analisi scanner](../04-development/NMAP_SCAN_REVIEW.md). Gate visuale ancora bloccato; SCAN-UX-01 resta [!].

ID/task, severità, ambiente/versione, atteso, osservato, passi minimi, evidenza sanificata, test di riproduzione, causa radice, fix e regressione.

## Gate finale

Test pertinenti e regressione verdi; build/typecheck/vet verdi; nessun blocco critico; Doc Sync; review del diff. La review è indipendente quando un secondo agente è disponibile e autorizzato; altrimenti dichiarare self-review e relativi limiti.
