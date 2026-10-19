# Product Requirements Document

Current M4 product decision (2026-10-19): PO selected automatic technical regulatory
context in Findings and Executive/Technical PDFs, replacing manual compliance
scorecards, questionnaires, identity associations and Audit workflow. The earlier
delivery below is superseded. [Current scope](../04-development/plans/M4-AUTOMATIC-CONTEXT.md).

M4 implementation (2026-10-19 working tree) is verified and awaiting PO acceptance: evidence-based
scorecards, explicit issue identity/lifecycle/SLA and Audit A4 in the existing
assessment/reporting workflow. Curated framework coverage and unknown states do
not establish certification. [Boundaries and gates](../04-development/plans/M4-COMPLIANCE.md).

Ultimo allineamento: 2026-10-19 · working tree · ROADMAP-COHERENCE-01.

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

### 1.3 Estensione approvata 2026-10-18: validazione e penetration testing

Le nuove milestone seguono M5 senza riordinare le priorità esistenti. Sono pianificate, non implementate; l'inserimento in roadmap non autorizza installazioni, acquisti o test su sistemi reali.

- **Milestone 6 (Exploit Intelligence & Applicability):** correlare asset, servizi, versioni e CVE con riferimenti pubblici e catalogo moduli Metasploit Framework. Mostrare prerequisiti, fonti e incertezza; nessuna esecuzione di moduli o codice pubblico.
- **Milestone 7 (Vulnerability Validation):** integrare un worker Metasploit separato attraverso API RPC per controlli selezionati, con perimetro autorizzato, permessi, timeout, audit ed evidenze. Un controllo `check` non è disponibile per ogni modulo né automaticamente innocuo.
- **Milestone 8 (Controlled Penetration Testing & Retest):** prove di sfruttamento esplicitamente autorizzate e selezionate dall'operatore, gestione job/sessioni, arresto, cleanup e report con impatto osservato, remediation e retest. Nessuno sfruttamento automatico derivato dalla sola correlazione CVE.

Flusso: azienda/asset/finding → candidati e prerequisiti (M6) → verifica autorizzata ed evidenze (M7) → eventuale prova di sfruttamento autorizzata (M8) → report → correzione → retest. Esiti distinti: potenziale, validata, sfruttamento dimostrato, non confermata, inconcludente, corretta e riverificata. Fallimento della prova non equivale ad assenza di vulnerabilità; l'Attack Path resta euristico fuori dalle evidenze puntualmente raccolte.

Scope, task e gate: [TASK_LIST](../04-development/TASK_LIST.md); contratti: FR-22..FR-24 in [REQUIREMENTS](REQUIREMENTS.md). Architettura definitiva, versione/licenze delle dipendenze, API/DTO e schema con migrazioni up/down saranno definiti prima dell'implementazione. Nessuna nuova API o dipendenza è introdotta da questa pianificazione.

### 1.4 Integrated assessment roadmap (approved 2026-10-19)

M4-M8 extend the existing company assessment workflow: company → department → asset → scan/finding → evaluation → remediation → verification → report. Inputs include CVEs, exposed services, configuration observations and evidence; not every assessment item requires a CVE. These are connected capabilities, not one independent tool per milestone. SSH/Firewall remains a tenant-scoped operational tool accessible without company selection; explicit remediation links do not implicitly assign SSH records to companies.

Reuse the implemented inventory, company context, CVE/confidence/detection metadata, remediation text, heuristic Attack Path and report hub. Existing Executive/Technical PDF and CSV exports, company consolidation and department sections remain the foundation. M4 adds an Audit profile to the same pipeline; M5 enriches summaries and remediation; M6 adds applicability; M7/M8 add authorized verification evidence. No parallel finding store, report hub or automatic remediation executor is planned.

Start M4 by defining persistent issue identity across scan observations, lifecycle, ownership/deadlines, evidence and consistent report scope/cutoff. These are future contracts, not current Finding capabilities. Separate detection confidence, applicability, validation and demonstrated exploitation. Mitigation applied, correction declared, correction verified and reopening are distinct; missing observations do not prove correction.

Keep M4 → M5 → M6 → M7 → M8 delivery order. M7/M8 share authorization/job/evidence contracts with separate permissions and execution gates. Resolve relevant Attack Path scope/freshness debt before relying on it. Notifications remain deferred; M4 can display SLA breaches without claiming external delivery. Public exploits may be absent or incompatible; successful exploitation is never guaranteed. Dependency approval, a dedicated lab and per-engagement authorization precede active tests.

This section refines older milestone summaries: M3 T1/T2/T3 are delivered and accepted; the eight-CVE generator remains until planned T7 replacement with mitigation sources. T4/T5 are incomplete/deferred and T6 pending. M4-M8 remain unimplemented. Approval covers product direction and documentation, not code, installations or live tests. See [ADR-28](../03-architecture/ADR/README.md#adr-28--integrated-assessment-lifecycle-and-reporting-2026-10-19).

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

## 2. Perimetro del Progetto (Scope Boundaries)

### 2.1 In-Scope approvato per milestone

Lo scope seguente descrive l'obiettivo di prodotto. Implementazione, gate e accettazioni effettive sono in [TASK_LIST](../04-development/TASK_LIST.md); M1.1 presenta gap di chiusura nella [matrice QA](../05-qa/M11_CLOSURE_QA.md). Non dedurre completamento dall'inclusione nello scope.
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
