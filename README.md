# NextGen-CyberSecurity

Piattaforma Go/React per ricognizione LAN, scansioni Nmap, analisi vulnerabilità, report e amministrazione difensiva remota multi-tenant. Stato delle funzionalità e roadmap: [PROJECT_STATE](docs/PROJECT_STATE.md).

## Quickstart locale

Con Docker e Docker Compose disponibili, dalla root:

```powershell
docker compose up --build
```

UI: http://localhost:3000 · API health: http://localhost:8080/api/v1/health. Configurazione, prerequisiti e limiti sono in [DEPLOYMENT](docs/06-devops/DEPLOYMENT.md). Usare un ambiente locale di sviluppo con valori propri; consultare `.env.example` senza condividere credenziali.

Per sviluppo nativo: [backend](docs/04-development/BACKEND_SETUP.md) e [frontend](docs/04-development/FRONTEND_SETUP.md).

## Documentazione e collaborazione

- [Indice documenti](docs/README.md): prodotto, UX, architettura, sviluppo, QA e DevOps.
- [Task e debito tecnico](docs/04-development/TASK_LIST.md).
- [Changelog](CHANGELOG.md).
- [Contratto operativo Codex](AGENTS.md) e [ruoli/playbook](.agent/README.md).

I test e i criteri di verifica sono in [TEST_PLAN](docs/05-qa/TEST_PLAN.md); gli esiti reali sono registrati nello stato del progetto.
