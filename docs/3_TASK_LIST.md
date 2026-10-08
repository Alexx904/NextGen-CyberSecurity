# 3. Interactive Task List & Implementation Backlog

> **Stato:** APPROVATO (Fase 2: System Design - Backlog Attivo Milestone 1.1)  
> **Ultimo Aggiornamento:** 2026-10-08  
> **Autore / Ruolo:** TechLead-Engineer-Agent (System Architect & Tech Lead)  
> **Documenti Correlati:** [0_REQUIREMENTS_AND_USE_CASES.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/0_REQUIREMENTS_AND_USE_CASES.md) | [1_PROJECT_CONTEXT.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/1_PROJECT_CONTEXT.md) | [2_ARCHITECTURE.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/2_ARCHITECTURE.md) | [4_CHANGELOG.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/4_CHANGELOG.md)

---

## 📋 Regole di Avanzamento (Fase 3 SDLC)
1. **Micro-Planning obbligatorio (`/plan [task-id]`):** Prima di toccare qualsiasi riga di codice, deve essere formulato e approvato un piano di implementazione dettagliato con firme, file e strategie di test.
2. **Esecuzione (`/execute`):** Solo dopo l'approvazione del piano si scrive il codice di produzione con i relativi unit test.
3. **Aggiornamento di Stato:** Il task viene spuntato `- [x]` al completamento della Definition of Done (DoD).
4. **Archiviazione:** I task TASK-01 fino a TASK-34 completati nella Milestone 1 sono archiviati in [4_CHANGELOG.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/4_CHANGELOG.md).

---

## 🧹 EPIC 12: Alive-Only Host Filtering & Range Data Sanitization (Feature 1)

### Story 12.1: Filtraggio e Bonifica Host Inattivi nei Risultati di Scansione
- [ ] **TASK-35: Filtraggio Host Inattivi nel Parser XML e Scan Orchestrator (Backend Go)**
  - **File coinvolti:** `backend/internal/scanner/nmap_parser.go`, `backend/internal/service/scan_orchestrator.go`, `backend/internal/scanner/nmap_parser_test.go`
  - **Dipendenze:** TASK-14, TASK-15
  - **DoD:** Il parser XML e lo scan orchestrator escludono a monte dalla generazione dei findings tutti gli host aventi stato `down`. Solo gli host attivi (`up`) vengono elaborati e persistiti; le metriche di riepilogo mantengono il conteggio aggregato `up_hosts` e `down_hosts`. Unit test con fixture di range con host offline.

- [ ] **TASK-36: Ottimizzazione UI e Reportistica per Visualizzazione Esclusiva Host Attivi**
  - **File coinvolti:** `frontend/src/pages/ScanDetail.tsx`, `frontend/src/components/FindingsTable.tsx`, `backend/pkg/report/pdf_generator.go`, `backend/pkg/report/csv_generator.go`
  - **Dipendenze:** TASK-35
  - **DoD:** La schermata di dettaglio scansione e la tabella findings mostrano solo gli asset operativi; l'export PDF e CSV esclude categoricamente righe vuote per host non responsivi. Toggle client-side per escludere facoltativamente anche gli host privi di porte aperte.

---

## 📑 EPIC 13: Centralized Reports Hub & Dual-Template White-Label (Feature 2)

### Story 13.1: Hub Centralizzato di Reportistica e Template Executive/Tecnico
- [ ] **TASK-37: Backend Endpoint REST e Motore Multi-Template Reports Hub**
  - **File coinvolti:** `backend/internal/delivery/http/handler/report_handler.go`, `backend/internal/service/report_service.go`, `backend/pkg/report/pdf_generator.go`, `backend/cmd/server/main.go`
  - **Dipendenze:** TASK-24, TASK-25
  - **DoD:** Endpoint `GET /api/v1/reports` per consultare lo storico globale delle scansioni/report con filtri per target, date, severità e cartelle. Supporto al parametro query `template=executive` (report manageriale sintetico con grafici e posture index) e `template=technical` (dettaglio tecnico esteso).

- [ ] **TASK-38: Frontend Reports Hub Dashboard (`frontend/src/pages/Reports.tsx`)**
  - **File coinvolti:** `frontend/src/pages/Reports.tsx`, `frontend/src/api/client.ts`, `frontend/src/App.tsx`
  - **Dipendenze:** TASK-37
  - **DoD:** Schermata Reports completa che sostituisce il segnaposto con una dashboard interattiva: elenco globale dei report, filtri istantanei, selettore template (Executive o Tecnico), download immediato PDF/CSV e visualizzazione dello stato.

---

## 🌳 EPIC 14: Hierarchical Tree Asset & Scan Folders (Feature 3)

### Story 14.1: Gerarchia ad Albero per Target e Campagne di Scansione
- [ ] **TASK-39: Database Migration e Backend Repository/Service Gruppi ad Albero (`parent_id`)**
  - **File coinvolti:** `backend/internal/repository/migrations/000004_tree_target_groups.up.sql`, `backend/internal/repository/migrations/000004_tree_target_groups.down.sql`, `backend/internal/domain/target.go`, `backend/internal/repository/postgres/target_group_repo.go`, `backend/internal/service/target_group_service.go`
  - **Dipendenze:** TASK-04, TASK-11
  - **DoD:** Script di migrazione SQL che introduce `parent_id UUID REFERENCES target_groups(id)` e `color_tag`. Service con metodi per recuperare la gerarchia ad albero, calcolo ricorsivo del conteggio target e determinazione del livello di rischio aggregato. Unit test completi.

- [ ] **TASK-40: Endpoint REST Gruppi ad Albero e Supporto Scansioni su Intera Cartella**
  - **File coinvolti:** `backend/internal/delivery/http/handler/target_group_handler.go`, `backend/internal/service/scan_service.go`, `backend/cmd/server/main.go`
  - **Dipendenze:** TASK-39, TASK-17
  - **DoD:** Endpoint REST `GET /api/v1/target-groups` (albero nidificato), `POST /api/v1/target-groups`, `PUT /api/v1/target-groups/{id}`, `DELETE /api/v1/target-groups/{id}`. Supporto all'avvio di una scansione sull'intero `group_id` ricorsivo.

- [ ] **TASK-41: Frontend Tree Explorer Component & Integrazione in Targets e Scans**
  - **File coinvolti:** `frontend/src/components/TreeExplorer.tsx`, `frontend/src/pages/Targets.tsx`, `frontend/src/pages/Scans.tsx`, `frontend/src/components/TargetModal.tsx`
  - **Dipendenze:** TASK-40
  - **DoD:** Componente navigatore ad albero a sinistra con nodi espandibili, badge contatore e badge colore di rischio. Modal per creare/rinominare cartelle e possibilità di filtrare target e scansioni cliccando sui rami dell'albero.

---

## ⚡ EPIC 15: High-Speed LAN Network Discovery (Feature 4 - Advanced IP Scanner)

### Story 15.1: Ricognizione LAN Rapida con Vendor OUI e Importazione Target
- [ ] **TASK-42: Motore di Discovery LAN Veloce con Database Locale OUI Vendor (Backend)**
  - **File coinvolti:** `backend/internal/scanner/oui_database.go`, `backend/internal/service/discovery_service.go`, `backend/internal/scanner/profiles.go`, `backend/internal/domain/discovery.go`
  - **Dipendenze:** TASK-13, TASK-14
  - **DoD:** Profilo Nmap ultra-veloce (`-sn -PR -T4`). Modulo di lookup OUI integrato che traduce i primi 3 ottetti del MAC nel produttore hardware (es. Apple, Cisco, Dell, Raspberry Pi). Calcolo latenza RTT e risoluzione hostname.

- [ ] **TASK-43: Endpoint REST Network Discovery e Importazione Massiva Target**
  - **File coinvolti:** `backend/internal/delivery/http/handler/discovery_handler.go`, `backend/cmd/server/main.go`
  - **Dipendenze:** TASK-42, TASK-40
  - **DoD:** Endpoint `POST /api/v1/network-discovery/scan`, `GET /api/v1/network-discovery/results`, e `POST /api/v1/network-discovery/import-targets` (importa più dispositivi come target in una cartella specifica dell'albero).

- [ ] **TASK-44: Frontend Network Discovery Dashboard (`frontend/src/pages/NetworkDiscovery.tsx`)**
  - **File coinvolti:** `frontend/src/pages/NetworkDiscovery.tsx`, `frontend/src/App.tsx`, `frontend/src/components/layout/AppLayout.tsx`
  - **Dipendenze:** TASK-43
  - **DoD:** Nuova vista dedicata nel menu laterale stile "Advanced IP Scanner": input subnet/range, barra di progresso, tabella in tempo reale con MAC, Produttore, IP, Hostname, Latenza, selezione multipla e pulsanti "Aggiungi a Cartella Target" e "Scansiona Subito".

---

## 🛡️ EPIC 16: Remote SSH Host Management, iptables Firewall GUI & Web Terminal (Feature 5)

### Story 16.1: Infrastruttura Dati, Crittografia AES-256-GCM e Client SSH Nativo
- [ ] **TASK-45: Database Migration e Modulo Crittografico Credenziali SSH (AES-256-GCM)**
  - **File coinvolti:** `backend/internal/repository/migrations/000005_ssh_servers_firewall.up.sql`, `backend/internal/repository/migrations/000005_ssh_servers_firewall.down.sql`, `backend/pkg/crypto/aes_gcm.go`, `backend/internal/domain/ssh_server.go`, `backend/internal/repository/postgres/ssh_server_repo.go`
  - **Dipendenze:** TASK-05
  - **DoD:** Tabella `ssh_servers` e `firewall_backups`. Modulo `pkg/crypto` con crittografia autenticata AES-256-GCM per proteggere password e chiavi private a riposo. Unit test crittografici al 100%.

- [ ] **TASK-46: Client SSH Go Nativo e Servizio Gestione Firewall con Anti-Lockout Timer**
  - **File coinvolti:** `backend/internal/service/ssh_service.go`, `backend/internal/service/firewall_service.go`, `backend/internal/domain/firewall.go`
  - **Dipendenze:** TASK-45
  - **DoD:** Client SSH basato su `golang.org/x/crypto/ssh` per password e chiavi private. Servizio firewall che esegue `iptables -L -n -v --line-numbers`, parsa le regole delle catene `INPUT`, `OUTPUT`, `FORWARD`, e applica modifiche attivando un goroutine timer di safety rollback (es. 60s) con restore automatico da snapshot se non confermato.

- [ ] **TASK-47: WebSocket Gateway per Web Terminal SSH PTY Interattivo**
  - **File coinvolti:** `backend/internal/delivery/ws/ssh_terminal_ws.go`, `backend/cmd/server/main.go`
  - **Dipendenze:** TASK-46
  - **DoD:** Endpoint `/ws/v1/ssh/servers/{id}/terminal` con upgrade WebSocket e validazione JWT. Allocazione PTY remoto e bridging bidirezionale dello stream stdin/stdout/stderr con supporto al resize terminale.

- [ ] **TASK-48: Endpoint REST Gestione Server SSH e Operazioni Firewall**
  - **File coinvolti:** `backend/internal/delivery/http/handler/ssh_firewall_handler.go`, `backend/cmd/server/main.go`
  - **Dipendenze:** TASK-46
  - **DoD:** Endpoint REST `GET /api/v1/ssh/servers`, `POST /api/v1/ssh/servers`, `POST /api/v1/ssh/servers/{id}/test-connection`, `DELETE /api/v1/ssh/servers/{id}`, `GET /api/v1/ssh/servers/{id}/firewall`, `POST /api/v1/ssh/servers/{id}/firewall/rules`, `DELETE /api/v1/ssh/servers/{id}/firewall/rules/{line}`, `POST /api/v1/ssh/servers/{id}/firewall/apply`, `POST /api/v1/ssh/servers/{id}/firewall/confirm`, `POST /api/v1/ssh/servers/{id}/firewall/rollback`.

- [ ] **TASK-49: Frontend Interfaccia Grafica Firewall & Web Terminal SSH**
  - **File coinvolti:** `frontend/src/pages/SSHServers.tsx`, `frontend/src/components/FirewallManager.tsx`, `frontend/src/components/WebTerminalModal.tsx`, `frontend/src/App.tsx`, `frontend/src/components/layout/AppLayout.tsx`, `frontend/package.json`
  - **Dipendenze:** TASK-47, TASK-48
  - **DoD:** Schermata per censimento server SSH; vista grafica a schede delle catene `INPUT`/`OUTPUT`/`FORWARD` con regole, azioni, protocolli e porte; modal guidato per nuova regola; banner countdown del Safety Rollback Timer con pulsante di conferma; modal terminale interattivo a schermo intero con `@xterm/xterm`.

---

## 🧪 EPIC 17: Full-Stack Verification & Integration Test (Milestone 1.1)

### Story 17.1: Collaudo E2E delle Nuove Funzionalità
- [ ] **TASK-50: Test di Integrazione e Validazione E2E Full Stack**
  - **File coinvolti:** `docker-compose.yml`, test scripts
  - **Dipendenze:** TASK-35 fino a TASK-49
  - **DoD:** Collaudo end-to-end con:
    1. Scansione subnet con verifica esclusione host inattivi.
    2. Navigazione e download report dall'Hub `/reports` con entrambi i template.
    3. Creazione cartelle ad albero e spostamento target.
    4. Esecuzione scansione rapida LAN Discovery con rilevamento vendor e import in cartella.
    5. Connessione a server SSH di test, apertura terminale Web e modifica sicura di una regola firewall iptables con conferma del safety timer.
