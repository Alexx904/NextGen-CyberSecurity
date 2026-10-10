---
description: Pipeline autonoma end-to-end per una nuova feature, dall'idea al pacchetto di accettazione
---

Segui le regole di `AGENTS.md`. Input: la descrizione della feature fornita dall'utente.

1. **Contesto.** Leggi `docs/PROJECT_STATE.md` e `docs/3_TASK_LIST.md`. Se mancano, esegui prima `/onboard` o `/init`. Leggi solo i documenti di dominio rilevanti.

2. **Intake.** Traduci la richiesta in: obiettivo, attori, use case, edge case, criteri di accettazione verificabili. Controlla che non contraddica requisiti o architettura esistenti e non duplichi feature già presenti. Aggiorna `docs/0_REQUIREMENTS.md`.

3. **Chiarimenti.** Chiedi all'utente SOLO se il dubbio cambia il comportamento visibile o è bloccante. Max 3 domande in un unico messaggio. Altrimenti procedi con assunzioni ragionevoli e reversibili, annotandole in `PROJECT_STATE.md`.

4. **Design e task.** Aggiorna `docs/2_ARCHITECTURE.md` (e ADR) se servono nuovi componenti, tabelle o contratti API. Scomponi in task piccoli con ID, file, dipendenze, test, DoD in `docs/3_TASK_LIST.md`. Un task deve essere completabile e verificabile in modo indipendente.

5. **Branch.** Crea `feat/<id>` dal branch principale. Verifica prima che il working tree sia pulito.

6. **Per ogni task, in ordine di dipendenza:**
   a. **Plan:** Implementation Plan con file, firme, logica, rischi, Test Strategy (happy/error/edge, mock), comandi di verifica e domini/documenti da aggiornare. Se rispetta l'architettura, auto-approva e prosegui.
   b. **Execute:** scrivi codice e test. Non toccare file fuori dal piano senza dichiararlo.
   c. **Verify:** esegui i test mirati. Applica il failure loop di `AGENTS.md` (max 3 tentativi per causa).
   d. **Regression:** esegui suite completa e build/typecheck. Confronta con la baseline in `PROJECT_STATE.md`.
   e. **Doc Sync (bloccante):** esegui il gate di `AGENTS.md` sez. 5. Elenca i file cambiati con `git diff --name-only`, trova i domini nella Mappa documenti di `docs/2_ARCHITECTURE.md`, aggiorna i file in `docs/domains/` e gli altri documenti toccati, aggiorna "Ultimo allineamento". Senza questo passaggio il task non può passare a `[~]`.
   f. **Self-review:** rileggi il diff come reviewer ostile (scope creep, segreti, edge case, test significativi, docs coerenti con il codice).
   g. **Commit** sul branch con messaggio Conventional Commits. Nessun push.

7. **Per UI o flussi web:** verifica con il browser subagent e salva screenshot o registrazione come evidenza.

8. **Chiusura.** Imposta i task in `[~]`, aggiorna `PROJECT_STATE.md` (sezione "In attesa di accettazione", log sessione) e consegna il **Pacchetto di Accettazione** come da `AGENTS.md` sez. 8, inclusa la riga "Documenti aggiornati".

Se resti bloccato, imposta il task `[!]`, documenta il motivo in `PROJECT_STATE.md` e riferisci all'utente cosa serve.