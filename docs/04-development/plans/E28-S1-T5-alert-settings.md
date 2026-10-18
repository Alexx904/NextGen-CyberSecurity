# E28-S1-T5 · Alert channel settings

2026-10-18 · working tree. Shared contract and backend evidence are in the
[T4 plan](E28-S1-T4-alerting.md); no duplicate provider protocol specification.

## Observable acceptance criteria

- Only tenant admins can manage/read destination metadata. Show loading, retry,
  errors and empty state; no fetch for SecOps/Auditor.
- Create initially disabled; explicit tenant/root-company scope and event set.
  Per-provider fields, required TLS and network-policy guidance, no credential
  readback. Update metadata without replacing config; changing provider requires
  replacement. Revision conflict asks user to reload.
- Single synthetic test needs explicit consent; config changes invalidate consent
  and previous result. Test acceptance is not final delivery and never activates
  or saves the channel. No automated external test by the coding agent.
- Two-step removal explains scrubbed config and already in-flight requests.
  Component-held drafts clear on cancel/save; no localStorage or diagnostic data.
- Responsive form, labels, focus on open and keyboard-operable controls. Browser
  QA on an explicitly authorized updated lab, real backend/DB, no mocked browser
  responses. Private-destination refusal is a valid negative test, not a claimed
  successful provider delivery.

## Development evidence

`npm.cmd test -- src/components/AlertChannelsCard.test.tsx` from frontend: six
tests PASS exit 0, 1.740s wall. Covers role guard, test consent/reset, no save on
test, no send on save, omitted config/revision, two-step delete and SMTP default
STARTTLS port. Backend protocol tests separately cover local TLS SMTP acceptance;
no public Slack/Teams/Telegram/SMTP account contacted. Final integrated and browser
gates pending.

## PO deferral — 2026-10-18

Deferred until the end of the project together with the paired backend/frontend alerting task. Prepared work is preserved; implementation, deployment and live testing are stopped. Task remains incomplete and unaccepted. The integrated run 20261018-193150-594acff4 passed backend tests/vet/build, 229 frontend tests/build and PostgreSQL integration; overall exit 1 due to Markdown EOF whitespace. Browser QA and live provider delivery remain unverified.
