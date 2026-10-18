# Documentation Index

Leggere prima [PROJECT_STATE](PROJECT_STATE.md) e il solo task pertinente in [TASK_LIST](04-development/TASK_LIST.md). Il contratto operativo è [AGENTS.md](../AGENTS.md); procedure e ruoli sono in [.agent](../.agent/README.md).

| Fase | Fonti canoniche | Responsabile |
|---|---|---|
| Prodotto | [PRD](01-product/PRD.md), [requisiti](01-product/REQUIREMENTS.md), [use case](01-product/USE_CASES.md), [storie](01-product/USER_STORIES.md) | Business Analyst; PO per decisioni di prodotto |
| UX/UI | [Flussi](02-design/UX_FLOW.md), [specifiche](02-design/UI_SPECIFICATIONS.md) | UI/UX Designer |
| Architettura | [HLD](03-architecture/HLD.md), [API](03-architecture/LLD/API_SPEC.md), [database](03-architecture/LLD/DATABASE_SCHEMA.md), [componenti](03-architecture/LLD/COMPONENT_DESIGN.md) | Architect e Lead Developer |
| Tracciabilità | [Mappa domini](03-architecture/DOMAIN_MAP.md), [ADR](03-architecture/ADR/README.md) | Architect |
| Sviluppo | [Backend](04-development/BACKEND_SETUP.md), [frontend](04-development/FRONTEND_SETUP.md), [task/debito](04-development/TASK_LIST.md), [piani](04-development/plans/README.md) | Lead Developer |
| QA | [Test plan](05-qa/TEST_PLAN.md) | QA Engineer |
| DevOps | [Deployment](06-devops/DEPLOYMENT.md), [CI/CD](06-devops/CI_CD_PIPELINE.md) | DevOps Engineer |
| Storico | [Changelog](../CHANGELOG.md) | Technical Writer |

## Lettura e fonti di verità

- L0: stato e task corrente. L1: dominio interessato da [DOMAIN_MAP](03-architecture/DOMAIN_MAP.md). L2: sezioni pertinenti dei documenti di fase.
- Requisiti: comportamento voluto; contratti: API attese; migrazioni: evoluzione eseguibile dello schema; codice: comportamento implementato; test: evidenza verificata. Una divergenza richiede analisi, non prevalenza automatica.
- I domini conservano invarianti e dettagli verticali; ogni definizione ha una sola sede, collegata dalle altre fonti. Le quattro [skill](../.agent/README.md#routing) possiedono procedure specializzate; ruoli e documenti non ne duplicano i passi.
- Stato corrente compatto; [snapshot storico](../.agent/history/PROJECT_STATE-2026-10-18.md) solo per ricostruire evidenze/accettazioni precedenti. Aggiornare lo stato per cambi persistenti, non a ogni turno; nessuna autorizzazione storica si estende a nuove operazioni.
- [DOMAIN_MAP](03-architecture/DOMAIN_MAP.md) è la tabella eseguibile dei percorsi, inclusi core/test/untracked. Il verificatore rifiuta assenza di mapping, glob obsoleti e ownership ambigue; non certifica accordo semantico fra codice e requisiti.
- FR/UC e stati di accettazione sono conservati. Nessun documento migrato costituisce nuova prova di conformità del codice.
- `docs/` contiene solo documentazione Markdown necessaria agli agenti. Manuali utente, screenshot, binari, archivi e stub obsoleti sono esclusi.
- Le evidenze visuali temporanee si salvano in `.agent/artifacts/<task-id>/`, fuori dalla documentazione canonica e ignorate da Git. I vecchi documenti numerati sono stati rimossi: usare esclusivamente i percorsi indicizzati sopra.

Ultimo allineamento: 2026-10-18 · working tree · AGENT-HARNESS-01.
