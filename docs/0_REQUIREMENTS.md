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
- **Milestone 2 (Differentiators):**
  - Attack Path Mapping (Grafo delle Vulnerabilità e lateral movement).
  - Continuous Compliance Score (NIS2, DORA, ISO 27001).

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

---

## 5. Requisiti Non Funzionali (NFR) Aggiuntivi

- **NFR-SEC-06 (Cifratura Credenziali SSH a Riposo):** Le password e le passphrase delle chiavi SSH devono essere cifrate in PostgreSQL con algoritmo crittografico autenticato AES-256-GCM. Le chiavi di decifratura sono gestite dal backend in memoria e non esposte alle API.
- **NFR-SEC-07 (Isolamento Sessioni WebSocket SSH):** Le connessioni WebSocket per il terminale remoto devono essere autenticate tramite JWT, vincolate al `tenant_id` dell'utente e terminate tempestivamente in caso di inattività o chiusura finestra.
- **NFR-REL-04 (Anti-Lockout Firewall Resilience):** Qualsiasi modifica remota alle regole del firewall deve essere atomica e reversibile. In caso di mancata conferma entro il timeout di sicurezza (Safety Rollback Timer), il backend deve eseguire il ripristino automatico dello script delle regole precedente.
- **NFR-PERF-04 (Discovery LAN Throughput):** La scansione LAN di discovery su una classe C (/24) deve completarsi entro 5-10 secondi grazie a probe paralleli ARP/ICMP senza esecuzione di port-scan profondo non richiesto.
- **NFR-PERF-05 (Real-Time SSE Event Latency):** Latenza di notifica push SSE inferiore a 200ms dal momento dell'elaborazione nel worker del backend alla ricezione nel client React.
- **NFR-REL-05 (Stream Resilience & Reconnection):** Meccanismo di auto-reconnect nativo del protocollo SSE in caso di temporanea instabilità di rete con gestione idempotente degli eventi già ricevuti.
- **NFR-DEV-01 (Zero-Overhead Debugging & Privacy):** Il logger client-side deve conservare al massimo gli ultimi 500 eventi in memoria circolare senza causare degrado di prestazioni o memory leak. Le password e le chiavi segrete inviate nei payload devono essere automaticamente oscurate (`[REDACTED]`) prima di essere visualizzate o copiate.

---

## 6. Decisioni Consolidate
- **Approvazione:** Requisiti e casi d'uso aggiornati e approvati in data 2026-10-09 (Feature 6: Real-Time Live Streaming via SSE & Feature 7: Debug & Testing Live Console Drawer).
- **Prossima Fase:** Fase 2 (System Design & Architecture Update) su `1_PROJECT_CONTEXT.md`, `2_ARCHITECTURE.md` e `3_TASK_LIST.md` tramite comando `/architect`.
