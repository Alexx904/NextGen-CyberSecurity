# AP-REVIEW-01 · Revisione Attack Path

Data: 2026-10-17. Baseline pubblicata prima della revisione: `b16264b`, branch `feat/nmap-scan-workflow-review`. Implementazione corrente: working tree. Review svolta dallo stesso agente; nessuna attestazione indipendente o pentest di rete.

## Giudizio tecnico

La baseline non era validabile come prova di attack path/lateral movement: costruiva connessioni non dimostrate, cammini non adiacenti e poteva non terminare su cicli. La revisione rende il motore utilizzabile per **triage di finding e ipotesi da verificare**, nei limiti documentati. Non dimostra catene di compromissione reali: mancano telemetria di raggiungibilità, ACL, identità/privilegi, prerequisiti e validazione exploit. Per questi ultimi servono evidenze di laboratorio autorizzato, non sostituibili con test sintetici.

## Finding e trattamento

| Priorità | Finding iniziale | Trattamento / evidenza |
|---|---|---|
| Alta | Cammini multi-hop con nodi consecutivi senza arco; `edge_ids` vuoti | Test rosso riprodotto; sequenze generate dagli archi e validate dal dominio; verde. |
| Alta | Esposizione Internet assunta anche su privati e FQDN | Radice ipotetica, accesso inferito per ogni asset; nessuna dichiarazione di Internet verificato. |
| Alta | CVSS alto o inventario SSH trasformati in pivot su reti arbitrarie | Inventario SSH escluso dalle inferenze; nessun pivot cross-subnet; /24 solo ipotesi debole dichiarata. Test negativi e positivo /24. |
| Alta | BFS a profondità crescente su cicli, possibile blocco UI | BFS visita ciascun nodo una volta; test ciclo raggiungibile e layer 0 della radice. |
| Alta | Filtro aziendale solo sulla lista e fallback globale | Proiezione completa con esclusione delle catene cross-scope e ricalcolo KPI/chokepoint; test helper e pagina su azienda vuota/CIDR. Non cambia l'autorizzazione tenant dell'API. |
| Media | TCP/UDP fusi, duplicati e troncamento silenzioso | Identità per protocollo, osservazione più recente per endpoint, deduplica, paginazione con spareggio ID, limiti dichiarati. Test storico/closed e 205 target/505 finding. |
| Media | Score descritto come probabilità; CVE come RCE; remediation garantita | Etichette, guida, requisiti e dominio aggiornati; formula storica preservata come indice euristico non calibrato. |
| Media | Selezione attiva causava fetch e risposte obsolete potevano sovrascrivere filtri | Fetch dipende dal filtro, risposte obsolete scartate; test selezione senza refetch. |
| Media | Guida assente e grafo difficile da leggere | Guida integrata, vista sequenziale e mappa completa, scorrimento, legenda, selezione tastiera. Prova visuale reale ancora aperta. |

## Evidenze eseguite

Tutti i comandi applicativi dalla rispettiva directory `backend`/`frontend`.

- Prima del fix: `go test ./internal/service -run TestAttackReview -count=1` FAIL, exit 1, tre difetti riprodotti (archi, esposizione, protocollo/deduplica).
- Dopo il fix: `go test ./internal/service ./internal/domain ./internal/delivery/http/handler -run 'Attack|Defense' -count=1` PASS, exit 0.
- `go test ./...`, `go build ./...`, `go vet ./...` PASS, exit 0. Cache Go locale; nessun test saltato per ottenere verde.
- `npm.cmd test` PASS, 23 file / 156 test; `npm.cmd run build` PASS, exit 0. Esecuzione fuori sandbox per permessi esbuild; TypeScript e Vite inclusi. Test specifici su archi esatti, stile inferito/osservato, selezione tastiera e reset zoom.
- Snapshot dal PostgreSQL locale di sviluppo, sola lettura: 11 tenant, 17 target, 20 finding, **0 CVE associate**. Campi limitati a target e finding senza banner/credenziali; artefatto ignorato `.agent/artifacts/AP-REVIEW-01/local-snapshot.json`.
- Harness temporaneo `go test ./internal/service -run TestAttackReviewLocalSnapshot -v -count=1` PASS, exit 0, soglie 0/7/9 per ogni tenant. Tenant con 7 target e 20 finding: 28 nodi, 27 archi, 7 percorsi inventariali a soglia 0 e 0 percorsi CVE a soglie 7/9. Nessuna catena CVE/laterale inventata in assenza di CVE. Harness conservato negli artefatti ignorati e rimosso dai sorgenti permanenti.
- UI automation: inventario browser vuoto; tentativi browser interno e Chrome restituiscono `Browser is not available`. Nessuno screenshot o prova responsive reale. Non è un PASS visuale.

## Gate residui e limiti

Doc Sync: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify-agent-docs.ps1` PASS (24 obbligatori, 59 Markdown, 148 link/ancore), exit 0; whitespace verificato con `git diff --check`. Diff rivisto dallo stesso agente: nessuna nuova dipendenza, migrazione, chiamata remota di mutazione, shell applicativa, log di credenziali o indebolimento dei test. Permessi tenant esistenti invariati; scope aziendale è presentazione. Difetti residui di evidenza/scala registrati in TECH-DEBT-08.

## Documenti aggiornati e motivo

| Documento | Motivo |
|---|---|
| `docs/domains/attack_path.md` | Contratto canonico del modello effettivo, formula reale, scope, grafo e limiti. |
| `docs/01-product/REQUIREMENTS.md` | FR-14 coerenti con analisi euristica, senza promesse di compromissione/probabilità. |
| `docs/01-product/USE_CASES.md` | UC-13 con lettura evidenze, verifica e remediation valutata. |
| `docs/02-design/UI_SPECIFICATIONS.md` | Vista percorso/completa, guida, tastiera, responsive atteso e gate visuale. |
| `docs/03-architecture/HLD.md` | Confini architetturali e letture limitate del modello. |
| `docs/03-architecture/ADR/README.md` | ADR-17 su evidenze, compatibilità e scelta 2D. |
| `docs/03-architecture/DOMAIN_MAP.md` | Copertura reale di domain/handler/types/API/componenti graph. |
| `docs/03-architecture/LLD/API_SPEC.md` | Metadati analysis additivi e validazione min_cvss. |
| `docs/03-architecture/LLD/DATABASE_SCHEMA.md` | Schema invariato, ordinamenti stabili per paginazione. |
| `docs/04-development/plans/AP-REVIEW-01.md` | Intake, finding, contratti e piano di test. |
| `docs/04-development/TASK_LIST.md` | Stato `[!]` del gate browser e TECH-DEBT-08. |
| `docs/05-qa/TEST_PLAN.md` | Riferimento alle evidenze e prova visuale residua. |
| `docs/05-qa/ATTACK_PATH_REVIEW.md` | Report e pacchetto tecnico di validazione. |
| `docs/PROJECT_STATE.md` | Checkpoint push, verifiche, decisioni e prossimo gate. |

AP-REVIEW-01 resta `[!]` per la verifica browser richiesta. Nessuna modifica dei dati, esecuzione exploit, nuova scansione, deploy o secondo push della revisione effettuati. Il push iniziale salva solo il checkpoint precedente. E28-S1-T1 resta distinto con gate Linux/SSH aperto.

Limiti intrinseci: modello euristico, dati storici non equivalenti allo stato attuale, cap 2.000/10.000/50, letture non transazionali durante mutazioni concorrenti. Il cap globale prima del filtro aziendale può sottorappresentare una singola azienda; avviso visibile. CIDR/range IPv4 e identificatori esatti supportati nello scope; FQDN non risolti implicitamente, CIDR IPv6 non espansi. 3D non introdotto per privilegiare la sequenza senza occlusioni; scelta progettuale, non risultato di user testing.

Per chiudere: prova `/attack-path` su desktop/mobile con dati sintetici controllati e screenshot, verifica vista percorso/completa, selezione nodi e archi, zoom/pan/reset, guida, errore/vuoto e cambio azienda. Successiva accettazione PO. Contratto in [dominio Attack Path](../domains/attack_path.md).
