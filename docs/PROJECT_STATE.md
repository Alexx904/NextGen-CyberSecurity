# PROJECT STATE

**Ultimo aggiornamento:** 2026-10-19 · working tree.
Current tasks: [TASK_LIST](04-development/TASK_LIST.md); prior evidence: [history](../.agent/history/PROJECT_STATE-2026-10-18.md). History grants no new authorization.

## Project Profile

NextGen-CyberSecurity: Go 1.22+ / Chi / pgx, React 18 / TypeScript / Vite; PostgreSQL 16 e Nmap nel container. Fonti operative: [Backend](04-development/BACKEND_SETUP.md), [Frontend](04-development/FRONTEND_SETUP.md), [Test Plan](05-qa/TEST_PLAN.md#strategia).
Runner root: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1`, Quick/Final/Full/Docs. Go rapido esclude DB; integration richiede TEST_DB_* esplicite e DB locale dedicato. Docker PO 3080, Vite 3000, QA separata 3081.
Preferenza PO persistente: lavorare su main senza nuovi branch; commit/push/merge/deploy richiedono scope esplicito. Lavoro preesistente M1.1 preservato; PO authorized commit/push of accepted deliveries on main on 2026-10-18; deferred T4/T5 code excluded.

## Baseline test (prima di ogni intervento)

Consultare, non rilanciare automaticamente; la freschezza dipende dagli input effettivi, non solo da HEAD.
- M4: All/DB + 239 frontend tests/build PASS (`20261019-141434-e9fde59f`); final backend PASS (`20261019-141915-3841cb57`), browser/HTTP/PDF PASS. Evidence/limits in the delivery below.
- TECH-DEBT-05/09: Final All/Integration PASS, 233 frontend tests/typecheck/build plus backend/DB/up-down (20261019-130817-5cbdbed6); initialization-order correction then Final Backend PASS (20261019-131043-b80beb8c). Real browser desktop/390px PASS on QA 3000/8081 and disposable nextgen_consent_test 55432. QA services stopped; 3080 deployment untouched. No real scans or provider sends.
- Publication snapshot without deferred alerting: backend test/vet/build, 220 frontend tests/build and PostgreSQL/up-down PASS in `20261018-194107-662e86c9`. Docs mapping correction verified separately; no application changes after this gate.
- DB/000010 up/down/up storico nel [piano SSH](04-development/plans/SSH-HARDEN-01.md), non nuovo PASS della working tree.
- AGENT-HARNESS-01: Docs/22 prove/preservazione PASS; report `20261018-114735-c1bb6cad`, dettagli nel piano. Nessuna suite applicativa attribuita.

## In corso

M4-AUTO-01: automatic context; app gates/250 tests/PDF/Docs PASS, browser check pending. [Evidence](04-development/plans/M4-AUTOMATIC-CONTEXT.md). PO deploy.

- **E28-S1-T4/T5 deferred until project end by PO (2026-10-18)**. Code preserved locally and excluded from publication. Both remain `[ ]`; browser/live gates pending. [Plan](04-development/plans/E28-S1-T4-alerting.md).

## Accepted by PO - 2026-10-18

All previous pending deliveries accepted: AGENT-HARNESS-01, M11-CLOSE-01, M11-DISCOVERY-UI-01, M11-REPORT-01, M11-FIREWALL-01, M11-PERSIST-01, CI-CONTEXT-01, README-GH-01, SSH-HARDEN-01, E28-S1-T1, E28-S1-T2, E28-S1-T3. ROADMAP-PENTEST-01 planning accepted too. IDs, evidence and limits archived in [CHANGELOG](../CHANGELOG.md#accepted-tasks-and-priority-reorganization--2026-10-18).

## Accepted by PO - 2026-10-19

- **TECH-DEBT-05 [x]**: mandatory inventory scan consent, recurring scope/configuration binding and transactional immutable audit; API/service/worker/scheduler/DB up-down/browser PASS. [Evidence and compatibility](04-development/plans/TECH-DEBT-05.md).
- **TECH-DEBT-09 [x]**: dependency remediation plus real Chromium desktop/390px compatibility and keyboard gates PASS. [Assessment/completion](05-qa/TECH-DEBT-09.md#browser-completion---2026-10-19).
PO explicitly accepted both deliveries with `OK 05/09`; test evidence and limitations remain unchanged.

UI-VPATCH-REMOVE-01 [x]: button removal accepted with `OK`; evidence in TASK_LIST. No deploy.

## In attesa di accettazione utente

Manual M4/E29, GUIDE and WORKFLOW deliveries are superseded, not accepted; historical states/evidence remain in their plans/task list. M5 not started.

ROADMAP-COHERENCE-01 [~]: M4-M8 documentation; Docs PASS (20261019-133714-bf03fe56), self-review/Doc Sync complete, awaits PO. T4/T5 deferred/unpublished.

## Bloccati / Domande aperte per l'utente

- TECH-DEBT-07 mitigation accepted with M11-FIREWALL-01; residual limits retained. Fixture firewall .3/SSH .4, rete interna `nextgen-m11-lab`; backend .2, DB test 55432. Limiti host/UID/stato temporaneo nel piano.
- SSH-CONNECT-01 è una diagnosi storica superata dalla connessione confermata dal PO; il relativo `[!]` nella task list è conservato come storia, non richiede ripetere vecchie domande.
- TECH-DEBT-08 modello Attack Path, 10 costo archivio report e debiti 01..04 restano nel registro task; 05/09 accettati dal PO il 2026-10-19.

## Assunzioni confermate

Azienda→reparto a due livelli nelle nuove scritture; SSH/clienti indipendenti dallo scope aziende; Attack Path euristico 2D, non prova di compromissione; Executive con cinque priorità per reparto; discovery non recupera identità storiche senza MAC verificabile.
DOCKER-CI: Linux amd64, artifact 7 giorni, PostgreSQL separato, nessun registry/deploy. Firewall PO: iptables/iptables-nft con modello corrente, nessun motore nft nativo. Pin host fornito tramite canale indipendentemente fidato.
Assunzione AGENT-HARNESS-01: skill condivise a livello repository, nessuna modifica a config personale o dipendenze; beneficio comportamentale su Sol da misurare.

## Decisioni recenti (ultime 10, dettagli negli ADR)

- ADR-30 replaces manual M4 with automatic finding context; retains data. [Decision](03-architecture/ADR/README.md#adr-30---automatic-scan-based-regulatory-context-replaces-manual-compliance-2026-10-19).

- ADR-28: shared M4-M8 lifecycle/reporting, design first in M4; SSH independent. [Decision](03-architecture/ADR/README.md#adr-28--integrated-assessment-lifecycle-and-reporting-2026-10-19).

- ADR-27: explicit inventory scan consent, immutable transactional snapshots and live scheduler/worker checks; legacy schedules require explicit replacement. [Decision](03-architecture/ADR/README.md#adr-27--inventory-scan-authorization-snapshots-2026-10-19).
- ADR-26: outbox alert transazionale, retry/lease, config cifrata, destinatari pubblici TLS e scope/versioni; limiti at-least-once. [Decisione](03-architecture/ADR/README.md#adr-26--transactional-alert-outbox-and-restricted-egress-2026-10-18).
- ADR-25: transazione firewall sull'host, lock per UID, watchdog senza kill PID, backup su stdin e revisioni editor; limiti reboot/amministratori esterni espliciti.
- ADR-24: snapshot PostgreSQL discovery e callback stdout host/finding; scrittura prima di SSE, replay tenant e limiti restart/pub-sub locale.
- ADR-23: contesto progressivo, skill specifiche e verificatori deterministici; [decisione](03-architecture/ADR/README.md#adr-23).
- ADR-22: archivio report completo via paginazione esistente; costo lineare esplicito.
- ADR-21: timer firewall obbligatorio, input validati e commenti quotati.
- ADR-20: pin host separato e deadline SSH; nessuna garanzia remota dedotta.

## Prossimi passi

UI-TOOLS-01: post-login tool hub planned; future features must justify placement in existing/independent tools. [Scope and gates](04-development/TASK_LIST.md#p4---technical-debt).

E28-S1-T7: P4 replacement of virtual patch generator with external mitigation sources, without DB growth/sync; implementation pending. [Scope](04-development/TASK_LIST.md#p4---technical-debt).

Stop at M4; M5-M8 pending. No install, deploy or external/live-target testing authorized.
M4 -> M5 -> M6/M7/M8 -> measured debt -> T4/T5 -> T6. [Queue](04-development/TASK_LIST.md). No deploy/external send authorized; T2/T3 QA remains limited to 3081/synthetic fixtures.

## Log sessioni

- 2026-10-19 - M4 implemented and verified; All/DB, browser desktop/390px/Auditor and rendered Audit PDFs PASS. No M5 work, application deploy, scans or external send.

[Earlier session summaries](../.agent/history/PROJECT_STATE-2026-10-19-before-M4.md).
