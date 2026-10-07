# 1. Project Context & Non-Functional Requirements (NFR)

> **Stato:** APPROVATO (Fase 2: System Design)  
> **Ultimo Aggiornamento:** 2026-10-07  
> **Autore / Ruolo:** TechLead-Engineer-Agent (System Architect & Tech Lead)  
> **Documento Correlato:** [0_REQUIREMENTS_AND_USE_CASES.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/0_REQUIREMENTS_AND_USE_CASES.md)

---

## 1. Executive Summary & Obiettivi del Progetto

**NextGen CyberSecurity Platform** è una soluzione all-in-one per la gestione del rischio e della vulnerabilità di rete, ingegnerizzata secondo moderni standard di sicurezza, modularità e alte prestazioni.

Il sistema supera il tradizionale paradigma del "report statico di vulnerabilità" fornendo:
- Un motore di scansione di rete black-box altamente performante ed elastico basato su **Go** e **Nmap**.
- Un catalogo locale e offline-ready di vulnerabilità (**CVE/CPE/CVSS**) con correlazione automatizzata.
- Un'architettura ibrida **Multi-Tenant / Single-Tenant** idonea sia per erogazione SaaS (MSSP) sia per installazione On-Premise (Enterprise isolate).
- Un motore di **Reporting White-Label** con personalizzazione grafica e loghi per ogni tenant.
- Una roadmap predisposta all'integrazione di **Attack Path Mapping** (grafi di attacco) e **Continuous Compliance Score** (NIS2, DORA, ISO 27001).

---

## 2. Perimetro del Progetto (Scope Boundaries)

### 2.1 In-Scope (Milestone 1 - MVP)
- **Asset Scope Management:** Gestione target IP/CIDR/domini con gruppi logici e blacklist.
- **Scanning Engine:** Esecuzione concorrente e sicura di Nmap in Docker con profili (Fast, Standard, Full).
- **Job Orchestrator:** Supporto a scansioni On-Demand e Schedulate (cron) con monitoraggio di stato in tempo reale e cancellazione graceful.
- **CVE Database & Correlator:** Import e indicizzazione locale del feed CVE/CPE con arricchimento CVSS v3.1 e matching sui banner rilevati.
- **Multi-Tenant Hybrid RBAC:** Isolamento logico dei dati per `tenant_id` e ruoli (SuperAdmin, Tenant Admin, SecOps, Auditor).
- **Web UI & Dashboard:** SPA React/TypeScript con dark mode, filtri avanzati, grafici e gestione asset.
- **White-Label Reporting:** Generazione PDF e CSV con personalizzazione del branding del cliente (logo, colori, intestazioni).

### 2.2 Out-of-Scope (Riservato a Milestone 2+)
- Generazione del grafo di attacco e percorsi di lateral movement (Milestone 2).
- Dashboard di Continuous Compliance NIS2 / DORA / ISO 27001 (Milestone 2).
- Moduli EASM, CSPM e Identity Leak Monitoring (Roadmap futura).
- Scansioni autenticate (credentialed) con agenti locali su endpoint (Roadmap futura).

---

## 3. Requisiti Non Funzionali (NFR) Dettagliati

### 3.1 Sicurezza Applicativa e Infrastrutturale (Security by Design)
- **NFR-SEC-01 (Prevenzione Command Injection):** L'invocazione di `nmap` da Go non deve **mai** passare per una shell di sistema (`sh -c` o `cmd /c`). Deve utilizzare rigorosamente `exec.CommandContext` passando argomenti in array validati e igienizzati contro flag injection o input non autorizzati.
- **NFR-SEC-02 (Principio del Minimo Privilegio nel Container):** Il container Go/Nmap non deve girare come root indiscriminato. Verrà configurato con un utente di servizio dotato esclusivamente delle Linux Capabilities necessarie per il raw packet probing (es. `CAP_NET_RAW`, `CAP_NET_ADMIN`, `CAP_NET_BIND_SERVICE`).
- **NFR-SEC-03 (Isolamento Multi-Tenant Rigoroso):** Ogni query al database deve incorporare esplicitamente il `tenant_id` derivato dal contesto di autenticazione JWT, prevenendo qualsiasi accesso cross-tenant (Broken Object Level Authorization - BOLA/IDOR).
- **NFR-SEC-04 (Autenticazione & Hashing):** Token di accesso stateless JWT a breve scadenza (15-30 min) associati a refresh token revocabili. Hashing delle credenziali utente tramite **Argon2id** (o **Bcrypt** con work factor $\ge 12$).
- **NFR-SEC-05 (Protezione dei Dati a Riposo e in Transito):** TLS 1.3 per il traffico Web/API. Supporto per cifratura a riposo su PostgreSQL.

### 3.2 Performance e Concorrenza
- **NFR-PERF-01 (Motore Asincrono Concorrente):** Le scansioni di rete non bloccano le chiamate API HTTP. Le scansioni vengono gestite da un worker pool asincrono in background che gestisce una coda interna con limiti di concorrenza configurabili per non saturare la banda host.
- **NFR-PERF-02 (Throughput di Parsing):** Il parsing dei file XML di Nmap deve utilizzare parser a streaming (`encoding/xml` decoder) per elaborare scansioni di grandi dimensioni senza picchi di memoria RAM.
- **NFR-PERF-03 (Indicizzazione Database):** Tutte le tabelle con filtraggio per `tenant_id`, `scan_id`, `cve_id` e severity devono disporre di indici compositi B-Tree e indici GIN sui campi `JSONB`.

### 3.3 Affidabilità, Resilienza e Graceful Shutdown
- **NFR-REL-01 (Graceful Cancellation):** Ogni esecuzione è vincolata a un `context.Context`. In caso di cancellazione manuale o arresto del server, il processo `nmap` correlato riceve `SIGTERM` e le risorse orfane vengono ripulite.
- **NFR-REL-02 (Transazionalità e Coerenza):** La registrazione dei risultati e delle vulnerabilità scoperte avviene all'interno di transazioni ACID PostgreSQL per evitare report orfani o incoerenti in caso di crash durante l'elaborazione.
- **NFR-REL-03 (Offline-Readiness):** Il sistema deve essere in grado di completare discovery, fingerprinting, correlazione CVE e generazione report anche in assenza totale di connettività Internet esterna.

### 3.4 Scalabilità e Manutenibilità
- **NFR-MAINT-01 (Architettura Clean / Hexagonal):** Backend strutturato secondo i principi della Clean Architecture (Domain, UseCases/Application, Adapters/Infrastructure, Delivery/HTTP), garantendo testabilità unitaria e disaccoppiamento dei driver Nmap e Postgres.
- **NFR-MAINT-02 (Continuous Integration & Testability):** Codice coperto da Unit Test e Integration Test con mock per Nmap e DB, puntando a una coverage $\ge 80\%$ sulla logica di business.
- **NFR-MAINT-03 (Structured Logging):** Utilizzo di `log/slog` nativo di Go con output JSON strutturato (`level`, `time`, `tenant_id`, `scan_id`, `message`).

