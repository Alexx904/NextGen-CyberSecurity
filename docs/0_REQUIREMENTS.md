# 0. Requisiti e Casi d'Uso (Requirements & Use Cases)

> **Stato:** APPROVATO (Fase 1 Conclusa - Aggiornamento Feature 6: Real-Time Scanner Streaming)  
> **Ultimo Aggiornamento:** 2026-10-09  
> **Autore / Ruolo:** TechLead-Engineer-Agent (System Architect & Product Owner)  
> **Documento di Origine:** [_legacy/Brainstorming_CyberSec_App.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/_legacy/Brainstorming_CyberSec_App.md)

---

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

## 3. Requisiti Funzionali (Functional Requirements - FR)

### FR-01: Gestione Target e Asset Scope
- **FR-01.1:** Inserimento e validazione target: IP singoli, range IP (`10.0.0.1-10.0.0.25`), blocchi CIDR (`192.168.1.0/24`) o FQDN (`scanme.org`).
- **FR-01.2:** Assegnazione ad asset group e verifica blacklist a livello tenant.
- **FR-01.3:** Attestazione di Consenso & Autorizzazione Legale (Scan Authorization Gate): Richiesta di attestazione esplicita di autorizzazione/ownership per target pubblici e reti terze prima dell'esecuzione di scansioni attive, registrata immutabilmente nell'audit log con timestamp e user_id (conformità normativa art. 615-ter c.p. / CFAA).
- **FR-01.4:** Deny-List e Isolamento Infrastruttura Interna: Blocco proattivo di target diretti verso loopback (`127.0.0.0/8`, `::1`), link-local/cloud metadata (`169.254.169.254`), container bridge Docker e host della piattaforma (`postgres`, `backend`), con verifica e risoluzione DNS preventiva anti-rebinding.

### FR-02: Motore di Scansione Basato su Nmap Containerizzato
- **FR-02.1:** Backend Go con esecuzione sicura di Nmap in container con Linux capabilities controllate.
- **FR-02.2:** Supporto a molteplici profili: *Fast Discovery*, *Standard Service Scan*, *Full Deep Scan*, e *LAN Device Discovery*.
- **FR-02.3:** Parsing streaming XML/JSON per servizi, OS e banner.
- **FR-02.4:** Graceful cancellation via `context.Context` e segnale SIGTERM.

### FR-03: Schedulazione e Trigger delle Scansioni
- **FR-03.1:** Trigger On-Demand e Schedulato (cron a 5 campi).
- **FR-03.2:** Ciclo di vita: `QUEUED`, `RUNNING`, `COMPLETED`, `FAILED`, `CANCELLED`.

### FR-04: Database Vulnerabilità Locale & Correlazione CVE (Offline-Ready)
- **FR-04.1:** Sincronizzazione locale feed CVE/CPE in PostgreSQL e correlazione banner offline.
- **FR-04.2:** Scoring CVSS v3.1 e classificazione CRITICAL, HIGH, MEDIUM, LOW, INFO.

### FR-05: Dashboard Interattiva e Storicizzazione
- **FR-05.1:** Metriche aggregate tenant, grafici severità e storico scansioni.

### FR-06: Motore di Reporting White-Label
- **FR-06.1:** Generazione PDF e CSV con template responsive e palette del tenant.

### FR-07: Esclusione Host Inattivi ("Alive-Only") nelle Scansioni su Range
- **FR-07.1:** Nelle scansioni di range IP e subnet CIDR, il motore di parsing esclude a monte gli host con stato `down`.
- **FR-07.2:** Vengono memorizzate e mostrate esclusivamente le macchine attive (`up`), salvando un contatore sintetico per gli host inattivi scartati (`down_hosts`).
- **FR-07.3:** I report PDF e CSV contengono solo nodi attivi con evidenze o servizi identificati.
- **FR-07.4:** Toggle client-side per escludere facoltativamente anche gli host reattivi al ping ma privi di porte aperte.

### FR-08: Hub Centralizzato di Reportistica White-Label (`/reports`)
- **FR-08.1:** Schermata dedicata `/reports` che elenca tutti i report generati e le scansioni completate del tenant.
- **FR-08.2:** Filtri avanzati per intervallo date, target, cartella dell'albero, profilo di scansione e livello di severità massima.
- **FR-08.3:** Download con un click del report PDF White-Label e dell'export CSV.
- **FR-08.4:** Supporto a due template selezionabili: *Executive Summary* (sintesi manageriale, grafici, indice di rischio) e *Technical Deep-Dive* (dettaglio tecnico porte, CPE, CVE, remediation).
- **FR-08.5:** Personalizzazione avanzata in Settings: ragione sociale, disclaimer legale, firma responsabile e palette colori.

### FR-09: Struttura ad Albero (Cartelle Gerarchiche) per Target e Scansioni
- **FR-09.1:** Modello ad albero ricorsivo per i gruppi target (`parent_id UUID`), supportando cartelle e sotto-cartelle illimitate (es. *Azienda > Sede Milano > DMZ Produzione*).
- **FR-09.2:** Componente UI *Tree Explorer* interattivo con navigazione cartelle, creazione, ridenominazione, eliminazione e conteggio asset ricorsivo.
- **FR-09.3:** Visualizzazione del livello di rischio aggregato sulla cartella (badge cromatico in base alla severità massima dei target contenuti).
- **FR-09.4:** Possibilità di avviare una scansione aggregata sull'intera cartella (scansiona tutti i target nel sotto-albero).
- **FR-09.5:** Raggruppamento e filtraggio delle scansioni per cartella nella vista Scansioni.

### FR-10: Network Discovery LAN ad Alta Velocità (Stile Advanced IP Scanner)
- **FR-10.1 (Pagina Dedicata):** Schermata autonoma `/network-discovery` con input subnet/range (es. `192.168.1.0/24`) e pulsante di avvio rapido.
- **FR-10.2 (Scansione Rapida):** Meccanismo di Ping Sweep e ARP scan ad altissima velocità (`nmap -sn -PR -T4`).
- **FR-10.3 (Dati Dispositivo Rilevati):** Tabella in tempo reale contenente:
  - Stato (Online / Offline).
  - Indirizzo IP.
  - Hostname (risoluzione DNS locale, NetBIOS, mDNS).
  - Indirizzo MAC.
  - Vendor / Produttore hardware della scheda di rete (lookup su database locale IEEE OUI).
  - Tempo di risposta / latenza (RTT in millisecondi).
  - Porte standard aperte rapide opzionali (22, 80, 443, 445, 3389).
- **FR-10.4 (Azioni Rapide):**
  - Selezione multipla dispositivi con aggiunta diretta ai Target (con scelta della cartella nell'albero).
  - Lancio immediato di scansione approfondita di sicurezza Nmap sui nodi selezionati.
  - Esportazione elenco dispositivi in formato CSV / JSON.
- **FR-10.5 (Profilo di Scansione):** Integrazione del profilo `LAN Device Discovery` anche nel flusso standard delle Scansioni e nello Scheduler.

### FR-11: Gestione Remota Server SSH, Firewall GUI (`iptables`/`nftables`) & Web Terminal
- **FR-11.1 (Gestione Host SSH):** Anagrafica server remoti per tenant: nome mnemonico, IP/FQDN, porta SSH (default 22), utente.
- **FR-11.2 (Autenticazione Sicura Flessibile):**
  - Supporto per autenticazione a **Chiave Privata** (upload file `.pem`/`.id_rsa` o testo con eventuale passphrase).
  - Supporto per autenticazione a **Password**, cifrata a riposo nel database PostgreSQL con algoritmo AES-256-GCM.
  - Verifica della connettività e dei permessi di esecuzione (`sudo`/`root`).
- **FR-11.3 (Interfaccia Grafica Firewall iptables/nftables):**
  - Rilevamento stato firewall e visualizzazione delle Catene principali (`INPUT`, `OUTPUT`, `FORWARD`).
  - Visualizzazione e modifica sicura della Policy predefinita di ciascuna catena (`ACCEPT` o `DROP`).
  - Tabella visuale delle regole ordinate per catena con parametri: azione (`ACCEPT`, `DROP`, `REJECT`), protocollo (`TCP`, `UDP`, `ICMP`, `ALL`), sorgente, destinazione, porte, interfaccia e commento descrittivo.
  - Modal guidato per aggiunta e modifica regole con validazione formale prima dell'invio.
  - Riordinamento ed eliminazione regole.
  - **Safety Rollback Timer (Anti-Lockout):** Quando viene applicata una policy restrittiva sulla catena INPUT, viene avviato un countdown di sicurezza (es. 60 secondi): se l'utente non conferma la continuità della sessione, il firewall esegue il ripristino automatico allo stato precedente per prevenire l'auto-esclusione dall'accesso SSH.
  - Salvataggio e ripristino configurazioni firewall (`iptables-save`/`iptables-restore`).
- **FR-11.4 (Web SSH Terminal Integrato):**
  - Emulazione terminale completa nel browser basata su WebSocket e xterm.js.
  - Supporto a comandi interattivi a schermo intero (es. `htop`, `nano`, `top`), colori ANSI e copia/incolla.
  - Modalità visualizzazione: sia finestra a schermo intero sia finestra agganciabile in basso (docked) per affiancare il terminale alla gestione grafica del firewall.

### FR-12: Real-Time Live Streaming Engine per Scansioni e Discovery (Server-Sent Events)
- **FR-12.1 (LAN Discovery Live Stream):** L'avvio di una ricognizione LAN predispone uno stream SSE (`GET /api/v1/network-discovery/stream?session_id=...`) che trasmette istantaneamente l'evento `device_discovered` con le informazioni di ciascun host (IP, MAC, Vendor OUI, Hostname, RTT) nel momento esatto in cui risponde al probe di rete, oltre all'avanzamento percentuale in tempo reale (0-100%).
- **FR-12.2 (Vulnerability Scan Live Stream):** L'esecuzione di una scansione di vulnerabilità Nmap espone uno stream SSE (`GET /api/v1/scans/{id}/stream`) che notifica in tempo reale i cambi di fase (Host Discovery, Port Scan, Service Detection, CVE Correlation), la percentuale di avanzamento stimata, i log operativi e le porte/findings scoperte progressivamente.
- **FR-12.3 (Persistenza Incrementale su PostgreSQL):** I risultati parziali (dispositivi LAN e finding di sicurezza) vengono salvati sul database in modo incrementale durante il ciclo di vita della scansione, garantendo la conservazione e il recupero dei dati parziali anche in caso di ricaricamento della pagina (refresh) o riconnessione del client.
- **FR-12.4 (Decoupling Connessione & Resilienza):** La disconnessione della sessione SSE da parte del browser non deve interrompere l'esecuzione del job nel backend (a meno di un comando esplicito di annullamento). Alla riconnessione, il client riceve lo snapshot aggiornato e continua ad ascoltare gli eventi successivi.

### FR-13: Finestra Laterale di Debug e Monitoraggio Attività (Debug & Testing Live Console Drawer)
- **FR-13.1 (Drawer Laterale Slide-Over):** Interfaccia a scomparsa sul lato destro dello schermo (`DebugConsoleDrawer`), accessibile tramite pulsante fluttuante (icona bug/terminale) in basso a destra e scorciatoia da tastiera globale (`Ctrl + Shift + D`).
- **FR-13.2 (Client-Side Full Interceptor):**
  - **Chiamate HTTP REST:** Intercettazione trasparente di richieste e risposte Axios (metodo HTTP, URL, query params, request body, status code HTTP, durata in millisecondi, response body o errore).
  - **Eventi Push SSE:** Registrazione automatica degli eventi in arrivo dagli stream Server-Sent Events (`device_discovered`, `phase`, `progress`, `finding_discovered`, `status_change`, `error`).
  - **Frame WebSocket:** Tracciamento dei frame e dello stato di connessione del terminale remoto SSH.
  - **Errori e Warning JS:** Cattura globale degli errori non gestiti (`window.onerror`, `unhandledrejection`) e delle chiamate a `console.error` / `console.warn`.
- **FR-13.3 (Filtro e Ispezione Rapida degli Errori):**
  - Barra di filtraggio rapido per livello: `ALL`, `ERROR`, `WARN`, `INFO`.
  - Filtro per sorgente: `API/HTTP`, `SSE`, `WS`, `RUNTIME`.
  - Campo di ricerca full-text sui log con evidenziazione.
  - Vista dettagliata espandibile per ogni riga di log con JSON formattato di request/response payload e stack trace.
  - Badge numerico sul pulsante fluttuante che segnala all'istante il numero di errori accumulati nella sessione.
- **FR-13.4 (Copia Diagnostica Pre-formattata per AI & Export):**
  - Pulsante dedicato *"Copia per AI (Markdown)"*: copia negli appunti un report diagnostico completo pronto per essere incollato direttamente all'agente AI, contenente timestamp, tipo di errore, URL/endpoint, payload, codice di stato e stack trace.
  - Pulsante *"Esporta JSON"*: download del file di diagnostica strutturato completo.
  - Pulsante *"Pulisci Log"*: svuota la cronologia in memoria della sessione.

### FR-14: Motore di Attack Path Mapping & Chaining Analysis (Milestone 2)
- **FR-14.1 (Modello a Grafo Topologico):** Modellazione del perimetro e dei target scansionati come grafo orientato con nodi (Entrypoint Esterno/Internet, Subnet, Host, Servizi di rete, Vulnerabilità/CVE) ed archi tipizzati (Raggiungibilità di rete, Exploitability remota, Pivot/Lateral movement, Credenziali note).
- **FR-14.2 (Correlazione Catene di Compromissione):** Rilevamento automatico di cammini multi-hop che consentono a un attaccante esterno di penetrare nella rete interna sfruttando vulnerabilità critiche (es. RCE su DMZ, banner SSH, weak services).
- **FR-14.3 (Cumulative Path Risk Score):** Algoritmo di scoring per quantificare la probabilità e l'impatto di ciascun percorso di attacco su scala 0-100 basato su metriche CVSS v3.1, vettori di attacco (Network/Local) e profondità di hop.
- **FR-14.4 (Identificazione Chokepoints di Remediation):** Analisi algoritmica di centralità per individuare il singolo nodo o la singola porta la cui chiusura/remediation interrompe il massimo numero di percorsi di attacco, ottimizzando il ROI di sicurezza.
- **FR-14.5 (Interfaccia Grafica Interattiva):** Visualizzatore a grafo interattivo (zoom, pan, selezione nodi, filtri di rischio, evidenziazione del Critical Path) nella schermata dedicata `/attack-path`.

### FR-15: Workflow-Categorized Navigation & Hierarchical Multi-Company IP Views (Workstation Multi-Company Workflow)
- **FR-15.1 (Riorganizzazione Dashboard Laterale per Flussi di Lavoro):** La barra laterale di navigazione (Desktop e Mobile Drawer) deve essere categorizzata logicamente per macro-aree d'azione ("cosa si deve fare"):
  - **Overview / Dashboard** (sintesi generale)
  - **Scansione di Rete & Asset:** *Target & Scope Management* (albero cartelle e bersagli censiti), *Network Discovery* (Advanced IP Scanner live LAN).
  - **Scansione Vulnerabilità & Audit:** *Scansioni Nmap* (engine, esecuzioni live e cron), *Vulnerabilità & CVE* (findings e correlazione), *Attack Path Mapping* (grafo minacce e chokepoints).
  - **Infrastruttura & Accesso Remoto:** *Server SSH & Firewall GUI* (iptables/nftables e web terminal).
  - **Reportistica & Amministrazione:** *Hub Report White-Label* (audit PDF/CSV), *Impostazioni Tenant* (branding e multi-tenancy).
- **FR-15.2 (Raggruppamento per Cartella/Azienda & Dettaglio IP in Scansioni `/scans`):**
  - Vista raggruppata per Cartella Scope/Azienda (Target Groups) e nodi non assegnati.
  - Card/Accordion collassabili per ciascuna cartella con metriche riassuntive (totale scansioni, nodi UP, conteggio vulnerabilità critiche/alte, job in esecuzione).
  - Quando la cartella viene aperta, visualizzazione specifica e gerarchica di ciascun indirizzo IP / Target e dei relativi job di scansione con stati, chip severità, durata e azioni (Dettaglio, download PDF Executive/Tecnico, CSV, cancellazione SIGTERM).
  - Controlli per "Espandi tutti", "Comprimi tutti" e toggle di switch istantaneo tra vista raggruppata e tabella lineare.
- **FR-15.3 (Raggruppamento per Cartella/Azienda & Dettaglio IP in Vulnerabilità `/findings`):**
  - Riorganizzazione dell'hub vulnerabilità per consentire la distinzione immediata tra aziende/clienti diversi.
  - Selezione e visualizzazione raggruppata per Cartella Scope -> Target/IP specifico -> Findings CVE associati con punteggi CVSS v3.1, porte e raccomandazioni di remediation.
  - Possibilità di filtrare o cercare vulnerabilità all'interno della specifica azienda o per singolo IP.
- **FR-15.4 (Raggruppamento per Cartella/Azienda & Dettaglio IP in Report White-Label `/reports`):**
  - Raggruppamento dei report generati per Cartella/Azienda Scope con card riassuntiva di Security Posture (punteggio medio 0-100 e Grade A-F calcolato per cliente) e totale audit completati.
  - Apertura/espansione della cartella dell'azienda per ispezionare gli specifici report generati per ciascun IP/target, con download diretto a 1-click (Executive PDF, Tecnico PDF, CSV RFC 4180).
  - Switch rapido tra vista raggruppata per azienda e archivio cronologico globale.

---

## 4. Casi d'Uso (Use Cases - UC)

### UC-01: Configurazione Target e Policy
- **Attore:** SecOps Engineer.
- **Flusso:** L'operatore accede alla sezione Target, inserisce un IP/CIDR/range, verifica che non sia in blacklist e lo assegna a una cartella nell'albero.

### UC-02: Esecuzione Scansione On-Demand con Esclusione Host Inattivi
- **Attore:** SecOps Engineer.
- **Flusso:** L'operatore seleziona un range IP. Il backend esegue la scansione, scarta a monte gli host down, memorizza solo gli host reattivi, correla i servizi alle CVE e aggiorna la dashboard.

### UC-03: Pianificazione Ricorrente su Cartella Asset
- **Attore:** SecOps Engineer.
- **Flusso:** L'operatore programma una scansione settimanale su una cartella dell'albero di asset. Lo scheduler esegue il job ricorsivo su tutti i nodi appartenenti alla cartella.

### UC-04: Analisi Findings e Triage Vulnerabilità
- **Attore:** SecOps Engineer / Auditor.
- **Flusso:** Consultazione della matrice vulnerabilità con filtri per CVSS, visualizzazione evidenze e note di remediation.

### UC-05: Personalizzazione Branding Tenant (White-Label)
- **Attore:** Tenant Admin.
- **Flusso:** Caricamento logo, selezione colori, intestazioni e note legali in Settings con anteprima live.

### UC-06: Gestione Centralizzata Report dall'Hub `/reports`
- **Attore:** SecOps Engineer / Auditor.
- **Flusso:** L'operatore accede a `/reports`, filtra le scansioni storiche, seleziona il template desiderato (Executive o Tecnico) e scarica direttamente il file PDF o l'export CSV.

### UC-07: Organizzazione Gerarchica Asset tramite Tree Explorer
- **Attore:** SecOps Engineer.
- **Flusso:** L'operatore crea cartelle e sotto-cartelle nell'albero, organizza i target per sede geografica o perimetro logico, e monitora gli indicatori di rischio aggregati per cartella.

### UC-08: Rilevamento Rapido Dispositivi LAN (Network Discovery)
- **Attore:** SecOps Engineer / Sysadmin.
- **Flusso:** L'utente accede alla sezione Network Discovery, inserisce la subnet locale, avvia la scansione e visualizza in pochi secondi tutti i dispositivi online completi di MAC, vendor e hostname. Seleziona i dispositivi rilevati e li aggiunge a una cartella di Target con un click.

### UC-09: Gestione Regole Firewall iptables da Interfaccia Grafica
- **Attore:** SecOps / Sysadmin.
- **Flusso:** L'utente seleziona un server Linux censito, visualizza le catene iptables attive, aggiunge una regola di blocco o apertura porta via form grafico, attiva il safety timer e applica la configurazione verificando l'assenza di lockout.

### UC-10: Accesso a Terminale Web SSH Remoto
- **Attore:** Sysadmin / SecOps.
- **Flusso:** L'utente apre la sessione di terminale web direttamente dal browser per eseguire verifiche diagnostiche su un server Linux registrato con credenziali sicure.

### UC-11: Monitoraggio in Tempo Reale della Scansione e Ricognizione LAN
- **Attore:** SecOps Engineer / Sysadmin.
- **Flusso:** L'utente avvia una scansione LAN o di vulnerabilità. Senza attendere il termine del processo, visualizza in tempo reale la progressione percentuale e i singoli dispositivi o porte che compaiono istantaneamente nella griglia a schermo. Se la pagina viene ricaricata, i dati parziali già salvati su DB vengono ricaricati e il flusso di streaming si riallaccia.

### UC-12: Debugging Interattivo, Triage Errori e Feedback Loop con Agente AI
- **Attore:** Sviluppatore / Tester / Sysadmin.
- **Flusso:** Durante l'esecuzione di qualsiasi operazione nella piattaforma (scansione, discovery, firewall, login), l'utente nota un'anomalia o un badge rosso di errore. Apre il pannello laterale (click sul pulsante fluttuante o `Ctrl + Shift + D`), seleziona il filtro `ERROR`, ispeziona la richiesta fallita con status code e corpo risposta, clicca "Copia per AI (Markdown)" e incolla il blocco diagnostico nella chat con l'agente. L'agente analizza la causa radice ed emette il fix immediato.

### UC-13: Esplorazione Visiva Catene di Attacco & Remediation dei Chokepoint (Milestone 2)
- **Attore:** SecOps Engineer / CISO.
- **Flusso:** L'utente accede alla schermata `/attack-path`. Il sistema elabora l'inventario dei target attivi, le porte esposte e le vulnerabilità correlate, costruendo la mappa topologica delle minacce. L'utente visualizza visivamente i percorsi d'attacco con il cammino critico (*Critical Path*) evidenziato dal perimetro esterno agli host interni. Cliccando su un percorso o su un asset bersaglio (es. database interno o server sensibile), il pannello laterale evidenzia il *Chokepoint* principale (es. porta o servizio la cui chiusura/remediation interrompe il maggior numero di catene di compromissione), fornendo una raccomandazione di remediation a massimo impatto e ROI di sicurezza.

### UC-14: Audit di Sicurezza Presso Diverse Aziende con lo Stesso Computer (Workstation Multi-Company Workflow)
- **Attore:** Security Consultant / Auditor / Sysadmin.
- **Flusso:** L'operatore si reca presso diverse aziende o filiali con il medesimo computer portatile. Utilizza il Network Discovery per scoprire velocemente la LAN locale e importa gli host attivi in una cartella Scope dedicata al cliente (es. "Azienda Alfa", "Azienda Beta"). Nelle sezioni *Scansioni*, *Vulnerabilità* e *Report*, naviga la nuova interfaccia raggruppata per cartella: visualizza i gruppi separati con metriche chiare, apre la cartella dell'azienda desiderata e ispeziona nello specifico ciascun IP con le relative scansioni, vulnerabilità rilevate e report white-label pronti al download, mantenendo il perimetro dei diversi clienti rigorosamente isolato e visivamente ordinato.

### FR-16: Personalizzazione Avanzata Profili di Scansione & Categorie Professionali
- **FR-16.1 (Categorizzazione dei Profili di Scansione):** Il sistema deve organizzare i profili di scansione in macro-categorie specialistiche:
  1. *Penetration Testing & Probing*: Audit ad alta intensità per simulazioni di attacco, ricognizione furtiva SYN stealth (`-sS`), full port scan (65535 porte `-p-`), verifica bypass firewall (`-sA`);
  2. *Falle di Sistema & Vulnerability Assessment*: Audit approfondito del sistema operativo (`-O`), correlazione versioni dei demoni di sistema (RPC, SMB, RDP, SSH, database) e mapping CVE;
  3. *Web Application & API Services*: Audit focalizzato sullo stack web, API gateway e reverse proxy (porte 80, 443, 8080, 8443, 3000, 5000);
  4. *Network & Asset Discovery*: Ping sweep e ricognizione rapida host UP su LAN/WAN (`-sn -PE -PS...`);
  5. *Compliance & Hardening Check*: Rilevamento protocolli insicuri o non cifrati (Telnet, FTP, Rsh, VNC non protetto) e baseline di conformità.
- **FR-16.2 (Configuratore Professionale Custom Scansione):** L'utente deve poter comporre scansioni personalizzate on-the-fly definendo granularmente: range porte (Top 100/1000, Tutte, porte custom separate da virgola o range), timing template (da T1 stealth a T4 aggressive), version intensity, OS detection, solo porte aperte (`--open`), ping bypass (`-Pn`).
- **FR-16.3 (Live Command Preview Sicuro):** L'interfaccia deve mostrare l'anteprima in tempo reale del comando Nmap generato, con validazione rigorosa per prevenire flag injection (`IsSafeArgument`).
- **FR-16.4 (Gestione e Salvataggio Profili Custom):** Possibilità di salvare una configurazione personalizzata come nuovo profilo riutilizzabile e schedulabile dal tenant, con persistenza su database (`POST /api/v1/scan-profiles`, `DELETE /api/v1/scan-profiles/{id}`).

### FR-17: Ricerca Vulnerabilità Online (NVD API 2.0 Cloud Lookup) & Script Vulnerabilità Nmap (--script vuln)
- **FR-17.1 (Ricerca CVE Online su Feed Ufficiale NIST NVD):** In aggiunta al catalogo locale, il motore di correlazione deve poter interrogare in tempo reale le API REST pubbliche ufficiali (NIST NVD API 2.0) tramite CPE 2.3 URI o keyword di prodotto/versione per identificare CVE note appena pubblicate.
- **FR-17.2 (Auto-Popolamento e Caching su PostgreSQL):** Ogni CVE recuperata tramite ricerca online deve essere automaticamente memorizzata nella tabella `cve_entries` di PostgreSQL (upsert atomico con indici CVSS), arricchendo in modo incrementale il database locale e garantendo velocità istantanea e resilienza offline per le scansioni successive.
- **FR-17.3 (Script Engine Nmap per Vulnerabilità Attive `--script vuln`):** I profili specialistici di audit (es. *Falle di Sistema*, *All-in-One 360°*) e il configuratore custom devono supportare l'esecuzione controllata degli script NSE di vulnerabilità (`--script=vuln`), rilevando configurazioni deboli (SMB EternalBlue, SSL POODLE, RDP, exploit web).
- **FR-17.4 (Parsing XML & Correlazione Risultati Script NSE):** Il parser Nmap e l'orchestratore devono decodificare i tag `<script>` associati a porte e host, estraendo automaticamente CVE ID menzionati nell'output, severità, banner e raccomandazioni di remediation da salvare come finding formali.

### UC-15: Scansione Specialistica e Personalizzata di Vulnerabilità
- **Attore:** Penetration Tester / Security Engineer / IT Auditor.
- **Flusso:** L'utente apre il modale *Nuova Scansione*. Seleziona il target (o la cartella dell'azienda). Può scegliere la modalità **Semplice / Guidata** per selezionare con 1 click un profilo preimpostato categorizzato in base all'obiettivo dell'audit (Pen Test, Falle di Sistema, Web App, Discovery, Compliance) con spiegazione chiara di impatto e tempo stimato; oppure passare alla modalità **Professionale** per personalizzare porte specifiche, livello di stealth, OS detection e flag addizionali, visualizzando l'anteprima del comando e salvando facoltativamente il profilo custom. La scansione viene validata ed eseguita in modo sicuro dal backend.

### FR-18: Qualità Grafica, Layout A4 & Coerenza Dati nei Report PDF (Executive & Technical)
- **FR-18.1 (Executive Summary Compatto A4):** Il report direzionale deve essere strutturato come un Executive Brief elegante di esattamente 2 pagine (pagina 1: Copertina, Sintesi Esecutiva, Statistiche KPI, Distribuzione Severità; pagina 2: Top Priority Findings con remediation dettagliata e Governance/Prossimi Passi), eliminando salti pagina artificiali che lasciano pagine semivuote.
- **FR-18.2 (Zero Testo Tagliato & Formattazione Multiriga Card):** Le card "Top Priorities" e le schede dei finding tecnici devono calcolare dinamicamente l'altezza in base al testo di remediation tramite wrapping multiriga, senza overflow orizzontale oltre i margini della pagina A4 e senza font miniaturizzati illegibili.
- **FR-18.3 (Separazione Righe CVE & Azione):** L'identificativo CVE e la label "Azione:" devono essere posizionati su righe distinte con spaziatura e gutter calibrati, eliminando sovrapposizioni di testo.
- **FR-18.4 (Colori Semantici & Metriche Coerenti):** Le KPI devono riflettere semanticamente il dato rappresentato (es. MAX CVSS colorato in base alla severità effettiva: arancione per High 7.0-8.9, rosso per Critical 9.0+, giallo per Medium; Host Attivi con colore neutro ardesia scuro).
- **FR-18.5 (Coerenza Punteggi Posture & Testo Qualitativo):** L'algoritmo di Posture Score deve normalizzare le penalità su perimetri multi-host per prevenire il collasso sistematico a 0/100 Grade F su reti ampie e garantire coerenza tra punteggio numerico, grado e testo esplicativo (nessuna menzione di "CVSS >= 9.0" se non ci sono vulnerabilità critiche).
- **FR-18.6 (Nomi Software Reali & Protocolli Insicuri):** Nei titoli e nelle raccomandazioni deve essere utilizzato il nome reale del prodotto/software ricavato dai banner (`dnsmasq`, `vsftpd`, `Dropbear`) anziché la denominazione generica della porta di rete (`domain`, `ftp`, `ssh`). Servizi storici e protocolli non cifrati attivi (Telnet, FTP, SMB, NetBIOS) devono ricevere opportuna evidenza con severità calcolata e remediation prescrittiva.
- **FR-18.7 (Impaginazione Fluida Report Tecnico):** Le schede dei finding nel report tecnico devono fluire verticalmente sfruttando l'intera pagina fino al footer (margine dinamico a 270mm anziché interruzioni premature a 225mm), eliminando vuoti anomali di 80-150px a fondo pagina.
- **FR-18.8 (Dati Temporali Sempre Valorizzati):** "Data Avvio" e "Durata Totale" devono essere sempre calcolati e visualizzati, aggregando le durate effettive degli audit ed eliminando diciture "N/A".

### UC-17: Generazione e Presentazione Report PDF Direzionale e Tecnico per Clienti
- **Attore:** CISO / Security Consultant / Auditor.
- **Flusso:** L'operatore genera i report dall'Hub `/reports` dopo aver completato l'audit su una cartella/azienda. L'Executive Summary viene stampato o inviato in formato A4 di 2 pagine, perfettamente leggibile, con colori conformi alle direttive di sicurezza e testo di raccomandazione completo senza troncamenti. Il Technical Report presenta tutte le evidenze host per host distribuite uniformemente senza spazi vuoti e con i nomi software corretti.

### FR-19: Active Defense, Chokepoint Auto-Remediation & SOAR Light (Milestone 3)
- **FR-19.1 (Neutralizzazione Chokepoint con 1-Click):** Dalla vista del grafo topologico (`/attack-path`), selezionando un nodo o un cammino identificato come *Chokepoint ad alto ROI*, l'operatore può cliccare su "Applica Regola di Mitigazione / Isola Host". Il sistema genera automaticamente la regola di blocco `iptables`/`nftables` ottimale mirata sulla porta o sull'indirizzo sorgente/destinazione.
- **FR-19.2 (Deployment Sicuro via SSH con Rollback):** La regola viene recapitata sul server remoto associato (se censito in *Server SSH*) tramite la sessione SSH cifrata AES-256-GCM, attivando il *Safety Rollback Timer* a 60 secondi: se l'operatore non conferma il corretto funzionamento, le regole precedenti vengono ripristinate automaticamente.
- **FR-19.3 (Virtual Patching & WAF Rule Generator):** Per ogni vulnerabilità CVE rilevata sui servizi esposti (HTTP, Reverse Proxy, SSH), il sistema offre un pannello di generazione regole di patching virtuale con snippet copia-e-incolla o salvataggio file per:
  - *Fail2ban* (jail e filter regex personalizzati per bloccare tentativi di probe);
  - *CrowdSec* (scenari e parser di blocco IP abusivi);
  - *ModSecurity / Nginx* (regole di rewrite o deny-list per exploit path noti).
- **FR-19.4 (Multi-Channel Alerting & Webhooks Hub):** Modulo centralizzato di gestione canali di notifica configurabile per tenant e per azienda:
  - Canali supportati: Slack (incoming webhook), Microsoft Teams (connector), Telegram Bot, Webhook REST generico (POST con payload JSON firmato HMAC-SHA256) ed Email SMTP.
  - Eventi triggerabili: Scoperta nuova vulnerabilità CRITICAL o HIGH, rilevamento nuovo host/porta mai censito (Shadow IT), fallimento o completamento scansione programmata, scadenza certificato SSL/TLS.
  - Test live del canale con invio messaggio di prova prima del salvataggio.

### FR-20: Continuous Compliance & Regulatory Frameworks (NIS2, DORA, ISO 27001) (Milestone 4)
- **FR-20.1 (Automated Regulatory Mapping Engine):** Correlazione algoritmica automatica tra i finding di sicurezza tecnici censiti nel database e i requisiti normativi:
  - *NIS2 (Direttiva UE 2022/2555 - Art. 21):* Politiche di analisi dei rischi, gestione e divulgazione delle vulnerabilità, crittografia avanzata e sicurezza delle reti;
  - *DORA (Regolamento UE 2022/2554):* Gestione del rischio ICT per entità finanziarie, monitoraggio continuativo e test di resilienza;
  - *ISO/IEC 27001:2022 (Controlli A.8.8 Gestione vulnerabilità tecniche, A.8.20 Sicurezza di rete, A.8.24 Uso della crittografia).*
- **FR-20.2 (Compliance Scorecard & Department Gap Analysis):** Cruscotto dedicato `/compliance` integrato con il selettore aziendale globale:
  - Calcolo del *Compliance Score* globale e per singolo framework (0-100%);
  - Gap Analysis con matrice dei controlli soddisfatti, parzialmente conformi o non conformi;
  - Tracciamento SLA di remediation (MTTR - Mean Time to Remediate) delle non-conformità con alert per violazioni dei termini prescritti dalla norma.
- **FR-20.3 (Audit-Ready Dossier PDF):** Generazione ed esportazione di un dossier formale A4 per revisori ed organi di controllo contenente:
  - Copertina formale con dati legali dell'azienda, data di audit e firmatario;
  - Checklist di conformità articolo per articolo con evidenze tecniche collegate;
  - Registro storico delle contromisure applicate e piano di adeguamento residuo.

### FR-21: AI Remediation Copilot & Conversational Security Assistant (Milestone 5)
- **FR-21.1 (AI Remediation Playbooks):** Generazione dinamica di guide di mitigazione prescrittive per ogni CVE e Attack Path:
  - Riconoscimento del sistema operativo target (Debian, Ubuntu, RHEL, Alpine, Windows Server) e del server software (`nginx`, `apache2`, `openssh`, `postgresql`);
  - Generazione di comandi Bash esatti, script Ansible di aggiornamento o direttive di configurazione hardened pronte per il deploy.
  - Supporto per backend LLM flessibile: modelli locali via Ollama (offline-ready, zero data leakage) o API esterne con mascheramento preventivo degli IP privati.
- **FR-21.2 (AI Executive Summaries nei Report):** Sintesi narrativa intelligente generata dall'AI per l'Executive Report PDF: traduce l'impatto tecnico di CVE e porte aperte in scenari di rischio di business comprensibili a board direzionali e figure non tecniche.
- **FR-21.3 (Conversational Assistant "Ask Infrastructure"):** Barra di ricerca e assistente conversazionale contestuale: l'analista può porre domande in linguaggio naturale (es. *"Quali host del reparto Vendite hanno versioni vulnerabili di Apache?"*, *"Come si propaga l'attaccante dal web server al database?"*) ricevendo risposte puntuali derivate dallo stato reale dell'inventario su PostgreSQL.

### UC-18: Neutralizzazione Attiva di un Chokepoint dall'Attack Path
- **Attore:** SecOps Engineer.
- **Flusso:** L'operatore visualizza il grafo di attacco su `/attack-path`. Nota che l'host intermedio espone una porta SSH o un servizio pivot che permette all'attaccante di accedere al DB. Clicca su "Neutralizza Chokepoint": la piattaforma propone una regola `iptables -I INPUT -p tcp --dport <port> -s <untrusted> -j DROP`, si connette in SSH al server e applica la regola con timer di rollback a 60s. L'operatore convalida la connettività e conferma la regola; il grafo si ricalcola istantaneamente mostrando l'interruzione della catena.

### UC-19: Audit di Conformità NIS2 per Cliente Aziendale
- **Attore:** CISO / IT Compliance Manager.
- **Flusso:** L'operatore seleziona l'azienda cliente nel selettore globale e accede alla sezione `/compliance`. Visualizza lo stato di conformità alla NIS2 (es. 82%). Individua un gap critico sull'Art. 21 dovuto a protocolli non cifrati (Telnet/HTTP) nel Reparto Produzione. Con un click genera l'Audit-Ready PDF Dossier che documenta la gap analysis e le azioni correttive pianificate per la presentazione al comitato di audit.

### UC-20: Generazione Playbook di Remediation Assistito da AI
- **Attore:** Sysadmin / Security Analyst.
- **Flusso:** Nella schermata Vulnerabilità o Dettaglio Scansione, l'utente seleziona una CVE critica (es. RCE su OpenSSH). Clicca su "Genera Playbook AI": il copilot analizza la versione esatta del sistema operativo e fornisce la procedura di fix step-by-step con comando di patching, configurazione `sshd_config` raccomandata e test di verifica del servizio post-aggiornamento.

---

## 5. Requisiti Non Funzionali (NFR) Aggiuntivi

- **NFR-SEC-06 (Cifratura Credenziali SSH a Riposo):** Le password e le passphrase delle chiavi SSH devono essere cifrate in PostgreSQL con algoritmo crittografico autenticato AES-256-GCM. Le chiavi di decifratura sono gestite dal backend in memoria e non esposte alle API.
- **NFR-SEC-07 (Isolamento Sessioni WebSocket SSH):** Le connessioni WebSocket per il terminale remoto devono essere autenticate tramite JWT, vincolate al `tenant_id` dell'utente e terminate tempestivamente in caso di inattività o chiusura finestra.
- **NFR-REL-04 (Anti-Lockout Firewall Resilience):** Qualsiasi modifica remota alle regole del firewall deve essere atomica e reversibile. In caso di mancata conferma entro il timeout di sicurezza (Safety Rollback Timer), il backend deve eseguire il ripristino automatico dello script delle regole precedente.
- **NFR-PERF-04 (Discovery LAN Throughput):** La scansione LAN di discovery su una classe C (/24) deve completarsi entro 5-10 secondi grazie a probe paralleli ARP/ICMP senza esecuzione di port-scan profondo non richiesto.
- **NFR-PERF-05 (Real-Time SSE Event Latency):** Latenza di notifica push SSE inferiore a 200ms dal momento dell'elaborazione nel worker del backend alla ricezione nel client React.
- **NFR-REL-05 (Stream Resilience & Reconnection):** Meccanismo di auto-reconnect nativo del protocollo SSE in caso di temporanea instabilità di rete con gestione idempotente degli eventi già ricevuti.
- **NFR-DEV-01 (Zero-Overhead Debugging & Privacy):** Il logger client-side deve conservare al massimo gli ultimi 500 eventi in memoria circolare senza causare degrado di prestazioni o memory leak. Le password e le chiavi segrete inviate nei payload devono essere automaticamente oscurate (`[REDACTED]`) prima di essere visualizzate o copiate.
- **NFR-SEC-08 (Zero Data Leakage su Assistente AI):** L'invio di prompt a provider LLM esterni deve garantire l'offuscamento preventivo di indirizzi IP privati, credenziali, nomi utente e token di sessione. Qualora sia configurato un motore locale (es. Ollama), l'elaborazione deve avvenire interamente on-premise senza alcuna connessione verso l'esterno.

---

## 6. Decisioni Consolidate
- **Approvazione:** Requisiti e casi d'uso aggiornati e approvati dal Product Owner in data 2026-10-11:
  - Approvata **Milestone 3** (Active Defense & SOAR Light: Chokepoint Remediation, Virtual Patching, Alerting Hub);
  - Approvata **Milestone 4** (Continuous Compliance: NIS2, DORA, ISO 27001, Scorecard, Dossier PDF);
  - Approvata **Milestone 5** (AI Remediation Copilot: Playbooks, AI Executive Summaries, Conversational Assistant).
- **Prossima Fase:** Formalizzazione in `1_PROJECT_CONTEXT.md`, `2_ARCHITECTURE.md` e pianificazione epiche in `3_TASK_LIST.md`.

