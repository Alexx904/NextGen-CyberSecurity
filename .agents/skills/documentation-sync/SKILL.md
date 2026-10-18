---
name: documentation-sync
description: Align repository documentation after behavior or governance changes, or inspect documentation drift on an explicit review request.
---

# Documentation synchronization

Determine actual changed files with `git diff --name-only`, `git diff --cached --name-only` and `git status --short`; separate preexisting work. Use [DOMAIN_MAP](../../../docs/03-architecture/DOMAIN_MAP.md) for application paths. Matching rows must identify one primary document; conflicting ownership fails validation. Shared changes may affect additional domains.

Compare only changed contracts/behavior with their canonical requirements, API/schema, domain, UI or operational sections. Requirements describe intent, migrations schema evolution, code implementation and tests evidence. Investigate divergence instead of declaring one source automatically correct. In an audit-only request report drift without rewriting files.

Update the document that owns each changed fact; use links elsewhere. Update indices only when routing/structure changes, ADR only for significant decisions, and PROJECT_STATE only when persistent state changes. Mark historical decisions superseded explicitly; preserve acceptance IDs/states and evidence, including when compacting logs. Historical authorization is scoped to its original task/action.

Keep current state concise; move concluded detail to linked history outside the bootstrap. Do not put screenshots, binaries or temporary logs in `docs/`; use ignored `.agent/artifacts/<task-id>/`. Use portable relative links and real dates/commits (`working tree` when applicable).

Run from root:

```text
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1 -Profile Docs
```

For governance migration also run `scripts/test-agent-harness.ps1` and inspect preserved IDs/states. Structural PASS does not prove semantic agreement or skill activation. Report documents changed and why, measured size differences, remaining uncertainty and any checks not executed. Do not rerun application suites after Markdown-only changes unless command behavior was affected.
