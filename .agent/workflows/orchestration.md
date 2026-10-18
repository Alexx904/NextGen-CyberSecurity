---
description: Orchestra ruoli, dipendenze e handoff proporzionati per singola feature
---

# Orchestration

Il Tech Lead mantiene scope, dipendenze, task e responsabilità finale. Leggere solo i [ruoli pertinenti](../README.md#ruoli-e-ownership) che aggiungono vincoli utili; il resto è coperto da AGENTS e skill. I ruoli possono essere svolti dallo stesso agente o da subagent disponibili e autorizzati. La definizione di un ruolo non avvia un processo.

## Gate per feature

| Gate | Condizione di ingresso/uscita | Validatore |
|---|---|---|
| Ready for design | Obiettivo, FR/UC, criteri osservabili e ambiguità rilevanti risolte | Tech Lead; PO per business |
| Ready for implementation | Contratti pertinenti, piano e rischi definiti; test strategy pronta | Tech Lead con Architect/Lead Developer |
| Verified | Implementazione, test obbligatori, review e Doc Sync completati con evidenze | Tech Lead con QA |
| Accepted | Risultato funzionale accettato | PO |

I gate valgono per il cambiamento corrente: non attendere il completamento dell'intera documentazione o delle milestone future. Un documento esistente può soddisfare il gate se ancora valido. Una sezione non pertinente può essere N/A con motivazione. Una verifica obbligatoria non eseguita non può essere N/A per aggirare un blocco.

QA e DevOps contribuiscono già al design; dopo contratti sufficienti Backend e Frontend possono procedere in parallelo su file distinti. Non richiedere al PO approvazioni tecniche interne già comprese nello scope autorizzato.

## Dimensionamento

- Bug locale: comportamento atteso, riproduzione, fix/test e documenti pertinenti.
- Feature UI: requisiti/storia, UX/stati, componenti, piano test e prova visuale.
- Feature API/DB o cross-domain: contratti e LLD prima del codice, compatibilità, migrazioni up/down, piano e review.
- Sola documentazione: inventario, fonti, mapping, controlli di riferimenti e preservazione contenuti.

## Delega e handoff

Prima della delega assegnare task, input minimi, file scrivibili, non-goal, dipendenze e output atteso. Limitare i worker ai task utili e separabili; nessuna scrittura concorrente sullo stesso file. I cambi globali e l'integrazione restano al Tech Lead.

Handoff minimo:

```markdown
Task / ruolo:
Obiettivo e scope:
Input / FR / UC:
File modificati e contratti:
Decisioni / assunzioni:
Verifiche e risultati:
Questioni aperte / blocchi:
Prossimo ruolo e gate:
```

Non dichiarare una review indipendente se effettuata dallo stesso agente. Se un requisito cambia, riaprire solo i gate e i documenti impattati.
