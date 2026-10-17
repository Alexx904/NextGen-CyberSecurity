# SSH-TOOLS-01 · Terminal proxy and firewall diagnostics

Ultimo allineamento: 2026-10-17 · working tree.

PO accepts SSH-CONNECT-02: connection works. New failures: terminal displays
WebSocket gateway error; firewall displays Axios HTTP 502.

## Findings and scope

Nginx and Vite had no `/ws/` proxy. Read-only HTTP upgrade request against the
existing frontend at localhost:3080 returned 200 SPA rather than reaching the
authenticated gateway. Added proxy upgrade forwarding in both environments.

Firewall status ran `sudo iptables` without PTY or a noninteractive flag; API
hid the cause and UI showed Axios's generic HTTP status. Added root/PATH support,
`sudo -n`, sanitized prerequisite/privilege errors, and API error text in the
existing UI. No remote permission changes, sudo password forwarding, firewall
mutation changes, watchdog changes, dependency changes or database changes.
Actual Debian privilege/package state remains unknown until PO's next test.

## Evidence

| Directory | Command / scope | Result | Wall seconds |
|---|---|---|---:|
| backend | `go test ./internal/service -run '^TestFirewallStatusUnattendedAccess$' -count=1`, before fix | FAIL exit 1 | 2.526 |
| frontend | `npm.cmd test -- src/components/FirewallManager.test.tsx`, before fix | FAIL exit 1, API message absent | 2.797 |
| backend | `go test ./internal/service ./internal/delivery/http/handler -run 'Firewall' -count=1` | PASS exit 0 | 3.359 |
| frontend | `npm.cmd test -- src/components/FirewallManager.test.tsx` | PASS exit 0 | 1.626 |
| root | `node .agent/artifacts/ssh-tools-01/proxy-smoke.mjs` | PASS exit 0, Vite and Nginx | 4.447 |
| backend | `go test ./...`, final | PASS exit 0 | 20.011 |
| backend | `go vet ./...`, final | PASS exit 0 | 0.948 |
| backend | `go build ./...`, final | PASS exit 0 | 1.846 |
| frontend | `npm.cmd test`, final, 175 tests | PASS exit 0 | 5.655 |
| frontend | `npm.cmd run build`, final | PASS exit 0 | 11.944 |
| root | `scripts/verify-agent-docs.ps1`, final | PASS exit 0 | 0.545 |
| root | `git diff --check`, final | PASS exit 0 | 0.057 |

Final runner: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1 -Profile Final -Scope All`.
Report: `.agent/artifacts/testing/20261017-164340-c24f99fd.json`.
Go cache: ignored `.agent/artifacts/go-cache`.
Frontend red first could not start in sandbox; reproduced with escalation.
One Quick runner invocation rejected a comma-joined package argument; direct
Go invocation above corrected it. Neither setup failure is counted as a test.

Proxy smoke uses synthetic WebSocket echo, full terminal path/query with a fake
ticket, bidirectional bytes, actual Vite configuration and Nginx configuration
with only the upstream address replaced for the fixture. Nginx uses the existing
local frontend image in an isolated container with an ephemeral loopback port;
container stopped and automatically removed. First image reference was stale:
Docker returned image-not-found before starting; used the available local image
without pulling. Application containers and target were not modified.

Handler tests verify 502/code/sanitized message; service tests cover missing
iptables/sudo and password-required/denied cases. Frontend test verifies visible
API message instead of generic 502. These do not prove real sudoers policy.

## Review and acceptance

Accepted by PO on 2026-10-17: «Ok perfetto è funzionante», following the generic
502 diagnosis. SSH-TOOLS-01 is `[x]`; no package/sudoers change or exact cause is
inferred from this confirmation. Real outcome is PO-reported, not agent-observed.
The follow-up below is historical; no requested output remains pending.
SSH-HARDEN-01/TECH-DEBT-07 watchdog/rollback gates remain separate and open.

Follow-up PO: still sees generic 502. Read-only runtime check confirms served
SSHServers-BeM5cCm3.js matches the verified local build and the running backend
binary includes FIREWALL_PRIVILEGES_REQUIRED. Unauthenticated request through
localhost:3080 to `/api/v1/ssh/servers` returns expected 401 JSON, curl exit 0.
Different latest image tags alone do not prove old code is active. Browser cache,
remote package availability and privileges remain unverified. Requested forced
refresh and Debian package/noninteractive read-only sudo checks; no deployment
or source change. Task reopened, no real-target success claimed.

Self hostile review: WS upgrade headers, unchanged path/query/auth, exact host
pin preserved, no credentials in diagnostics, root bypass only after `id -u`,
no automatic grants, safe static shell command, no changes to writes/watchdog,
and preservation of all unrelated working-tree changes. No independent review.
No new layout; no browser visual or real Debian PTY/firewall claim. PostgreSQL
integration N/A because no repository/schema changes. Rebuild backend/frontend,
then terminal should connect; firewall should show rules or an actionable error.

Doc Sync: SSH domain (behavior), API_SPEC (errors), DOMAIN_MAP (paths),
DEPLOYMENT (activation/prerequisites), TASK_LIST/PROJECT_STATE (state), this QA
(evidence). SSH-HARDEN-01/TECH-DEBT-07 Linux/watchdog gate remains open.
