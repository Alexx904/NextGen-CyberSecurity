# E28-S1-T2 · Chokepoint mitigation UI

2026-10-18 · working tree. Reuse verified E28-S1-T1 and firewall safety UI.

## Contract before implementation

- SecOps/TenantAdmin/SuperAdmin can open mitigation for current SERVICE chokepoints
  with literal IPv4 TCP/UDP endpoints. Auditor remains read-only. Unsupported node
  types explain the limit; no arbitrary commands or inferred server selection.
- Load tenant SSH inventory, offer only exact matching host IPs excluding the TCP
  SSH management port. User explicitly chooses server and requests backend preview.
  Changing server invalidates preview/acknowledgement. Missing pin blocks apply.
- Render returned destination/protocol/port, command and timeout, with operational
  impact and explicit acknowledgement. Backend revalidates the graph at apply.
  Submit only node ID and server_id; no editable rule payload.
- Successful mutation opens existing FirewallManager for remote countdown,
  confirmation/rollback. Failed/uncertain apply exposes error and a manual status
  check, never retries or confirms automatically. Scope/graph changes dismiss
  an unapplied modal; applied server safety view remains available.
- Loading/error/empty, keyboard focus/escape and responsive scroll are required.
  No new dependency/schema/backend API. Frontend gate plus existing backend proofs.

## QA environment boundary

Current QA company's real graph has zero chokepoints, and partial tenant coverage
is explicitly reported (inventory/path cap). Do not claim complete UI mutation from that empty screen or
inject mocked browser responses. Prepare a separate local QA app/database with
clearly synthetic graph fixtures for end-to-end UI, then request explicit deploy/
fixture authorization. Main app/customer data and real SSH hosts remain untouched.
Component tests cover preview, stale response, acknowledgement, roles, errors and
safety handoff; actual remote effect already proved by T1's isolated integration.

Prepared (not started) bundle under `.agent/artifacts/E28-S1-T2/`: Compose with
frontend at `http://127.0.0.1:3081`, separate backend and tmpfs PostgreSQL database
`nextgen_e28_ui_qa`. Backend/DB are not published and use internal networks;
backend reaches only the internal fixture network at reserved .6. Main app .2,
firewall .3, SSH fixture .4 and unused down-test address .5 remain unchanged.
127.0.0.1 origin separates cookies/localStorage from the main localhost session.
No main environment, secrets or customer data are copied. Dedicated synthetic
operator, company, completed scan and two clearly marked future CVE fixtures seed
the graph; this is test data, not real vulnerability evidence. SQL checks database
name before INSERT-only seed. Benign standard-library HTTP service on fixture port
45696 compiled PASS exit 0 (0.607s); not yet started. Compose config PASS exit 0.
New QA credentials are test-only; real SSH/customer credentials are not used.

## Development verification

Frontend targeted command from frontend:
`npm.cmd test -- --run src/components/ChokepointRemediationModal.test.tsx src/pages/AttackPath.test.tsx src/api/attackPath.test.ts`
PASS exit 0, 22 tests/three files, 2.523s wall (runner 1.81s). Includes Auditor,
explicit server, missing pin, matching/management port, consent, stale preview,
uncertain apply and manual safety handoff. Initial fixture used an unsupported
matcher; replaced with separate exact call/count assertions without weakening.
Final integrated gate and complete isolated-browser workflow still pending.

## Closure verification

PO explicitly authorized the separate QA deployment and fixture firewall tests in
this session on 2026-10-18. Compose up --wait PASS exit 0, 6.614s. Seed inserted
only into the named disposable database (name guard PASS). Main app unchanged.

Final Frontend from repository root:
`powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1 -Profile Final -Scope Frontend`
PASS, report `20261018-183418-df186fee`: frontend npm test 215 tests, 6.827s;
tsc/Vite build 12.090s; docs and diff checks exit 0. First sandbox attempt failed
before test startup on esbuild directory access; rerun outside sandbox passed.
Backend/API/schema were not changed for T2; prior T1 proofs remain separate.

Browser against actual QA backend/PostgreSQL and pinned OpenSSH fixture:

- Explicit server required, preview displays INPUT/DROP and endpoint 45696/tcp;
  consent required. Changing selection clears preview and consent.
- Benign HTTP baseline PASS (`docker exec nextgen-e28-ui-qa-backend wget -T 3 -qO- http://172.30.77.4:45696`, exit 0, 0.106s).
- Apply opens FirewallManager with remote countdown (56s observed); HTTP timeout
  exit 1, 3.089s is the expected blocked result. Confirm removes timer and retains
  DROP; manual rollback restores HTTP exit 0, 0.091s.
- Second apply left unconfirmed: expiry is reported without falsely asserting
  rollback. Explicit reload verifies timer absent and original two fixture rules;
  HTTP restored exit 0, 0.084s. Automatic UI refresh is not claimed.
- Preview visually checked at 1440x1000 and 390x844. Tab/Shift+Tab wrap inside
  modal; Escape closes and restores opener focus. Viewport override reset.
- Artifacts: `.agent/artifacts/E28-S1-T2/preview-desktop.png`,
  `preview-mobile.png`, `timer-mobile.png`, `rollback-mobile.png`.

Self-review covered tenant inventory association, role guard, stale responses,
consent invalidation and uncertain mutation handoff. No new finding detected in
the inspected T2 scope. QA graph/CVEs are synthetic; observed firewall/SSH/HTTP
effects are real within the authorized lab. No customer hosts or external sends.
Technical completion `[~]`; PO acceptance remains separate.
