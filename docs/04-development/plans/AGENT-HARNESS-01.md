# AGENT-HARNESS-01 · Progressive repository context

Ultimo allineamento: 2026-10-18 · working tree. Richiesta PO: applicare i miglioramenti dell'audit dopo verifica delle pratiche aggiornate. Scope: governance, documentazione e relativi verificatori; nessuna modifica applicativa, dipendenza, configurazione personale o pubblicazione remota.

## Contratto e fonti

- [Codex best practices](https://learn.chatgpt.com/docs/learn/best-practices): istruzioni pratiche e rimandi ai dettagli specifici.
- [Build skills](https://learn.chatgpt.com/docs/build-skills): discovery repository in `.agents/skills`, metadati iniziali e caricamento progressivo del corpo/riferimenti. Le skill restano invocabili implicitamente; nessun requisito di plugin/MCP per queste procedure locali.
- [GPT-6 family](https://developers.openai.com/api/docs/guides/latest-model?model=gpt-6.1-sol): calibrare autonomia e verifica sul lavoro. La guida espone comportamenti osservati con Astra, da valutare sul modello/workload scelto.
- [Rethinking skills and prompts](https://developers.openai.com/blog/rethinking-skills-and-prompts-for-gpt-6-astra): descrizioni precise, contesto pertinente e rimozione di vincoli ridondanti; indicazioni Astra non assunte come benchmark Sol.

Consultate il 2026-10-18. Si mantengono convenzioni e gate locali; non si adotta la struttura di un esempio ufficiale come requisito del prodotto.

## Implementazione

AGENTS resta contratto sempre attivo con numeri di sezione compatibili. Quattro skill specifiche possiedono verifica, migrazioni, review e Doc Sync; i playbook restano ingressi conversazionali. Ruoli consultati per competenze distintive, handoff condiviso. PROJECT_STATE contiene fatti correnti; snapshot storico esterno a docs preserva contenuto/accettazioni, con link ribasati e autorizzazioni chiaramente storiche.

DOMAIN_MAP è la tabella canonica leggibile anche dal verificatore, senza copia JSON. Ogni sorgente nel perimetro ha un documento primario; glob senza riscontri, sorgenti nuove senza mapping e ownership ambigua falliscono. Percorsi trasversali hanno un dominio core esplicito, senza catch-all che nasconda nuove aree. Test e file non tracciati sono inclusi.

Il verificatore esclude artifacts temporanei, include le skill, controlla manifest semplici/name unici/riferimenti e limiti delle liste di stato. Budget caratteri AGENTS 10.000 e stato 8.000 sono avvisi, non permesso di omettere invarianti. Fixture isolate verificano rifiuti reali e regressione artifacts; nessuna scansione o DB applicativo coinvolti.

## Misurazione e preservazione

Baseline iniziale e originali in `.agent/artifacts/agent-harness-01/` ignorati da Git. Il confronto usa gli stessi file AGENTS/PROJECT_STATE prima/dopo; spostare dettagli nello storico riduce bootstrap, non il patrimonio documentale totale. Caratteri non equivalgono a token, quota o costo. Nessun tokenizer o API a pagamento introdotti.

| Testo obbligatorio | Caratteri prima | Caratteri dopo |
|---|---:|---:|
| AGENTS.md | 15.832 | 8.918 |
| PROJECT_STATE.md | 47.213 | 6.900 |
| Totale dei due file | 63.045 | 15.818 |

Riduzione del testo dei due file: 74,9%, sulla baseline di questo intervento e a contenuti finali dello stato. Non include metadati delle skill, task, tool, reasoning o cache; non è una misura di risparmio token/costo o un confronto di qualità.

Controllare ID/stati task prima/dopo e testo dello snapshot normalizzato salvo ribasamento link. Non promuovere task applicativi né attribuire a questo lavoro le modifiche M1.1 preesistenti.

## Documenti aggiornati

| Fonti | Motivo |
|---|---|
| AGENTS, .agent/README, feature/feedback/orchestration e ruoli | Contratto compatto, routing e handoff comuni senza procedure duplicate |
| Quattro SKILL.md e ingressi verify/db-migrate/review/sync-docs; release | Procedure specializzate e comando DB corretto in un'unica sede |
| PROJECT_STATE, snapshot storico e link CHANGELOG | Stato corrente distinto da storia, preservando accettazioni/evidenze |
| DOMAIN_MAP e core_and_security | Copertura eseguibile dei percorsi reali e ownership trasversale |
| docs/README, PRD, TASK_LIST, TEST_PLAN, ADR-23 e nota ADR-06 | Indice coerente, rimozione ripetizioni, scope distinto da completamento e supersessione esplicita |
| Questo piano | Fonti, verifica, misure e limiti della consegna |

## Verifica e limiti

Comandi dalla root: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1 -Profile Docs` e `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/test-agent-harness.ps1`.

Gate 2026-10-18, dalla root, tutti exit 0: profilo Docs (verificatore 0,792s; diff 0,065s; invocazione complessiva 1,318s), report `.agent/artifacts/testing/20261018-114735-c1bb6cad.json`; 22 prove harness in 2,276s (invocazione 2,680s). PASS preservazione: snapshot completo salvo link ribasati e 26 ID/stati task preesistenti invariati. Self-review: confini letture/path, esclusione artifacts prima della ricorsione, mapping senza fallback generale, nessuna nuova dipendenza o sorgente applicativo. Sole annotazioni Markdown successive richiedono nuovo gate Docs, non nuove suite applicative.

Difetti introdotti corretti durante sviluppo: default del path radice valutato troppo presto in PowerShell, asserzione fixture con conversione array/bool, link storico spostato e tre destinazioni mancanti nella copia isolata. Controlli verdi solo dopo correzione, nessuna asserzione indebolita. Non richiede suite backend/frontend o PostgreSQL perché non cambia sorgenti o comando eseguibile del runner applicativo; il comando DB documentato è riallineato al tag già esistente.

Python assente nell'ambiente: `quick_validate.py` della skill-creator non eseguibile. Manifest vincolati a due scalari semplici, verificati dal controllore PowerShell senza dipendenze; questo controllo non certifica il routing del modello. PO autorizza installazione se più veloce: nessun vantaggio dimostrato rispetto ai pochi secondi misurati, quindi nessuna installazione o modifica alla configurazione personale.

Valutazione comportamentale successiva: confrontare istruzioni prima/dopo con stesso modello GPT-6.1 Sol, effort, tool, snapshot e ambiente su bug locale, UI, API comune, migrazione, Markdown, feedback ambiguo e tree sporca. Misurare criteri soddisfatti, gate omessi, token input/output/reasoning/cached quando esposti, latenza, interventi PO e lavoro correttivo. Ripetere i casi e analizzare gli errori; adottare una variante solo se preserva i gate critici. Nessun benchmark Sol o risparmio token dichiarato da questa verifica strutturale.
