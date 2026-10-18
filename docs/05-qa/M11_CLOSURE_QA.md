# M1.1 · Matrice di chiusura

Ultimo allineamento: 2026-10-18 · working tree. Ricognizione statica e primo
incremento firewall; non attestazione di completamento della milestone.

## Criteri e gap

| Requisito | Evidenza presente | Gap da chiudere |
|---|---|---|
| FR-07 alive-only | Test orchestratore `TestScanOrchestrator_AliveOnlyFiltering`, parser range/down | Collaudo integrato UI/export con host up/down e filtro senza porte |
| FR-08 report | Report service/handler, PDF/CSV e report aziendali accettati; M11-REPORT-01 aggiunge filtri periodo/profilo/target/cartella e paginazione completa, scope cronologico, test automatici PASS | Collaudo browser e render PDF aperti; costo caricamento storico da misurare |
| FR-09 albero | Gerarchia azienda/reparto e report per reparto accettati (COMPANY-UX-02) | Collaudo browser; prevale il contratto aggiornato a due livelli |
| FR-10 discovery | Test service/handler, parsing e identità DISCOVERY-01 accettati | Collaudo LAN isolata e revisione azioni rapide e riconnessione |
| FR-11 SSH/firewall | Terminale/proxy accettati, pinning e deadline con test; PO conferma iptables/iptables-nft con modello attuale, niente nft nativo | Gate visuale pinning; watchdog Linux/multiistanza TECH-DEBT-07; policy, modifica/riordino regole e terminale docked richiesti ma non dimostrati dalle API/UI correnti; collaudo iptables-nft |
| FR-12 SSE/persistenza | Test `realtime_e2e_test.go`, risultati discovery progressivi in `sessionResults` | La memoria del processo non prova PostgreSQL FR-12.3: definire e implementare persistenza discovery e verificare recupero; verificare anche scansioni durante esecuzione Nmap, non solo eventi dopo parsing |
| FR-13 diagnostica | Test logger/exporter/drawer/context e debugConsoleE2E | Collaudo browser e verifica redazione integrata |

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
