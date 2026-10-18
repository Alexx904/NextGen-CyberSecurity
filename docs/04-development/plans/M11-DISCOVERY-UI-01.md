# M11-DISCOVERY-UI-01 · Discovery actions and scan handoff

2026-10-18 · working tree. FR-10.4/10.5 completion inside M11-CLOSE-01.
Browser verified copy/import/recovery on isolated Docker hosts; final inventory
found missing CSV/JSON export and selected-device scan handoff. FR-10.5's former
LAN preset is superseded by accepted FEAT-WORKFLOW-CLEANUP-02/SCAN-UX-01; see
CHANGELOG and scanner domain. The eight accepted presets remain unchanged.

## Contract and risk before implementation

- Export the explicitly visible/selected device set, clearly labelled. Preserve
  snapshot metadata/partial completion in JSON. Quote CSV and neutralize spreadsheet
  formula prefixes in untrusted hostname/vendor strings. No external transmission.
- Selected-device scan opens the existing import/destination flow first, then
  NewScanModal restricted to those literal IP values in the chosen group. Fresh
  tenant inventory determines target IDs; blacklist remains enforced. Never submit
  a whole group as a substitute for selected devices.
- Existing group launches retain their contract. Restricted launches use existing
  POST /scans with target_id per eligible device, handling partial failures and
  retrying only definite rejected requests; network/5xx uncertainty blocks retry
  and requires history review. Show exact count and exclusions. Scheduling
  uses the same restricted eligible set. No API/schema/dependency change.
- Retain the eight accepted presets and current intrusive-consent policy.
  Import may open configuration; it never starts a scan automatically.
- Refresh CompanyContext after successful import to keep sidebar/count/scope fresh.

## Verification

Frontend scope: unit export hostile text, selected-target handoff with same-group
unselected/blacklisted devices, partial failure retry, restricted schedules. Existing
group launch tests must remain unchanged. Final Frontend gate, Doc Sync/self-review,
browser on authorized local build with only 172.30.77.3/.4 fixtures. No automatic
customer scans or recurring QA schedules. PostgreSQL unchanged; no DB rerun.

FR-07 already observed on QA range .3–.5: completed 2 up/3 total, 1 down excluded;
only .4:22 service in table and CSV, open-port toggle remains one row. PDF contains
summary and only observed service evidence, with range text retained as metadata.
Artifacts in `.agent/artifacts/M11-FIREWALL-01/alive-only.*`.

## Implementation and evidence

Restricted launch sends at most four requests concurrently, deduplicates IPs and
honors any blacklisted duplicate inside the selected subtree. Definite HTTP
rejections retain successful acknowledgements; network/5xx uncertainty disables
resubmission pending history review. Existing whole-group behavior is unchanged.
NetworkDiscovery component integration proves import selection reaches the scan
modal and only one selected target is submitted despite another same-group target.

Targeted frontend tests: 35 tests PASS exit 0, 2.306s wall (runner 1.66s).
Preset contract: eight accepted presets retained; an attempted LAN preset exposed
the existing no-host-only assertion. Checked accepted cleanup history, removed the
attempt and clarified historical FR-10.5; no assertion weakened.

Root `scripts/verify.ps1 -Profile Final -Scope Frontend` initially FAIL on policy
test acting before status load. Added enabled-state wait, preserving assertions.
Final PASS exit 0: 201 tests 5.634s, build/typecheck 11.126s, docs 0.620s, diff
0.072s; report `20261018-174149-be1394db.json`. Backend/DB sources and environment
unchanged since prior gate; no backend/database PASS inferred from frontend.
Self-review by implementing agent: CSV injection, selected-only requests,
blacklist/duplicate handling, bounded concurrency, uncertain-response retry and
import-without-launch checked. Corrected-build browser remains pending.

## FR-13 transport check discovered during QA

Actual terminal worked, but diagnostic export contained REST only: WebTerminalModal
never called the existing WS logger. Add connection states and frame direction/
length metadata, without terminal input/output payloads or ticket query strings.
Terminal content can contain unstructured secrets and must never enter the debug
buffer. Verify with synthetic sensitive input/output and real browser source filter.
Frontend scope only; include this correction in the pending frontend gate/image.

WS regression RED exit 1 (1.978s wall), GREEN exit 0 (1.966s wall), two tests.
Final Frontend after WS correction PASS exit 0: 202 tests 5.623s, build/typecheck
11.445s, docs 0.617s/diff 0.066s. Report `20261018-174546-eafd5b25.json`.
Backend/PostgreSQL unchanged. New local frontend image pending browser activation.

PO explicitly authorized frontend-only deployment `m11-discovery-ui-20261018`:
Compose PASS exit 0 (0.958s), frontend HTTP 200; backend/DB not recreated.
Browser exported two visible devices CSV and one selected device JSON, with the
company destination correctly preselected. Import succeeded, but company refresh
unmounted the route and lost the pending scan dialog. CompanyRoute regression
RED exit 1 (1.889s wall); preserve the mounted page while an already valid company
refreshes, keeping initial/unresolved selection gating. Targeted guard/handoff/
modal regression GREEN, 39 tests exit 0 (2.816s wall). Server auth unchanged.
Fresh frontend gate/image/browser required for this actual fix.

Final corrected source: `Final -Scope Frontend` PASS exit 0, report
`20261018-175253-fd10722f`: 204 tests 6.227s/build 12.841s/docs 0.669s/diff 0.073s.
Image rebuilt PASS; same authorized frontend-only QA deployment completed exit 0
(1.056s), HTTP 200. Browser handoff now survives refresh, reuses existing inventory,
shows one eligible .4 target and starts exactly one job. QA company retains three
targets; scan history grows from two to three jobs, only new .4 job at 17:54 local.
CSV two visible devices and JSON one selected .4 verified from downloaded files;
JSON preserves total_discovered=2 and completed snapshot metadata.
WS filter, Markdown copy and JSON export PASS: 102 total entries/64 WS entries
with connecting/open/frame/disconnect-requested; terminal sentinel, ticket query
and JWT-shaped values absent. Mobile override 390×844: document/client width both
375 (scrollbar deducted), toolbar fits; viewport reset afterwards. Screenshots and
downloaded evidence in `.agent/artifacts/M11-DISCOVERY-UI-01/`.

## Focused accumulated-risk audit

Five completed increments prompted self-audit of auth/scope, storage-before-SSE,
remote ownership/concurrency, async UI refresh and diagnostics. This audit found
and fixed stale department, numeric protocol, route-unmount and missing WS logging.
Selected-only requests, blacklist duplicates, bounded HTTP concurrency, uncertain
response blocking and no terminal payload in logs checked. No independent reviewer
or customer-host validation claimed. Host reboot/temporary-state loss, per-UID
firewall coordination, process-local live pub/sub and no automatic job resume remain
explicit limitations in owning plans/domains. M11-DISCOVERY-UI-01 `[~]`, awaits PO.
