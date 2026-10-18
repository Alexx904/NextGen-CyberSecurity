# Codex workflow registry

Ultimo allineamento: 2026-10-18 · working tree.

`AGENTS.md` è il contratto sempre attivo. I playbook sono procedure del repository; gli alias sotto non registrano slash command. Le quattro skill in `.agents/skills/` usano discovery nativa Codex; se non esposte dal runtime leggere il relativo SKILL.md direttamente. Una skill e il playbook di ingresso condividono una sola procedura, senza caricarla due volte.

## Routing

| Intento / alias | Ingresso |
|---|---|
| Nuovo progetto / init | [init](workflows/init.md) |
| Repository esistente / onboard | [onboard](workflows/onboard.md) |
| Nuova capacità / feature | [feature](workflows/feature.md) |
| OK, KO, modifica / feedback | [feedback](workflows/feedback.md) |
| Bug / fix | [fix](workflows/fix.md) |
| Refactor | [refactor](workflows/refactor.md) |
| DB / db-migrate | [db-migrate](workflows/db-migrate.md) → [postgres-migration](../.agents/skills/postgres-migration/SKILL.md) |
| Verifica / verify | [verify](workflows/verify.md) → [repo-verification](../.agents/skills/repo-verification/SKILL.md) |
| Review | [review](workflows/review.md) → [security-review](../.agents/skills/security-review/SKILL.md) |
| Doc Sync | [sync-docs](workflows/sync-docs.md) → [documentation-sync](../.agents/skills/documentation-sync/SKILL.md) |
| Audit periodico | [audit](workflows/audit.md) |
| Stato / status | [status](workflows/status.md) |
| Commit esplicito | [commit](workflows/commit.md) |
| Release esplicita | [release](workflows/release.md) |

Leggere il solo ingresso pertinente e le estensioni richieste dal rischio. Un'analisi non autorizza implementazione; le procedure non ampliano i permessi. Non creare branch se la preferenza PO o lo stato Git lo escludono.

## Ruoli e ownership

[Orchestration](workflows/orchestration.md) contiene gate e handoff comuni. Consultare i ruoli solo per vincoli/output distintivi: [Business Analyst](roles/business-analyst.md), [UI/UX](roles/ui-ux-designer.md), [Architect](roles/software-architect.md), [Developer](roles/lead-developer.md), [QA](roles/qa-engineer.md), [DevOps](roles/devops-engineer.md), [Writer](roles/technical-writer.md). Nessuna attivazione automatica di subagent.

## Manutenzione

Una fonte per fatto: [indice documenti](../docs/README.md). Stato corrente compatto; [storia](history/PROJECT_STATE-2026-10-18.md) solo per ricostruzione, mai nuova autorizzazione. Evidenze temporanee in artifacts ignorati, esclusi dai verificatori canonici.

Il [piano harness](../docs/04-development/plans/AGENT-HARNESS-01.md) documenta fonti ufficiali, misure e limiti. `scripts/verify-agent-docs.ps1` controlla link, skill, stato e mapping; `scripts/test-agent-harness.ps1` ne verifica i rifiuti su fixture isolate. Un PASS strutturale non prova efficacia del modello o attivazione delle skill.
