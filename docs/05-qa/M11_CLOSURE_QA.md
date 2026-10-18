# M1.1 · Matrice di chiusura

Ultimo allineamento: 2026-10-18 · working tree. Current outcome: M1.1 `[~]`,
implemented and verified; PO acceptance separate. The matrix owns current gates;
sections below retain historical failures/environment limits for each increment.

## Criteri e gap

| Requisito | Evidenza presente | Gap da chiudere |
|---|---|---|
| FR-07 alive-only | Browser real QA range .3–.5: 2 up/3 total, 1 offline excluded; one open service .4:22, open-port toggle consistent. CSV one service, PDF two A4 pages rendered/inspected with summary and only service evidence | Range text remains metadata, not a persisted inactive host; no customer network used |
| FR-08 report | Filtri e archivio automatici PASS; browser 3080: 21 report, tre pagine, periodo invertito, combinazione profilo/cartella/target e isolamento Casa→Casa 2; Executive PDF A4 di quattro pagine renderizzato e ispezionato | Costo storico da misurare; media score dello storico UI e score consolidato PDF hanno basi diverse, non sono confronti equivalenti |
| FR-09 albero | Browser Casa 14 target, Reparto1 sette; cambio Casa 2 zero. Difetto reparto obsoleto riprodotto/corretto rosso→verde e ricollaudato nella build aggiornata | Prevale il contratto accettato a due livelli; nessuna nuova scrittura su aziende cliente |
| FR-10 discovery | Isolated .3/.4 discovered, copy IP/import company choice/recovery PASS; new CSV two visible/JSON one selected, only selected .4 launches one new job in three-target QA company; mobile toolbar 390px PASS | Historical FR-10.5 superseded by accepted cleanup/SCAN-UX; no host-only preset reintroduced |
| FR-11 SSH/firewall | M11-FIREWALL-01: nft/legacy quattro prove ciascuno, OpenSSH disconnect/late mutation PASS; browser expiry/conferma/policy/riordino/sostituzione e terminale I/O PASS, 192 test frontend/gate All PASS | Limiti UID/host/state descritti nel piano; accettazione PO separata, nessun server cliente usato |
| FR-12 SSE/persistenza | Gate All/DB PASS; migrazione 000011/up-down-up; callback host/finding e errori DB prima del publish. Browser con Nmap sul solo host Docker sintetico: snapshot completo e parziale conservati dopo refresh e restart backend | Nessuna prova su LAN cliente; nessuna ripresa automatica del processo o pub/sub live distribuito |
| FR-13 diagnostica | Browser REST export redacts synthetic SSH password. WS logger fix RED→GREEN; real WS filter/Markdown/JSON show 64 state/frame metadata entries; no terminal sentinel, ticket query or JWT-shaped value | Raw terminal content intentionally excluded, rather than relying on free-text secret redaction; no external sharing |

Non riesumare requisiti precedenti superati dalle accettazioni PO (es. albero
arbitrariamente profondo). M3 remediation chokepoint, notifiche e virtual patching
restano fuori da M1.1, salvo dipendenza tecnica condivisa del firewall.

## Primo incremento firewall

Riproduzione dalla directory `backend`, cache Go locale ignorata:
`go test ./internal/domain ./internal/service -run 'TestFirewallRuleParams_RejectUnsafeNetworkFields|TestFirewallMutationCannotDisableSafetyTimer|TestFirewallRejectsInvalidTimeoutBeforeSSH' -count=1`
FAIL atteso exit 1: cinque campi ostili accettati, flag false aggira snapshot,
timeout negativo arriva all'esecutore SSH.

Dopo fix: `go test ./internal/domain ./internal/service -run 'TestFirewall|TestBuildIptables|TestBuildChokepoint' -count=1`
PASS exit 0 (ultima esecuzione 2,659s wall; domain 0,274s, service 0,960s).
Test aggiunti per confini IPv4/CIDR/interfacce, rifiuto prima di SSH, commenti con
metacaratteri preservati, protezione add/delete non disattivabile, timeout invalidi.
Aggiornato l'atteso esatto del test chokepoint per il quoting sicuro condiviso.

Da `frontend`: `npm.cmd test -- --run src/components/FirewallManager.test.tsx`
PASS exit 0, 2 test, 13,82s runner. Primo avvio sandbox FAIL exit 1 per accesso
esbuild; esecuzione con escalation riuscita. Nessuna asserzione indebolita.

## Ambiente e limiti

`cua.getState()` restituisce `apps: []`, `browsers: []`: nessun collaudo visuale
eseguito o screenshot prodotto. Docker `version` con permessi estesi PASS exit 0,
server 29.8.2; inventario immagini locale letto. Nessun container di laboratorio
avviato, nessun host remoto contattato e nessuna regola reale modificata.
Nessuna migrazione o repository DB cambiato dal primo incremento.
Self-review in prima persona; limiti watchdog non chiusi dalle validazioni input.

## Gate finale primo incremento

Dalla root: `scripts/verify.ps1 -Profile Final -Scope All` con GOCACHE locale.
Backend `go test ./...` PASS exit 0 (20,816s), `go vet ./...` PASS exit 0
(1,189s), `go build ./...` PASS exit 0 (1,432s). Frontend 168 test PASS,
build FAIL exit 2: fixture `chains: {}` non soddisfaceva `Record<ChainName,
FirewallChain>`. Corrette le tre catene della fixture senza cast o indebolimenti.
Report `.agent/artifacts/testing/20261018-112428-c3f82bab.json`.

Ripetuto solo il lato invalidato, dalla root:
`scripts/verify.ps1 -Profile Final -Scope Frontend` PASS exit 0:
168 test/28 file (5,032s), build/typecheck (15,301s), verificatore documentale
(0,760s; 66 Markdown/233 link), diff check (0,051s). Report
`.agent/artifacts/testing/20261018-112524-3edeff41.json`. Backend invariato,
evidenze precedenti riutilizzate. PostgreSQL non rieseguito: nessuna modifica
repository/schema/SQL nel primo incremento; nessun PASS DB dedotto.

## M11-REPORT-01 · Secondo incremento · 2026-10-18

Decisione PO acquisita: iptables e iptables-nft con modello attuale. Contratto
nel piano M1.1; nessun motore nftables nativo aggiunto.

Frontend: archivio paginato caricato completamente; filtri locali inclusivi UTC,
profilo/target/cartella/severità esatti; scope aziendale anche in cronologico,
navigazione a 10 righe. Conteggio incoerente, pagina duplicata o incompleta espone
errore/riprova; risposte superate non aggiornano stato. Download consolidati
restano completi con nota UI. Nessuna modifica backend/API/DB/dipendenze.

Mirati da `frontend`: `npm.cmd test -- --run src/pages/Reports.test.tsx src/utils/reportArchive.test.ts`
PASS exit 0, 11 test/2 file, 1,941s wall (runner 1,31s). Prima prova 1 test accordion
FAIL: il vecchio selettore testuale puntava alla nuova option cartella; corretto a
`getByRole('heading')`, mantenute le asserzioni espansione/collasso.

Gate dalla root `scripts/verify.ps1 -Profile Final -Scope Frontend` PASS exit 0:
175 test/29 file 4,995s, build/typecheck 10,805s, docs 0,699s, diff 0,053s.
Report `.agent/artifacts/testing/20261018-113105-c03e6e33.json`.
Baseline backend riutilizzata dal primo incremento: sorgenti, test, dipendenze,
configurazione e ambiente backend invariati. Nessun PASS visuale/PDF/LAN/DB dedotto.
Self-review: isolamento scope, race richieste, errori pagine parziali, nomi filtri,
perimetro download, nessun segreto o dipendenza aggiunti. Limite: costo lineare
dello storico; TECH-DEBT-10, ADR-22. Task resta `[ ]` per gate browser/PDF aperto.

## Browser e PDF · 2026-10-18

Il PO ha reso disponibile il browser ed eseguito il login su localhost:3080.
Archivio Casa: 21 report su tre pagine (10/10/1), verifica pagina 2 e 3, reset
pagina dopo filtri; profilo TCP 100 + cartella Casa → sette righe, target
192.168.1.1 nella cartella Casa → una riga. Periodo iniziale successivo al finale:
messaggio esplicito e storico vuoto. Cambio scope Casa 2 → nessun report Casa.
Nessuna scansione, importazione o modifica di dati applicativi.

Executive consolidato Casa scaricato tramite UI. Copia e quattro PNG in
`.agent/artifacts/M11-REPORT-01/`. `pdfinfo`: A4, quattro pagine, 9057 byte.
Poppler eseguito in container temporaneo con pacchetto Alpine `poppler-utils` e
font Liberation; rendering/testo PASS exit 0 (2,950s). Primo rendering senza font
produceva diagnostica Helvetica; corretto l'ambiente, non il PDF. Ispezionate
tutte le quattro pagine: cover, posture/priorità reparto, priorità senza reparto
e inventario; nessun taglio/sovrapposizione rilevato. Non prova correttezza delle CVE.

SSH: modale creazione con campo pin e istruzione canale fidato; modifica server
esistente, cambio endpoint a indirizzo TEST-NET azzera il pin; Annulla preserva
l'host registrato. Screenshot osservato nel tool; nessuna connessione SSH o
mutazione firewall avviata.

Difetto riprodotto al login: provider aziende montato prima dell'autenticazione,
due richieste inventario premature e contesto vuoto fino a refresh. Test
`src/App.test.tsx` rosso exit 1 (1,699s), provider spostato dentro le rotte protette;
verde con regressione contesto/discovery: 14 test exit 0 (1,959s). Fix non ancora
collaudato nella build Docker; nessun deploy implicito.

## Gate integrato e deploy locale autorizzato

Root: `scripts/verify.ps1 -Profile Final -Scope All -Integration`, DB locale
dedicato `nextgen_discovery_test` su 55432 con TEST_DB_* esplicite. PASS exit 0:
backend suite 22,846s, vet 1,112s, build 1,255s; frontend 184 test 5,578s,
build/typecheck 11,427s; PostgreSQL 5,609s; docs 0,598s/diff 0,059s.
Report `.agent/artifacts/testing/20261018-124105-426c02b0.json`.
Self-review, non review indipendente: SQL parametrico/FK tenant, topic tenant,
ordine commit/SSE, pipe con join e cancellazione, fallimenti storage senza
notifiche di successo, raw XML preservato, nessuna dipendenza applicativa aggiunta.

PO ha poi autorizzato il deploy locale dei soli container app. Immagini verificate
`m11-persist-20261018`, Compose exit 0 (6,665s); backend healthy, frontend e
`/healthz` HTTP 200. Migrazione applicata all'avvio senza fermare PostgreSQL.
Discovery reale sul solo container 172.30.77.3 in rete Docker interna isolata:
un host trovato, nessun import. Refresh/restart conservano il risultato completo;
refresh immediato dopo il primo device durante enrichment conserva uno snapshot
parziale, con banner corretto anche dopo restart. Regola TCP temporanea del solo
container rimossa. Dettagli nel [piano persistence](../04-development/plans/M11-PERSIST-01.md).
M11-REPORT-01, SSH-HARDEN-01 e M11-PERSIST-01 passano a `[~]`; M1.1 resta aperta.
