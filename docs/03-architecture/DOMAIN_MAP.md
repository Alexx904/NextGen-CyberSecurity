# Domain Map

> Responsabile: Software Architect. Percorsi repository relativi alla root. Le righe per milestone future sono prenotazioni; i documenti mancanti vanno creati prima della relativa implementazione.

## 5. Documentazione Dettagliata per Dominio

Per garantire modularità ed efficienza di contesto, i dettagli tecnici e i contratti specifici sono frammentati nei file di dominio:

1. **[Scanning & LAN Network Discovery](../domains/scanning_and_discovery.md)**: modale minimale con profili tecnici ed editor Nmap, anteprima coerente, validazione semantica backend e classificazione custom. Contratti storici preservati; discovery host con arricchimento porte opzionale, scansioni ricorsive per gruppo e SSE.
2. **[Remote SSH & iptables Firewall Manager](../domains/ssh_and_firewall.md)**: Inventario clienti autonomo dal filtro aziende, Gestione credenziali e cifratura AES-256-GCM, Anti-Lockout Safety Timer, ispezione e modifica regole iptables, WebSocket Web Terminal SSH PTY.
3. **[Asset Tree Management & Centralized Reports Hub](../domains/reporting_and_assets.md)**: Workstation Multi-Company Workflow, albero gerarchico asset (`TargetGroup`), dossier consolidati per azienda (`GET /api/v1/reports/company/{groupId}/download`), algoritmo Posture Score normalizzato e perfezionamento tipografico A4 (Executive 2 pagine dense, technical continuous flow fino a 270mm, wrapping multiriga fluido).
4. **[Debug & Testing Live Console Drawer](../domains/debug_console.md)**: In-Memory Ring Buffer (500 log), intercettori Axios/SSE/WS/JS, sanificazione credenziali `[REDACTED]` ed esportatore Markdown preformattato per AI.
5. **[Attack Path Mapping & Threat Chaining](../domains/attack_path.md)**: Analisi euristica di associazioni e ipotesi di accesso, indice di priorità e chokepoint dei percorsi modellati; vista 2D percorso/completa, scope aziendale e guida integrata. Non prova di compromissione o lateral movement.
6. **[Vulnerability Intelligence & CVE Correlation](../domains/vulnerability_intelligence.md)**: Catalogo locale offline-ready, client REST NIST NVD API 2.0 (timeout 4s), auto-caching PostgreSQL su `cve_entries`, correlazione script NSE (`--script=vuln`, `--script=vulners`) ed endpoint `/api/v1/cve/stats` e `/api/v1/cve/sync`.
7. **[Active Defense & SOAR Light](../domains/active_defense_and_soar.md) (M3):** preview e blocco porta chokepoint SERVICE. Isolamento host, UI, virtual patching e alerting restano pianificati.
8. **Compliance & Governance (M4, pianificato):** Mappatura NIS2/DORA/ISO 27001, scorecard per reparto, dossier PDF audit-ready. Documento di dominio da creare all'avvio della milestone.
9. **AI Security Copilot & Assistant (M5, pianificato):** Generazione dinamica playbook di mitigazione, sintesi narrative executive per report, query semantiche "Ask Infrastructure". Documento di dominio da creare all'avvio della milestone.

### Mappa documenti Full-Stack (Glob -> Dominio -> Documento)

Estensione E28-S1-T1 (2026-10-17, working tree): `backend/internal/domain/defense*`,
`backend/internal/delivery/http/handler/defense*` e `frontend/src/types/defense.ts`
appartengono al dominio [Active Defense](../domains/active_defense_and_soar.md).
Il wiring in `backend/cmd/server/main.go` collega i domini coinvolti.
Tabella vincolante per il gate **Doc Sync**. Deve coprire il 100% del codice applicativo (Backend Go, Frontend React, Migrazioni SQL):

| Percorso (glob Backend / Frontend / DB) | Dominio | Documento di Riferimento |
| :--- | :--- | :--- |
| `backend/internal/domain/scan.go`<br/>`backend/internal/domain/event.go` (payload scansioni)<br/>`backend/internal/delivery/http/handler/scan_handler*`<br/>`frontend/src/types/scan.ts` | scanning & discovery, contratti condivisi | `docs/domains/scanning_and_discovery.md` |
| `backend/internal/scanner/**`<br/>`backend/internal/service/scan*`<br/>`backend/internal/service/discovery*`<br/>`backend/internal/delivery/scan*`<br/>`backend/internal/delivery/discovery*`<br/>`frontend/src/pages/Scans*`<br/>`frontend/src/pages/NetworkDiscovery*`<br/>`frontend/src/components/scans/**`<br/>`frontend/src/components/NewScanModal*`<br/>`frontend/src/components/discovery/**`<br/>`frontend/src/components/Scan*` | scanning & discovery | `docs/domains/scanning_and_discovery.md` |
| `backend/internal/domain/ssh_server*`<br/>`frontend/src/types/ssh.ts`<br/>`frontend/src/components/SSHServerModal*`<br/>`backend/internal/service/ssh*`<br/>`backend/internal/service/firewall*`<br/>`backend/internal/delivery/ssh*`<br/>`backend/internal/delivery/firewall*`<br/>`frontend/src/pages/SSHServers*`<br/>`frontend/src/components/ssh/**`<br/>`frontend/src/components/Terminal*` | ssh & firewall | `docs/domains/ssh_and_firewall.md` |
| `frontend/src/App.tsx`<br/>`frontend/src/components/layout/AppLayout*`<br/>`frontend/src/components/auth/CompanyRoute*`<br/>`frontend/src/components/GroupModal*`<br/>`backend/internal/service/target*`<br/>`backend/internal/service/report*`<br/>`backend/pkg/report/**`<br/>`backend/internal/delivery/target*`<br/>`backend/internal/delivery/report*`<br/>`backend/internal/delivery/tenant*`<br/>`frontend/src/pages/Companies*`<br/>`frontend/src/pages/Targets*`<br/>`frontend/src/pages/Reports*`<br/>`frontend/src/pages/Dashboard*`<br/>`frontend/src/context/CompanyContext*`<br/>`frontend/src/components/Company*`<br/>`frontend/src/components/TreeExplorer*`<br/>`frontend/src/components/Target*` | reporting & assets | `docs/domains/reporting_and_assets.md` |
| `frontend/src/pages/DebugConsole*`<br/>`frontend/src/components/Debug*`<br/>`frontend/src/utils/debugLogger*` | debug_console | `docs/domains/debug_console.md` |
| `backend/internal/domain/attack_path*`<br/>`backend/internal/service/attack_path*`<br/>`backend/internal/delivery/http/handler/attack_path*`<br/>`frontend/src/types/attackPath.ts`<br/>`frontend/src/api/attackPath.ts`<br/>`frontend/src/pages/AttackPath*`<br/>`frontend/src/components/graph/**` | attack_path | `docs/domains/attack_path.md` |
| `backend/internal/service/cve*`<br/>`backend/internal/delivery/cve*`<br/>`frontend/src/pages/Findings*`<br/>`frontend/src/components/findings/**`<br/>`frontend/src/components/Finding*` | vulnerability intelligence | `docs/domains/vulnerability_intelligence.md` |
| `backend/internal/service/defense*`<br/>`backend/internal/service/alert*`<br/>`backend/internal/delivery/defense*`<br/>`backend/internal/delivery/alert*`<br/>`frontend/src/pages/ActiveDefense*`<br/>`frontend/src/components/defense/**`<br/>`frontend/src/components/alert/**` | active defense & soar | `docs/domains/active_defense_and_soar.md` |
| `backend/internal/service/compliance*`<br/>`backend/internal/delivery/compliance*`<br/>`frontend/src/pages/Compliance*`<br/>`frontend/src/components/compliance/**` | compliance & governance | `docs/domains/compliance_and_governance.md` |
| `backend/internal/service/copilot*`<br/>`backend/internal/delivery/copilot*`<br/>`frontend/src/components/copilot/**` | ai security copilot | `docs/domains/ai_copilot.md` |
| `backend/internal/repository/migrations/**`<br/>`backend/internal/repository/postgres/**` | database & core erd | `docs/03-architecture/LLD/DATABASE_SCHEMA.md` |

---
