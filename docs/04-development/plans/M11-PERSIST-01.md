# M11-PERSIST-01 · Durable discovery snapshots

2026-10-18 · working tree. Scope: FR-12.1..3 durable discovery recovery and
host callbacks during Nmap, including incremental vulnerability findings.

## Contract before implementation

- PostgreSQL is authoritative in the server wiring. Existing memory-only service
  construction remains available for isolated tests, not as a database-error fallback.
- Migration 000011 adds tenant-owned discovery sessions and devices. Session IDs
  cannot be reused, including across tenants; insertion conflicts before scanning.
  All reads and writes include tenant ID. Device IPs are unique per session.
- Save the session before starting Nmap, each assembled device before its SSE
  notification, and completion before the terminal event. Database errors fail the
  request; never acknowledge an uncommitted device. Partial data survives failure.
- Existing REST/SSE payloads remain compatible. Last-results includes the newest
  started session, even when incomplete. A failed/interrupted job is not resumed
  automatically; zero completed_at means an incomplete snapshot, not success.
- Store normalized session metadata and one JSONB device per IP; avoid rewriting
  the entire growing snapshot. Read metadata and devices in a single SQL statement.
- Subscribe to live SSE before reading replay, so persisted events between replay
  and subscription cannot disappear. Consumers already deduplicate device IPs.
- Connect the existing streaming parser to runner stdout through a pipe. A host
  callback fires for each complete XML host before process exit; callback errors
  cancel the process and retain committed partial data. Final raw XML is preserved.
  Discovery first persists observed identity, then upserts enriched identity for
  the same IP. UI accepts updates as well as first observations. Scan findings
  use the same callback and final parsing skips already processed host IPs.
- Down 000011 removes only these new tables and loses their history. Back up both
  tables before any application rollback; test down/up only on disposable fixtures.
  No application database rollback is authorized by this plan.

## Verification

Risk: tenant isolation, duplicate session writers, write-before-publish ordering,
partial failures/cancellation, recovery through a new service/repository instance,
SSE replay. Targeted service/handler tests, dedicated PostgreSQL integration tests
and migration up/down/up; final gate All with Integration. No new dependency.
Browser refresh/reconnection remains a real visual gate when a browser is available.

## Evidence and limits

Targeted PostgreSQL tests PASS exit 0 (3,213s wall, package 1,751s): partial and
completed recovery through a fresh service/repository; tenant read/write/FK
isolation, duplicate IDs, upsert, append after completion, newest partial session,
up/down/up migration and previous host-pin schema preservation.
Container `nextgen-discovery-test-20261018`, PostgreSQL 16 Alpine, port 55432
bound only to 127.0.0.1, database `nextgen_discovery_test`. Explicit TEST_DB_*
configuration; no application DB/credentials used. Installation/preparation separate.

Live scanner tests use a child process blocked on an acknowledgement from the
host callback: process exit cannot precede delivery. Raw XML remains parseable;
callback error stops the child. Service tests cover write-before-publish ordering,
failed/cancelled writes, no cache fallback, scanner not launched on reused IDs,
live host save while runner remains active, finding persistence before runner
return and no duplicate findings after final parse. No red baseline was captured
for this new capability; existing discovery regression remained green.

Frontend targeted recovery/API tests: 10 PASS exit 0 (1,956s wall), partial snapshot,
completion after reload, retry, HTTP/network failure semantics. Later regression
with login/context: 14 PASS exit 0 (1,959s wall). Browser tests on port 3080 verify
the existing report/pinning build, not the newly modified discovery source.
The PO subsequently authorized updating the local application containers; see below.
No automatic job resumption or cross-instance live pub/sub is claimed.

## Integrated gate and real browser recovery

Root: `scripts/verify.ps1 -Profile Final -Scope All -Integration`, explicit TEST_DB_*
on the dedicated local database. PASS exit 0: backend suite 22.846s, vet 1.112s,
build 1.255s; frontend 184 tests 5.578s, typecheck/build 11.427s; PostgreSQL
integration 5.609s; docs 0.598s and diff 0.059s. Report
`.agent/artifacts/testing/20261018-124105-426c02b0.json`.

Both Docker images `m11-persist-20261018` built PASS, separate from the gate.
The PO explicitly approved updating only backend/frontend on localhost:3080,
including migration 000011. Compose recreation PASS exit 0 (6.665s), backend
healthy, frontend and `/healthz` HTTP 200. PostgreSQL container/volume kept running.
No remote publication, commit or push. This approval does not authorize future deploys.

Real Nmap/UI test: `nextgen-firewall-lab-20261018`, sole target 172.30.77.3 on
internal Docker network `nextgen-m11-lab` (172.30.77.0/29), no exposed lab ports.
The backend was temporarily connected at .2. Discovery without port enrichment
found one host; browser refresh and backend restart retained it. For the partial
case, TCP was temporarily dropped only in the fixture container. Reload immediately
after first live host (during enrichment) retained one device with the incomplete
snapshot banner, also after backend restart. The temporary rule was removed.
Screenshots observed in browser tool. No company target imported or real LAN/SSH
host contacted. No inference that an interrupted job resumed automatically.
