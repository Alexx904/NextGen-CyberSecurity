# Mappatura Documentazione Legacy (Legacy Documentation Mapping)

Questo documento traccia la provenienza e la destinazione dei file storici del repository durante la migrazione alla struttura autonoma v6.0 (`/onboard`).

| File Originale | Percorso Legacy | Destinazione / Dove è confluito | Note sul Contenuto |
| :--- | :--- | :--- | :--- |
| `docs/Agent.md` | `docs/_legacy/Agent.md` | Sostituito da `AGENTS.md` (root) e `.agent/workflows/` | Regole operative, protocollo di verifica e definizioni comandi dell'agente. I principi cardine sono confluiti nelle regole centrali. |
| `docs/Brainstorming_CyberSec_App.md` | `docs/_legacy/Brainstorming_CyberSec_App.md` | `docs/0_REQUIREMENTS.md` e `docs/1_PROJECT_CONTEXT.md` | Visione iniziale del progetto, motivazione dello stack Go + Nmap + Postgres + React, e roadmap ad alto livello (Milestone 1, 1.1 e 2). |
| `docs/0_REQUIREMENTS_AND_USE_CASES.md` | `docs/0_REQUIREMENTS.md` (rinominato) | `docs/0_REQUIREMENTS.md` | Requisiti funzionali e use cases aggiornati per uniformità con la convenzione `0_REQUIREMENTS.md`. |

