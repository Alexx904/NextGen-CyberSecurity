# Implementation Plans

Creare `<task-id>-<short-name>.md` per feature complesse, cross-domain o con API/DB; per modifiche locali basta il task. I piani precedenti non vengono ricreati se già rimossi dal PO.

Ogni piano contiene: obiettivo/non-goal, FR/UC/storie, file e contratti interessati, dipendenze, scelte/ADR, strategia e comandi di test, migrazione/rollback se pertinenti, documenti da aggiornare e stato dei gate.

Stati tecnici: draft → ready → implemented → verified. `ready` è validato dal Tech Lead; l'accettazione funzionale del PO resta separata in TASK_LIST. Eventuali assunzioni visibili o cambi incompatibili richiedono decisione del PO.
