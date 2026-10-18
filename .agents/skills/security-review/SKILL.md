---
name: security-review
description: Review a repository diff for security and reliability when review is requested or a change affects security-sensitive behavior.
---

# Security review

Define base and scope; inspect staged, unstaged and relevant untracked files, separating preexisting work. Audit is read-only unless fixes were requested. Read only affected domains/contracts and [orchestration](../../../.agent/workflows/orchestration.md) if a handoff is needed.

Check concrete failure paths:

- Commands/SQL: untrusted data remains arguments/parameters; shell quoting and Nmap option boundaries match the actual invocation. No secrets in logs, reports or staged content; inspect only nonsensitive sources.
- Auth/tenant: REST, SSE and WS enforce the intended JWT/ticket, role and tenant/company scope. Check target ownership, stream-ticket lifetime, network scope and DNS rebinding where relevant.
- Lifecycle: cancellation, goroutine/channel ownership, blocked operations, HTTP bodies, SSE/WS listeners, timers and shared-state races. Local SSH cancellation does not prove remote termination or rollback.
- UI/tests: loading/empty/error, focus and keyboard behavior, meaningful assertions and negative/edge cases. No weakened assertions, hidden skips or synthetic evidence presented as real UI/DB validation.
- Contracts: compatibility, migration rollback, affected-domain mapping and current documentation. Use `working tree` when no relevant commit exists.

Report findings by severity with file/line, triggering condition, impact, evidence and smallest remedy. Distinguish verified defects from hypotheses and self-review from independent review. Record substantive unresolved findings in Tech Debt. A clean review means no finding detected in the inspected scope, not an absolute guarantee.
