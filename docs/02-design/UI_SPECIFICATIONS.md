# UI Specifications

SSH-HARDEN-01 · 2026-10-17 · working tree: modale SSH con campo opzionale
"Fingerprint SHA256 della chiave host", distinto dalle credenziali client.
Spiegazione inline: canale fidato obbligatorio per verificarlo; senza pin le
connessioni sono bloccate. Legacy modificabile e registrazione inventario senza
pin consentita. Cambio host/porta svuota il campo; rotazione tramite sostituzione
esplicita. Form scorrevole entro 90vh. Test DOM coprono enrollment/reset; prova
visuale autonoma non eseguita (browser runtime indisponibile).

> Responsabile: UI/UX Designer. Baseline ricavata da `frontend/tailwind.config.js`, `src/App.tsx` e documentazione dei domini; non costituisce un nuovo audit UI.

## Sistema visivo esistente

React 18, Tailwind CSS, icone Lucide. Token `cyber`: dark `#0a0d14`, card `#111726`, border `#1e293b`, accent `#06b6d4`, hover `#0891b2`. Severity: critical `#dc2626`, high `#ea580c`, medium `#eab308`, low `#3b82f6`, info `#64748b`.

## Componenti e stati richiesti

COMPANY-UX-02 · 2026-10-17: unica dashboard con CRUD aziende; route aziende precedente rediretta. Scope globale: sidebar operativa disabilitata e route guard; SSH/Firewall sempre accessibile da sezione Tool indipendente. Clienti SSH autonomi in form e tabella raggruppata con filtro. Target mostra sola gerarchia Azienda → Reparto, parent bloccato all'azienda corrente. Attack Path: selettore reparto/azienda, nessun filtro CVSS. Report PDF consolidato con inventario IP per reparto su pagine aggiuntive; CSV reparto additivo. COMPANY-PDF-03: nei PDF aziendali con reparti, anche schede CVE separate per reparto, interruzione pagina tra reparti e intestazione ripetuta nelle continuazioni. Tecnico completo; Executive massimo cinque priorità per reparto con conteggio. Riepilogo consolidato mantenuto. Atteso responsive; browser/PDF render ancora da collaudare. [QA](../05-qa/COMPANY_WORKFLOW_QA.md).

- Layout: navigazione desktop/mobile, indicazione azienda attiva, contenuto e fallback di caricamento.
- Form: label, validazione contestuale, errore remoto, stato di invio e protezione da invii duplicati.
- Tabelle e card: dati, loading, empty, error e azioni coerenti con i permessi.
- Modali/drawer: titolo, focus, tastiera, chiusura coerente e ritorno del focus. Questi sono criteri da verificare, non conformità già dimostrata.
- Operazioni invasive: spiegare impatto e consenso; mostrare conferma e recupero dove previsti.

## Specifiche per incremento

### AP-REVIEW-01 · 2026-10-17 · working tree

`/attack-path`: vista percorso 2D con sequenza sinistra→destra e scorrimento orizzontale per mantenere le dimensioni delle card; mappa completa con layout finito su cicli. Archi osservati continui, inferenze tratteggiate, ipotesi laterali arancioni; motivazione al passaggio del puntatore. Selezione tastiera Tab/Invio/Spazio, zoom/pan Pointer Events e adattamento vista. Nessuna animazione pulsante necessaria per leggere il grafo. Guida espandibile integrata nel prodotto, con prerequisiti, procedura, esempio e limiti.

Scope aziendale limita anche canvas, KPI e chokepoint; selezione vuota non ripiega su altri dati. Banner distingue ipotesi e osservazioni; etichetta indice euristico, avvisi di copertura parziale e metadati di provenienza. Stati loading/empty/error conservati; errori di aggiornamento rimuovono il grafo precedente. Desktop: grafo e pannello laterale; mobile: colonne impilate e canvas scorrevole. 3D non aggiunto per evitare occlusioni nella lettura della catena. Verifica browser reale ancora aperta: [report QA](../05-qa/ATTACK_PATH_REVIEW.md), nessuna conformità visuale dichiarata.

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
