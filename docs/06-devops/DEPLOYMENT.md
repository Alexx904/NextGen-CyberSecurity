# Deployment

> Responsabile: DevOps Engineer. Procedura locale derivata da Compose; staging e produzione non risultano definiti nei file ispezionati.

## Ambiente locale verificabile

Dalla root, con Docker e Docker Compose disponibili:

```powershell
docker compose up --build
docker compose ps
```

UI: `http://localhost:3000`; API: `http://localhost:8080/api/v1`; health: `http://localhost:8080/api/v1/health`. Porte configurabili tramite ambiente. PostgreSQL usa volume persistente `postgres_data`; il backend applica migrazioni all'avvio. Compose assegna `NET_RAW` al backend per Nmap.

La presenza di healthcheck non dimostra test end-to-end. Questa migrazione non ha avviato container né eseguito scansioni.

## Staging/produzione: gate prima del rilascio

Definire destinazione, TLS/reverse proxy, gestione segreti, backup/ripristino provati, compatibilità schema/binario, autorizzazioni, monitoraggio e rollback. I default di sviluppo nei manifest non sono configurazione di produzione. Conservare stabile la chiave di cifratura SSH tramite gestione segreti; non pubblicarla.

Non usare `docker compose down -v` come normale stop: rimuove dati persistenti. Ogni deploy o azione su sistemi remoti richiede scope e autorizzazione del PO.

Un rollback del binario non ripristina automaticamente lo schema. Il piano di rilascio deve distinguere ripristino del codice, migrazione down e recupero dati.
