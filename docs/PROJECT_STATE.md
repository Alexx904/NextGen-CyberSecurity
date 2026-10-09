# PROJECT STATE

> Memoria viva dell'agente. Letta a inizio sessione, aggiornata a fine di ogni turno.
> Non cancellare sezioni: se vuote, scrivi "nessuno".

**Ultimo aggiornamento:** 2026-10-09 22:00  
**Modalità progetto:** esistente (onboarded il 2026-10-09, validato dall'utente)  
**Fase:** Manutenzione & Sviluppo Milestone 2  

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
| 2026-10-09 21:49 | `go test ./...` (backend) | PASS (exit code 0, 14 package) | Nessuno. Tutti i test passano al 100%. |
| 2026-10-09 21:49 | `go build ./...` (backend) | PASS (exit code 0) | Nessuno. Compilazione pulita. |
| 2026-10-09 21:49 | `go vet ./...` (backend) | PASS (exit code 0) | Nessuno. Zero anomalie rilevate. |
| 2026-10-09 21:50 | `npm test` (frontend Vitest) | PASS (exit code 0, 10 file, 65 test) | Nessuno. (Nota: warning informativi su act() in DebugConsole e React Router v7 future flags). |
| 2026-10-09 21:50 | `npm run build` (frontend tsc + Vite) | PASS (exit code 0) | Nessuno. (Nota: warning chunk size > 500kB su bundle principale, da ottimizzare con code-splitting). |

## In corso

- (nessuno)

## In attesa di accettazione utente

- (nessuno)

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
- **ADR-01:** Modularizzazione di `2_ARCHITECTURE.md` con scomposizione per domini in `docs/domains/` (`scanning_and_discovery.md`, `ssh_and_firewall.md`, `reporting_and_assets.md`, `debug_console.md`).
- **ADR-02:** Preservazione della cronologia Git via `git mv` per file storici e tracciamento in `docs/_legacy/MAPPING.md`.
- **ADR-03:** Creazione del branch dedicato `chore/agent-onboarding` per tutte le modifiche strutturali.

## Prossimi passi

1. Chiedere all'utente conferma per eseguire il merge di `chore/agent-onboarding` su `main`.
2. Implementare `ONBOARD-01` (Pipeline CI/CD GitHub Actions in `.github/workflows/ci.yml`).
3. Implementare `ONBOARD-03` (Code-splitting bundle frontend in `frontend/src/App.tsx` e `frontend/vite.config.ts`).
4. Avviare intake e system design per `FEAT-M2-01` (Attack Path Mapping).

## Log sessioni (ultime 5, la più recente in alto)

- 2026-10-09 · Task ONBOARD-03 accettato dal PO (`[x]`). Ottimizzazione bundle e code-splitting consolidati su main · Test e Build VERDI
- 2026-10-09 · Esecuzione ONBOARD-03: implementato code-splitting con React.lazy in App.tsx e manualChunks in vite.config.ts. Dimensione chunk principale ridotta da 1.2 MB a 91 kB, 0 warning Vite · Test 65/65 PASS
- 2026-10-09 · Task ONBOARD-01 accettato dal PO (`[x]`). Pipeline CI/CD GitHub Actions consolidata · Test VERDI
- 2026-10-09 · Esecuzione ONBOARD-01: creata pipeline GitHub Actions in `.github/workflows/ci.yml` per backend (Go 1.22 test/build/vet) e frontend (Node 20 vitest/build) · Sintassi YAML verificata
- 2026-10-09 · Merge branch `chore/agent-onboarding` su `main` e creazione branch feature `feat/ONBOARD-01-ci-cd` · Baseline VERDE
- 2026-10-09 · Validazione onboarding da parte del PO. Registrate scelte operative: CI/CD (Sì), Code-Splitting (Sì), Attack Path Mapping (Priorità Milestone 2) · Baseline VERDE
- 2026-10-09 · Esecuzione `/onboard`: inventario, baseline test 100% verde (Go + Vitest + build), modularizzazione `docs/domains/`, creazione `MAPPING.md` e configurazione memoria viva `PROJECT_STATE.md` · Baseline VERDE (exit code 0)
