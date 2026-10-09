# Dominio: Debug & Testing Live Console Drawer

> **Dominio:** Osservabilità Client-Side, Diagnostica Interattiva ed Esportazione per AI  
> **File di Riferimento:** `frontend/src/types/debug.ts`, `frontend/src/utils/debugLogger.ts`, `frontend/src/utils/markdownExporter.ts`, `frontend/src/context/DebugContext.tsx`, `frontend/src/components/debug/DebugConsoleDrawer.tsx`, `frontend/src/components/debug/DebugFloatingToggle.tsx`

---

## 1. Architettura del Logger Client-Side

### 1.1 In-Memory Ring Buffer
- Singleton `debugLogger` con capacità circolare fissa a 500 entry (FIFO eviction) per garantire zero perdite di memoria.
- Livelli supportati: `info`, `warn`, `error`.
- Sorgenti supportate: `api` (Axios), `sse` (Server-Sent Events), `ws` (WebSocket SSH), `runtime` (Window Error / Unhandled Rejection), `app` (log personalizzati).

### 1.2 Intercettazione Trasparente e Zero Impatto
- Interceptor Axios su richiesta, risposta ed errore con calcolo preciso dei tempi di risposta (`durationMs`).
- Event listener globali su `window.onerror` e `window.onunhandledrejection`.
- Sanificazione ricorsiva dei dati sensibili (`password`, `access_token`, `refresh_token`, `secret`, `Authorization`) sostituiti con `"[REDACTED]"`.

---

## 2. Esperienza Utente & Drawer Interattivo

- Accessibile in qualunque schermata tramite:
  - Pulsante fluttuante `DebugFloatingToggle.tsx` con badge numerico di errori non letti.
  - Scorciatoia globale da tastiera `Ctrl + Shift + D`.
- Slide-over panel fluido con barra di ricerca full-text e filtri istantanei per sorgente e livello.
- Dettaglio interattivo di ogni evento con JSON syntax highlighting per payload di richiesta e risposta.

---

## 3. Esportazione Diagnostica per AI

- Funzionalità a un click: pulsante *"Copia per AI (Markdown)"*.
- Genera un prompt preformattato completo di:
  1. Header diagnostico (Timestamp, rotta corrente, User Agent, conteggio errori).
  2. Tabella riassuntiva degli errori.
  3. Dettaglio completo degli stack trace e dei corpi di errore HTTP.
- Consente al Product Owner di incollare direttamente i log nella chat con l'agente per una risoluzione immediata dei problemi senza dover aprire i DevTools del browser.

