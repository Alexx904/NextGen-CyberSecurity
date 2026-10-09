---
description: Gestisce l'esito dell'utente su una feature consegnata, tra accettazione (OK), rifiuto (KO) o richiesta di modifica
---

Segui le regole di `AGENTS.md`. Input: il messaggio dell'utente sull'esito di una feature.

1. **Identifica la feature.** Leggi "In attesa di accettazione" in `docs/PROJECT_STATE.md`. Se il messaggio non indica chiaramente a quale feature si riferisce e ce n'è più di una, elenca le feature in attesa e chiedi quale.

2. **Classifica l'esito.**

   **OK (funziona):**
   - Imposta i task a `[x]` in `docs/3_TASK_LIST.md`.
   - Se i task completati superano 10, archivia in `docs/4_CHANGELOG.md`.
   - Rimuovi la feature da "In attesa di accettazione" e aggiorna i documenti interessati.
   - Conferma le assunzioni fatte (rimuovile da "Assunzioni da confermare") oppure chiedi conferma se erano rilevanti.
   - Proponi l'eventuale merge/commit (`/commit`), senza eseguire push né merge.

   **KO (non funziona):**
   - Raccogli solo i dettagli essenziali mancanti (cosa ha fatto, cosa ha visto, cosa si aspettava). Se hai già abbastanza, non fare domande.
   - Riporta il task in `[ ]` e registra il bug in `docs/3_TASK_LIST.md`.
   - Riproduci il problema con un test che fallisce, individua la causa radice, correggi, riesegui test mirati e suite completa (failure loop, max 3 tentativi).
   - Verifica che il comportamento rifiutato non dipenda da un requisito mal interpretato: se sì, correggi anche `0_REQUIREMENTS.md`.
   - Consegna un nuovo Pacchetto di Accettazione, con il task in `[~]`.

   **Modifica (cambia/aggiungi/togli):**
   - Se la modifica è piccola e locale: trattala come mini-feature (intake breve, piano, test, verifica).
   - Se modifica requisiti o architettura già accettati: esegui l'analisi d'impatto sui task `[x]` e sui test esistenti, aggiorna docs e task list, poi segui la pipeline di `/feature`.
   - Segnala all'utente cosa cambia di ciò che aveva già accettato.

   **Indicazione generica (preferenza, vincolo, priorità):**
   - Registrala in `PROJECT_STATE.md` o nel documento pertinente (ADR se è una decisione tecnica) e applicala da ora in poi.

3. **Chiudi il turno** con "Stato" e "Serve da te", aggiornando `PROJECT_STATE.md`.
