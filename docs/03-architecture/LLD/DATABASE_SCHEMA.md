# Database Schema

> Responsabile: Lead Developer; revisione Software Architect. Schema descrittivo migrato, non nuova verifica del database.
> Fonte eseguibile: migrazioni versionate in `backend/internal/repository/migrations/`. Aggiornare questo documento con ogni modifica dello schema.

## 4. Schema Database (ERD Esteso)

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
        string target_subnet
        integer total_hosts
        integer alive_hosts
        string status "RUNNING, COMPLETED, FAILED"
        timestamp started_at
        timestamp completed_at
        timestamp created_at
    }

    DISCOVERY_DEVICES {
        uuid id PK
        uuid session_id FK
        uuid tenant_id FK
        string ip
        string mac_address
        string vendor
        string hostname
        float latency_ms
        jsonb ports_open
        timestamp created_at
    }
```

---
