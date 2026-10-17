---
description: Prepara un commit atomico Conventional Commits senza push
---

# Commit

1. Verifica che il PO abbia richiesto il commit. Ispeziona status e diff; separa modifiche preesistenti o non pertinenti.
2. Convalida i gate richiesti dal task e Doc Sync: riutilizza evidenze ancora valide secondo `AGENTS.md` sez. 8, riesegui solo i controlli invalidati. Non committare se test obbligatori falliscono o segreti/file sensibili sono presenti.
3. Proponi scope e messaggio Conventional Commits (`type(scope): summary`) coerenti con il diff.
4. Aggiungi allo staging solo file esplicitamente inclusi nel task; mostra lo staged diff e verifica assenza di contenuti estranei.
5. Crea il commit solo su richiesta, poi riporta hash e file. Mai push, merge, amend o riscrittura della history impliciti.
