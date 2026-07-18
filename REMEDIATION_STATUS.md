# MailMerge-Go Post-PR Remediation Status

Last updated: 2026-07-18

Branch: `agent/post-pr-full-remediation`

Draft PR: #2 — Post-PR full remediation: stabilization and campaign safety

## Status legend

- **COMPLETE** — implemented and validated by the applicable automated or platform check.
- **IN PROGRESS** — implementation is present but additional validation or scope remains.
- **DEFERRED** — intentionally assigned to a later release milestone.

## Executive status

The release has moved from a broken baseline to a substantially hardened release candidate. The frontend lockfile is synchronized; frontend installation, tests, and coverage run; Go race tests run with coverage; dependency scans are blocking; CycloneDX SBOMs are generated; and a Windows Wails build is part of every CI run.

Campaign execution now persists immutable run snapshots. Failed recipients can be retried by persisted campaign ID after an application restart, and each retry is stored as a distinct child run with parent linkage and an incremented run number. The frontend now exposes draft-only delivery, structured preflight review, an accessible test-send workflow, campaign history, retry/delete/clear controls, Windows-aware attachment deduplication, and an error boundary.

The PR remains a draft until the current post-integration CI run confirms the frontend production build, Windows build/root tests, and golangci-lint against the final source.

## Phase 0 — Stabilize Main and Fix CI

**Status: IN PROGRESS — final integrated CI validation pending**

Completed:

- Pinned Node 22.23.1 consistently across CI and release workflows.
- Synchronized `frontend/package-lock.json`; `npm ci` succeeds.
- Migrated golangci-lint to the v2 configuration and action line.
- `gofmt`, `go vet`, and Go race tests execute as required checks.
- `govulncheck` and production `npm audit` are blocking checks and currently pass.
- Removed all temporary write-enabled remediation jobs and staging files.

Remaining:

- Confirm golangci-lint is green on the final integrated source.
- Confirm the final frontend production build and Windows package/root tests are green.
- Configure the documented branch-protection required checks in repository settings.

## Phase 1 — Campaign Reliability and Retry Safety

**Status: COMPLETE**

- Retry selects only recipients whose latest attempt failed.
- Every retry performs fresh preflight against current sender availability, suppression, templates, addressing, attachments, duplicate handling, and capabilities.
- Attempt history is preserved and retry attempts are marked with `trigger: retry`.
- Runner-owned start, finish, and duration values are persisted.
- Distinct completed, completed-with-failures, stopped-on-failure, cancelled, preflight-failed, and runtime-failed states are retained.
- Campaign results return durable campaign IDs, parent IDs, and run numbers.
- Retry by persisted campaign ID works after application restart.
- Each retry is persisted as a separate linked history record.

## Phase 2 — Outlook COM Reliability

**Status: IN PROGRESS**

Completed:

- `RPC_E_CHANGED_MODE` fails STA initialization instead of silently continuing.
- Sender failures are classified as recipient, transient, fatal, or cancelled.
- Fatal sender failures terminate campaigns safely.
- COM worker shutdown waits for worker completion.
- Unsupported multiple-account/shared-mailbox capabilities are no longer advertised.
- Draft-only delivery is implemented through Outlook `MailItem.Save` and is exposed in test and campaign UI.

Remaining:

- Expand Windows-specific tests for `S_FALSE`, unexpected HRESULTs, close-during-command, and post-close submission.
- Sender-account selection remains deferred; its capability remains false.

## Phase 3 — Campaign Persistence and History

**Status: COMPLETE**

- Campaign history stores immutable templates, contacts, headers, attachments, CC/BCC templates, duplicate policy, send options, draft mode, sender type, timing, result, and retry lineage.
- Application execution populates the full snapshot.
- JSON writes use temporary files, file synchronization, and atomic rename.
- Repository create/update semantics are explicit.
- Legacy records are normalized on load with lineage and timing backfill.
- Corrupt records are isolated during list operations.
- Persistence tests cover create, update, list, reload, delete, duplicate/missing semantics, corrupt records, legacy normalization, traversal rejection, and temporary-file cleanup.
- Campaign-history APIs support detail retrieval, retry, delete-one, and clear-all.

## Phase 4 — Product Workflow Completion

**Status: IN PROGRESS**

Completed:

- Accessible test-send modal with destination, merge-data contact, overwrite-email option, send/draft selection, and rendered preview.
- Structured preflight modal showing recipients, estimated duration, attachment size, warnings, blocking issues, and delivery mode.
- Campaign-history review, retry, delete-one, and two-step clear-all workflow.
- Draft-only campaign mode.

Remaining:

- Manual duplicate-resolution UI.
- Suppression management UI.
- Contact edit/exclude/cleaned-export workflow.
- Multi-sheet Excel selection, column mapping, and reusable mapping profiles.

## Phase 5 — Frontend Maintainability

**Status: IN PROGRESS**

Completed:

- Removed browser `prompt()` and the critical send `window.confirm()` workflow.
- Added an application error boundary.
- Fixed stale captured settings during theme updates.
- Added typed Wails bindings for persisted campaign operations.
- Split test-send, preflight, history, and error-boundary concerns into dedicated components.

Remaining:

- Continue decomposing `App.tsx` into focused workflow hooks and API wrappers.
- Normalize categorized Wails errors across all screens.

## Phase 6 — Attachment Reliability

**Status: IN PROGRESS**

Completed:

- Attachment metadata is cached per unique resolved path during render/preflight.
- Static attachment paths are not repeatedly stat-ed for every recipient.
- Frontend attachment selection and file drop deduplicate Windows paths case-insensitively and normalize separators.

Remaining:

- Validate packaged-app native file-drop behavior.
- Add personalized-path preview and unique-path cache regression coverage.

## Phase 7 — Import Scalability

**Status: IN PROGRESS**

Existing import limits and stable contact IDs remain in place.

Remaining:

- First-class imported dataset/schema model.
- Multi-sheet and column-mapping workflow.
- Virtualized contact review and explicit 10,000-contact performance validation.

## Phase 8 — Security and Privacy

**Status: IN PROGRESS**

Completed:

- Reachable Go vulnerability scanning is blocking and currently passes.
- Production npm high-severity audit is blocking and currently passes.
- Existing sanitizer tests cover script and JavaScript URL payloads.
- History delete-one and clear-all controls are available.
- Release and CI generate dependency SBOM evidence.

Remaining:

- Expand sanitizer tests for encoded/event-handler/SVG vectors.
- Complete sensitive-log redaction audit.
- Add configurable history retention and export-before-delete.
- Add suppression audit metadata.

## Phase 9 — Tests

**Status: IN PROGRESS**

Completed:

- Backend race tests generate coverage evidence.
- Frontend Vitest runs generate coverage evidence.
- Added retry fresh-preflight, timing/state, cancellation, draft mode, attachment cache, persistence, retry-after-restart, and immutable-snapshot coverage.
- Windows CI now compiles/tests the root Wails-bound package after building.

Remaining:

- Outlook COM edge-case coverage on Windows.
- Additional frontend interaction tests for the new dialogs/history workflow.
- Define and enforce minimum coverage thresholds after baseline review.

## Phase 10 — CI/CD and Release Engineering

**Status: COMPLETE**

- CI generates backend and frontend coverage artifacts.
- CI generates Go and frontend CycloneDX SBOMs.
- Security scans are real gates.
- The Windows executable is built and retained as an artifact.
- A manual Release Candidate workflow validates tests, scans, coverage, SBOMs, and a Windows package without publishing a release.
- Tagged releases generate SBOMs, release metadata, comprehensive checksums, and optional Authenticode signing through secrets.
- The release checklist documents go/no-go validation and recommended required checks.

Operational item outside source control:

- Configure `main` branch protection to require the five checks listed in `dev_docs/release-checklist.md`.

## Phase 11 — Microsoft Graph Readiness

**Status: DEFERRED**

The `EmailSender` abstraction remains compatible with a future Graph implementation. Delegated authentication, secure token storage, `Mail.Send`, draft creation, shared mailbox behavior, throttling, tenant restrictions, and sender selection are intentionally deferred beyond this COM-focused release.

## Phase 12 — Documentation

**Status: IN PROGRESS**

Completed:

- Maintained this phase-by-phase status document.
- Added a release go/no-go checklist and branch-protection guidance.
- Release workflows now describe generated evidence through their artifact names and manifests.

Remaining:

- Complete final README/CHANGELOG wording after integrated CI is green.
- Expand end-user campaign history, draft mode, and privacy documentation.

## Current validation evidence

Confirmed on the integrated branch before the final nullability correction:

- `npm ci`: passed.
- Frontend tests and coverage: passed.
- `gofmt`: passed.
- `go vet ./backend/...`: passed.
- Go race tests and backend coverage: passed.
- `govulncheck ./backend/...`: passed.
- `npm audit --omit=dev --audit-level=high`: passed.
- Go/frontend CycloneDX SBOM generation: passed.

The final CI run is revalidating:

- Frontend TypeScript/Vite production build after the retry nullability correction.
- golangci-lint v2 findings against the integrated source.
- Windows Wails build and root-package persistence/retry tests.

## Release readiness

**CONDITIONALLY NOT READY UNTIL FINAL CI IS GREEN**

There are no longer known dependency-install or security-scan blockers. The remaining release gate is confirmation that every final integrated required check passes. The PR should remain draft until that evidence is available and the Windows/Outlook manual checklist is completed.
