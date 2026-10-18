# E28-S1-T4 · Durable alerting hub

2026-10-18 · working tree. Design in progress; no new deployment, outbound
message or provider credentials authorized by this plan.

## Contract before implementation

- TenantAdmin/SuperAdmin manage channels; Auditor/SecOps cannot change destinations
  or send arbitrary test messages. Channels may cover tenant or a root company.
  Existing company scope remains enforced by tenant-qualified lookups.
- `/api/v1/alert-channels`: GET list, POST create; `/{id}` PUT update and DELETE
  soft delete. `POST /test` tests a complete unsaved configuration with a fixed
  synthetic message; no arbitrary message body. Explicit UI acknowledgement that
  the message leaves the system is required. Provider acceptance is distinct from
  final delivery. No server-side external test during development.
- Config contains provider endpoint/token or SMTP settings/HMAC secret. It is
  encrypted with tenant-derived AES-GCM, bound to channel identity, never returned
  in list/update/error/log output. Omitted config on update preserves credentials;
  provider change requires full replacement. No shared unencrypted URL fields.
- Providers: Slack incoming webhook, Teams Workflows Adaptive Card, Telegram Bot
  sendMessage, generic JSON webhook with timestamp/event-ID/HMAC-SHA256, SMTP
  with mandatory TLS (implicit 465 or STARTTLS 587). Only HTTPS public destinations,
  DNS validated at connection time, no redirects, no proxy environment forwarding.
  Private relays/webhooks need a separately reviewed explicit allowlist; not an
  automatic unrestricted SSRF exception. No new application dependency.
- Migration 000012 pairs up/down. Channel, observation and outbox tables; tenant
  ownership and company validation. Source DB triggers enqueue in the same
  transaction as findings/discovery changes, avoiding an in-memory event gap.
  Existing source history baselined without sending retroactive notifications.
  Events: newly observed HIGH/CRITICAL CVE, new host/port and meaningful discovery
  identity/port changes. Latency/timestamps excluded from identity changes.
- Outbox at-least-once, bounded exponential retry, persistent attempt/dead state,
  unique dedup keys, leased claims with tokens to reject stale acknowledgements.
  Disable/delete prevents new claims; an in-flight request may complete. Provider
  idempotency is not guaranteed; never claim exactly-once delivery.
- Scan schedule terminal status and TLS-certificate expiry require their own
  observed data contracts. Do not infer certificate expiry from banner text or
  silently claim every FR-19.4 trigger implemented from the T4 DoD alone.
- Gate: source/trigger atomicity, tenant/company refusal, migration up/down/up on
  dedicated DB, lease/retry/concurrency, SSRF/DNS/redirect/TLS refusals, HMAC exact
  bytes, redaction, provider response handling, cancellation/shutdown; Final All
  + PostgreSQL, self security review, Doc Sync. T5 provides settings/browser QA.

## External protocol references checked 2026-10-18

- [Slack incoming webhooks](https://api.slack.com/messaging/webhooks)
- [Teams Workflows](https://learn.microsoft.com/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook)
- [Teams connector retirement](https://devblogs.microsoft.com/microsoft365dev/retirement-of-office-365-connectors-within-microsoft-teams/)
- [Telegram Bot API](https://core.telegram.org/bots/api)

No access to a provider account, secret, customer destination or real message is
required to implement or test the adapters with isolated HTTP/SMTP fixtures.

## PO deferral — 2026-10-18

Deferred until the end of the project together with the paired backend/frontend alerting task. Prepared work is preserved; implementation, deployment and live testing are stopped. Task remains incomplete and unaccepted. The integrated run 20261018-193150-594acff4 passed backend tests/vet/build, 229 frontend tests/build and PostgreSQL integration; overall exit 1 due to Markdown EOF whitespace. Browser QA and live provider delivery remain unverified.
