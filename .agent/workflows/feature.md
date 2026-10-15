---
description: Pipeline autonoma end-to-end per una nuova feature, dall'idea al pacchetto di accettazione
---

Segui `AGENTS.md` e l'indice `.agent/README.md`. Input: la descrizione della feature fornita dal PO.

1. **Contesto & Layering.** Leggi `docs/PROJECT_STATE.md` e solo lo Sprint attivo in `docs/04-development/TASK_LIST.md` (Livello 0). Se mancano, applica `onboard.md` o `init.md`. Leggi solo il documento di dominio rilevante tramite la Mappa Full-Stack in `docs/03-architecture/DOMAIN_MAP.md` (Livello 1). Non caricare guide utente o file globali per intero (Livelli 2 e 3).

2. **Orchestrazione.** Applica `orchestration.md`, seleziona ruoli e gate pertinenti; QA e DevOps contribuiscono al design. Contratti API/schema/componenti in `docs/03-architecture/LLD/` prima dell'implementazione. Per feature UI aggiorna `docs/02-design/`; i documenti già validi soddisfano il gate senza riscritture.

3. **Intake.** Traduci la richiesta in: obiettivo, attori, use case, edge case, criteri di accettazione verificabili. Controlla che non contraddica requisiti o architettura esistenti e non duplichi feature già presenti. Aggiorna requisiti, use case e storie pertinenti in `docs/01-product/`, senza duplicare definizioni.

4. **Chiarimenti.** Chiedi all'utente SOLO se il dubbio cambia il comportamento visibile o è bloccante. Max 3 domande in un unico messaggio. Altrimenti procedi con assunzioni ragionevoli e reversibili, annotandole in `PROJECT_STATE.md`.

5. **Design, Contratti & Task (Contract-First):**
   - Se tocca API o DB: definisci prima le interfacce TypeScript (`frontend/src/types/`) e le struct Go (`delivery/dto/` o `domain/`).
   - Se tocca il database: prepara la migrazione versionata `.up.sql` e `.down.sql`.
   - Aggiorna `docs/03-architecture/HLD.md` (e ADR se serve). Scomponi in task con DoD a 4 punti in `docs/04-development/TASK_LIST.md`.

6. **Isolamento Git.** Ispeziona branch e working tree. Crea `feat/<id>` solo se la feature è isolabile e non esistono modifiche preesistenti da spostare; non fare stash, reset o checkout distruttivi per “pulire” il repository.

7. **Per ogni task, in ordine di dipendenza:**
   a. **Plan:** Implementation Plan con file, firme, logica, rischi, Test Strategy (happy/error/edge, mock), comandi di verifica e domini/documenti da aggiornare. Se rispetta l'architettura, auto-approva e prosegui.
   b. **Execute:** scrivi codice e test. Non toccare file fuori dal piano senza dichiararlo.
   c. **Verify:** esegui i test mirati. Se ci sono migrazioni SQL, verifica sia `up` che `down`. Applica il failure loop di `AGENTS.md` (max 3 tentativi per causa).
   d. **Regression:** esegui suite completa e build/typecheck. Confronta con la baseline in `PROJECT_STATE.md`.
   e. **Doc Sync Full-Stack (bloccante):** esegui il gate di `AGENTS.md` sez. 7. Elenca i file cambiati con `git diff --name-only`, trova i domini nella Mappa documenti Full-Stack, aggiorna i file in `docs/domains/` e gli altri documenti toccati, aggiorna "Ultimo allineamento". Senza questo passaggio il task non può passare a `[~]`.
   f. **Hostile Review:** rileggi il diff come reviewer avversario ostile. Usa un subagent QA solo se disponibile, consentito dal runtime e utile su uno scope separabile; cerca injection, leak di risorse/goroutine, race condition, credenziali esposte, permessi non validati o test deboli.
   g. **Commit:** esegui `commit.md` solo se richiesto dal PO. Nessun push o merge implicito.

8. **Per UI o flussi web:** usa le capacità browser/computer-use disponibili e salva screenshot reali in `.agent/artifacts/<task-id>/`, fuori da `docs/` e ignorati da Git. Se il runtime non le offre, dichiara la verifica visuale non eseguita e non sostituirla con un'affermazione presunta.

9. **Chiusura.** Imposta i task in `[~]`, aggiorna `PROJECT_STATE.md` (sezione "In attesa di accettazione", log sessione, eventuale debito tecnico registrato) e consegna il **Pacchetto di Accettazione** come da `AGENTS.md` sez. 10.

Se resti bloccato, imposta il task `[!]`, documenta il motivo in `PROJECT_STATE.md` e riferisci all'utente cosa serve.
