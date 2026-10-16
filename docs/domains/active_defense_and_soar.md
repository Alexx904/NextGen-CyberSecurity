# Active Defense & SOAR Light

Ultimo allineamento: 2026-10-17 · working tree · E28-S1-T1 in implementazione.

## Contratto backend

Decisione PO: bloccare una porta sul server del chokepoint; isolamento host separato.
Solo nodi SERVICE presenti nei chokepoint del grafo corrente del tenant, IPv4 letterale,
TCP/UDP e porta 1–65535. Il server SSH selezionato deve appartenere allo stesso tenant
e avere host IPv4 coincidente. DNS e IPv6 non sono supportati in questo incremento
iptables; nessuna risoluzione o selezione automatica ambigua. La porta TCP di gestione
SSH viene rifiutata. HOST e VULNERABILITY non sono convertiti implicitamente in blocchi.

- `POST /api/v1/attack-paths/chokepoints/{id}/preview`
- `POST /api/v1/attack-paths/chokepoints/{id}/remediate`

Entrambi richiedono ruolo SecOps/admin e body `{"server_id":"<uuid>"}`.
La risposta data contiene `node_id`, `server_id`, `rule`, `command`,
`safety_timeout_seconds` (60), e `firewall` solo dopo applicazione.
La regola è INPUT, posizione 1, DROP, destinazione IP del servizio, protocollo e porta
derivati dal grafo. Non sono ammessi comandi, IP o porte arbitrari nel body.
La preview non esegue SSH; l'applicazione rigenera e rivalida il grafo.

Errori: 400 richiesta/nodo non supportato, 401 autenticazione assente,
403 ruolo insufficiente, 404 chokepoint/server assente, 409 timer già attivo,
502 errore di esecuzione. Non esporre dettagli remoti nella risposta HTTP.

## Sicurezza e lifecycle

Riusare [SSH/firewall](ssh_and_firewall.md) per snapshot, watchdog remoto obbligatorio,
conferma e rollback. Nessuna conferma automatica. Le API esistenti sul server SSH
permettono status, conferma e ripristino. Il blocco interessa anche connessioni esistenti
alla porta selezionata. Non attesta che l'intero percorso di attacco sia neutralizzato.

## QA

Test generatore, isolamento tenant e associazione IP, porta SSH, nodo obsoleto,
preparazione watchdog fallita, concorrenza, errori SSH e rollback. Nessuna esecuzione
remota implicita: la prova reale richiede un laboratorio identificato.
