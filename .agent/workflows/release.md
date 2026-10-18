---
description: Preparazione al rilascio di versione, aggiornamento changelog, semver bump e audit finale
---

Segui le regole di `AGENTS.md`. Questo workflow prepara un rilascio formale di Milestone o Minor version della piattaforma.

1. **Verifica Pre-Rilascio (Pre-flight Checks):**
   - Verifica branch, working tree e scostamento da `main`; non eseguire fetch/pull o merge impliciti.
   - Verifica che nessun task nella lista attiva di `docs/04-development/TASK_LIST.md` sia in stato `[!]` (bloccato) o `[~]` (in attesa di accettazione).
   - Verifica lo stato del debito tecnico in `docs/04-development/TASK_LIST.md`: nessun debito con priorità "Critica" deve essere aperto.

2. **Suite Completa di Baseline (Zero-Assumption):**
   - Applica [repo-verification](../../.agents/skills/repo-verification/SKILL.md) con profilo Full: entrambi i lati, build/vet e PostgreSQL dedicato. Rollback, browser/PDF e budget bundle pertinenti restano prove aggiuntive; non duplicare comandi già validi.
   - Registra una sintesi della baseline corrente in PROJECT_STATE e collega il report dettagliato; dati/servizi mutati invalidano le evidenze di integrazione.

3. **Versioning Semantico (SemVer):**
   - Determina il nuovo tag di versione seguendo SemVer (`vX.Y.Z`):
     - `MAJOR` (X): rotture architetturali, cambi radicali di protocollo/DB.
     - `MINOR` (Y): nuova Milestone, nuove feature o domini introdotti.
     - `PATCH` (Z): bugfix, rifiniture, hardening.
   - Aggiorna il numero di versione nei manifest applicativi solo dopo conferma della versione da parte del PO.

4. **Archiviazione Changelog & Task:**
   - Sposta i task accettati `[x]` dallo Sprint attivo in `docs/04-development/TASK_LIST.md` al file di archivio storico `CHANGELOG.md` sotto l'intestazione della nuova versione `## [vX.Y.Z] - YYYY-MM-DD`.
   - Includi una sintesi discorsiva ad alto livello dei benefici per gli stakeholder e la lista dei contributi principali.

5. **Allineamento Stato di Progetto:**
   - Aggiorna `docs/PROJECT_STATE.md`:
     - Aggiorna "Fase" con la nuova versione rilasciata e la prossima milestone.
     - Pulisci la lista "Accettati recentemente" spostandola nello storico.
     - Inserisci una riga di log sintetica e collega le evidenze del rilascio.

6. **Proposta di Tag e Merge al Product Owner:**
   - Prepara il messaggio di commit Conventional Commits: `chore(release): release vX.Y.Z`.
   - Proponi commit e tag. Creazione del tag, merge, push e deploy richiedono ciascuno richiesta esplicita del PO.
