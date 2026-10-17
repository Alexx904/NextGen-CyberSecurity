# AP-REVIEW-01 · Validazione e usabilità Attack Path

## Intake e criterio di utilità

Richiesta PO 2026-10-17: verificare correttezza cyber, funzionamento e leggibilità di `/attack-path`, correggere problemi e integrare una guida. Checkpoint pre-intervento pubblicato: `b16264b`. Dati presenti e scenari sintetici autorizzati; nessun exploit o test remoto. 3D facoltativo solo se migliora comprensione.

Il prodotto deve aiutare a prioritizzare e verificare ipotesi, distinguendo associazioni osservate da accesso e compromissione non provati. Non è una validazione di lateral movement reale. CVSS esprime severità, non probabilità: fonti [FIRST](https://www.first.org/cvss/user-guide.html) e [MITRE TA0008](https://attack.mitre.org/tactics/TA0008/).

## Finding iniziali e piano

- Alto: percorsi multi-hop con coppie di nodi prive di arco; `edge_ids` non popolati. Validare sequenza e identità degli archi, test rosso/verde.
- Alto: Internet assunto su reti private, pivot fra reti arbitrarie e trust dedotto dall'inventario SSH. Usare origine di analisi ipotetica; pivot solo come ipotesi esplicita di vicinanza IPv4 /24, mai prova di subnet reale, RCE o credenziali.
- Alto: scope aziendale applicato alla sola lista con fallback globale. Proiettare nodi, archi, percorsi, chokepoint e KPI nello scope; nessun fallback. Scartare catene che escono dallo scope.
- Alto: layout con BFS a profondità crescente non termina su cicli. Distribuzione deterministica finita e vista dedicata al percorso selezionato.
- Medio: collisione TCP/UDP, duplicati storici, limiti di lettura taciuti, richieste ripetute su selezione. Distinguere protocolli, usare ultima osservazione per endpoint e segnalare copertura incompleta.
- Medio: score presentato come probabilità, CVE come RCE e chokepoint come garanzia. Rinominare in indice euristico e priorità dei soli percorsi modellati.

## Contratti prima del codice

API esistente e interfaccia `AttackPathService` preservate. `AttackGraph.analysis` additivo: `model`, `warnings`, `paths_truncated`, `inventory_truncated`, `target_count`, `finding_count`, `max_paths`. `AttackPath.edge_ids` sempre completo, un arco orientato per coppia consecutiva; hop count coerente. Nessuna migrazione o dipendenza.

UI: 2D con vista percorso per leggere la catena da sinistra a destra, vista completa per contesto, adattamento vista, selezione tastiera, legenda osservato/ipotesi e guida espandibile. Nessuna modalità 3D obbligatoria: privilegiare sequenza e assenza di occlusione.

## File, test e gate

Backend domain/service/handler Attack Path e test; frontend types/page/graph, helper puri scope/layout e test. Documenti: dominio, requisiti FR-14, UI, API, ADR, task, QA, PROJECT_STATE.

Verifica locale: cache Go esclusa in `.gitignore`; snapshot e harness temporanei conservati solo negli artefatti ignorati. Nessun dato locale aggiunto al diff applicativo.

Test: adiacenza e edge IDs, protocollo, deduplica, nessuna falsa esposizione o trust, min_cvss invalidi, limite percorsi, determinismo, cicli, scope vuoto/CIDR/cross-company, selezione senza refetch, error/loading, guida. Suite complete Go/Vitest, build/vet, doc verifier e diff check. Prova browser con dati sintetici dichiarati e screenshot in `.agent/artifacts/AP-REVIEW-01/`. Review finale svolta dallo stesso agente.
