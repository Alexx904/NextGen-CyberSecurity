# Deployment

## COMPANY-UX-02 · Upgrade/rollback · 2026-10-17

Backend e frontend aggiornati insieme. Migrazioni 000008 (etichette clienti SSH) e 000009 (guard gerarchia) applicate prima dell'uso della nuova API; avvio server esegue il migrator esistente. Nessun deploy applicativo eseguito in questa sessione. `docker compose up --build` è un'istruzione per la prova del PO, non un'azione già effettuata.

Down 000009 elimina trigger/funzione senza riscrivere gruppi; down 000008 rimuove client_name e perde le sole etichette: esportarle prima del rollback operativo. Up/down e idempotenza provati nel solo DB dedicato `nextgen_company_ux02_test_20261017`, mantenuto per ripetibilità. Nessuna modifica a schema/dati del DB applicativo; lettura sola per verificare assenza sotto-reparti esistenti.

## E28-S1-T1 · Limiti operativi (2026-10-17, working tree)

Remediation IPv4 su host Linux con iptables e privilegi sudo già richiesti dal
gestore firewall. Nessuna migrazione o dipendenza nuova. Snapshot e watchdog remoto
sono riusati; serializzazione e timer locali presuppongono una singola istanza backend.
Conferma/rollback remoto e interruzione connessioni devono essere provati su laboratorio
prima dell'uso operativo. Il client SSH attuale non verifica la host key; nessuna
attestazione di canale autenticato al server finché TECH-DEBT-06 rimane aperto.
Non è stato eseguito alcun deploy né comando SSH reale durante questo task.

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

## SSH-HARDEN-01 · Nota operativa · 2026-10-17 · working tree

La migrazione 000010 aggiunge `ssh_servers.host_key_fingerprint` vuoto per
l'inventario esistente. Prima delle connessioni occorre configurare il pin SHA256
della chiave del server tramite un canale fidato. Non copiare il fingerprint
della chiave client e non accettare automaticamente una chiave ricevuta in rete.
Rotazioni richiedono verifica indipendente e aggiornamento esplicito. Down perde
i pin host; salvare i pin autorizzati prima di un rollback operativo e riconfigurarli
dopo il re-upgrade. Nessuna migrazione o deploy applicativo eseguito in questa
sessione. Timeout locali non sostituiscono watchdog/transazione Linux remota;
E28-S1-T1 e TECH-DEBT-07 restano aperti.
