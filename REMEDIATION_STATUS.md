# MailMerge-Go Post-PR Remediation Status

Last updated: 2026-07-18

Branch: `agent/post-pr-full-remediation`

Draft PR: #2 — Post-PR full remediation: stabilization and campaign safety

## Status legend

- **COMPLETE** — implemented and validated by the applicable automated or platform check.
- **IN PROGRESS** — implementation is present but additional product scope or manual validation remains.
- **DEFERRED** — intentionally assigned to a later release milestone.

## Executive status

The automated release baseline is green. The permanent CI workflow now passes all five required jobs:

1. Go formatting, vet, race tests with coverage, and golangci-lint.
2. Frontend dependency installation, tests with coverage, and production build.
3. Reachable Go vulnerability scanning and production npm audit.
4. Go and frontend CycloneDX SBOM generation.
5. Windows Wails compilation, root-package persistence/retry tests, and executable artifact upload.

Campaign execution now persists immutable run snapshots. Failed recipients can be retried by persisted campaign ID after an application restart, and every retry is stored as a distinct child run with parent linkage and an incremented run number. The frontend exposes draft-only delivery, structured preflight review, an accessible test-send workflow, campaign history, retry/delete/clear controls, Windows-aware attachment deduplication, and an application error boundary.

The PR remains a draft only because the manual Windows/classic-Outlook compatibility checklist has not yet been completed on a clean target system. There are no known automated build, lint, dependency, security, SBOM, frontend, persistence, or Windows compilation blockers.

## Phase 0 — Stabilize Main and Fix CI

**Status: COMPLETE**

- Pinned Node 22.23.1 consistently across CI and release workflows.
- Synchronized `frontend/package-lock.json`; `npm ci` succeeds.
- Migrated golangci-lint to the v2 configuration and action line.
- `gofmt`, `go vet`, Go race tests, and golangci-lint pass.
- Frontend tests, coverage, and the TypeScript/Vite production build pass.
- `govulncheck` and production `npm audit` are blocking checks and pass.
- Windows Wails compilation and root-package tests pass.
- Removed all temporary write-enabled remediation jobs and staging files.
- Recommended branch-protection checks are documented in `dev_docs/release-checklist.md`.

Operational item outside source control:

- Configure `main` branch protection to require the five permanent CI jobs.

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
- Headless campaign execution no longer requires a Wails runtime context.

## Phase 2 — Outlook COM Reliability

**Status: IN PROGRESS — manual compatibility matrix remains**

Completed:

- `RPC_E_CHANGED_MODE` fails STA initialization instead of silently continuing.
- Sender failures are classified as recipient, transient, fatal, or cancelled.
- Fatal sender failures terminate campaigns safely.
- COM worker shutdown waits for worker completion.
- Unsupported multiple-account/shared-mailbox capabilities are no longer advertised.
- Draft-only delivery is implemented through Outlook `MailItem.Save` and exposed in test and campaign UI.
- Wails progress emission is isolated from campaign execution and is installed only by the Wails lifecycle context.

Remaining manual or future work:

- Execute the clean-Windows/classic-Outlook matrix in `dev_docs/release-checklist.md`.
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
- Persistence tests cover create, update, list, reload, delete, duplicate/missing semantics, corrupt records, legacy normalization, traversal rejection, temporary-file cleanup, draft-mode round trip, immutable snapshots, and retry after restart.
- Campaign-history APIs support detail retrieval, retry, delete-one, and clear-all.

## Phase 4 — Product Workflow Completion

**Status: IN PROGRESS**

Completed:

- Accessible test-send modal with destination, merge-data contact, overwrite-email option, send/draft selection, and rendered preview.
- Structured preflight modal showing recipients, estimated duration, attachment size, warnings, blocking issues, and delivery mode.
- Campaign-history review, retry, delete-one, and two-step clear-all workflow.
- Draft-only campaign mode.

Remaining product enhancements:

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
- Added durable frontend coverage and compiler-diagnostic artifacts.

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

- Validate packaged-app native file-drop behavior manually.
- Add personalized-path preview and unique-path cache regression coverage.

## Phase 7 — Import Scalability

**Status: IN PROGRESS**

Completed foundation:

- Existing import limits, stable contact IDs, BOM handling, duplicate-header detection, and blank-row handling remain enforced.

Remaining:

- First-class imported dataset/schema model.
- Multi-sheet and column-mapping workflow.
- Virtualized contact review and explicit 10,000-contact performance validation.

## Phase 8 — Security and Privacy

**Status: IN PROGRESS**

Completed:

- Reachable Go vulnerability scanning is blocking and passes.
- Production npm high-severity audit is blocking and passes.
- Existing sanitizer tests cover script and JavaScript URL payloads.
- History delete-one and clear-all controls are available.
- CI and tagged releases generate dependency SBOM evidence.
- Tagged releases generate a manifest and comprehensive SHA-256 checksums.
- Optional Authenticode signing hooks are available through repository secrets; no certificate material is committed.

Remaining:

- Expand sanitizer tests for encoded/event-handler/SVG vectors.
- Complete sensitive-log redaction audit.
- Add configurable history retention and export-before-delete.
- Add suppression audit metadata.

## Phase 9 — Tests

**Status: IN PROGRESS — release-critical baseline complete**

Completed:

- Backend race tests generate coverage evidence.
- Frontend Vitest runs generate coverage evidence.
- Added retry fresh-preflight, timing/state, cancellation, draft mode, attachment cache, persistence, retry-after-restart, immutable-snapshot, draft persistence, and headless execution coverage.
- Windows CI compiles/tests the root Wails-bound package after building.
- All permanent automated test, lint, build, and security gates pass.

Remaining enhancements:

- Outlook COM edge-case coverage on Windows.
- Additional frontend interaction tests for the new dialogs/history workflow.
- Define and enforce minimum coverage thresholds after baseline review.

## Phase 10 — CI/CD and Release Engineering

**Status: COMPLETE**

- CI generates backend and frontend coverage artifacts.
- CI generates Go and frontend CycloneDX SBOMs.
- Security scans are real gates.
- The Windows executable is built and retained as an artifact.
- Compiler, linter, vulnerability, Wails-build, and root-test diagnostics are retained when useful.
- A manual Release Candidate workflow validates tests, scans, coverage, SBOMs, root-package tests, and a Windows package without publishing a release.
- Tagged releases generate SBOMs, release metadata, comprehensive checksums, and optional Authenticode signing through secrets.
- The release checklist documents go/no-go validation, rollback, and recommended required checks.

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
- Added release-evidence documentation.
- Updated `CHANGELOG.md` and `CONTRIBUTING.md` for the hardened release process.
- Release workflows describe generated evidence through artifact names and manifests.

Remaining:

- Expand end-user documentation for campaign history, draft mode, retention, and privacy controls.

## Current automated validation evidence

The permanent CI workflow is green for the integrated application source:

- `npm ci`: passed.
- Frontend tests and coverage: passed.
- Frontend TypeScript/Vite production build: passed.
- `gofmt`: passed.
- `go vet ./backend/...`: passed.
- Go race tests and backend coverage: passed.
- golangci-lint v2: passed.
- `govulncheck ./backend/...`: passed.
- `npm audit --omit=dev --audit-level=high`: passed.
- Go/frontend CycloneDX SBOM generation: passed.
- Windows Wails build: passed.
- Windows root-package persistence, restart-safe retry, and headless-context tests: passed.
- Windows executable artifact upload: passed.

## Release readiness

**AUTOMATED RELEASE GATES COMPLETE — MANUAL WINDOWS/OUTLOOK VALIDATION REQUIRED**

The codebase has no known automated build, test, lint, dependency, security, SBOM, frontend, persistence, or Windows packaging blocker. Before publishing a version tag, execute `dev_docs/release-checklist.md` on a clean supported Windows system with classic Outlook and a configured mail account. PR #2 should remain draft until that manual compatibility and smoke-test evidence is recorded.