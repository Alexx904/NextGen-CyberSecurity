---
description: Refactoring behavior-preserving con caratterizzazione, metriche e regressione
---

# Refactor

1. Definisci odore, perimetro, comportamento invariato e beneficio misurabile. Se cambia il comportamento, usa `feature.md`.
2. Esegui test esistenti; aggiungi test di caratterizzazione per percorsi non protetti.
3. Registra il task e collega l'eventuale voce Tech Debt. Pianifica passi piccoli e reversibili.
4. Refactor senza mescolare feature; verifica dopo ogni passo significativo.
5. Esegui test mirati per passo e una sola suite finale completa per lato coinvolto con build/typecheck/lint, secondo `AGENTS.md` sez. 8; poi Doc Sync e hostile review. Non ripetere controlli non invalidati.
6. Chiudi o aggiorna la voce Tech Debt e consegna differenze strutturali, invarianti ed evidenze.
