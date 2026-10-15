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

Da `backend`: `go test ./...`, `go build ./...`, `go vet ./...`. Per test mirati usare il package reale. I servizi richiesti e le eventuali esclusioni interne devono essere riportati nell'evidenza.

## Migrazioni

Coppie numerate `.up.sql`/`.down.sql`; contratti e schema prima del codice. Applicazione e rollback solo su DB di test identificato. Il migrator esistente contiene riallineamento automatico della versione: non interpretarlo come rollback dei dati; richiede analisi prima di un rilascio su schema divergente.

Vedi [deployment](../06-devops/DEPLOYMENT.md) e [schema](../03-architecture/LLD/DATABASE_SCHEMA.md).
