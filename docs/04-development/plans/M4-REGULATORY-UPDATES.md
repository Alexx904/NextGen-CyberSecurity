# Regulatory catalogue updates - proposed design

Deferred design. [Automatic finding context](M4-AUTOMATIC-CONTEXT.md) interprets
scan results; it does not activate regulatory monitoring or manual compliance.

Updated: 2026-10-19 - working tree. Design only, requested by PO after M4.
No collector, scheduled task, database migration or automatic activation is implemented.

## Objective and boundaries

Detect relevant official publications, explain their impact and release reviewed,
versioned control mappings without silently changing previous assessments. The
system must distinguish publication monitoring from legal interpretation and
activation. It must never claim complete regulatory coverage or determine an
organization's legal applicability from technical findings alone.

## Sources and review workflow

Start with a manual source register and reviewed application releases; add automated
discovery only after its operating cost, source access and ownership are approved.
[EUR-Lex RSS alerts](https://eur-lex.europa.eu/content/help/my-eurlex/my-rss-feeds.html?locale=en)
support monitoring documents and saved searches. The
[ESMA DORA reference page](https://www.esma.europa.eu/esmas-activities/digital-finance-and-innovation/digital-operational-resilience-act-dora)
separates primary legislation, technical standards/delegated regulations and guidance;
track these separately rather than treating every publication as an effective rule.
Include NIS2 and applicable national implementing sources in a jurisdiction-specific
register agreed with the compliance reviewer. ISO changes require a separate,
licensed review process; do not scrape or redistribute protected standard text.

Proposed states: detected -> triaged -> reviewed -> approved -> released, with
rejected and superseded branches. A detected publication is a review candidate,
never an active rule. Record official identifier/URL, document type, jurisdiction,
publication date, effective/applicability dates if explicitly established, retrieval
time, content digest, reviewer, rationale and affected controls. Unknown dates stay
unknown. Distinguish drafts, corrections, guidance and binding instruments. Human
compliance review decides relevance; engineering review validates mappings and tests;
the PO approves product scope and release. No external messages are sent by default.

## Versioning and compatibility contract

The first phase keeps the existing code-based catalogue. Each reviewed change gets
a new rule version, source references, release notes and an impact list: added,
changed, retired controls and affected manual evidence. Never repurpose a control
identifier for a different obligation. Changed semantics require explicit revision
and reassessment; preserve prior evidence and label it as requiring review rather
than silently treating it as proof of the new obligation. Newly added controls start
unknown. Existing technical gaps remain visible. Show denominator and score changes
as catalogue changes, not inferred remediation progress.

Before automated activation is considered, extend the current snapshot contract to
retain an immutable catalogue manifest/digest and assessment revision. Today's
rule-version label and MVCC identifier alone cannot reproduce an old report.
Persisted manifests, revision-bound evidence and export references require separate
API/schema design and versioned up/down migrations. Preserve downloaded historical
reports; do not rewrite evidence or claim historical reproducibility before this
storage exists. Releases must record activation time and previous version; rollback
restores the prior catalogue while retaining evidence and activation history.

## Proposed user experience

Show installed catalogue version, last successful source check, last reviewed date
and pending reviews as distinct fields. A failed check displays stale monitoring;
it does not imply no updates. Reviewers see the official source, semantic diff,
effective dates, affected controls/evidence and expected denominator impact. Readers
see a notice after activation and which assessments need review. Existing Refresh
continues to reload assessment data; it never activates a catalogue. Do not display
these future monitoring fields as working features until implemented and verified.

## Collector constraints and failures

If approved, use a daily configurable polling job with bounded retries/backoff,
conditional requests and deduplication by official identifier/revision/digest.
Support manual registration when a feed is unavailable. Use an allowlist of official
HTTPS sources, validate every redirect and reject private/local destinations; apply
timeouts, response-size limits and safe XML parsing without external entities.
Treat retrieved text as untrusted data, never executable instructions. No tenant
assessment data or credentials leave the application for source retrieval. Source
outages, malformed feeds and duplicate items must not alter the active catalogue.
Any new dependency, hosted provider, credential or ongoing cost needs separate review.

## Delivery slices and acceptance gates

1. Source register plus manual review/release checklist, named operational owner and
   review cadence; verify one synthetic amendment from detection to release notes.
2. Immutable catalogue revisions, evidence reassessment and activation history;
   define DTO/auth/errors and migrations first. Verify tenant isolation, rollback,
   old/new report provenance, control retirement and denominator changes.
3. Optional discovery collector and reviewer UI; verify duplicate/out-of-order items,
   source outage, drafts versus effective acts, stale status, malicious URLs/XML and
   concurrent approvals. Browser desktop/mobile/keyboard and read-only roles required.

Each code slice needs the applicable Final gates; storage changes also require real
PostgreSQL and up/down checks. Use synthetic publications for tests, not assertions
that a live feed is complete. Regulatory reviewer sign-off is separate from software
test PASS. Implementation, scheduling and source-account setup remain pending scope
approval; this document does not start M5 or change current installed rules.
