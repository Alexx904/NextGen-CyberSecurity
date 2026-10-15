---
description: Analizza un progetto esistente, riorganizza la documentazione senza perdere nulla e crea la struttura di gestione autonoma
---

Segui le regole di `AGENTS.md`. Obiettivo: portare un progetto già avviato nella struttura autonoma **senza modificare il codice applicativo e senza cancellare nulla**.

## 0. Sicurezza
- Verifica il working tree e considera ogni modifica preesistente proprietà del PO. Se puoi lavorare senza sovrapposizioni, procedi e dichiarale; chiedi solo se i file necessari sono in conflitto.
- Crea `chore/agent-onboarding` solo se il tree è pulito e il PO ha autorizzato la gestione del branch.
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
Ricava feature e use case da rotte, schermate, comandi, test e documentazione esistente. Scrivili in `docs/01-product/REQUIREMENTS.md` marcandoli come "dedotti, da validare". Non inventare funzionalità.

## 5. Migrazione della documentazione
- Per ogni documento esistente, decidi la destinazione (requisiti, contesto, architettura, dominio, ADR, changelog) e **integra il contenuto utile** nei nuovi file, segnalando conflitti tra documenti e codice (analizzare il conflitto senza prevalenza automatica del codice).
- Conserva la provenienza dei contenuti migrati nei documenti canonici e nella cronologia Git. Non creare archivi, manuali, screenshot o stub dentro `docs/`; ogni rimozione materiale deve rientrare nella richiesta del PO.
- Il README resta al suo posto: aggiungi solo un link alla cartella `docs/`.

## 6. Generazione struttura
Crea o aggiorna: `docs/01-product/PRD.md`, `docs/03-architecture/HLD.md` (con `docs/domains/` se grande), `docs/04-development/TASK_LIST.md`, `CHANGELOG.md`, `docs/PROJECT_STATE.md` (con Project Profile compilato e baseline). Verifica che `AGENTS.md` e `.agent/workflows/` siano presenti.

Nella task list inserisci un backlog iniziale prioritizzato (Alta/Media/Bassa), tutto in stato `[ ]`:
- test falliti nella baseline;
- moduli critici senza test (proponi test di caratterizzazione, che fotografano il comportamento attuale prima di qualunque refactoring);
- debito tecnico evidente, dipendenze obsolete o vulnerabili, segreti esposti nel repo;
- TODO/FIXME rilevanti.

## 7. Chiusura
- Proponi un commit Conventional Commits; eseguilo solo su richiesta del PO. Nessun push né merge.
- Riporta all'utente, in linguaggio non tecnico: stack rilevato, stato di salute (build, test, baseline), cosa hai riorganizzato, principali rischi trovati.
- Poni **max 3 domande** raggruppate, solo su ciò che è ⚠️ o ❓ e ha impatto reale.
- Segna i documenti generati come "bozza da validare". Il Tech Lead valida i contenuti tecnici; le ambiguità di business e l'accettazione funzionale restano al PO.

Usa l'alberatura canonica in `docs/README.md` (`01-product`–`06-devops`), `PROJECT_STATE.md`, `docs/domains/`, `CHANGELOG.md` in root e i ruoli in `.agent/roles/`. I vecchi nomi numerati sono stati eliminati: non ricrearli.
