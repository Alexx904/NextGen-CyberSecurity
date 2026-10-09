# PROJECT STATE

> Memoria viva dell'agente. Letta a inizio sessione, aggiornata a fine di ogni turno.
> Non cancellare sezioni: se vuote, scrivi "nessuno".

**Ultimo aggiornamento:** 2026-10-09 22:32  
**Modalità progetto:** esistente (onboarded il 2026-10-09, validato dall'utente)  
**Fase:** Sviluppo Milestone 2 (Attack Path Mapping)  

## Project Profile

```
Nome progetto:           NextGen-CyberSecurity Platform
Linguaggi/stack:         Golang 1.22+ (Chi Router, pgx/v5, gorilla/websocket, robfig/cron) + React 18 / TypeScript 5.6 (Vite, Tailwind CSS, xterm.js, Recharts)
Gestore pacchetti:       Go Modules (backend), npm (frontend)
Comando test mirati:     go test -v ./internal/<package>/... (backend), npx vitest run src/<path>/<file>.test.ts (frontend)
Comando suite completa:  go test ./... (backend), npm test (frontend)
Comando build/typecheck: go build ./... (backend), npm run build (frontend)
Comando lint/format:     go vet ./... (backend)
Percorsi per livello:    unit=backend/pkg, backend/internal/domain, backend/internal/scanner; service=backend/internal/service; integration=backend/internal/delivery, backend/internal/repository/postgres; frontend=frontend/src/tests, frontend/src/**/*.test.tsx
Prerequisiti ambiente:   Docker & Docker Compose (PostgreSQL 16, Nmap su Linux container)
Come avviare l'app:      docker compose up --build (oppure 'go run ./cmd/server' + 'cd frontend && npm run dev')
Branch principale:       main (onboarding completato su chore/agent-onboarding)
Comandi vietati:         rm -rf fuori repo, git reset --hard non confermati, git push --force, DROP database senza conferma esplicita
```

## Baseline test (prima di ogni intervento)

| Data | Comando | Esito | Fallimenti preesistenti noti |
|---|---|---|---|
| 2026-10-09 22:30 | `go test ./...` (backend) | PASS (exit code 0, 14 package, inclusi handler e service attack path) | Nessuno. Tutti i test passano al 100%. |
| 2026-10-09 22:29 | `go build ./...` (backend) | PASS (exit code 0) | Nessuno. Compilazione pulita. |
| 2026-10-09 22:29 | `go vet ./...` (backend) | PASS (exit code 0) | Nessuno. Zero anomalie rilevate. |
| 2026-10-09 22:31 | `npm test` (frontend Vitest) | PASS (exit code 0, 10 file, 65 test) | Nessuno. |
| 2026-10-09 22:31 | `npm run build` (frontend tsc + Vite) | PASS (exit code 0) | Nessuno. Tutti i chunk ottimizzati. |

## In corso

- (nessuno)

## In attesa di accettazione utente

- **FEAT-M2-02 · Motore di Calcolo Lateral Movement & API Grafo (Backend Go)**: Servizio di aggregazione topologica, calcolo dei pivot multi-hop verso reti interne, identificazione chokepoint e registrazione endpoint REST `/api/v1/attack-paths/graph` e `/api/v1/attack-paths/chokepoints`.

## Bloccati / Domande aperte per l'utente

- (nessuno)

## Assunzioni confermate

- Struttura documentale autonoma v6.0 approvata dal Product Owner.
- Archiviazione storica dei file legacy in `docs/_legacy/` confermata.
- Pipeline CI/CD GitHub Actions da implementare con priorità P1.
- Code-splitting bundle frontend da implementare con priorità P1/P2.
- Feature cardine di Milestone 2 identificata: **Attack Path Mapping** (grafo visuale vulnerabilità e percorsi di attacco).

## Decisioni recenti (ultime 10, dettagli negli ADR)

- **PO-01 (2026-10-09):** Approvazione implementazione pipeline CI/CD GitHub Actions per backend e frontend.
- **PO-02 (2026-10-09):** Approvazione ottimizzazione bundle frontend tramite code-splitting e `React.lazy`.
- **PO-03 (2026-10-09):** Priorità strategica Milestone 2 assegnata a Attack Path Mapping.
- **ADR-01:** Modularizzazione di `2_ARCHITECTURE.md` con scomposizione per domini in `docs/domains/` (`scanning_and_discovery.md`, `ssh_and_firewall.md`, `reporting_and_assets.md`, `debug_console.md`, `attack_path.md`).
- **ADR-02:** Preservazione della cronologia Git via `git mv` per file storici e tracciamento in `docs/_legacy/MAPPING.md`.
- **ADR-03:** Creazione del branch dedicato `chore/agent-onboarding` per tutte le modifiche strutturali.

## Prossimi passi

1. Attendere feedback / accettazione da parte del PO sul task `FEAT-M2-02`.
2. Procedere all'implementazione del componente UI Frontend: `FEAT-M2-03 · Componente UI Visual Graph Explorer (Frontend React)`.

## Log sessioni (ultime 5, la più recente in alto)

- 2026-10-09 · Esecuzione FEAT-M2-02: implementato AttackPathService (correlazione DMZ/RCE -> internal DB/host, mitigazione chokepoints), AttackPathHandler con test unitari completi, cablaggio rotte in main.go · Backend test/vet/build e Frontend vitest/build 100% VERDI
- 2026-10-09 · Task FEAT-M2-01 accettato e verificato dal PO (`[x]`). Ingegneria dei requisiti e modello di dominio Attack Path consolidati · Test e Build VERDI
- 2026-10-09 · Esecuzione FEAT-M2-01: formalizzati FR-14 e UC-13 in 0_REQUIREMENTS.md, redatto docs/domains/attack_path.md, implementato modello di dominio Go attack_path.go con algoritmi di Cumulative Risk Scoring e Chokepoints · Test domain PASS 100%
- 2026-10-09 · Task ONBOARD-03 accettato dal PO (`[x]`). Ottimizzazione bundle e code-splitting consolidati su main · Test e Build VERDI
- 2026-10-09 · Esecuzione ONBOARD-03: implementato code-splitting con React.lazy in App.tsx e manualChunks in vite.config.ts. Dimensione chunk principale ridotta da 1.2 MB a 91 kB, 0 warning Vite · Test 65/65 PASS
