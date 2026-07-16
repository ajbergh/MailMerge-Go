# MailMerge Go

A production-ready Windows desktop application for Outlook-based email mail merge. Built with Go backend (using COM bindings) and React frontend, bundled with Wails.

## Features

### Contact Management
- **Contact Import**: Import contacts from CSV or Excel (.xlsx) files
- **Custom Merge Fields**: Use any column from your file as a merge field (v1.2)
- **Contact Filtering**: Search and filter contacts by name or email (v1.2)
- **Contact Selection**: Select specific contacts with checkboxes (v1.2)
- **Duplicate Detection**: Automatic warning for duplicate email addresses (v1.2)
- **Recent Files**: Quick access to recently opened contact files (v1.3)

### Email Composition
- **Rich Text Editor**: Compose emails in plain text or HTML with a WYSIWYG editor
- **Dynamic Merge Fields**: Merge fields auto-generated from your file columns
- **Email Preview**: Preview how emails look for specific contacts before sending (v1.2)
- **Email Templates**: Save and load email templates for reuse (v1.3)
- **Built-in Templates**: 6 starter templates included (Welcome, Newsletter, Meeting, etc.) (v1.3)
- **CC/BCC Support**: Add CC and BCC recipients with optional merge field support (v1.4)

### Attachments
- **Multiple Attachments**: Add multiple file attachments to your emails
- **Drag & Drop**: Drag files directly onto the attachment area (v1.2)
- **Size Tracking**: View individual and total attachment sizes (v1.2)

### Sending
- **Test Emails**: Send test emails before bulk sending
- **Real-time Progress**: Track sending progress with live updates
- **Retry Failed**: Retry sending to failed contacts with one click (v1.2)
- **Rate Limiting**: Configure sending delay (100ms-5000ms) for Outlook stability (v1.3)
- **Export Logs**: Export send results to CSV for record-keeping
- **Outlook Integration**: Uses Microsoft Outlook COM automation for reliable email delivery

### UI/UX
- **Dark/Light Mode**: Toggle between dark and light themes
- **Settings Panel**: Centralized settings with theme, delay, and notification options (v1.3)
- **Keyboard Shortcuts**: Power user shortcuts for common actions (v1.3)
- **Responsive Design**: Clean, modern interface that works on any screen size

### Keyboard Shortcuts (v1.3)
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
- Go 1.21 or later
- Node.js 18 or later
- Wails CLI v2

## Installation

### Pre-built Binary

Download the latest release from the [Releases](../../releases) page and run `MailMergeApp.exe`.

### Build from Source

1. Install [Go](https://go.dev/dl/)
2. Install [Node.js](https://nodejs.org/)
3. Install Wails CLI:
   ```powershell
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

4. Clone the repository and build:
   ```powershell
   cd MailMergeApp
   .\build.ps1
   ```

5. The executable will be at `build/bin/MailMergeApp.exe`

## Build Script Options

The included PowerShell build script (`build.ps1`) provides several options:

```powershell
.\build.ps1                 # Production build
.\build.ps1 -Clean          # Clean build directory before building
.\build.ps1 -Dev            # Development build (includes devtools)
.\build.ps1 -Debug          # Build with debug symbols
.\build.ps1 -NoPackage      # Skip packaging step
.\build.ps1 -Help           # Show all options
```

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

**Any additional columns** in your file become available as merge fields! For example, if your file has a "Company" column, you can use `{Company}` in your email.

Column names are case-insensitive. Sample file included in `samples/contacts.csv`.

After import:
- **Duplicates**: Any duplicate email addresses will be highlighted with a warning icon
- **Selection**: All contacts are selected by default. Use checkboxes to select/deselect
- **Search**: Use the search box to filter contacts by name or email

### 2. Compose Email

Enter your email subject and body. Merge field buttons are automatically generated based on your imported file columns.

**Common Merge Fields**:
- `{FirstName}` - Recipient's first name
- `{LastName}` - Recipient's last name
- `{Email}` - Recipient's email address
- Plus any custom columns from your file!

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
├── app.go                 # Main app with Wails bindings
├── main.go                # Application entry point
├── backend/
│   ├── models/
│   │   └── contact.go     # Data models (Contact, Template, Settings)
│   └── services/
│       ├── file_service.go      # CSV/Excel parsing
│       ├── merge_service.go     # Template merging
│       ├── outlook_service.go   # Outlook COM automation
│       ├── template_service.go  # Email template management (v1.3)
│       └── settings_service.go  # App settings persistence (v1.3)
├── frontend/
│   ├── src/
│   │   ├── App.tsx        # Main React component
│   │   ├── components/    # UI components
│   │   │   ├── TemplateManager.tsx  # Template selection (v1.3)
│   │   │   ├── SettingsModal.tsx    # Settings panel (v1.3)
│   │   │   └── ...                  # Other components
│   │   ├── styles/        # CSS styles
│   │   └── types/         # TypeScript types
│   └── wailsjs/           # Generated Wails bindings
├── samples/
│   └── contacts.csv       # Sample contact file
└── build/
    └── bin/               # Built executables
```

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
- A 500ms delay between emails helps maintain COM stability

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
