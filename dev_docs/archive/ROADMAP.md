# Archived MailMerge Go Product Roadmap

> Historical planning document. It includes proposals that are not implemented
> and a superseded cross-platform Outlook vision. See the root
> [ROADMAP](../../ROADMAP.md) for the current authoritative status.

> **Vision**: A cross-platform, feature-rich mail merge application that empowers users to send personalized bulk emails through Microsoft Outlook on both Windows and macOS.

---

## Table of Contents

1. [Current State Assessment](#current-state-assessment)
2. [Phase 1: Foundation & Stability](#phase-1-foundation--stability-v12) ✅ **COMPLETED**
3. [Phase 2: Enhanced User Experience](#phase-2-enhanced-user-experience-v13) ✅ **COMPLETED**
4. [Phase 3: Advanced Features](#phase-3-advanced-features-v14)
5. [Phase 4: macOS Support](#phase-4-macos-support-v20)
6. [Phase 5: Enterprise & Collaboration](#phase-5-enterprise--collaboration-v21)
7. [Technical Debt & Maintenance](#technical-debt--maintenance)
8. [Implementation Timeline](#implementation-timeline)

---

## Current State Assessment

### Existing Capabilities ✅
- CSV and Excel (.xlsx) contact import
- **Custom merge fields from any column** (Phase 1.1)
- **Contact filtering with search** (Phase 1.1)
- **Contact selection with checkboxes** (Phase 1.1)
- **Duplicate email detection** (Phase 1.1)
- Rich text (HTML) and plain text email composition
- **Email preview modal** (Phase 1.2)
- File attachments support
- **Drag-and-drop attachments** (Phase 1.3)
- **Attachment size display & warnings** (Phase 1.3)
- Real-time progress tracking during sends
- **Retry failed emails** (Phase 1.4)
- Export logs to CSV
- Dark/light theme support
- Outlook COM automation (Windows-only)
- **Email templates with save/load** (Phase 2.1)
- **6 built-in starter templates** (Phase 2.1)
- **Settings panel with persistence** (Phase 2.3)
- **Keyboard shortcuts** (Phase 2.3)
- **Recent files list** (Phase 2.4)
- **Rate limiting configuration** (Phase 2.2)

### Current Limitations ⚠️
- Windows-only (COM automation not available on macOS)
- No scheduling capability
- No batch sending with pauses
- No unsubscribe list management
- No column mapping for non-standard files

---

## Phase 1: Foundation & Stability (v1.2) ✅ COMPLETED

**Goal**: Solidify the foundation, improve reliability, and enhance core functionality.

**Status**: ✅ All Phase 1 features have been implemented.

### 1.1 Contact Management Improvements ✅

#### Custom Merge Fields ✅ IMPLEMENTED
Support any column from the contact file as a merge field.

**Implementation Details**:
- Modified `FileService.ParseContactFile` to extract all column headers
- Updated `Contact` model with `CustomFields map[string]string`
- Updated `MergeService.renderTemplate` to check CustomFields for any field
- Added `GetMergeFieldsFromHeaders` to generate merge field buttons dynamically
- Auto-generate merge field buttons based on detected columns

```go
// Updated Contact model with custom fields
type Contact struct {
    FirstName    string            `json:"firstName"`
    LastName     string            `json:"lastName"`
    Email        string            `json:"email"`
    CustomFields map[string]string `json:"customFields"` // All other columns
}
```

#### Contact Filtering & Selection ✅ IMPLEMENTED
Allow users to select specific contacts for sending.

**Implementation Details**:
- Added checkbox column to ContactTable component
- Added "Select All" / "Deselect All" buttons
- Added search input to filter by name or email
- Contacts filtered in real-time using `useMemo`
- Selection state stored as `Set<number>` in App component

**Components Updated**:
- `ContactTable.tsx` - Complete rewrite with selection and filtering
- `App.tsx` - Added `selectedIndices` state and handlers

#### Duplicate Detection ✅ IMPLEMENTED
Warn users about duplicate email addresses.

**Implementation Details**:
- Added `detectDuplicates()` function in `FileService`
- Returns duplicate emails in `ParseResult.Duplicates`
- Visual warning badge on contacts with duplicate emails
- Alert shown at top of ContactTable when duplicates exist

### 1.2 Email Composition Enhancements ✅

#### Email Preview Modal ✅ IMPLEMENTED
Show exactly how the email will look for a specific contact.

**Implementation Details**:
- Added `PreviewModal` component with contact selector
- Preview calls backend `PreviewMergeForContact` to render templates
- Shows rendered subject and body with actual contact values
- HTML emails displayed in sandboxed iframe
- Attachments list shown with file names

**Components Added**:
- `PreviewModal.tsx` - New component for email preview
- `app.go` - Added `PreviewMergeForContact` method

#### Subject Line Merge Fields ✅ VERIFIED
Merge fields work correctly in subject lines.

**Status**: Already implemented, verified working with custom fields

### 1.3 Attachment Improvements ✅

#### Attachment Validation ✅ IMPLEMENTED
Validate attachments before sending.

**Implementation Details**:
- File size displayed for each attachment
- Total attachment size calculated and displayed
- Warning shown when total exceeds 20MB
- Uses `formatFileSize` helper for human-readable sizes

#### Drag-and-Drop Attachments ✅ IMPLEMENTED
Enable dropping files directly onto the attachment area.

**Implementation Details**:
- Added HTML5 drag-and-drop event handlers
- Visual feedback during drag (border color change, background highlight)
- Supports dropping multiple files simultaneously
- `onFilesDropped` callback prop for handling dropped files

**Components Updated**:
- `AttachmentManager.tsx` - Added drag-drop functionality

### 1.4 Error Handling & Resilience ✅

#### Retry Failed Emails ✅ IMPLEMENTED
Option to retry failed sends.

**Implementation Details**:
- Failed contacts stored in `SendResult.FailedContacts`
- "Retry Failed" button appears in ResultsSummary when failures exist
- Retry merges results with previous totals
- Shows combined logs from all attempts

**Components Updated**:
- `ResultsSummary.tsx` - Added retry button and handler
- `App.tsx` - Added `handleRetryFailed` callback
- `backend/models/contact.go` - Added `FailedContacts` to SendResult

#### Connection Recovery
Handle Outlook connection issues gracefully.

**Status**: ⏳ Deferred to Phase 2 - Requires more complex state management

---

## Phase 2: Enhanced User Experience (v1.3) ✅ COMPLETED

**Goal**: Improve usability, add productivity features, and polish the interface.

**Status**: ✅ All Phase 2 features have been implemented.

### 2.1 Email Templates ✅

#### Save & Load Templates ✅ IMPLEMENTED
Persist email templates for reuse.

**Implementation Details**:
- Created `TemplateService` in `backend/services/template_service.go`
- Templates stored as JSON files in `%AppData%/MailMergeGo/templates/`
- Each template has: id, name, subject, body, isHTML, isBuiltIn, createdAt, updatedAt
- CRUD operations: GetAllTemplates, GetTemplate, SaveTemplate, DeleteTemplate

**Data Structure**:
```go
type EmailTemplate struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Subject   string    `json:"subject"`
    Body      string    `json:"body"`
    IsHTML    bool      `json:"isHTML"`
    IsBuiltIn bool      `json:"isBuiltIn"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

**UI Changes**:
- `TemplateManager` component with dropdown selector
- Save button opens modal with name input and preview
- Delete button with confirmation (double-click within 3 seconds)
- User templates separated from built-in templates

#### Built-in Starter Templates ✅ IMPLEMENTED
Provide common email templates out of the box.

**Included Templates**:
- Welcome Email - Professional greeting template
- Newsletter Template - News and updates format
- Meeting Request - Meeting invitation with details
- Follow-up Email - After initial contact
- Thank You Note - Appreciation message
- Plain Text Template - Simple text-only format

### 2.2 Advanced Sending Options (Partial)

#### Scheduling
⏳ Deferred to Phase 3 - Requires background service architecture

#### Rate Limiting Configuration ✅ IMPLEMENTED
Allow users to adjust sending speed.

**Implementation Details**:
- Added `sendingDelay` setting (100-5000ms range)
- Settings panel with preset buttons: Fast (100ms), Normal (500ms), Slow (2s)
- Custom delay input field for specific values
- Persisted in `settings.json`

#### Send in Batches
⏳ Deferred to Phase 3 - Requires additional state management

### 2.3 UI/UX Improvements ✅

#### Improved Progress Display
✅ Already implemented in Phase 1 with real-time progress tracking

#### Keyboard Shortcuts ✅ IMPLEMENTED
Add keyboard shortcuts for power users.

**Implemented Shortcuts**:
- `Ctrl+O` - Open contact file
- `Ctrl+S` - Save as template (triggers save modal)
- `Ctrl+Enter` - Send test email
- `Ctrl+Shift+Enter` - Send to all selected contacts
- `Ctrl+,` - Open settings modal
- `Ctrl+D` - Toggle dark mode
- `Escape` - Close modals

**Implementation Details**:
- Global keyboard event listener in App component
- Ignores shortcuts when typing in input/textarea fields
- Shortcuts displayed in Settings modal and tooltips

#### Settings Panel ✅ IMPLEMENTED
Centralized settings management.

**Settings Included**:
- Theme preference (Light/Dark/System) with visual toggle buttons
- Default email format (HTML/Plain text) dropdown
- Sending delay configuration with presets and custom input
- Confirm before sending toggle
- Sound notifications toggle
- Auto-save templates toggle
- Recent files list with clear button

**Implementation Details**:
- Created `SettingsService` in `backend/services/settings_service.go`
- Settings persisted to `%AppData%/MailMergeGo/settings.json`
- `SettingsModal` component with tabbed interface (General, Sending, Files, Shortcuts)
- Settings button added to header next to theme toggle

### 2.4 Contact File Improvements ✅

#### Remember Recent Files ✅ IMPLEMENTED
Quick access to recently used contact files.

**Implementation Details**:
- Recent files stored in settings (up to 10 files)
- Recent Files tab in Settings modal
- Click to open recent file directly
- "Clear Recent Files" button
- `AddRecentFile`, `GetRecentFiles`, `ClearRecentFiles` methods

#### Column Mapping
⏳ Deferred to Phase 3 - Current implementation auto-detects standard column names

---

## Phase 3: Advanced Features (v1.4) 🚧 IN PROGRESS

**Goal**: Add power-user features and automation capabilities.

**Status**: 🚧 Partially complete - CC/BCC Support (3.4) implemented.

### 3.1 Email Tracking & Analytics

#### Basic Open Tracking
Track when emails are opened (HTML emails only).

**Implementation**:
- Option to insert tracking pixel
- Local server endpoint to receive tracking pings
- Note: Requires app running or external service

**Privacy Considerations**:
- Make opt-in only
- Clear disclosure to users
- Consider privacy regulations (GDPR, CAN-SPAM)

### 3.2 Unsubscribe Management

#### Unsubscribe List
Maintain list of emails that should not receive emails.

**Implementation**:
- Persistent unsubscribe list (JSON/SQLite)
- Auto-check against list before sending
- Skip unsubscribed contacts with warning
- Add manual entry option
- Import/export unsubscribe list

### 3.3 Personalized Attachments

#### Per-Contact Attachments
Send different attachments based on contact data.

**Implementation**:
- Support file path merge fields in attachment list
- Example: `C:\Reports\{Email}_report.pdf`
- Validate all files exist before send
- Error handling for missing files

### 3.4 CC/BCC Support ✅ IMPLEMENTED

#### Carbon Copy Options ✅
Add CC and BCC recipients with optional merge field support.

**Status**: ✅ Completed - Full CC/BCC support with merge field templates.

**Implementation Details**:

**Backend Changes**:
- Added `CC`, `BCC`, `CCTemplate`, `BCCTemplate` fields to `EmailRequest` model
- Added `CC`, `BCC`, `CCTemplate`, `BCCTemplate` fields to `TestEmailRequest` model
- Updated `OutlookService.SendTestEmail` to accept and set CC/BCC via COM
- Updated `OutlookService.SendBulkEmails` to accept CC/BCC static values and templates
- Updated `OutlookService.sendSingleEmail` to set CC/BCC properties on Outlook mail item
- CC/BCC templates are rendered per-contact using `MergeService` for dynamic recipients

```go
// CC/BCC COM integration in sendSingleEmail
if cc != "" {
    _, err = oleutil.PutProperty(mailItem, "CC", cc)
}
if bcc != "" {
    _, err = oleutil.PutProperty(mailItem, "BCC", bcc)
}
```

**Frontend Changes**:
- Added collapsible CC/BCC section to `EmailEditor` component
- CC/BCC fields support merge field insertion at cursor position
- Auto-expand CC/BCC section when values are present
- Preview tab shows rendered CC/BCC with sample data
- `App.tsx` manages CC/BCC state and passes to backend requests
- Smart detection: fields with `{` are sent as templates for per-contact rendering

**User Features**:
- ✅ Static CC/BCC (same recipients for all emails)
- ✅ Dynamic CC/BCC using merge fields (e.g., `{ManagerEmail}`)
- ✅ Multiple recipients (comma-separated)
- ✅ Works with both test emails and bulk sends
- ✅ Preview shows CC/BCC with sample data rendered
- ✅ Clean collapsible UI to reduce clutter

### 3.5 Auto-Save & Recovery

#### Draft Auto-Save
Automatically save work in progress.

**Implementation**:
- Auto-save every 30 seconds when changes detected
- Save to local draft file
- Prompt to restore on startup if draft exists
- Clear draft after successful send

---

## Phase 4: macOS Support (v2.0)

**Goal**: Bring full MailMerge Go functionality to macOS users with Microsoft Outlook.

### 4.1 Architecture Changes

#### Platform Abstraction Layer
Create abstraction for email sending.

**Implementation**:
```go
// EmailSender interface for platform-specific implementations
type EmailSender interface {
    CheckAvailable() error
    SendEmail(request EmailRequest) error
    SendBulk(requests []EmailRequest, progress chan<- ProgressUpdate) SendResult
}

// Platform implementations
type WindowsOutlookSender struct { /* COM automation */ }
type MacOSOutlookSender struct { /* AppleScript/osascript */ }
```

**Build Tags**:
```go
// outlook_windows.go
//go:build windows

// outlook_darwin.go  
//go:build darwin
```

### 4.2 macOS Outlook Integration

#### AppleScript Automation
Use AppleScript to automate Outlook on macOS.

**Implementation**:
```applescript
tell application "Microsoft Outlook"
    set newMessage to make new outgoing message with properties {subject:"Hello", content:"Body text"}
    make new recipient at newMessage with properties {email address:{address:"user@example.com"}}
    send newMessage
end tell
```

**Go Integration**:
```go
func (s *MacOSOutlookSender) SendEmail(req EmailRequest) error {
    script := fmt.Sprintf(`
        tell application "Microsoft Outlook"
            set newMessage to make new outgoing message with properties {subject:"%s", content:"%s"}
            make new recipient at newMessage with properties {email address:{address:"%s"}}
            send newMessage
        end tell
    `, req.Subject, req.Body, req.Email)
    
    cmd := exec.Command("osascript", "-e", script)
    return cmd.Run()
}
```

**Challenges & Solutions**:

| Challenge | Solution |
|-----------|----------|
| AppleScript escaping | Use `osascript -` stdin method with proper escaping |
| HTML emails | Use `content type` property in AppleScript |
| Attachments | Use `make new attachment` in AppleScript |
| Rate limiting | Same delay approach as Windows |
| Error handling | Parse osascript stderr for error messages |

### 4.3 macOS-Specific UI

#### Native macOS Appearance
Match macOS design conventions.

**Implementation**:
- Wails supports native macOS window controls
- Use SF Symbols or similar icon set
- macOS-style menu bar
- Support for macOS accent colors

### 4.4 macOS Build Configuration

#### Build Scripts
Create macOS build process.

**Files**:
```
build/
├── darwin/
│   ├── Info.plist          # Already exists
│   └── entitlements.plist  # New: for code signing
└── Makefile                # Cross-platform build
```

**Build Commands**:
```bash
# macOS build
wails build -platform darwin/universal

# Code signing (for distribution)
codesign --sign "Developer ID" --entitlements entitlements.plist MailMergeApp.app
```

### 4.5 Cross-Platform Testing

#### Test Matrix
Ensure consistent behavior across platforms.

| Feature | Windows Test | macOS Test |
|---------|--------------|------------|
| Outlook detection | COM check | AppleScript check |
| Send single email | ✓ | ✓ |
| Send bulk | ✓ | ✓ |
| HTML formatting | ✓ | ✓ |
| Attachments | ✓ | ✓ |
| Progress events | ✓ | ✓ |

---

## Phase 5: Enterprise & Collaboration (v2.1)

**Goal**: Add features for team use and larger organizations.

### 5.1 Contact Lists Management

#### Named Contact Lists
Save and manage multiple contact lists.

**Implementation**:
- Import contacts to named list
- Switch between lists
- Merge/combine lists
- Export lists

### 5.2 Email Campaign History

#### Send History
Track past email campaigns.

**Implementation**:
- Store campaign metadata in SQLite
- View past campaigns: date, subject, recipient count
- Resend campaigns
- Export campaign reports

### 5.3 Outlook Profile Selection

#### Multiple Outlook Profiles
Support users with multiple Outlook profiles.

**Implementation**:
- Detect available Outlook profiles
- Profile selector in settings
- Remember selected profile

### 5.4 Integration Options

#### API for Automation
Enable programmatic access.

**Potential Integrations**:
- Command-line interface (CLI mode)
- HTTP API for local automation
- Integration with CRM systems

---

## Technical Debt & Maintenance

### Code Quality
- [ ] Add comprehensive unit tests (target: 80% coverage)
- [ ] Add integration tests for Outlook operations
- [ ] Set up CI/CD pipeline with GitHub Actions
- [ ] Add linting (golangci-lint, ESLint)
- [ ] Document all exported functions

### Performance
- [ ] Profile and optimize large contact file parsing
- [ ] Implement virtual scrolling for 10,000+ contacts
- [ ] Lazy load components for faster startup

### Security
- [ ] Validate all file paths to prevent path traversal
- [ ] Sanitize HTML content to prevent XSS
- [ ] Sign Windows executables (code signing certificate)

### Accessibility
- [ ] Add ARIA labels throughout
- [ ] Ensure keyboard navigation works everywhere
- [ ] Test with screen readers
- [ ] Add high contrast mode support

---

## Implementation Timeline

### Q1 2025 - Phase 1 (v1.2)
| Week | Focus Area |
|------|------------|
| 1-2 | Custom merge fields & dynamic field detection |
| 3-4 | Contact filtering & selection |
| 5-6 | Email preview modal |
| 7-8 | Attachment validation & drag-drop |
| 9-10 | Retry failed emails & connection recovery |
| 11-12 | Testing & documentation |

### Q2 2025 - Phase 2 (v1.3)
| Week | Focus Area |
|------|------------|
| 1-3 | Email templates (save/load) |
| 4-6 | Scheduling & rate limiting |
| 7-8 | UI improvements & keyboard shortcuts |
| 9-10 | Settings panel & recent files |
| 11-12 | Column mapping & testing |

### Q3 2025 - Phase 3 (v1.4)
| Week | Focus Area |
|------|------------|
| 1-4 | Unsubscribe management |
| 5-6 | CC/BCC support |
| 7-8 | Personalized attachments |
| 9-10 | Auto-save & recovery |
| 11-12 | Testing & documentation |

### Q4 2025 - Phase 4 (v2.0)
| Week | Focus Area |
|------|------------|
| 1-2 | Platform abstraction layer design |
| 3-5 | macOS AppleScript integration |
| 6-7 | macOS-specific UI adjustments |
| 8-9 | Cross-platform testing |
| 10-11 | macOS build & distribution |
| 12 | Release & documentation |

### Q1 2026 - Phase 5 (v2.1)
| Week | Focus Area |
|------|------------|
| 1-4 | Contact lists management |
| 5-8 | Email campaign history |
| 9-10 | Multiple Outlook profiles |
| 11-12 | Documentation & release |

---

## Feature Request Tracking

New feature requests should be added here with:
- **Title**: Brief description
- **Priority**: High / Medium / Low
- **Phase**: Which phase it fits into
- **Requested By**: Source of request
- **Status**: Proposed / Approved / In Progress / Complete

| # | Title | Priority | Phase | Status |
|---|-------|----------|-------|--------|
| 1 | Custom merge fields | High | 1 | Proposed |
| 2 | macOS support | High | 4 | Proposed |
| 3 | Email templates | Medium | 2 | Proposed |
| 4 | Scheduling | Medium | 2 | Proposed |
| 5 | Unsubscribe list | Medium | 3 | Proposed |

---

## Contributing

If you'd like to contribute to any roadmap item:
1. Check the issue tracker for existing work
2. Discuss approach in the issue before starting
3. Follow the coding standards in CONTRIBUTING.md
4. Submit PR with tests and documentation

---

## Version History

| Version | Release Date | Key Features |
|---------|--------------|--------------|
| v1.0 | 2024 | Initial release - Windows only |
| v1.1 | 2024 | Dark mode, build script, documentation |
| v1.2 | Q1 2025 | Custom fields, filtering, preview |
| v1.3 | Q2 2025 | Templates, scheduling, settings |
| v1.4 | Q3 2025 | Unsubscribe, CC/BCC, auto-save |
| v2.0 | Q4 2025 | macOS support |
| v2.1 | Q1 2026 | Enterprise features |

---

*Last Updated: December 2024*
*Maintained by: MailMerge Go Team*
