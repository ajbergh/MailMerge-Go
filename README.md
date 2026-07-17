# MailMerge Go

A Windows desktop application for Outlook-based email mail merge. Built with a Go
backend (classic Outlook COM automation) and a React/TypeScript frontend, bundled
with Wails v2.

> **Status:** actively developed, not yet certified production-ready. Core sending
> is guarded by a campaign preflight and a typed result model; automated tests
> cover the campaign engine without Outlook. See [dev_docs/](dev_docs/) for
> architecture and the [ROADMAP](ROADMAP.md) for what is implemented vs planned.
>
> **Outlook support:** classic desktop Outlook only. The **New Outlook** app does
> **not** expose COM automation and is not supported for sending.

## Features

### Contact Management
- **Contact Import**: Import contacts from CSV or Excel (.xlsx) files
- **Custom Merge Fields**: Use any imported column as a merge field
- **Contact Filtering and Selection**: Search contacts and select recipients by stable ID
- **Duplicate Handling**: Detect duplicate recipient addresses before sending and apply the configured policy
- **Recent Files**: Quick access to recently opened contact files

### Email Composition
- **Rich Text Editor**: Compose emails in plain text or HTML with a WYSIWYG editor
- **Dynamic Merge Fields**: Merge fields auto-generated from your file columns
- **Email Preview**: Render a preview through the same backend renderer used for sending
- **Email Templates**: Save and load reusable templates, including built-in templates
- **CC/BCC Support**: Add static or personalized CC and BCC recipient lists

### Attachments
- **Multiple Attachments**: Add multiple file attachments to your emails
- **Drag & Drop**: Drag files directly onto the attachment area
- **Size Tracking**: View individual and total attachment sizes

### Sending
- **Test Emails**: Send test emails before bulk sending
- **Real-time Progress**: Track sending progress with live updates
- **Retry Failed**: Retry only recipients whose latest attempt failed
- **Rate Limiting**: Configure validated inter-message and batch pacing
- **Export Logs**: Export send results to CSV for record-keeping
- **Outlook Integration**: Uses Microsoft Outlook COM automation for reliable email delivery

### UI/UX
- **Dark/Light Mode**: Toggle between dark and light themes
- **Settings Panel**: Centralized settings with theme, delay, and notification options
- **Keyboard Shortcuts**: Power user shortcuts for common actions
- **Responsive Design**: Clean, modern interface that works on any screen size

### Keyboard Shortcuts
- `Ctrl+O` - Open contact file
- `Ctrl+S` - Save current email as template
- `Ctrl+Enter` - Send test email
- `Ctrl+Shift+Enter` - Send to all selected contacts
- `Ctrl+,` - Open settings
- `Ctrl+D` - Toggle dark/light mode
- `Escape` - Close any open modal

## Requirements

### For Users
- Windows 10 or Windows 11
- Microsoft Outlook installed and configured with an email account

### For Developers
- Go 1.24 or later (matches `go.mod`; CI pins 1.24.x)
- Node.js 20.19+ or 22.12+ (required by Vite 8)
- Wails CLI v2.11.0

## Installation

### Pre-built Binary

Download the latest release from the [Releases](../../releases) page and run `MailMergeApp.exe`.

### Build from Source

1. Install [Go](https://go.dev/dl/)
2. Install [Node.js](https://nodejs.org/)
3. Install Wails CLI:
   ```powershell
   go install github.com/wailsapp/wails/v2/cmd/wails@v2.11.0
   ```

4. Clone the repository and build:
   ```powershell
   cd MailMerge-Go
   .\scripts\build-windows.ps1
   ```

5. The executable will be at `build/bin/MailMergeApp.exe`

## Build Scripts

The supported Windows entry point is `scripts/build-windows.ps1`:

```powershell
.\scripts\build-windows.ps1                       # Production build
.\scripts\build-windows.ps1 -Clean                # Clean build output first
.\scripts\build-windows.ps1 -Debug                # Build with debug symbols
.\scripts\build-windows.ps1 -SkipFrontendInstall  # Reuse installed dependencies
```

`scripts/build-linux.sh` and `scripts/build-macos.sh` provide corresponding
Wails targets. Run them on their native platforms with the Wails system
dependencies installed. They build the desktop shell only; sending still requires
classic Outlook on Windows.

## Development

Run in development mode with hot-reload:
```powershell
wails dev
```

This will run a Vite development server with fast hot reload for frontend changes. 
A dev server also runs on http://localhost:34115 for browser-based debugging with access to Go methods.

## Usage

### 1. Import Contacts

Click "Select Contact List" and choose a CSV or Excel file. Required columns:
- `Email` (required) - Must contain valid email addresses

Optional but recommended columns:
- `FirstName` - Used for personalization
- `LastName` - Used for personalization

**Any additional columns** in your file become available as merge fields. For
example, a `Company` column becomes `{{company}}`.

Column names are case-insensitive. Sample file included in `samples/contacts.csv`.

After import:
- **Duplicates**: Duplicate recipient groups are surfaced in preflight and resolved by the configured policy before sending
- **Selection**: All contacts are selected by default. Use checkboxes to select/deselect
- **Search**: Use the search box to filter contacts by name or email

### 2. Compose Email

Enter your email subject and body. Merge field buttons are automatically generated based on your imported file columns.

**Merge field syntax** (see [dev_docs/merge-fields.md](dev_docs/merge-fields.md)):
- Canonical: `{{field_id}}` — e.g. `{{first_name}}`, `{{account_manager}}`
- With fallback: `{{first_name|there}}` — uses the fallback when the value is blank
- Legacy single-brace `{FirstName}` is still supported for backward compatibility

Columns with spaces, hyphens, punctuation, or non-ASCII characters are supported
(e.g. `Account Manager` → `{{account_manager}}`). Unknown fields are reported and
**block sending** rather than silently rendering as empty. Merge values are
HTML-escaped in HTML emails.

**Common fields**: `{{first_name}}`, `{{last_name}}`, `{{email}}`, plus any column
from your file.

Toggle between plain text and rich text (HTML) mode.

**Preview**: Click "Preview Email" to see exactly how your email will look for any specific contact.

### 3. Add Attachments (Optional)

Click "Add Attachments" to attach files, or **drag and drop** files directly onto the attachment area. File sizes are displayed, and you'll see a warning if total size exceeds 20MB.

### 4. Send Test Email

Before sending to all recipients, click "Send Test Email" to verify your template looks correct.

### 5. Send to Selected

Click "Send to Selected" to begin sending to all selected contacts. Progress will be displayed in real-time.

### 6. Review Results & Retry

After sending completes:
- View success/failure summary
- **Retry Failed**: If any emails failed, click "Retry Failed" to attempt them again
- Export the results log to CSV for your records

## Project Structure

```
MailMergeApp/
├── app.go                 # Wails bindings (thin controller over the engine)
├── main.go                # Application entry point + build-time version vars
├── backend/
│   ├── models/            # Shared data structs (Contact, Settings, Template, ...)
│   ├── mergefield/        # Canonical merge-field model + renderer + diagnostics
│   ├── email/             # net/mail validation, address lists, duplicate policy
│   ├── htmlutil/          # HTML normalization + sanitizer (bluemonday)
│   ├── campaign/          # EmailSender iface, FakeSender, preflight, runner, results
│   ├── outlook/           # Classic-Outlook COM sender (dedicated STA worker) + stub
│   ├── graph/             # Microsoft Graph sender (design-stage stub)
│   ├── storage/           # Campaign history repository (JSON impl)
│   └── services/          # File import, templates, settings, suppression persistence
├── frontend/
│   ├── src/
│   │   ├── App.tsx        # Main React component
│   │   ├── components/    # UI components (+ *.test.tsx)
│   │   ├── styles/        # CSS
│   │   └── types/         # TypeScript types
│   └── wailsjs/           # Generated Wails bindings
├── .github/workflows/     # CI + release pipelines
├── scripts/               # Native-platform Wails build entry points
├── dev_docs/              # Architecture, merge fields, testing, security, ...
├── samples/contacts.csv   # Sample contact file
└── build/bin/             # Built executables
```

## Testing

```powershell
# Go: unit tests for the campaign engine, merge fields, email, services, storage
go test ./backend/...
go vet ./backend/...

# Frontend: type check, unit tests, production build
cd frontend
npm ci
npm run test
npm run build
```

Outlook COM is isolated behind an `EmailSender` interface, so the campaign engine
is fully testable without Outlook installed. Windows/Outlook integration tests are
opt-in behind a build tag:

```powershell
go test -tags outlookintegration ./backend/outlook/...
```

See [dev_docs/testing.md](dev_docs/testing.md).

## Continuous Integration

GitHub Actions (`.github/workflows/ci.yml`) runs on every pull request and push to
`main`: gofmt check, `go vet`, `go test -race`, golangci-lint, frontend
test/build (plus lint when configured), a Wails Windows build with an uploaded unsigned artifact, and
dependency scanning (govulncheck + npm audit). Releases are cut only from version
tags via `release.yml`.

## Data Storage & Privacy

Application data is stored locally under your Windows user profile
(`%AppData%\MailMergeGo`):

- `settings.json` — application settings
- `templates/*.json` — saved email templates
- `suppression.json` — local do-not-send list
- `campaigns/*.json` — campaign run history

No contact data, credentials, or email content is sent anywhere by this
application; email is submitted through your local Outlook profile. No tracking
pixels or unsubscribe links are added to messages. See
[dev_docs/security.md](dev_docs/security.md).

## Security Notes

- Imported contact values are HTML-escaped before being placed into HTML emails,
  and authored HTML bodies are sanitized (scripts, event handlers, unsafe URLs,
  iframes, and embedded objects are removed).
- Previews are additionally sanitized and rendered in a sandboxed iframe.
- Template IDs are backend-generated and cannot traverse directories; built-in
  templates cannot be overwritten.
- Settings, templates, suppression list, and campaign history are written
  atomically (temp file + rename).
- "Submitted" reflects acceptance by Outlook for sending; it does not confirm
  delivery.

## Troubleshooting

### "Outlook is not available"

- Ensure Microsoft Outlook is installed (not the web version)
- Make sure you have configured an email account in Outlook
- Try opening Outlook manually first and ensure it's fully loaded
- Close any Outlook security dialogs that may be blocking

### "Failed to send email" or "Interface marshalled for different thread"

This COM threading error has been fixed in the latest version. If you encounter it:
- Ensure you're running the latest build
- The application now properly locks OS threads for COM operations
- Verify the configured sending delay and any Outlook security prompts before retrying

### "Failed to create mail item"

- Check that Outlook is not showing any dialogs or security prompts
- Verify your email account is properly configured and can send emails manually
- Check if your IT department has security policies blocking COM automation
- Try restarting Outlook and the application

### Attachments not working

- Verify the file paths are correct and files exist
- Ensure files are not locked by another application
- Check that you have read permissions on the files

### Contact file not parsing correctly

- Ensure your file has a header row
- Required column: `Email` (case-insensitive)
- Optional columns: `FirstName`, `LastName` (or `First Name`, `first_name`)
- Verify the file is not open in another application

## License

MIT License

## Credits

Built with:
- [Wails](https://wails.io/) - Desktop application framework
- [go-ole](https://github.com/go-ole/go-ole) - COM bindings for Go
- [excelize](https://github.com/xuri/excelize) - Excel file processing
- [React](https://react.dev/) - UI framework
- [Lucide](https://lucide.dev/) - Icons
- [React Quill](https://github.com/zenoamaro/react-quill) - Rich text editor

## Configuration

You can configure the project by editing `wails.json`. More information about the project settings can be found
in the [Wails Documentation](https://wails.io/docs/reference/project-config).
