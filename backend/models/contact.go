/*
Models Package - Data Structures

This package defines all data structures used throughout the MailMerge application.
These models are shared between the Go backend and the frontend via Wails bindings.

All structs with JSON tags are automatically serialized/deserialized when passed
between Go and JavaScript.

Phase 1 Updates (v1.2):
  - Contact now includes CustomFields map for dynamic merge fields
  - ParseResult includes Warnings and Duplicates for better feedback
  - SendResult includes FailedContacts for retry functionality

Phase 2 Updates (v1.3):
  - Added EmailTemplate model for template persistence
  - Added AppSettings model for user preferences

Phase 3 Updates (v1.4):
  - EmailRequest and TestEmailRequest now support CC and BCC fields
*/
package models

import "time"

// Contact represents a single recipient from the mailing list.
// At minimum, an email address is required. First and last names
// are optional but enable personalization via merge fields.
// CustomFields contains any additional columns from the import file.
type Contact struct {
	ID           string            `json:"id"`                     // Stable per-import identifier for selection/dedup/editing
	FirstName    string            `json:"firstName"`              // Recipient's first name (optional)
	LastName     string            `json:"lastName"`               // Recipient's last name (optional)
	Email        string            `json:"email"`                  // Recipient's email address (required)
	CustomFields map[string]string `json:"customFields,omitempty"` // Additional columns from import file
}

// GetField retrieves a field value by name (case-insensitive).
// Checks FirstName, LastName, Email first, then CustomFields.
func (c *Contact) GetField(name string) string {
	switch name {
	case "firstname":
		return c.FirstName
	case "lastname":
		return c.LastName
	case "email":
		return c.Email
	default:
		if c.CustomFields != nil {
			return c.CustomFields[name]
		}
		return ""
	}
}

// EmailLog represents the result of sending an email to a single contact.
// Used for tracking success/failure and for generating export reports.
type EmailLog struct {
	FirstName    string    `json:"firstName"`              // Contact's first name
	LastName     string    `json:"lastName"`               // Contact's last name
	Email        string    `json:"email"`                  // Contact's email address
	Status       string    `json:"status"`                 // "Success" or "Failure"
	ErrorMessage string    `json:"errorMessage,omitempty"` // Error details if failed
	Timestamp    time.Time `json:"timestamp"`              // When the send was attempted
}

// ParseResult represents the result of parsing a contact file (CSV or Excel).
// Contains both successfully parsed contacts and any parsing errors encountered.
// Phase 1: Now includes warnings, duplicates, and detected column headers.
type ParseResult struct {
	Contacts   []Contact `json:"contacts"`             // Successfully parsed contacts
	Errors     []string  `json:"errors,omitempty"`     // Row-level parsing errors
	Warnings   []string  `json:"warnings,omitempty"`   // Non-fatal warnings (e.g., duplicates)
	Total      int       `json:"total"`                // Total number of valid contacts
	Headers    []string  `json:"headers,omitempty"`    // All detected column headers
	Duplicates []string  `json:"duplicates,omitempty"` // List of duplicate email addresses
}

// EmailRequest represents a request to send bulk emails.
// Contains all data needed to execute a mail merge operation.
// Phase 3: Added CC and BCC fields for carbon copy recipients.
type EmailRequest struct {
	Contacts        []Contact `json:"contacts"`              // List of recipients
	SubjectTemplate string    `json:"subjectTemplate"`       // Subject with merge fields
	BodyTemplate    string    `json:"bodyTemplate"`          // Body with merge fields
	IsHTML          bool      `json:"isHTML"`                // True for HTML email, false for plain text
	Attachments     []string  `json:"attachments"`           // Absolute file paths to attach
	CC              string    `json:"cc,omitempty"`          // Static CC addresses (comma-separated)
	BCC             string    `json:"bcc,omitempty"`         // Static BCC addresses (comma-separated)
	CCTemplate      string    `json:"ccTemplate,omitempty"`  // CC with merge fields support
	BCCTemplate     string    `json:"bccTemplate,omitempty"` // BCC with merge fields support
}

// TestEmailRequest represents a request to send a single test email.
// Includes sample data to preview how merge fields will be rendered.
// Phase 3: Added CC and BCC fields for carbon copy recipients.
type TestEmailRequest struct {
	TestAddress     string   `json:"testAddress"`           // Email address to send test to
	SubjectTemplate string   `json:"subjectTemplate"`       // Subject with merge fields
	BodyTemplate    string   `json:"bodyTemplate"`          // Body with merge fields
	IsHTML          bool     `json:"isHTML"`                // True for HTML email
	Attachments     []string `json:"attachments"`           // File paths to attach
	CC              string   `json:"cc,omitempty"`          // Static CC addresses (comma-separated)
	BCC             string   `json:"bcc,omitempty"`         // Static BCC addresses (comma-separated)
	CCTemplate      string   `json:"ccTemplate,omitempty"`  // CC with merge fields support
	BCCTemplate     string   `json:"bccTemplate,omitempty"` // BCC with merge fields support
	// Sample data for merge preview
	SampleFirstName string `json:"sampleFirstName"` // Sample first name for preview
	SampleLastName  string `json:"sampleLastName"`  // Sample last name for preview
}

// ProgressUpdate represents a real-time progress update during bulk sending.
// Emitted as an event to the frontend for live progress tracking.
type ProgressUpdate struct {
	Current   int    `json:"current"`           // Current email number being processed
	Total     int    `json:"total"`             // Total number of emails to send
	Status    string `json:"status"`            // "sending", "success", "failure", "complete"
	Email     string `json:"email"`             // Email address currently being processed
	Message   string `json:"message,omitempty"` // Additional status message
	Timestamp string `json:"timestamp"`         // ISO 8601 timestamp
}

// SendResult represents the final result of a bulk send operation.
// Contains aggregate counts and detailed logs for each email sent.
// Phase 1: Now includes FailedContacts for retry functionality.
type SendResult struct {
	TotalSent      int        `json:"totalSent"`                // Number of emails sent successfully
	TotalFailed    int        `json:"totalFailed"`              // Number of emails that failed
	Logs           []EmailLog `json:"logs"`                     // Detailed log for each email
	FailedContacts []Contact  `json:"failedContacts,omitempty"` // Contacts that failed, for retry
}

// FileInfo represents metadata about an uploaded or selected file.
// Used to display file information in the UI.
type FileInfo struct {
	Name string `json:"name"` // File name without path
	Path string `json:"path"` // Absolute file path
	Size int64  `json:"size"` // File size in bytes
}

// ==================== Phase 2 Models ====================

// EmailTemplate represents a saved email template for reuse.
// Templates can be saved, loaded, and managed by the user.
// Phase 2: New model for template persistence.
type EmailTemplate struct {
	ID        string    `json:"id"`        // Unique identifier (UUID)
	Name      string    `json:"name"`      // User-friendly template name
	Subject   string    `json:"subject"`   // Subject line with merge fields
	Body      string    `json:"body"`      // Body content with merge fields
	IsHTML    bool      `json:"isHTML"`    // True for HTML format, false for plain text
	IsBuiltIn bool      `json:"isBuiltIn"` // True if this is a built-in template
	CreatedAt time.Time `json:"createdAt"` // When the template was created
	UpdatedAt time.Time `json:"updatedAt"` // When the template was last modified
}

// AppSettings represents user preferences and application settings.
// These are persisted to disk and loaded on startup.
// Phase 2: New model for centralized settings management.
type AppSettings struct {
	// Display Settings
	Theme         string `json:"theme"`         // "light", "dark", or "system"
	DefaultFormat string `json:"defaultFormat"` // "html" or "plaintext"

	// Sending Settings
	SendingDelay   int  `json:"sendingDelay"`   // Delay between emails in milliseconds
	ConfirmSend    bool `json:"confirmSend"`    // Show confirmation before sending
	SoundEnabled   bool `json:"soundEnabled"`   // Play sound on completion
	AutoSaveTempls bool `json:"autoSaveTempls"` // Auto-save templates on exit

	// Recent Files
	RecentFiles []string `json:"recentFiles"` // Last 10 opened contact files
}

// DefaultSettings returns the default application settings.
func DefaultSettings() *AppSettings {
	return &AppSettings{
		Theme:          "light",
		DefaultFormat:  "html",
		SendingDelay:   500,
		ConfirmSend:    true,
		SoundEnabled:   true,
		AutoSaveTempls: true,
		RecentFiles:    []string{},
	}
}
