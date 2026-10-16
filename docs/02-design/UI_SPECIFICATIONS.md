# UI Specifications

> Responsabile: UI/UX Designer. Baseline ricavata da `frontend/tailwind.config.js`, `src/App.tsx` e documentazione dei domini; non costituisce un nuovo audit UI.

## Sistema visivo esistente

React 18, Tailwind CSS, icone Lucide. Token `cyber`: dark `#0a0d14`, card `#111726`, border `#1e293b`, accent `#06b6d4`, hover `#0891b2`. Severity: critical `#dc2626`, high `#ea580c`, medium `#eab308`, low `#3b82f6`, info `#64748b`.

## Componenti e stati richiesti

- Layout: navigazione desktop/mobile, indicazione azienda attiva, contenuto e fallback di caricamento.
- Form: label, validazione contestuale, errore remoto, stato di invio e protezione da invii duplicati.
- Tabelle e card: dati, loading, empty, error e azioni coerenti con i permessi.
- Modali/drawer: titolo, focus, tastiera, chiusura coerente e ritorno del focus. Questi sono criteri da verificare, non conformità già dimostrata.
- Operazioni invasive: spiegare impatto e consenso; mostrare conferma e recupero dove previsti.

## Specifiche per incremento

Collegare requisito/storia, route, componenti interessati, comportamento responsive, stati, accessibilità, criteri visuali e screenshot di verifica. Usare Mermaid per flussi e sequenze; usare wireframe o immagini quando servono a definire il layout.

Dettaglio per funzionalità: [DOMAIN_MAP](../03-architecture/DOMAIN_MAP.md). L'accettazione di Epic 27.1 resta indipendente dalla migrazione documentale.

## Nuova scansione Nmap · SCAN-UX-01

SCAN-COPY-01 · 2026-10-16: descrizioni dei profili limitate allo scopo della scansione e alle attività eseguite. Niente confronti con altre funzionalità o elenchi di ciò che il profilo non fa. Nessun suggerimento Discovery sotto il selettore; tempi e controlli mantengono la propria collocazione.

Ultimo allineamento: 2026-10-15 · working tree. Sostituisce il disegno Epic 27.1 su richiesta del PO.

Modale compatta (massimo 42rem, altezza 94vh, scroll interno). Ordine: azienda/reparto e conteggio target; modalità Profili o Comando Nmap; selettore di profilo oppure textarea; anteprima; pianificazione opzionale; eventuale consenso; azione finale. Dettagli salvati/timeout e limiti collassati. Nessun catalogo per categorie o compositore duplicato. Dal 2026-10-16, su richiesta PO, presente durata indicativa per IP con ipotesi esplicite (SCAN-TIME-01).

Profili nel [dominio scanner](../domains/scanning_and_discovery.md). Il comando precarica esattamente il preset; modifiche libere senza alterare il preset canonico. Pianificazione con fuso del server e segnalazione fallimenti parziali.

Loading/error/empty bloccano invio; form bloccato durante submit. Label associate, errori annunciati, focus contenuto e restituito alla chiusura, Escape, aria-pressed. Test componenti non equivalgono a conformità WCAG completa. Verifica visuale desktop/mobile pendente: browser indisponibile nel runtime.

## Progresso del gruppo · SCAN-TIME-01

Accettazione funzionale PO del 2026-10-16: scansioni testate e dichiarate correttamente funzionanti; chiusi i gate pendenti di SCAN-UX-01/SCAN-TIME-01 con evidenza del PO. Le note di indisponibilità browser sotto descrivono i limiti delle verifiche autonome dell'agente, non un blocco attuale della consegna.

Ultimo allineamento: 2026-10-16 · working tree. Il PO conferma scope intero gruppo. Card compatta nella pagina Scansioni per ultimo lotto lanciato: percentuale stimata, barra accessibile, conteggi per stato e intervallo residuo quando disponibile. Aggiornamento ogni 5s; nessun punto percentuale inventato in assenza di telemetria. Messaggio «in calcolo» in coda, statistiche scadute o fase senza ETA. Il 100% indica tutti i job terminati e mantiene visibili errori/annullamenti. Ricaricamento recupera ID lotto della sessione tenant; la card non aggrega lo storico e non viene mescolata con i filtri tabellari.

Dettaglio singolo continua a mostrare SSE e barra totale del job (senza la vecchia barra duplicata), distinguendo percentuale della fase da stima totale. Test automatici includono polling, ripristino sessione, gruppi oltre 50 job, errori e assenza di ETA; visuale reale ancora non disponibile.
