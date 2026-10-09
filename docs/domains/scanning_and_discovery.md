# Dominio: Scanning & LAN Network Discovery

> **Dominio:** Network Reconnaissance, Vulnerability Scanning & Live SSE Streaming  
> **File di Riferimento:** `backend/internal/scanner/`, `backend/internal/service/scan_orchestrator.go`, `backend/internal/service/discovery_service.go`, `backend/internal/delivery/http/handler/discovery_handler.go`, `backend/internal/delivery/http/handler/scan_handler.go`, `frontend/src/pages/NetworkDiscovery.tsx`, `frontend/src/pages/ScanDetail.tsx`

---

## 1. Scanner Nmap Containerizzato & Parsing Streaming

### 1.1 Esecuzione Sicura (`NFR-SEC-01`)
- Il wrapper Go (`backend/internal/scanner/nmap_runner.go`) invoca il binario Nmap containerizzato senza mai passare per la shell di sistema (`sh` / `bash`), utilizzando rigorosamente `exec.CommandContext(ctx, "nmap", args...)`.
- Gli argomenti passati sono validati e sanitizzati contro flag injection da `backend/pkg/validator/target_validator.go`.
- Il container Go/Nmap opera con Linux Capabilities strettamente confinate (`CAP_NET_RAW`, `CAP_NET_ADMIN`, `CAP_NET_BIND_SERVICE`).

### 1.2 Profili di Scansione (`backend/internal/scanner/profiles.go`)
- **Fast Discovery (`-sS -T4 --top-ports 100 -n`):** Individuazione veloce porte comuni senza risoluzione DNS inversa.
- **Standard Service Scan (`-sV -O -T3 --top-ports 1000`):** Service fingerprinting e rilevazione sistema operativo.
- **Full Deep Scan (`-sS -sV -O -p- -sC -T3`):** Scansione esaustiva di tutte le 65535 porte TCP con default scripts Nmap.
- **LAN Device Discovery (`-sn -PR` / ARP scan):** Sweep ad altissima velocità su subnet locali.

### 1.3 Filtro Host Spenti (Alive-Only Host Filter)
- Nei range IP o blocchi CIDR, gli host marcati come `down` non generano record superflui nel database. Vengono salvati esclusivamente gli host con stato `up`, mantenendo il database e la dashboard privi di rumore.

---

## 2. High-Speed LAN Network Discovery

Modulo dedicato (`/network-discovery`) per il censimento istantaneo dei nodi attivi in rete locale:
1. **Sweep ARP / Ping:** Invio di probe a bassa latenza su subnet definite dall'operatore.
2. **Lookup IEEE OUI:** Database locale integrato (`backend/internal/scanner/oui_database.go`) contenente prefissi MAC IEEE per identificare immediatamente il produttore hardware (es. Apple, Dell, Raspberry Pi, Cisco).
3. **Banner Grabbing Leggero:** Ispezione rapida di porte comuni (80, 443, 8080) per catturare l'intestazione HTTP `Server` e il titolo HTML della pagina.
4. **Import Immediato:** Capacità di importare con un click i nodi scoperti all'interno delle cartelle Target per successive scansioni di vulnerabilità.

---

## 3. Real-Time Streaming via Server-Sent Events (SSE)

### 3.1 Endpoint di Streaming
- `GET /api/v1/network-discovery/stream?session_id={session_id}&token={jwt}`
- `GET /api/v1/scans/{id}/stream?token={jwt}`

Header di risposta: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`.

### 3.2 Tipi di Evento
- `device_discovered`: Emesso in tempo reale per ciascun host che risponde al probe LAN, con IP, MAC, Vendor, Hostname e latenza.
- `phase`: Notifica di passaggio fase di scansione (`HOST_DISCOVERY`, `PORT_SCAN`, `SERVICE_DETECTION`, `CVE_CORRELATION`).
- `progress`: Percentuale di avanzamento dinamica (`percentage: float64`, `message: string`).
- `finding_discovered`: Notifica di porta aperta / servizio / CVE individuata con persistenza incrementale su PostgreSQL.
- `status_change`: Notifica di conclusione (`COMPLETED`, `FAILED`, `CANCELLED`) con statistiche riassuntive.

