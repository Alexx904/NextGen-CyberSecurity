# 3. Interactive Task List & Implementation Backlog

> **Stato:** APPROVATO (Fase 2: System Design - Backlog Attivo Milestone 1.2: Real-Time Live Streaming)  
> **Ultimo Aggiornamento:** 2026-10-09  
> **Autore / Ruolo:** TechLead-Engineer-Agent (System Architect & Tech Lead)  
> **Documenti Correlati:** [0_REQUIREMENTS_AND_USE_CASES.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/0_REQUIREMENTS_AND_USE_CASES.md) | [1_PROJECT_CONTEXT.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/1_PROJECT_CONTEXT.md) | [2_ARCHITECTURE.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/2_ARCHITECTURE.md) | [4_CHANGELOG.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/4_CHANGELOG.md)

---

## 📋 Regole di Avanzamento (Fase 3 SDLC)
1. **Micro-Planning obbligatorio (`/plan [task-id]`):** Prima di toccare qualsiasi riga di codice, deve essere formulato e approvato un piano di implementazione dettagliato con firme, file coinvolti e Test Strategy esplicita.
2. **Esecuzione (`/execute`):** Solo dopo l'approvazione del piano si scrive il codice di produzione con i relativi test unitari/integrazione.
3. **Esecuzione & Validazione Test (Zero-Assumption Rule):** Prima di completare il task, i test devono essere eseguiti tramite shell con esito positivo al 100% (exit code 0). È formalmente vietato spuntare il task senza evidenza comprovata di successo.
4. **Aggiornamento di Stato:** Il task viene spuntato `- [x]` al completamento della Definition of Done (DoD) e verifica dei test.
5. **Archiviazione:** I task TASK-01 fino a TASK-50 completati nelle Milestone 1 e 1.1 sono archiviati in [4_CHANGELOG.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/4_CHANGELOG.md).

---

## ⚡ EPIC 18: Real-Time Live Streaming Engine via Server-Sent Events (Feature 6)

### Story 18.1: Modello Eventi SSE e In-Memory Event Broadcaster Hub
- [x] **TASK-51: Modello Eventi SSE e Servizio In-Memory Event Broadcaster (Backend Go)**
  - **File coinvolti:** `backend/internal/domain/event.go`, `backend/internal/service/event_broadcaster.go`, `backend/internal/service/event_broadcaster_test.go`
  - **Dipendenze:** TASK-06
  - **DoD:** Creazione del modello di dominio `event.go` con tipizzazione per `DeviceDiscoveredEvent`, `ScanProgressEvent`, `ScanPhaseEvent`, `ScanFindingEvent`, `ScanStatusEvent`. Creazione del servizio thread-safe `EventBroadcaster` con canali Go multiplexati per tenant/sessione/scan, registrazione subscriber, buffer non bloccante con drop policy controllata e chiusura pulita su `context.Done`. Unit test completi sul broadcast concorrente con coverage >= 90%.

### Story 18.2: Streaming in Tempo Reale per LAN Network Discovery
- [x] **TASK-52: Live LAN Discovery Engine e Streaming Endpoint SSE (Backend Go)**
  - **File coinvolti:** `backend/internal/domain/discovery.go`, `backend/internal/service/discovery_service.go`, `backend/internal/delivery/http/handler/discovery_handler.go`, `backend/internal/service/discovery_service_test.go`, `backend/cmd/server/main.go`
  - **Dipendenze:** TASK-51, TASK-43
  - **DoD:** Aggiornamento del motore di discovery per emettere eventi push non appena un host risponde al probe (con risoluzione OUI vendor immediata) e aggiornamento della percentuale di sweep. Implementazione endpoint SSE `GET /api/v1/network-discovery/stream` con flushing `http.Flusher`, header `text/event-stream`, autorizzazione JWT e persistenza incrementale dei dispositivi scoperti nella sessione. Unit test completi.

### Story 18.3: Streaming Fasi, Progresso e Scoperte Progressive Scansioni Nmap
- [x] **TASK-53: Streaming Progressivo Scansioni Nmap e Scrittura Incrementale Findings (Backend Go)**
  - **File coinvolti:** `backend/internal/service/scan_orchestrator.go`, `backend/internal/scanner/nmap_runner.go`, `backend/internal/delivery/http/handler/scan_handler.go`, `backend/internal/service/scan_orchestrator_test.go`, `backend/cmd/server/main.go`
  - **Dipendenze:** TASK-51, TASK-35
  - **DoD:** Modifica di `scan_orchestrator` e `nmap_runner` per monitorare l'esecuzione Nmap in tempo reale con flag `--stats-every` e streaming delle porte scoperte. Salvataggio incrementale su PostgreSQL di ogni finding/porta aperta correlata senza attendere il termine del processo; emissione eventi `phase`, `progress`, `finding_discovered` e `status_change` verso `EventBroadcaster`. Implementazione endpoint SSE `GET /api/v1/scans/{id}/stream`. Unit test con simulazione eventi live.

### Story 18.4: Integrazione Frontend Real-Time LAN Discovery
- [x] **TASK-54: Client SSE e Aggiornamento Tabella Live in Network Discovery (Frontend React)**
  - **File coinvolti:** `frontend/src/api/discovery.ts`, `frontend/src/pages/NetworkDiscovery.tsx`, `frontend/src/hooks/useSSE.ts`
  - **Dipendenze:** TASK-52
  - **DoD:** Creazione hook riutilizzabile o client `EventSource` con autenticazione JWT e gestione auto-reconnection. Aggiornamento di `NetworkDiscovery.tsx` in modo che all'avvio della scansione la tabella venga popolata in tempo reale con animazione riga per riga per ogni dispositivo emesso dall'evento `device_discovered`, con barra di avanzamento e conteggio incrementale. Mantenimento delle azioni rapide (aggiungi a cartella target / scansione immediata) operative sui nodi già comparsi a schermo.

### Story 18.5: Integrazione Frontend Real-Time Scan Detail
- [x] **TASK-55: Monitoraggio Live di Scansione e Findings Progressivi in ScanDetail (Frontend React)**
  - **File coinvolti:** `frontend/src/pages/ScanDetail.tsx`, `frontend/src/components/ScanProgressBadge.tsx`, `frontend/src/api/scan.ts`
  - **Dipendenze:** TASK-53
  - **DoD:** Aggiornamento di `ScanDetail.tsx` per agganciarsi allo stream SSE `/api/v1/scans/{id}/stream` quando la scansione è `RUNNING` o `QUEUED`. Visualizzazione in diretta della fase corrente, percentuale animata, e inserimento immediato dei finding nella tabella Findings man mano che arrivano, senza necessità di polling o refresh manuale della pagina.

### Story 18.6: Collaudo E2E e Verifica delle Performance Real-Time
- [x] **TASK-56: Test di Integrazione e Validazione E2E per Funzionalità Real-Time (Milestone 1.2)**
  - **File coinvolti:** `docker-compose.yml`, test scripts / unit test
  - **Dipendenze:** TASK-51 fino a TASK-55
  - **DoD:** Test end-to-end con:
    1. Avvio scansione LAN Discovery e ricezione streaming push dei dispositivi con latenza < 200ms.
    2. Verifica persistenza incrementale su refresh browser.
    3. Avvio scansione Nmap con ricezione eventi di cambio fase, progresso e finding live.
    4. Verifica disconnessione/riconnessione client senza interruzione del job in background.

---

## 🛠️ EPIC 19: Debug & Testing Live Console Drawer (Feature 7)

### Story 19.1: Core Logging Engine, Redaction e Ring Buffer
- [x] **TASK-57: Ring Buffer Logger, Sensitive Data Sanitizer & Event Bus (Frontend TypeScript)**
  - **File coinvolti:** `frontend/src/types/debug.ts`, `frontend/src/utils/debugLogger.ts`, `frontend/src/utils/debugLogger.test.ts`
  - **Dipendenze:** Nessuna
  - **DoD:** Definizione tipi `DebugEntry`, `DebugLogLevel`, `DebugLogSource`. Implementazione del singleton thread-safe `debugLogger` con memoria circolare (max 500 eventi), rimozione FIFO degli elementi in eccesso, mascheramento ricorsivo di token e password sensibili (`[REDACTED]`), metodi helper `logApiRequest`, `logApiResponse`, `logApiError`, `logSSEEvent`, `logWSEvent`, `logRuntimeError`, `logCustom`, ed emissione eventi con listener pub/sub. Suite di unit test dedicata con copertura >= 90%.

### Story 19.2: Intercettazione Globale Rete ed Errori di Runtime
- [x] **TASK-58: Intercettore Trasparente Axios, Handlers Globali e DebugContext (Frontend React)**
  - **File coinvolti:** `frontend/src/api/client.ts`, `frontend/src/context/DebugContext.tsx`, `frontend/src/App.tsx`
  - **Dipendenze:** TASK-57
  - **DoD:** Integrazione degli interceptor Axios in `api/client.ts` per calcolare tempo di esecuzione (`durationMs`), status HTTP e payload di richiesta/risposta/errore verso `debugLogger`. Registrazione listener `window.addEventListener('error')` e `unhandledrejection`. Creazione di `DebugContext` e `DebugProvider` per distribuire lo stato reattivo dei log, il contatore errori non letti e la visibilità del pannello all'intera applicazione.

### Story 19.3: Interfaccia Utente Slide-Over, Filtri e Formattatore AI Markdown
- [x] **TASK-59: Componente UI Slide-Over Drawer, Floating Toggle & Markdown Exporter (Frontend React)**
  - **File coinvolti:** `frontend/src/components/debug/DebugConsoleDrawer.tsx`, `frontend/src/components/debug/DebugFloatingToggle.tsx`, `frontend/src/utils/markdownExporter.ts`, `frontend/src/components/layout/AppLayout.tsx`
  - **Dipendenze:** TASK-58
  - **DoD:** Realizzazione del drawer laterale destro con transizione CSS fluida, apribile tramite pulsante fluttuante con badge contatore errori o combinazione tasti `Ctrl + Shift + D`. Barra strumenti con filtri per livello (`ALL`, `ERROR`, `WARN`, `INFO`), sorgente (`API`, `SSE`, `WS`, `RUNTIME`), barra di ricerca rapida testuale. Visualizzatore espandibile per ciascun record di log con JSON formattato. Pulsante "Copia per AI (Markdown)" che genera un blocco formattato pronto per la chat, pulsante "Esporta JSON" e "Pulisci Log".

### Story 19.4: Build Type-Safety & Test di Integrazione
- [x] **TASK-60: Verifica Type-Safety e Collaudo E2E Console di Debug (Frontend React)**
  - **File coinvolti:** `frontend/src/App.tsx`, `frontend/package.json`, test scripts
  - **Dipendenze:** TASK-57, TASK-58, TASK-59
  - **DoD:** Esecuzione di `npm run build` (`tsc && vite build`) con esito 0 errori. Esecuzione suite unit test. Verifica funzionale dell'apertura del pannello, registrazione di una chiamata HTTP reale (es. `/api/v1/health`), simulazione di un errore 4xx/5xx con incremento del badge contatore ed emissione del report Markdown formattato.

