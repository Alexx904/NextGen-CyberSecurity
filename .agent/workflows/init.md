---
description: Inizializza un nuovo progetto con discovery, struttura documentale e baseline Codex
---

# Init

1. Raccogli idea, utenti, risultato principale, vincoli e non-goal. Chiedi solo decisioni bloccanti, massimo 3 per turno.
2. Definisci MVP, criteri di accettazione, rischi e stack; chiedi conferma solo per scelte costose, irreversibili o visibili.
3. Crea la struttura minima prevista da `AGENTS.md`: sezioni `01-product`–`06-devops`, `PROJECT_STATE.md`, domini, changelog in root e ruoli/playbook.
4. Compila il Project Profile con comandi reali derivati dai manifest, non inventati.
5. Crea task con DoD verificabile e un primo vertical slice; evita scaffolding non richiesto.
6. Esegui build/test/lint disponibili, registra la baseline e consegna lo stato. Non inizializzare servizi remoti, push o deploy senza richiesta.

Usa l'alberatura canonica in `docs/README.md` (`01-product`–`06-devops`), `PROJECT_STATE.md`, `docs/domains/`, `CHANGELOG.md` in root e i ruoli in `.agent/roles/`. I vecchi nomi numerati sono stati eliminati: non ricrearli. In `docs/` conservare solo Markdown necessari agli agenti; nessun manuale utente, screenshot o archivio.
