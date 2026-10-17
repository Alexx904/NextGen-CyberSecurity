# Backend Setup

> Responsabile: Lead Developer. Comandi derivati da manifest, CI, Compose e configurazione; non eseguiti durante questa migrazione Markdown.

Go 1.22 come baseline CI; PostgreSQL 16; Nmap e capability di rete sono forniti dal container. Configurazione in `backend/internal/config/config.go`; migrazioni incorporate dal package `backend/internal/repository/migrations` e applicate all'avvio dal migrator PostgreSQL.

## Avvio locale

Dalla root: `docker compose up --build`. Per sviluppo nativo, predisporre prima un PostgreSQL di sviluppo, Nmap e le variabili nel processo, poi dalla directory `backend`:

```powershell
go mod download
go run ./cmd/server
```

Il backend legge variabili di ambiente; non presumere che carichi automaticamente un file `.env`. Riferimento sicuro per i nomi: `.env.example`. Variabili principali: `PORT`, `APP_ENV`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, `JWT_SECRET`, `SSH_MASTER_KEY`. Non inserire valori reali nei documenti.

## Verifica

Ultimo allineamento: 2026-10-17 · working tree. Da `backend`: test mirati durante lo sviluppo, una verifica finale con `go test ./...`, `go build ./...`, `go vet ./...` per il lato backend. Il percorso rapido esclude PostgreSQL tramite build tag: per repository/DB aggiungere `go test -tags=integration -count=1 ./internal/repository/postgres` con `TEST_DB_*` esplicite e database locale dedicato; nessun fallback su `DB_*` e nessuno skip. Dalla root il runner `scripts/verify.ps1 -Profile Final -Scope Backend -Integration` registra durate/esiti. Parametri, profili e criteri di riuso nel [Test Plan](../05-qa/TEST_PLAN.md).

## Migrazioni

Coppie numerate `.up.sql`/`.down.sql`; contratti e schema prima del codice. Applicazione e rollback solo su DB di test identificato. Il migrator esistente contiene riallineamento automatico della versione: non interpretarlo come rollback dei dati; richiede analisi prima di un rilascio su schema divergente.

Vedi [deployment](../06-devops/DEPLOYMENT.md) e [schema](../03-architecture/LLD/DATABASE_SCHEMA.md).
