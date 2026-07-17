# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/), and the project aims to follow
semantic versioning.

## [Unreleased] — Remediation

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
- GitHub Actions CI + tag-based release; golangci-lint config; reproducible
  version stamping (ldflags); `npm ci` in build.
- Documentation: architecture, merge fields, campaign lifecycle, Outlook
  compatibility, testing, security, Graph sender design; CONTRIBUTING/SECURITY.

### Changed
- Aligned the Wails module with the pinned v2.11.0 CLI, switched Wails frontend
  installs to `npm ci`, and added Windows, Linux, and macOS build entry points.
- Upgraded the frontend build/test toolchain to Vite 8, Vitest 4, and TypeScript
  6; Wails timestamp bindings now expose RFC 3339 JSON timestamps as TypeScript
  strings.
- Updated current documentation and active code comments to describe the
  canonical merge pipeline, supported build commands, and current toolchain.
- `SendBulkEmails` returns a typed `CampaignResult` (was an untyped summary);
  Outlook COM removed from `backend/services`.
- Removed the hard-coded 500 ms send delay (now sourced from settings).
- README corrected: no unqualified "production-ready" claim, Go 1.24, New Outlook
  is unsupported for sending.

### Removed
- Unused `@tanstack/react-table` frontend dependency.
- Hard-coded `1.3.0` / `Your Name` app metadata (now build-time version vars).
