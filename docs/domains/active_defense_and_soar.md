# Active Defense & SOAR Light

> Publication scope 2026-10-18: alert-channel / outbox / migration 000012 sections describe deferred T4/T5 design and local prototype evidence. Their code and routes are excluded from this release; published migration is 000011. Alert-specific commands are future gates.


Ultimo allineamento: 2026-10-18 · working tree · E28-S1-T1 `[~]`, awaits PO.

Operational test uses the named isolated SSH fixture, production defense/SSH/
firewall services and a real loopback HTTP endpoint. Deterministic tenant graph
and in-memory backup fixture; not a full-graph/PostgreSQL E2E claim. Preview and
foreign/stale rejections preserve revision. Remediation blocks fresh HTTP requests,
fresh-instance confirmation retains the block, rollback restores reachability and
initial revision. PASS 3.01s; [evidence](../04-development/plans/E28-S1-T1-chokepoint-remediation.md).
Shared watchdog nft/legacy/OpenSSH gates now PASS under ADR-25; host/UID/state
limitations remain in the SSH domain. UI remediation remains E28-S1-T2.

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

E28-S1-T2 UI (working tree): SecOps/TenantAdmin/SuperAdmin can evaluate a current
SERVICE chokepoint; unsupported node types remain disabled, Auditor read-only.
Choose a matching tenant SSH server explicitly, request backend preview and
acknowledge service impact before applying. Missing host pin blocks mutation.
Only node ID/server_id are sent; no client-supplied commands or ports. Server
change invalidates preview/consent; stale responses are ignored. Successful apply
opens the existing firewall safety manager; uncertain failure permits manual
status verification without retry. Graph/scope changes dismiss unapplied dialogs.
[Plan and isolated QA boundary](../04-development/plans/E28-S1-T2-chokepoint-ui.md).
Final frontend gate PASS (215 tests/build). Authorized isolated browser workflow
PASS: preview/consent, remote countdown, HTTP DROP, confirm, manual rollback and
unconfirmed expiry with restored HTTP. After expiry, explicit reload verified
remote state; automatic refresh is not attested. Desktop/mobile, focus cycling
and Escape checked. E28-S1-T2 `[~]`, awaits PO; evidence in linked plan.

Riusare [SSH/firewall](ssh_and_firewall.md) per snapshot, watchdog remoto obbligatorio,
conferma e rollback. Nessuna conferma automatica. Le API esistenti sul server SSH
permettono status, conferma e ripristino. Il blocco interessa anche connessioni esistenti
alla porta selezionata. Non attesta che l'intero percorso di attacco sia neutralizzato.

## QA

## Virtual patch templates (E28-S1-T3, in progress)

`GET /api/v1/findings/{id}/virtual-patches` is an authenticated, tenant-scoped
read for all existing authenticated roles, including Auditor. It never executes
or installs files. Missing/foreign finding 404, invalid UUID 400, missing tenant
401, unexpected error 500 with redacted detail; success has `Cache-Control:
no-store`. Body contains finding/CVE, catalog revision, support status/reason,
limitations, source URLs and `{engine, filename, content}` files.

Catalog `2026-10-18.2` covers eight CVEs across Apache HTTP Server, Citrix
ADC/Gateway, Grafana, VMware vCenter and Confluence on TCP HTTP-family services.
The [catalog matrix](../04-development/plans/E28-S1-T3-virtual-patching.md#expanded-catalog-2026-10-182)
distinguishes URI indicators from endpoint restrictions and their operational
impact. Response includes product, strategy, operational impact and covered CVEs.
Applicability is unverified:
operator checks Apache version, configuration and proxy path. Fail2ban filter
and initially disabled jail, CrowdSec nginx acquisition/scenario (existing nginx
parser required), ModSecurity rule and Nginx server guard are exported. Templates
do not interpolate finding banners, addresses or arbitrary client parameters.
Other CVEs/services return no files and a clear unsupported reason.

Finding detail opens a keyboard-accessible modal with per-engine preview,
copy/download, sources and limitations. Failed clipboard/download does not report
success. Late responses after close/finding change are ignored. Server/App wiring
and the vulnerability detail are shared consumers of this contract.

These are conservative indicators/restrictions with false positives and bypass limitations.
Log-driven engines react after the first request. Raw URI, authentic client IP,
trusted proxy configuration and a CrowdSec remediation profile/bouncer are
prerequisites. Native engine checks and browser verification are distinct from
template unit tests. No vulnerability fix or protection guarantee is asserted.
Sources/design and current gate status: [T3 plan](../04-development/plans/E28-S1-T3-virtual-patching.md).

E28-S1-T3 `[~]`: Final All PASS (220 frontend tests), native Nginx syntax for
all eight entries and actual QA browser export/unsupported/mobile flows PASS.
Native Fail2ban/CrowdSec/ModSecurity checks not run; no installation/prevention
claim. PO acceptance remains separate.

### Remediation QA

Test generatore, isolamento tenant e associazione IP, porta SSH, nodo obsoleto,
preparazione watchdog fallita, concorrenza, errori SSH e rollback. Nessuna esecuzione
remota implicita: la prova reale richiede un laboratorio identificato.

## Alerting hub (E28-S1-T4/T5, in progress)

TenantAdmin/SuperAdmin configure channels in Settings; other roles do not fetch
channel metadata or change recipients. Tenant or root-company scope; seven event
choices: `finding.high`, `finding.critical`, `host.discovered`, `host.changed`,
`port.discovered`, `scan.completed`, `scan.failed`. Critical/High CVEs are newly
observed per tenant/host/protocol/port/CVE, with High→Critical escalation and no
downgrade notification. New company assignment/channel creation does not replay
historical findings. Host signature includes MAC/vendor/hostname/status and
sorted distinct open ports, excluding latency/timestamps. Discovery with unknown
or ambiguous company routes only to tenant-wide channels. Scheduled terminal
states are observed directly; certificate-expiry alerts are unavailable.

Metadata CRUD and test API: [API specification](../03-architecture/LLD/API_SPEC.md#alert-channel-management-e28-s1-t4t5-working-tree).
Configuration fields by provider: HTTPS `endpoint` (Slack/Teams/Webhook),
`hmac_secret` minimum 32 bytes for generic Webhook; Telegram `token`/`chat_id`;
SMTP `smtp_host`, port 465/587, from/to address list (maximum 10 recipients),
optional paired username/password. Configuration is write-only and tenant/channel/
provider-bound AES-GCM, using the existing master-key mechanism with distinct
AAD. Alert encryption/activation/test refuses the known default or a master key
shorter than 32 bytes (503), without changing existing SSH behavior. No credential
or endpoint is returned. New channels default disabled in UI;
editing metadata omits config and preserves secrets. Provider changes require a
new config; optimistic `expected_updated_at` protects against stale writes.

Only public HTTPS on 443 and SMTP implicit TLS 465/required STARTTLS 587. DNS
answers are validated immediately before dialing literal addresses; any private,
loopback, link-local, documentation, benchmark, multicast or special-use answer
rejects the connection. No environment proxy or redirect. Private relays and
internal webhook targets are unsupported in this increment. HTTP response bodies
are bounded/redacted; 429/5xx retry with bounded provider delays. Slack requires
`ok`, Telegram checks JSON `ok`, Teams Workflows uses Adaptive Card payloads.
SMTP never authenticates before verified TLS; DATA acceptance is not retried due
only to QUIT failure. No provider-account or final-delivery guarantee is claimed.

Generic webhook: exact JSON bytes, `X-NextGen-Timestamp` Unix seconds,
`X-NextGen-Event-ID` stable event UUID and `X-NextGen-Signature` containing
`sha256=<hex HMAC-SHA256(secret, timestamp + "." + eventID + "." + body)>`.
Receivers should enforce timestamp freshness and deduplicate event IDs. Payload
contains event/tenant/company identifiers, occurrence time, type, source entity,
IP/port/CVE/severity when present; no raw banners. Retries may duplicate messages.

Source DB triggers and durable outbox use ADR-26. Six attempts, exponential
5..160s base with small deterministic jitter/provider delay up to one hour,
30-second lease and token-checked completion. Config/scope/filter changes cancel
old unclaimed messages; workers recheck live metadata. Disable/delete cannot
recall a request already in flight. Automatic audit retention/purge is not yet
implemented. Root-company reparent/delete with channels is rejected to preserve
recipient scope; remove/reassign channels first.

Settings: loading/error/empty, per-provider fields, tenant/company/event choices,
disabled default, metadata edit without credential reveal and two-step removal.
Explicit acknowledgement gates a single synthetic test before saving; changing
config invalidates test consent/result. Test uses random event ID, zero tenant
ID, no host/CVE, does not persist/activate the channel and never retries. Config
is kept only in component state and cleared on close/save; debugger redacts the
entire `config` object and standalone `endpoint`, including URL-path secrets.
Accepted test response asks the operator to verify receipt; no false delivery
claim. Mandatory final/browser gates still open; [plan](../04-development/plans/E28-S1-T4-alerting.md).
