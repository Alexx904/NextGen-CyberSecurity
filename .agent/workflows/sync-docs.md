---
description: Rileva e corregge il drift tra codice, contratti, task e documentazione
---

# Sync Docs

1. Determina lo scope reale dal diff e separa modifiche preesistenti estranee.
2. Mappa ogni file applicativo tramite la Mappa Full-Stack; segnala glob mancanti e documenti di dominio inesistenti.
3. Confronta codice e test con requisiti, API/DTO, schema/ERD, flussi e regole di business. Il codice non prevale automaticamente: segnala i conflitti di prodotto.
4. Mostra prima un report di divergenza; se la richiesta include l'allineamento, aggiorna solo le fonti pertinenti.
5. Usa link relativi, aggiorna date/hash senza inventarli e controlla riferimenti a file rimossi.
6. Verifica con `rg`, controllo link locale e `git diff --check`; riporta documenti aggiornati e divergenze residue.
