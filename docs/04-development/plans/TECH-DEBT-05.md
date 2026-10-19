# TECH-DEBT-05 - Server-side inventory scan authorization

Alignment: 2026-10-19, working tree. Implementation, final automatic gates and real-browser checks complete; accepted by PO on 2026-10-19. No legal/regulatory compliance claim.

## Policy and scope

PO selected explicit consent for every inventory scan on 2026-10-19, including SAFE
profiles. Recurring consent applies only to the authorized scope and configuration.
Scope: manual target/group creation, the legacy group endpoint, scheduler lifecycle,
queued workers, transactional audit and scan/schedule UI. Standalone LAN discovery
retains its separate existing contract. No public-target support, application deploy,
notification send, credential access or real Nmap scan is included.

## Implemented contract

Creation requests include `authorization`: `accepted: true`,
`statement: "scan-authorization-v1"`, exact `targets: [{id,value}]`, `nmap_args`,
`timeout_seconds`, `cron_expression` (empty for manual scans). Actor/tenant come
from authenticated context and acceptance time from server UTC. Request-supplied
identity/time fields cannot override them. Consent does not bypass RBAC, blacklist,
network policy or profile ownership. Missing, stale, cross-tenant or mismatched
consent returns 409 `SCAN_AUTHORIZATION_REQUIRED` before creating covered work.

Group resolution loads every page and validates the whole eligible snapshot
before any job is created. Each job retains narrowed target evidence with its
original actor/time. Ordinary operational failures can still give partial group
results; queue delivery is not a transactional outbox or an idempotency guarantee.

Scheduler triggers reload the source row and validate live target/profile/cron;
legacy, disabled, deleted or stale schedules cannot authorize work. Workers compare
live target/profile with the job snapshot immediately before Nmap; recurring jobs
also recheck their source schedule. Configure all repositories before starting
workers. Revocation here means suspension/deletion of a schedule; an already-started
process is not recalled. Account-role changes do not independently rewrite consent.

Runtime metadata updates write only last/next-run fields conditionally on matching
enabled configuration; they cannot resurrect a disabled schedule. Browser QA exposed
a preexisting UTC/local mismatch in next-run display. A failing dedicated test proved
it; calculations now use the same server-local clock as cron and store UTC instants.

## Storage and compatibility

Migration `000013_scan_authorization` adds nullable `consent_snapshot JSONB` to
scans/schedules, serialized publicly as `authorization`. Old rows are not backfilled
with invented consent. Historical completed scans remain readable; legacy queued
scans fail before Nmap. Legacy/stale schedules require suspension and replacement
through an explicitly confirmed configuration. Responses expose current
`authorization_required`; the UI explains recovery and blocks reactivation.

A source-insert trigger writes `scan_authorization_audit` in the same PostgreSQL
transaction. The table stores tenant/user, object type/ID, database recorded_at and
the evidence JSON. Mutation/truncation of evidence and changes to source snapshots
are rejected. Inventory foreign keys are intentionally absent so evidence survives
scan/schedule/tenant removal. No automatic purge/retention period is introduced;
storage grows with authorizations. Database owners can remove protections: this
is application-level immutability, not external notarization or proof of ownership.

Down drops the new evidence table, columns, triggers and functions and loses that
history; back up the columns/table before application rollback. Actual up/down/up
checks use only disposable `nextgen_consent_test`, PostgreSQL 16 on 127.0.0.1:55432.
Preexisting deferred alert migration 000012 remains untouched and unpublished.

## UI and real browser evidence

The main scan dialog requires confirmation for every profile and invalidates it
when options, scope, timeout or recurrence change. Quick scans send the exact
confirmed target/profile; recursive group entry reuses the main dialog. Existing
loading/error/partial-result behavior remains. Disabled launch buttons are outside
the keyboard focus cycle. Schedule rows distinguish current consent from requested
activation status.

On 2026-10-19 the browser became available after an initially empty inventory.
The PO supplied `http://localhost:3080/`; it was opened read-only as the existing
application. Updated sources were checked separately through Vite at 3000 and a
real Go backend at 8081 against the disposable test DB. Synthetic account/company,
target, suspended legacy schedule and historical COMPLETED record were used.
No fixture record is evidence of an actual Nmap run. No external provider was used.

Observed: default-profile launch disabled until confirmation; profile/cron edits
invalidate it; a confirmed weekly schedule saves through the real API, creates one
audit record and can be suspended. Legacy consent-required row is non-reactivatable.
Desktop and 390px dialog render/scroll correctly. Login, dashboard, reports and
Settings theme/layout were inspected; final visual completion is recorded in the
dependency assessment. Browser screenshots are ignored artifacts in
`.agent/artifacts/TECH-DEBT-05/` (consent, schedules, reports, Settings and login).
An initial empty-history report error recovered after adding the synthetic history;
this does not establish report correctness on every empty-data state.

## Verification and self-review

Risk: security-sensitive identity/scope, persisted evidence and scheduled execution.
Required sides: backend + frontend, PostgreSQL, migration rollback and browser.

- Initial backend regression detected old positive fixtures without consent; those
  now explicitly authorize their original target/profile. Assertions are preserved.
- Domain tests reject absent consent, old statement, missing actor, duplicate scope,
  changed target/options/timeout/cron and cross-tenant profiles. Context copying and
  per-target narrowing preserve original acceptance identity/time.
- Service tests prove missing/stale consent creates no records or queue entries;
  workers reject missing/changed snapshots before Nmap. Scheduler tests cover legacy,
  disabled/deleted stale callbacks, recurrence edits and server-clock consistency.
- HTTP tests prove authenticated actor/time override untrusted body fields and
  both manual/group consent failures map to 409.
- PostgreSQL tests prove evidence round-trip/tenant isolation, atomic rejection,
  immutable evidence/source snapshot, audit survival after source deletion,
  schedule round-trip and metadata not restoring disabled state, plus 13 down/up.
- Frontend tests cover mandatory default-profile/quick-scan consent, exact payload,
  recurrence invalidation, discovery-selected targets and blocked legacy activation.

Targeted frontend: 34 tests PASS before four additional final cases. Dedicated
PostgreSQL consent test PASS (Go package 1.076 s, process 2.174 s). Earlier SQL
failures (reserved column name, then statement literal during renaming) were fixed
and retested on the disposable DB; no application database was migrated. The first
Quick runner invocation rejected a malformed package list; another unsupported
argument was rejected. Sandbox cache/filesystem failures were rerun with approval.

Final All/Integration report `20261019-130817-5cbdbed6`: exit 0; backend tests
24.249 s, vet 1.148 s, build 1.429 s; frontend 41 files / 233 tests 8.234 s,
typecheck/build 11.404 s; PostgreSQL 6.731 s; docs 0.682 s; diff check 0.078 s.
Initialization-order self-review fix then Final Backend PASS in `20261019-131043-b80beb8c`: tests 1.177 s (cached unchanged packages), vet 0.462 s, build 1.266 s; exit 0. Frontend and repository/migration inputs unchanged. No frontend lint command exists.

Self-review covers parameterized SQL, tenant/profile boundaries, identity derivation,
append-only evidence and rollback, worker/scheduler lifecycle, UI errors/focus and
preservation of preexisting work. This is not independent review. Canonical updates:
API/schema/domain/UI, ADR-27, path mapping and task/state; no unrelated dependency
or deferred alerting implementation is attributed to this task.


Final browser completion: desktop 1280px and mobile 390px login/dashboard/reports/Settings/modal PASS. Keyboard focus is visible; Shift+Tab wraps to Cancel, Tab to Close, Escape restores New Scan. Updated backend restarted after the final initialization correction; no Nmap call was made. QA servers and the disposable PostgreSQL container are stopped after checks. The existing 3080 application was not rebuilt or redeployed.
