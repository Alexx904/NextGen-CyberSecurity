# API Specification

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
