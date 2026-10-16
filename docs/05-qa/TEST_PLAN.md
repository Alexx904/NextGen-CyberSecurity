# Test Plan

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

Per ogni feature collegare FR/UC/storia → criterio osservabile → test → esito. Coprire happy/error/edge, isolamento tenant, autorizzazioni, input ostili e lifecycle risorse nei percorsi pertinenti. Per bug dimostrare riproduzione rossa e test verde; per refactor fissare prima gli invarianti.

| Livello | Comandi e scope | Evidenza richiesta |
|---|---|---|
| Backend | Da `backend`: test package, poi `go test ./...`, `go vet ./...`, `go build ./...` | Exit code e package; indicare skip e prerequisiti reali |
| Frontend | Da `frontend`: test file, poi `npm test`, `npm run build` | Exit code, file/test e typecheck |
| Integrazione/E2E | API, DB di test, migrazioni up/down, flussi browser interessati | Ambiente, dati, output e screenshot reali per UI |
| Documentazione | `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify-agent-docs.ps1`, confronto dei contenuti migrati, `git diff --check` | Esito e perimetro; nessuna suite applicativa per soli Markdown e verificatore |

Lo script `scripts/e2e_test.ps1` esiste ma esegue operazioni sulla piattaforma: usarlo solo dopo avere identificato destinazione e credenziali di test. Un test saltato internamente non prova l'integrazione con un servizio reale.

Salvare le evidenze visuali temporanee in `.agent/artifacts/<task-id>/` (ignorato da Git). In `docs/` restano strategia, criteri e risultati testuali, senza screenshot o allegati binari.

## Test case per feature

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
