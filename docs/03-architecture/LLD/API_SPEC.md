# API Specification

> Publication scope 2026-10-18: alert-channel / outbox / migration 000012 sections describe deferred T4/T5 design and local prototype evidence. Their code and routes are excluded from this release; published migration is 000011. Alert-specific commands are future gates.


> Responsabile: Lead Developer; revisione Software Architect. Inventario storico migrato: non equivale a una verifica endpoint per endpoint.
> Per nuove modifiche definire metodo/path, auth/tenant scope, request/response, validazione, errori, compatibilità e test prima dell'implementazione. Per SSE/WS includere eventi, riconnessione e lifetime.

## 6. Contratti API REST, SSE e WebSocket (Riepilogo)

Tutte le API rispondono sotto `/api/v1` con payload standard `{ "success": true, "data": ..., "error": null }` e autenticazione Bearer JWT:

- `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`, `GET /api/v1/auth/me`, `POST /api/v1/auth/stream-ticket` (Ticket Monouso TTL 30s)
- `GET /api/v1/tenants/current`, `PUT /api/v1/tenants/current/branding`
- `GET`, `POST`, `DELETE /api/v1/targets`, `GET /api/v1/targets/{id}`
- `GET`, `POST`, `PUT`, `DELETE /api/v1/target-groups`, `GET /api/v1/target-groups/{id}`, `POST /api/v1/target-groups/{id}/scan`
- `GET`, `POST`, `DELETE /api/v1/scan-profiles`
- `POST /api/v1/network-discovery/scan`, `GET /api/v1/network-discovery/results`, `POST /api/v1/network-discovery/import-targets`
- `GET /api/v1/network-discovery/stream?session_id={id}&ticket={ticket}` (SSE con ticket monouso, fallback `?token=`)

M11-PERSIST-01 · 2026-10-18 · working tree: payload discovery invariati.
POST scan rifiuta con 409 `SESSION_CONFLICT` un ID già utilizzato, prima di Nmap.
GET results restituisce la sessione più recente del tenant per avvio/ID, anche
parziale; 404 solo in assenza di sessioni. `completed_at` zero Go
(`0001-01-01T00:00:00Z`) indica completamento non confermato. Snapshot PostgreSQL
recuperabili dopo restart; nessuna ripresa automatica del job.
SSE può ripetere `device_discovered` per aggiornare lo stesso IP con arricchimenti;
client esegue upsert per IP. Snapshot e terminale persistiti prima della rispettiva
notifica. Replay sottoscritto prima della lettura, isolato per tenant/sessione;
errore di lettura DB → evento error e chiusura. Dettagli e limiti nel
[dominio scanning](../../domains/scanning_and_discovery.md#5-api-e-streaming).
- `POST /api/v1/scans`, `GET /api/v1/scans`, `GET /api/v1/scans/{id}`, `POST /api/v1/scans/{id}/cancel`
- `GET /api/v1/scans/{id}/findings`, `GET /api/v1/scans/{id}/diff` (Scan Diff: NEW, RESOLVED, PERSISTENT)
- `GET /api/v1/scans/{id}/stream?ticket={ticket}` (SSE con ticket monouso, fallback `?token=`)
- `GET /api/v1/reports`, `GET /api/v1/reports/{id}/download` (PDF/CSV), `GET /api/v1/reports/company/{groupId}/download` (PDF/CSV)
- `GET /api/v1/scans/{id}/reports/pdf`, `GET /api/v1/scans/{id}/reports/csv`
- `GET /api/v1/schedules`, `GET /api/v1/schedules/{id}`, `POST /api/v1/schedules`, `PUT /api/v1/schedules/{id}/toggle`, `DELETE /api/v1/schedules/{id}`
- `GET /api/v1/cve/stats`, `POST /api/v1/cve/sync`
- `GET /api/v1/attack-paths/graph`, `GET /api/v1/attack-paths/chokepoints`
- `GET`, `POST`, `PUT`, `DELETE /api/v1/ssh/servers`, `GET /api/v1/ssh/servers/{id}`, `POST /api/v1/ssh/servers/{id}/test-connection`
- `GET /api/v1/ssh/servers/{id}/firewall`, `POST /api/v1/ssh/servers/{id}/firewall/rules`, `DELETE /api/v1/ssh/servers/{id}/firewall/rules/{line}`
- `POST /api/v1/ssh/servers/{id}/firewall/apply`, `POST /api/v1/ssh/servers/{id}/firewall/confirm`, `POST /api/v1/ssh/servers/{id}/firewall/rollback`
- `GET /ws/v1/ssh/servers/{id}/terminal?ticket={ticket}` (WebSocket PTY con ticket monouso, fallback `?token=`)
- *(M3)* `POST /api/v1/attack-paths/chokepoints/{id}/remediate`, `POST /api/v1/virtual-patching/generate`, `GET/POST/PUT/DELETE /api/v1/alert-channels`, `POST /api/v1/alert-channels/{id}/test`
- *(M4)* `GET /api/v1/compliance/overview`, `GET /api/v1/compliance/frameworks/{framework}`, `GET /api/v1/compliance/dossier/pdf`
- *(M5)* `POST /api/v1/copilot/playbook`, `POST /api/v1/copilot/ask`, `POST /api/v1/copilot/executive-summary`

---

## Autorità del contratto

### M11-FIREWALL-01 · Revision and remote lifecycle · 2026-10-18 · working tree

All paths below use `/api/v1/ssh/servers/{id}/firewall`. Existing authentication,
tenant SSH inventory and SecOps/TenantAdmin/SuperAdmin authorization apply.

- GET adds `revision`, a SHA256 of canonical rules under the remote lock.
- PUT `/policy`: `chain`, `policy` ACCEPT/DROP, `expected_revision`, optional
  `safety_timeout_seconds` (0 defaults to 60; accepted 10..300).
- PUT `/rules/{line}`: complete add-rule DTO, required `expected_revision`;
  replaces the line using the supplied fields, with no preservation of omitted
  advanced flags. Positive line, zero `position`.
- POST `/rules/move`: `chain`, positive/different `from` and `to`, required
  `expected_revision`, optional timeout. Preserves full saved rule data.
- Existing POST `/rules` accepts optional `expected_revision`; existing DELETE
  `/rules/{line}` accepts the same optional query parameter. New UI always sends it.
- New edit payloads are bounded to 32 KiB and reject trailing JSON. Invalid input
  is 400; stale revision is 409 `FIREWALL_STATE_CHANGED`; pending conflict is 409
  `SAFETY_TIMER_ACTIVE`; expired/absent confirmation is 409 `NO_SAFETY_TIMER_ACTIVE`.
  Missing remote prerequisites/privileges retain sanitized 502 diagnostics.
- Snapshot, watchdog readiness and mutation are one host operation; DB backup
  failure leaves the remote watchdog armed and returns error. Confirm checks the
  matching tenant/server backup; rollback carries snapshot data on SSH stdin.
  [Operational boundaries](../../domains/ssh_and_firewall.md).

### M11-CLOSE-01 · Mutazioni firewall · 2026-10-18 · working tree

POST regole: source/destination vuoti o IPv4/CIDR IPv4; interfacce ASCII max 15
caratteri, wildcard `+` finale; commento UTF-8 max 255 byte senza controlli.
Safety timer obbligatorio per POST/DELETE: flag legacy false non lo disattiva.
Timeout omesso/0 = 60s, altri valori ammessi 10..300s; input invalidi HTTP 400
`VALIDATION_FAILED`. Nessuna nuova API o migrazione. Restano aperti i limiti
operativi watchdog nel [dominio SSH](../../domains/ssh_and_firewall.md).

### SSH-TOOLS-01 · 2026-10-17 · working tree

`GET /api/v1/ssh/servers/{id}/firewall`: isolamento tenant e RBAC SecOps/admin
invariati. Errori 502 additivi `FIREWALL_UNAVAILABLE` (iptables non disponibile)
e `FIREWALL_PRIVILEGES_REQUIRED` (root/sudo non interattivo necessari), con
messaggi sanificati; altri errori SSH conservano `SSH_EXECUTION_FAILED`.
Nessuno stderr remoto arbitrario incluso nei nuovi messaggi. Nginx e Vite
inoltrano `/ws/` con upgrade al backend, preservando path/query e controlli
del ticket monouso; nessun cambiamento al protocollo terminale o ai permessi.

### COMPANY-UX-02 · 2026-10-17 · working tree

- `POST/PUT /target-groups`: padre deve essere un'azienda radice; nodo con figli non può diventare reparto; violazioni restituiscono errore validazione anche se intercettate dal trigger DB. Contratto tenant invariato.
- SSH create/read/list/update: `client_name` stringa additiva, max 160 caratteri; create opzionale/default vuoto; update omesso conserva, vuoto azzera. Non correlato alle aziende. Ruoli e segreti invariati.
- `GET /reports/company/{groupId}/download`: stesso endpoint per azienda o reparto. Aggrega discendenti e pagine, PDF con inventario per reparto, CSV con colonna finale `reparto`. CSV singola scansione invariato.
- Attack Path UI non invia `min_cvss`, parametro API compatibile. Il guard aziendale è client-side e non modifica i permessi API.

### AP-REVIEW-01 · 2026-10-17 · working tree

`GET /attack-paths/graph`: `min_cvss` deve essere finito in [0,10], altrimenti 400 `VALIDATION_ERROR`. Risposta additiva `analysis` con modello, avvisi, conteggi letti, cap 50 e flag inventario/percorsi troncati. `paths[].edge_ids` completo e orientato, coerente con nodi e passaggi. Gli ID generati appartengono allo snapshot e distinguono TCP/UDP; nessuna persistenza garantita fra modelli. `cumulative_risk` rimane un indice euristico, non probabilità. `GET /chokepoints` usa lo stesso modello con soglia 0. JWT e perimetro tenant invariati; scope aziendale è una proiezione frontend, senza nuova autorizzazione per azienda. Contratto completo e limiti nel [dominio Attack Path](../../domains/attack_path.md).

### E28-S1-T1 · 2026-10-17 · working tree

Preview ed esecuzione del blocco porta chokepoint: contratto canonico nel
[dominio Active Defense](../../domains/active_defense_and_soar.md#contratto-backend).
POST autenticati con RBAC SecOps/admin; conferma e rollback riusano le API firewall.
API additive, nessuna migrazione DB. Nessuna UI di remediation inclusa.

### SCAN-TIME-01 · 2026-10-16 · working tree

- `GET /api/v1/scans?ids=<uuid,...>`: variante additiva autenticata per il monitor del lotto. Da 1 a 100 ID, UUID non nulli, deduplicati; ogni scansione letta mediante GetScan con tenant corrente. Non applica limit/offset allo storico. Envelope invariato con items/total/limit/offset. 400 su input vuoto/invalido/eccessivo, 404 se una scansione è assente o non appartiene al tenant (nessuna risposta parziale), 500 su errore repository. Il vecchio elenco paginato resta invariato senza ids.
- `ScanResponse.progress` opzionale in GET per ID e lettura per lotto: snapshot attivo in memoria, non colonna DB. Proprietà aggiuntive SSE progress: phase, phase_percentage (percentuale della sola fase Nmap), remaining_seconds (stima complessiva del job), observed_at UTC. percentage esistente rappresenta stima totale pesata del job. Omissione di ETA significa indisponibile; nessun uso del timeout come countdown.
- Snapshot aggiornato durante Nmap e nelle fasi successive; cancellato a stato terminale e cleanup del job. Accesso allo snapshot solo dopo verifica tenant. Riavvio del backend perde telemetria, non inventa avanzamento; ripresa da nuove statistiche quando disponibili.
- La UI conserva gli ID dell'ultimo lancio per tenant nella sessione del browser e legge blocchi da 50 ogni 5s senza sovrapposizione. I job terminali non sono riletti. La route dettaglio continua a usare SSE con stream-ticket.

Contratto, formula ETA e test in [SCAN_PROGRESS_PLAN](../../04-development/SCAN_PROGRESS_PLAN.md).

Questo file governa l'inventario dei contratti; i dettagli verticali sono collegati da [DOMAIN_MAP](../DOMAIN_MAP.md). DTO e test verificano l'implementazione. Se viene introdotto OpenAPI, lo schema diventa la fonte strutturata e questo file ne conserva guida e link, senza duplicare i payload.
## Virtual patch advisory export (E28-S1-T3, working tree)

`GET /api/v1/findings/{id}/virtual-patches`: authenticated tenant read, all existing
roles including Auditor. UUID required; 400 invalid, 401 missing tenant, 404
missing/foreign finding, 500 unexpected failure with redacted message. Standard
`data` envelope, `Cache-Control: no-store`. Response: `finding_id`, `cve_id`,
`catalog_revision`, `supported`, `reason`, `limitations: string[]`,
`product`, `strategy` (uri_indicator/endpoint_restriction), `operational_impact`,
`covered_cves: string[]`, `sources: string[]`,
`files: {engine, filename, content}[]`. Unsupported findings
return 200 with `supported: false` and empty files. No body, arbitrary regex,
installation or external retrieval. [Coverage and limits](../../domains/active_defense_and_soar.md#virtual-patch-templates-e28-s1-t3-in-progress).

## Alert channel management (E28-S1-T4/T5, working tree)

All routes require authenticated tenant and TenantAdmin/SuperAdmin; SecOps and
Auditor receive 403. `GET /api/v1/alert-channels/` lists metadata only.
`POST /api/v1/alert-channels/` creates (201), `PUT /api/v1/alert-channels/{id}`
updates (200), `DELETE /api/v1/alert-channels/{id}` soft-deletes/scrubs config
(204). Save body: `name`, `kind` (slack/teams/telegram/webhook/smtp), optional
`company_id` (root company, absence means tenant), `enabled`, unique `events`
array, write-only `config`. Update requires `expected_updated_at` matching list
metadata; omitted config preserves existing credentials, provider change needs
replacement. Responses contain id/tenant/company/name/kind/enabled/events and
created/updated timestamps, never endpoint/token/password/config. Error 400
invalid/unknown/trailing/oversized JSON or config, 404 foreign/missing company or
channel, 409 stale update, 500 unexpected failure with redacted detail. No-store.
Alert key unavailable/default/shorter than 32 bytes: 503
`ALERT_ENCRYPTION_UNAVAILABLE`. `DELETE /target-groups/{id}` returns 409
`GROUP_IN_USE` if channel references prevent deleting the company; remove/reassign
channels first.

`POST /api/v1/alert-channels/test`: body `{kind, config,
confirmed_external_send: true}` sends one fixed synthetic event, not saved or
retried. No real tenant/host/CVE data in that test event. Success 200 means
provider acceptance and requires checking receipt separately; rejected test 502
returns a fixed redacted code. Missing acknowledgement 400. No arbitrary message
or target URL in a GET/query string. Config fields and HMAC receiver contract,
event semantics and network policy: [Active Defense](../../domains/active_defense_and_soar.md#alerting-hub-e28-s1-t4t5-in-progress).
