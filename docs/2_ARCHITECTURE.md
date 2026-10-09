# 2. System Architecture & Technical Specifications

> **Stato:** APPROVATO (Fase 2: System Design - Struttura Autonoma v6.0)  
> **Ultimo Aggiornamento:** 2026-10-09  
> **Autore / Ruolo:** TechLead-Engineer-Agent (System Architect)  
> **Documenti Correlati:** [0_REQUIREMENTS.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/0_REQUIREMENTS.md) | [1_PROJECT_CONTEXT.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/1_PROJECT_CONTEXT.md)

---

## 1. Diagramma Architetturale di Sistema

La piattaforma adotta un'architettura containerizzata multi-servizio Clean Architecture con disaccoppiamento tra il layer API REST/SSE/WebSocket, il motore di ricognizione/scansione Nmap, il gestore di connessioni SSH sicure e firewall, il database relazionale PostgreSQL e la Single Page Application (SPA).

```mermaid
graph TD
    Client["Browser / Utente (SecOps / Admin / Auditor)"]
    
    subgraph Frontend_Layer ["Frontend Container"]
        SPA["React 18 + TypeScript + Vite + Tailwind CSS"]
        UI_Modules["UI Modules:
- Dashboard & Findings
- Tree Asset Explorer
- LAN Discovery (Real-Time Live Table)
- Scan Detail (Real-Time Progress & Findings Stream)
- Centralized Reports Hub
- SSH Host & iptables GUI + Web Terminal (xterm.js)
- Debug & Testing Live Console Drawer (Ctrl+Shift+D)"]
        NginxProxy["Reverse Proxy (Nginx)"]
    end

    subgraph Backend_Layer ["Backend Container (Golang 1.22 App)"]
        API["REST & SSE Router (Chi Router)"]
        AuthMiddleware["Tenant & RBAC JWT Middleware"]
        WSGateway["WebSocket Gateway (gorilla/websocket)"]
        EventHub["Real-Time SSE Event Broadcaster (Pub/Sub Hub)"]
        
        subgraph Core_Services ["Application Services (Clean Architecture)"]
            TargetService["Target & Tree Asset Service"]
            ScanOrchestrator["Scan Orchestrator (Incremental Findings & Alive-Only Filter)"]
            DiscoveryEngine["LAN Discovery Engine (Real-Time Live Host Sweeper)"]
            SchedulerService["Scheduler Service (Cron Worker)"]
            CVERegistry["CVE/CPE Correlation Engine"]
            ReportsHubService["Reports Hub Engine (Executive & Tech PDF/CSV)"]
            SSHManager["SSH Remote Host Manager (Crypto AES-256-GCM)"]
            FirewallService["iptables/nftables Service & Anti-Lockout Timer"]
        end

        subgraph OS_Tools ["System & Remote Execution"]
            NmapBinary["Nmap Binary (Secure exec.CommandContext)"]
            XMLParser["Stream XML Parser (Nmap Output)"]
            SSHClient["Native Go SSH Client (golang.org/x/crypto/ssh)"]
        end
    end

    subgraph Remote_Infra ["Target Systems / Remote Linux Servers"]
        RemoteServer["Remote Linux Server (iptables / SSH Daemon port 22)"]
    end

    subgraph Data_Layer ["Data Container (PostgreSQL 16)"]
        PostgresDB[("PostgreSQL Database
- Multi-Tenant Tables
- Target Tree Groups (parent_id)
- Local CVE/CPE Catalog
- Scan History & Alive Findings (Incremental Writes)
- SSH Servers (Encrypted Secrets)
- Firewall Rule Snapshots")]
    end

    Client -->|HTTPS / Port 443| NginxProxy
    NginxProxy -->|Static Assets| SPA
    NginxProxy -->|Proxy /api/v1| API
    NginxProxy -->|Live SSE text/event-stream| API
    NginxProxy -->|Upgrade /ws/v1| WSGateway
    
    API --> AuthMiddleware
    AuthMiddleware --> Core_Services
    API --> EventHub
    WSGateway --> SSHManager
    
    ScanOrchestrator -->|Push Live Events| EventHub
    DiscoveryEngine -->|Push Discovered Devices| EventHub
    
    ScanOrchestrator -->|Incremental DB Writes| PostgresDB
    DiscoveryEngine -->|Incremental Device Persistence| PostgresDB
    
    ScanOrchestrator -->|Raw Commands| NmapBinary
    DiscoveryEngine -->|Fast Ping/ARP Sweep| NmapBinary
    NmapBinary -->|Raw XML Stream| XMLParser
    XMLParser -->|Filtered Up Hosts| CVERegistry
    
    SSHManager -->|Encrypted Session| SSHClient
    FirewallService -->|iptables Commands & Safety Rollback| SSHClient
    SSHClient -->|SSH Port 22 / Pseudo-TTY| RemoteServer
    
    Core_Services -->|SQL Queries & Migrations| PostgresDB
    ReportsHubService -->|Query Findings & Tenant Branding| PostgresDB
```

---

## 2. Stack Tecnologico Completo

| Componente | Tecnologia | Versione / Note |
| :--- | :--- | :--- |
| **Backend Language** | **Golang** | `1.22+` (Goroutine, worker pool, type safety) |
| **HTTP Router & SSE** | **go-chi/chi/v5** + `http.Flusher` | Router HTTP snello con supporto nativo Server-Sent Events (`text/event-stream`) |
| **WebSocket Gateway** | **gorilla/websocket** | Gateway WebSocket dedicato per il terminale remoto interattivo PTY |
| **Event Broadcaster** | **In-Memory Pub/Sub Hub Go** | Canali Go bufferizzati con multiplexing per tenant/sessione e safety timeout |
| **SSH Client Nativo** | **golang.org/x/crypto/ssh** | Client SSH Go conforme RFC 4251 con supporto chiavi RSA/Ed25519 e password |
| **Crittografia a Riposo** | **Crypto AES-256-GCM** | Cifratura autenticata hardware-accelerata per credenziali SSH |
| **Database & Driver** | **PostgreSQL 16** + **jackc/pgx/v5** | Relational DB con supporto JSONB, indici B-Tree e salvataggio incrementale findings |
| **Database Migrations** | **golang-migrate/migrate/v4** | Migrazioni SQL versionate e riproducibili |
| **Network Scanner** | **Nmap** (containerizzato) | `7.94+` con flag minimi di privilegio e parsing streaming |
| **Task Scheduling** | **robfig/cron/v3** | Cron scheduler in-memory con persistenza DB |
| **PDF Reporting** | **jung-kurt/gofpdf** | Generatore nativo PDF (layout Executive e Tecnico) |
| **Frontend Framework** | **React 18** + **TypeScript** | Web UI reattiva modulare |
| **Build Tool Frontend** | **Vite 5+** | Bundling veloce, HMR e build ottimizzata |
| **Live Stream Client** | **Native EventSource / Fetch SSE** | Connessione stream persistente con riconnessione automatica e parsing eventi |
| **Terminale Web** | **@xterm/xterm** + **@xterm/addon-fit** | Emulatore terminale ANSI ad alte prestazioni nel browser |
| **Debug & Diagnostics** | **Custom Ring-Buffer Logger + Interceptors** | Cattura trasparente chiamate Axios, eventi SSE, WebSocket ed eccezioni JS con export Markdown |
| **UI & Styling** | **Tailwind CSS** + **Lucide Icons** | Design system moderno con dark mode nativo e grafici **Recharts** |
| **Containerization** | **Docker & Docker Compose** | Multi-container: `postgres`, `backend` (con Nmap), `frontend` (con Nginx) |

---

## 3. Design Pattern e Struttura del Codice (Clean Architecture)

```
backend/
├── cmd/
│   ├── server/             # Entrypoint applicazione principale (main.go)
│   └── cve-sync/           # Utility CLI per sincronizzare il database CVE locale
├── internal/
│   ├── domain/             # Entità di business pure, errori di dominio e interfacce
│   ├── repository/         # Implementazione persistenza dati PostgreSQL (SQL/pgx)
│   │   ├── postgres/       # Repository concreti per entità
│   │   └── migrations/     # File di migrazione .sql versionati
│   ├── service/            # Casi d'uso e logica di business
│   ├── scanner/            # Wrapper Nmap, parser XML e profili (Alive-Only)
│   └── delivery/           # Handlers HTTP REST, SSE, WebSocket e Middleware
└── pkg/
    ├── crypto/             # AES-256-GCM encryption/decryption helper
    ├── auth/               # JWT token & Password Hasher
    ├── report/             # Generatori PDF e CSV
    └── validator/          # Validatore IP/CIDR/Porte/Firewall rules
```

---

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
```

---

## 5. Documentazione Dettagliata per Dominio

Per garantire modularità ed efficienza di contesto, i dettagli tecnici e i contratti specifici sono frammentati nei file di dominio:

1. **[Scanning & LAN Network Discovery](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/domains/scanning_and_discovery.md)**: Profili Nmap, filtro host spenti (Alive-Only), scansione LAN ad alta velocità, database IEEE OUI e specifiche stream SSE (`/api/v1/network-discovery/stream`, `/api/v1/scans/{id}/stream`).
2. **[Remote SSH & iptables Firewall Manager](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/domains/ssh_and_firewall.md)**: Gestione credenziali e cifratura AES-256-GCM, Anti-Lockout Safety Timer, ispezione e modifica regole iptables, WebSocket Web Terminal SSH PTY.
3. **[Asset Tree Management & Centralized Reports Hub](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/domains/reporting_and_assets.md)**: Albero gerarchico asset (`TargetGroup`), tipi di target e blacklist, archivio storico `/reports` e motori di esportazione PDF White-Label (Executive/Technical) e CSV RFC 4180.
4. **[Debug & Testing Live Console Drawer](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/domains/debug_console.md)**: In-Memory Ring Buffer (500 log), intercettori Axios/SSE/WS/JS, sanificazione credenziali `[REDACTED]` ed esportatore Markdown preformattato per AI.

---

## 6. Contratti API REST, SSE e WebSocket (Riepilogo)

Tutte le API rispondono sotto `/api/v1` con payload standard `{ "success": true, "data": ..., "error": null }` e autenticazione Bearer JWT:

- `POST /api/v1/auth/login`, `GET /api/v1/auth/me`
- `GET`, `POST`, `PUT`, `DELETE /api/v1/targets`
- `GET`, `POST`, `PUT`, `DELETE /api/v1/target-groups`
- `POST /api/v1/network-discovery/scan`, `GET /api/v1/network-discovery/results`, `POST /api/v1/network-discovery/import-targets`
- `GET /api/v1/network-discovery/stream?session_id={id}&token={jwt}` (SSE)
- `POST /api/v1/scans`, `GET /api/v1/scans`, `GET /api/v1/scans/{id}`, `POST /api/v1/scans/{id}/cancel`
- `GET /api/v1/scans/{id}/stream?token={jwt}` (SSE)
- `GET /api/v1/reports`, `GET /api/v1/reports/{scan_id}/download` (PDF/CSV)
- `GET`, `POST`, `DELETE /api/v1/ssh/servers`, `POST /api/v1/ssh/servers/{id}/test-connection`
- `GET`, `POST`, `DELETE /api/v1/ssh/servers/{id}/firewall/rules`
- `POST /api/v1/ssh/servers/{id}/firewall/apply-with-safety-timer`, `confirm`, `rollback`
- `GET /ws/v1/ssh/servers/{id}/terminal?token={jwt}` (WebSocket PTY)
