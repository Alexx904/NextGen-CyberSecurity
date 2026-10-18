# Deployment

> Publication scope 2026-10-18: alert-channel / outbox / migration 000012 sections describe deferred T4/T5 design and local prototype evidence. Their code and routes are excluded from this release; published migration is 000011. Alert-specific commands are future gates.


## DOCKER-CI-01 · Caricamento artifact

Ultimo allineamento: 2026-10-17 · working tree. Contratto operativo del pacchetto definito nella [pipeline CI](CI_CD_PIPELINE.md#docker-ci-01--pacchetto-scaricabile). Dopo una CI riuscita, scaricare ed estrarre l'artifact `nextgen-docker-linux-amd64-<sha>` dalla pagina Actions. Su Windows è richiesto Docker Desktop con container Linux, architettura amd64. Dalla cartella estratta:

```powershell
Get-FileHash .\nextgen-images.tar.gz -Algorithm SHA256
# Confrontare il valore con SHA256SUMS prima di caricare.
docker load --input nextgen-images.tar.gz
docker compose -f docker-compose.yml -f compose.images.yml up -d --no-build
docker compose -f docker-compose.yml -f compose.images.yml ps
```

UI su `http://localhost:3000`. PostgreSQL viene scaricato se assente; database e runtime non sono inclusi nell'artifact. La directory di estrazione determina il progetto Compose e quindi il volume predefinito: per aggiornare un'installazione esistente mantenere directory/nome progetto e configurazione originali. Non avviare un secondo stack sulle stesse porte. I manifest conservano i default di sviluppo e i requisiti di segreti/backup descritti sotto; non sono una configurazione di produzione. Nessun volume viene eliminato dalla procedura. La prima esecuzione Actions resta da verificare dopo push autorizzato.

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

SSH-TOOLS-01: la configurazione frontend inoltra `/ws/` con HTTP/1.1 e header
Upgrade/Connection; timeout proxy di lettura/scrittura 3600s. Vite ha `ws: true`
sullo stesso percorso. Richiesto rebuild frontend per aggiornare Nginx, oltre
al backend per la diagnostica firewall:
`docker compose up -d --build --no-deps backend frontend`. Comando di attivazione
per il PO, non eseguito durante il fix.

La lettura firewall richiede iptables installato e privilegi effettivi root
oppure sudo non interattivo autorizzato al comando di lettura. Il servizio usa
PATH con sbin e `sudo -n`, senza riusare password SSH per sudo. Una sessione SSH
riuscita non dimostra questi privilegi. Su Debian verificare nella sessione del
PO `command -v iptables` e `sudo -n iptables -L -n -v --line-numbers` (sola
lettura). Nessuna installazione o modifica sudoers automatica. Mutazioni e
watchdog now use the host-owned transaction described in the
[SSH domain](../domains/ssh_and_firewall.md): flock, stat, nohup, date, sha256sum,
awk, core file/sleep utilities and iptables/save/restore (including restore test).
Non-root users require noninteractive authorization for the individual tools;
no privileged shell/install/sudoers change is performed automatically. State is
private per SSH UID under /tmp; reboot/deleted state and other administrators
are outside the guarantee. Linux nft/legacy and real OpenSSH tests PASS in the
isolated lab; no customer-host validation is inferred.

La migrazione 000010 aggiunge `ssh_servers.host_key_fingerprint` vuoto per
l'inventario esistente. Prima delle connessioni occorre configurare il pin SHA256
della chiave del server tramite un canale fidato. Non copiare il fingerprint
della chiave client e non accettare automaticamente una chiave ricevuta in rete.
Rotazioni richiedono verifica indipendente e aggiornamento esplicito. Down perde
i pin host; salvare i pin autorizzati prima di un rollback operativo e riconfigurarli
dopo il re-upgrade. Nessuna migrazione o deploy applicativo eseguito in questa
sessione. Timeout locali non sostituiscono watchdog/transazione Linux remota;
E28-S1-T1 e TECH-DEBT-07 restano aperti.
## Notification channel deployment (E28-S1-T4/T5)

Migration 000012 and server worker must be deployed together. No channels are
automatically created, enabled or contacted at startup. Existing source history
is baselined without retroactive notifications. Production-sized baseline time,
queue throughput and audit retention have not been benchmarked; plan maintenance
and backup before migration. Down destroys the three alert tables and cannot
recall external messages. Root company deletion/reparenting requires channel
reassignment/removal first.

Alerting requires an operator-provisioned, stable, non-default `SSH_MASTER_KEY`
of at least 32 bytes; default/short key refuses alert encryption/activation/test.
Keep the existing master-key backup/rotation boundary; replacing it without
preserving old key access makes stored configuration undecryptable. No key or
provider credential is returned in API responses or diagnostics. Do not put
credentials in URLs passed to shell commands or logs.

Egress must support public HTTPS 443 and SMTP TLS 465/587 with trusted certificates.
No redirects or environment proxy, no private relay/webhook exception. Tenant
admins explicitly choose destinations and consent to synthetic tests. Teams
targets current Workflows with Adaptive Cards. Provider acceptance does not prove
receipt; possible duplicate retries and already in-flight sends after disable
remain explicit. [Protocol and queue contract](../domains/active_defense_and_soar.md#alerting-hub-e28-s1-t4t5-in-progress).
