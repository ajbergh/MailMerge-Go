# Roadmap

This roadmap separates **implemented and verified** functionality from **planned**
work. Prior date-stamped roadmap content was archived to
[dev_docs/archive/ROADMAP.md](dev_docs/archive/ROADMAP.md).

## Implemented (verified by tests / build)

- Canonical merge-field system (spaces, hyphens, punctuation, Unicode), fallbacks,
  collision detection, single unified renderer for preview/test/bulk.
- `net/mail` email validation; comma/semicolon address lists; duplicate-resolution
  policy.
- Campaign preflight that blocks unsafe/invalid sends; typed campaign results with
  per-recipient attempt history (fatal failures are not shown as success).
- Validated send options (delay/batch/pause/continue-on-error) sourced from
  settings — the hard-coded 500 ms delay is gone.
- `EmailSender` abstraction with a deterministic `FakeSender`; Outlook COM isolated
  on a dedicated STA worker thread; capability detection.
- Cancellation and retry (failed-only, no double counting).
- HTML escaping of merge values + sanitization of authored HTML; sandboxed,
  sanitized previews.
- Atomic, validated persistence for settings/templates/suppression; campaign
  history behind a repository interface.
- Contact import: BOM handling, stable contact IDs, duplicate-header detection,
  blank-row skipping, size/row/column limits.
- Stable-ID contact selection in the UI; attachment size/total wiring; accessible
  labels.
- CI (tests/lint/build/scan), tag-based release, reproducible version stamping.

## In progress / partial

- Frontend state refactor into focused hooks/stores (wired to the new engine;
  full decomposition pending).
- Broader accessibility pass (focus trapping, screen-reader progress
  announcements).
- Excel sheet selection and column mapping; saved import profiles.

## Planned

- SQLite persistence behind the existing repository interfaces.
- Microsoft Graph sender (see [dev_docs/graph-sender-design.md](dev_docs/graph-sender-design.md)),
  behind a feature flag with auth/security/integration tests.
- Pause/resume during a run; scheduled sends.
- Windows installer polish, code signing, SBOM.

## Release criteria (per milestone)

A milestone is "done" only when: its Go/frontend tests pass, `go vet` +
golangci-lint + `tsc` are clean, the Wails Windows build succeeds in CI, and the
relevant documentation is updated. The app is not called production-ready until
all automated release gates pass and Outlook COM paths are verified on a Windows
workstation with classic Outlook.
