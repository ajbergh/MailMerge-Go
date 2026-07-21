# MailMerge-Go Remediation Status

Last updated: 2026-07-21

Implementation branch: `agent/post-pr-full-remediation`  
Integration PR: #2 — Post-PR full remediation: stabilization and campaign safety

## Readiness summary

The remediation branch has a green automated integration baseline across five permanent CI jobs:

1. Go formatting, vet, race tests with coverage, and golangci-lint.
2. Frontend installation, tests with coverage, and production build.
3. Reachable Go vulnerability scanning and production npm audit.
4. Go and frontend CycloneDX SBOM generation.
5. Windows Wails compilation, root-package tests, and executable artifact upload.

The implementation is ready to enter `main` after the final documentation head passes those same jobs. Publishing a version tag remains separately gated by the clean-Windows/classic-Outlook checklist in `dev_docs/release-checklist.md`.

## Phase status

### Phase 0 — Stabilize Main and Fix CI

**COMPLETE**

- Pinned Node 22.23.1 across CI and release workflows.
- Synchronized the frontend lockfile.
- Migrated golangci-lint to v2.
- Made tests, lint, production builds, vulnerability scans, npm audit, SBOM generation, and Windows packaging real gates.
- Documented recommended branch-protection checks.

Operational follow-up: configure the documented protections for `main` in repository settings.

### Phase 1 — Campaign Reliability and Retry Safety

**COMPLETE**

- Retry selects only recipients whose latest attempt failed.
- Retry performs fresh preflight against the current sender, suppression list, templates, addresses, attachments, duplicate policy, and capabilities.
- Campaign states distinguish complete, partial, stopped, cancelled, preflight-failed, and runtime-failed outcomes.
- Attempt history, timing, campaign IDs, parent IDs, and run numbers are retained.
- Retry works from persisted history after application restart.
- Headless execution no longer requires a Wails runtime context.

### Phase 2 — Outlook COM Reliability

**IN PROGRESS — manual compatibility validation remains**

Completed:

- Hardened STA initialization, fatal/transient/recipient/cancelled error classification, cancellation, and worker shutdown.
- Corrected advertised capabilities.
- Added Outlook Drafts delivery for tests and campaigns.
- Isolated Wails progress emission from campaign execution.

Remaining:

- Complete the clean-Windows/classic-Outlook release matrix.
- Expand tests for additional HRESULT and shutdown edge cases.
- Sender-account selection remains deferred and unadvertised.

### Phase 3 — Campaign Persistence and History

**COMPLETE**

- Stores immutable campaign inputs, result details, timing, delivery mode, and retry lineage.
- Uses validated atomic JSON persistence and normalizes legacy records.
- Isolates corrupt records during list operations.
- Supports history detail, retry, delete-one, and clear-all APIs.
- Covers restart-safe retry, immutable snapshots, draft persistence, corruption, traversal, and cleanup in tests.

### Phase 4 — Product Workflow Completion

**IN PROGRESS — release-critical workflows complete**

Completed:

- Accessible test-send dialog with independent destination and merge-data contact.
- Rendered test preview and send-or-draft selection.
- Structured preflight review.
- Campaign-history review, retry, delete-one, and clear-all.

Future enhancements:

- Manual duplicate resolution.
- Suppression-management UI.
- Contact editing, exclusion, and cleaned export.
- Multi-sheet Excel selection and column-mapping profiles.

### Phase 5 — Frontend Maintainability

**IN PROGRESS**

Completed:

- Removed browser prompt/critical confirm flows.
- Added an error boundary and fixed stale theme settings.
- Added typed Wails history bindings.
- Split test-send, preflight, history, and error-boundary concerns into components.

Remaining:

- Continue decomposing `App.tsx`.
- Normalize categorized Wails errors across all screens.

### Phase 6 — Attachment Reliability

**IN PROGRESS — release-critical handling complete**

- Caches metadata per unique resolved path.
- Avoids repeated stats for static attachments.
- Deduplicates equivalent Windows paths.

Remaining: packaged-app file-drop validation and personalized-path cache coverage.

### Phase 7 — Import Scalability

**IN PROGRESS**

The current importer enforces limits and handles stable IDs, BOMs, duplicate headers, and blank rows. Dataset/schema modeling, multi-sheet mapping, virtualization, and explicit 10,000-contact validation remain future work.

### Phase 8 — Security and Privacy

**IN PROGRESS — automated security baseline complete**

Completed:

- Blocking Go vulnerability and production npm scans.
- HTML sanitization and escaped merge values.
- History deletion controls.
- SBOMs, release manifest, checksums, and optional secret-backed signing hooks.
- README disclosure of local campaign-history contents and user-managed retention.

Remaining: expanded sanitizer vectors, log-redaction audit, configurable retention/export-before-delete, and suppression audit metadata.

### Phase 9 — Tests

**IN PROGRESS — release-critical baseline complete**

- Backend race tests and frontend tests generate coverage artifacts.
- Added campaign state, retry, cancellation, persistence, attachment, draft, restart, snapshot, and headless-execution coverage.
- Windows CI builds the application and tests root Wails-bound APIs.

Remaining: broader Outlook edge-case and frontend interaction coverage, plus reviewed coverage thresholds.

### Phase 10 — CI/CD and Release Engineering

**COMPLETE**

- CI retains coverage, diagnostics, SBOMs, and a Windows executable.
- Release Candidate workflow validates without publishing.
- Tagged releases generate metadata, SBOMs, checksums, and optional Authenticode signing.
- Release checklist documents go/no-go, rollback, and required checks.

### Phase 11 — Microsoft Graph Readiness

**DEFERRED**

The sender abstraction remains compatible with a future Graph implementation. Authentication, token storage, Graph send/draft behavior, shared mailboxes, throttling, tenant restrictions, and sender selection are outside this COM-focused release.

### Phase 12 — Documentation

**COMPLETE**

- Maintained this phase-by-phase status record.
- Consolidated the changelog into one authoritative Unreleased section.
- Updated README coverage of preflight, draft mode, persisted history, restart-safe retry, privacy, testing, release evidence, and known limitations.
- Updated contributor, release-checklist, and release-evidence documentation.

## Final decision

**Merge readiness:** ready after the final documentation commit passes all five permanent CI jobs.  
**Versioned release readiness:** manual Windows/classic-Outlook validation is still required before tagging or publishing a release.
