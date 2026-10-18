# M11-FIREWALL-01 · Remote firewall transaction

2026-10-18 · working tree. Scope: TECH-DEBT-07 and FR-11 mutation lifecycle,
then policy/rule replacement/reordering and terminal docking. Existing iptables
and iptables-nft model remains; no native nftables API.

## Contract before implementation

- Snapshot, detached watchdog and mutation form one remote operation under a
  `flock` lock. All application instances using the managed SSH account share
  that lock, irrespective of process memory or database server ID. Other host
  administrators/accounts remain outside this coordination boundary.
- Root runs iptables directly; other accounts use `sudo -n` for the individual
  iptables commands. No privileged shell or automatic sudoers/remote installation.
  Linux prerequisites: sh, flock, stat, nohup, date, sha256sum, awk, core file/sleep
  utilities and iptables tools. Snapshot input uses SSH stdin without base64.
- State lives in a verified private directory owned by the SSH UID, mode 0700,
  outside predictable world-writable snapshot filenames. Transaction IDs are
  random UUIDs; state includes tenant/server ownership and expiry.
- A watchdog is ready before any mutation. Its expiry restores only its own
  pending transaction under the same lock. Confirmation updates/removes the
  pending marker under lock; it does not kill a potentially reused PID.
- Backend restart loses no ownership/timer state. A later instance reads remote
  state and verifies the corresponding tenant-owned DB backup before confirming.
  DB failure after remote application leaves the watchdog armed and returns error.
- Mutation and rollback cannot cross a pending transaction owned by another
  tenant/server. Fixed operation arguments are validated before SSH. Rollback
  snapshot travels on SSH stdin, outside shell command arguments/heredoc data.
- Failed mutation triggers immediate restore; if restore fails, preserve the
  pending marker and detached retry. Never report successful cancellation/revert
  from a local deadline alone.

## QA and operational boundary

Use only `nextgen-firewall-lab-20261018` on internal Docker network
`nextgen-m11-lab`, 172.30.77.3, NET_ADMIN limited to its own network namespace.
Confirmed iptables 1.8.9 (nf_tables). No host networking or customer credentials.
Docker exec tests prove Linux command/state behavior. A second disposable fixture,
`nextgen-ssh-lab-20261018` (hostname `nextgen-ssh-fixture`, 172.30.77.4), runs
OpenSSH with fixture-only synthetic authentication and a pinned public host key.
Its loopback publication 55222 was unreachable from Windows (FAIL); tests instead
run as a cross-compiled Linux binary inside this named container against port 22.
`TEST_FIREWALL_SSH_IN_CONTAINER=1` also verifies the hostname before any local
fixture command. No remote customer action authorized.

Required checks: atomic overlap rejection, second service instance confirmation,
expiry/disconnection, stale confirmation after expiry, rollback error retaining
protection, hostile inputs, both iptables backends when available. Backend gate
for lifecycle; All for new routes/UI, plus browser on a version explicitly deployed.
Existing DB backup schema suffices; no migration planned for this increment.

## Additive API and UI contract

- GET firewall adds `revision`, SHA256 of canonical `iptables -S` output (no
  volatile counters), computed while holding the remote lock.
- PUT `/{id}/firewall/policy`: chain INPUT/OUTPUT/FORWARD, policy ACCEPT/DROP,
  `expected_revision` and `safety_timeout_seconds`.
- PUT `/{id}/firewall/rules/{line}`: full existing rule-fields DTO plus required
  `expected_revision`; replaces the selected line with the form fields.
- POST `/{id}/firewall/rules/move`: chain, `from`, `to`, required revision/timeout.
  Reorders the complete saved filter-chain rule without reconstructing its flags;
  all other chains/tables remain in the candidate snapshot.
- Existing add/delete accept optional revision (legacy compatibility); the new
  UI supplies it. Check expected revision under the same lock before snapshot or
  mutation. Stale state → 409 `FIREWALL_STATE_CHANGED`, reload required.
- All mutation routes remain SecOps/TenantAdmin/SuperAdmin, tenant/server checked through SSH
  inventory. No mutation or timer starts for invalid fields/revision.
- UI disables mutations during pending transactions; confirmation unavailable
  after expiry. Expired-but-pending state remains visible and is polled, without
  inventing a new countdown or assuming successful rollback.
- Replacing rules uses an explicit full-replacement form; advanced original
  options cannot be silently claimed as preserved. Moving preserves raw rule data.
- Terminal opens only on explicit user action and can be docked in firewall view;
  existing standalone SSH terminal remains supported.

## Verification evidence · 2026-10-18

- Before service replacement, actual Linux multi-instance regression was RED:
  the second instance accepted an overlapping operation (exit 1, 3.306s wall).
- Backend directory, `TEST_FIREWALL_CONTAINER=nextgen-firewall-lab-20261018`:
  `go test -tags=integration ./internal/service -run 'TestRemoteFirewallLinux|TestFirewallServiceLinux' -count=1 -v`
  PASS exit 0, package 16.648s. Four tests cover overlap/foreign owner, fresh
  instance confirmation, revision/editor operations, failed rollback retaining
  protection and expiry. Fixture rules are restored by cleanup.
- Compile preparation from backend: `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c -tags=integration -o ../.agent/artifacts/M11-FIREWALL-01/service-linux.test ./internal/service`,
  PASS exit 0, 1.355s on the final test source. Docker copy PASS.
- Execute inside the dedicated SSH fixture with explicit TEST_FIREWALL_SSH_CONTAINER,
  TEST_FIREWALL_SSH_PORT=22 and TEST_FIREWALL_SSH_IN_CONTAINER=1:
  `/tmp/service-linux.test -test.run=^TestFirewallSSH -test.v -test.count=1`.
  PASS exit 0: disconnect/expiry 10.90s; timeout with observed delayed remote
  command 10.72s. Production SSHService/pin/command transport used. First disconnect
  attempt failed on a residual synthetic baseline rule; removed that exact fixture
  rule and added a precondition rejecting contaminated baseline before rerun.
- Backend targeted handler/editor tests PASS exit 0, 2.281s wall; transport stdin
  and firewall units PASS exit 0, 2.441s wall. HTTP tests include unauthenticated,
  Auditor, malformed/trailing payloads and sanitized stale-state conflict.
- Frontend directory: `npm.cmd test -- --run src/components/FirewallManager.test.tsx`
  PASS exit 0, eight tests, 2.051s wall; pending expiry, revision ordering,
  replacement/interface/ICMP-port, policy and lazy docked terminal.
- `npm.cmd test -- --run src/components/WebTerminalModal.test.tsx`: PASS exit 0,
  one test, 2.266s wall; close before ticket resolves creates no WebSocket.
  Initial fixture lacked xterm loadAddon and failed; completed the fixture.
- Self-review found terminal async-unmount race and Windows shell line-ending
  risk; fixed session generation guards/observer teardown and `*.sh eol=lf`.
  Review is by the implementing agent, not independent.

Root `scripts/verify.ps1 -Profile Final -Scope All`: PASS exit 0; backend suite
26.858s/vet 0.791s/build 1.282s; frontend 191 tests 6.012s/build 11.440s;
docs 0.652s/diff 0.072s. Report `20261018-133322-543b162f.json` under ignored
testing artifacts. Prior PostgreSQL PASS reused: schema/repository/tests,
dependencies, configuration and DB environment unchanged in this increment.
Both Docker images `m11-firewall-20261018` built successfully, exit 0.

PO paused and then resumed work. Local deploy `m11-firewall-20261018` explicitly
authorized and executed: Compose exit 0, 7.044s; backend healthy, HTTP frontend/
healthz 200; PostgreSQL not recreated. Separate PO authorization permits SSH
inventory fixture `QA M11 SSH Lab`, synthetic password and Docker-verified pin.
Connection PASS 113ms; no customer server used.

Browser: 10-second pending mutation disabled editing; expiry removed the rule.
Explicit confirmation accepted two synthetic DROP rules; move reversed their
order retaining ports/comments. Docked SSH terminal echoed `QA_DOCKED_OK`.
FORWARD policy changed ACCEPT to DROP with pending state and explicit confirmation.
The policy click returned a browser timeout, but fresh UI showed the mutation;
no automatic retry was issued.

Replacement found numeric protocol `6` from iptables-nft disabling the editor.
Normalize 0/1/6/17 to all/icmp/tcp/udp; unknown numbers preserved. Parser test
RED exit 1 (4.107s wall), GREEN exit 0 (4.352s wall). Fresh gate/image/browser needed.

FR-09 browser found stale department across companies: Casa/Reparto1 to Casa 2
retained seven old targets. `Targets.test.tsx` RED exit 1 (1.817s wall), GREEN
exit 0 (1.651s wall). Reset department, bulk selection and mutation dialogs on
company change; always intersect department with company. Not yet deployed.
Fresh integrated gate `Final -Scope All` PASS exit 0, report
`20261018-172209-43bb2250.json`: backend suite 22.618s/vet 0.631s/build 1.315s;
frontend 192 tests 6.025s/build 11.165s; docs 0.699s/diff 0.076s.
Both corrected images rebuilt PASS, authorized local deployment completed
exit 0 (6.766s). Browser: TCP recognized, replacement 45692→45693 preserved
comment and confirmed; company switch now removes stale department/targets.
Screenshots `firewall-docked.png` and `company-switch.png` in ignored task artifacts.

Legacy backend available: selected iptables-legacy only inside .3 fixture,
reran the same four Linux tests PASS exit 0, package 15.882s, then restored
automatic nft alternative in finally. OpenSSH .4 fixture remained nft.
Real diagnostic JSON export includes two redaction markers, no synthetic password
and no JWT-shaped value; valid JSON saved as `debug-ssh-redacted.json`.
Self-review and Doc Sync complete; task `[~]`, acceptance PO separate.
No customer machines used; fixture inventory remains labelled QA.
