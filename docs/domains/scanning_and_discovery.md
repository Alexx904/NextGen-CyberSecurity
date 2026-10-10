# Dominio: Scanning & LAN Network Discovery

> **Dominio:** Network Reconnaissance, Vulnerability Scanning & Live SSE Streaming  
> **File di Riferimento:** `backend/internal/scanner/`, `backend/internal/service/scan_orchestrator.go`, `backend/internal/service/discovery_service.go`, `backend/internal/delivery/http/handler/discovery_handler.go`, `backend/internal/delivery/http/handler/scan_handler.go`, `frontend/src/pages/NetworkDiscovery.tsx`, `frontend/src/pages/ScanDetail.tsx`

---

## 1. Scanner Nmap Containerizzato & Parsing Streaming

### 1.1 Esecuzione Sicura (`NFR-SEC-01`)
- Il wrapper Go (`backend/internal/scanner/nmap_runner.go`) invoca il binario Nmap containerizzato senza mai passare per la shell di sistema (`sh` / `bash`), utilizzando rigorosamente `exec.CommandContext(ctx, "nmap", args...)`.
- Gli argomenti passati sono validati e sanitizzati contro flag injection da `backend/pkg/validator/target_validator.go`.
- Il container Go/Nmap opera con Linux Capabilities strettamente confinate (`CAP_NET_RAW`, `CAP_NET_ADMIN`, `CAP_NET_BIND_SERVICE`).

### 1.2 Profili di Scansione Specializzati e Macro-Categorie (`FR-16`, `FR-17`, `SEC-P1-01`, `SEC-P1-02`, `SEC-P1-03`)
Il sistema supporta 14 profili di scansione preconfigurati raggruppati in 6 macro-aree specialistiche (`domain.ScanProfileCategory`) classificati secondo tassonomia di invasività (`SAFE`, `INTRUSIVE`, `DANGEROUS`):

1. **ALL_IN_ONE:**
   - **All-in-One 360° Comprehensive Audit (`-sS -sV -O --script "vuln and safe" -T4 --open`):** Profilo generalista completo con script non distruttivi (`INTRUSIVE`).
2. **PEN_TEST (Penetration Testing & Probing):**
   - **Full Deep Scan (`-p- -sV -O -T4`):** Scansione esaustiva di tutte le 65535 porte TCP, OS detection e versioning completo (`INTRUSIVE`).
   - **Stealth SYN Penetration Audit (`-sS -Pn -T2 --open`):** Ricognizione stealth a basso rumore senza handshake TCP completo (`SAFE`).
   - **Firewall & Filter Bypass Check (`-sA -Pn -T3`):** Mappatura filtri firewall perimetrali e IDS con pacchetti TCP ACK (`SAFE`). Gestione accurata degli stati porta `unfiltered` e `filtered` per evitare falsi positivi di porta aperta.
3. **SYSTEM_FLAWS (Falle di Sistema & Vulnerability Assessment):**
   - **Standard Service Scan (`-sV -T4`):** Scansione standard delle top 1000 porte TCP con rilevamento versioni dei servizi (`SAFE`).
   - **OS & System Daemon Flaws Audit (`-sV -O --version-intensity 7 --script "vuln and safe" -T4`):** Rilevamento dettagliato del sistema operativo e vulnerabilità sicure (`INTRUSIVE`).
   - **Critical Services & RCE Surface (`-sV -p 21,22,23,25,53,80,110,135,139,143,443,445,1433,1521,3306,3389,5432,5900,6379,8080,8443 --script "vuln and safe" -T4`):** Audit mirato sui vettori critici più esposti (`INTRUSIVE`).
4. **WEB_APP (Web Application & API Services):**
   - **Web & API Surface Audit (`-sV -p 80,443,8000,8080,8443,8888,3000,5000,9000 -T4 --open`):** Rilevamento mirato di server HTTP/HTTPS e API REST (`SAFE`).
   - **Full Web Stack Deep Versioning (`-sV --version-all -p 80,443,8008,8080,8443,9090,9443 -T4`):** Interrogazione intensiva dei banner HTTP e intestazioni TLS (`SAFE`).
5. **DISCOVERY (Network & Asset Discovery):**
   - **Fast Discovery (`-T4 -F`):** Discovery veloce degli host attivi e delle top 100 porte TCP (`SAFE`).
   - **LAN Device Discovery (`-sn -PE -PS... -PU... -T4`):** Sweep veloce su subnet locali con abilitazione risoluzione reverse DNS (`SAFE`).
   - **Fast Ping & Host Sweep (LAN) (`-sn -PE -PS22,80,443,445 -T4`):** Ricognizione LAN ultra-rapida senza port scan per mappare host vivi (`SAFE`).
   - **UDP Specialized Services & Critical Infra (`-sU -p U:53,69,123,161,514,623,5353 -T4`):** Profilo specialistico UDP dedicato per DNS, TFTP, NTP, SNMP, Syslog, IPMI e mDNS senza intasare le code di scansione TCP (`SAFE`, SEC-P1-01).
6. **COMPLIANCE (Compliance & Hardening Check):**
   - **Legacy & Insecure Protocol Check (`-sV -p 21,23,79,110,143,512,513,514 -T3`):** Rilevamento servizi TCP in chiaro e protocolli legacy (porta 69 UDP rimossa per evitare timeout su socket TCP, SEC-P1-02) (`SAFE`).

### 1.3 Modalità di Composizione e Tassonomia di Invasività (`NewScanModal.tsx`, `SEC-P1-03`)
- Classificazione visiva con badge di invasività (`SAFE` in smeraldo, `INTRUSIVE` in ambra, `DANGEROUS` in rosso).
- Gate di sicurezza: per lanciare profili `INTRUSIVE` o `DANGEROUS` l'utente deve selezionare esplicitamente un checkbox di autorizzazione ad alto impatto di rete prima che il pulsante di avvio si abiliti.

### 1.7 Analisi Differenziale Scansioni (Scan Diff, `SEC-P2-03`)
- Endpoint `GET /api/v1/scans/{id}/diff`: confronta i finding della scansione specificata con la scansione completata più recente sullo stesso target.
- Classificazione automatica di ciascun finding:
  - `NEW`: Vulnerabilità rilevata per la prima volta nella scansione corrente.
  - `PERSISTENT`: Vulnerabilità già presente nella scansione precedente e non ancora rimediata.
  - `RESOLVED`: Vulnerabilità presente nella scansione precedente che non è più stata rilevata (remediation confermata).
- Identificazione della tupla finding: `(HostIP, Port, Protocol, FindingKey)` per evitare disallineamenti di ID.

---

## 4. Specifiche Contratti API REST Profili e Scansioni
- `GET /api/v1/scan-profiles`: Elenco di tutti i profili di sistema e custom del tenant.
- `POST /api/v1/scan-profiles`: Creazione e salvataggio di un nuovo profilo custom per il tenant.
- `DELETE /api/v1/scan-profiles/{id}`: Eliminazione di un profilo custom salvato.
- `GET /api/v1/scans/{id}/diff`: Report differenziale con conteggio `total_new`, `total_resolved`, `total_persistent` e array di `items`.

---

## 5. File chiave
- `backend/internal/scanner/nmap_runner.go`, `profiles.go`, `nmap_parser.go`, `oui_database.go`
- `backend/pkg/validator/target_validator.go`, `backend/pkg/auth/stream_ticket.go`
- `backend/internal/service/scan_orchestrator.go`, `scan_service.go`, `discovery_service.go`
- `backend/internal/delivery/http/handler/scan_handler.go`, `discovery_handler.go`
- `frontend/src/components/NewScanModal.tsx`, `frontend/src/pages/NetworkDiscovery.tsx`, `frontend/src/pages/ScanDetail.tsx`, `frontend/src/api/scan.ts`

## 6. Regole di business
- Il target non può appartenere alla deny-list interna (loopback, container di rete, link-local, cloud metadata).
- L'utente deve confermare formalmente la titolarità / autorizzazione al penetration test prima del lancio della scansione (art. 615-ter c.p.).
- I profili invasivi richiedono checkbox esplicito di autorizzazione prima del lancio (`SEC-P1-03`).
- Separatore `--` obbligatorio prima del target nell'invocazione di Nmap.
- I token JWT a lungo termine non devono comparire negli URL di streaming (uso di stream-ticket monouso a 30s).
- Hardening container: concesso solo `CAP_NET_RAW`, rimosso `CAP_NET_ADMIN`, abilitato `NMAP_PRIVILEGED=1` (`SEC-P1-05`).

## 7. Dipendenze da altri domini
- **reporting_and_assets:** fornitura dei dati scansionati e host vivi per il calcolo del Posture Score e la generazione dei report consolidati.
- **attack_path:** fornitura di porte aperte, vulnerabilità e subnet per la mappatura del grafo di attacco e lateral movement.
- **vulnerability_intelligence:** correlazione delle CVE individuate tramite Nmap NSE con il database NVD e offline catalog.

## 8. Test
- Unit test Go: `backend/pkg/validator/target_validator_test.go`, `backend/internal/scanner/profiles_test.go`, `nmap_runner_test.go`
- Service & Orchestrator test: `backend/internal/service/scan_orchestrator_test.go`, `scan_service_test.go`, `discovery_service_test.go`
- Handler test: `backend/internal/delivery/http/handler/scan_handler_test.go`
- Frontend test: `frontend/src/components/NewScanModal.test.tsx`, `frontend/src/pages/ScanDetail.test.tsx`, `frontend/src/hooks/useSSE.test.ts`

## 9. Ultimo allineamento
- 2026-10-10 (Epic 25 SEC-P1 Overhaul Scanner & Epic 26 SEC-P2 Attack Path & Diff implementati e verificati al 100%)


