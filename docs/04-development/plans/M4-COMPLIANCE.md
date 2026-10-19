# M4 - Compliance implementation and verification

Historical delivery: manual UI/API/Audit are superseded by PO's selection of
[automatic finding context](M4-AUTOMATIC-CONTEXT.md). Evidence below describes the
earlier implementation, not the current user workflow.

Updated: 2026-10-19 - working tree. Extends [the issue contract](E29-S1-T1.md)
and ADR-28. PO requested completion of M4; M5 and deferred alerting are excluded.

## First-time user guide: workflow and interpretation

The Compliance page includes a collapsible first-use guide above the scope filters,
available to readers and writers even when assessment loading fails. Future
regulatory maintenance is specified in the [proposed update design](M4-REGULATORY-UPDATES.md);
monitoring and automatic activation are not implemented.

Guide follow-up verification: root `scripts/verify.ps1 -Profile Final -Scope Frontend`,
report `20261019-190633-b8be16e0`: 239 tests PASS (6.171s), TypeScript/Vite build
PASS (10.872s), Docs PASS (0.673s), diff check PASS (0.056s), all exit 0.
Self-review: native keyboard-operable disclosure, static text without new input,
network calls or dependencies; existing roles and assessment behavior preserved.
The session had no connected browser. PO took responsibility for visual verification
and requested marking it done, with feedback if layout is wrong; this is not an
agent-observed browser PASS. No deployment was performed.

Open **Compliance & Remediation**, select the company and, if needed, a department,
then choose a framework. Read the score together with assessed coverage, unknown
controls and exclusions: the percentage measures recorded evidence against the
configured controls, not legal compliance or certification. An empty assessment
starts with unknown controls; it does not prove either compliance or a breach.
Satisfied controls earn one point, partial controls half a point, and unknown or
not-satisfied controls zero. An explicitly justified not-applicable assessment
removes that control from the denominator. For example, one satisfied, one partial
and one unknown control produce 50%, with two of three controls assessed. These
percentages cover only the curated controls implemented for the selected framework.

For each control, review the technical findings and record an assessment with its
source, version and reason. An open technical gap cannot be hidden by a positive
manual assessment. A completed scan or an absence of findings does not establish
full coverage or prove that a control is satisfied. To track remediation, register
a stable asset and explicitly associate the relevant finding with it; reuse that
asset for later observations even if its IP changes. Association is an operator
action, not automatic matching. Set an owner and SLA, record mitigation, then
declare correction. Mark correction as verified only with a referenced verification
covering the issue and confirming its absence. This action records your attestation;
it does not run a new test. A later associated positive observation reopens the
issue while retaining its history. The deadline runs from the first positive
observation in the current cycle; mean MTTR uses completed, verified cycles only,
so a low mean does not imply that unresolved issues have been fixed.

Use **Refresh** to retrieve current recorded data. Auditors can inspect history and
export, while authorized SecOps/admin users can also record changes. From Compliance
or Reports, export the Audit PDF with the legal entity, reviewer and review period.
It is a consistent snapshot of the current assessment and its supporting history;
the declared period labels the review and does not filter historical records.
Review unknowns, exclusions, open gaps and source versions before sharing it. The
PDF is an unsigned review dossier, not an independent certification. Compliance
scores and the technical Posture Score measure different things.

**Regulatory updates are not automatic.** The application ships a curated,
versioned mapping catalogue (currently `2026-10-19.1`); it does not monitor EU
publications, download new regulations or automatically adopt amendments. Updating
the catalogue requires reviewing the relevant official sources, revising the
implemented controls/mappings and their version, testing the changes and releasing
the updated application. Existing manual evidence must then be reviewed for
continued relevance. Refreshing the page recalculates the assessment from recorded
data using the installed rules; it does not update those rules or determine which
legal obligations apply to the organization. ISO/IEC 27001 is a separate standard,
not an EU regulation.

## Contracts before implementation

- T1: persistent issues use the existing domain aggregate. Stable assets are
  explicitly registered in a tenant/company or department and linked to selected
  finding observations by an operator. Neither IP nor target-range UUID silently
  identifies a host. Reuse the asset UUID for subsequent observations or address
  changes; ambiguous/unlinked observations stay visible and excluded from MTTR.
- T2: additive migration 000014 creates assets, unique issue identities,
  append-only ordered issue command/evidence events and control assessments.
  Transactions lock an issue before replay/appending; expected version prevents
  lost updates. Source finding/scan/target/group tenant and asset group must agree.
  Historical import cannot override newer lifecycle events. No automatic closure
  on absence. No migration of historical findings or rewrite of old migrations.
- T3: versioned curated mappings for NIS2 Art.21(2)(a-j), DORA ICT risk/testing
  controls and ISO 27001:2022 A.8.8/A.8.20/A.8.24. Mapping is technical evidence
  relevance, never legal applicability or certification. Manual evidence supplies
  unknown/satisfied/partial/not_satisfied/not_applicable states. Current open
  technical gaps prevent a satisfied manual claim from hiding the gap.
- T4: `/compliance` extends assessment navigation, scoped by the global company
  selector and optional department. Server validates root/department membership
  in the authenticated tenant. Existing tenant roles have tenant-wide read access;
  company selection is a scope filter, not a new entitlement system. SecOps/admin
  may write; Auditors read/export. Client and server both enforce this distinction.
  All configured controls remain in the denominator except explicit evidenced
  not-applicable decisions. Unknown contributes zero credit; partial half credit;
  also show assessed coverage, exclusions and per-framework/per-department totals.
- T5: explicit per-issue versioned SLA duration/owner configuration and lifecycle
  actions; elapsed current-cycle MTTR only for verified corrections. Mean MTTR
  averages completed cycles retained in history, not open/unresolved observations.
  Breaches appear in-app; no external notification. Manual verification requires
  issue-scoped source/version/reference, full coverage and explicit not-present
  result, actor and reason; it is an operator attestation, not a new scan executor.
- T6: Audit PDF through the existing report package and `/reports` route family,
  downloadable from Reports and Compliance. One repeatable-read dataset supplies
  scope names, current assessments, source observations, issue transitions,
  source/rule versions, cutoff and limitations. No parallel report/finding archive.
  Legal entity, reviewer/signatory and period are explicit export inputs; unsigned
  content is labeled as such. Export is a review dossier, never a certificate.

API: `GET /compliance?company_id=&department_id=` returns one complete snapshot;
`POST /compliance/assets` registers identity; `POST /compliance/observations`
links a persisted finding to that asset; `POST /compliance/issues/{id}/events`
accepts action, expected_version, reason and action-specific configuration or
verification. `POST /compliance/evidence` appends a control assessment.
`POST /reports/audit` exports the same snapshot semantics to PDF. UUIDs, enum,
size and positive duration validation precede writes. Missing auth 401, denied
role 403, out-of-tenant/scope resources 404, invalid input 400, stale version or
observation identity conflict 409; internal errors never expose SQL.

Schema: assets unique `(tenant_id,id)` and `(tenant_id,group_id,name)`; issues
unique complete identity and composite asset FK; events primary key
`(tenant_id,issue_id,version)` and immutable command JSONB. Evidence references
tenant/group and actor/source/rule version. Explicit RESTRICT preserves history
when referenced assets/groups/tenants are deleted. No ON DELETE CASCADE on audit.
`created_at` is server assigned; append-only records have no mutable updated_at.
Down removes only these new tables/triggers/functions and loses their data;
application rollback requires backing up all four tables first. Rollback tests
run only in a fresh disposable M4 database, never application data.

## Gates

Risk: cross-tenant/source confusion, optimistic concurrency, false compliance,
stale historical observations, duplicate/reopening errors and inconsistent PDF.
Targeted domain/service/handler tests, real PostgreSQL isolation/atomicity and
000014 up/down/up, meaningful frontend behavior/error/scope tests. Final:
`scripts/verify.ps1 -Profile Final -Scope All -Integration` with explicit
TEST_DB_* against disposable localhost M4 PostgreSQL. Real browser desktop/mobile,
read-only Auditor, evidence/issue/SLA/export flows; render representative Audit
PDF pages, including empty and long content. Security self-review and Doc Sync.
Use workspace Go cache. No application deployment, real scan or external send.

## Regulatory references

- [NIS2 official text](https://eur-lex.europa.eu/eli/dir/2022/2555/oj/eng)
- [DORA official text](https://eur-lex.europa.eu/eli/reg/2022/2554/oj/eng)
- [ISO official catalogue](https://www.iso.org/standard/27001)

Curated rule version `2026-10-19.1`. Paraphrases and control identifiers only;
no reproduction of the ISO standard. No jurisdiction-specific legal assessment.

## Execution evidence

M4 / E29-S1-T1..T6 implemented and verified on 2026-10-19, awaiting PO
acceptance (`[~]`). PO asked to stop at M4; no M5 implementation started.

### Automated gates

Root runner: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1`.
Go cache was workspace-local `.agent/artifacts/go-build-cache`; Node/esbuild
required the approved sandbox elevation. Explicit TEST_DB_* identified PostgreSQL
16 at `127.0.0.1:55436`, user `m4test`, disposable `nextgen_m4_gate_test`.
Browser/HTTP fixtures used a different database, `nextgen_m4_test`, in the same
new container. No rollback ran against the QA service database or application DB.

Final All/Integration report `20261019-141434-e9fde59f`:

| Directory | Command | Duration | Exit / result |
|---|---|---|---|
| backend | `go test ./...` | 24.737s | 0 / PASS |
| backend | `go vet ./...` | 0.753s | 0 / PASS |
| backend | `go build ./...` | 1.378s | 0 / PASS |
| frontend | `npm.cmd test` | 5.702s | 0 / PASS, 239 tests / 42 files |
| frontend | `npm.cmd run build` | 10.658s | 0 / PASS, TypeScript + Vite |
| backend | `go test -tags=integration -count=1 ./internal/repository/postgres` | 7.991s | 0 / PASS |
| root | docs verifier / `git diff --check` | 0.636s / 0.061s | both 0 / PASS |

Security/provenance review refinements include source locking, immutable source-conflict
checks, original host/target/group/confidence/method snapshots, protected-scope
deletion assertions and scan profile/authorization provenance. The group-deletion
409 message now explains protected compliance references as well as channels.
Supplemental backend gate `20261019-141915-3841cb57`, same runner with
`-Profile Final -Scope Backend -Integration`, is authoritative for final backend:
tests 21.280s, vet 0.774s, build 1.312s, DB 7.543s, docs 0.769s, diff 0.055s;
every exit 0. Intermediate backend gate `20261019-141609-c765aae0` also passed.
Frontend sources/tests/dependencies/configuration remained unchanged after its
All gate, so its 239 tests/build are reused without attributing new browser/DB
results to them. Reports are in ignored `.agent/artifacts/testing/`.

Integration assertions: tenant/group isolation, same-group source binding,
idempotent replay, changed-source/rebinding conflict, simultaneous same-version
writes (one success, one 409), no partial writes after invalid verification,
old observations rejected after remediation, changed-IP recurrence preserving
identity/history, original source surviving source deletion, immutable audit and
referenced group deletion protection. Migration 14 down removes its schema;
up recreates an empty stream, never fabricated history. Earlier migration
round-trip suites also passed on the dedicated database.

### Real HTTP, browser and PDF

QA: Vite `127.0.0.1:3002`, backend 8082, new disposable PostgreSQL on 55436.
Synthetic accounts, companies, targets and persisted completed scan fixtures only;
no Nmap job, remote retest, exploit, notification/provider send or application
deployment. Existing 3080/8080/5432 services were not modified or restarted.

- Real HTTP PASS: asset registration, observation association/replay, configured
  SLA breach, manual evidence unable to hide an open gap, mitigation/declaration/
  referenced verification and completed-cycle MTTR. Auditor read PASS, write 403.
  Ignored scripts: `.agent/artifacts/M4/http-smoke.ps1` and `export-pdfs.ps1`.
- Connected Chromium desktop PASS: company/reparto selection, denominator display,
  evidence input (ISO score 33.3% to 50% after an evidenced partial control),
  asset registration and association, named owner/one-hour SLA with active breach,
  distinct mitigation/correction/verification actions, stored provenance and MTTR,
  persistence after refresh, shared Audit form/download from Reports.
- Browser Auditor PASS: read/evidence/history/export UI available; evidence/asset
  writes and issue update buttons absent. Separate real HTTP role enforcement
  and handler tests prove this is not only a visual restriction.
- Empty-company switch PASS: 0/28 assessed, 28 unknown, no issues/cycles or leaked
  previous-company content. Return to populated company restored correct data.
- Mobile 390x844 PASS: Compliance score/table and Reports Audit form readable;
  document width 375 with a 390px viewport (scrollbar included), no page overflow.
  Keyboard Tab from Framework reached Refresh with visible cyan outline.
  Temporary viewport override reset. No captured TypeError in the inspected log.
- Real Audit export rendered through Windows PDF WinRT: empty 3 pages, long
  evidence/unbroken references 6 pages, browser-downloaded updated dossier 7 pages.
  All pages rendered; cover/checklist, wrapped evidence, source versions/IDs,
  policy/history, current and mean MTTR, cutoff and footers visually inspected.
  No overlap, clipping or blank trailing pages detected. Rendering used built-in
  Windows APIs because Python/Poppler were unavailable; no dependency installed.

Evidence: ignored `.agent/artifacts/M4/`, including `browser-desktop.jpg`,
`browser-mobile.jpg`, `browser-sla.jpg`, `browser-final.jpg`, `audit-empty.pdf`,
`audit-long.pdf`, `browser-audit.pdf` and rendered page PNGs. The browser gate
initially lacked a connected browser; the PO connected it and the blocker was
resolved. A long browser/download wait reset the tool session; final mobile and
Auditor checks resumed against the unchanged sources, with successful token refresh.

### Review, synchronization and limits

Security self-review, not independent review: scoped joins/foreign keys, trusted
actor/time, locked source identity, atomic expected-version events, immutable
history, input/body limits, generic error responses, read-only role, UI stale
responses and single-snapshot export inspected. No unresolved blocking finding.
Source locking/conflicting-source checks and protected group conflict text were
corrected during review; tests and final backend gate cover final sources.

Doc Sync preserved the eleven preexisting roadmap edits and prior domain/SSH
increment. Updated product intent/use case/UI/HLD/ADR-29, actual path ownership,
API/schema, compliance/reporting domain, QA/task/state for implemented behavior.
Earlier session summaries were preserved in history to keep L0 compact. No
verifier/routing/governance migration, branch, commit, push or deployment occurred.

Limits: curated framework subsets and manual attestations, explicit identity
resolution, no automatic evidence expiry or cross-department asset transfer,
no digital signature/certification, declared review period rather than historical
filter, no stored snapshot replay token, and no external SLA notifications.
Whole-history memory/query/export growth is tracked as TECH-DEBT-11 with a
measurement/materialization plan. These limits are documented in the domain, with applicable cautions in the UI/PDF;
no unexecuted mandatory gate is being counted as PASS.

Cleanup: verified and stopped only the recorded M4 backend/Vite processes and
`nextgen-m4-test-20261019` container. Databases/evidence retained, no drop/removal;
the agent-created browser tab was closed and its viewport reset. Application
services were untouched. Future DB checks require restarting the dedicated test
service; the historical PASS is not evidence of current service availability.
Final source/dependency SHA-256 manifest: ignored `M4/final-inputs.json`; no
application/dependency/config/test changes after the applicable final gates.

Recorded toolchain: Go 1.27.2 windows/amd64, Node 24.15.0, npm 11.12.1.
Size inventory at delivery: 18 untracked application/test files, 2489 lines,
including the preexisting domain increment; tracked wiring/helpers are separate.
Post-cleanup Docs PASS (`20261019-185626-3ebb8f63`, root, verifier 0.794s,
diff 0.062s, both exit 0); the L0 size advisory was then compacted without removing
invariants, followed by the final Docs recheck. No application suite was rerun
after Markdown-only delivery updates.
