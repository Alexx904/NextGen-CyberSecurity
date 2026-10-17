# SSH-HARDEN-01 · Host identity and bounded commands

Ultimo allineamento: 2026-10-17 · working tree.

## Scope and contract

### Follow-up SSH-CONNECT-02 · 2026-10-17

PO reports a host-key mismatch from `/ssh-servers`, while Windows OpenSSH to
`alex@192.168.1.18` succeeds. This supersedes the previous TCP-timeout symptom;
it does not prove which key is stored or offered by the real target. No target
credentials, known_hosts files or application database were read.

Reproduced locally: Go's current host-key preference selects ECDSA on a server
offering ECDSA/RSA/Ed25519, rejecting valid pins for its other keys. Fix: remove
each rejected key family (including RSA signature aliases) and reconnect with
the remaining algorithms. Every callback checks the original exact pin before
authentication. A single shared 10s context bounds all attempts. Transport and
authentication failures are not retried. Defaults mirror the pinned x/crypto
v0.27.0 dependency; review this list when upgrading it. No dependency change,
API/schema/UI change, automatic enrollment, or pin rotation is introduced.

Evidence from root unless stated otherwise:

- Baseline `scripts/verify.ps1 -Profile Quick -Scope Backend -BackendPackages ./internal/service`: PASS exit 0, 16.489s.
- Red from backend: `go test ./internal/service -run '^TestSSHMultipleHostKeys$' -count=1`: FAIL exit 1, 2.464s wall; Ed25519/RSA valid pins rejected and diagnostic incomplete.
- Green from backend: `go test ./internal/service -run '^TestSSH(MultipleHostKeys|HostIdentityBeforeAuthentication|HandshakeCancellation|CommandCancellation)$' -count=1`: PASS exit 0, 3.887s wall, 2.124s package.
- Final `scripts/verify.ps1 -Profile Final -Scope Backend`: PASS exit 0; backend `go test ./...` 17.840s, `go vet ./...` 0.871s, `go build ./...` 1.470s; docs 0.671s, diff 0.055s.
- Report: `.agent/artifacts/testing/20261017-163723-b83c9a1f.json`; GOCACHE in ignored `.agent/artifacts/go-cache`.

Final suite includes multi-key success for each family, unknown pin rejection
with all observed fingerprints, no password sent to mismatched keys, and one
authentication attempt on a trusted key when login fails. Existing cancellation
tests pass. Self hostile review covered finite family removal, RSA aliases,
unchanged trust, shared deadlines, connection closure, tenant isolation and
preservation of unrelated changes. No independent review claimed.
Frontend/DB integration N/A: no frontend/repository/schema changes. No visual
claim, target connection or deployment. To activate in the local Docker app:
`docker compose up -d --build --no-deps backend`, then test the server card.
Real Debian outcome and correctness of its saved pin remain unverified.

First increment of TECH-DEBT-06/07; existing discovery changes are outside scope.
Add `host_key_fingerprint` (canonical OpenSSH SHA256, empty for legacy inventory)
to SSHServer, create/update DTOs, public responses and the registration form.
Update omission preserves the pin; explicit empty clears it. Changing host/port
without an explicit pin clears trust and resets connection status. Pin changes
reset status. No trust-on-first-use or discovery endpoint: the operator obtains
the server fingerprint over an independently trusted channel. A missing or
mismatched pin prevents authentication, commands and web terminal access.
The client public-key fingerprint retains its existing meaning.

Migration 000010 adds a separate non-null column defaulting to empty. Down drops
only host pins; inventory, client fingerprints and encrypted secrets survive.
Reapplying up requires re-enrollment: no migration infers server trust.

SSH dial/handshake is bounded at 10 seconds, respects earlier context deadlines
and cancellation. TestConnection and ExecuteCommand have an overall 30-second
budget; closing the dedicated connection interrupts blocked session operations.
Web terminals retain their existing lifetime after the bounded handshake.

## Implementation and verification

Domain/DTO/SQL contracts first; service callback and context lifecycle next;
repository, handler and form wiring; red/green loopback tests for unpinned and
wrong keys, explicit rotation, endpoint changes, handshake and command stalls.
UI tests cover pin submission and legacy inventory. Integration tests cover
create/list/update and tenant isolation on a dedicated PostgreSQL container.
Migration up/down/up is tested separately on that test database. Final gate:
`scripts/verify.ps1 -Profile Final -Scope All -Integration`, plus docs/diff.
Self hostile review; no application database or remote client host is modified.

## Remaining operational gate

Transport cancellation does not prove remote process termination. TECH-DEBT-07
and E28-S1-T1 stay open until a host-side snapshot/watchdog/apply transaction,
cross-instance coordination and Linux disconnect/timeout/rollback tests exist.
Do not claim that this increment closes firewall safety or authorizes operational
remediation. TECH-DEBT-05 and lower-priority debts follow separately.

## Evidence · 2026-10-17

Initial red: backend `go test ./internal/service -run '^TestSSHHostIdentityBeforeAuthentication$' -count=1`
reports FAIL (Go exit 1): unpinned host accepted. Green loopback proves missing/
mismatched pin rejection and trusted pin success. The stall fixture waits for
connection close; command start is signaled before asserting its deadline.

Root final command: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1 -Profile Final -Scope All -Integration`, PASS exit 0.
`GOCACHE` points inside ignored `.agent/artifacts/go-cache`. DB uses explicit
`TEST_DB_*`, loopback port 55439, `nextgen_ssh_test` on an isolated PostgreSQL 16
container, not the application DB. Docker/Vite required toolchain escalation;
no approval rejection. Initial DOM click did not trigger submit in the test
runtime; explicit submit with a separate input-validity assertion passes.

| Directory | Command | Seconds | Exit |
|---|---|---:|---:|
| backend | `go test ./...` | 21.260 | 0 |
| backend | `go vet ./...` | 0.796 | 0 |
| backend | `go build ./...` | 1.353 | 0 |
| frontend | `npm.cmd test` (27 files, 166 tests) | 10.865 | 0 |
| frontend | `npm.cmd run build` | 14.767 | 0 |
| backend | `go test -tags=integration -count=1 ./internal/repository/postgres` | 4.905 | 0 |
| root | `scripts/verify-agent-docs.ps1` | 0.470 | 0 |
| root | `git diff --check` | 0.052 | 0 |

DB integration includes pin create/get/list/rotation with tenant filtering and
real migration down/up: host column absent after down, encrypted credential,
client label and client fingerprint preserved; after up pin empty, never inferred.
The complete suite includes preexisting Discovery changes without attributing
them to this task. No application test rerun after Markdown-only updates.

Self hostile review covered missing/mismatched identity, tenant access, invalid
canonical base64, stale pins on endpoint change, connection lifetime/cancellation,
SQL alignment, rollback data loss and legacy behavior. No new dependencies.
Browser inventory empty; iab creation returns unavailable. No UI visual or Linux
firewall claim, no `[~]` while the UI gate remains open.

Updated documents and purpose: ssh_and_firewall (behavior/API), REQUIREMENTS
(host identity), UI_SPECIFICATIONS (enrollment/reset), DATABASE_SCHEMA (column),
ADR-20 (trust/lifecycle), DEPLOYMENT (legacy enrollment/down loss), TASK_LIST and
PROJECT_STATE (evidence/open gates), this plan (scope/QA). Discovery acceptance
and preexisting changes preserved.
