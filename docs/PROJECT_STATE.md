# PROJECT STATE

> Memoria viva dell'agente. Letta a inizio sessione, aggiornata a fine di ogni turno.
> Non cancellare sezioni: se vuote, scrivi "nessuno".

**Ultimo aggiornamento:** 2026-10-09 21:55  
**Modalità progetto:** esistente (onboarded il 2026-10-09)  
**Fase:** Onboarded / Manutenzione & Roadmap Milestone 2  

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
Branch principale:       main (onboarding branch: chore/agent-onboarding)
Comandi vietati:         rm -rf fuori repo, git reset --hard non confermati, git push --force, DROP database senza conferma esplicita
```

## Baseline test (prima di ogni intervento)

| Data | Comando | Esito | Fallimenti preesistenti noti |
|---|---|---|---|
| 2026-10-09 21:49 | `go test ./...` (backend) | PASS (exit code 0, 14 package) | Nessuno. Tutti i test passano al 100%. |
| 2026-10-09 21:49 | `go build ./...` (backend) | PASS (exit code 0) | Nessuno. Compilazione pulita. |
| 2026-10-09 21:49 | `go vet ./...` (backend) | PASS (exit code 0) | Nessuno. Zero anomalie rilevate. |
| 2026-10-09 21:50 | `npm test` (frontend Vitest) | PASS (exit code 0, 10 file, 65 test) | Nessuno. (Nota: warning informativi su act() in DebugConsole e React Router v7 future flags). |
| 2026-10-09 21:50 | `npm run build` (frontend tsc + Vite) | PASS (exit code 0) | Nessuno. (Nota: warning chunk size > 500kB su bundle principale, da ottimizzare con code-splitting). |

## In corso

- (nessuno)

## In attesa di accettazione utente

| Task | Feature | Come provarla (sintesi) | Consegnata il |
|---|---|---|---|
| ONBOARD | Onboarding Progetto & Struttura Autonoma v6.0 | Consultare `docs/` (PROJECT_STATE, 0_REQUIREMENTS, 2_ARCHITECTURE, 3_TASK_LIST) e verificare che build e test siano verdi. | 2026-10-09 |

## Bloccati / Domande aperte per l'utente

1. **CI/CD Pipeline:** Vuoi che impostiamo una pipeline GitHub Actions in `.github/workflows/ci.yml` per eseguire automaticamente i test e la build a ogni push?
2. **Ottimizzazione Bundle Frontend:** Desideri dare priorità al code-splitting con `React.lazy` su Vite per snellire il file JavaScript principale (attualmente > 1.2 MB non compresso)?
3. **Roadmap Milestone 2:** Tra **Attack Path Mapping** (grafo interattivo di lateral movement) e **Continuous Compliance Score** (NIS2 / DORA / ISO 27001), quale feature strategica ha priorità per il prossimo sprint?

## Assunzioni fatte (da confermare)

- L'ambiente primario di esecuzione per la produzione e per i collaudi E2E live con Nmap è Docker Compose (`docker-compose.yml`), con Linux capabilities `NET_RAW` e `NET_ADMIN`.
- La struttura dei documenti legacy (`Agent.md`, `Brainstorming_CyberSec_App.md`) è archiviata e conservata storicamente in `docs/_legacy/` con mapping tracciato in `docs/_legacy/MAPPING.md`.
- I 60 task delle Milestone 1, 1.1 e 1.2 sono completati, testati e consolidati nel changelog (`docs/4_CHANGELOG.md`).

## Decisioni recenti (ultime 10, dettagli negli ADR)

- **ADR-01:** Modularizzazione di `2_ARCHITECTURE.md` con scomposizione per domini in `docs/domains/` (`scanning_and_discovery.md`, `ssh_and_firewall.md`, `reporting_and_assets.md`, `debug_console.md`) per garantire efficienza di contesto e rispetto del limite di 400 righe.
- **ADR-02:** Preservazione totale della cronologia Git via `git mv` per file storici e introduzione del registro `docs/_legacy/MAPPING.md`.
- **ADR-03:** Creazione del branch dedicato `chore/agent-onboarding` per tutte le modifiche strutturali e documentali senza impattare il codice sorgente o la configurazione di runtime.

## Prossimi passi

1. Attendere feedback e validazione dell'utente sull'onboarding e risposta alle 3 domande aperte.
2. Procedere sul task approvato del backlog (`3_TASK_LIST.md`).

## Log sessioni (ultime 5, la più recente in alto)

- 2026-10-09 · Esecuzione `/onboard`: inventario, baseline test 100% verde (Go + Vitest + build), modularizzazione `docs/domains/`, creazione `MAPPING.md` e configurazione memoria viva `PROJECT_STATE.md` · Baseline VERDE (exit code 0)
