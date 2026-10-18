---
name: repo-verification
description: Verify this repository's tests and builds for an explicit verification request or the final gate of a code change.
---

# Repository verification

Work from the repository root. Apply `AGENTS.md` section 8; command ownership and environment requirements are in [TEST_PLAN](../../../docs/05-qa/TEST_PLAN.md#strategia). Read only the relevant strategy/runner sections, not historical feature reports.

Select the smallest applicable profile of `scripts/verify.ps1`:

| Request | Arguments |
|---|---|
| Backend iteration | `-Profile Quick -Scope Backend -BackendPackages ./internal/service` (choose actual packages) |
| Frontend iteration | `-Profile Quick -Scope Frontend -FrontendFiles src/pages/Reports.test.tsx` (choose actual files) |
| Final code gate | `-Profile Final -Scope Backend`, `Frontend` or `All` |
| Repository/DB involved | Add `-Integration` to Final Backend/All |
| Explicit global verification | `-Profile Full` (All and PostgreSQL) |
| Markdown/governance checks | `-Profile Docs` |

Invoke with `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1` on Windows. Check manifest/CI if commands or toolchain changed. Quick requires explicit targets and does not replace the final gate.

Integration requires explicit `TEST_DB_*` in the process and a dedicated local test database. Never load `.env`, fall back to application DB variables, or report a missing DB as a skip. Do not run migrations concurrently on one DB. Browser/PDF checks and migration rollback are additional evidence when pertinent; the runner does not prove them.

Before reuse, compare actual relevant sources/tests including untracked files, dependencies, configuration, toolchain and environment. Reports contain timings and outcomes, not an automatic freshness certificate. External data changes invalidate integration evidence.

Report scope, command, directory, duration, exit code, failures and unexecuted checks. On a verification-only request diagnose/report failures; change application code only if the user requested correction. Update the current baseline only when the run is complete and representative. Keep verbose output in ignored artifacts.
