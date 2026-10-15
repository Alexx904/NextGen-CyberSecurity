---
description: Hostile Code Review & Security Audit prima del merge o dell'accettazione finale
---

Segui le regole di `AGENTS.md`. Questo workflow esegue un audit avversario e di conformità architetturale sui commit o sul diff del branch corrente.

1. **Analisi del Diff.**
   - Esegui `git diff <branch-principale>...HEAD` e controlla la lista dei file modificati.
   - Verifica che nessun file estraneo o di configurazione sensibile (`.env`, certificati, chiavi private, log) sia stato incluso.

2. **Checklist di Sicurezza Avversaria (Hostile Checklist):**
   - **Injection:** Ci sono chiamate a comandi di sistema (`exec.Command`, `sh`) o query SQL concatenate senza parametrizzazione? (In Nmap: verificare separatore POSIX `--`; in SQL: query parametrizzate `$1, $2`).
   - **Autenticazione & Scope:** Tutti gli endpoint REST/SSE/WS sono protetti da middleware JWT o StreamTicket? I target controllati appartengono davvero all'azienda attiva (`CompanyContext` / `CheckTargetScope`)?
   - **Perimetro di Rete:** C'è controllo contro loopback, bridge Docker, reti link-local e DNS rebinding?
   - **Sanificazione Log:** Le credenziali, password SSH o token vengono mascherati con `[REDACTED]`?

3. **Checklist di Affidabilità & Concorrenza:**
   - **Goroutine & Concurrency:** Tutti i canali bufferizzati hanno timeout/defer cancel? I worker pool gestiscono il contesto di cancellazione (`ctx.Done()`)?
   - **Memory Leaks:** Ci sono listener WebSocket/SSE o timer non chiusi? I response body HTTP sono chiusi con `defer resp.Body.Close()`?
   - **Race Conditions:** Gli accessi a mappe o slice condivise sono protetti da `sync.RWMutex` o atomici?

4. **Checklist Frontend & UX:**
   - Ci sono warning TypeScript o `any` ingiustificati?
   - I rendering condizionali gestiscono stati di caricamento (`loading`), vuoto (`empty`) ed errore (`error`)?
   - Le modali o drawer hanno trap del focus e gestione chiusura con tasto ESC?

5. **Verifica della Suite e dei Test:**
   - Le asserzioni dei test sono significative o si limitano a `assert.NoError` senza verificare i dati restituiti?
   - Ci sono test `skip`, commentati o con tolleranze allargate?

6. **Doc Sync Verification:**
   - Ogni file nel diff ha una riga corrispondente nella Mappa Documenti Full-Stack di `docs/2_ARCHITECTURE.md`?
   - I documenti di dominio sono stati aggiornati con la data e l'hash recenti?

7. **Esito:**
   - Se emergono problemi: genera un report sintetico con livello di criticità (CRITICO / MEDIO / BASSO), crea i task correttivi o inserisci nel `Tech Debt Backlog` di `docs/3_TASK_LIST.md`.
   - Se pulito: conferma al Product Owner che il diff è pronto per il merge o per l'accettazione.

