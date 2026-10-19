# Guided compliance workflow

Superseded by [automatic finding context](M4-AUTOMATIC-CONTEXT.md). The manual page
is no longer mounted; the following scope and evidence are historical.

Updated: 2026-10-19 - working tree. PO authorized implementation of the guided
workflow after M4. Regulatory monitoring remains design-only; M5 is not started.

## Observable contract

Generate an actionable work queue from the scoped assessment snapshot: unassociated
positive observations, unresolved issues and unknown/partial/gap controls. Explain
the next action and show affected controls. Provide a route to the existing scan
workflow, preserving its authorization and consent requirements. Automatically
refresh recorded data while the page is visible, pausing during edits and writes.

Asset candidates come only from same-group historical associations sharing a host
address. They are suggestions, never proof of identity. Confirm identity before
linking; ambiguous candidates require selection. New asset creation and association
are separate audited API operations: if the second fails, retain the created asset
and permit retry without claiming success. No new backend, schema or dependency.

Offer explicit bulk owner/SLA configuration for selected unresolved issues, using
their current expected versions and a user-specified policy/reason. Report partial
success per issue and reload server state after failures. Never retry a stale version
silently or overwrite existing policy automatically. Existing lifecycle editor
continues to require evidence before verified closure; no remote test is executed
by a queue action. Unknown organizational controls need human evidence.

## Verification plan

Frontend risk: wrong asset/scope, duplicate writes, stale snapshots while editing,
partial operations, misleading closure and read-only role bypass. Test ambiguous
matching, cross-group rejection, partial failures, expected versions, polling pause,
queue generation and Auditor behavior. Run Final Frontend and Docs; backend/DB
contracts remain unchanged. Inspect real browser if connected, otherwise explicitly
record missing visual evidence or PO-owned verification. Self-review and Doc Sync.

## Limits

This implements an assisted workflow, not unattended scan execution, asset identity
discovery or automatic legal assessment. Stable identity cannot safely be inferred
from an address. Retesting uses the existing authorized scan workflow; its absence
of findings alone cannot establish complete issue coverage. Regulatory discovery and
catalogue activation remain in the separate regulatory-update design.

## Execution evidence

Final gate from repository root:
`powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1 -Profile Final -Scope Frontend`.
Report `20261019-191432-70ead8c8`, all exit 0: frontend `npm.cmd test` 6.118s,
246 tests/43 files; `npm.cmd run build` 11.313s (TypeScript/Vite); root Docs 0.673s
and diff check 0.063s. Targeted 13 tests passed beforehand (report
`20261019-191416-cdd97cdf`, 1.481s). No application source changed after the final
gate. An initial sandbox run failed before tests because esbuild could not read
its configuration; the approved elevated run passed. A malformed Quick target
was rejected by the runner and corrected to the two explicit test paths.

Security/reliability self-review (not independent review): proposals require same
group and a historical observation, ambiguous matches are not preselected, writes
reuse existing tenant/role checks, issue updates retain expected versions, partial
failure is explicit and no remote operation is started. React renders untrusted
names as text. Polling is cleaned up on unmount and pauses during open disclosures,
focus in forms, writes and lifecycle editing; stale scope responses remain guarded.
New files match existing `frontend/src/pages/Compliance*` domain ownership. Doc Sync
updates requirements, UI/domain contracts, QA, task/state and this plan; unrelated
working-tree changes are preserved. No API/schema/PDF changes, dependencies,
backend/DB reruns, branch, commit or deployment.

Visual verification is pending: browser inventory returned no connected surfaces.
PO explicitly chose "Verifico io il nuovo flusso". This assigns the check to PO;
it does not assert that it has already passed. Earlier visual acceptance of the
static guide does not establish a PASS for this new flow. M4-WORKFLOW-01 stays `[ ]`
until the remaining visual evidence is supplied.
