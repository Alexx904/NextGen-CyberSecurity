# Dominio: Asset Tree Management & Centralized Reports Hub

> **Dominio:** Organizzazione Gerarchica Asset, Workstation Multi-Company Workflow e Generazione Report White-Label
> **File di Riferimento:** `backend/internal/domain/target.go`, `backend/internal/domain/report.go`, `backend/internal/service/target_group_service.go`, `backend/internal/service/report_service.go`, `backend/pkg/report/pdf_generator.go`, `backend/pkg/report/csv_generator.go`, `frontend/src/components/TreeExplorer.tsx`, `frontend/src/pages/Targets.tsx`, `frontend/src/pages/Reports.tsx`

---

## 1. Organizzazione Gerarchica Asset & Workstation Multi-Company Workflow (`FR-09`, `FR-15`)

### 1.1 Struttura Cartelle Ricorsiva & Company-Centric Scope
- Modello `TargetGroup` auto-referenziale su PostgreSQL con colonna `parent_id` (foreign key verso `target_groups.id`).
- I `TargetGroup` fungono sia da contenitori logici/reparti sia da **Aziende / Clienti distinti** per consulenti e auditor che operano su più perimetri dalla medesima workstation.
- Supporto ad alberi di profondità arbitraria (es. *Cliente Acme Corp -> Sede Milano -> Server DMZ*).
- Calcolo e aggregazione ricorsiva delle statistiche di rischio (totale host, numero di porte aperte, vulnerabilità critiche).
- Scansioni esclusive per azienda e workflow sequenziale operativo (Passo 1: Dashboard -> Passo 2: Discovery -> Passo 3: Target/Aziende -> Passo 4: Scansioni -> Passo 5: Analisi Minacce -> Passo 6: Report).

### 1.2 Target Supportati e Gestione Massiva
- IP singoli (v4 e v6).
- Range IP sequenziali (es. `192.168.1.1-192.168.1.50`).
- Blocchi CIDR (es. `10.0.0.0/24`).
- Nomi a dominio / FQDN (es. `scanme.nmap.org`).
- Blacklist di esclusione a livello di tenant per prevenire scansioni accidentali su nodi vietati.
- Interfaccia `/targets` con selezione massiva via checkbox per eliminazione rapida o spostamento tra cartelle aziendali.

---

## 2. Centralized Reports Hub (`/reports`) & Dossier Consolidati per Azienda (`FR-08`, `FR-18`)

### 2.1 Archivio e Raggruppamento per Azienda
- Archivio globale delle scansioni storiche con filtraggio multi-parametro: per data/periodo, target, cartella dell'albero, severità massima rilevata e profilo di scansione.
- Visualizzazione raggruppata per Cartella/Azienda Scope con card di riepilogo Posture Score (0-100, Rating A-F), audit completati e download diretto a 1-click.

### 2.2 Report Consolidati Aziendali (Multi-Host)
- **Endpoint:** `GET /api/v1/reports/company/{groupId}/download?type=pdf|csv&template=executive|technical`
- Aggrega tutti i target e le scansioni completate associate alla cartella aziendale (`group_id`).
- Deduplica intelligente dei finding multi-host tramite chiave composta `HostIP:Port:VulnKey`.
- Calcolo della postura globale dell'azienda integrata in un unico documento direzionale o tecnico.

### 2.3 Algoritmo di Normalizzazione Posture Score (`domain.CalculatePostureScore`)
Il punteggio di postura globale (0-100) e il rating a lettere (A-F) adottano una formula normalizzata per evitare il collasso su perimetri ampi:

$$\text{Deduction} = 25 \times N_{\text{Critical}} + 15 \times N_{\text{High}} + 5 \times N_{\text{Medium}} + 1 \times N_{\text{Low}}$$

$$\text{Deduction}_{\text{norm}} = \frac{\text{Deduction}}{\sqrt{\max(1, \text{UpHosts})}}$$

$$\text{PostureScore} = \text{clamp}\left(0, \, 100, \, \text{round}(100 - \text{Deduction}_{\text{norm}})\right)$$

- **Salvaguardia 0-Critical:** In assenza di vulnerabilità critiche ($N_{\text{Critical}} = 0$), il punteggio non collassa mai a 0/100, ma preserva un pavimento minimo di 25 (fascia D) garantendo coerenza qualitativa e quantitativa.
- **Scala di Rating:**
  - **A** ($\ge 90$): Postura eccellente, solo rilievi informativi o marginali.
  - **B** ($75 - 89$): Buona postura, poche anomalie di media entità.
  - **C** ($50 - 74$): Postura moderata, presenza di vulnerabilità High.
  - **D** ($25 - 49$): Postura degradata, vulnerabilità multiple da sanare.
  - **F** ($< 25$): Rischio critico elevato, presenza di vulnerabilità Critical attive (CVSS $\ge 9.0$).

### 2.4 Perfezionamento Grafico e Tipografico dei Report PDF A4 (`FEAT-REPORT-PDF-UX-01`)
- **Executive Summary Compatto (Esattamente 2 Pagine A4):**
  - **Pagina 1:** Copertina con branding tenant, Sintesi Direzionale, Statistiche KPI (Host, Porte, Vulnerabilità, MAX CVSS, Postura), Distribuzione Severità e Tabella Host Attivi.
  - **Pagina 2:** Top Priority Findings con remediation dettagliata e Governance / Prossimi Passi d'Azione.
  - Eliminazione di salti pagina arbitrari per un impaginato denso e professionale.
- **Wrapping Dinamico Multiriga:** Calcolo geometrico dell'altezza delle card di remediation in base al testo effettivo: zero testo troncato e nessun ricorso a font miniaturizzati illegibili.
- **Separazione Riga CVE e Azione:** Identificativo CVE e label "Azione:" posizionati su righe distinte con spaziatura e gutter calibrati per massima leggibilità.
- **Colori Semantici KPI:**
  - MAX CVSS dinamico: Rosso per Critical ($\ge 9.0$), Arancione per High ($7.0 - 8.9$), Giallo per Medium ($4.0 - 6.9$).
  - Host Attivi: Colore neutro ardesia scuro (`#334155`).
- **Nomi Software Reali:** Risoluzione nei titoli e descrizioni del vero prodotto/demone ricavato dai banner (`dnsmasq`, `vsftpd`, `Dropbear`, `Apache HTTP Server`, ecc.) al posto di etichette generiche di porta (`domain`, `ftp`, `ssh`).
- **Impaginazione Continua Report Tecnico:** Le schede dei finding scorrono sfruttando l'intera pagina fino al margine di 270mm (eliminando vuoti anomali di 80-150px a fondo pagina).
- **Date e Durate Valorizzate:** Aggregazione esatta dei tempi di scansione, eliminando diciture "N/A" su "Data Avvio" e "Durata Totale".

---

## 4. Gestione Aziende, Reparti & CompanyContext Globale (`Epic 27`, `FEAT-COMPANY-SCOPE`)

### 4.1 Modello Gerarchico Azienda / Reparto
- **Aziende Clienti:** Nodi di livello radice in `target_groups` caratterizzati da `parent_id IS NULL`. Rappresentano le entità legali o clienti primari gestiti dal consulente o team MSSP.
- **Reparti Aziendali:** Nodi discendenti in `target_groups` con `parent_id != NULL`. Rappresentano aree, uffici o subnet distinte (es. *Reparto IT*, *Finance*, *DMZ Produzione*).
- **Target IP Associati:** Ciascun target appartiene a un reparto specifico (`group_id`) o direttamente all'azienda radice. La chiusura ricorsiva (`getAllDescendantGroupIds`) mappa istantaneamente tutti gli IP all'azienda madre.

### 4.2 CompanyContext Globale (`frontend/src/context/CompanyContext.tsx`)
- Stato centralizzato React: `selectedCompanyId`, `selectedCompany`, `departments`, `companyGroupIds`, `companyTargetIds`, `companyTargetValues`.
- Persistenza automatica in `localStorage` (`nextgen_active_company_id`) per preservare lo scope attivo tra ricaricamenti pagina.
- Helper unificati di filtraggio automatico:
  - `filterTargets(targets)`: restituisce solo i target dello scope.
  - `filterScans(scans)`: filtra le scansioni dei soli target dell'azienda.
  - `filterFindings(findings)`: isola le vulnerabilità CVE dei target dell'azienda.
  - `isTargetInActiveScope(targetIdOrValue)`: test booleano rapido su ID o stringa IP/hostname.
- Fallback sicuro `DEFAULT_COMPANY_CONTEXT` per esecuzione trasparente in test isolati senza eccezioni.

### 4.3 Navigazione e Dashboard `/companies`
- **Sidebar Dropdown:** Selettore globale con badge stato, ricerca rapida e commutazione istantanea tra "Tutte le Aziende" e l'azienda cliente desiderata.
- **Pagina `/companies`:** Cruscotto direzionale con metriche aggregate (totale aziende, reparti censiti, IP gestiti, media posture), color tag personalizzato (#hex), operazioni CRUD complete e tasto di attivazione scope.
- **Contestualizzazione Viste:**
  - `/targets`: visualizza esclusivamente i reparti dell'azienda selezionata e crea sotto-reparti per l'azienda attiva.
  - `/dashboard`: card posture per l'azienda e per ciascuno dei suoi reparti, banner scope attivo.
  - `/scans` & `NewScanModal`: scansioni e schedulazioni filtrate; modale pre-seleziona l'azienda attiva e limita la scelta ai soli suoi reparti.
  - `/findings`: ispezione per reparto ed IP dell'azienda attiva, eliminando switcher duplicati.
  - `/reports`: aggregazione sui reparti dell'azienda e download istantaneo del dossier consolidato (Executive PDF, Tecnico PDF, CSV).
  - `/attack-path`: focalizzazione del grafo topologico sui bersagli dell'azienda attiva.

---

## 5. Ultimo Allineamento
- **Data:** 2026-10-11
- **Hash commit:** In attesa di accettazione PO (Epic 27 FEAT-COMPANY-SCOPE)
