# 1. Project Context & Non-Functional Requirements (NFR)

> **Stato:** APPROVATO (Fase 2: System Design - Revisione Milestone 1.1: Real-Time Streaming)  
> **Ultimo Aggiornamento:** 2026-10-09  
> **Autore / Ruolo:** TechLead-Engineer-Agent (System Architect & Tech Lead)  
> **Documento Correlato:** [0_REQUIREMENTS.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/0_REQUIREMENTS.md)

---

## 1. Executive Summary & Obiettivi del Progetto

**NextGen CyberSecurity Platform** è una soluzione all-in-one per la gestione del rischio e della vulnerabilità di rete, la ricognizione ad alta velocità e l'amministrazione difensiva dei server, ingegnerizzata secondo moderni standard di sicurezza, modularità e alte prestazioni.

Il sistema supera il tradizionale paradigma del "report statico di vulnerabilità" fornendo:
- Un motore di scansione di rete black-box altamente performante ed elastico basato su **Go** e **Nmap**.
- Un catalogo locale e offline-ready di vulnerabilità (**CVE/CPE/CVSS**) con correlazione automatizzata.
- Un motore di **Real-Time Live Streaming (Server-Sent Events)** che proietta all'istante sullo schermo i dispositivi rilevati e le evidenze di sicurezza senza attendere la conclusione delle scansioni.
- Un modulo di **High-Speed Network Discovery** (stile Advanced IP Scanner) per il censimento istantaneo dei dispositivi LAN, hardware vendor e latenza.
- Un'organizzazione gerarchica ad albero (**Nested Tree Explorer**) per asset e campagne di scansione.
- Un modulo di **Remote SSH Host & iptables/nftables Firewall Management** con safety rollback timer e terminale Web SSH integrato (xterm.js).
- Una **Live Debug & Testing Console (Slide-over Drawer)** per l'intercettazione e la diagnosi in tempo reale di tutto il traffico API REST, streaming SSE, WebSocket ed errori JavaScript con esportazione Markdown per fixing assistito.
- Un'architettura ibrida **Multi-Tenant / Single-Tenant** idonea sia per erogazione SaaS (MSSP) sia per installazione On-Premise (Enterprise isolate).
- Un hub centralizzato di **Reporting White-Label** con personalizzazione grafica e loghi per ogni tenant (Executive & Technical reports).

---

## 2. Perimetro del Progetto (Scope Boundaries)

### 2.1 In-Scope (Milestone 1, 1.1, 2 Completate & Milestone 3, 4, 5 Approvate)
- **Asset Scope Management, CompanyContext & Tree Explorer:** Gestione target IP/CIDR/domini con struttura gerarchica ad albero e scope globale per azienda cliente (Epic 27).
- **Scanning Engine & Alive-Only Host Filter:** Esecuzione concorrente e sicura di Nmap in Docker con profili specialistici ed esclusione nativa degli host spenti (`down`) nei range IP.
- **Real-Time Live Streaming Engine (SSE):** Streaming istantaneo Server-Sent Events per LAN Network Discovery e Scansioni Nmap con salvataggio incrementale su PostgreSQL.
- **Debug & Testing Live Console Drawer:** Finestra laterale (slide-over panel `Ctrl+Shift+D`) per intercettazione diagnostica (HTTP, SSE, WebSocket, JS) ed export AI.
- **Job Orchestrator:** Supporto a scansioni On-Demand e Schedulate (cron) a livello di target singolo o intera cartella dell'albero.
- **CVE Database & Correlator:** Catalogo locale offline-ready con matching CVE/CPE, arricchimento CVSS e client REST NIST NVD API 2.0.
- **LAN Device Discovery:** Modalità ad alta velocità con Ping Sweep, lookup OUI MAC Vendor e integrazione target.
- **Remote SSH & Firewall GUI (iptables):** Gestione credenziali cifrate AES-256-GCM, Anti-Lockout Safety Timer, terminale Web SSH interattivo (xterm.js).
- **Centralized Reports Hub:** Archivio storico, download PDF White-Label (Executive compresso A4 e Technical continuous flow) e CSV consolidati per azienda.
- **Multi-Tenant Hybrid RBAC:** Isolamento logico dei dati per `tenant_id` e ruoli (SuperAdmin, Tenant Admin, SecOps, Auditor).
- **Attack Path Mapping & Visual Graph Explorer (Milestone 2):** Grafo di compromissione multi-hop, inferenza percorsi di lateral movement, rilevamento nodi chokepoint ad alto ROI con prescrizione di remediation e canvas SVG interattivo con controlli zoom/pan.
- **Active Defense & SOAR Light (Milestone 3 - Approvata):** Neutralizzazione 1-click dei chokepoint con firewall rules auto-generate via SSH, generatore di regole di virtual patching (Fail2ban, CrowdSec, ModSecurity/Nginx) e hub di alerting multi-canale (Slack, Teams, Telegram, Webhook REST, SMTP).
- **Continuous Compliance NIS2 / DORA / ISO 27001 (Milestone 4 - Approvata):** Correlazione automatica di finding tecnici e regole firewall rispetto a NIS2 Art. 21, DORA e ISO 27001; cruscotto `/compliance` con scorecard, gap analysis, SLA MTTR e dossier audit-ready PDF.
- **AI Remediation Copilot & Security Assistant (Milestone 5 - Approvata):** Generazione automatica di playbook di mitigazione per OS e software target, sintesi esecutive AI nei report per board direzionali e assistente conversazionale "Ask Infrastructure".

### 2.2 Out-of-Scope (Roadmap Futura)
- Moduli EASM (External Attack Surface Management) e passive OSINT subdomain enumeration.
- Dark Web / Identity Leak Monitoring continuativo.
- Scansioni autenticate (credentialed) con agenti locali su endpoint (EDR/agent telemetry).
- Integrazione tool esterni containerizzati (Nuclei, testssl.sh, ssh-audit).

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
- **NFR-PERF-05 (Real-Time SSE Event Latency):** La latenza tra la scoperta di un host/porta/finding nel worker e la propagazione dell'evento push Server-Sent Events verso il frontend React deve essere inferiore a 200ms.

### 3.3 Affidabilità, Resilienza e Graceful Shutdown
- **NFR-REL-01 (Graceful Cancellation):** Vincolo al `context.Context` con propagazione di SIGTERM ai processi attivi.
- **NFR-REL-02 (Transazionalità e Coerenza):** Registrazione atomica dei risultati e delle transizioni di stato su PostgreSQL.
- **NFR-REL-03 (Offline-Readiness):** Operatività completa (discovery, fingerprinting, correlazione CVE e report PDF) anche in assenza di connessione Internet.
- **NFR-REL-04 (Anti-Lockout Firewall Resilience):** Modifiche a `iptables` supportate da **Safety Rollback Timer**: se l'utente non conferma la sessione entro il timeout prestabilito (es. 60s), il backend esegue automaticamente il ripristino delle regole precedenti salvate in backup.
- **NFR-REL-05 (Stream Decoupling & Reconnection Resilience):** La disconnessione o chiusura accidentale della sessione SSE dal browser non deve arrestare il job di scansione in esecuzione nel backend; alla riconnessione o refresh della pagina, lo stato parziale memorizzato su PostgreSQL viene ripristinato e lo stream si riallaccia istantaneamente.

### 3.4 Scalabilità e Manutenibilità
- **NFR-MAINT-01 (Clean Architecture):** Backend strutturato secondo la Clean Architecture (Domain, UseCases, Adapters, Delivery HTTP/WS).
- **NFR-MAINT-02 (Continuous Integration & Testability):** Copertura con Unit Test e Integration Test con mock.
- **NFR-MAINT-03 (Structured Logging):** Utilizzo di `log/slog` nativo di Go con output JSON strutturato.
- **NFR-DEV-01 (Zero-Overhead Debugging & Privacy):** Il logger client-side deve memorizzare in buffer circolare al massimo gli ultimi 500 eventi senza degradare le performance dell'applicazione né provocare memory leak. I campi sensibili (password, token, bearer, passphrase) devono essere mascherati preventivamente (`[REDACTED]`).

