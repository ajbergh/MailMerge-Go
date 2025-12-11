/*
App - Main Application Controller

This file contains the primary application struct that serves as the bridge
between the frontend UI and backend services. All methods exported on the App
struct are automatically bound to the frontend and can be called from JavaScript.

The App orchestrates:
  - File operations (selecting and parsing contact files)
  - Email composition and template merging
  - Outlook integration for sending emails
  - Log export functionality
  - Phase 2: Template management and settings persistence

Thread Safety:
  - All Outlook operations are internally synchronized
  - File operations are stateless and thread-safe
*/
package main

import (
	"MailMergeApp/backend/models"
	"MailMergeApp/backend/services"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds the application state and service dependencies.
// It is the primary struct bound to the frontend, providing all
// functionality accessible from the UI.
// Phase 2: Added template and settings services.
type App struct {
	ctx             context.Context           // Wails context for runtime operations
	fileService     *services.FileService     // Handles CSV/Excel file parsing
	mergeService    *services.MergeService    // Handles template variable replacement
	outlookService  *services.OutlookService  // Handles Outlook COM automation
	templateService *services.TemplateService // Phase 2: Handles email template persistence
	settingsService *services.SettingsService // Phase 2: Handles app settings persistence
}

// NewApp creates a new App application struct with initialized services.
// This is called once during application startup.
// Phase 2: Added template and settings service initialization.
func NewApp() *App {
	templateService, err := services.NewTemplateService()
	if err != nil {
		fmt.Printf("Warning: failed to initialize template service: %v\n", err)
	}

	settingsService, err := services.NewSettingsService()
	if err != nil {
		fmt.Printf("Warning: failed to initialize settings service: %v\n", err)
	}

	return &App{
		fileService:     services.NewFileService(),
		mergeService:    services.NewMergeService(),
		outlookService:  services.NewOutlookService(),
		templateService: templateService,
		settingsService: settingsService,
	}
}

// startup is called when the app starts. It receives the Wails context
// which is used for runtime operations like dialogs and events.
// This method initializes any context-dependent services.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.outlookService.SetContext(ctx)
}

// ================== File Operations ==================

// SelectContactFile opens a native file dialog allowing the user to select
// a contact list file. Supports both CSV and Excel (.xlsx) formats.
//
// Returns:
//   - string: The full path to the selected file, or empty if cancelled
//   - error: Any error encountered while opening the dialog
func (a *App) SelectContactFile() (string, error) {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Contact List",
		Filters: []runtime.FileFilter{
			{DisplayName: "Contact Files (*.csv, *.xlsx)", Pattern: "*.csv;*.xlsx"},
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
			{DisplayName: "Excel Files (*.xlsx)", Pattern: "*.xlsx"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}
	return file, nil
}

// ParseContactFile validates and parses a contact file (CSV or Excel).
// The file must contain at minimum an "Email" column. "FirstName" and
// "LastName" columns are optional but recommended.
//
// Parameters:
//   - filePath: Absolute path to the contact file
//
// Returns:
//   - *models.ParseResult: Contains parsed contacts and any parsing errors
//   - error: Fatal errors that prevented parsing
func (a *App) ParseContactFile(filePath string) (*models.ParseResult, error) {
	if err := a.fileService.ValidateFile(filePath); err != nil {
		return nil, err
	}
	return a.fileService.ParseContactFile(filePath)
}

// SelectAttachments opens a native multi-file dialog allowing the user
// to select one or more files to attach to the emails.
//
// Returns:
//   - []string: Array of full paths to selected files
//   - error: Any error encountered while opening the dialog
func (a *App) SelectAttachments() ([]string, error) {
	files, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Attachments",
		Filters: []runtime.FileFilter{
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open file dialog: %w", err)
	}
	return files, nil
}

// GetFileInfo returns metadata about a file at the specified path.
// Used to display file information in the UI.
//
// Parameters:
//   - filePath: Absolute path to the file
//
// Returns:
//   - *models.FileInfo: File metadata (name, path, size)
//   - error: If file cannot be accessed
func (a *App) GetFileInfo(filePath string) (*models.FileInfo, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return &models.FileInfo{
		Name: info.Name(),
		Path: filePath,
		Size: info.Size(),
	}, nil
}

// ================== Merge Operations ==================

// PreviewMerge renders the subject and body templates with sample data
// to show the user how their merged email will look.
//
// Parameters:
//   - subjectTemplate: Email subject with optional merge fields like {FirstName}
//   - bodyTemplate: Email body with optional merge fields
//   - sampleFirstName, sampleLastName, sampleEmail: Sample data for preview
//
// Returns:
//   - *services.MergeResult: Rendered subject and body with fields replaced
func (a *App) PreviewMerge(subjectTemplate, bodyTemplate, sampleFirstName, sampleLastName, sampleEmail string) *services.MergeResult {
	result := a.mergeService.PreviewMerge(subjectTemplate, bodyTemplate, sampleFirstName, sampleLastName, sampleEmail)
	return &result
}

// GetMergeFields returns the list of available merge fields that can be
// used in email templates. Returns default fields: {FirstName}, {LastName}, {Email}
// For dynamic fields from imported files, use GetMergeFieldsFromHeaders.
func (a *App) GetMergeFields() []string {
	return a.mergeService.GetAvailableFields()
}

// GetMergeFieldsFromHeaders converts column headers from an imported file
// to merge field format. Phase 1: Enables dynamic merge fields from any column.
//
// Parameters:
//   - headers: Column headers from the imported file
//
// Returns:
//   - []string: Merge fields in {FieldName} format
func (a *App) GetMergeFieldsFromHeaders(headers []string) []string {
	return a.mergeService.GetMergeFieldsFromHeaders(headers)
}

// PreviewMergeForContact renders the subject and body templates for a specific contact.
// Phase 1: Added for the email preview modal feature.
//
// Parameters:
//   - subjectTemplate: Email subject with merge fields
//   - bodyTemplate: Email body with merge fields
//   - contact: Contact data to use for preview
//
// Returns:
//   - *services.MergeResult: Rendered subject and body
func (a *App) PreviewMergeForContact(subjectTemplate, bodyTemplate string, contact models.Contact) *services.MergeResult {
	result := a.mergeService.PreviewMergeWithContact(subjectTemplate, bodyTemplate, contact)
	return &result
}

// ValidateTemplate checks a template string for invalid merge fields.
// Returns a list of unrecognized field names found in the template.
//
// Parameters:
//   - template: Template string to validate
//
// Returns:
//   - []string: List of invalid field names (empty if all valid)
func (a *App) ValidateTemplate(template string) []string {
	return a.mergeService.ValidateTemplate(template)
}

// ================== Email Operations ==================

// CheckOutlookInstalled verifies that Microsoft Outlook is installed
// and accessible via COM automation. Should be called on app startup.
//
// Returns:
//   - error: Non-nil if Outlook is not available
func (a *App) CheckOutlookInstalled() error {
	return a.outlookService.CheckOutlookInstalled()
}

// SendTestEmail sends a single test email to verify the configuration.
// Uses sample data to render the template before sending.
// Phase 3: Now supports CC and BCC recipients.
//
// Parameters:
//   - request: Contains test address, templates, sample data, CC, and BCC
//
// Returns:
//   - error: Non-nil if send failed
func (a *App) SendTestEmail(request models.TestEmailRequest) error {
	// Render templates with sample data
	merged := a.mergeService.PreviewMerge(
		request.SubjectTemplate,
		request.BodyTemplate,
		request.SampleFirstName,
		request.SampleLastName,
		request.TestAddress,
	)

	// Render CC/BCC templates if provided (use static values for test)
	cc := request.CC
	if request.CCTemplate != "" {
		ccMerged := a.mergeService.PreviewMerge("", request.CCTemplate, request.SampleFirstName, request.SampleLastName, request.TestAddress)
		cc = ccMerged.Body
	}
	bcc := request.BCC
	if request.BCCTemplate != "" {
		bccMerged := a.mergeService.PreviewMerge("", request.BCCTemplate, request.SampleFirstName, request.SampleLastName, request.TestAddress)
		bcc = bccMerged.Body
	}

	return a.outlookService.SendTestEmail(
		request.TestAddress,
		merged.Subject,
		merged.Body,
		request.IsHTML,
		request.Attachments,
		cc,
		bcc,
	)
}

// SendBulkEmails sends personalized emails to all contacts in the list.
// Progress updates are emitted as events for real-time UI feedback.
// Phase 3: Now supports CC and BCC with optional merge field templates.
//
// Each email is rendered with the contact's data replacing merge fields,
// then sent via Outlook COM automation with a delay between sends.
//
// Parameters:
//   - request: Contains contacts, templates, HTML flag, attachments, CC, and BCC
//
// Returns:
//   - *models.SendResult: Summary with success/failure counts and logs
func (a *App) SendBulkEmails(request models.EmailRequest) *models.SendResult {
	return a.outlookService.SendBulkEmails(
		request.Contacts,
		request.SubjectTemplate,
		request.BodyTemplate,
		request.IsHTML,
		request.Attachments,
		request.CC,
		request.BCC,
		request.CCTemplate,
		request.BCCTemplate,
	)
}

// ================== Export Operations ==================

// ExportLogsToCSV exports email send logs to a CSV file for record-keeping.
// Opens a native save dialog for the user to choose the destination.
//
// The CSV includes columns: FirstName, LastName, Email, Status, ErrorMessage, Timestamp
//
// Parameters:
//   - logs: Array of email log entries to export
//
// Returns:
//   - string: Path to the saved file, or empty if cancelled
//   - error: Any error during file creation or writing
func (a *App) ExportLogsToCSV(logs []models.EmailLog) (string, error) {
	// Open save dialog
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Email Logs",
		DefaultFilename: fmt.Sprintf("email_logs_%s.csv", time.Now().Format("2006-01-02_15-04-05")),
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open save dialog: %w", err)
	}

	if savePath == "" {
		return "", nil // User cancelled
	}

	// Ensure .csv extension
	if filepath.Ext(savePath) != ".csv" {
		savePath += ".csv"
	}

	// Create the file
	file, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"FirstName", "LastName", "Email", "Status", "ErrorMessage", "Timestamp"}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write header: %w", err)
	}

	// Write data rows
	for _, log := range logs {
		row := []string{
			log.FirstName,
			log.LastName,
			log.Email,
			log.Status,
			log.ErrorMessage,
			log.Timestamp.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write row: %w", err)
		}
	}

	return savePath, nil
}

// ================== Utility Methods ==================

// GetAppInfo returns application metadata for display in the UI.
// Useful for about dialogs or version checking.
func (a *App) GetAppInfo() map[string]string {
	return map[string]string{
		"name":    "MailMerge Go",
		"version": "1.3.0",
		"author":  "Your Name",
	}
}

// ================== Phase 2: Template Operations ==================

// GetAllTemplates returns all saved email templates.
// Templates are sorted with built-in templates first, then user templates alphabetically.
//
// Returns:
//   - []models.EmailTemplate: List of all templates
func (a *App) GetAllTemplates() []models.EmailTemplate {
	if a.templateService == nil {
		return []models.EmailTemplate{}
	}
	return a.templateService.GetAllTemplates()
}

// GetTemplate returns a single template by ID.
//
// Parameters:
//   - id: Template ID to retrieve
//
// Returns:
//   - *models.EmailTemplate: The template, or nil if not found
func (a *App) GetTemplate(id string) *models.EmailTemplate {
	if a.templateService == nil {
		return nil
	}
	return a.templateService.GetTemplate(id)
}

// SaveTemplate saves a new or updated email template.
// If the template has no ID, a new one will be generated.
//
// Parameters:
//   - template: Template data to save
//
// Returns:
//   - *models.EmailTemplate: The saved template with updated timestamps
//   - error: Any error encountered during save
func (a *App) SaveTemplate(template models.EmailTemplate) (*models.EmailTemplate, error) {
	if a.templateService == nil {
		return nil, fmt.Errorf("template service not available")
	}
	return a.templateService.SaveTemplate(template)
}

// DeleteTemplate removes a template by ID.
// Built-in templates cannot be deleted.
//
// Parameters:
//   - id: Template ID to delete
//
// Returns:
//   - error: Any error encountered during deletion
func (a *App) DeleteTemplate(id string) error {
	if a.templateService == nil {
		return fmt.Errorf("template service not available")
	}
	return a.templateService.DeleteTemplate(id)
}

// ================== Phase 2: Settings Operations ==================

// GetSettings returns the current application settings.
//
// Returns:
//   - *models.AppSettings: Current settings
func (a *App) GetSettings() *models.AppSettings {
	if a.settingsService == nil {
		return models.DefaultSettings()
	}
	return a.settingsService.GetSettings()
}

// UpdateSettings saves updated application settings.
//
// Parameters:
//   - settings: New settings to apply
//
// Returns:
//   - error: Any error encountered during save
func (a *App) UpdateSettings(settings models.AppSettings) error {
	if a.settingsService == nil {
		return fmt.Errorf("settings service not available")
	}
	return a.settingsService.UpdateSettings(settings)
}

// GetRecentFiles returns the list of recently opened contact files.
//
// Returns:
//   - []string: File paths of recent files
func (a *App) GetRecentFiles() []string {
	if a.settingsService == nil {
		return []string{}
	}
	return a.settingsService.GetRecentFiles()
}

// AddRecentFile adds a file path to the recent files list.
//
// Parameters:
//   - filePath: Path to add to recent files
//
// Returns:
//   - error: Any error encountered during save
func (a *App) AddRecentFile(filePath string) error {
	if a.settingsService == nil {
		return nil
	}
	return a.settingsService.AddRecentFile(filePath)
}

// ClearRecentFiles clears the recent files list.
//
// Returns:
//   - error: Any error encountered during save
func (a *App) ClearRecentFiles() error {
	if a.settingsService == nil {
		return nil
	}
	return a.settingsService.ClearRecentFiles()
}

// GetSendingDelay returns the current sending delay in milliseconds.
//
// Returns:
//   - int: Delay between emails in milliseconds
func (a *App) GetSendingDelay() int {
	if a.settingsService == nil {
		return 500
	}
	return a.settingsService.GetSendingDelay()
}

// SetSendingDelay updates the sending delay.
//
// Parameters:
//   - delay: Delay in milliseconds (100-5000)
//
// Returns:
//   - error: Any error encountered during save
func (a *App) SetSendingDelay(delay int) error {
	if a.settingsService == nil {
		return nil
	}
	return a.settingsService.SetSendingDelay(delay)
}
