# Dominio: Scanning & LAN Network Discovery

> **Dominio:** Network Reconnaissance, Vulnerability Scanning & Live SSE Streaming  
> **File di Riferimento:** `backend/internal/scanner/`, `backend/internal/service/scan_orchestrator.go`, `backend/internal/service/discovery_service.go`, `backend/internal/delivery/http/handler/discovery_handler.go`, `backend/internal/delivery/http/handler/scan_handler.go`, `frontend/src/pages/NetworkDiscovery.tsx`, `frontend/src/pages/ScanDetail.tsx`

---

## 1. Scanner Nmap Containerizzato & Parsing Streaming

### 1.1 Esecuzione Sicura (`NFR-SEC-01`)
- Il wrapper Go (`backend/internal/scanner/nmap_runner.go`) invoca il binario Nmap containerizzato senza mai passare per la shell di sistema (`sh` / `bash`), utilizzando rigorosamente `exec.CommandContext(ctx, "nmap", args...)`.
- Gli argomenti passati sono validati e sanitizzati contro flag injection da `backend/pkg/validator/target_validator.go`.
- Il container Go/Nmap opera con Linux Capabilities strettamente confinate (`CAP_NET_RAW`, `CAP_NET_ADMIN`, `CAP_NET_BIND_SERVICE`).

### 1.2 Profili di Scansione Specializzati e Macro-Categorie (`FR-16`, `FR-17`)
Il sistema supporta 13 profili di scansione preconfigurati raggruppati in 6 macro-aree specialistiche (`domain.ScanProfileCategory`):

1. **ALL_IN_ONE:**
   - **All-in-One 360° Comprehensive Audit (`-sS -sV -O --script=vuln -T4 --open`):** Profilo generalista completo che mappa host vivi, scansiona porte standard, rileva versioni demoni (`-sV`), fingerprinting OS (`-O`) e interroga vulnerabilità note (`--script=vuln`).
2. **PEN_TEST (Penetration Testing & Probing):**
   - **Full Deep Scan (`-p- -sV -O -T4`):** Scansione esaustiva di tutte le 65535 porte TCP, OS detection e versioning completo.
   - **Stealth SYN Penetration Audit (`-sS -Pn -T2 --open`):** Ricognizione stealth a basso rumore senza handshake TCP completo, elusione baseline firewall e solo porte aperte.
   - **Firewall & Filter Bypass Check (`-sA -Pn -T3`):** Mappatura filtri firewall perimetrali e IDS con pacchetti TCP ACK per verificare se i firewall sono stateful o stateless.
3. **SYSTEM_FLAWS (Falle di Sistema & Vulnerability Assessment):**
   - **Standard Service Scan (`-sV -T4`):** Scansione standard delle top 1000 porte TCP con rilevamento versioni dei servizi.
   - **OS & System Daemon Flaws Audit (`-sV -O --version-intensity 7 --script=vuln -T4`):** Rilevamento dettagliato del sistema operativo, fingerprinting demoni e script attivi di vulnerabilità Nmap.
   - **Critical Services & RCE Surface (`-sV -p 21,22,23,25,53,80,110,135,139,143,443,445,1433,1521,3306,3389,5432,5900,6379,8080,8443 --script=vuln -T4`):** Audit mirato sui vettori critici più esposti a exploit remoti e vulnerabilità RCE/PrivEsc.
4. **WEB_APP (Web Application & API Services):**
   - **Web & API Surface Audit (`-sV -p 80,443,8000,8080,8443,8888,3000,5000,9000 -T4 --open`):** Rilevamento mirato di server HTTP/HTTPS, API REST e reverse proxy applicativi.
   - **Full Web Stack Deep Versioning (`-sV --version-all -p 80,443,8008,8080,8443,9090,9443 -T4`):** Interrogazione intensiva dei banner HTTP e intestazioni TLS per scovare componenti obsoleti.
5. **DISCOVERY (Network & Asset Discovery):**
   - **Fast Discovery (`-T4 -F`):** Discovery veloce degli host attivi e delle top 100 porte TCP.
   - **LAN Device Discovery (`-sn -PE -PS... -PU... -T4 -n`):** Sweep ad altissima velocità su subnet locali con probe ICMP/SYN/UDP.
   - **Fast Ping & Host Sweep (LAN) (`-sn -PE -PS22,80,443,445 -T4 -n`):** Ricognizione LAN ultra-rapida senza port scan per mappare tutti gli host accesi.
6. **COMPLIANCE (Compliance & Hardening Check):**
   - **Legacy & Insecure Protocol Check (`-sV -p 21,23,69,79,110,143,512,513,514 -T3`):** Rilevamento servizi in chiaro e protocolli legacy obsoleti (Telnet, FTP, TFTP, Rsh, POP3).

### 1.3 Modalità di Composizione e Configuratore Avanzato (`NewScanModal.tsx`)
- **Modalità Guidata / Semplice:** Selezione con 1-click del profilo per obiettivo con badge tempo stimato e impatto sulla rete.
- **Modalità "Componi Multi-Obiettivo":** Selezione combinata tramite checkbox di categorie di audit (es. Web App + Falle di Sistema + Script Vulnerabilità).
- **Modalità Professionale:**
  - Definizione granulare di porte (Top 100/1000, Tutte 65535, porte custom separate da virgola o range).
  - Timing template (-T1 .. -T4), Version Intensity (0-9), OS Detection (`-O`), ping bypass (`-Pn`), solo porte aperte (`--open`).
  - Abilitazione selettiva script vulnerabilità `--script=vuln` e `--script=vulners`.
  - **Live Command Preview:** Anteprima istantanea e reattiva del comando Nmap generato con validazione preventiva client e server.

### 1.4 Validazione Rigorosa degli Argomenti & Scope Check (`backend/pkg/validator/target_validator.go`, `backend/internal/scanner/profiles.go`)
- **Deny-List & Scope Check (`CheckTargetScope`):**
  - Blocco assoluto di loopback IPv4/IPv6 (`127.0.0.0/8`, `::1/128`), link-local (`169.254.0.0/16`, `fe80::/10`), range bridge Docker (`172.17.0.0/16` .. `172.20.0.0/16`) e hostname infrastrutturali (`localhost`, `postgres`, `backend`, `frontend`, `.internal`, `.local`).
  - Risoluzione DNS preventiva (`net.LookupIP`) per contrastare attacchi di DNS rebinding verso host interni riservati.
  - Blocco di subnet eccessivamente ampie (CIDR `/8` o prefissi con più di 65536 host per IPv4).
- **`IsSafeArgument(arg)`:** Verifica assenza di metacaratteri shell malevoli (`;`, `&`, `|`, `$`, `>`, `<`, ecc.) e blacklist severa di flag vietati (`DisallowedFlags`: `--script-args`, output su file `-oN/-oX/-oA`, caricamento file `-iL`, privilegi `--privileged`).
- **Delimitatore `--` POSIX:** `BuildSafeArguments` posiziona il delimitatore standard `--` immediatamente prima del target, prevenendo qualsiasi attacco di flag injection via hostname ostili.
- **Default Offline & Safe Scripts:** Disabilitazione per default di script di terze parti come `vulners` che trasmettono dati tecnici di rete all'esterno senza consenso esplicito; utilizzo di script sicuri e locali (`--script "vuln and safe"`).

### 1.5 Scansioni a Livello di Cartella / Azienda
- Endpoint `POST /api/v1/target-groups/{id}/scan`: esecuzione concorrente e ordinata della scansione su tutti gli IP appartenenti al perimetro dell'azienda/cartella.

### 1.6 Filtro Host Spenti (Alive-Only Host Filter)
- Nei range IP o blocchi CIDR, gli host marcati come `down` non generano record superflui nel database. Vengono salvati esclusivamente gli host con stato `up`, mantenendo il database e la dashboard privi di rumore.

---

## 2. High-Speed LAN Network Discovery

Modulo dedicato (`/network-discovery`) per il censimento istantaneo dei nodi attivi in rete locale:
1. **Sweep ARP / Ping:** Invio di probe a bassa latenza su subnet definite dall'operatore (`-sn -PR -T4`).
2. **Lookup IEEE OUI:** Database locale integrato (`backend/internal/scanner/oui_database.go`) contenente prefissi MAC IEEE per identificare immediatamente il produttore hardware (es. Apple, Dell, Raspberry Pi, Cisco).
3. **Banner Grabbing Leggero:** Ispezione rapida di porte comuni (80, 443, 8080) per catturare l'intestazione HTTP `Server` e il titolo HTML della pagina.
4. **Import e Assegnazione Aziendale:** Assegnazione diretta dei nodi scoperti a una cartella/azienda esistente o creazione 1-click di una nuova azienda cliente.

---

## 3. Real-Time Streaming via Server-Sent Events (SSE) & Stream-Ticket Monouso

### 3.1 Endpoint di Streaming
- `GET /api/v1/network-discovery/stream?session_id={session_id}&ticket={stream_ticket}`
- `GET /api/v1/scans/{id}/stream?ticket={stream_ticket}`
- *(Supporto trasparente di fallback a `?token={jwt}` mantenuto per retrocompatibilità).*

Header di risposta: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`.

### 3.2 Tipi di Evento
- `device_discovered`: Emesso in tempo reale per ciascun host che risponde al probe LAN, con IP, MAC, Vendor, Hostname e latenza.
- `phase`: Notifica di passaggio fase di scansione (`HOST_DISCOVERY`, `PORT_SCAN`, `SERVICE_DETECTION`, `CVE_CORRELATION`).
- `progress`: Percentuale di avanzamento dinamica (`percentage: float64`, `message: string`).
- `finding_discovered`: Notifica di porta aperta / servizio / CVE individuata con persistenza incrementale su PostgreSQL.
- `status_change`: Notifica di conclusione (`COMPLETED`, `FAILED`, `CANCELLED`) con statistiche riassuntive.

---

## 4. Specifiche Contratti API REST Profili di Scansione

- `GET /api/v1/scan-profiles`: Elenco di tutti i profili di sistema e custom del tenant.
- `POST /api/v1/scan-profiles`: Creazione e salvataggio di un nuovo profilo custom per il tenant.
- `DELETE /api/v1/scan-profiles/{id}`: Eliminazione di un profilo custom salvato.

---

## 5. File chiave
- `backend/internal/scanner/nmap_runner.go`, `profiles.go`, `nmap_parser.go`, `oui_database.go`
- `backend/pkg/validator/target_validator.go`, `backend/pkg/auth/stream_ticket.go`
- `backend/internal/service/scan_orchestrator.go`, `discovery_service.go`
- `backend/internal/delivery/http/handler/scan_handler.go`, `discovery_handler.go`
- `frontend/src/components/NewScanModal.tsx`, `frontend/src/pages/NetworkDiscovery.tsx`, `frontend/src/pages/ScanDetail.tsx`

## 6. Regole di business
- Il target non può appartenere alla deny-list interna (loopback, container di rete, link-local, cloud metadata).
- L'utente deve confermare formalmente la titolarità / autorizzazione al penetration test prima del lancio della scansione (art. 615-ter c.p.).
- Separatore `--` obbligatorio prima del target nell'invocazione di Nmap.
- I token JWT a lungo termine non devono comparire negli URL di streaming (uso di stream-ticket monouso a 30s).

## 7. Dipendenze da altri domini
- **reporting_and_assets:** fornitura dei dati scansionati e host vivi per il calcolo del Posture Score e la generazione dei report consolidati.
- **attack_path:** fornitura di porte aperte, vulnerabilità e subnet per la mappatura del grafo di attacco e lateral movement.
- **vulnerability_intelligence:** correlazione delle CVE individuate tramite Nmap NSE con il database NVD e offline catalog.

## 8. Test
- Unit test Go: `backend/pkg/validator/target_validator_test.go`, `backend/internal/scanner/profiles_test.go`, `nmap_runner_test.go`
- Service & Orchestrator test: `backend/internal/service/scan_orchestrator_test.go`, `discovery_service_test.go`
- Frontend test: `frontend/src/components/NewScanModal.test.tsx`, `frontend/src/pages/ScanDetail.test.tsx`, `frontend/src/hooks/useSSE.test.ts`

## 9. Ultimo allineamento
- 2026-10-10 (Epic 24 SEC-P0 Hardening implementato e verificato al 100%)


