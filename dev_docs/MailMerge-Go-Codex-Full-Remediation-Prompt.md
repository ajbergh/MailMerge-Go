# Implementation Prompt: MailMerge-Go Full Remediation and Production-Readiness Plan

---

# Implementation Status

**Branch:** `remediation/full-remediation` · **Last updated:** 2026-07-16

**Progress summary:** Backend correctness/architecture core is implemented and unit-tested
without Outlook: canonical merge fields, unified renderer, `net/mail` validation,
duplicate policy, validated `SendOptions`, full preflight, typed campaign results,
`EmailSender` + `FakeSender`, dedicated Outlook STA worker (COM isolated from the
runner), HTML escaping + sanitization, suppression list, atomic persistence.
`go build ./...`, `go vet ./...`, and `go test ./backend/...` all pass; the backend
cross-compiles for Linux. Remaining: frontend refactor/wiring, campaign persistence
store, CI/packaging, docs. See table.

**Baseline (captured on branch creation):**
- `go build ./...` failed: `pattern all:frontend/dist: no matching files found` (root package cannot build until the frontend is built or a `frontend/dist` stub exists).
- `go vet ./...` failed for the same reason.
- `go test ./...` → `FAIL MailMergeApp [setup failed]`; `backend/models` and `backend/services` reported **no test files**.
- Toolchain present: Go 1.25.4, Node 24.11.1, npm 11.6.2, Wails v2.11.0.

**Legend:** ✅ done · 🚧 in progress · ⏳ pending · ⏭️ deferred (documented)

| # | Milestone / Section | Status | Notes |
|---|---|---|---|
| — | Baseline, branch, build scaffolding | ✅ | Feature branch + `frontend/dist` stub; baseline recorded. |
| 1.1 | Canonical merge-field system | ✅ | `backend/mergefield`: canonical IDs, collision detection, aliases, tests. |
| 1.2 | Unified template rendering | ✅ | Single `campaign.Renderer` → one `RenderedMessage` for preview/test/bulk. |
| 1.3 | Representative test sends | ✅ | Backend: full contact, no `{{email}}` overwrite by default, same pipeline. Frontend wiring pending. |
| 1.4 | Campaign preflight | ✅ | `campaign.Preflighter` full checklist; `PreflightCampaign` binding. Frontend display pending. |
| 1.5 | Email validation (`net/mail`) | ✅ | `backend/email`; comma/semicolon lists; tests. |
| 1.6 | Duplicate recipient resolution | ✅ | Policy engine (keep first/last/exclude/all/manual); applied pre-count; tests. |
| 1.7 | Attachment remediation | 🚧 | Backend resolves/validates (incl. personalized paths); frontend drag-drop/size wiring pending. |
| 1.8 | Sending settings affect behavior | ✅ | Validated `SendOptions` from settings; 500ms hard-code removed; tests. Sound/auto-save pending. |
| 1.9 | Typed campaign results / fatal errors | ✅ | Typed states; fatal ≠ empty success; tests. |
| 2.10 | EmailSender abstraction + FakeSender | ✅ | Interface + deterministic FakeSender + capabilities; tests. |
| 2.11 | Dedicated Outlook STA worker | ✅ | `backend/outlook` locked STA thread; COM out of runner; non-Windows stub. Needs on-device COM test. |
| 2.12 | Outlook capability detection | ✅ | Actionable status; no false New-Outlook claims. |
| 2.13 | Cancellation / pause / resume | ✅ | Context cancellation + `CancelCampaign`; cancelled recipients preserved. Pause/resume deferred. |
| 2.14 | Retry semantics | ✅ | Attempt-history retry, no double count; `RetryFailed` binding; tests. |
| 3.15 | Safe template storage | 🚧 | Atomic helper ready; hardening in progress. |
| 3.16 | Safe settings persistence | 🚧 | Schema version + policy added; validation/merge in progress. |
| 3.17 | HTML security | ✅ | Escape merge values + bluemonday sanitize authored HTML; tests. |
| 3.18 | Campaign persistence | ⏳ | Repository interfaces planned (JSON first). |
| 3.19 | Suppression / unsubscribe | ✅ | `SuppressionService` (atomic, normalized); preflight blocks. Frontend UI pending. |
| 4.20 | Contact import improvements | ⏳ | BOM/limits/sheet-selection pending. |
| 4.21 | Contact editing and review | 🚧 | Stable `Contact.ID` added; UI editing pending. |
| 4.22 | Frontend state refactor | ⏳ | Hooks/wiring to new bindings pending. |
| 4.23 | Accessibility | ⏳ | |
| 5.24 | Go tests | ✅ | mergefield/email/campaign/htmlutil covered; more to add for services. |
| 5.25 | Windows Outlook integration tests | ⏳ | Build-tagged harness pending. |
| 5.26 | Frontend tests | ⏳ | |
| 5.27 | Static analysis & formatting | 🚧 | gofmt/vet clean; golangci config + frontend lint pending. |
| 5.28 | GitHub Actions CI | ⏳ | |
| 5.29 | Reproducible builds | 🚧 | Version ldflags vars added; build.ps1/CI pinning pending. |
| 5.30 | Packaging and release quality | ⏳ | |
| 6.31 | Microsoft Graph readiness | ⏳ | Design doc + interface pending. |
| — | Documentation remediation | ⏳ | README/ROADMAP/dev_docs pending. |

---

## Role

Act as the senior software engineer responsible for remediating and modernizing the following repository:

- Repository: `https://github.com/ajbergh/MailMerge-Go`
- Default branch: `main`
- Primary stack:
  - Go
  - Wails v2
  - React
  - TypeScript
  - Microsoft Outlook COM automation on Windows
  - CSV and Excel contact import

Your goal is to transform the current functional prototype into a reliable, testable, secure, and maintainable mail-merge application.

Do not treat the current README or roadmap as proof that a feature works. Verify behavior directly in the source and through tests.

---

## Primary Objectives

Implement the complete remediation plan described below.

The finished application must:

1. Render merge fields correctly and consistently.
2. Prevent unsafe or invalid campaigns from being sent.
3. Make test sends representative of real bulk sends.
4. Respect all configured sending options.
5. Correctly handle attachments, duplicates, failures, cancellation, and retries.
6. Isolate Outlook COM behavior behind a testable sender abstraction.
7. Use a safe COM threading and initialization model.
8. Add automated tests, linting, CI, and reproducible builds.
9. Correct misleading documentation, metadata, and roadmap claims.
10. Establish an architecture that can support Microsoft Graph in a later phase.

---

# Mandatory Working Rules

## Repository Safety

- Work from the latest `main`.
- Create a dedicated feature branch.
- Do not rewrite repository history.
- Preserve existing user-facing functionality unless explicitly replaced by a safer implementation.
- Prefer small, reviewable commits grouped by remediation milestone.
- Do not commit generated binaries, build output, credentials, local configuration, or Outlook data.
- Do not remove existing functionality merely because it is difficult to test. Introduce an abstraction and isolate it.

## Verification

Before changing code:

1. Inspect the complete repository structure.
2. Read:
   - `README.md`
   - `dev_docs/archive/ROADMAP.md` (the roadmap was moved here during the "Repo Organization" commit; there is no root `ROADMAP.md`)
   - `app.go`
   - `main.go`
   - `backend/models/contact.go`
   - all files under `backend/services/`
   - `frontend/src/App.tsx`
   - all frontend components
   - `frontend/package.json`
   - `go.mod`
   - `wails.json`
   - `build.ps1`
3. Search for all Wails-bound methods and confirm generated frontend bindings match the Go API.
4. Run the existing build and test commands before modifying code.
5. Record any baseline failures in the final implementation report.

Do not claim Outlook COM behavior is verified unless it was tested on Windows with classic Outlook.

---

# Verified Baseline Findings

The problem statements in this document were confirmed against the current `main` during a source review. Treat the following as verified anchors, not assumptions:

- **Merge parsing** uses `\{(\w+)\}` and silently returns `""` for unknown fields — `backend/services/merge_service.go` (`NewMergeService`, `renderTemplate`). `\w+` cannot match headers with spaces, hyphens, or non-ASCII characters.
- **Frontend preview** does its own substitution with hard-coded sample values instead of the backend renderer — `frontend/src/components/EmailEditor.tsx` (`.replace(/\{FirstName\}/gi, 'John')`, etc.).
- **Email validation** is a bare `@`/`.` substring check — `backend/services/file_service.go` (`extractContact`).
- **Duplicates** are only warned about; every row is still selected by default — `backend/services/file_service.go` (`detectDuplicates`) and `frontend/src/App.tsx` (`handleSelectFile` selects all indices).
- **Test sends** carry only `SampleFirstName`/`SampleLastName` and reuse the test address as the `{Email}` value; no custom fields flow through — `app.go` (`SendTestEmail`) and `backend/models/contact.go` (`TestEmailRequest`).
- **The 500 ms inter-message delay is hard-coded** and ignores the persisted `SendingDelay` (`GetSendingDelay`/`SetSendingDelay` exist but are never read by the send path) — `backend/services/outlook_service.go` (`SendBulkEmails`).
- **AttachmentManager** exposes `onFilesDropped` and `attachmentInfo`, but `App.tsx` renders it without either prop, so drag-drop and size totals are inert (`attachmentInfo` is always undefined) — `frontend/src/components/AttachmentManager.tsx` and `frontend/src/App.tsx`.
- **Fatal attachment / COM / Outlook failures** return an empty `SendResult` (0 sent / 0 failed) that reads as success — `backend/services/outlook_service.go` (`SendBulkEmails`).
- **COM is initialized and torn down inside each method**, and `CheckOutlookInstalled` uses an unchecked `err.(*ole.OleError)` type assertion — `backend/services/outlook_service.go`. (The send methods use the safe comma-ok form; `CheckOutlookInstalled` does not.)
- **Contacts have no stable ID**; selection identity is index-based (`Set<number>`) — `backend/models/contact.go` and `frontend/src/App.tsx`.
- **Template and settings writes are non-atomic** `os.WriteFile` calls; `SaveTemplate` trusts caller-supplied IDs (enabling path traversal and overwrite of built-in template files) and `UpdateSettings` performs no validation; `GetRecentFiles` returns the shared backing slice — `backend/services/template_service.go`, `backend/services/settings_service.go`.
- **No CI exists** (`.github/workflows/` is absent), and `@tanstack/react-table` is listed in `frontend/package.json` but is not imported anywhere in the source.
- **App metadata is hard-coded** `"1.3.0"` / `"Your Name"` in `app.go` (`GetAppInfo`); `README.md` states Go 1.21 while `go.mod` requires `go 1.24.0`; `build.ps1` uses `npm install` and an unpinned Wails CLI.
- **`README.md` still describes the app as "production-ready."** The roadmap now lives at `dev_docs/archive/ROADMAP.md` (there is no root `ROADMAP.md`).

Re-verify each of these before implementing, and record any drift in the baseline findings note.

---

# Architectural Target

Refactor toward the following conceptual layers:

```text
UI / Wails bindings
        |
Campaign application service
        |
+-------------------------+
| Campaign preflight      |
| Template renderer       |
| Contact normalization   |
| Campaign runner         |
| Campaign persistence    |
+-------------------------+
        |
EmailSender interface
        |
+----------------------+------------------+
| Classic Outlook COM  | Fake/Test Sender |
+----------------------+------------------+
```

Use dependency injection rather than creating platform services directly throughout the application.

Recommended interfaces may be adjusted if a better idiomatic Go design is justified.

```go
type EmailSender interface {
    Capabilities(ctx context.Context) SenderCapabilities
    Preflight(ctx context.Context, campaign Campaign) PreflightResult
    Send(ctx context.Context, message RenderedMessage) SendReceipt
}

type CampaignRenderer interface {
    ValidateTemplate(template Template, schema ContactSchema) ValidationResult
    Render(template Template, contact Contact) (RenderedMessage, error)
}

type CampaignRunner interface {
    Run(ctx context.Context, campaign Campaign, progress ProgressSink) CampaignResult
}
```

---

# Milestone 1: Correctness and Send Safety

This milestone is the highest priority. Do not start advanced feature work until these items are complete.

## 1. Canonical Merge-Field System

### Current Problem

The UI generates merge fields from raw column headers, but the backend regex only supports `\w+`. Fields containing spaces, hyphens, punctuation, or non-ASCII characters may appear in the UI but fail to render. Unknown fields are silently replaced with empty strings.

### Required Changes

Implement a canonical merge-field model.

Each imported column must have:

- Original display name
- Canonical identifier
- Normalized lookup key
- Optional aliases
- Data type, initially string
- Required/optional state where applicable

Example:

```go
type MergeField struct {
    ID          string   `json:"id"`
    DisplayName string   `json:"displayName"`
    LookupKey   string   `json:"lookupKey"`
    Aliases     []string `json:"aliases,omitempty"`
}
```

Requirements:

- Support source headers such as:
  - `First Name`
  - `Account-Manager`
  - `Customer ID`
  - `E-mail Address`
  - Unicode column names
- Define and document a deterministic canonicalization algorithm.
- Detect canonical-name collisions during import.
- Do not silently choose one colliding field.
- Expose merge-field metadata to the frontend.
- Ensure inserted field syntax can be reliably parsed.
- Preserve case-insensitive lookup where appropriate.
- Replace the current `\{(\w+)\}` parsing approach.
- Ensure preview, test send, and bulk send use the same renderer.

Choose one supported syntax and document it. A suitable option is:

```text
{{field_id}}
```

Support migration from the existing `{FirstName}` syntax if practical. At minimum, preserve the standard legacy fields:

- `{FirstName}`
- `{LastName}`
- `{Email}`

### Unknown and Missing Fields

Unknown fields must never silently become empty strings during a send.

Implement validation statuses such as:

- Valid
- Unknown field
- Known field with missing value
- Canonical collision
- Invalid syntax
- Unsupported transform

Default behavior:

- Block campaign sending when the template references an unknown field.
- Warn when a known field is blank for one or more recipients.
- Allow the user to define a fallback value before sending.

Add fallback syntax or structured fallback behavior, for example:

```text
{{first_name|there}}
```

Do not add an overly complex template language. Keep the first implementation deterministic and secure.

---

## 2. Unified Template Rendering

Create one rendering path for:

- Subject
- HTML body
- Plain-text body
- To
- CC
- BCC
- Personalized attachment paths
- Preview
- Test send
- Bulk send
- Retry

Do not maintain separate frontend-only replacement logic.

Remove or replace all ad hoc frontend preview substitutions such as direct `.replace()` calls for `FirstName`, `LastName`, and `Email`.

The renderer must return structured diagnostics.

Example:

```go
type RenderDiagnostic struct {
    Severity string
    Code     string
    FieldID  string
    Message  string
}

type RenderedMessage struct {
    To          []string
    CC          []string
    BCC         []string
    Subject     string
    HTMLBody    string
    TextBody    string
    Attachments []ResolvedAttachment
    Diagnostics []RenderDiagnostic
}
```

---

## 3. Representative Test Sends

### Current Problem

Test sends only use limited sample fields and do not include the selected contact’s complete custom data.

### Required Changes

- Replace sample first-name/last-name arguments with a complete `Contact`.
- Let the user select which imported contact is used for a test send.
- Render all custom fields exactly as bulk send would.
- Clearly show:
  - Actual test recipient
  - Contact whose merge data is being used
  - Rendered To/CC/BCC
  - Rendered subject/body
  - Resolved attachments
- The test recipient address must not overwrite the contact’s `{Email}` merge value unless the user explicitly selects that behavior.
- Test sends must pass through the same campaign preflight and rendering pipeline.

Acceptance criterion:

> Given the same contact and template, preview, test send, and bulk send must generate equivalent rendered subject, body, CC, BCC, and attachments.

---

## 4. Campaign Preflight

Implement a first-class preflight operation that runs before any test or bulk send.

The preflight must detect:

- Unknown merge fields
- Invalid merge syntax
- Missing required values
- Blank rendered subjects
- Blank rendered bodies
- Invalid To addresses
- Invalid CC/BCC addresses
- Duplicate recipients
- Recipients appearing in both To and CC/BCC where inappropriate
- Missing attachments
- Attachment paths resolving to directories
- Oversized attachments
- Duplicate attachments
- Unsupported sender state
- Outlook unavailable
- Zero selected recipients
- Suppressed recipients, when suppression support is added
- Estimated campaign duration
- Estimated number of send attempts
- Whether the selected sender supports the requested features

Return structured results.

```go
type PreflightIssue struct {
    Severity     string
    Code         string
    Message      string
    ContactIndex *int
    Email        string
    FieldID      string
}

type PreflightResult struct {
    CanSend           bool
    Errors            []PreflightIssue
    Warnings          []PreflightIssue
    RecipientCount    int
    EstimatedDuration time.Duration
}
```

The frontend must display preflight results before sending.

Bulk sending must be impossible while `CanSend` is false.

---

## 5. Email Validation

Replace the current `@` and `.` validation.

Use Go’s standard mail-address parsing where appropriate:

```go
net/mail
```

Requirements:

- Normalize surrounding whitespace.
- Support semicolon- and comma-delimited Outlook address lists.
- Reject malformed addresses before send.
- Preserve display names when valid.
- Normalize email addresses for duplicate detection.
- Use a documented duplicate-normalization policy.
- Do not over-normalize provider-specific addresses by removing dots or plus tags.

---

## 6. Duplicate Recipient Resolution

The current behavior warns about duplicates but selects all rows by default.

Implement an explicit duplicate policy:

- Keep first
- Keep last
- Exclude all duplicates
- Keep all
- Manual review

Default to `Keep first`.

Show duplicate groups in the UI and allow the user to inspect conflicting field values.

Apply the policy before preflight recipient counts are calculated.

---

## 7. Attachment Remediation

### Existing Gaps

The attachment component exposes drag/drop and file-info properties, but the parent does not wire them. File sizes and total-size warnings therefore do not work reliably.

### Required Changes

- Implement supported Wails-native file-drop handling for Windows.
- Retrieve file metadata through the backend.
- Display:
  - File name
  - Full path where appropriate
  - Individual size
  - Total size
  - Missing/unreadable status
- Deduplicate attachment paths.
- Prevent directory attachments.
- Validate attachments before entering sending state.
- Use a configurable warning limit.
- Distinguish warning limits from hard sender-specific limits.
- Ensure clicking a remove button does not trigger the add-file dialog through event bubbling.
- Add support for personalized attachment templates, but only after canonical merge rendering is complete.

Personalized attachment example:

```text
C:\Reports\{{customer_id}}.pdf
```

Preflight must resolve and validate every personalized attachment before sending unless the user explicitly selects a documented lazy-resolution mode.

---

## 8. Sending Settings Must Affect Behavior

Create a validated `SendOptions` model.

```go
type SendOptions struct {
    DelayBetweenMessages time.Duration
    BatchSize            int
    PauseBetweenBatches  time.Duration
    ConfirmBeforeSend    bool
    ContinueOnError      bool
}
```

Requirements:

- Replace the hard-coded 500 ms delay.
- Read the value from validated application settings.
- Enforce supported min/max limits.
- Use the same options for sends and retries unless explicitly changed.
- Make `ConfirmSend` affect actual UI behavior.
- Implement completion sound only when enabled.
- Either implement auto-save behavior or remove the setting until implemented.
- Every visible setting must have a test proving it changes behavior.

---

## 9. Typed Campaign Results and Fatal Errors

Do not return an empty successful-looking result after a fatal preflight or Outlook failure.

Implement explicit campaign states.

```go
type CampaignState string

const (
    CampaignCompleted       CampaignState = "completed"
    CampaignCancelled       CampaignState = "cancelled"
    CampaignPreflightFailed CampaignState = "preflight_failed"
    CampaignRuntimeFailed   CampaignState = "runtime_failed"
)

type RecipientStatus string

const (
    RecipientSubmitted RecipientStatus = "submitted"
    RecipientFailed    RecipientStatus = "failed"
    RecipientSkipped   RecipientStatus = "skipped"
    RecipientCancelled RecipientStatus = "cancelled"
)

type CampaignResult struct {
    State            CampaignState
    Attempted        int
    Submitted        int
    Failed           int
    Skipped          int
    Cancelled        int
    FatalError       string
    RecipientResults []RecipientResult
}
```

Requirements:

- Preflight failures return `preflight_failed`.
- Outlook creation failure returns `runtime_failed`.
- Missing global attachment returns `preflight_failed`.
- Individual send failures create recipient results.
- Retry results must not inflate totals incorrectly.
- Preserve attempt history rather than flattening it into ambiguous counters.
- Display submitted versus delivered accurately. Do not claim confirmed delivery.

---

# Milestone 2: Testable Sending Architecture

## 10. EmailSender Abstraction

Introduce a sender interface and isolate Outlook COM code.

Required implementations:

1. `ClassicOutlookCOMSender`
2. `FakeSender`
3. `DraftOnlySender`, if feasible

The fake sender must support deterministic behaviors:

- Always succeeds
- Always fails
- Fails for selected recipients
- Delayed responses
- Cancellation
- Simulated Outlook-unavailable state

This sender must enable full campaign-runner tests without Outlook installed.

Do not let the campaign runner import `go-ole`.

---

## 11. Dedicated Outlook STA Worker

COM operations must execute on one dedicated, locked OS thread.

Implement an Outlook worker with:

- `runtime.LockOSThread()`
- Explicit COM initialization
- A request channel
- A response channel or future/result mechanism
- Controlled startup and shutdown
- Context cancellation
- Serialized COM access
- Clean release of COM objects
- Accurate error propagation

Do not initialize and tear down COM independently in multiple unrelated methods unless the design proves it is safe.

Handle COM initialization results correctly:

- `S_OK`
- `S_FALSE`
- `RPC_E_CHANGED_MODE`
- Unexpected error types

Only call `CoUninitialize` when the current code path successfully acquired a COM initialization reference.

Never use an unsafe direct type assertion on an arbitrary `error`.

---

## 12. Outlook Capability Detection

Distinguish:

- Classic Outlook available
- New Outlook detected
- Outlook unavailable
- Outlook COM inaccessible
- Outlook running but blocked by policy
- Account/profile unavailable

Return actionable diagnostics rather than a generic “Outlook not available.”

Do not falsely claim support for new Outlook if COM automation is unavailable.

Expose sender capabilities:

```go
type SenderCapabilities struct {
    SupportsHTML             bool
    SupportsAttachments      bool
    SupportsMultipleAccounts bool
    SupportsSharedMailbox    bool
    SupportsDraftOnly        bool
    SupportsScheduling       bool
}
```

---

## 13. Cancellation, Pause, and Resume

At minimum implement:

- Cancel current campaign
- Stop before the next recipient
- Mark unattempted recipients as cancelled
- Preserve completed recipient results
- Do not retry already successful recipients

Optional if architecture permits:

- Pause/resume during the current session
- Batch sending
- Pause between batches

Use `context.Context`; do not create global cancellation flags without synchronization.

---

## 14. Retry Semantics

Retries must operate on recipient attempt records.

Requirements:

- Retry only failed recipients by default.
- Allow selected failed recipients.
- Preserve original attempt.
- Add a new attempt record.
- Do not double-count prior successes.
- Retain the exact rendered message or source snapshot needed to explain what was retried.
- Clearly show current failures versus historical failures.

---

# Milestone 3: Persistence, Security, and Data Integrity

## 15. Safe Template Storage

Harden template persistence.

Requirements:

- Template IDs generated only by the backend.
- Reject path separators and invalid IDs.
- Prevent built-in templates from being overwritten through a crafted ID.
- Validate template name, subject, body, and format.
- Write to a temporary file.
- Flush and atomically rename.
- Update in-memory state only after durable write succeeds.
- Recover gracefully from one corrupted template file.
- Log corruption without exposing sensitive template contents.
- Add import/export support only after validation exists.

Consider moving templates and settings to SQLite under the persistence milestone.

---

## 16. Safe Settings Persistence

- Validate all settings before replacing the current settings object.
- Merge stored settings with defaults to support schema evolution.
- Add a settings schema version.
- Handle corrupted settings files.
- Use atomic writes.
- Return copies of slices and maps, not shared mutable backing storage.
- Ensure `GetRecentFiles` does not expose mutable internal state.
- Remove recent-file entries that no longer exist, or mark them unavailable.
- Normalize paths for duplicate comparison.

---

## 17. HTML Security

The template HTML and imported merge values require separate trust handling.

Requirements:

- Escape merge values inserted into HTML by default.
- Do not allow imported contact data to inject arbitrary HTML.
- Sanitize HTML used in UI previews.
- Use a well-maintained sanitizer rather than regex-only sanitization.
- Preserve a safe subset needed by the rich text editor.
- Validate and sanitize links.
- Do not permit scripts, event handlers, embedded objects, unsafe URLs, or uncontrolled iframes.
- Keep preview sandboxing restrictive.
- Ensure plain-text messages never pass through HTML rendering.
- Add tests for script tags, event attributes, malformed markup, encoded payloads, and malicious merge data.

If trusted raw-HTML fields are ever supported, make them explicit, disabled by default, and clearly labeled.

---

## 18. Campaign Persistence

Introduce SQLite if feasible within this implementation.

Suggested data model:

- Schema migrations
- Campaigns
- Campaign recipient snapshots
- Message templates
- Recipient attempts
- Attachments
- Suppression entries
- Application metadata

Persist enough information to:

- Review a past campaign
- Resume an interrupted campaign safely
- Retry only failed recipients
- Export campaign results
- Explain what was rendered and attempted

Do not store Outlook credentials or access tokens in plaintext.

If SQLite is deferred, implement repository interfaces so JSON persistence can later be replaced without rewriting the application service layer.

---

## 19. Suppression and Unsubscribe Management

Implement a local suppression list before open tracking.

Requirements:

- Manually add/remove addresses.
- Import/export suppression list.
- Normalize addresses consistently.
- Block suppressed recipients during preflight.
- Show skipped recipient counts and reason.
- Require an explicit override workflow if overrides are allowed.
- Record override decisions in campaign history.
- Do not automatically add tracking pixels or unsubscribe links without explicit configuration.

---

# Milestone 4: Import and User Experience

## 20. Contact Import Improvements

Implement:

- CSV UTF-8 BOM handling
- Configurable or detected delimiter
- Quoted-field support through the standard CSV parser
- Clear row-level errors
- Excel sheet selection
- Column mapping
- Saved import profiles
- Empty-row handling
- Duplicate header detection
- Canonical-field collision detection
- Large-file safeguards
- Optional preview of the first rows before final import

Do not load an unbounded workbook into memory without limits.

Add configurable limits for:

- File size
- Row count
- Column count
- Cell size

Surface failures clearly.

---

## 21. Contact Editing and Review

Allow users to:

- Review imported values
- Filter by any field
- Edit a row before sending
- Exclude a row
- Resolve duplicate groups
- Inspect fields missing for selected recipients
- Export the cleaned recipient list

Replace index-only selection identity with stable contact IDs so filtering, editing, sorting, and deduplication do not corrupt selection state.

---

## 22. Frontend State Refactor

`App.tsx` currently holds too much state and workflow logic.

Refactor into focused hooks or stores, such as:

- `useContactImport`
- `useCampaignDraft`
- `useCampaignPreflight`
- `useCampaignRunner`
- `useTemplates`
- `useSettings`
- `useOutlookStatus`

Requirements:

- Avoid stale closure bugs in settings updates.
- Use functional updates where required.
- Remove duplicate theme state sources if possible.
- Use one authoritative settings source.
- Keep Wails calls behind typed service wrappers.
- Avoid `prompt()` and generic `window.confirm()` for critical workflows.
- Use accessible application modals.
- Add error boundaries or equivalent failure handling.

---

## 23. Accessibility

Implement:

- Proper labels for all controls
- Keyboard-operable merge-field buttons
- Focus trapping in modals
- Focus restoration after modal close
- Escape-to-close where safe
- Screen-reader announcements for send progress
- Semantic progress elements
- High-contrast compatibility
- No critical information conveyed by color alone
- Accessible duplicate and error indicators

---

# Milestone 5: Testing and Engineering Quality

## 24. Go Tests

Add table-driven unit tests for:

### File Service

- Standard CSV
- UTF-8 BOM
- Missing headers
- Duplicate headers
- Missing email column
- Alternate email header aliases
- Invalid email addresses
- Blank rows
- Duplicate addresses
- Canonical merge-field collisions
- Excel files with multiple sheets
- Excel sheet selection
- Large and malformed inputs

### Merge Service

- Standard fields
- Custom fields
- Spaces and punctuation in source headers
- Unicode headers
- Unknown fields
- Missing values
- Fallbacks
- HTML escaping
- Plain-text rendering
- Subject rendering
- CC/BCC rendering
- Personalized attachment paths
- Collision detection
- Invalid syntax

### Settings and Templates

- Defaults
- Schema migration
- Invalid settings
- Atomic save failure
- Corrupted files
- Recent-files copying and normalization
- Invalid template IDs
- Built-in overwrite attempts
- Atomic template writes

### Campaign Runner

Using `FakeSender`, test:

- Complete success
- Partial failure
- Fatal sender failure
- Cancellation
- Retry
- Duplicate exclusion
- Suppression
- Delay application using an injectable clock
- Batch pauses
- Progress ordering
- Accurate result totals
- No double counting

Avoid real sleeping in unit tests. Inject a clock or sleeper interface.

---

## 25. Windows Outlook Integration Tests

Add opt-in integration tests using build tags.

Example:

```go
//go:build windows && outlookintegration
```

Test:

- Outlook capability detection
- COM initialization helper
- Create a draft
- Set To/CC/BCC
- Set HTML body
- Set plain-text body
- Add attachment
- Release COM objects
- Cancellation between messages

Default CI must not require Outlook.

Document how to run integration tests on a Windows workstation with classic Outlook.

Prefer draft creation over sending real email during automated integration testing.

---

## 26. Frontend Tests

Add an appropriate React testing stack compatible with the chosen frontend versions.

Test:

- Contact selection with filtering and stable IDs
- Duplicate-resolution workflow
- Merge-field insertion
- Preflight error display
- Send disabled when validation fails
- Test-send contact selection
- Attachment drag/drop
- Attachment size display
- Settings behavior
- Confirmation-setting behavior
- Cancellation
- Retry UI
- Accessible modal behavior
- Preview sanitization

---

## 27. Static Analysis and Formatting

Add and enforce:

### Go

- `gofmt`
- `go test ./...`
- `go vet ./...`
- `staticcheck` or `golangci-lint`
- Race detector where supported

### Frontend

- TypeScript strict mode where practical
- ESLint
- Prettier
- Production build
- Dependency audit with documented exceptions

Remove unused dependencies such as packages imported into `package.json` but not used in the source.

---

## 28. GitHub Actions

Create CI workflows that run on pull requests and pushes to `main`.

At minimum:

1. Go unit tests
2. Go vet/static analysis
3. Frontend install with `npm ci`
4. Frontend lint
5. Frontend tests
6. TypeScript build
7. Wails Windows build
8. Artifact upload for the unsigned test build
9. Dependency vulnerability scanning
10. License or SBOM generation if practical

Pin action versions.

Do not publish a release from every commit.

Use a separate release workflow triggered by a version tag.

---

## 29. Reproducible Builds

- Align README Go requirements with `go.mod`.
- Decide the supported Go version and pin it in CI.
- Pin the Wails CLI version used by builds.
- Use `npm ci`, not `npm install`, in reproducible build paths.
- Verify `package-lock.json` is current.
- Include version information through build-time variables.
- Remove the hard-coded `1.3.0` and `Your Name`.
- Derive app version from a single source.
- Embed commit SHA and build date where useful.
- Keep debug and production builds distinct.

---

## 30. Packaging and Release Quality

Implement or document:

- Windows installer build
- Application metadata
- Upgrade behavior
- User-data preservation
- Uninstall behavior
- Code-signing hooks
- Release checksums
- SBOM
- Release notes
- Compatibility statement:
  - Windows versions
  - Classic Outlook versions
  - New Outlook limitations
- Rollback guidance

Do not claim the application is production-ready until automated release gates pass.

---

# Milestone 6: Microsoft Graph Readiness

Do not make Graph mandatory for the initial remediation, but prepare the architecture.

## 31. Graph Sender Design

Create a design document and interfaces for a future `MicrosoftGraphSender`.

Address:

- Delegated authentication
- Token storage
- `Mail.Send`
- Draft creation
- Attachments
- Shared mailboxes
- Send As / Send on Behalf Of
- Submission versus delivery semantics
- Rate limits and retry-after handling
- Tenant administrator restrictions
- Personal Microsoft accounts versus organizational accounts

Keep Graph-specific code out of the core campaign runner.

If implementing Graph during this work, place it behind a feature flag and include authentication, security, and integration tests.

---

# Documentation Remediation

Update all repository documentation to reflect actual behavior.

## README

Correct:

- Production-readiness claims
- Supported Go version
- Supported Outlook type
- New Outlook limitations
- Implemented versus planned features
- Build steps
- Test steps
- CI status
- Data storage locations
- Privacy and compliance considerations
- Troubleshooting
- Security limitations
- Release installation instructions

## ROADMAP

The prior roadmap was moved to `dev_docs/archive/ROADMAP.md` during the repo reorganization and is no longer surfaced at the repo root. Revise it and restore a current, authoritative roadmap (either at the root as `ROADMAP.md` or under `dev_docs/`, whichever the project prefers). When revising:

- Remove stale dates.
- Correct completed/incomplete statuses.
- Separate verified functionality from planned functionality.
- Prioritize correctness, testing, and Graph migration over tracking pixels.
- Add explicit release criteria for each milestone.

## Additional Documentation

Place technical/developer documentation under the existing `dev_docs/` directory (the repo already uses it for internal docs). Keep the community-health files at the repo root, where GitHub surfaces them specially.

Create:

- `dev_docs/architecture.md`
- `dev_docs/merge-fields.md`
- `dev_docs/outlook-compatibility.md`
- `dev_docs/testing.md`
- `dev_docs/security.md`
- `dev_docs/campaign-lifecycle.md`
- `CONTRIBUTING.md` (repo root)
- `SECURITY.md` (repo root)
- `CHANGELOG.md` (repo root)

---

# Suggested Delivery Sequence

Use the following order unless repository findings justify a documented adjustment.

## Phase A: Baseline and Safety

1. Baseline build and test report
2. Canonical merge fields
3. Unified rendering
4. Template validation
5. Representative test sends
6. Typed preflight
7. Typed campaign results
8. Duplicate policy
9. Attachment wiring
10. Settings behavior fixes

## Phase B: Architecture

1. `EmailSender` interface
2. Fake sender
3. Campaign runner
4. Dedicated Outlook STA worker
5. Cancellation
6. Retry attempt history

## Phase C: Persistence and Security

1. Atomic settings/templates
2. HTML escaping and sanitization
3. Campaign storage
4. Suppression list

## Phase D: Import and UI

1. Stable contact IDs
2. Column mapping
3. Sheet selection
4. Contact review/editing
5. Frontend state refactor
6. Accessibility

## Phase E: Quality and Release

1. Unit tests
2. Frontend tests
3. Integration-test harness
4. Linting
5. CI
6. Reproducible build
7. Documentation
8. Packaging and release gates

---

# Required Acceptance Criteria

The implementation is not complete until all applicable criteria pass.

## Rendering

- A column called `Account Manager` can be inserted and rendered.
- A column called `Customer-ID` can be inserted and rendered.
- Unicode column names can be represented without collision.
- Unknown fields block the send.
- Missing values are reported before sending.
- Preview, test send, and bulk send use the same renderer.

## Sending

- Configured delay is honored.
- Campaign can be cancelled.
- Fatal Outlook errors are not shown as zero-failure success.
- Retries do not double-count successful recipients.
- Duplicate default behavior sends only once per normalized address.
- Preflight blocks invalid To/CC/BCC addresses.
- All attachment failures are visible before send when resolvable.

## Outlook

- COM is isolated behind an interface.
- COM work executes on a dedicated STA thread.
- COM initialization errors cannot panic through unsafe assertions.
- COM object release paths are tested or documented.
- New Outlook limitations are clearly communicated.

## Security

- Merge values are escaped in HTML.
- Preview HTML is sanitized.
- Template IDs cannot traverse directories.
- Writes are atomic.
- Sensitive data is not logged.
- Suppressed recipients cannot be sent accidentally.

## Engineering

- Go tests pass.
- Frontend tests pass.
- Linting passes.
- Type checking passes.
- Windows Wails build passes.
- CI is active.
- README matches actual requirements and capabilities.
- No unimplemented feature is described as complete.

---

# Required Final Deliverables

At the end of the work, provide:

1. A concise implementation summary.
2. A list of changed files grouped by milestone.
3. A list of architectural decisions and tradeoffs.
4. Test commands run and their results.
5. Build commands run and their results.
6. Any Outlook integration steps not testable in the current environment.
7. Remaining known issues.
8. Security considerations.
9. Migration notes for user settings/templates.
10. Recommended follow-on work, especially Microsoft Graph support.

Also update the repository with appropriate issues or TODO references for any explicitly deferred item. Do not leave silent partial implementations.

---

# Definition of Done

The project is considered remediated only when:

- Core campaigns are deterministic and preflighted.
- Merge fields cannot silently disappear.
- Test sends accurately represent real sends.
- All visible sending settings affect actual behavior.
- Outlook COM is isolated and safely threaded.
- Campaign failures and retries are accurately represented.
- Automated tests cover the campaign engine without Outlook.
- CI validates every pull request.
- Security-sensitive rendering and persistence paths are hardened.
- Documentation accurately represents the application.

Begin by inspecting the current repository and producing a short baseline findings note in the pull request description. Then implement the work in the ordered phases above.
