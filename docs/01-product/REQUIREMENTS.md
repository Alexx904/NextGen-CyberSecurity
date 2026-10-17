# Functional and Non-Functional Requirements

> Responsabile: Business Analyst. Migrazione documentale: 2026-10-15. Gli stati funzionali restano quelli di [PROJECT_STATE](../PROJECT_STATE.md).

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
- **FR-09.1:** Nuove gerarchie limitate ad Azienda → Reparto (`parent_id UUID`): vietati reparti dentro reparti in UI, service e DB. Nessuna riscrittura automatica di strutture storiche.
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
  - SSH-HARDEN-01 (2026-10-17, working tree): pin SHA256 della chiave host separato dalla chiave client, verificato su canale fidato e configurato esplicitamente. Pin assente/discordante blocca tutte le connessioni, inclusi terminale e remediation. Rotazione esplicita; cambio endpoint richiede nuova autorizzazione del pin. Nessun enrollment automatico al primo accesso. Handshake massimo 10s, comandi/test connessione massimo 30s con cancellazione; queste scadenze non garantiscono la terminazione del processo remoto.
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
- **FR-14.1 (Modello a Grafo Topologico):** Modellazione dei target, servizi e CVE come grafo orientato. Distinguere associazioni osservate da accesso e sfruttabilità ipotetici; non assumere esposizione Internet, RCE o trust dall'inventario SSH.
- **FR-14.2 (Ipotesi di Percorso):** Percorsi strutturalmente validi, con archi orientati per ogni passaggio. Le ipotesi laterali /24 devono dichiarare che routing, maschera reale, ACL, controllo della sorgente e credenziali non sono verificati; nessuna prova di lateral movement dichiarata.
- **FR-14.3 (Indice di Priorità):** Indice euristico 0–100 derivato da CVSS e passaggi logici, senza presentarlo come probabilità o rischio economico. Esporre limiti di inventario, cap dei percorsi e provenienza/storicità dei finding.
- **FR-14.4 (Chokepoint nel Modello):** Classificare i nodi intermedi per numero e indice aggregato dei soli percorsi rappresentati. Raccomandazioni da verificare con impatto operativo, senza garanzie di neutralizzazione o ROI misurato.
- **FR-14.5 (Interfaccia Grafica Interattiva):** `/attack-path` con vista percorso 2D leggibile, mappa completa terminante su cicli, zoom/pan, selezione tastiera, filtri CVSS e guida integrata. Scope aziendale applicato anche a grafo, KPI e chokepoint, senza fallback globale su scope vuoto. 3D facoltativo subordinato alla comprensione.

### FR-15: Workflow-Categorized Navigation & Hierarchical Multi-Company IP Views (Workstation Multi-Company Workflow)

- **COMPANY-UX-02:** Dashboard e CRUD aziende in `/`, con redirect da `/companies`. Scope globale permette solo panoramica per le feature aziendali, con selezione obbligatoria prima delle route operative. Eccezione: SSH/Firewall sempre accessibile, clienti autonomi per server. Reparti per target, scansioni separate, report e Attack Path; niente seconda dashboard reparto. Nessun filtro CVSS nella UI Attack Path.
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

### FR-16: Profili tecnici ed editor Nmap
- **FR-16.1 (Profili veritieri):** Due modalità, Profili e Comando Nmap, senza compositore multi-obiettivo duplicato. Nomi espliciti per porte/protocolli, versioni, NSE e filtraggio ACK; nessuna promessa di penetration test, invisibilità, bypass o conformità. Categorie e ID storici restano compatibili nelle API. Dettaglio nel [dominio scanner](../domains/scanning_and_discovery.md).
- **FR-16.2 (Editor professionale):** Opzioni modificabili direttamente, prefisso nmap facoltativo; porte, timing T0–T5, versioni, OS e sonde entro l'allowlist server. Target separati e output XML gestito. -Pn salta discovery, non aggira firewall. Una sola tecnica TCP e un solo selettore porte; TCP+UDP con lista unica.
- **FR-16.3 (Anteprima e validazione):** Anteprima delle opzioni realmente inviate, nessun matching per nome/categoria. Controllo client preliminare e validazione backend autorevole di sintassi, combinazioni e intervalli. Nessun badge anticipato «sicuro»; rischio e durata dipendono da rete e target.
- **FR-16.4 (Gestione e Salvataggio Profili Custom):** Possibilità di salvare una configurazione personalizzata come nuovo profilo riutilizzabile e schedulabile dal tenant, con persistenza su database (`POST /api/v1/scan-profiles`, `DELETE /api/v1/scan-profiles/{id}`).
- **FR-16.5 (Tempi e progresso del gruppo, SCAN-TIME-01):** Ogni nuovo preset espone una durata indicativa per IP con ipotesi esplicite e timeout distinto. Dopo il lancio del gruppo la pagina Scansioni mostra progresso complessivo del lotto, job in coda/in esecuzione/riusciti/falliti/annullati ed ETA approssimativa quando le statistiche Nmap sono sufficienti. Lotto identificato dagli ID restituiti, non dallo storico. Telemetria assente/scaduta non deve produrre un countdown fittizio; 100% significa tutti i job terminati, anche con errori. Valori orientativi, formula e limiti nel dominio scanner.

### FR-17: Ricerca Vulnerabilità Online (NVD API 2.0 Cloud Lookup) & Script Vulnerabilità Nmap (--script vuln)
- **FR-17.1 (Ricerca CVE Online su Feed Ufficiale NIST NVD):** In aggiunta al catalogo locale, il motore di correlazione deve poter interrogare in tempo reale le API REST pubbliche ufficiali (NIST NVD API 2.0) tramite CPE 2.3 URI o keyword di prodotto/versione per identificare CVE note appena pubblicate.
- **FR-17.2 (Auto-Popolamento e Caching su PostgreSQL):** Ogni CVE recuperata tramite ricerca online deve essere automaticamente memorizzata nella tabella `cve_entries` di PostgreSQL (upsert atomico con indici CVSS), arricchendo in modo incrementale il database locale e garantendo velocità istantanea e resilienza offline per le scansioni successive.
- **FR-17.3 (Script Engine Nmap):** Profilo Controlli vulnerabilità NSE ed editor supportano `--script=vuln` entro l'allowlist. La categoria può includere script intrusivi, richiede conferma UI e classificazione persistita dal server; copertura dipendente dagli script installati, esiti da verificare professionalmente. Non equivale a exploitation completa o penetration test.
- **FR-17.4 (Parsing XML & Correlazione Risultati Script NSE):** Il parser Nmap e l'orchestratore devono decodificare i tag `<script>` associati a porte e host, estraendo automaticamente CVE ID menzionati nell'output, severità, banner e raccomandazioni di remediation da salvare come finding formali.

### FR-18: Qualità Grafica, Layout A4 & Coerenza Dati nei Report PDF (Executive & Technical)

- **COMPANY-UX-02:** Dossier aziendale include scansioni e target dei reparti, con inventario IP/target separato per reparto, anche senza finding. PDF consolidato multipagina ammesso, download reparto conservato. CSV consolidato aggiunge reparto; CSV scansione invariato.
- **COMPANY-PDF-03:** Nei PDF dell'intera azienda, se esistono reparti, anche le schede CVE/finding sono separate per reparto. Tecnico completo; Executive con cinque priorità per reparto e conteggio totale. Riepilogo aziendale consolidato, sezione “Senza reparto” per target diretti e indicazione esplicita dei reparti senza finding disponibili. Aziende senza reparti mantengono il formato non raggruppato.
- **FR-18.1 (Executive Summary Compatto A4):** Il report direzionale deve essere strutturato come un Executive Brief elegante di esattamente 2 pagine (pagina 1: Copertina, Sintesi Esecutiva, Statistiche KPI, Distribuzione Severità; pagina 2: Top Priority Findings con remediation dettagliata e Governance/Prossimi Passi), eliminando salti pagina artificiali che lasciano pagine semivuote.
- **FR-18.2 (Zero Testo Tagliato & Formattazione Multiriga Card):** Le card "Top Priorities" e le schede dei finding tecnici devono calcolare dinamicamente l'altezza in base al testo di remediation tramite wrapping multiriga, senza overflow orizzontale oltre i margini della pagina A4 e senza font miniaturizzati illegibili.
- **FR-18.3 (Separazione Righe CVE & Azione):** L'identificativo CVE e la label "Azione:" devono essere posizionati su righe distinte con spaziatura e gutter calibrati, eliminando sovrapposizioni di testo.
- **FR-18.4 (Colori Semantici & Metriche Coerenti):** Le KPI devono riflettere semanticamente il dato rappresentato (es. MAX CVSS colorato in base alla severità effettiva: arancione per High 7.0-8.9, rosso per Critical 9.0+, giallo per Medium; Host Attivi con colore neutro ardesia scuro).
- **FR-18.5 (Coerenza Punteggi Posture & Testo Qualitativo):** L'algoritmo di Posture Score deve normalizzare le penalità su perimetri multi-host per prevenire il collasso sistematico a 0/100 Grade F su reti ampie e garantire coerenza tra punteggio numerico, grado e testo esplicativo (nessuna menzione di "CVSS >= 9.0" se non ci sono vulnerabilità critiche).
- **FR-18.6 (Nomi Software Reali & Protocolli Insicuri):** Nei titoli e nelle raccomandazioni deve essere utilizzato il nome reale del prodotto/software ricavato dai banner (`dnsmasq`, `vsftpd`, `Dropbear`) anziché la denominazione generica della porta di rete (`domain`, `ftp`, `ssh`). Servizi storici e protocolli non cifrati attivi (Telnet, FTP, SMB, NetBIOS) devono ricevere opportuna evidenza con severità calcolata e remediation prescrittiva.
- **FR-18.7 (Impaginazione Fluida Report Tecnico):** Le schede dei finding nel report tecnico devono fluire verticalmente sfruttando l'intera pagina fino al footer (margine dinamico a 270mm anziché interruzioni premature a 225mm), eliminando vuoti anomali di 80-150px a fondo pagina.
- **FR-18.8 (Dati Temporali Sempre Valorizzati):** "Data Avvio" e "Durata Totale" devono essere sempre calcolati e visualizzati, aggregando le durate effettive degli audit ed eliminando diciture "N/A".

### FR-19: Active Defense, Chokepoint Auto-Remediation & SOAR Light (Milestone 3)

Perimetro E28-S1-T1 concordato col PO il 2026-10-17: blocco porta sul server del
chokepoint; isolamento host da definire separatamente. Primo incremento IPv4,
servizi TCP/UDP e host SSH registrato con IP coincidente; porta SSH esclusa.
Contratto e limiti nel [dominio Active Defense](../domains/active_defense_and_soar.md).
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
- **Riferimenti correnti:** contesto in `docs/01-product/PRD.md`, architettura in `docs/03-architecture/HLD.md` e pianificazione epiche in `docs/04-development/TASK_LIST.md`.
