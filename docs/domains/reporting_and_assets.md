# Dominio: Asset Tree Management & Centralized Reports Hub

> **Dominio:** Organizzazione Gerarchica Asset, Gestione Target e Generazione Report  
> **File di Riferimento:** `backend/internal/domain/target.go`, `backend/internal/domain/report.go`, `backend/internal/service/target_group_service.go`, `backend/internal/service/report_service.go`, `backend/pkg/report/pdf_generator.go`, `backend/pkg/report/csv_generator.go`, `frontend/src/components/TreeExplorer.tsx`, `frontend/src/pages/Targets.tsx`, `frontend/src/pages/Reports.tsx`

---

## 1. Organizzazione Gerarchica Asset (Nested Tree Explorer)

### 1.1 Struttura Cartelle Ricorsiva
- Modello `TargetGroup` auto-referenziale su PostgreSQL con colonna `parent_id` (foreign key verso `target_groups.id`).
- Supporto ad alberi di profondità arbitraria per raggruppare target per reparto, sede geografica o ambiente (es. *HQ -> DataCenter Milano -> Server DMZ*).
- Calcolo e aggregazione ricorsiva delle statistiche di rischio (totale host, numero di porte aperte, vulnerabilità critiche).

### 1.2 Target Supportati
- IP singoli (v4 e v6).
- Range IP sequenziali (es. `192.168.1.1-192.168.1.50`).
- Blocchi CIDR (es. `10.0.0.0/24`).
- Nomi a dominio / FQDN (es. `scanme.nmap.org`).
- Blacklist di esclusione a livello di tenant per prevenire scansioni accidentali su nodi vietati.

---

## 2. Centralized Reports Hub (`/reports`)

### 2.1 Archivio e Filtri Avanzati
- Archivio globale delle scansioni storiche con filtraggio multi-parametro: per data/periodo, target, cartella dell'albero, severità massima rilevata e profilo di scansione.

### 2.2 Motore di Generazione White-Label
- **PDF Executive Report:** Sintesi per il management con indicatori grafici, posture di rischio aggregata e raccomandazioni strategiche con branding del tenant.
- **PDF Technical Deep-Dive Report:** Analisi puntuale di ogni host, porte aperte, banner di servizio, CVE correlate con punteggio CVSS v3.1 e passaggi di remediation dettagliati.
- **Esportazione CSV RFC 4180:** Export tabellare grezzo compatibile con Excel, SIEM ed elaborazioni esterne.
