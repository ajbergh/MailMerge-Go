# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/), and the project aims to follow
semantic versioning.

## [Unreleased] — Full remediation and release hardening

### Added

- Canonical merge-field support through `backend/mergefield`, including
  `{{field_id}}`, `{{field|fallback}}`, legacy `{Field}` compatibility, collision
  detection, structured diagnostics, and support for Unicode and punctuated column
  names.
- `net/mail`-based email validation, address-list parsing, and explicit duplicate
  policy handling through `backend/email`.
- A platform-independent campaign engine with a single renderer, full preflight,
  cancellation, validated send options, typed campaign states, per-recipient
  attempt history, and deterministic fake-sender tests.
- A classic-Outlook COM sender running on a dedicated STA worker, with actionable
  readiness reporting and a non-Windows stub.
- HTML normalization and sanitization, plus DOMPurify-sanitized previews in a
  sandboxed iframe.
- Local suppression-list and campaign-history repositories.
- Persisted immutable campaign snapshots with durable campaign IDs, parent/child
  retry lineage, run numbers, and retry-after-restart support.
- Campaign History UI with refresh, retry-failed, delete-one, and clear-all actions.
- An accessible test-send workflow with independent destination and merge-data
  contact, optional `{{email}}` replacement, rendered preview, and Outlook Drafts
  mode.
- Structured preflight review showing recipients, warnings, blocking errors,
  estimated duration, attachment size, and delivery mode.
- Draft-only delivery for test messages and full campaigns.
- An application error boundary and Windows-aware attachment-path deduplication.
- Backend and frontend CI coverage artifacts.
- CycloneDX SBOM generation for Go and frontend production dependencies.
- A manual Release Candidate workflow and documented release go/no-go checklist.
- Tagged-release manifests, comprehensive SHA-256 checksums, and optional
  Authenticode signing hooks driven by repository secrets.
- Windows root-package tests covering persisted retries, immutable snapshots,
  draft-mode persistence, and headless campaign execution.
- Architecture, merge-field, campaign-lifecycle, Outlook compatibility, testing,
  security, release-evidence, and future Graph-sender documentation.

### Changed

- `SendBulkEmails` now returns a typed `CampaignResult`; Outlook COM is isolated
  from application services behind `campaign.EmailSender`.
- Preview, test send, and campaign execution use the same backend renderer.
- Campaign result semantics distinguish completed, completed-with-failures,
  stopped-on-failure, cancelled, preflight-failed, and runtime-failed outcomes.
- Retry selects only recipients whose latest attempt failed and performs fresh
  preflight against the current environment.
- Campaign history now performs legacy-record normalization and atomic durable
  writes.
- Settings and templates use validated atomic persistence with schema migration.
- Contact import handles UTF-8 BOMs, stable contact IDs, duplicate headers, blank
  rows, and file/row/column limits.
- Outlook COM startup, error classification, capability reporting, cancellation,
  and worker shutdown are hardened.
- Unsupported sender-account and shared-mailbox selection are no longer advertised.
- Browser prompt/confirm send flows were replaced with accessible application
  dialogs.
- Frontend sending uses persisted campaign IDs for restart-safe retry.
- Attachment metadata is cached per resolved path during rendering and preflight.
- Send pacing is sourced from validated settings instead of a hard-coded delay.
- The frontend toolchain was upgraded to Vite 8, Vitest 4, and TypeScript 6.
- Wails timestamp bindings expose RFC 3339 JSON timestamps as TypeScript strings.
- Wails frontend installs use `npm ci`, the lockfile is synchronized, and Node
  22.23.1 is pinned across CI and release workflows.
- golangci-lint was migrated to the v2 configuration and action line.
- `govulncheck` and production npm audit are blocking release gates.
- Linux, macOS, and Windows build entry points are documented consistently while
  Outlook sending remains Windows-only.
- README and contributor documentation now describe campaign history, draft mode,
  local-data retention, CI evidence, and the distinction between merge readiness
  and manual release validation.

### Security

- Imported values are escaped before insertion into HTML messages, and authored
  HTML is sanitized before preview and send.
- Dependency scans fail CI on reachable Go vulnerabilities or high-severity
  production npm findings.
- Template and campaign identifiers are validated against path traversal.
- Release artifacts include CycloneDX SBOMs, a release manifest, and SHA-256
  checksums.
- Authenticode signing is optional and uses repository secrets; certificate and
  password material is never stored in the repository.

### Removed

- The unused `@tanstack/react-table` frontend dependency.
- The hard-coded 500 ms send delay.
- Hard-coded application version and author metadata in favor of build-time values.
- Unqualified production-readiness claims and unsupported New Outlook capability
  claims.
