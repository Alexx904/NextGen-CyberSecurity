# 0. Requisiti e Casi d'Uso (Requirements & Use Cases)

> **Stato:** APPROVATO (Fase 1 Conclusa)  
> **Ultimo Aggiornamento:** 2026-10-07  
> **Autore / Ruolo:** TechLead-Engineer-Agent (System Architect & Product Owner)  
> **Documento di Origine:** [Brainstorming_CyberSec_App.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/Brainstorming_CyberSec_App.md)

---

## 1. Visione del Prodotto e Strategia di Release

### 1.1 Visione Complessiva
Piattaforma di CyberSecurity moderna e modulare concepita per trasformare il vulnerability scanning da attività meramente tecnica a leva strategica di gestione del rischio informatico e conformità normativa.

### 1.2 Strategia di Rilascio Incrementale (Phasing)
- **Milestone 1 (MVP - In corso di specifica):** Focus esclusivo su un **Network Vulnerability Scanner** robusto, affidabile e performante:
  - Scansione Black-Box / Unauthenticated (Host Discovery, Port Scanning, Service Fingerprinting).
  - Orchestrazione nativa di **Nmap** tramite worker Go in ambiente containerizzato.
  - Sincronizzazione locale offline-ready del database delle vulnerabilità (**CVE / CPE / CVSS**).
  - Esecuzione scansioni sia **On-Demand** che **Schedulate** (ricorrenti).
  - Gestione **Multi-Tenant Ibrida** (SaaS MSSP o On-Premise Single-Tenant).
  - Reporting **White-Label** personalizzabile con stili e logo aziendale.
- **Milestone 2 (Differentiators):**
  - **Attack Path Mapping (Grafo delle Vulnerabilità):** topologia di rete e concatenazione dei vettori di attacco.
  - **Continuous Compliance Score:** mapping normativo automatico (NIS2, DORA, ISO 27001) e KPI manageriali.

---

## 2. Modello di Tenancy e Attori del Sistema

### 2.1 Architettura Multi-Tenant Ibrida
- **Isolamento Logico Rigoroso:** Ogni risorsa di dominio (`Target`, `Scan`, `Finding`, `ReportTemplate`, `User`) è associata a un `tenant_id`.
- **Flessibilità di Deployment:**
  - **Multi-Tenant (SaaS / MSSP):** Più aziende/clienti indipendenti gestiti da una singola istanza con separazione logica garantita da filtri a livello di query/database.
  - **Single-Tenant (On-Premise Enterprise):** Installazione con una singola organizzazione predefinita trasparente all'utente, senza overhead di gestione multi-cliente.

### 2.2 Ruoli Utente (RBAC)
| Ruolo | Descrizione | Permessi Principali |
| :--- | :--- | :--- |
| **SuperAdmin (Platform / MSSP)** | Amministratore globale della piattaforma | Creazione e gestione tenant, configurazione parametri di sistema, monitoraggio globale |
| **Tenant Admin (Org Manager)** | Responsabile organizzazione cliente | Gestione utenti del tenant, impostazioni di white-labeling (logo e colori del report), policy target |
| **SecOps / Security Engineer** | Operatore tecnico | Definizione target/scope, lancio scansioni on-demand, configurazione schedulazioni, analisi vulnerabilità |
| **Auditor / Read-Only** | Stakeholder / Management / Auditor | Accesso alla dashboard, consultazione findings e download dei report approvati |

---

## 3. Requisiti Funzionali (Functional Requirements - FR) — Milestone 1 (MVP)

### FR-01: Gestione Target e Asset Scope
- **FR-01.1:** Inserimento e validazione target: IP singoli (`192.168.1.50`), range IP (`10.0.0.1-10.0.0.25`), blocchi CIDR (`192.168.1.0/24`) o FQDN/domini (`scanme.org`).
- **FR-01.2:** Creazione di "Asset Groups" logici (es. "Server Web DMZ", "Infrastruttura Interna").
- **FR-01.3:** Gestione Blacklist / Exclusion List a livello di tenant (IP da non scansionare mai per policy o contratti).

### FR-02: Motore di Scansione Basato su Nmap Containerizzato
- **FR-02.1:** Backend in Go che orchestra l'esecuzione del binario `nmap` preinstallato nell'immagine Docker con privilege management controllato.
- **FR-02.2:** Supporto a molteplici **Profili di Scansione**:
  - *Fast Discovery:* Discovery host attivi e top 100 porte TCP (`-T4 -F`).
  - *Standard Service Scan:* Top 1000 porte TCP con rilevamento versioni dei servizi (`-sV -T4`).
  - *Full Deep Scan:* Tutte le porte 1-65535, OS detection e service versioning (`-p- -sV -O`).
- **FR-02.3:** Parsing ad alte prestazioni dell'output strutturato XML/JSON di Nmap per estrarre porte, protocolli, software, versioni e banner.
- **FR-02.4:** Graceful cancellation delle scansioni in corso tramite propagazione del segnale di interruzione (`SIGINT`/`SIGTERM` al processo `nmap` via `context.Context`).

### FR-03: Schedulazione e Trigger delle Scansioni
- **FR-03.1 (On-Demand):** Avvio immediato di una scansione su uno o più target selezionati.
- **FR-03.2 (Scheduled):** Schedulatore integrato per esecuzioni periodiche (giornaliere, settimanali, mensili o con espressione cron personalizzata).
- **FR-03.3:** Gestione stati del ciclo di vita: `QUEUED`, `RUNNING`, `COMPLETED`, `FAILED`, `CANCELLED`.
- **FR-03.4:** Tracciamento del tempo di esecuzione e log sintetico di scansione per troubleshooting.

### FR-04: Database Vulnerabilità Locale & Correlazione CVE (Offline-Ready)
- **FR-04.1:** Sincronizzazione programmata / on-demand del catalogo pubblico CVE e CPE all'interno di tabelle dedicate in PostgreSQL.
- **FR-04.2:** Funzionamento 100% offline del motore di matching una volta sincronizzato il database (adatto per installazioni on-premise isolate/air-gapped).
- **FR-04.3:** Algoritmo di matching tra i banner/CPE estratti da Nmap e il catalogo CVE locale.
- **FR-04.4:** Calcolo e arricchimento con CVSS Score v3.1 (Base Score, Attack Vector, Privileges Required, User Interaction, Scope, Confidentiality, Integrity, Availability) e classificazione di severità (CRITICAL, HIGH, MEDIUM, LOW, INFO).

### FR-05: Dashboard Interattiva e Storicizzazione
- **FR-05.1:** Visualizzazione web interattiva delle metriche aggregate del tenant (distribuzione per severità, host più vulnerabili, ultime scansioni).
- **FR-05.2:** Tabella interattiva dei findings con filtri avanzati (per severità, porta, IP, CVE ID, data).
- **FR-05.3:** Storico delle scansioni con possibilità di consultare l'evoluzione temporale (identificazione di nuove vulnerabilità emerse rispetto alla scansione precedente).

### FR-06: Motore di Reporting White-Label Personalizzabile
- **FR-06.1:** Generazione ed esportazione report nei formati **PDF** e **CSV**.
- **FR-06.2:** Personalizzazione White-Label:
  - Caricamento e rendering del logo aziendale del tenant.
  - Personalizzazione palette colori / tema e intestazione/piè di pagina.
  - Inserimento dati personalizzati dell'organizzazione (ragione sociale, note legali, contatti).
- **FR-06.3:** Struttura del report: Sezione Executive Summary (grafici a torta delle severità, KPI di conformità tecnica) e Sezione Technical Detail (schede dettagliate per ogni vulnerabilità con remediation raccomandata).

---

## 4. Casi d'Uso (Use Cases - UC) — Milestone 1

### UC-01: Configurazione Target e Policy
- **Attore:** SecOps Engineer.
- **Flusso:** L'operatore accede alla sezione Target, inserisce un IP/CIDR, verifica che non sia in blacklist e lo assegna a un gruppo logico del tenant.

### UC-02: Esecuzione Scansione On-Demand
- **Attore:** SecOps Engineer.
- **Flusso:** L'operatore seleziona il target, sceglie il profilo (es. Standard Service Scan) e avvia la scansione. Il backend lancia il processo Nmap, ne monitora l'avanzamento, parsa i risultati, li correla alle CVE locali e notifica il completamento sulla dashboard.

### UC-03: Configurazione Scansione Pianificata
- **Attore:** SecOps Engineer.
- **Flusso:** L'operatore pianifica una scansione settimanale su un gruppo di asset. Lo scheduler interno avvia il task al momento stabilito e registra il nuovo report storico.

### UC-04: Analisi Risultati e Triage Vulnerabilità
- **Attore:** SecOps Engineer / Auditor.
- **Flusso:** L'utente visualizza l'elenco delle vulnerabilità ordinate per CVSS Score, esamina il dettaglio di un CVE critico (vettore d'attacco, servizio impattato, remediation) e visualizza le evidenze del banner Nmap.

### UC-05: Personalizzazione Branding Tenant (White-Label)
- **Attore:** Tenant Admin.
- **Flusso:** L'amministratore dell'organizzazione carica il file immagine del logo aziendale, imposta i colori primari del report e salva le preferenze per il tenant.

### UC-06: Generazione ed Esportazione Report Personalizzato
- **Attore:** SecOps Engineer / Auditor.
- **Flusso:** L'operatore richiede l'esportazione del report PDF per una determinata scansione. Il sistema genera il documento applicando il template con il logo e i colori del tenant e fornisce il file per il download.

---

## 5. Requisiti Non Funzionali (NFR)

- **NFR-01 (Concorrenza & Robustezza Backend):** Backend scritto in Go, architettura modulare con gestione di goroutine pool per evitare starvation della CPU e della memoria durante il parsing Nmap e il matching CVE.
- **NFR-02 (Sicurezza & Privilegi Nmap):** Esecuzione di Nmap all'interno di container Docker con Linux capabilities minime necessarie (es. `CAP_NET_RAW`, `CAP_NET_ADMIN` se richiesto per SYN scan, senza concedere permessi root indiscriminati).
- **NFR-03 (Isolamento Multi-Tenant):** Tutte le entità dati sono rigorosamente vincolate a `tenant_id`. Nessuna query può bypassare il filtro del tenant.
- **NFR-04 (Operatività Offline / Air-Gapped):** Nessuna dipendenza a runtime da connessioni esterne durante le scansioni o l'arricchimento CVE; sincronizzazione del feed gestita in modo asincrono.
- **NFR-05 (Autenticazione & Autorizzazione):** Autenticazione basata su JWT con refresh token, password hashing con Argon2id / Bcrypt e autorizzazione granulare Role-Based Access Control (RBAC).

---

## 6. Decisioni Consolidate (Fase 1 Conclusa)
- **Stack Frontend Selezionato:** React + TypeScript + Vite + Tailwind CSS.
- **Database CVE:** Locale PostgreSQL (Offline-Ready / Sincronizzazione asincrona feed).
- **Scanner:** Orchestrazione Nmap in container Docker via worker Go concorrenti.
- **Reporting:** White-label personalizzabile con loghi aziendali e palette colori.
- **Approvazione:** Requisiti approvati dall'utente in data 2026-10-07. Transizione autorizzata alla Fase 2 (System Design).
