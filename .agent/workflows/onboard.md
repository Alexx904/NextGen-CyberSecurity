---
description: Analizza un progetto esistente, riorganizza la documentazione senza perdere nulla e crea la struttura di gestione autonoma
---

Segui le regole di `AGENTS.md`. Obiettivo: portare un progetto già avviato nella struttura autonoma **senza modificare il codice applicativo e senza cancellare nulla**.

## 0. Sicurezza
- Verifica che il working tree git sia pulito. Se non lo è, fermati e chiedi all'utente di committare o stashare.
- Crea il branch `chore/agent-onboarding`.
- Usa `git mv` per spostare file (preserva la cronologia). Nessuna cancellazione. Nessuna modifica al codice sorgente, ai test o alla configurazione di build.

## 1. Inventario
Mappa il repository: struttura delle cartelle, linguaggi, framework, manifest delle dipendenze, CI/CD, Docker/infra, configurazione test, migrazioni DB, variabili d'ambiente (solo i nomi, mai i valori). Elenca tutti i documenti esistenti (README, `docs/`, wiki locali, ADR, CONTRIBUTING, note sparse, TODO/FIXME nel codice).

## 2. Baseline
Individua i comandi di build, test e lint leggendo manifest, Makefile, CI e README. Eseguili davvero (installando le dipendenze solo se necessario e indicandolo). Registra in `docs/PROJECT_STATE.md` esito e **fallimenti preesistenti**: sono la baseline, non vanno corretti ora né nascosti. Se qualcosa non è eseguibile (servizi, credenziali), annota cosa serve.

## 3. Reverse engineering dell'architettura
Ricostruisci dal codice: moduli e layer, flusso dei dati, modello dati (ERD testuale), endpoint e contratti API, integrazioni esterne, meccanismi di auth, pattern ricorrenti, convenzioni. **Etichetta ogni affermazione:**
- ✅ verificato nel codice (indica il percorso del file)
- ⚠️ dedotto (ipotesi ragionevole)
- ❓ sconosciuto (da chiedere all'utente)

Mai presentare un'ipotesi come fatto.

## 4. Requisiti dedotti
Ricava feature e use case da rotte, schermate, comandi, test e documentazione esistente. Scrivili in `docs/0_REQUIREMENTS.md` marcandoli come "dedotti, da validare". Non inventare funzionalità.

## 5. Migrazione della documentazione
- Per ogni documento esistente, decidi la destinazione (requisiti, contesto, architettura, dominio, ADR, changelog) e **integra il contenuto utile** nei nuovi file, segnalando conflitti tra documenti e codice (vince il codice, il conflitto va registrato).
- Sposta gli originali in `docs/_legacy/` con `git mv` e crea `docs/_legacy/MAPPING.md` (vecchio file → dove è confluito).
- Il README resta al suo posto: aggiungi solo un link alla cartella `docs/`.

## 6. Generazione struttura
Crea o aggiorna: `docs/1_PROJECT_CONTEXT.md`, `docs/2_ARCHITECTURE.md` (con `docs/domains/` se grande), `docs/3_TASK_LIST.md`, `docs/4_CHANGELOG.md`, `docs/PROJECT_STATE.md` (con Project Profile compilato e baseline). Verifica che `AGENTS.md` e `.agent/workflows/` siano presenti.

Nella task list inserisci un backlog iniziale prioritizzato (Alta/Media/Bassa), tutto in stato `[ ]`:
- test falliti nella baseline;
- moduli critici senza test (proponi test di caratterizzazione, che fotografano il comportamento attuale prima di qualunque refactoring);
- debito tecnico evidente, dipendenze obsolete o vulnerabili, segreti esposti nel repo;
- TODO/FIXME rilevanti.

## 7. Chiusura
- Commit sul branch `chore/agent-onboarding` con Conventional Commits. Nessun push né merge.
- Riporta all'utente, in linguaggio non tecnico: stack rilevato, stato di salute (build, test, baseline), cosa hai riorganizzato, principali rischi trovati.
- Poni **max 5 domande** raggruppate, solo su ciò che è ⚠️ o ❓ e ha impatto reale.
- Segna i documenti generati come "bozza da validare". Diventano riferimento ufficiale solo dopo la conferma dell'utente. Fino ad allora, in caso di dubbio, fa fede il codice.
