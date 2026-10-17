# COMPANY-UX-02 · Evidenze e accettazione

Data 2026-10-17, working tree. Modifiche AP-REVIEW-01 preesistenti preservate, non accettate implicitamente. Nessun nuovo commit/push/deploy.

## Incremento COMPANY-PDF-03 · CVE per reparto

Implementati raggruppamento delle schede tecniche complete e top cinque priorità Executive per ogni reparto, mantenendo riepilogo aziendale. Per reparti successivi e continuazioni si gestiscono pagina/intestazione. Target diretti in “Senza reparto”; reparti vuoti indicati senza dichiarazione di assenza vulnerabilità. Identità basata su gruppo/target/scan, non solo IP o nome visualizzato: un vero reparto chiamato “Senza reparto” resta distinto dall'etichetta di fallback. Aziende senza reparti e download singola scansione/reparto mantengono il flusso non raggruppato.

Rischio locale backend; nessuna query repository, migrazione, API pubblica, frontend o dipendenza modificata in questo incremento. Nessuna integrazione DB eseguita né PASS DB dedotto dai test unitari. Scenario service con repository simulati e generatore reale in memoria verificati separatamente; nessun E2E visuale dichiarato.

| Directory | Comando / evidenza | Esito e durata |
|---|---|---|
| backend | `go test ./pkg/report ./internal/service -run 'TestCompanyReportsInclude\|TestCompanyPDFFindings\|TestDepartmentFinding' -count=1` | PASS exit 0, 1,976 s comando; due template, IP uguali, sanitizzazione, sezioni vuote, limite Executive locale e paginazione. Primo tentativo FAIL per separatore stream nel parser di test, corretto senza cambiare asserzioni. |
| root, `GOCACHE=backend/.gocache` | `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1 -Profile Final -Scope Backend` | PASS exit 0. Suite backend 16,296 s, vet 1,034 s, build 1,340 s; documenti 0,587 s e diff 0,061 s. Include nuovi test service per target diretti, collisione etichetta e azienda senza reparti. |

Evidenza locale ignorata da Git: `.agent/artifacts/testing/20261017-091358-c1a52cac.json`. Test PDF ispezionano gli stream effettivi generati (ordine intestazioni/CVE) e la paginazione; non sono rendering visuale. Self-review dello stesso agente: scope tenant preservato, nessuna attribuzione per solo IP, filtro finding invalidi coerente nei due flussi, copia degli slice prima di ordinare, nessuna omissione tecnica dovuta al top cinque Executive. Gate renderer aperto `[!]`.

Prova PO: in `/reports`, scegliere l'azienda con reparti e scaricare i PDF **Tecnico** e **Executive** dell'intera azienda. Atteso riepilogo unico e schede “Reparto: …”; tutte le CVE nel Tecnico, prime cinque priorità di ciascun reparto nell'Executive con totale e rinvio al Tecnico. Verificare anche un target diretto aziendale, un reparto vuoto e lo stesso IP in reparti diversi. Per azienda senza reparti: schede senza suddivisione artificiale. Assunzione: Executive resta una sintesi, non un elenco esaustivo dei finding.

Documenti sincronizzati nell'incremento:

| Documento | Motivo |
|---|---|
| [Dominio](../domains/reporting_and_assets.md) | Contratto interno e attribuzione finding per reparto. |
| [Requisiti](../01-product/REQUIREMENTS.md) | Atteso PDF aziendale con CVE per reparto. |
| [Design](../02-design/UI_SPECIFICATIONS.md) | Sezioni, continuazioni e differenza Executive/Tecnico. |
| [Piano](../04-development/plans/COMPANY-UX-02.md) | Scope backend, firme, rischi e strategia di test. |
| [Task](../04-development/TASK_LIST.md) | Evidenze e gate renderer aperto. |
| Questo report QA | Scenari, risultati, limiti e procedura di accettazione. |
| [Stato progetto](../PROJECT_STATE.md) | Memoria, assunzione Executive e prossima verifica. |

## Risultato

Dashboard con CRUD aziende e metriche aggregate, redirect route aziende precedente, obbligo azienda per route operative. SSH/Firewall indipendente con campo cliente autonomo, filtro e gruppi interni. Reparti massimi due livelli nelle nuove scritture UI/API/DB. Attack Path senza filtro CVSS e con selettore reparto. Dossier aziendali includono i reparti e inventario IP separato, CSV reparto additivo; download reparto preservato.

Corretto anche il difetto preesistente del filtro report per sola radice, che ometteva i reparti; mantenuti IP uguali su target/reparti differenti. Target senza finding inclusi nell'inventario PDF. Errori lettura finding propagati, letture consolidate paginate. CompanyContext aggiornato dopo CRUD e paginato; niente default globale per reparto vuoto.

## Comandi ed esiti

| Directory / contesto | Comando | Risultato |
|---|---|---|
| backend | `go test ./internal/service -run TestCompanyHierarchyRejectsNestedDepartments -count=1` prima del fix | FAIL exit 1, sotto-reparto accettato. |
| backend | `go test ./internal/service -run 'Company\|TargetGroup\|Report' -count=1` dopo fix | PASS exit 0; fixture legacy introdotte direttamente nel repository mock per preservare test di lettura ricorsiva. Nessuna asserzione indebolita. |
| backend | test mirati `Company`, `Department`, `SSHClientLabel`, `TargetGroup` in service/pkg report | PASS exit 0: due reparti con IP uguale, target senza finding, download reparto isolato, CSV e PDF multipagina in memoria; update SSH omesso/vuoto/overflow. |
| frontend | `npm.cmd test` | PASS exit 0: 26 file, 164 test. Route guard (assenza/stale/loading/valida/SSH), reparto AP vuoto/CIDR, nessun filtro CVSS, gerarchia editor e filtro clienti SSH inclusi. |
| frontend | `npm.cmd run build` | PASS exit 0, TypeScript e Vite. Eseguito fuori sandbox per esbuild. |
| backend, DB_NAME test esplicito | `go test ./...`, `go build ./...`, `go vet ./...` | PASS exit 0. Test versione massima migrazione aggiornato da 7 a 9, senza allentare il controllo. |
| backend, DB_NAME test esplicito | `go test ./internal/repository/postgres -run 'TestRunMigrationsIdempotence\|TestSSHServerRepo_Integration' -v -count=1` | PASS exit 0: migrazioni applicate e idempotenti, CRUD repository reale. Successiva suite completa comprende asserzioni persistenza/aggiornamento cliente. |
| PostgreSQL di test | down 000009/000008, verifica assenza colonna, up 000008/000009, verifica colonna/trigger, ROLLBACK | PASS exit 0. |
| PostgreSQL di test | creazione root/reparto, tentativo sotto-reparto e spostamento root con figli, ROLLBACK | PASS exit 0: entrambi rifiutati dal trigger, nessun dato sintetico conservato da questa prova. |

DB dedicato creato: `nextgen_company_ux02_test_20261017`. Tutte le mutazioni di verifica sono circoscritte a quel database, mantenuto disponibile per ripetere i test. Database applicativo solo letto per contare sotto-reparti esistenti (0); schema e dati applicativi non modificati. Non sono state lette credenziali né contattati server dei clienti.

## Gate visuali aperti

Inventario computer-use vuoto (`apps: [], browsers: []`). Renderer PDF assente sia sul sistema sia nel container backend verificato. Test PDF in memoria provano contenuto e paginazione, non allineamento visivo; nessun PDF finale o screenshot dichiarato verificato. La skill PDF richiede rendering e ispezione PNG prima di consegnare un artefatto PDF: non viene consegnato un PDF autonomo. Il task rimane `[!]` fino a prove visuali UI/PDF.

Per provare: avviare la versione aggiornata con `docker compose up --build` (applica le migrazioni al database configurato), aprire `/`, creare/selezionare azienda e reparti. Con scope globale provare URL `/targets` e `/attack-path` (redirect), `/ssh-servers` (accessibile). Verificare CRUD azienda in dashboard, soli reparti diretti, scansioni separate, selettore reparto AP, download PDF azienda/reparto e inventario IP. Provare su desktop/mobile. Nessuna nuova scansione remota richiesta per i soli controlli di navigazione/anagrafica.

## Assunzioni, rischi e self-review

- Clienti SSH: etichette autonome per tenant, non nuove ACL per cliente; campi vuoti ammessi per server esistenti. Confermato dal PO.
- Reparto AP limita l'intera catena al reparto; intera azienda mantiene ipotesi tra reparti. Parametro backend min_cvss preservato, rimosso solo dalla UI.
- PDF consolidato ora può superare due pagine per non troncare l'inventario. I report scansione restano invariati. Rollback migrazione 000008 perde etichette cliente: esportarle prima dell'uso operativo.
- Review svolta dallo stesso agente su tenant scope, query parametrizzate, output cliente senza segreti, trigger/cicli, report cross-reparto, CSV formula nel campo aggiunto, no nuove dipendenze. Guard navigazione distinto dall'autorizzazione API. Restano limiti AP euristici e gate SSH Linux documentati separatamente.

## Documenti sincronizzati

Verifica documentale finale: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify-agent-docs.ps1` PASS exit 0 (24 obbligatori, 61 Markdown, 160 link/ancore); `git diff --check` PASS exit 0. Diff unstaged/staged/status ispezionati; nessun file staged, segreto o nuova dipendenza. Modifiche AP preesistenti riconosciute e preservate.

- Domini asset/report, SSH, scanning e Attack Path: nuovi comportamenti e confini.
- Requisiti, use case, UI: gerarchia, dashboard unica, selettori e flusso operativo.
- HLD, ADR-18, API, schema, DOMAIN_MAP: architettura, compatibilità client_name, migrazioni e percorsi reali.
- Deployment: ordine upgrade e rollback, database di test.
- Piano COMPANY-UX-02, TASK_LIST, TEST_PLAN e PROJECT_STATE: verifiche, stato e gate residui.

| File | Motivo |
|---|---|
| `docs/domains/reporting_and_assets.md` | Contratto aggiornato dashboard, gerarchia e dossier reparto/azienda. |
| `docs/domains/ssh_and_firewall.md` | Indipendenza e client_name autonomo/compatibile. |
| `docs/domains/attack_path.md` | Rimozione filtro UI e proiezione reparto. |
| `docs/domains/scanning_and_discovery.md` | Scope azienda obbligatorio e reparti diretti. |
| `docs/01-product/REQUIREMENTS.md` | FR-09/15/18 allineati alla richiesta. |
| `docs/01-product/USE_CASES.md` | UC-14 con flusso unico ed eccezione SSH. |
| `docs/02-design/UI_SPECIFICATIONS.md` | Stati, navigazione e nuovi controlli. |
| `docs/03-architecture/HLD.md` | Confini fra aziende e tool SSH. |
| `docs/03-architecture/ADR/README.md` | ADR-18 e tradeoff multipagina/compatibilità. |
| `docs/03-architecture/DOMAIN_MAP.md` | Copertura route guard, modal, SSH domain/types. |
| `docs/03-architecture/LLD/API_SPEC.md` | Campi additivi, validazioni e API report compatibili. |
| `docs/03-architecture/LLD/DATABASE_SCHEMA.md` | Migrazioni 000008/000009 e rollback. |
| `docs/06-devops/DEPLOYMENT.md` | Upgrade, perdita etichette nel down e DB test. |
| `docs/04-development/plans/COMPANY-UX-02.md` | Intake, contratti, chiarimento PO e test. |
| `docs/04-development/TASK_LIST.md` | Stato gate e prossimo ID migrazione libero. |
| `docs/05-qa/TEST_PLAN.md` | Indice delle nuove evidenze. |
| `docs/05-qa/COMPANY_WORKFLOW_QA.md` | Pacchetto verifiche e prova guidata. |
| `docs/PROJECT_STATE.md` | Memoria, assunzioni, risultati e blocchi. |
