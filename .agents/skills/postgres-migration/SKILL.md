---
name: postgres-migration
description: Design or change PostgreSQL schema and versioned migrations in this repository, including compatibility and rollback verification.
---

# PostgreSQL migration

Read the affected [schema](../../../docs/03-architecture/LLD/DATABASE_SCHEMA.md), domain and existing migrator before changing them. Preserve key types, constraints and real package paths. Define expected schema, compatibility, Go models and affected TypeScript contracts before implementation; do not invent DTO directories.

Inspect `backend/internal/repository/migrations/` and choose the next unused sequential number for paired `.up.sql`/`.down.sql` files. Do not rewrite a published migration. Explain data loss and obtain authorization for destructive operations. A down restores schema, not deleted data; define backup/restore when required.

Use `IF NOT EXISTS` only when it cannot conceal divergent schema. Choose ON DELETE behavior and indexes explicitly; do not default to CASCADE. Specify update semantics for audit columns: DEFAULT does not maintain `updated_at`. Align repository persistence and public contracts only where affected.

Verify against an identified dedicated local test DB with explicit `TEST_DB_*`. From `backend`, run:

```text
go test -tags=integration -count=1 ./internal/repository/postgres
```

Prove forward/schema assertions and down/up round-trip for the changed migration. A fast Go test run does not cover tagged DB tests; missing configuration/DB is FAIL. Do not invoke rollback on application data or a DB whose purpose is unknown.

Complete the final affected-side gate using [verification guidance](../repo-verification/SKILL.md). Update schema/ERD, affected domain and evidence links. Report compatibility, data-loss limits and unexecuted checks; never infer production safety from a test rollback.
