# E28-S1-T3 · Virtual patch generation

2026-10-18 · working tree. Implementation in progress; no deployment or remote
installation is included. Existing unrelated working tree changes are preserved.

## Contract before code

- Authenticated tenant read endpoint `GET /api/v1/findings/{id}/virtual-patches`:
  load the finding by tenant/UUID; never accept a CVE, regex, path or command from
  the client. Invalid UUID 400, missing/foreign finding 404, unexpected error 500.
- Return finding ID, CVE, catalog revision, support status/reason, limitations,
  authoritative source links and an array of named text files. Unsupported CVEs
  return an empty file array with an explicit reason, never generic protection.
- Curated coverage expanded after PO rejected the two-CVE initial scope:
  eight CVEs across Apache HTTP Server, Citrix ADC/Gateway, Grafana, VMware vCenter
  and Atlassian Confluence. Four export targets: Fail2ban filter/jail, CrowdSec scenario
  using the existing nginx parser, ModSecurity rule, Nginx raw-request guard.
  This covers known path indicators and requires operator review/testing. It
  does not assert exploitability or complete bypass resistance. Upgrade remains
  the remediation. Other network/SSH CVEs explicitly remain unsupported.
- Static catalog templates only; no untrusted banner interpolation. Logs must
  preserve the raw request and authentic client IP; trusted proxy setup and
  enforcement/bouncer are prerequisites. No automatic install/reload/SSH.
- Finding detail offers a separate accessible modal with loading/error/empty,
  engine selection, copy/download and source/limitation text. Scope changes or
  close discard pending responses; keyboard focus returns to the opener.
- No dependency or schema change. Gate Final All, meaningful service/handler/UI
  tests, browser QA on the authorized isolated app, self security review and Doc
  Sync. Native engine validation is separate from Go/JS template tests and must
  be reported explicitly if unavailable; no claim that these prove prevention.

## Sources checked on 2026-10-18

- [Apache security advisory](https://httpd.apache.org/security/vulnerabilities_24.html)
- [Fail2ban filter development](https://fail2ban.readthedocs.io/en/latest/filters.html)
- [CrowdSec scenario format](https://docs.crowdsec.net/docs/log_processor/scenarios/format/)
- [ModSecurity reference](https://github.com/owasp-modsecurity/ModSecurity/wiki/Reference-Manual-(v2.x))
- [Nginx core module](https://nginx.org/en/docs/http/ngx_http_core_module.html)

Catalog scope is an implementation choice to provide reviewed templates without
inventing CVE signatures; task remains incomplete until coverage and required
verification satisfy the DoD.

## Expanded catalog 2026-10-18.2

| CVE | Product | Strategy | Operational effect |
|---|---|---|---|
| CVE-2021-41773 / CVE-2021-42013 | Apache HTTP Server | Known raw URI traversal indicators | Encoded benign paths may match; bypass variants remain possible |
| CVE-2019-19781 | Citrix ADC/Gateway | Conservative /vpns/ restriction | May interrupt legitimate features; does not reproduce the entire vendor responder policy |
| CVE-2021-43798 | Grafana | Traversal under /public/plugins/ | Normal plugin assets excluded by test; normalization/bypass limits remain |
| CVE-2021-21972 / CVE-2021-21973 | VMware vCenter | Restrict /ui/vropspluginui | Disables legitimate plugin traffic through this control; origin workaround remains separate |
| CVE-2023-22515 | Confluence | Restrict /setup/ | Blocks setup, including legitimate setup |
| CVE-2023-22518 | Confluence | Restrict three setup-restore endpoints | Blocks legitimate restore too |

Additional primary sources:

- [Citrix mitigation](https://support.citrix.com/external/article/CTX267679/mitigation-steps-for-cve201919781.html)
- [Grafana advisory](https://grafana.com/blog/an-update-on-0day-cve-2021-43798-grafana-directory-traversal/)
- [VMware workaround](https://knowledge.broadcom.com/external/article/317865/workaround-instructions-for-cve202121972.html)
- [Confluence CVE-2023-22515](https://confluence.atlassian.com/security/cve-2023-22515-broken-access-control-vulnerability-in-confluence-data-center-and-server-1295682276.html)
- [Confluence CVE-2023-22518](https://confluence.atlassian.com/security/cve-2023-22518-improper-authorization-vulnerability-in-confluence-data-center-and-confluence-server-1311473907.html)

These original templates are derived advisory restrictions, not certified vendor
rules. Body/header attacks, alternate encodings, direct-origin access and prior
compromise are not covered. SSH CVEs without a meaningful HTTP/log signature
remain unsupported; no invented WAF fix for an SSH flaw. Response additionally
exposes product, strategy, operational impact and all catalog CVE IDs. Filenames,
Fail2ban references, CrowdSec scenario names and local ModSecurity IDs are unique
per CVE; review local ID collisions before installation.

## Verification and review

All commands below exit 0 unless explicitly described as a failing reproduction.
From backend, targeted `go test ./internal/service ./internal/delivery/http/handler
-run 'VirtualPatch|TraversalIndicator' -count=1`: 3.437s wall. Covers tenant/ID
guards, error redaction, unsupported services/CVEs, banner non-interpolation,
eight catalog positive/benign paths, all three restore endpoints, unique local
ModSecurity IDs and exported Fail2ban matching after timestamp removal. Initial
timestamp-removal test failed (2.740s); corrected filter passed (3.077s).

Final integrated gate from root:
`powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1 -Profile Final -Scope All`
PASS report `20261018-185831-e92556f2`: Go tests 23.472s, vet 0.783s, build
1.410s; frontend 220 tests 6.541s, tsc/Vite 12.729s; Docs 0.657s, diff 0.082s.
No repository/schema/dependency changed by T3; no new PostgreSQL suite attributed.
Earlier gate failed only on a new trailing blank line in API_SPEC, corrected;
the final report above is on the expanded integrated source.

Additional native check, from backend:
`$env:TEST_VIRTUALPATCH_NGINX_CONTAINER='nextgen-e28-ui-qa-frontend'; go test -tags=integration ./internal/service -run '^TestVirtualPatchNginxSyntax$' -count=1 -v`
PASS 2.778s wall, native Nginx syntax for eight generated snippets, no reload or
new listener. Initial stdin-only attempt failed because Nginx rereads config;
replaced input with a fixed temporary file in the explicitly named QA container.
Fail2ban, CrowdSec and ModSecurity native checks NOT RUN: no engines installed
or dependencies added. This is not a prevention/bypass-resistance claim.

PO explicitly authorized deploying T3 QA images and synthetic fixtures on 3081.
Same T3 lab updated to expanded catalog after the PO requested broader coverage;
no main app deployment, real CVE test target or external send. Browser loaded
actual tenant findings from PostgreSQL: supported Apache and Confluence, all
four engine selectors, exact clipboard content, real Nginx download and source
links; unsupported synthetic CVE has no export controls. Downloaded file passed
native `nginx -t` (0.084s) before catalog expansion. Expanded Confluence panel
shows eight-CVE catalog and restore impact. Escape restores finding opener;
desktop 1440x1000 and mobile 390x844 inspected; overrides reset. Screenshots and
download live under `.agent/artifacts/E28-S1-T3/` (including
`catalog-expanded-desktop.png` and `catalog-expanded-mobile.png`).

Self security review (not independent) inspected all new files plus the two-line
server route/wiring change and finding-detail integration, separating existing
M1.1 edits. No new command execution, arbitrary file path, tenant bypass or
credential use detected. Static source URLs and content only; no CVE payload or
finding text evaluated. Known limits remain visible in the UI and exports.
Technical completion `[~]`; acceptance and any additional required catalog scope
remain PO decisions.
