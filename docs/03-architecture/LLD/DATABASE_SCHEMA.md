# Database Schema

> Publication scope 2026-10-18: alert-channel / outbox / migration 000012 sections describe deferred T4/T5 design and local prototype evidence. Their code and routes are excluded from this release; published migration is 000011. Alert-specific commands are future gates.


> Responsabile: Lead Developer; revisione Software Architect. Schema descrittivo migrato, non nuova verifica del database.
> Fonte eseguibile: migrazioni versionate in `backend/internal/repository/migrations/`. Aggiornare questo documento con ogni modifica dello schema.

## 4. Schema Database (ERD Esteso)

E28-S1-T4 · 2026-10-18 · working tree: `000012` adds `alert_channels`
(tenant, optional root company, provider, enabled/events, encrypted config,
created/updated/deleted timestamps), `alert_observations` (tenant/entity key,
normalized identity signature) and `alert_jobs` (event payload, state, attempts,
next attempt, lease/token and redacted result code). Composite tenant/channel
and tenant/company foreign keys prevent cross-tenant references. Company FK
uses deferred NO ACTION so a tenant cascade can remove all owned data; removing
an in-use company alone fails. Guard triggers reject reparenting a company with
non-deleted channels. Deleting a channel soft-deletes metadata, disables it,
clears company/config and preserves delivery history. Pending old messages are
cancelled on relevant configuration changes; in-flight requests may complete.

Finding/discovery/scheduled terminal-status triggers enqueue in the source
transaction. Observation locks serialize source writes per tenant, with no new
messages for unchanged findings or latency/time/port-order-only updates. Existing
history is baselined at migration time without notifications. Delivery is
at-least-once with six attempts and 30-second leased claims; token/revision
checks reject stale ownership. Claim index covers pending/sending due times.
No automatic retention purge or production-scale migration benchmark is claimed.
Down drops only alert tables/triggers/functions and its composite company index;
it destroys alert config/observations/history, cannot recall sent messages and
requires a backup of the three tables for application rollback. Up/down/up is
tested only on the disposable local `nextgen_discovery_test` database.
[Contract/evidence](../../04-development/plans/E28-S1-T4-alerting.md), ADR-26.

M11-PERSIST-01 · 2026-10-18 · working tree: `000011` aggiunge
`discovery_sessions` e `discovery_devices`. ID sessione globale univoco;
FK composta `(tenant_id, session_id)` impedisce dispositivi di tenant diversi.
Un dispositivo JSONB per IP (`INET`), upsert durante la scansione; metadati e
dispositivi letti con un unico snapshot SQL. Indice tenant/data/ID per l'ultima
sessione, anche parziale; nessuno stato COMPLETED inferito da un riavvio.
CASCADE tenant→sessioni→dispositivi intenzionale per dati posseduti dal tenant.
Down rimuove solo le due nuove tabelle e perde lo storico discovery: eseguire
backup di entrambe prima di rollback applicativi. Prove solo sul DB sacrificabile
`nextgen_discovery_test`, porta locale 55432; [piano](../../04-development/plans/M11-PERSIST-01.md).

TEST-OPT-01 · Ultimo allineamento 2026-10-17 · working tree: schema e migrazioni invariati. Test repository PostgreSQL con tag `integration`, configurazione `TEST_DB_*` esclusiva su host locale e database dedicato `_test`/`_test_<id>`; nessun fallback su DB applicativo e indisponibilità = FAIL. Percorso rapido senza DB. Contratto operativo e comandi nel [Test Plan](../../05-qa/TEST_PLAN.md).

COMPANY-UX-02 · 2026-10-17 · working tree: `000008` aggiunge `ssh_servers.client_name VARCHAR(160) NOT NULL DEFAULT ''`, senza FK aziendale. `000009` aggiunge trigger `company_department_depth` e funzione associata: nuove scritture parent solo sotto radice del tenant, nodo con figli non spostabile sotto radice. Lock advisory transazionale per tenant sulle assegnazioni parent, nessuna migrazione dati legacy. Down 000009 rimuove trigger/funzione; down 000008 rimuove le etichette cliente. Up/down verificati solo sul DB dedicato `nextgen_company_ux02_test_20261017`.

AP-REVIEW-01 · 2026-10-17 · working tree: schema invariato, nessuna migrazione. Le letture paginate dei target aggiungono `id ASC` come spareggio a `created_at DESC`; i finding aggiungono `id ASC` dopo severità/porta. Ordinamento stabile a dati fermi, senza garanzia di snapshot consistente durante scritture concorrenti. Serve a evitare perdita/duplicazione per pareggi nella costruzione del [grafo euristico](../../domains/attack_path.md).

```mermaid
erDiagram
    TENANTS ||--o{ USERS : "contains"
    TENANTS ||--o{ TARGET_GROUPS : "owns"
    TENANTS ||--o{ TARGETS : "owns"
    TENANTS ||--o{ SCANS : "runs"
    TENANTS ||--o{ SCHEDULES : "schedules"
    TENANTS ||--o{ SSH_SERVERS : "manages"

    TARGET_GROUPS ||--o{ TARGET_GROUPS : "parent_of"
    TARGET_GROUPS ||--o{ TARGETS : "contains"
    TARGET_GROUPS ||--o{ SCANS : "targeted_by_group"

    TARGETS ||--o{ SCANS : "targeted_by"
    TARGETS ||--o{ SCHEDULES : "scheduled_for"

    SCAN_PROFILES ||--o{ SCANS : "configured_with"
    SCAN_PROFILES ||--o{ SCHEDULES : "used_by"

    SCANS ||--o{ FINDINGS : "produces"
    CVE_ENTRIES ||--o{ FINDINGS : "correlated_with"

    SSH_SERVERS ||--o{ FIREWALL_BACKUPS : "has_snapshots"
    TENANTS ||--o{ DISCOVERY_SESSIONS : "performs"
    DISCOVERY_SESSIONS ||--o{ DISCOVERY_DEVICES : "discovers"

    TARGET_GROUPS {
        uuid id PK
        uuid tenant_id FK
        uuid parent_id FK "Auto-referenziale per albero gerarchico"
        string name
        string description
        string color_tag
        timestamp created_at
        timestamp updated_at
    }

    TARGETS {
        uuid id PK
        uuid tenant_id FK
        uuid group_id FK
        string target_type
        string value
        string description
        boolean is_blacklisted
        timestamp created_at
    }

    SSH_SERVERS {
        uuid id PK
        uuid tenant_id FK
        string name
        string host
        integer port
        string username
        string auth_type "KEY o PASSWORD"
        text encrypted_secret "AES-256-GCM cipher"
        text public_key_fingerprint
        text host_key_fingerprint "SHA256 server pin, empty = untrusted"
        string status "ONLINE, OFFLINE, UNCHECKED"
        timestamp last_connected_at
        timestamp created_at
        timestamp updated_at
    }

    FIREWALL_BACKUPS {
        uuid id PK
        uuid tenant_id FK
        uuid server_id FK
        text raw_iptables_rules
        string label
        timestamp created_at
    }

    SCANS {
        uuid id PK
        uuid tenant_id FK
        uuid target_id FK "Opzionale se lanciato su cartella"
        uuid group_id FK "Opzionale per scansioni su cartella"
        uuid profile_id FK
        string trigger_type
        string status
        timestamp started_at
        timestamp completed_at
        jsonb summary_stats
        text error_message
        timestamp created_at
    }

    FINDINGS {
        uuid id PK
        uuid scan_id FK
        uuid tenant_id FK
        string host_ip
        string hostname
        integer port
        string protocol
        string service_name
        string service_version
        string cve_id FK
        float cvss_score
        string severity
        text raw_banner
        text remediation
        timestamp created_at
    }

    TENANTS {
        uuid id PK
        string name
        string slug
        string logo_url
        string primary_color
        string secondary_color
        timestamp created_at
        timestamp updated_at
    }

    USERS {
        uuid id PK
        uuid tenant_id FK
        string email
        string password_hash
        string role "SuperAdmin, TenantAdmin, SecOps, Auditor"
        string full_name
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }

    SCAN_PROFILES {
        uuid id PK
        uuid tenant_id FK "Opzionale per profili custom"
        string name
        string description
        string category "discovery, system_flaws, pentest, web_app, compliance, all_in_one"
        text raw_args
        integer timeout_seconds
        boolean is_default
        boolean is_custom
        timestamp created_at
        timestamp updated_at
    }

    CVE_ENTRIES {
        string cve_id PK
        text description
        float cvss_v31_score
        string cvss_v31_severity
        jsonb cpe_matchers
        timestamp published_at
        timestamp last_modified_at
        timestamp created_at
    }

    DISCOVERY_SESSIONS {
        uuid id PK
        uuid tenant_id FK
        string subnet
        timestamp started_at
        timestamp completed_at
        float duration_seconds
    }

    DISCOVERY_DEVICES {
        uuid session_id PK,FK
        uuid tenant_id PK,FK
        inet ip PK
        jsonb device
    }
```

---

## TECH-DEBT-05 - Authorization evidence (2026-10-19, working tree)

Migration `000013_scan_authorization` adds nullable `consent_snapshot JSONB` to
`scans` and `schedules`. Null preserves historical reads without fabricating
consent; application services/workers reject covered execution without evidence.
The public JSON name is `authorization`. Snapshots contain the versioned explicit
statement, exact target IDs/values, Nmap arguments, timeout, recurrence, tenant,
actor and server acceptance time; recurring evidence also records schedule ID.
Source snapshot changes are rejected by triggers; status/timestamps remain mutable.

`scan_authorization_audit`: UUID primary key, tenant/user UUIDs, object type
(scans/schedules), object UUID, database-generated recorded_at and evidence JSONB;
unique (object_type, object_id), index (tenant_id, recorded_at DESC, id). No inventory
FK is intentional: evidence survives scan/schedule/tenant deletion. Source inserts
and audit insertion share the source transaction. UPDATE/DELETE/TRUNCATE evidence
operations are rejected; PUBLIC mutation privileges are revoked. A database owner
can remove triggers; this is application-level immutability, not external notarization.
No automatic retention/purge policy or unlimited storage claim is introduced.

Runtime schedule writes update only last/next-run timestamps conditionally on the
same enabled schedule/configuration, never overwriting concurrent disable/edit.
Down removes the two snapshot columns, new audit table, triggers and functions;
it loses authorization history. Back up those columns/table before application
rollback. Up/down/up tests use only disposable `nextgen_consent_test` on 55432.
Migration 000012 remains the preexisting deferred alerting prototype; no release
or deployment of that work is implied by testing the integrated local tree.
[ADR-27](../ADR/README.md#adr-27--inventory-scan-authorization-snapshots-2026-10-19),
[implementation and evidence](../../04-development/plans/TECH-DEBT-05.md).
