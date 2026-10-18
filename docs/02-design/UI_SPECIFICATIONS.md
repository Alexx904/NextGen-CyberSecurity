# UI Specifications

> Publication scope 2026-10-18: alert-channel / outbox / migration 000012 sections describe deferred T4/T5 design and local prototype evidence. Their code and routes are excluded from this release; published migration is 000011. Alert-specific commands are future gates.


E28-S1-T2 · 2026-10-18 · working tree: "Neutralizza Chokepoint" opens a focused,
scrolling mitigation dialog for supported SERVICE nodes. Explicit server choice,
backend preview, returned rule/command/timeout and impact acknowledgement precede
"Applica blocco porta". Missing pin/empty inventory/loading/unsupported nodes block
apply. A fresh server selection clears preview and consent. Success opens existing
FirewallManager; uncertain outcomes offer "Verifica stato firewall" without retry.
Auditor has no mutation action. Scope change dismisses an unapplied dialog.
Keyboard focus containment/Escape supported; complete browser gate pending.

M11-DISCOVERY-UI-01 · 2026-10-18 · working tree: discovery CSV/JSON buttons label
the exported count and visible/selected set. Selected scan action opens import
with company destination preselected, then existing scan configuration scoped
to those IPs; no launch occurs on import alone. Empty/blacklisted/out-of-scope
selection cannot fall back to the whole company. Confirmed partial launches are
not repeated; uncertain network/5xx outcomes block retry pending history review.
Toolbar wraps on smaller widths. Eight accepted scanner presets remain unchanged.

M11-FIREWALL-01 · 2026-10-18 · working tree: firewall policy ACCEPT/DROP,
full rule replacement and up/down ordering use the loaded revision. Pending
transactions disable mutations; expired pending state remains visible, polls
the host, and cannot be confirmed. Missing revision explains why edits are disabled.
Replacement warns about advanced flags; changing TCP/UDP to ICMP/all clears port.
Rule dialog supports Escape/Tab focus containment and returns focus to its opener.
Terminal opens lazily on request, docked below the firewall with optional full
screen; closing while ticket acquisition is pending cannot open a late session.
Browser check of this build is pending.

M11-REPORT-01 · 2026-10-18 · working tree: Report con filtri etichettati per giorni
UTC inclusivi, profilo, target e cartella; severità massima esatta. Opzioni dallo
storico completo nello scope. Periodo invertito con `role=alert`, reset filtri,
cronologico 10 righe/pagina e pulsanti precedente/successiva disabilitati ai limiti.
Nota esplicita: i download aziendali restano dossier completi. Caricamento parziale
o incoerente espone errore/riprova anziché metriche tronche. Layout filtri 1/2/5
colonne responsivo; prova visuale autonoma ancora aperta.

M11-CLOSE-01 · 2026-10-18 · working tree: nel form firewall il safety timer
è dichiarato obbligatorio e non esiste più la checkbox per disattivarlo.
Timeout modificabile 10..300s, default 60s. Errori API di validazione esposti
dal flusso esistente; test DOM della protezione obbligatoria. Nessun collaudo
visuale autonomo disponibile: inventario browser vuoto.

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
## Virtual patch panel (E28-S1-T3)

Finding detail opens a separate keyboard-accessible panel with product, catalog
revision/coverage, strategy and operational impact before the engine selector.
Fail2ban/CrowdSec/ModSecurity/Nginx previews have per-file copy/download. Missing
coverage has an explicit unsupported state without exports. Load/export errors
remain visible; Escape closes and restores focus, Tab cycles inside. Long text
and snippets wrap/scroll on mobile. [Behavior and verified limits](../domains/active_defense_and_soar.md#virtual-patch-templates-e28-s1-t3-in-progress).

## Notification channels in Settings (E28-S1-T5)

Tenant administrators see an independent channel card below branding. List,
loading/error/empty, inline form and two-step removal are keyboard operable.
New channel defaults disabled, tenant/company scope and explicit event selection;
provider fields explain TLS/public destination policy. Existing credentials are
never displayed: metadata edits omit configuration, replacement is explicit and
required on provider change. Test button requires acknowledgement and tests the
draft without saving/activating. Config changes clear acknowledgement/result.
Accepted test asks operator to check receipt; failure never claims delivery.
Draft secrets clear on close/save and do not enter diagnostic exports.
[Contract and limits](../domains/active_defense_and_soar.md#alerting-hub-e28-s1-t4t5-in-progress).
