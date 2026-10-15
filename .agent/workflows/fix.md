---
description: Riproduzione, causa radice e correzione verificata di un bug
---

# Fix

1. Identifica comportamento atteso e osservato da requisiti, test, log non sensibili e feedback del PO.
2. Registra un task bug `[ ]`; esegui la baseline mirata prima di modificare codice.
3. Aggiungi il test minimo che riproduce il difetto e dimostra il fallimento. Se impraticabile, documenta il limite.
4. Trova la causa radice; correggi il punto di origine con la modifica minima, senza mascherare sintomi o indebolire test.
5. Esegui test di riproduzione, test del dominio e regressione completa. Applica il limite di tre tentativi per la stessa causa.
6. Esegui Doc Sync e hostile review, aggiorna il task a `[~]` solo con tutti i gate obbligatori verdi.
