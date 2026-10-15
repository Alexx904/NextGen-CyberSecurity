# Codex Playbook Registry

Codex scopre automaticamente `AGENTS.md`; i file in `.agent/workflows/` sono **playbook di progetto**, non slash command registrati. Un messaggio come `/verify` o “verifica tutto” indica a Codex di applicare `workflows/verify.md` nel rispetto di `AGENTS.md` e delle istruzioni di runtime.

## Routing

| Intento | Alias conversazionale | Playbook |
|---|---|---|
| Nuovo progetto | `/init` | [`workflows/init.md`](workflows/init.md) |
| Onboarding repository | `/onboard` | [`workflows/onboard.md`](workflows/onboard.md) |
| Nuova feature | `/feature` | [`workflows/feature.md`](workflows/feature.md) |
| Accettazione, KO o modifica | `/feedback` | [`workflows/feedback.md`](workflows/feedback.md) |
| Bug fix | `/fix` | [`workflows/fix.md`](workflows/fix.md) |
| Refactoring | `/refactor` | [`workflows/refactor.md`](workflows/refactor.md) |
| Migrazione database | `/db-migrate` | [`workflows/db-migrate.md`](workflows/db-migrate.md) |
| Verifica completa | `/verify` | [`workflows/verify.md`](workflows/verify.md) |
| Review ostile | `/review` | [`workflows/review.md`](workflows/review.md) |
| Sincronizzazione documenti | `/sync-docs` | [`workflows/sync-docs.md`](workflows/sync-docs.md) |
| Audit periodico | `/audit` | [`workflows/audit.md`](workflows/audit.md) |
| Stato progetto | `/status` | [`workflows/status.md`](workflows/status.md) |
| Preparazione commit | `/commit` | [`workflows/commit.md`](workflows/commit.md) |
| Preparazione release | `/release` | [`workflows/release.md`](workflows/release.md) |

## Ruoli e ownership

Il Tech Lead orchestra secondo [orchestration.md](workflows/orchestration.md); il PO umano mantiene autorità sul prodotto. I ruoli sono attivati in base alla feature e non richiedono sette processi permanenti.

| Ruolo | Istruzioni |
|---|---|
| Business Analyst | [roles/business-analyst.md](roles/business-analyst.md) |
| UI/UX Designer | [roles/ui-ux-designer.md](roles/ui-ux-designer.md) |
| Software Architect | [roles/software-architect.md](roles/software-architect.md) |
| Lead Developer (Backend/Frontend) | [roles/lead-developer.md](roles/lead-developer.md) |
| QA Engineer | [roles/qa-engineer.md](roles/qa-engineer.md) |
| DevOps Engineer | [roles/devops-engineer.md](roles/devops-engineer.md) |
| Technical Writer | [roles/technical-writer.md](roles/technical-writer.md) |

## Regole comuni

1. Leggi integralmente il playbook selezionato prima di agire.
2. I playbook non ampliano i permessi: conferme, sandbox e sicurezza di `AGENTS.md` restano vincolanti.
3. Se più playbook sono applicabili, usa il minimo insieme e dichiarane l'ordine.
4. I playbook descrivono risultati e gate, non nomi rigidi di tool: usa le capacità realmente disponibili nel runtime Codex.
5. Aggiorna questo indice quando aggiungi, rinomini o rimuovi un playbook.
6. `docs/` ospita solo documentazione Markdown necessaria agli agenti. Salvare screenshot e altre evidenze temporanee in `.agent/artifacts/<task-id>/` (ignorato da Git); non ricreare manuali, archivi o stub eliminati.
