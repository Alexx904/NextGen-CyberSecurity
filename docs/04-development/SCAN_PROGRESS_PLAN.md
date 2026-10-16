# SCAN-TIME-01 · Durate indicative e avanzamento

Ultimo allineamento: 2026-10-16 · working tree. Estensione richiesta dal PO di SCAN-UX-01; modifiche precedenti preservate.

## Contratto e piano prima del codice

- Otto preset: intervallo di durata illustrativo per singolo IP raggiungibile su LAN stabile, T3, carico moderato. Non benchmark né SLA; subnet, filtri, perdita, rate limiting, UDP/NSE e code possono superarlo. Timeout rimane limite separato.
- Runner: callback opzionale `ProfileConfig.OnProgress`; `--stats-every 5s` aggiunto internamente prima del delimitatore target, fuori dall'input utente. Osservare XML taskprogress senza alterare RawXML, con buffer incrementale limitato e tolleranza a chunk/malformed/missing remaining.
- SSE `progress`: mantenere campi esistenti, aggiungere facoltativi phase, phase_percentage, remaining_seconds. Percentage è una stima pesata monotona complessiva del job, non la percentuale nativa della sola fase. ETA euristica dalle statistiche del motore, mai countdown del timeout. Nessuna migrazione DB.
- Dettaglio: barra accessibile e ETA, fase nativa distinta; nessuna percentuale fittizia prima della telemetria, coda senza ETA. Stime scadute/oltrepassate mostrate come in aggiornamento. Completamento solo da stato server.
- Scope confermato dal PO: intero gruppo di target. Monitorare gli ID restituiti dal lancio, senza includere lo storico. Il progresso dei job alimenta la barra di lotto, i terminali sono distinti per successo/errore/annullamento; 100% del lavoro terminato non significa tutti riusciti.
- Contratto additivo: `ScanResponse.progress` opzionale, snapshot in memoria della telemetria attiva con observed_at; `GET /scans?ids=<uuid,...>` legge fino a 100 ID del tenant, niente paginazione implicita. Il client usa blocchi da 50 e polling non sovrapposto ogni 5s; terminali non ripollati. Lotto in sessionStorage associato al tenant, ID sempre nuovamente verificati dal server, non allo storico.
- ETA lotto: somma dei tempi residui osservati / capacità di esecuzione osservata, più limite inferiore del job più lungo; coda stimata dalla durata media prevista dai job osservati. Nessuna ETA con soli job in coda o telemetria incompleta/scaduta, nessuna promessa sulle risorse condivise. Assenza di cardinalità IP per CIDR non moltiplicata arbitrariamente.
- Test: XML frammentato/invalidi/buffer, trasmissione runner e SSE, monotonicità/ETA, rendering/coda/stale/completamento e regressioni. Suite Go e frontend, build/vet, Doc Sync, self-review. Browser reale se disponibile; mai scansioni su reti non autorizzate.

Fonte: [Nmap output](https://nmap.org/book/man-output.html) documenta statistiche periodiche anche su XML; [DTD ufficiale](https://nmap.org/book/nmap-dtd.html) descrive taskprogress. La durata totale non è fornita con precisione da Nmap e viene esplicitamente stimata.

## Pacchetto di revisione e risultati

Implementato monitor dell'intero gruppo, confermato dal PO. Le durate per profilo sono per singolo IP: TCP 1000/versioni 1–5 min, TCP 100 10–60 s, TCP completo 10–60 min, web 15–120 s, NSE 5–30 min, UDP 1–10 min, legacy 15–120 s, ACK 30–180 s. Ipotesi LAN stabile, T3, IP raggiungibile; non benchmark. Il totale dipende da scope e parallelismo, non dalla semplice moltiplicazione dei record target.

La barra media i job del lotto, con high-water mark per job e stati terminali a 100; il conteggio distingue riusciti/falliti/annullati. Rimanente job estrapolato dal ritmo totale pesato dopo almeno 5s e 10%, mai dalla sola discovery, con residuo Nmap come limite inferiore. Residuo lotto da lavoro in esecuzione + durata media prevista per job in coda, diviso parallelismo osservato, e limite del job più lungo. Intervalli UI ±50% lotto e ±30% dettaglio sono margini euristici, non confidenza statistica. Snapshot fase successiva invalida ETA vecchie; dati scaduti dopo 30s o senza basi sufficienti indicano «in calcolo».

Prova: pagina Scansioni → Nuova scansione, selezionare un profilo e controllare durata/timeout distinti; lanciare su un gruppo di laboratorio autorizzato. La card lotto compare nella pagina Scansioni con barra, conteggi ed ETA quando disponibili. Ricaricare per verificare ripristino sessione tenant. Il 100% non deve nascondere job falliti o annullati.

Verifiche eseguite il 2026-10-16:

| Directory | Comando | Esito |
|---|---|---|
| frontend | `npm.cmd test` | PASS, exit 0, 21 file / 145 test |
| frontend | `npm.cmd run build` | PASS, exit 0, TypeScript + Vite 8,20 s |
| backend | `go test ./...` | PASS, exit 0 |
| backend | `go build ./...` | PASS, exit 0 |
| backend | `go vet ./...` | PASS, exit 0 |
| root | `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify-agent-docs.ps1` | PASS, exit 0, 55 Markdown / 121 link-ancore |
| root | `git diff --check` | PASS, exit 0 |

Test con XML frammentato, attributi assenti/NaN/valori fuori range, buffer limitato, helper process senza rete, preservazione XML e flag prima di --; stime monotone, snapshot tenant/cleanup, API batch invalidi/limiti/tenant estraneo. Frontend: conteggi, coda, errori, stale, chunk da 50, stop terminali/unmount, ripristino lotto senza includere storico. Self-review svolta dallo stesso agente su injection, tenant, snapshot/lock/lifetime, polling/abort e significato delle stime. Nessuna review indipendente dichiarata.

Browser inventory dell'agente 2026-10-16: apps/browsers vuoti. Il PO successivamente dichiara di aver eseguito i test delle scansioni con esito positivo e accetta SCAN-TIME-01, autorizzando il push. Gate funzionale soddisfatto dalla prova riferita dal PO; nessun PASS visuale autonomo, screenshot o benchmark inventato. Il processo helper prova il cablaggio di telemetria, non l'accuratezza delle stime su una rete reale. Nessuna nuova dipendenza o migrazione DB; modifiche preesistenti SCAN-UX-01 preservate.

Documenti aggiornati e motivo:

- `docs/01-product/REQUIREMENTS.md`: FR-16.5 tempi e scope gruppo.
- `docs/02-design/UI_SPECIFICATIONS.md`: intervalli profili, card lotto e stati.
- `docs/03-architecture/HLD.md`: pipeline telemetria/snapshot/client.
- `docs/03-architecture/DOMAIN_MAP.md`: mapping DTO e handler toccati.
- `docs/03-architecture/LLD/API_SPEC.md`: campi opzionali e filtro batch tenant.
- `docs/03-architecture/ADR/README.md`: ADR-15 e limiti architetturali.
- `docs/domains/scanning_and_discovery.md`: intervalli, formule, contratti e limiti.
- `docs/04-development/SCAN_PROGRESS_PLAN.md`: piano, fonti, esiti e pacchetto.
- `docs/04-development/TASK_LIST.md`: task e accettazione PO dopo la sua prova funzionale.
- `docs/05-qa/TEST_PLAN.md`: test telemetria/ETA/lotto e verifiche mancanti.
- `docs/PROJECT_STATE.md`: consegna, assunzioni e stato reale.
