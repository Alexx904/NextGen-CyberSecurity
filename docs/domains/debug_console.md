# Dominio: Debug & Testing Live Console Drawer

> **Dominio:** Osservabilità Client-Side, Diagnostica Interattiva ed Esportazione per AI  
> **File di Riferimento:** `frontend/src/types/debug.ts`, `frontend/src/utils/debugLogger.ts`, `frontend/src/utils/markdownExporter.ts`, `frontend/src/context/DebugContext.tsx`, `frontend/src/components/debug/DebugConsoleDrawer.tsx`, `frontend/src/components/debug/DebugFloatingToggle.tsx`

---

## 1. Architettura del Logger Client-Side

### 1.1 In-Memory Ring Buffer
- Singleton `debugLogger` con capacità circolare fissa a 500 entry (FIFO eviction) per garantire zero perdite di memoria.
- Livelli supportati: `info`, `warn`, `error`.
- Sorgenti supportate: `api` (Axios), `sse` (Server-Sent Events), `ws` (WebSocket SSH), `runtime` (Window Error / Unhandled Rejection), `app` (log personalizzati).

### 1.2 Intercettazione Trasparente e Sanificazione Dati Sensibili (`NFR-DEV-01`, `NFR-SEC-04`)
- Interceptor Axios su richiesta, risposta ed errore con calcolo preciso dei tempi di risposta (`durationMs`).
- Event listener globali su `window.onerror` e `window.onunhandledrejection`.
- **Sanificazione Ricorsiva & Redazione Query Params:**
  - Mascheramento a schermo e in memoria di password, passphrase, chiavi e token bearer (`Authorization: Bearer [REDACTED]`).
  - Regex dedicata `QUERY_SECRET_PARAM_REGEX` per mascherare query parameters sensibili (`ticket=`, `token=`, `passphrase=`, `secret=`, `api_key=`) in qualsiasi URL o log testuale (trasformati in `[REDACTED]`).
  - Sanificazione preventiva della rotta attiva (`activeRoute`) e passaggio di redazione globale su tutto il Markdown generato per AI.

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
  1. Header diagnostico (Timestamp, rotta corrente sanitizzata, User Agent, conteggio errori).
  2. Tabella riassuntiva degli errori.
  3. Dettaglio completo degli stack trace e dei corpi di errore HTTP senza alcun token, ticket o segreto esposto.
- Consente al Product Owner di incollare direttamente i log nella chat con l'agente per una risoluzione immediata dei problemi senza dover aprire i DevTools del browser né rischiare il leak di credenziali.

---

## 4. File chiave
- `frontend/src/types/debug.ts`
- `frontend/src/utils/debugLogger.ts`, `markdownExporter.ts`
- `frontend/src/context/DebugContext.tsx`
- `frontend/src/components/debug/DebugConsoleDrawer.tsx`, `DebugFloatingToggle.tsx`

## 5. Modello dati
- `DebugLogEntry`: `id`, `timestamp`, `level` (`info`, `warn`, `error`), `source` (`api`, `sse`, `ws`, `runtime`, `app`), `message`, `data`, `durationMs`.

## 6. API e contratti
- Componente esclusivamente client-side: nessuna API remota, nessun overhead o tracking server-side.

## 7. Flussi principali
1. Intercettazione automatica chiamate Axios, flussi SSE e WebSocket.
2. Applicazione filtri di redazione (password, bearer, query string secrets).
3. Memorizzazione nel ring buffer circolare (500 elementi).
4. Visualizzazione a drawer o esportazione markdown clipboard 1-click.

## 8. Regole di business
- Divieto assoluto di memorizzare o esportare chiavi di sicurezza, credenziali in chiaro, JWT o stream-ticket.
- Capacità massima fissa a 500 entry per evitare memory leak client-side.

## 9. Dipendenze da altri domini
- Intercetta chiamate ed eventi da tutti i domini applicativi (scanning, ssh/firewall, report, attack path).

## 10. Test
- Unit test Vitest: `frontend/src/utils/debugLogger.test.ts`, `frontend/src/utils/markdownExporter.test.ts`
- Component test: `frontend/src/components/debug/DebugConsoleDrawer.test.tsx`, `frontend/src/context/DebugContext.test.tsx`, `frontend/src/tests/debugConsoleE2E.test.tsx`

## 11. Ultimo allineamento
- 2026-10-10 (Epic 24 SEC-P0: Redazione token/ticket query params e sanitizzazione Markdown per AI implementata e verificata al 100%)

