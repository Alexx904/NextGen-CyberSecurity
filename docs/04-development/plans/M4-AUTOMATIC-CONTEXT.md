# Automatic regulatory context from scan findings

Updated: 2026-10-19 - working tree. PO selected fully automatic technical context
instead of manual compliance assessments. This supersedes the manual M4 user
experience, guide and assisted workflow; it does not certify regulatory compliance.

## Contract and acceptance criteria

Use the existing authorized scan, Findings detail and Executive/Technical report
flows. No registration of assessment assets, association of observations, manual
control status, questionnaire, compliance score or SLA policy is required. The
former `/compliance` page redirects to `/findings`; its navigation item and manual
Audit export are removed. Manual `/api/v1/compliance` and `POST /reports/audit`
routes are no longer registered (404). Existing migration 000014 and persisted
history are retained; no rollback/drop or destructive cleanup is performed.
Legacy implementation modules remain dormant for history/compatibility review,
not mounted in the application or included in the frontend entry bundle.

For each REST and live/replayed SSE finding, derive additive `regulatory_context`
from the current installed catalogue. Preserve finding/scan IDs, source confidence,
detection method and port state. The object contains rule_version, confidence,
detection_method, implications[] and a limitation. Each implication has kind,
observation, meaning, action and references[] (identifier, short paraphrase, source).
No provider, AI, external request or new dependency is used at runtime. Existing
stored findings receive context on read without data migration. PDF generation
uses the same pure domain function and includes finding/scan provenance.

Rules `2026-10-19.2`:

- CVE correlation -> vulnerability management: NIS2 21(2)(e), DORA 9 and ISO A.8.8.
  Preserve correlation confidence; never imply exploitability or legal violation.
- Service identified as telnet/ftp/http -> review transport protection: NIS2
  21(2)(h), DORA 9(2), ISO A.8.24. The service name does not prove unencrypted
  transmission; TLS/tunnels/configuration remain uncertain. Do not infer the
  service from port number alone or label HTTPS as cleartext.
- Negative/non-open port observations, host discovery and invalid host placeholders
  produce no specific implication. Unknown services without CVEs also produce none;
  explicitly say that no matching rule does not establish security or compliance.

Nmap cannot establish organizational procedures, training, actual backup restoration,
legal applicability or certification. No automatic compliance percentage is shown.
Catalogue maintenance remains versioned software releases; automatic interpretation
of scan results is distinct from monitoring new legislation. The earlier regulatory
monitoring design is deferred, not activated by this request.

## Sources and interpretation

Official sources checked 2026-10-19:
[NIS2 text](https://eur-lex.europa.eu/legal-content/EN-IT/TXT/?uri=CELEX%3A32022L2555),
[DORA text](https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX%3A32022R2554),
[ISO catalogue](https://www.iso.org/standard/27001).
EUR-Lex ELI pages returned a browser challenge; indexed official legal-content
versions supported the cited vulnerability/cryptography/protection themes. ISO
catalogue confirms the standard/edition and organizational scope; control identifiers
reuse the curated M4 references. These are engineering relevance mappings and
paraphrases, not exhaustive legal interpretation or reproduction of standard text.

## Verification and review

Risk: false compliance, confidence loss, REST/SSE divergence, hidden manual flows,
tenant leakage and PDF layout regressions. Target domain boundaries, DTO provenance,
SSE JSON compatibility, frontend missing/empty/context states and no input forms.
Verify Final All (tests/vet/build and frontend tests/typecheck/build); no DB/schema
or repository change in this follow-up. Retained DB tests do not constitute a new
integration PASS. Render both PDF templates and inspect appended context pages.
Real browser verification depends on a connected surface or PO check; do not reuse
prior screenshots as proof of this replacement. Security self-review and Doc Sync.

## Execution evidence and remaining gate

Integrated command from repository root:
`powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1 -Profile Final -Scope All`.
Workspace Go cache; Node/esbuild required approved elevation. Report
`20261019-192754-e36fd35c` records application PASS: backend `go test ./...`
31.223s, vet 0.934s, build 1.478s; frontend 250 tests/44 files 6.463s and
TypeScript/Vite build 11.017s, each exit 0. The runner's overall exit was 1 solely
because Docs found a renamed historical API anchor. Correcting that Markdown link
then running `-Profile Docs` passed (`20261019-192914-71886374`, verifier 0.838s,
diff 0.066s, exit 0). No app/test/dependency/config input changed after the integrated
application checks, so those results are retained without repeating suites.

Targeted domain/handler/report tests PASS (`20261019-192508-001a90f0`, 3.579s).
Frontend targeted PASS (`20261019-192721-7ffd9568`, 1.403s) includes actual legacy
route redirection and automatic/no-form/missing-context states. The redirect test
initially used an incorrect stored company key and failed; correcting the fixture
to the application's real key passed. No production guard was weakened.

PDF test generates real Executive and Technical files with both mapping types;
it decompresses PDF streams and checks version, confidence, CVE and official source
URLs. Both three-page fixtures rendered with Windows PDF WinRT; all six pages were
visually inspected, with readable context, complete references, footer separation
and no clipping. Artifacts: ignored `.agent/artifacts/M4-AUTO/`, including PDFs and
page PNGs. These are synthetic fixtures, not reports from a real target scan.

Security self-review, not independent review: source names rendered as escaped text,
official reference URLs fixed in domain rules, tenant authorization unchanged,
REST/SSE confidence preserved, no remote calls/credentials/provider or automatic
issue mutation. Negative port states produce empty context, and transport names
do not establish actual encryption state. Manual server routes are unregistered;
the frontend entry bundle no longer contains Compliance/Audit chunks. Legacy
source remains dormant and historical DB records are not deleted.

Doc Sync covers product requirements/use cases, current UI, ADR-30/HLD/API/schema,
functional domains, QA/task/state and supersession notices in old plans. Existing
roadmap/SSH/other work is preserved. No branch, commit, push, deployment or live
scan. Repository/DB integration was not rerun: this follow-up changes no repository,
migration or database configuration, and no new DB PASS is claimed.

Browser inventory returned no connected surfaces. The PO is deploying and has
previously chosen to perform visual checks; this replacement still needs its own
confirmation. M4-AUTO-01 remains `[ ]` until that gate passes. To check: open an
existing CVE or HTTP/FTP/Telnet finding, verify the automatic context and no forms,
download an Executive/Technical PDF, and confirm the old Compliance URL redirects.
