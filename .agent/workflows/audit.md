---
description: Audit periodico di sicurezza, qualità, drift documentale e debito tecnico
---

# Audit

1. Definisci commit/base e scope; l'audit è read-only salvo richiesta esplicita di correzione.
2. Controlla secret exposure senza aprire file sensibili, dipendenze/configurazioni, injection, auth/tenant scope, concorrenza, lifecycle risorse ed error handling.
3. Verifica test mancanti o deboli, `skip`, baseline, CI, migrazioni up/down e copertura della Mappa Full-Stack.
4. Confronta requisiti, architettura, domini, task e codice per drift e riferimenti orfani.
5. Riporta findings ordinati per severità con file/riga, evidenza, impatto e rimedio. Nessun finding significa “nessun problema rilevato nello scope”, non garanzia assoluta.
6. Registra finding non corretti nel Tech Debt; modifica codice solo se il PO ha chiesto anche il fix.
