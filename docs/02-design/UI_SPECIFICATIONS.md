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
