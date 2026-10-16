# Nmap scan workflow review

Ultimo allineamento: 2026-10-15 · working tree · task SCAN-UX-01.

## Obiettivo e criteri

Ridurre Nuova Scansione a Profili e Comando Nmap. Rimuovere il compositore multi-obiettivo e il catalogo per categorie dalla modale; mantenere scope aziendale, blacklist, pianificazione e riuso dei profili tenant. Nessuna scansione reale di reti durante la verifica.

La UI descrive porte, protocolli e sonde effettive. Port scanning, service detection, vulnerability scanning e penetration testing sono attività distinte (NIST SP 800-115). Nmap non costituisce da solo un penetration test, un assessment applicativo OWASP o una certificazione di conformità. SYN è rilevabile; ACK distingue filtered/unfiltered, non open/closed. `--script=vuln` include controlli potenzialmente intrusivi: nessuna promessa di sicurezza o copertura completa. Versioni e CVE correlate richiedono verifica, anche per backport e falsi positivi/negativi.

## Contratti prima dell'implementazione

- API invariata: `POST /scan-profiles` riceve argomenti, nome, timeout e categoria CUSTOM; `POST /scans` riceve group_id/profile_id; scheduler usa target_id/profile_id/cron_expression. Nessuna migrazione DB.
- Editor: accetta opzioni oppure prefisso letterale `nmap`; target selezionati separatamente, output XML e binario gestiti dal server. Nessuna shell, file, target inline o script arbitrario. Allowlist server autorevole; controllo client preliminare senza badge di validazione server anticipata.
- Profili rapidi definiti da argomenti espliciti: riutilizzo soltanto per equivalenza degli argomenti, mai per nome o categoria. Anteprima e payload condividono la stessa configurazione.
- Parser server: rifiuta tecniche TCP incompatibili, selettori porte multipli, porte fuori range/invertite, top-ports fuori range e combinazioni incoerenti di host-only, ACK, versioning e output open-only. Conserva combinazioni TCP+UDP valide.
- Invasività custom determinata server-side dagli argomenti e persistita; UI richiede conferma per NSE, tutte le porte, timing T4/T5, version-all e ACK. La classificazione SAFE legacy non significa assenza di rischio.
- Target: caricamento paginato, discendenti del gruppo e blacklist esclusa dal conteggio eseguibile; nessun fallback ad altre aziende quando lo scope attivo manca. Stato di caricamento/errore blocca invio.

## Piano e verifiche

1. Test rossi per parser semantico e classificazione custom; implementazione in scanner/profiles.go e service/scan_service.go.
2. Sostituzione della modale con selettore compatto di profili, editor libero, dettagli opzionali e preview. Test su argomenti reali, profili discordanti, autorizzazione, scope, errori e cron.
3. Suite Go, build, vet; suite frontend e build; controllo documentale e whitespace. Verifica browser locale se disponibile, senza avviare scansioni reali.
4. Doc Sync di dominio/requisiti/UI/HLD/stato/task; self-review di injection, limiti parser, autorizzazioni e isolamento tenant.

## Fonti tecniche

- [NIST SP 800-115](https://csrc.nist.gov/pubs/sp/800/115/final): distingue discovery, vulnerability scanning e penetration testing; guida metodologica, non attestazione di conformità del prodotto.
- [Nmap scan techniques](https://nmap.org/book/man-port-scanning-techniques.html): SYN, ACK, UDP e relativi stati.
- [Nmap NSE usage](https://nmap.org/book/nse-usage.html): categorie di script sovrapposte e rischi.
- [Nmap host discovery](https://nmap.org/book/man-host-discovery.html): `-sn` evita il port scan; `-Pn` salta la discovery, non aggira un firewall.

## Baseline

Working tree iniziale pulita su main. Branch locale feat/nmap-scan-workflow-review. Baseline frontend: 14 test modale PASS; backend scanner/service PASS (exit 0). Esecuzione con permessi estesi per limiti sandbox su esbuild/cache Go; nessuna dipendenza aggiunta.

## Risultati della review tecnica

| Evidenza | Valutazione | Intervento |
|---|---|---|
| GUIDED, MULTI_OBJECTIVE e PROFESSIONAL costruivano combinazioni degli stessi flag | Ridondanza UX reale | Due modalità, profili espliciti + editor; rimossi compositore e catalogo duplicati |
| Card generalista SAFE senza NSE poteva selezionare All-in-One con --script=vuln; quick poteva selezionare profili -sn | Incoerenza tra descrizione e azione | Matching solo su argomenti e timeout, fonte unica per preview/payload |
| Penetration Test, CVE Exploit, bypass e compliance | Termini non sostenuti dalla copertura | Nuovi nomi tecnici per porte/servizi/NSE/ACK; mantenuti ID storici fuori dal catalogo UI |
| Discovery Stage 2 usava -F anche con include_common_ports=false | Traffico non richiesto e sovrapposizione non dichiarata | Test rosso/verde e rispetto del flag; sovrapposizione TCP 100 spiegata |
| Parser accettava -sS -sA, -p80 -p443, porte invalide e ACK --open | Configurazioni incoerenti | Test rosso/verde; controlli semantici client/server |
| Nuovi profili custom non assegnavano Invasiveness | Metadati non coerenti con gli argomenti | Classificazione server e test della persistenza |
| Consenso solo UI, nessun campo di attestazione in scans/schedules | FR-01.3 incompleto, aggirabile con API diretta | Gap dichiarato e TECH-DEBT-05; non certificata conformità o autorizzazione immutabile |

La scansione TCP 100 resta utile per risultati persistiti/pianificati sugli asset censiti; per il semplice inventario si usa Discovery. Con arricchimento porte attivo in Discovery esiste sovrapposizione tecnica, non equivalenza dell'intero flusso. Il bug del flag è stato incluso nello scope perché cambia direttamente il confronto richiesto dal PO.

Valutazione: configurazione e terminologia più coerenti con un uso professionale di Nmap; nessuna validazione completa del motore di vulnerability assessment o attestazione di conformità. Copertura NSE dipendente dall'immagine installata, falsi positivi/negativi e backport da verificare in laboratorio. Nessuna scansione reale o test su reti terze.

Verifica visuale dell'agente tentata con browser tool: inventario apps/browsers vuoto, creazione iab restituisce «Browser is not available: iab». Il 2026-10-16 il PO riferisce test delle scansioni riusciti e accetta SCAN-UX-01, autorizzando il push. Gate funzionale chiuso con evidenza del PO; nessuno screenshot o PASS visuale autonomo inventato. TECH-DEBT-05 rimane aperto.

## Pacchetto di revisione

Come provarla: avviare l'ambiente locale previsto dal progetto, aprire Scansioni → Nuova scansione. Selezionare azienda/reparto e un profilo; controllare le opzioni in anteprima. Passare a Comando Nmap, modificare ad esempio `nmap -sS -sV -p 22,80,443 -T3`; il target non va nel comando. Provare un comando incoerente (`-sS -sA`), poi `--script=vuln`: il primo blocca l'invio, il secondo richiede conferma. Verificare layout desktop/mobile prima dell'accettazione; l'avvio effettivo va provato solo su laboratorio esplicitamente autorizzato.

Evidenze finali: da frontend `npm.cmd test` PASS (19 file, 136 test, exit 0), `npm.cmd run build` PASS (exit 0, 7,30 s); da backend `go test ./...`, `go build ./...`, `go vet ./...` PASS (exit 0). Dalla root `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify-agent-docs.ps1` PASS (exit 0: 54 Markdown, 114 link/ancore). Whitespace verificato separatamente con `git diff --check`.

Documenti aggiornati:

- `docs/01-product/REQUIREMENTS.md`: FR-16/17 e terminologia professionale.
- `docs/02-design/UI_SPECIFICATIONS.md`: layout minimale, stati e accessibilità.
- `docs/03-architecture/HLD.md`: flusso frontend/backend compatibile.
- `docs/03-architecture/DOMAIN_MAP.md`: descrizione del dominio aggiornata.
- `docs/03-architecture/ADR/README.md`: ADR-14, scelta e compatibilità storica.
- `docs/domains/scanning_and_discovery.md`: contratti effettivi, profili, Discovery e limiti.
- `docs/04-development/NMAP_SCAN_REVIEW.md`: analisi esperta, piano, fonti e risultati.
- `docs/04-development/TASK_LIST.md`: accettazione SCAN-UX-01 dopo prova funzionale PO e TECH-DEBT-05.
- `docs/05-qa/TEST_PLAN.md`: casi, riproduzioni rosso/verde e confini delle evidenze.
- `docs/PROJECT_STATE.md`: consegna, assunzioni e blocchi senza accettazione implicita.

Assunzioni: T3 di default; TCP 100 mantenuto per esecuzioni persistite; nomi/ID storici non migrati; persistenza automatica delle configurazioni necessaria per l'API esistente. Nessuna nuova dipendenza. Limiti: visuale/E2E reale non eseguiti, attestazione consenso solo UI, classificazione impatto euristica, idempotenza remota non garantita nei timeout scheduler.
