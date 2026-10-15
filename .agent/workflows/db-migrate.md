---
description: Procedura contrattuale per la creazione e la verifica di migrazioni del database (up/down)
---

Segui le regole di `AGENTS.md`. Questo workflow definisce i controlli di integrità, tracciabilità e rollback delle migrazioni SQL.

1. **Intake & Specifiche del Modello Dati:**
   - Identifica le entità da aggiungere o modificare.
   - Verifica la conformità con la Clean Architecture: entità in `backend/internal/domain/`, handler e DTO nei package effettivamente esistenti sotto `backend/internal/delivery/http/`. Non creare directory ipotetiche per rispettare un esempio.
   - Se la modifica è distruttiva (DROP COLUMN, DROP TABLE), fermati ed esponi l'impatto al Product Owner.

2. **Numerazione Sequenziale della Migrazione:**
   - Ispeziona `backend/internal/repository/migrations/` per individuare l'ultimo numero sequenziale (es. `000007_...`).
   - Assegna il numero successivo (es. `000008_<nome_esplicativo>.up.sql` e `000008_<nome_esplicativo>.down.sql`).

3. **Scrittura File SQL:**
   - **`*.up.sql`:**
     - Definisci precondizioni e schema finale atteso. Usa `IF NOT EXISTS` solo se non nasconde uno schema divergente; una migrazione versionata non deve fingere successo.
     - Mantieni tipi di chiave e convenzioni esistenti; scegli esplicitamente vincoli, comportamento ON DELETE e indici in base al dominio, senza imporre CASCADE o nuove chiavi a ogni modifica.
     - Aggiungi colonne di audit quando previste dal modello; specifica come viene aggiornato `updated_at`, perché il solo DEFAULT non gestisce gli update.
   - **`*.down.sql`:**
     - Ripristina esattamente lo stato precedente. L'uso di `CASCADE` richiede analisi delle dipendenze e test dedicato: non applicarlo come default.
     - Definisci precondizioni e perdita di dati eventuale; verifica il ripristino dello schema. Un down non garantisce il recupero dei dati cancellati: serve un piano di backup/restore quando pertinente.

4. **Allineamento Modelli Go:**
   - Aggiorna o crea le struct in `backend/internal/domain/`.
   - Aggiungi i tag JSON, DB (`db:"..."`) e i metodi di validazione.
   - Aggiorna il repository PostgreSQL in `backend/internal/repository/postgres/`.

5. **Allineamento Tipi Frontend:**
   - Aggiorna i tipi corrispondenti in `frontend/src/types/` per preservare la coerenza end-to-end (Contract-First).

6. **Verifica Zero-Assumption (Rollback & Forward Test):**
   - Esegui i test del repository Go: `go test -v ./internal/repository/postgres/...`.
   - Verifica applicazione, schema risultante e rollback su un DB dichiaratamente di test; mai su database di produzione o non identificato.

7. **Doc Sync (ERD & Mappa):**
   - Aggiorna il diagramma ERD in `docs/03-architecture/LLD/DATABASE_SCHEMA.md`.
   - Aggiorna il file di dominio corrispondente in `docs/domains/`.
   - Registra l'avvenuta migrazione nel log di `docs/PROJECT_STATE.md`.
