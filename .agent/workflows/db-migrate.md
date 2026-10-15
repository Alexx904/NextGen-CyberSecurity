---
description: Procedura contrattuale per la creazione e la verifica di migrazioni del database (up/down)
---

Segui le regole di `AGENTS.md`. Questo workflow garantisce l'integrità del database, la tracciabilità delle migrazioni SQL e la garanzia di rollback.

1. **Intake & Specifiche del Modello Dati:**
   - Identifica le entità da aggiungere o modificare.
   - Verifica la conformità con la Clean Architecture: entità di dominio in `backend/internal/domain/`, DTO in `backend/internal/delivery/dto/`.
   - Se la modifica è distruttiva (DROP COLUMN, DROP TABLE), fermati ed esponi l'impatto al Product Owner.

2. **Numerazione Sequenziale della Migrazione:**
   - Ispeziona `backend/internal/repository/migrations/` per individuare l'ultimo numero sequenziale (es. `000007_...`).
   - Assegna il numero successivo (es. `000008_<nome_esplicativo>.up.sql` e `000008_<nome_esplicativo>.down.sql`).

3. **Scrittura File SQL:**
   - **`*.up.sql`:**
     - Usa `CREATE TABLE IF NOT EXISTS`, `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`.
     - Includi sempre chiavi primarie UUID o BIGSERIAL, foreign keys con vincoli di integrità (`ON DELETE CASCADE` o `ON DELETE SET NULL`) e indici B-Tree per colonne interrogate di frequente o chiavi esterne.
     - Aggiungi colonne di audit: `created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()`, `updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()`.
   - **`*.down.sql`:**
     - Ripristina esattamente lo stato precedente (`DROP TABLE IF EXISTS ... CASCADE;` o `ALTER TABLE ... DROP COLUMN IF EXISTS ...;`).
     - Assicurati che il rollback sia completamente idempotente e pulito.

4. **Allineamento Modelli Go:**
   - Aggiorna o crea le struct in `backend/internal/domain/`.
   - Aggiungi i tag JSON, DB (`db:"..."`) e i metodi di validazione.
   - Aggiorna il repository PostgreSQL in `backend/internal/repository/postgres/`.

5. **Allineamento Tipi Frontend:**
   - Aggiorna i tipi corrispondenti in `frontend/src/types/` per preservare la coerenza end-to-end (Contract-First).

6. **Verifica Zero-Assumption (Rollback & Forward Test):**
   - Esegui i test del repository Go: `go test -v ./internal/repository/postgres/...`.
   - Verifica che l'applicazione e il rollback della migrazione funzionino senza errori su DB di test.

7. **Doc Sync (ERD & Mappa):**
   - Aggiorna il diagramma ERD in `docs/2_ARCHITECTURE.md` (Sez. 4).
   - Aggiorna il file di dominio corrispondente in `docs/domains/`.
   - Registra l'avvenuta migrazione nel log di `docs/PROJECT_STATE.md`.

