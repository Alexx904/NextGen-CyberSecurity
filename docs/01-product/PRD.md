# Product Requirements Document

> Responsabile: Business Analyst; decisioni di prodotto al PO. Migrazione documentale: 2026-10-15. Gli stati funzionali restano quelli di [PROJECT_STATE](../PROJECT_STATE.md).

## 1. Visione del Prodotto e Strategia di Release

### 1.1 Visione Complessiva
Piattaforma di CyberSecurity moderna e modulare concepita per trasformare il vulnerability scanning, la ricognizione di rete e la gestione difensiva dei server da attività frammentate a una suite centralizzata multi-tenant per SecOps e Sysadmin.

### 1.2 Strategia di Rilascio Incrementale (Phasing)
- **Milestone 1 (MVP Consolidato):**
  - Scansione Black-Box (Host Discovery, Port Scanning, Service Fingerprinting con Nmap).
  - Correlazione CVE/CPE offline-ready e calcolo CVSS.
  - Multi-Tenancy ibrido (SaaS MSSP o On-Premise Enterprise).
  - Orchestrazione scansioni On-Demand e Schedulate.
- **Milestone 1.1 (Operational & Defense Expansion - Nuove Feature Approvate):**
  - **Feature 1 (Alive-Only Host Filter):** Esclusione totale degli host inattivi/spenti nei range IP scansionati, con persistenza e reportistica pulita.
  - **Feature 2 (Centralized Reports Hub):** Hub centralizzato di reportistica White-Label (`/reports`) con storico globale, filtri multi-parametro e doppio template (Executive e Tecnico).
  - **Feature 3 (Hierarchical Tree Asset & Scan Folders):** Organizzazione gerarchica ad albero (cartelle e sotto-cartelle annidate) per Target e Scansioni.
  - **Feature 4 (High-Speed LAN Network Discovery):** Modulo stile *Advanced IP Scanner* (pagina dedicata + profilo di scansione) con Ping Sweep / ARP Scan, lookup IEEE OUI del produttore hardware, hostname NetBIOS/mDNS e azioni rapide.
  - **Feature 5 (Remote SSH Host & iptables Firewall Manager):** Gestione credenziali SSH (chiavi private o password cifrate AES-256-GCM), interfaccia grafica visuale per regole firewall (`iptables`/`nftables`) con timer di anti-lockout e terminale Web SSH interattivo integrato (WebSocket + xterm.js).
  - **Feature 6 (Real-Time Live Streaming Engine via SSE):** Flusso dati in tempo reale via Server-Sent Events per LAN Network Discovery e Scansioni Nmap. I dispositivi LAN compaiono all'istante all'interno della griglia man mano che rispondono, e le scansioni trasmettono fasi, percentuali di avanzamento e finding progressivi con salvataggio incrementale su PostgreSQL.
  - **Feature 7 (Debug & Testing Live Console Drawer):** Finestra laterale (slide-over panel) dedicata al monitoraggio in tempo reale di tutte le attività (chiamate HTTP REST con endpoint, status, latenza e payload, eventi SSE ricevuti, stream WebSocket, errori/warning di runtime JavaScript). Filtro rapido degli errori, visualizzazione del dettaglio evento e pulsante "Copia Report Diagnostico per AI" preformattato in Markdown per consentire all'utente di inviare istantaneamente stack trace e contesti operativi all'agente per l'immediata risoluzione dei bug.
- **Milestone 2 (Differentiators - Completata):**
  - Attack Path Mapping (Grafo delle Vulnerabilità, lateral movement e nodi chokepoint).
  - Gestione Aziende, Reparti & Scope Globale (Epic 27 - Multi-Company Hierarchical Workstation).
- **Milestone 3 (Active Defense & SOAR Light - In Pianificazione):**
  - One-Click Chokepoint Remediation & Firewall Isolation (integrazione diretta Attack Path con modulo SSH/iptables).
  - Virtual Patching & WAF Rule Generator (regole automatiche Fail2ban, CrowdSec, ModSecurity, Nginx basate sulle CVE scoperte).
  - Multi-Channel Alerting & Webhooks Hub (notifiche istantanee Slack, Microsoft Teams, Telegram, Webhooks REST generici ed Email SMTP).
- **Milestone 4 (Continuous Compliance & Regulatory Frameworks - In Pianificazione):**
  - Automated Regulatory Mapping Engine (correlazione algoritmica di findings, posture e firewall con NIS2 Art. 21, DORA e ISO/IEC 27001:2022).
  - Compliance Scorecard & Department Gap Analysis (cruscotto con percentuali di adempimento, gap critici e tracciamento SLA MTTR).
  - Audit-Ready PDF Dossier (dossier formale A4 per revisori, auditor esterni e organi di vigilanza).
- **Milestone 5 (AI Remediation Copilot & Conversational Security Assistant - In Pianificazione):**
  - AI Remediation Playbooks (script di remediation contestualizzati per OS, container e versioni software rilevate).
  - AI Executive Summaries nei Report (sintesi narrativa chiara e discorsiva orientata al board e al C-level).
  - Interfaccia conversazionale "Ask Infrastructure" (interrogazione in linguaggio naturale dello stato di sicurezza e inventario del tenant).

---

## 2. Modello di Tenancy e Attori del Sistema

### 2.1 Architettura Multi-Tenant Ibrida
- **Isolamento Logico Rigoroso:** Tutte le entità (`Target`, `TargetGroup`, `Scan`, `Finding`, `ReportTemplate`, `SSHServer`, `FirewallRule`) sono vincolate al `tenant_id`.
- **Flessibilità di Deployment:** SaaS multi-tenant o Single-tenant on-premise.

### 2.2 Ruoli Utente (RBAC)
| Ruolo | Descrizione | Permessi Principali |
| :--- | :--- | :--- |
| **SuperAdmin (Platform / MSSP)** | Amministratore globale della piattaforma | Creazione e gestione tenant, configurazione parametri di sistema, monitoraggio globale |
| **Tenant Admin (Org Manager)** | Responsabile organizzazione cliente | Gestione utenti del tenant, impostazioni di white-labeling (logo e colori), gestione server SSH e policy firewall |
| **SecOps / Security Engineer** | Operatore tecnico / Sysadmin | Definizione target in cartelle ad albero, discovery LAN, scansioni Nmap, gestione firewall iptables e accesso terminale SSH |
| **Auditor / Read-Only** | Stakeholder / Management / Auditor | Accesso alla dashboard, consultazione findings, visualizzazione inventario e download report approvati |

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

## Indicatori di successo

I target tecnici esistenti sono in [REQUIREMENTS.md](REQUIREMENTS.md). KPI di adozione, soddisfazione e risultati commerciali non sono ancora concordati con il PO: non attribuire valori inventati. Le milestone pianificate non sono funzionalità rilasciate.
