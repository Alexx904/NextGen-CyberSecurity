# Compliance and assessment lifecycle

Current scope: [automatic finding context](../04-development/plans/M4-AUTOMATIC-CONTEXT.md)
in Findings and Executive/Technical PDFs, without manual compliance input. Manual
UI/API and Audit export below are retired; code/schema/history remain. Remaining
lifecycle sections describe historical contracts, not an active product workflow.
Nmap results never produce a legal compliance score.

Updated: 2026-10-19 - working tree. M4 implemented and verified, awaiting PO
acceptance. Evidence and limits: [M4 plan](../04-development/plans/M4-COMPLIANCE.md).

## Runtime contracts

The guided queue derives next actions from each snapshot: unlinked positive
observations, unresolved issues (SLA breaches first) and unknown/partial/gap controls.
Framework selection filters control tasks; technical tasks remain scope-wide.
Same-group historical host matches suggest assets but require identity confirmation;
ambiguous matches stay unselected. Creation followed by failed association retains
the created asset for retry. Explicit bulk owner/SLA configuration preserves expected
versions, reports per-issue failures and never silently retries conflicts. Control
tasks open the existing evidence form with the relevant control and group selected.
Every 30 seconds, visible idle pages refresh data; edits, writes and open disclosures
pause polling. Existing scan navigation retains consent, and verified closure still
requires referenced coverage/absence attestation. No unattended matching or retest
executor is introduced. [Scope and verification](../04-development/plans/M4-WORKFLOW.md).

The Compliance page has an expandable first-use guide covering score denominators,
unknowns, evidence, explicit asset association, remediation, SLA/MTTR and Audit PDF
interpretation. It explicitly distinguishes refreshing recorded data from updating
the installed regulatory mappings. Regulatory monitoring is not implemented; see
the [proposed maintenance design](../04-development/plans/M4-REGULATORY-UPDATES.md).

`/compliance` extends assessment navigation and the company selector. Server
validates root/department membership in the JWT tenant. Selection filters scope;
tenant readers retain tenant-wide access. SecOps/admin write; Auditors read/export.
Stable assets are explicitly registered and linked to selected observations by an
operator. IP/hostname/range/banner never silently identify a host. Reuse asset UUID
across scans and address changes; unresolved observations remain outside MTTR.

Unique issues use tenant/asset/transport/port/key. Source snapshots retain
finding/scan/target/group, host, confidence/method and versioned provenance.
Replay is idempotent; changed source evidence/rebinding conflicts. Old observations
cannot overwrite newer lifecycle actions. Missing/incomplete scans never close
issues. Positive evidence reopens them. Mitigation, declared correction and verified
correction are distinct; verification requires a referenced operator attestation,
complete issue coverage and explicit absence. No remote test is executed or authorized.
Writes lock/replay/append with an expected version. Assignment is an active tenant
user; SLA is explicit policy ID/version/hours (0 removes it). Deadline anchors to
current-cycle first observation, with no pauses or assumed regulatory deadline.
Current elapsed remediation time exists only after verification. Mean MTTR includes
all verified completed cycles in scope, excludes unresolved cycles and retains
pre-recurrence history. Breach means unresolved and strictly past deadline.

## Controls and scores

Curated rules `2026-10-19.1`: ten NIS2 Art.21(2) categories, fifteen DORA ICT
risk/testing articles, ISO 27001:2022 A.8.8/A.8.20/A.8.24. Sources/identifiers and
paraphrases only; DORA/ISO subsets are explicit, not full regulatory audits.
CVEs relate to vulnerability/network controls; FTP/Telnet and cleartext HTTP relate
to selected cryptography controls. These conservative gap indicators are not legal
violations. Closed/filtered ports are excluded; detection confidence remains visible.

Manual evidence appends actor/time/rationale/reference/source-version and status:
unknown, satisfied, partial, not_satisfied, not_applicable. Latest evidence per
group/control applies. Open technical gaps dominate positive claims and exclusions.
Missing/incomplete latest coverage prevents positive technical credit. Complete
scans without findings still require manual evidence. Organizational controls stay
unknown until assessed; company evidence never silently asserts department evidence.

Score credit: satisfied 1, partial 0.5, otherwise 0. Unknown remains in denominator;
only evidenced not-applicable controls are excluded. All excluded gives N/A.
Assessed/applicable, unknown and exclusion counts accompany every score. Company
cards aggregate direct scope and departments conservatively: gap, unknown, partial,
then satisfied. Exclusion is neutral unless every group excludes the control.

## Snapshot and reporting

Repeatable-read snapshots include server cutoff, MVCC identifier, stable ordering,
latest attempt per target (including incomplete), configured-target counts and full
issue/evidence history. Snapshot ID documents visibility, not a stored replay token.
Whole scoped histories are loaded without silent truncation; high-volume pagination
and materialization remain scalability limitations.

[API](../03-architecture/LLD/API_SPEC.md#historical-m4-compliance-api---retired),
[schema/rollback](../03-architecture/LLD/DATABASE_SCHEMA.md#m4-assessment-lifecycle).
Audit uses existing `pkg/report` and Reports UI, with no parallel archive. One
authorized snapshot supplies all A4 sections. The cover records declared legal
entity/details, audit date, optional reviewer (no digital signature), declared
review period, scope and cutoff. Period labels the review; it does not filter the
current-state snapshot. Checklist, denominators, evidence, source IDs/versions,
remediation history/policies/owners and limitations follow. Never a certificate.

## Verification

Domain: unknown/exclusion denominators, incomplete coverage, rollup, retained gaps,
replay/MTTR. PostgreSQL: scope, deduplication/conflicts, concurrent versions,
immutable audit, old observations/reopening and disposable 000014 down/up. HTTP:
authentication, Auditor write denial, strict DTO, trusted actor, error mapping and
one authorized snapshot/PDF. UI: unknown/empty, evidence, Auditor, retry and stale
company responses. Final/browser/PDF evidence: M4 plan.

Scan provenance includes profile ID and the immutable authorization/parameter
snapshot where present; legacy/synthetic scans explicitly lack that snapshot.
COMPLETED means that profile finished, not that every host/port/control was
assessed. Empty company containers do not count as extra unscanned assets in
technical rollups. Manual evidence has no automatic expiry in this increment;
reviewers must inspect its timestamp/source version. Cross-department asset
transfers require a future audited reassignment design; current associations
require the original group, while address changes can reuse the asset identity.
