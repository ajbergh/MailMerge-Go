# MailMerge-Go Post-PR Remediation Status

Last updated: 2026-07-18

Branch: `agent/post-pr-full-remediation`

Draft PR: #2 — Post-PR full remediation: stabilization and campaign safety

Source plan: `MailMerge-Go-Codex-Post-PR-Full-Remediation-Plan(2).md`

## Status legend

- **COMPLETE** — implementation for the phase is complete and validated by the applicable automated checks.
- **IN PROGRESS** — implementation has started but required work or validation remains.
- **BLOCKED** — work is waiting on a prerequisite or unresolved failure.
- **NOT STARTED** — no implementation change has been made for this phase on this branch.
- **DEFERRED** — intentionally postponed to a later milestone/branch.

> No phase is marked COMPLETE while its required CI or platform-specific validation is still red or unavailable.

## Executive status

The remediation branch is active and intentionally remains a draft PR. The first implementation slice addresses CI reproducibility, campaign retry safety, campaign result semantics, Outlook COM failure handling, persistence foundations, attachment metadata caching, and regression tests.

The current gating issue is still Phase 0: frontend `npm ci` fails in GitHub Actions before frontend tests/build can execute. A failure-only `npm-ci-diagnostics` artifact has been added so the exact resolver/install output can be inspected and fixed rather than guessed. Security scanning is now blocking instead of globally ignored; this has surfaced a `govulncheck` failure that also requires triage.

## Phase 0 — Stabilize Main and Fix CI

**Status: IN PROGRESS / RELEASE BLOCKER**

Implemented:

- Pinned the same explicit Node runtime in CI and release workflows.
- Updated the golangci-lint version while retaining the existing v1 configuration format.
- Removed blanket `|| true` behavior from `govulncheck` and production `npm audit`.
- Added a failure-only `npm-ci-diagnostics` Actions artifact for exact npm install failure analysis.
- Preserved `npm ci` as the only dependency installation command in CI/release paths.
- Opened draft PR #2 so every branch update exercises the actual GitHub Actions pipeline.

Validated so far:

- GitHub Actions checkout/setup succeeds.
- `gofmt` succeeds on the remediation branch.
- `go vet ./backend/...` succeeds on the remediation branch.

Remaining:

- Fix the root cause of `npm ci` failure.
- Run frontend tests and production build after install succeeds.
- Run/validate the Windows Wails build after install succeeds.
- Triage the now-blocking `govulncheck` finding.
- Run and fix golangci-lint after race tests pass.
- Document recommended branch-protection required checks.
- Confirm release workflow end-to-end after CI is green.

## Phase 1 — Campaign Reliability and Retry Safety

**Status: IN PROGRESS**

Implemented:

- Retry now identifies only recipients whose latest status failed.
- Retry builds the actual retry recipient subset and performs a fresh preflight before sending.
- Retry preflight now re-evaluates sender availability, suppression, attachments, templates, addresses, duplicate handling, and sender capabilities.
- Historical recipient attempts are preserved and retry attempts are tagged with `trigger: retry`.
- Runner owns `StartedAt`, `FinishedAt`, and `Duration` for terminal results.
- Added explicit `completed_with_failures` and `stopped_on_failure` campaign states in addition to completed, cancelled, preflight failed, and runtime failed.
- Added regression tests for retry after suppression changes, sender unavailability, and removed attachments.

Remaining:

- Finish immutable campaign snapshot integration at the application API layer.
- Change retry entry points to operate by persisted campaign ID rather than frontend `lastRequest` / backend `lastResult` coupling.
- Persist every retry as a distinct linked history record.
- Validate timing/state semantics through green race tests and frontend consumers.

## Phase 2 — Outlook COM Reliability

**Status: IN PROGRESS**

Implemented:

- `RPC_E_CHANGED_MODE` now fails STA worker initialization instead of silently continuing.
- Added structured sender error kinds: recipient, transient, fatal, cancelled.
- Fatal sender errors stop the campaign.
- Outlook startup/availability errors are classified separately from per-message transient errors.
- `Close()` now waits for the COM worker to stop, and submit waits account for worker shutdown.
- Corrected Outlook capabilities: multiple-account and shared-mailbox support are false until explicitly implemented.
- Added draft-only delivery plumbing through `Campaign` -> `RenderedMessage` -> Outlook `MailItem.Save`.
- Fake sender understands draft mode and structured error kinds.

Remaining:

- Add Windows-specific unit/integration coverage around COM initialization outcomes and shutdown races.
- Validate `S_FALSE`, `RPC_E_CHANGED_MODE`, unexpected COM errors, close-during-command, and post-close submission on Windows.
- Expose draft mode in the user-facing test-send/campaign workflow.
- Sender account selection remains deferred and capability remains false.

## Phase 3 — Campaign Persistence and History

**Status: IN PROGRESS**

Implemented:

- Expanded `CampaignRecord` into a fuller run snapshot schema containing original templates, contacts, headers, attachments, CC/BCC, duplicate policy, send options, sender type, result timing, and parent campaign linkage.
- Expanded the repository contract with `Create` and `Update` while retaining `Save` compatibility for existing callers.
- JSON persistence writes through a temporary file, flushes with `Sync`, and renames into place.
- Corrupt individual history files remain isolated during list operations.
- Added `ParentCampaignID` and `RunNumber` fields for non-destructive retry lineage.

Remaining:

- Wire application campaign execution to populate the full snapshot fields.
- Add retry-by-campaign-ID API and persist retries as new linked records.
- Add history detail/duplicate/retry/delete UI.
- Add persistence migration/backfill handling for older JSON records.
- Add persistence tests for create/update/list/reload/corrupt-file/write-failure behavior.

## Phase 4 — Product Workflow Completion

**Status: NOT STARTED**

Remaining:

- Accessible test-send modal with destination, merge-data contact, overwrite-email option, send/draft mode, and rendered preview.
- Structured preflight review modal.
- Duplicate-resolution UI including manual resolution.
- Suppression management UI.
- Contact review/edit/exclude/cleaned-export workflow.
- Multi-sheet Excel selection.
- Column mapping and reusable mapping profiles.
- Import preview with headers, samples, counts, warnings, sheet, and mapping.

Reason not yet started: Phase 0 remains a release blocker; the implementation plan explicitly prioritizes green CI before broad product workflow work.

## Phase 5 — Frontend Maintainability

**Status: NOT STARTED**

Remaining:

- Decompose `App.tsx` into focused workflow hooks.
- Add typed Wails API wrapper modules.
- Make persisted settings the single source of truth and remove stale captured-setting writes.
- Replace browser `prompt()` / critical `window.confirm()` flows with accessible dialogs.
- Add an error boundary and normalized categorized Wails error handling.

Reason not yet started: frontend dependency installation is currently red, so large frontend refactors would be difficult to validate safely.

## Phase 6 — Attachment Reliability

**Status: IN PROGRESS**

Implemented:

- Renderer now caches resolved attachment metadata by resolved path for the lifetime of a render/preflight pass.
- Static attachment paths are no longer re-statted once per recipient in the same renderer pass.
- Personalized attachment paths naturally cache once per unique resolved path.
- Added a regression test asserting one stat for a repeated static attachment path.

Remaining:

- Implement/validate Wails-native packaged-app file drop behavior.
- Deduplicate attachment paths immediately in the frontend with Windows-aware normalization.
- Add explicit per-recipient personalized-attachment preview UX.
- Add a unique-personalized-path cache regression test.

## Phase 7 — Import Scalability

**Status: NOT STARTED**

Remaining:

- Introduce first-class `ImportedDataset` / preserved import schema.
- Enforce explicit file, row, column, and cell-size limits.
- Add large-contact-list virtualization/debounced search and validate ~10,000 contacts.

## Phase 8 — Security and Privacy

**Status: IN PROGRESS**

Implemented:

- Production dependency security checks are now blocking rather than globally ignored.
- Existing HTML sanitizer regression coverage remains in place for script and javascript URL payloads.

Remaining:

- Expand HTML security tests for event handlers, malformed/encoded payloads, SVG vectors, and malicious merge values.
- Audit and redact sensitive logging across campaign/import/export paths.
- Add history retention settings and clear-all/delete-one/export-before-delete workflows.
- Expand suppression records with audit metadata: added date, source, optional reason.
- Resolve or explicitly document current `govulncheck` findings.

## Phase 9 — Tests

**Status: IN PROGRESS**

Implemented:

- Updated campaign state expectations for `completed_with_failures` and `stopped_on_failure`.
- Added runner timing assertions.
- Added retry fresh-preflight tests for suppression changes, sender unavailability, and attachment removal.
- Added retry trigger/history assertions.
- Added draft-mode sender-flow coverage.
- Added static attachment metadata cache coverage.

Remaining:

- Windows Outlook COM initialization/shutdown/error-classification tests.
- Persistence create/update/reload/atomic-failure tests.
- Retry-after-restart test using persisted campaign ID.
- Personalized attachment unique-path cache test.
- Frontend workflow tests listed in the implementation plan.
- Coverage reporting and threshold enforcement.

## Phase 10 — CI/CD and Release Engineering

**Status: IN PROGRESS**

Implemented:

- CI and release workflows share an explicitly pinned Node version.
- Release remains tag-triggered.
- Existing checksum generation and build metadata injection are preserved.
- Security checks have been converted into real gates.

Remaining:

- Achieve green required CI.
- Add backend/frontend coverage reporting.
- Add SBOM generation.
- Document and enforce release checklist/gates.
- Prepare optional Authenticode signing hooks without committing certificates.
- Verify version/commit/build-date display in the About UI.

## Phase 11 — Microsoft Graph Readiness

**Status: DEFERRED**

Current architecture remains compatible with a future Graph sender through `EmailSender`.

Remaining future milestone:

- Dedicated `MicrosoftGraphSender` implementation.
- Delegated auth and secure token storage.
- Mail.Send and draft creation.
- Attachments, shared mailbox, Send As / Send on Behalf Of.
- Throttling / Retry-After handling and 202 Accepted semantics.
- Tenant restrictions and sender-selection UI.

Per the implementation plan, COM is not being replaced during this stabilization slice.

## Phase 12 — Documentation

**Status: IN PROGRESS**

Implemented:

- Added this phase-by-phase remediation status document.
- Draft PR documents the current implementation scope and remains non-mergeable by policy until required gates pass.

Remaining:

- Update README runtime prerequisites after the final Node/npm fix is confirmed.
- Update ROADMAP, CHANGELOG, SECURITY, and CONTRIBUTING.
- Document campaign lifecycle, snapshot persistence schema, retry semantics, sender error classification, Outlook compatibility, and release checklist.
- Add Graph sender design document.

## Current CI evidence

Baseline merged-PR CI findings reviewed before remediation:

- Go format: passed.
- Go vet: passed.
- Go race tests: passed on the original remediation PR.
- golangci-lint: failed on the original remediation PR.
- Frontend `npm ci`: failed, blocking frontend tests/build.
- Windows Wails build: blocked at frontend dependency installation.
- Security workflow previously did not provide meaningful gating because scans were globally ignored.

Remediation-branch CI findings observed so far:

- `gofmt`: passing.
- `go vet`: passing.
- Initial race-test run failed after the intentional campaign state change; regression expectations have since been updated and expanded.
- Frontend `npm ci`: still failing; exact diagnostics artifact now enabled.
- `govulncheck`: now surfaces a real blocking failure and requires triage.
- Windows build remains dependent on resolving frontend install.

## Release readiness

**NOT READY FOR RELEASE**

Blocking conditions:

1. `npm ci` is red.
2. Frontend tests/build have not run successfully in current CI.
3. Windows Wails build is not green.
4. Security scan is red and unresolved.
5. Lint has not yet completed successfully on the current remediation branch.
6. Campaign snapshot/retry-by-ID persistence is not fully wired.
7. Required product workflow and frontend-maintainability phases remain outstanding.

## Next execution order

1. Pull and inspect `npm-ci-diagnostics`, fix the exact package/lock/runtime issue, and regenerate the lockfile only if required.
2. Triage and remediate the blocking `govulncheck` finding.
3. Get Go race tests and golangci-lint green.
4. Get frontend tests/build green.
5. Get the Windows Wails build and artifact upload green.
6. Finish persisted campaign snapshot + retry-by-ID + retry-run persistence.
7. Complete Outlook-specific tests and expose draft mode in UX.
8. Proceed through product workflow, frontend decomposition, import scalability, privacy, coverage, release gates, and documentation in the order defined by the implementation plan.
