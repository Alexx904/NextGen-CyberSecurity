# TECH-DEBT-09 dependency advisory assessment

Alignment: 2026-10-19, working tree. Automatic and real-browser compatibility gates complete; accepted by PO on 2026-10-19. Original 2026-10-18 evidence and limits are preserved below.

## Scope and acceptance criteria

Update existing frontend dependencies, distinguish browser runtime from build/test
exposure, preserve application behavior, pass the frontend Final gate and document
remaining findings. Backend, database and deferred alerting implementation are
outside this change. Existing uncommitted sources are preserved; the frontend
suite includes the local alerting tests without accepting or publishing that work.

## Assessment and remediation

The npm registry audit initially reported 16 affected packages: 3 critical,
6 high and 7 moderate. These are package counts, not unique advisory counts.
Raw before/after reports are in ignored `.agent/artifacts/TECH-DEBT-09/`.

| Boundary | Findings | Action |
|---|---|---|
| Browser runtime | React Router open redirect and SSR hydration constructor injection; this project is a client SPA, so SSR exposure is conditional | Upgrade react-router-dom 6 to 7.18.4; retain React 18 |
| Development/build | Vite file access/Windows path and esbuild development-server issues | Upgrade Vite 5 to 7.3.7 and React plugin to 5.2.0 |
| Tests | Vitest UI/mocker file access and tinypool prototype-pollution gadgets | Upgrade Vitest and coverage provider together to 4.1.11, removing tinypool from the installed tree |
| CSS build | Tailwind 3 transitively uses vulnerable braces and postcss-selector-parser | Upgrade to Tailwind 4.3.3 and official @tailwindcss/postcss 4.3.3, removing the affected chain |

Runtime and full audits now both exit 0 with zero reported vulnerabilities.
The intermediate seven remaining build packages had these root causes:

- [GHSA-vfj7-8cjw-p6xm](https://github.com/advisories/GHSA-vfj7-8cjw-p6xm): braces stack exhaustion on deeply nested patterns; propagated through micromatch, fast-glob, chokidar and tailwindcss.
- [GHSA-rj75-hqrm-r3gf](https://github.com/advisories/GHSA-rj75-hqrm-r3gf): selector parser CPU exhaustion; propagated through postcss-nested and tailwindcss.

These packages process local source/configuration during CSS generation and file
watching. They are not the Node runtime of the deployed frontend: its Docker final
stage serves compiled static assets with Nginx. The affected dependency chains
have been removed. Audit coverage does not assess the Nginx base image or prove
absence of undisclosed vulnerabilities.

PO explicitly authorized Tailwind 4 plus the official MIT-licensed
`@tailwindcss/postcss` plugin from npm, without cost, on 2026-10-18.
Registry metadata confirms the plugin's version/license and official Tailwind
node/oxide dependencies. Autoprefixer is removed because Tailwind 4 handles
prefixing. No forced audit fix, advisory suppression or overrides were used.

## Compatibility and review

Node baseline is 22.12+ on the 22 line, or 24+; CI uses Node 22 and Docker already
uses Node 22. Local verification uses Node 24.15.0. The manifest declares the
requirement. Vite 7 avoids a further bundler migration to Vite 8.

Router 7 removes the v6 `future` test props; the corresponding flags and obsolete
warning suppression were removed. Vitest 4 rejects spying on happy-dom's missing
`window.confirm`; the firewall test now stubs that browser API explicitly and
restores globals after each test. Assertions and application logic are unchanged.
The initial full run demonstrated the failure, followed by eight passing firewall
tests. See [Vitest migration guidance](https://vitest.dev/guide/migration/).

Tailwind loads the existing theme explicitly with `@config`; source discovery is
limited to index.html and src through CSS `@source` declarations. Old content
configuration was removed to keep one source list. Across the existing UI,
`shadow-sm`, `backdrop-blur-sm`, `outline-none` and `flex-shrink-0` were migrated to
`shadow-xs`, `backdrop-blur-xs`, `outline-hidden` and `shrink-0` respectively.
Base CSS preserves v3 border/placeholder defaults and pointer cursors on enabled
buttons. Existing Settings alert-card import/mount and all other preexisting work
remain intact; only its five focus utility tokens belong to this migration.

Tailwind 4 requires Safari 16.4+, Chrome 111+ or Firefox 128+; older browsers are
outside this toolchain's support. Default palettes, gradient interpolation and
space/divide selectors have upstream changes, so compiled CSS checks cannot prove
visual parity. See the [official upgrade guide](https://tailwindcss.com/docs/upgrade-guide).
Compiled CSS inspection confirms custom background/border utilities, shadow,
blur, shrink, accessible focus outline and responsive grid output. An initial
probe for unused `text-severity-critical` was invalid; no source references it,
so omission by the compiler is expected. Evidence: `css-check.json` in artifacts.

Self-review covers manifest/lockfile consistency, peer dependencies, runtime/build
boundaries, unchanged assertions and preserved unrelated edits. This is not an
independent security review. No new service, credential, deployment or external
message is involved. Browser inventory returned no apps or browsers; real UI
verification is not executed. An explicit attempt to open the integrated browser
also returned `Browser is not available: iab`. AGENTS sections 5/6 therefore keep
the task `[ ]` until the real visual gate is available.

## Verification

Commands run from repository root, except the runner's npm steps in `frontend`.

| Command | Result |
|---|---|
| `npm.cmd audit --json --prefix frontend` | Initial exit 1, 16 affected packages; sandbox network attempt failed before the successful registry request |
| `npm.cmd ci --prefix frontend` | PASS, exit 0; 217 packages installed, 218 audited, 23 s; preparation, not a test |
| `npm.cmd audit --omit=dev --json --prefix frontend` | PASS, exit 0, no runtime findings, 0.906 s |
| `npm.cmd audit --json --prefix frontend` after Tailwind 4 | PASS, exit 0, no findings including dev/build, 0.998 s |
| `scripts/verify.ps1 -Profile Quick -Scope Frontend -FrontendFiles src/App.test.tsx` | PASS, 1 test, 1.363 s; first sandbox attempt failed on filesystem access |
| Quick targeted Reports/ScanDetail/debugConsoleE2E | PASS, 3 files / 10 tests, 1.989 s |
| Quick FirewallManager | PASS, 8 tests, 1.714 s; report `20261018-195201-d0740a5c` |

Final runner `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1
-Profile Final -Scope Frontend`, report `20261018-201404-9eb98c27`: PASS, exit 0.
40 files / 229 tests PASS (6.383 s), typecheck/build PASS (11.665 s), document
structure PASS (0.663 s), diff check PASS (0.090 s). Earlier pre-Tailwind runner
failures and their corrections are superseded by this complete integrated run.
Only documentation changes follow this gate, so no application suite is rerun.
The assessment adds one Markdown file; setup/core/UI/task/state documentation
records the toolchain contract, evidence and remaining real-browser gate.
No frontend lint command exists. Backend/PostgreSQL tests are not required for
this dependency-only frontend increment. CI and Docker execution are not claimed.

## Original browser gate (2026-10-18; completed below)

In a connected browser, run the updated frontend and inspect login, dashboard,
reports, Settings and a modal at desktop and narrow widths. Check custom theme,
gradients/spacing, keyboard focus, enabled/disabled buttons, and navigation after
login. No real scans, firewall changes or external alert sends are needed.
Until this is observed, automatic compatibility is PASS and visual compatibility
is NOT EXECUTED; the task must not be marked implemented-and-verified `[~]`.

## Browser completion - 2026-10-19

The original remaining gate is now completed. The PO reconnected the in-app browser
and supplied localhost:3080; that deployment was inspected read-only. Updated
working-tree sources were exercised separately through Vite 3000, real Go backend
8081 and disposable PostgreSQL nextgen_consent_test on 55432. Synthetic account,
company, target, historical completed scan and schedule fixtures were used; no real
Nmap operation, provider send or update of the 3080 deployment occurred.

Login, dashboard, reports, Settings and the scan modal were visually inspected at
desktop (1280px) and narrow (390px) widths. Theme, gradients, responsive grids,
spacing, modal scrolling, enabled/disabled controls and post-login navigation render
correctly in this connected Chromium browser. Keyboard Shift+Tab wraps to the last
enabled control, Tab wraps to Close, and Escape returns focus to New Scan; the
focused control has a visible browser outline. This is a compatibility check, not
pixel parity against an earlier captured baseline or cross-browser certification.
Screenshots are in ignored `.agent/artifacts/TECH-DEBT-05/` with task-specific names.

The integrated Final All/Integration gate reports 233 frontend tests plus successful
typecheck/build in `20261019-130817-5cbdbed6`; dependency/configuration inputs are
unchanged since the zero-vulnerability audits on 2026-10-18, so no new audit result
is claimed. The consent increment adds four frontend cases; it does not change
package versions. TECH-DEBT-09 is `[x]`, accepted by PO on 2026-10-19.
