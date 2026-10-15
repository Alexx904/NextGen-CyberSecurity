---
description: Pipeline autonoma end-to-end per una nuova feature, dall'idea al pacchetto di accettazione
---

Segui le regole di `AGENTS.md`. Input: la descrizione della feature fornita dall'utente.

1. **Contesto & Layering.** Leggi `docs/PROJECT_STATE.md` e solo lo Sprint attivo in `docs/3_TASK_LIST.md` (Livello 0). Se mancano, esegui prima `/onboard` o `/init`. Leggi solo il documento di dominio rilevante tramite la Mappa Full-Stack (Livello 1). Non caricare guide utente o file globali per intero (Livelli 2 e 3).

2. **Intake.** Traduci la richiesta in: obiettivo, attori, use case, edge case, criteri di accettazione verificabili. Controlla che non contraddica requisiti o architettura esistenti e non duplichi feature già presenti. Aggiorna `docs/0_REQUIREMENTS.md`.

3. **Chiarimenti.** Chiedi all'utente SOLO se il dubbio cambia il comportamento visibile o è bloccante. Max 3 domande in un unico messaggio. Altrimenti procedi con assunzioni ragionevoli e reversibili, annotandole in `PROJECT_STATE.md`.

4. **Design, Contratti & Task (Contract-First):**
   - Se tocca API o DB: definisci prima le interfacce TypeScript (`frontend/src/types/`) e le struct Go (`delivery/dto/` o `domain/`).
   - Se tocca il database: prepara la migrazione versionata `.up.sql` e `.down.sql`.
   - Aggiorna `docs/2_ARCHITECTURE.md` (e ADR se serve). Scomponi in task con DoD a 4 punti in `docs/3_TASK_LIST.md`.

5. **Branch.** Crea `feat/<id>` dal branch principale. Verifica prima che il working tree sia pulito.

6. **Per ogni task, in ordine di dipendenza:**
   a. **Plan:** Implementation Plan con file, firme, logica, rischi, Test Strategy (happy/error/edge, mock), comandi di verifica e domini/documenti da aggiornare. Se rispetta l'architettura, auto-approva e prosegui.
   b. **Execute:** scrivi codice e test. Non toccare file fuori dal piano senza dichiararlo.
   c. **Verify:** esegui i test mirati. Se ci sono migrazioni SQL, verifica sia `up` che `down`. Applica il failure loop di `AGENTS.md` (max 3 tentativi per causa).
   d. **Regression:** esegui suite completa e build/typecheck. Confronta con la baseline in `PROJECT_STATE.md`.
   e. **Doc Sync Full-Stack (bloccante):** esegui il gate di `AGENTS.md` sez. 5. Elenca i file cambiati con `git diff --name-only`, trova i domini nella Mappa documenti Full-Stack, aggiorna i file in `docs/domains/` e gli altri documenti toccati, aggiorna "Ultimo allineamento". Senza questo passaggio il task non può passare a `[~]`.
   f. **Hostile Review:** rileggi il diff come reviewer avversario ostile (o invoca subagent QA/research se la feature è estesa): cerca injection, leak di risorse/goroutine, race condition, credenziali esposte, permessi non validati o test deboli.
   g. **Commit** sul branch con messaggio Conventional Commits atomico. Nessun push.

7. **Per UI o flussi web:** verifica con il browser subagent e salva screenshot o registrazione come evidenza visuale.

8. **Chiusura.** Imposta i task in `[~]`, aggiorna `PROJECT_STATE.md` (sezione "In attesa di accettazione", log sessione, eventuale debito tecnico registrato) e consegna il **Pacchetto di Accettazione** come da `AGENTS.md` sez. 8.

Se resti bloccato, imposta il task `[!]`, documenta il motivo in `PROJECT_STATE.md` e riferisci all'utente cosa serve.