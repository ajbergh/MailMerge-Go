# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/), and the project aims to follow
semantic versioning.

## [Unreleased] — Release hardening

### Added
- Persisted immutable campaign snapshots with durable campaign IDs, parent/child
  retry lineage, run numbers, and retry-after-restart support.
- Campaign-history UI with refresh, retry-failed, delete-one, and clear-all actions.
- Accessible test-send workflow with independent destination and merge-data contact,
  optional `{{email}}` replacement, rendered preview, and save-to-Drafts mode.
- Structured preflight review showing recipients, warnings, blocking errors,
  estimated duration, attachment size, and delivery mode.
- Application error boundary and Windows-aware attachment-path deduplication.
- Backend and frontend coverage artifacts in CI.
- CycloneDX SBOM generation for Go and frontend dependencies.
- Manual release-candidate workflow and a documented release go/no-go checklist.
- Tagged-release manifest, expanded checksums, and optional Authenticode signing
  hooks driven by repository secrets.

### Changed
- Migrated golangci-lint to its v2 configuration and action line.
- Made `govulncheck` and production npm audit blocking release gates.
- Corrected campaign result semantics for partial failure, stopped-on-failure,
  cancellation, preflight failure, and runtime failure.
- Hardened Outlook COM startup, error classification, and worker shutdown.
- Corrected Outlook capability reporting so unsupported account/shared-mailbox
  selection is not advertised.
- Campaign history now performs legacy-record normalization and atomic durable writes.
- Replaced browser prompt/confirm send flows with accessible application dialogs.
- Synchronized the frontend lockfile and pinned Node 22.23.1 across CI and release.

### Security
- Dependency scans now fail CI on reachable Go vulnerabilities or high-severity
  production npm findings.
- Release artifacts include dependency SBOMs and SHA-256 checksums.

## [Unreleased] — Initial remediation foundation

### Added
- Canonical merge-field system (`backend/mergefield`): `{{field_id}}` and legacy
  `{Field}` syntax, `{{field|fallback}}`, collision detection, Unicode/space/hyphen
  support, structured diagnostics.
- `net/mail`-based email validation and address-list parsing (`backend/email`),
  with an explicit duplicate-resolution policy.
- Campaign engine (`backend/campaign`): `EmailSender` interface, deterministic
  `FakeSender`, single `Renderer`, full preflight, context-cancellable runner with
  an injectable clock, typed `CampaignResult` with per-recipient attempt history,
  validated `SendOptions`.
- Classic-Outlook COM sender on a dedicated STA worker thread (`backend/outlook`),
  with a non-Windows stub and actionable capability detection.
- HTML sanitization (bluemonday) + email normalization (`backend/htmlutil`);
  DOMPurify-sanitized, sandboxed previews.
- Local suppression list and campaign history (repository interface + JSON store).
- Atomic, validated persistence for settings and templates; settings schema
  versioning/migration.
- Contact import improvements: UTF-8 BOM handling, stable contact IDs, duplicate
  header detection, blank-row skipping, size/row/column limits.
- Frontend: preflight-gated sending, Cancel, backend-driven previews, stable-ID
  selection, attachment size wiring, accessibility labels, Vitest tests.
- GitHub Actions CI + tag-based release; reproducible version stamping (ldflags);
  `npm ci` in build.
- Documentation: architecture, merge fields, campaign lifecycle, Outlook
  compatibility, testing, security, Graph sender design; CONTRIBUTING/SECURITY.

### Changed
- Aligned the Wails module with the pinned v2.11.0 CLI, switched Wails frontend
  installs to `npm ci`, and added Windows, Linux, and macOS build entry points.
- Upgraded the frontend build/test toolchain to Vite 8, Vitest 4, and TypeScript
  6; Wails timestamp bindings expose RFC 3339 JSON timestamps as TypeScript strings.
- `SendBulkEmails` returns a typed `CampaignResult`; Outlook COM is isolated from
  `backend/services`.
- Removed the hard-coded 500 ms send delay; pacing is sourced from settings.
- README corrected: no unqualified production-ready claim, Go 1.24, and New Outlook
  is unsupported for sending.

### Removed
- Unused `@tanstack/react-table` frontend dependency.
- Hard-coded application version/author metadata in favor of build-time values.
