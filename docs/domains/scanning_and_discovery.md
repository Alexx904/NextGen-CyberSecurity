# Dominio: Scanning & LAN Network Discovery

## M11-DISCOVERY-UI-01 · Selected devices and exports · 2026-10-18 · working tree

Discovery exports CSV/JSON for the explicitly labelled visible/selected set.
CSV quotes cells and neutralizes spreadsheet formula prefixes; JSON preserves
snapshot metadata, partial status and exported count. Local download only.
Selected-device action opens import/destination, then existing scan configuration
restricted to the selected IP values in fresh inventory and the chosen subtree.
It sends existing target_id requests, never a whole-group fallback; blacklist and
company scope apply, duplicate IP entries are deduplicated. Definite rejected
requests can be retried without repeating confirmed successes; network/5xx
uncertainty blocks retry and requires history review. Import alone does not run
scans. Restricted scheduling uses the same eligible set. Successful import refreshes
company context. Eight accepted presets remain; historical FR-10.5 is superseded.
[Plan and QA](../04-development/plans/M11-DISCOVERY-UI-01.md). Browser export,
single selected job, existing-device reuse and 390px toolbar PASS; task `[~]`.

> Dominio: port/service discovery, vulnerability scanning e SSE. File: `backend/internal/scanner/`, `backend/internal/service/scan*`, `backend/internal/service/discovery_service.go`, handler scan/discovery, `frontend/src/components/NewScanModal.tsx`, `frontend/src/components/scans/`, pagine Scans/ScanDetail/NetworkDiscovery.

## 1. Contratto di esecuzione

COMPANY-UX-02 · 2026-10-17 · working tree: azienda obbligatoria per navigare nelle scansioni; scelta azienda/reparto e gruppi esistenti preservati. Reparti solo figli diretti dell'azienda. CompanyContext paginato e aggiornato dopo CRUD target/reparti per mantenere coerenti le scelte. Motore, profili e consenso invariati.

- `NmapRunner` usa `exec.CommandContext` senza shell. `BuildSafeArguments` convalida target e scope, forza `-oX -`, valida gli argomenti, aggiunge `-6` per IPv6 e antepone `--` al target.
- Output XML parsato a streaming; timeout e cancellazione del processo gestiti dal runner. Capabilities container: solo `CAP_NET_RAW`, con `NMAP_PRIVILEGED=1` (SEC-P1-05).
- Allowlist: tecniche TCP SYN/connect/ACK/NULL/FIN/Xmas, UDP, discovery, rilevamento servizi/OS, timing T0–T5, porte, probe SYN/UDP, rate/retry e due selettori NSE `vuln`/`vulners`. Non supporta l'intera CLI Nmap.
- Vietati target inline, file di input/output, caricamento di script o database arbitrari, script-args, shell e modifica del binario. Non esporre la CLI come terminale di sistema.
- Validazione semantica: una sola tecnica TCP, un solo selettore porte, porte numeriche 0–65535 e intervalli crescenti, top-ports 1–65535; blocco di `-sn` con port scanning/versioni/OS e di ACK con `--open`, `-sV`, `-O`. TCP+UDP resta supportato con una lista porte unica (`-p T:80,443,U:53,161`).

## 2. UX attiva: due modalità (SCAN-UX-01)

La proposta Epic 27.1 è sostituita nella modale da **Profili** e **Comando Nmap**. Rimossi compositore multi-obiettivo, card promozionali, catalogo per categorie, stime di durata e badge che dichiaravano parametri già sicuri senza risposta server.

### Profili

Aggiornamento testi SCAN-COPY-01 (2026-10-16): le otto descrizioni espongono soltanto scopo e attività del profilo, con termini tecnici aderenti ai flag. Rimossi confronti con Discovery, formulazioni negative e avvertenze dalle descrizioni; eliminato anche il suggerimento Discovery sotto il selettore. Durate, anteprima, limiti generali e consenso restano nei rispettivi elementi UI. Nessuna modifica a nomi, argomenti o classificazione. I profili già persistiti conservano la descrizione storica; il selettore usa i testi canonici aggiornati.

Fonte eseguibile: `SCAN_PRESETS` in `frontend/src/components/scans/scanConfiguration.ts`. Tutti usano T3; nessun profilo host-only duplicato con Discovery.

| Nome UI | Opzioni determinanti | Limite per target |
|---|---|---|
| Porte TCP e versioni (1000) | `-sS -sV --top-ports 1000 -T3` | 60 minuti |
| Porte TCP frequenti (100) | `-sS -F -T3` | 15 minuti |
| Tutte le porte TCP (1–65535) | `-sS -sV -p- -T3` | 240 minuti |
| Servizi su porte web | `-sS -sV -p 80,443,3000,5000,8000,8080,8443,8888,9000 -T3` | 20 minuti |
| Controlli vulnerabilità NSE | `-sS -sV --top-ports 1000 --script=vuln -T3` | 60 minuti |
| Porte UDP selezionate | `-sU -p U:53,69,123,161,514,623,5353 -T3` | 20 minuti |
| Servizi TCP su porte legacy | `-sS -sV -p 21,23,79,110,143,512,513,514 -T3` | 15 minuti |
| Filtraggio TCP ACK | `-sA -Pn --top-ports 1000 -T3` | 30 minuti |

I timeout non sono stime. Porte web non equivale a test applicativo/API, porte legacy non dimostra traffico in chiaro, ACK non dimostra porte aperte o bypass, NSE non equivale a un penetration test. SYN è rilevabile. `open|filtered` UDP è ambiguo. La correlazione CPE/versione/CVE richiede conferma professionale, specialmente per backport e banner incompleti.

Da SCAN-TIME-01 la UI espone anche intervalli indicativi per singolo IP raggiungibile su LAN stabile a T3: TCP 1000/versioni 1–5 min, TCP 100 10–60 s, TCP completo 10–60 min, web 15–120 s, NSE 5–30 min, UDP 1–10 min, legacy 15–120 s, ACK 30–180 s. Sono ipotesi di pianificazione, non misurazioni né SLA, e possono essere superate. Nessuna moltiplicazione automatica per record CIDR/range o uguaglianza al timeout.

### Avanzamento del gruppo e tempo residuo (SCAN-TIME-01)

Scope confermato dal PO: intero gruppo lanciato insieme. La pagina Scansioni monitora esattamente gli ID restituiti da POST /scans, salva l'ultimo lotto in sessionStorage per tenant e ripristina il monitor al ricaricamento quando l'elenco corrente identifica il tenant. Le selezioni aziendali limitano la visibilità del monitor; un nuovo lancio sostituisce il monitor precedente senza interromperne i job.

Runner con callback opzionale osserva XML taskprogress tramite writer incrementale limitato, senza alterare RawXML; flag interno `--stats-every 5s`, riflesso nell'anteprima. Gli aggiornamenti nativi descrivono una fase, non l'intero job. L'orchestratore usa finestre euristiche 0–5 discovery, 5–65 porte, 65–80 versioni/OS, 80–90 NSE, 90–92 parsing e 92+ correlazione, senza tornare indietro durante Nmap. Stato COMPLETED resta l'unica prova del completamento; snapshot di fase invalidano ETA precedenti. Nessun cronometro artificiale fino al timeout.

Progresso lotto: media dei progressi dei job (terminali = 100, coda = 0); job di dimensioni diverse hanno lo stesso peso, quindi è una stima, non percentuale di pacchetti o IP analizzati. Conteggi separati riusciti/falliti/annullati. ETA: residui dei job in esecuzione più lavoro atteso in coda, divisi per parallelismo osservato, con limite inferiore dato dal job residuo più lungo. La UI espone intervallo ±50%, non intervallo statistico di confidenza; per il dettaglio ±30%. Senza telemetria sufficiente, se più vecchia di 30s, esaurita o durante correlazione senza stima, mostra «in calcolo/in aggiornamento».

Lettura batch e proprietà additive in [API_SPEC](../03-architecture/LLD/API_SPEC.md). Polling sequenziale a blocchi di 50, ogni 5s dopo fine richiesta, cancellazione richieste al teardown, nessun ripoll dei terminali; errore visibile e retry. Non richiede una connessione SSE per ciascun target e non dipende dalle prime 50 scansioni dello storico. Snapshot volatili eliminati al termine; nessuna migrazione DB. Limiti: parallelismo condiviso, CIDR, reti filtrate, NSE e correlazione rendono i tempi approssimativi; nessun benchmark reale effettuato.

### Comando Nmap

- Editor libero di opzioni, con prefisso `nmap` facoltativo; esempio `nmap -sS -sV -p 22,80,443 -T3`. Target scelti con gruppo aziendale, separatamente dal comando.
- Il profilo rapido selezionato viene precaricato al passaggio all'editor. Dettagli opzionali: carica profilo custom, nome e timeout 1–1440 minuti. Le nuove configurazioni sono persistite nel tenant; il vecchio controllo «salva» non descriveva correttamente il salvataggio già obbligatorio per eseguire.
- Anteprima unica delle opzioni inviate, con output XML gestito e placeholder per target. Feedback client preliminare; validazione definitiva backend al salvataggio e nuovamente all'esecuzione.
- Riutilizzo profilo soltanto per argomenti equivalenti (normalizzazione spazi, `-p` e `--script`) e identico timeout; se l'operatore specifica un nome, deve coincidere. Mai selezione per corrispondenza di nome/categoria.
- Profili di sistema storici, ID e categorie API restano invariati per compatibilità di scansioni/pianificazioni esistenti. Non vengono mostrati come catalogo nella nuova modale; nessuna riscrittura dei dati custom del tenant.

### Scope, stati e accessibilità

- Gruppi ricorsivi dell'azienda attiva; nessun fallback a tutte le aziende se il nodo attivo manca. Target caricati con paginazione completa, blacklist esclusa dal conteggio eseguibile; on-demand il server ricalcola i target del gruppo.
- Loading/error/empty bloccano invio. Label associate, errori con `role=alert`, contenimento e ritorno del focus, Escape, dettagli chiusi per ridurre densità, contenuto scrollabile e larghezza adattiva.
- Scheduler mantiene cron e target discendenti non bloccati; fallimenti parziali mostrati, ritentati solo i target mancanti nella stessa configurazione e nella stessa apertura. Non garantisce idempotenza in caso di timeout di rete con esito remoto ignoto.

## 3. Impatto e autorizzazione

`ClassifyInvasiveness` classifica server-side i nuovi profili: NSE vuln/vulners, tutte le porte (`-p-`, `-p1-65535`, `-p0-65535`), T4/T5, version-all, ACK o min-rate sono INTRUSIVE; altri SAFE per compatibilità del contratto. SAFE non è una garanzia di assenza di impatto; la classificazione è euristica e non una policy completa di rate limiting o valutazione del rischio della rete.

TECH-DEBT-05 · 2026-10-19 · working tree: PO selected explicit consent for every
inventory scan, including configurations classified SAFE. Manual target/group and
recurring creation compare the operator-confirmed target IDs/values, Nmap options,
timeout and recurrence against live tenant inventory/profile. Missing or stale
consent returns 409 `SCAN_AUTHORIZATION_REQUIRED` before any launch. Actor and UTC
acceptance time come from authenticated server context. The UI resets confirmation
on configuration/scope changes; quick and group entry points follow the same rule.

Each job stores its target-bound authorization; migration 000013 writes append-only
audit evidence in the same transaction as scan/schedule creation. Workers compare
the snapshot again before Nmap. Scheduled triggers reload current state, reject
legacy/stale authorization and inactive/deleted schedules; queued recurring jobs
also recheck their source schedule. Runtime timestamp writes cannot re-enable a
concurrently disabled schedule. Next-run calculations use the same server-local
clock as cron, stored as UTC instants. Existing schedules require suspension and
replacement with explicitly authorized configuration; the UI labels them
"Autorizzazione richiesta" and blocks reactivation. Historical completed scans
remain readable; legacy queued scans fail before Nmap. An already-started process
is not recalled by later schedule changes. Consent is an operator attestation,
not proof of legal ownership or regulatory compliance. Standalone LAN discovery
retains its separate existing contract. [Design and evidence](../04-development/plans/TECH-DEBT-05.md).

## 4. Network Discovery e sovrapposizione

DISCOVERY-01 · 2026-10-17 · working tree: normalizzazione MAC unicast valida;
cache identità riutilizzata soltanto con MAC corrente identico e TTL di 15 minuti.
Il DNS fallback contribuisce agli indizi vendor; il vendor Nmap precede le euristiche.
Software generico (Windows, Linux, nginx, Samba) non identifica il produttore hardware.
Probe HTTP limitato a 2 secondi per host, senza redirect e con URL IPv6 corretto.
Nessuna modifica API/DB o aumento dei probe Nmap; inferenze di marca restano euristiche,
OUI identifica l'assegnatario dell'interfaccia, non necessariamente il produttore del dispositivo.

- Stage 1: `-sn -PE -PS21,22,80,443,445,8080,8443,3389,53,8000 -PU53,137,161,5353 -T4`, con discovery host, DNS e arricchimento MAC/OUI/cache. `-sn` non effettua port scanning.
- Stage 2: `-T4 -F -n` sugli host attivi, soltanto se `include_common_ports=true` e da 1 a 64 host, concorrenza massima 8. Da SCAN-UX-01 l'opzione è realmente rispettata: in precedenza era ignorata.
- Esiste sovrapposizione tecnica fra Stage 2 e il profilo TCP 100; i flussi restano distinti: Discovery inventaria/importa asset, scansioni produce esecuzioni persistite, finding, storico e pianificazioni sugli asset censiti. I profili host-only sono tolti dalla modale, non dal contratto storico.

## 5. API e streaming

M11-PERSIST-01 · 2026-10-18 · working tree: il server inietta un repository
PostgreSQL discovery. Sessione creata prima di Nmap; ogni dispositivo viene
confermato su DB prima dell'evento SSE; completamento persistito prima dell'evento
terminale. Nessun fallback alla memoria se il DB fallisce. ID sessione riutilizzato
→ 409 prima dello scanner. Letture e scritture filtrate per tenant; topic live
includono tenant e sessione, anche se il client si sottoscrive prima dell'avvio.

Il runner collega `StreamNmapXML` a stdout durante il processo tramite pipe:
callback per ciascun elemento host XML completo, non per singolo pacchetto/probe.
Host up discovery salvati subito con identità osservata; DNS/HTTP/porte aggiornano
lo stesso IP dopo lo sweep. UI sostituisce il dispositivo esistente, senza duplicarlo.
L'orchestratore correla/persiste anche i finding al completamento XML di ciascun
host; errore DB interrompe il job, senza pubblicare finding non confermati.
Il parsing finale supporta runner legacy senza duplicare gli host già processati.
Un errore callback cancella il processo; timeout/cancellazione mantengono i dati
già confermati e non pubblicano successo. XML finale mantenuto per statistiche.

Replay SSE: sottoscrizione prima della lettura dello snapshot; sono possibili
duplicati dello stesso IP, gestiti come aggiornamenti dal client. Un errore DB nel
replay produce errore visibile e chiusura. Ultimi risultati: sessione più recente
per data avvio/ID, anche incompleta, recuperabile da una nuova istanza del servizio.
`completed_at` zero indica snapshot parziale; UI lo esplicita e permette ricarica.
Solo HTTP 404 è archivio vuoto: errori di rete/auth/DB restano visibili. Nessuna
ripresa automatica del processo dopo restart; nessun nuovo probe LAN eseguito da
questi test. Trasporto live ancora locale all'istanza; replay durevole su riconnessione.
Schema/rollback nel [database](../03-architecture/LLD/DATABASE_SCHEMA.md);
contratto e gate nel [piano](../04-development/plans/M11-PERSIST-01.md).

- `GET /api/v1/scan-profiles`: profili sistema/custom tenant.
- `POST /api/v1/scan-profiles`: nome, nmap_args, descrizione, timeout_seconds, categoria; invasività calcolata server-side.
- `DELETE /api/v1/scan-profiles/{id}`: eliminazione custom tenant; contratto mantenuto, azione non esposta nella modale minimale.
- `POST /api/v1/scans`: target_id o group_id e profile_id; scansione ricorsiva per gruppo. `POST /api/v1/schedules`: target_id, profile_id, cron_expression, is_active.
- `GET /api/v1/scans/{id}/stream`: SSE con stream-ticket monouso, mai JWT permanente in URL.
- `GET /api/v1/scans/{id}/diff`: NEW/RESOLVED/PERSISTENT, confronto con precedente scansione completata sullo stesso target, chiave `(HostIP, Port, Protocol, FindingKey)`. «RESOLVED» significa non rilevato nel confronto, non prova autonoma di remediation se copertura/condizioni differiscono.

## 6. Confini e dipendenze

Target soggetti a validazione, deny-list e isolamento tenant. Scanner non fornisce automaticamente exploitation, verifica manuale, assessment autenticato o conformità ISO/NIS2. Dipendenze: [asset/reporting](reporting_and_assets.md), [CVE](vulnerability_intelligence.md), [attack path](attack_path.md).

## 7. Verifiche e fonti

Test: `backend/internal/scanner/profile_semantics_test.go`, `profiles_test.go`, `nmap_runner_test.go`, test dei servizi scan/discovery, handler; frontend `NewScanModal.test.tsx` e `scans/scanConfiguration.test.ts`.

Analisi, piano, fonti ufficiali Nmap/NIST e limiti di validazione in [NMAP_SCAN_REVIEW](../04-development/NMAP_SCAN_REVIEW.md). Nessuna migrazione DB né nuova dipendenza. Test di processo/parser usano fixture; nessuna scansione reale autorizzata o eseguita.

## 8. Ultimo allineamento

2026-10-17 · working tree · DISCOVERY-01 / SCAN-UX-01 / SCAN-TIME-01. Backend suite/vet/build PASS; [analisi discovery](../04-development/NETWORK_DISCOVERY_REVIEW.md). Nessuna scansione di laboratorio eseguita per DISCOVERY-01. Verifica visuale precedente non eseguita: inventario browser vuoto nel runtime.
