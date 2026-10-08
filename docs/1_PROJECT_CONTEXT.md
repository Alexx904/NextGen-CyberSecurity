# 1. Project Context & Non-Functional Requirements (NFR)

> **Stato:** APPROVATO (Fase 2: System Design - Revisione Milestone 1.1)  
> **Ultimo Aggiornamento:** 2026-10-08  
> **Autore / Ruolo:** TechLead-Engineer-Agent (System Architect & Tech Lead)  
> **Documento Correlato:** [0_REQUIREMENTS_AND_USE_CASES.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/0_REQUIREMENTS_AND_USE_CASES.md)

---

## 1. Executive Summary & Obiettivi del Progetto

**NextGen CyberSecurity Platform** è una soluzione all-in-one per la gestione del rischio e della vulnerabilità di rete, la ricognizione ad alta velocità e l'amministrazione difensiva dei server, ingegnerizzata secondo moderni standard di sicurezza, modularità e alte prestazioni.

Il sistema supera il tradizionale paradigma del "report statico di vulnerabilità" fornendo:
- Un motore di scansione di rete black-box altamente performante ed elastico basato su **Go** e **Nmap**.
- Un catalogo locale e offline-ready di vulnerabilità (**CVE/CPE/CVSS**) con correlazione automatizzata.
- Un modulo di **High-Speed Network Discovery** (stile Advanced IP Scanner) per il censimento istantaneo dei dispositivi LAN, hardware vendor e latenza.
- Un'organizzazione gerarchica ad albero (**Nested Tree Explorer**) per asset e campagne di scansione.
- Un modulo di **Remote SSH Host & iptables/nftables Firewall Management** con safety rollback timer e terminale Web SSH integrato (xterm.js).
- Un'architettura ibrida **Multi-Tenant / Single-Tenant** idonea sia per erogazione SaaS (MSSP) sia per installazione On-Premise (Enterprise isolate).
- Un hub centralizzato di **Reporting White-Label** con personalizzazione grafica e loghi per ogni tenant (Executive & Technical reports).

---

## 2. Perimetro del Progetto (Scope Boundaries)

### 2.1 In-Scope (Milestone 1 & 1.1)
- **Asset Scope Management & Tree Explorer:** Gestione target IP/CIDR/domini con struttura gerarchica ad albero (cartelle/sotto-cartelle), blacklist e metriche di rischio aggregate.
- **Scanning Engine & Alive-Only Host Filter:** Esecuzione concorrente e sicura di Nmap in Docker con profili (Fast, Standard, Full, LAN Discovery) ed esclusione nativa degli host spenti (`down`) nei range IP.
- **Job Orchestrator:** Supporto a scansioni On-Demand e Schedulate (cron) a livello di target singolo o intera cartella dell'albero.
- **CVE Database & Correlator:** Import e indicizzazione locale del feed CVE/CPE con arricchimento CVSS v3.1 e matching sui banner rilevati.
- **LAN Device Discovery (Advanced IP Scanner mode):** Schermata dedicata `/network-discovery` ad altissima velocità (Ping Sweep / ARP Scan) con identificazione OUI MAC Vendor, hostname e import rapido nei Target.
- **Remote SSH & Firewall GUI (iptables):** Gestione credenziali SSH (chiavi private o password cifrate AES-256-GCM), tabella visuale delle regole firewall (`INPUT`, `OUTPUT`, `FORWARD`), timer di anti-lockout protection e terminale Web SSH interattivo (WebSocket + xterm.js).
- **Centralized Reports Hub:** Pagina dedicata `/reports` con archivio storico, download diretto PDF White-Label e CSV, e doppio layout (Executive Summary vs Technical Deep-Dive).
- **Multi-Tenant Hybrid RBAC:** Isolamento logico dei dati per `tenant_id` e ruoli (SuperAdmin, Tenant Admin, SecOps, Auditor).

### 2.2 Out-of-Scope (Riservato a Milestone 2+)
- Generazione del grafo di attacco e percorsi di lateral movement (Milestone 2).
- Dashboard di Continuous Compliance NIS2 / DORA / ISO 27001 (Milestone 2).
- Moduli EASM, CSPM e Identity Leak Monitoring (Roadmap futura).
- Scansioni autenticate (credentialed) con agenti locali su endpoint (Roadmap futura).

---

## 3. Requisiti Non Funzionali (NFR) Dettagliati

### 3.1 Sicurezza Applicativa e Infrastrutturale (Security by Design)
- **NFR-SEC-01 (Prevenzione Command Injection):** L'invocazione di `nmap` da Go non deve **mai** passare per una shell di sistema. Deve utilizzare rigorosamente `exec.CommandContext` passando argomenti in array validati e igienizzati contro flag injection.
- **NFR-SEC-02 (Principio del Minimo Privilegio nel Container):** Il container Go/Nmap non deve girare come root indiscriminato. Linux Capabilities controllate (`CAP_NET_RAW`, `CAP_NET_ADMIN`, `CAP_NET_BIND_SERVICE`).
- **NFR-SEC-03 (Isolamento Multi-Tenant Rigoroso):** Ogni query al database deve incorporare esplicitamente il `tenant_id` derivato dal contesto di autenticazione JWT, prevenendo qualsiasi accesso cross-tenant.
- **NFR-SEC-04 (Autenticazione & Hashing):** Token di accesso stateless JWT a breve scadenza (15-30 min) associati a refresh token revocabili. Hashing delle credenziali utente tramite **Argon2id** (o **Bcrypt** con work factor $\ge 12$).
- **NFR-SEC-05 (Protezione dei Dati a Riposo e in Transito):** TLS 1.3 per il traffico Web/API. Cifratura a riposo PostgreSQL.
- **NFR-SEC-06 (Cifratura Credenziali SSH a Riposo):** Le credenziali dei server remoti (password e passphrase chiavi SSH) devono essere memorizzate cifrate nel database con algoritmo **AES-256-GCM** utilizzando una master key derivata per tenant.
- **NFR-SEC-07 (Isolamento Sessioni WebSocket SSH):** Le sessioni di terminale Web SSH devono essere autenticate via JWT, protette da timeout di inattività e terminate immediatamente alla disconnessione del client.

### 3.2 Performance e Concorrenza
- **NFR-PERF-01 (Motore Asincrono Concorrente):** Le scansioni di rete non bloccano le chiamate API HTTP; gestite da worker pool concorrente con limite di concorrenza.
- **NFR-PERF-02 (Throughput di Parsing):** Parsing XML Nmap a streaming (`encoding/xml` decoder) per gestire scansioni su ampi range senza picchi di memoria.
- **NFR-PERF-03 (Indicizzazione Database):** Tabelle con indici compositi B-Tree su `tenant_id`, `group_id`, `scan_id`, `cve_id` e indici GIN sui campi `JSONB`.
- **NFR-PERF-04 (Throughput LAN Discovery):** La scansione rapida di una classe C `/24` nella modalità Network Discovery deve completarsi entro 5-10 secondi grazie a probe ICMP/ARP paralleli.

### 3.3 Affidabilità, Resilienza e Graceful Shutdown
- **NFR-REL-01 (Graceful Cancellation):** Vincolo al `context.Context` con propagazione di SIGTERM ai processi attivi.
- **NFR-REL-02 (Transazionalità e Coerenza):** Registrazione atomica dei risultati e delle transizioni di stato su PostgreSQL.
- **NFR-REL-03 (Offline-Readiness):** Operatività completa (discovery, fingerprinting, correlazione CVE e report PDF) anche in assenza di connessione Internet.
- **NFR-REL-04 (Anti-Lockout Firewall Resilience):** Modifiche a `iptables` supportate da **Safety Rollback Timer**: se l'utente non conferma la sessione entro il timeout prestabilito (es. 60s), il backend esegue automaticamente il ripristino delle regole precedenti salvate in backup.

### 3.4 Scalabilità e Manutenibilità
- **NFR-MAINT-01 (Clean Architecture):** Backend strutturato secondo la Clean Architecture (Domain, UseCases, Adapters, Delivery HTTP/WS).
- **NFR-MAINT-02 (Continuous Integration & Testability):** Copertura con Unit Test e Integration Test con mock.
- **NFR-MAINT-03 (Structured Logging):** Utilizzo di `log/slog` nativo di Go con output JSON strutturato.
