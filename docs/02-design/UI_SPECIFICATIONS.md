# UI Specifications

## Current M4 - automatic finding context

The Findings table explains where to find automatic interpretation. Its existing
detail modal includes observed fact, technical meaning, suggested action, original
confidence/method, catalogue version and official links. Missing server context and
empty mappings have explicit messages without implying compliance. No manual forms
or compliance percentages are part of this feature. The previous Compliance nav is
removed; `/compliance` redirects to Findings. Reports retain normal Executive/Technical
exports with automatic context, without the manual Audit form. Sections below on
manual M4 are historical/superseded. [Gates](../04-development/plans/M4-AUTOMATIC-CONTEXT.md).

## M4 verified implementation - 2026-10-19

Guided follow-up: a "Cosa devo fare adesso?" queue precedes the scorecard, with
scope-wide technical tasks and framework-filtered control tasks. Identity forms
suggest historical same-group candidates and require confirmation; policy forms
apply named owner/SLA to explicitly selected issues with per-item results. Actions
focus the existing lifecycle/evidence editor. Visible idle pages refresh every 30s,
pausing during editing or open disclosures. Scan links use the existing consent
workflow. [Verification status](../04-development/plans/M4-WORKFLOW.md).

The Compliance page offers a native expandable first-use guide above its filters,
including score interpretation and the absence of automatic regulatory updates.
It remains accessible while data loads or fails, for read-only and write roles.

`/compliance` sits in assessment navigation, scoped by the global company and an
optional department. It displays framework/global score, denominators, unknowns,
exclusions, scan/target coverage and per-department cards. Evidence, stable asset
association, issue owner/policy/actions/history and in-app SLA breach are distinct
sections. Empty, loading, failed/retry and stale scope responses are handled;
Auditors see read/export only. Labeled fields, native disclosures and visible
keyboard focus support desktop/mobile use. Shared Audit form in Reports and
Compliance collects declared legal/reviewer/period data. [Semantics/limits](../domains/compliance.md).

## Planned integrated M4-M8 experience - 2026-10-19

M4-M8 extend assessment with company/department context; SSH/Firewall remains independent. The verified M4 screen is described above; this section describes remaining integrated expansion. UI-TOOLS-01 owns the final hub design.

- Enrich existing finding details with detection, context, applicability, remediation, verification history and control evidence; no duplicate inventory or report archive.
- Preserve Executive/Technical/CSV exports and add an Audit profile in the report hub. Show period/cutoff, scope, filters and coverage. Empty/partial/unknown data cannot imply safety or compliance.
- Distinguish detection confidence from applicability, validation and exploitation; distinguish AI advice, action applied, correction declared and correction reverified.
- Lifecycle/owner/deadline and linked observations need loading/error/empty/read-only states. Server authorization, deep links, company context, keyboard/focus and mobile behavior must be preserved.
- M7/M8 require explicit scope/technique/window confirmation, status/stop and cleanup/residual-effect results. No execution from a CVE or AI response without separate authorization. Missing provider/context cannot fabricate advice.

TECH-DEBT-09 · 2026-10-19 · working tree: Tailwind 4 migration retains custom
brand colors, v3 small shadows/blurs, transparent keyboard outlines and enabled
button pointer cursors. Supported browser minimums: Safari 16.4, Chrome 111,
Firefox 128. Automatic gates and connected Chromium desktop/390px visual and
keyboard checks pass; PO accepted the delivery on 2026-10-19. Evidence and browser limits are
listed in the [dependency assessment](../05-qa/TECH-DEBT-09.md#browser-completion---2026-10-19).

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

2026-10-19 PO change: removed "Genera virtual patch" and its modal mount/import
from Finding details. The generator has no active UI entry point. The panel behavior
below is historical; backend template endpoints are unchanged by this UI removal.

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

## TECH-DEBT-05 - Mandatory scan consent (2026-10-19, working tree)

All inventory scan configurations require the ownership/authorization checkbox,
including SAFE profiles. Changes to command, target selection, group, timeout or
recurrence reset it. Scheduled consent text explains the configuration-bound
recurring scope. Quick scans bind their exact target/profile; recursive group
entry uses the main scan dialog. API rejections appear in the dialog without
launching or closing it. Legacy/stale schedules show "Autorizzazione richiesta",
explain suspension/recreation, and cannot be reactivated without current consent.
Disabled launch controls remain outside the keyboard focus cycle. No standalone
LAN discovery consent redesign is included. [Contract](../04-development/plans/TECH-DEBT-05.md).
