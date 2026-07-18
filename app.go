/*
App - Main Application Controller

This file contains the primary application struct that serves as the bridge
between the frontend UI and backend services. All methods exported on the App
struct are automatically bound to the frontend and can be called from JavaScript.

The App orchestrates:
  - File operations (selecting and parsing contact files)
  - Preview, preflight, and campaign execution through the shared renderer
  - Sender capability checks and local Outlook integration
  - Template, settings, suppression, and campaign-history persistence
  - Log export and Wails runtime dialogs/events

Thread safety: the App serializes access to the active campaign cancellation
function and last result. The Outlook sender owns its dedicated COM worker.
*/
package main

import (
	"MailMergeApp/backend/campaign"
	"MailMergeApp/backend/email"
	"MailMergeApp/backend/models"
	"MailMergeApp/backend/outlook"
	"MailMergeApp/backend/services"
	"MailMergeApp/backend/storage"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App holds the Wails-bound application state and service dependencies.
type App struct {
	ctx             context.Context              // Wails context for runtime operations
	fileService     *services.FileService        // Handles CSV/Excel file parsing
	mergeService    *services.MergeService       // Preview/validate merge helper
	sender          campaign.EmailSender         // Outlook COM (or stub) behind the sender interface
	runner          *campaign.Runner             // Platform-independent campaign engine
	templateService *services.TemplateService    // Handles email template persistence
	settingsService *services.SettingsService    // Handles app settings persistence
	suppression     *services.SuppressionService // Suppression / unsubscribe list
	history         storage.CampaignRepository   // Campaign run history (JSON-backed)

	mu             sync.Mutex         // guards cancel and lastResult
	cancel         context.CancelFunc // cancels the in-flight campaign, if any
	lastResult     *campaign.CampaignResult
	lastCampaignID string

	version   string // build-time version metadata
	commit    string
	buildDate string
}

// NewApp constructs the application services and wires them into the campaign
// runner. Persistent-service initialization failures are non-fatal: the related
// UI operation returns an availability error instead.
func NewApp() *App {
	templateService, err := services.NewTemplateService()
	if err != nil {
		fmt.Printf("Warning: failed to initialize template service: %v\n", err)
	}

	settingsService, err := services.NewSettingsService()
	if err != nil {
		fmt.Printf("Warning: failed to initialize settings service: %v\n", err)
	}

	suppression, err := services.NewSuppressionService()
	if err != nil {
		fmt.Printf("Warning: failed to initialize suppression service: %v\n", err)
	}

	sender := outlook.New()

	var history storage.CampaignRepository
	if appData, derr := os.UserConfigDir(); derr == nil {
		if repo, herr := storage.NewJSONCampaignRepository(filepath.Join(appData, "MailMergeGo", "campaigns")); herr == nil {
			history = repo
		} else {
			fmt.Printf("Warning: failed to initialize campaign history: %v\n", herr)
		}
	}

	return &App{
		fileService:     services.NewFileService(),
		mergeService:    services.NewMergeService(),
		sender:          sender,
		runner:          campaign.NewRunner(sender),
		templateService: templateService,
		settingsService: settingsService,
		suppression:     suppression,
		history:         history,
	}
}

// persistRun stores an immutable campaign snapshot and returns the result with
// durable campaign lineage metadata. Persistence is best-effort: a send result is
// still returned when history storage is unavailable, but retry-by-ID will not be
// available for that run.
func (a *App) persistRun(c campaign.Campaign, res campaign.CampaignResult, parentID string, runNumber int) campaign.CampaignResult {
	if a.history == nil {
		return res
	}
	if runNumber < 1 {
		runNumber = 1
	}

	id := uuid.NewString()
	res.CampaignID = id
	res.ParentCampaignID = parentID
	res.RunNumber = runNumber

	rec := storage.CampaignRecord{
		ID:               id,
		ParentCampaignID: parentID,
		RunNumber:        runNumber,
		CreatedAt:        time.Now(),
		StartedAt:        res.StartedAt,
		FinishedAt:       res.FinishedAt,
		Duration:         res.Duration,
		Subject:          c.SubjectTemplate,
		SubjectTemplate:  c.SubjectTemplate,
		BodyTemplate:     c.BodyTemplate,
		IsHTML:           c.IsHTML,
		DraftOnly:        c.DraftOnly,
		Headers:          cloneStrings(c.Headers),
		Contacts:         cloneContacts(c.Contacts),
		Attachments:      cloneStrings(c.Attachments),
		CCTemplate:       c.CCTemplate,
		BCCTemplate:      c.BCCTemplate,
		DuplicatePolicy:  c.DuplicatePolicy,
		SendOptions:      c.Options,
		SenderType:       string(campaign.StateClassicOutlook),
		RecipientCount:   len(c.Contacts),
		State:            res.State,
		Result:           res,
	}
	if err := a.history.Create(rec); err != nil {
		fmt.Printf("Warning: failed to record campaign run: %v\n", err)
		res.CampaignID = ""
		res.ParentCampaignID = ""
		res.RunNumber = 0
	}
	return res
}

// GetCampaignHistory returns past campaign runs, newest first.
func (a *App) GetCampaignHistory() []storage.CampaignRecord {
	if a.history == nil {
		return []storage.CampaignRecord{}
	}
	records, err := a.history.List()
	if err != nil {
		return []storage.CampaignRecord{}
	}
	return records
}

// GetCampaign returns the immutable snapshot for one campaign run.
func (a *App) GetCampaign(id string) (*storage.CampaignRecord, error) {
	if a.history == nil {
		return nil, fmt.Errorf("campaign history is unavailable")
	}
	return a.history.Get(id)
}

// DeleteCampaign deletes one campaign-history record. It never deletes linked
// parent or child runs implicitly.
func (a *App) DeleteCampaign(id string) error {
	if a.history == nil {
		return fmt.Errorf("campaign history is unavailable")
	}
	if err := a.history.Delete(id); err != nil {
		return err
	}
	a.mu.Lock()
	if a.lastCampaignID == id {
		a.lastCampaignID = ""
		a.lastResult = nil
	}
	a.mu.Unlock()
	return nil
}

// ClearCampaignHistory deletes every local campaign-history record.
func (a *App) ClearCampaignHistory() error {
	if a.history == nil {
		return fmt.Errorf("campaign history is unavailable")
	}
	records, err := a.history.List()
	if err != nil {
		return err
	}
	for _, rec := range records {
		if err := a.history.Delete(rec.ID); err != nil {
			return fmt.Errorf("delete campaign %s: %w", rec.ID, err)
		}
	}
	a.mu.Lock()
	a.lastCampaignID = ""
	a.lastResult = nil
	a.mu.Unlock()
	return nil
}

func (a *App) campaignFromRecord(rec storage.CampaignRecord) campaign.Campaign {
	var suppressed map[string]bool
	if a.suppression != nil {
		suppressed = a.suppression.Set()
	}
	return campaign.Campaign{
		Headers:         cloneStrings(rec.Headers),
		Contacts:        cloneContacts(rec.Contacts),
		SubjectTemplate: rec.SubjectTemplate,
		BodyTemplate:    rec.BodyTemplate,
		IsHTML:          rec.IsHTML,
		DraftOnly:       rec.DraftOnly,
		Attachments:     cloneStrings(rec.Attachments),
		CCTemplate:      rec.CCTemplate,
		BCCTemplate:     rec.BCCTemplate,
		Options:         rec.SendOptions.Normalized(),
		DuplicatePolicy: rec.DuplicatePolicy,
		Suppressed:      suppressed,
	}
}

func cloneStrings(in []string) []string {
	return append([]string(nil), in...)
}

func cloneContacts(in []models.Contact) []models.Contact {
	out := make([]models.Contact, len(in))
	for i, contact := range in {
		out[i] = contact
		if contact.CustomFields != nil {
			out[i].CustomFields = make(map[string]string, len(contact.CustomFields))
			for key, value := range contact.CustomFields {
				out[i].CustomFields[key] = value
			}
		}
	}
	return out
}

// startup is called when the app starts. It receives the Wails context
// which is used for runtime operations like dialogs and events.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown releases the Outlook COM worker cleanly.
func (a *App) shutdown(ctx context.Context) {
	if closer, ok := a.sender.(interface{ Close() }); ok {
		closer.Close()
	}
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

// GetMergeFields returns the standard canonical merge tokens. For imported
// columns, use GetMergeFieldsFromHeaders to create the matching canonical tokens.
func (a *App) GetMergeFields() []string {
	return a.mergeService.GetAvailableFields()
}

// GetMergeFieldsFromHeaders canonicalizes imported column headers and returns the
// corresponding merge tokens (for example, `Account Manager` becomes
// `{{account_manager}}`).
//
// Parameters:
//   - headers: Column headers from the imported file
//
// Returns:
//   - []string: Canonical merge tokens in {{field_id}} format
func (a *App) GetMergeFieldsFromHeaders(headers []string) []string {
	return a.mergeService.GetMergeFieldsFromHeaders(headers)
}

// PreviewMergeForContact renders subject and body templates for one contact using
// the same merge service used by the campaign pipeline.
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

// CheckOutlookInstalled verifies that a usable sender (classic Outlook via COM)
// is available. Returns an actionable error when it is not.
func (a *App) CheckOutlookInstalled() error {
	status := a.sender.Preflight(a.ctx)
	if !status.Available {
		return fmt.Errorf("%s", status.Message)
	}
	return nil
}

// GetOutlookStatus returns a structured sender-readiness report for the UI.
func (a *App) GetOutlookStatus() campaign.SenderStatus {
	return a.sender.Preflight(a.ctx)
}

// GetSenderCapabilities reports what the active sender supports.
func (a *App) GetSenderCapabilities() campaign.SenderCapabilities {
	return a.sender.Capabilities(a.ctx)
}

// PreflightCampaign validates a campaign and returns structured results. The UI
// must display these before sending; bulk sending is impossible while CanSend
// is false.
func (a *App) PreflightCampaign(request models.EmailRequest) campaign.PreflightResult {
	c := a.buildCampaign(request, true)
	return campaign.NewPreflighter().Preflight(a.ctx, c, a.sender)
}

// SendTestEmail sends a single representative test email through the same
// preflight and rendering pipeline as a bulk send, using the selected contact's
// full data. The test recipient does not overwrite the contact's {{email}}
// value unless OverwriteEmail is set.
func (a *App) SendTestEmail(request models.TestEmailRequest) error {
	contact := request.Contact
	if contact.Email == "" {
		contact = models.Contact{
			FirstName: firstNonEmpty(request.SampleFirstName, "John"),
			LastName:  firstNonEmpty(request.SampleLastName, "Doe"),
			Email:     "sample@example.com",
		}
	}
	c := campaign.Campaign{
		Headers:         headersForContacts([]models.Contact{contact}),
		Contacts:        []models.Contact{contact},
		SubjectTemplate: request.SubjectTemplate,
		BodyTemplate:    request.BodyTemplate,
		IsHTML:          request.IsHTML,
		DraftOnly:       request.DraftOnly,
		Attachments:     request.Attachments,
		CCTemplate:      firstNonEmpty(request.CCTemplate, request.CC),
		BCCTemplate:     firstNonEmpty(request.BCCTemplate, request.BCC),
		ToOverride:      request.TestAddress,
		OverrideEmail:   request.OverwriteEmail,
		Options:         campaign.SendOptions{DelayBetweenMessages: 0, ContinueOnError: true},
		DuplicatePolicy: email.PolicyKeepAll,
	}

	ctx, cancel := a.beginRun()
	defer a.endRun(cancel)

	res := a.runner.Run(ctx, c, a.progressSink())
	switch res.State {
	case campaign.CampaignCompleted:
		if res.Submitted == 0 {
			return fmt.Errorf("test send failed: %s", firstRecipientError(res))
		}
		return nil
	case campaign.CampaignPreflightFailed:
		return fmt.Errorf("test send blocked by preflight: %s", preflightSummary(res.Preflight))
	default:
		return fmt.Errorf("test send %s: %s", res.State, firstNonEmpty(res.FatalError, firstRecipientError(res)))
	}
}

// SendBulkEmails runs a full campaign through the engine and returns a typed
// result. A fatal preflight or Outlook failure is never reported as an empty
// success.
func (a *App) SendBulkEmails(request models.EmailRequest) campaign.CampaignResult {
	c := a.buildCampaign(request, true)

	ctx, cancel := a.beginRun()
	defer a.endRun(cancel)

	res := a.runner.Run(ctx, c, a.progressSink())
	res = a.persistRun(c, res, "", 1)

	a.mu.Lock()
	a.lastResult = &res
	a.lastCampaignID = res.CampaignID
	a.mu.Unlock()
	return res
}

// RetryCampaign reconstructs a campaign from its persisted immutable snapshot,
// performs fresh preflight against the current environment, and persists the retry
// as a new child record. This remains safe after an application restart.
func (a *App) RetryCampaign(campaignID string) campaign.CampaignResult {
	if a.history == nil {
		return campaign.CampaignResult{State: campaign.CampaignPreflightFailed, FatalError: "campaign history is unavailable"}
	}
	rec, err := a.history.Get(campaignID)
	if err != nil {
		return campaign.CampaignResult{State: campaign.CampaignPreflightFailed, FatalError: fmt.Sprintf("load campaign %s: %v", campaignID, err)}
	}
	c := a.campaignFromRecord(*rec)
	if len(c.Contacts) == 0 {
		return campaign.CampaignResult{State: campaign.CampaignPreflightFailed, FatalError: "stored campaign does not contain recipient data"}
	}

	ctx, cancel := a.beginRun()
	defer a.endRun(cancel)

	res := a.runner.Retry(ctx, c, rec.Result, nil, a.progressSink())
	res = a.persistRun(c, res, rec.ID, rec.RunNumber+1)

	a.mu.Lock()
	a.lastResult = &res
	a.lastCampaignID = res.CampaignID
	a.mu.Unlock()
	return res
}

// RetryFailed is retained for Wails/API compatibility. New callers should use
// RetryCampaign with the CampaignID returned by SendBulkEmails.
func (a *App) RetryFailed(request models.EmailRequest) campaign.CampaignResult {
	a.mu.Lock()
	campaignID := a.lastCampaignID
	prev := a.lastResult
	a.mu.Unlock()
	if campaignID != "" {
		return a.RetryCampaign(campaignID)
	}
	if prev == nil {
		return campaign.CampaignResult{State: campaign.CampaignPreflightFailed, FatalError: "no previous campaign to retry"}
	}

	c := a.buildCampaign(request, true)
	ctx, cancel := a.beginRun()
	defer a.endRun(cancel)

	res := a.runner.Retry(ctx, c, *prev, nil, a.progressSink())
	a.mu.Lock()
	a.lastResult = &res
	a.mu.Unlock()
	return res
}

// CancelCampaign cancels the in-flight campaign, if any. Unattempted recipients
// are marked cancelled; completed recipient results are preserved.
func (a *App) CancelCampaign() {
	a.mu.Lock()
	cancel := a.cancel
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// beginRun creates a cancelable context for a campaign and stores its cancel func.
func (a *App) beginRun() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(a.ctx)
	a.mu.Lock()
	a.cancel = cancel
	a.mu.Unlock()
	return ctx, cancel
}

func (a *App) endRun(cancel context.CancelFunc) {
	cancel()
	a.mu.Lock()
	a.cancel = nil
	a.mu.Unlock()
}

// buildCampaign maps a frontend EmailRequest into a campaign, sourcing send
// options from validated settings and the suppression list from its service.
func (a *App) buildCampaign(request models.EmailRequest, applySuppression bool) campaign.Campaign {
	var suppressed map[string]bool
	if applySuppression && a.suppression != nil {
		suppressed = a.suppression.Set()
	}
	return campaign.Campaign{
		Headers:         headersForContacts(request.Contacts),
		Contacts:        request.Contacts,
		SubjectTemplate: request.SubjectTemplate,
		BodyTemplate:    request.BodyTemplate,
		IsHTML:          request.IsHTML,
		DraftOnly:       request.DraftOnly,
		Attachments:     request.Attachments,
		CCTemplate:      firstNonEmpty(request.CCTemplate, request.CC),
		BCCTemplate:     firstNonEmpty(request.BCCTemplate, request.BCC),
		Options:         a.sendOptions(),
		DuplicatePolicy: a.duplicatePolicy(),
		Suppressed:      suppressed,
	}
}

// sendOptions reads pacing/behavior from validated application settings.
func (a *App) sendOptions() campaign.SendOptions {
	opts := campaign.DefaultSendOptions()
	if a.settingsService != nil {
		s := a.settingsService.GetSettings()
		opts.DelayBetweenMessages = time.Duration(s.SendingDelay) * time.Millisecond
		opts.ConfirmBeforeSend = s.ConfirmSend
	}
	return opts.Normalized()
}

func (a *App) duplicatePolicy() email.Policy {
	if a.settingsService != nil {
		if p := email.Policy(a.settingsService.GetSettings().DuplicatePolicy); p.Valid() {
			return p
		}
	}
	return email.DefaultPolicy
}

// progressSink forwards campaign progress to the frontend's "email:progress"
// event channel (unchanged payload shape).
func (a *App) progressSink() campaign.ProgressSink {
	return campaign.ProgressFunc(func(p campaign.Progress) {
		if a.ctx == nil {
			return
		}
		runtime.EventsEmit(a.ctx, "email:progress", models.ProgressUpdate{
			Current:   p.Current,
			Total:     p.Total,
			Status:    p.Status,
			Email:     p.Email,
			Message:   p.Message,
			Timestamp: time.Now().Format(time.RFC3339),
		})
	})
}

// ---- campaign helpers ----

// headersForContacts derives the union of column headers from the contacts'
// custom fields (every contact carries all imported columns), giving the merge
// schema the full field set for unknown-field detection.
func headersForContacts(contacts []models.Contact) []string {
	seen := map[string]bool{}
	var headers []string
	for _, c := range contacts {
		for k := range c.CustomFields {
			if !seen[k] {
				seen[k] = true
				headers = append(headers, k)
			}
		}
	}
	return headers
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func firstRecipientError(res campaign.CampaignResult) string {
	for _, r := range res.RecipientResults {
		if len(r.Attempts) > 0 {
			last := r.Attempts[len(r.Attempts)-1]
			if last.Error != "" {
				return last.Error
			}
		}
	}
	return "unknown error"
}

func preflightSummary(pf *campaign.PreflightResult) string {
	if pf == nil || len(pf.Errors) == 0 {
		return "validation failed"
	}
	msg := pf.Errors[0].Message
	if len(pf.Errors) > 1 {
		msg = fmt.Sprintf("%s (and %d more)", msg, len(pf.Errors)-1)
	}
	return msg
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

// SetVersionInfo records build-time version metadata (called from main).
func (a *App) SetVersionInfo(version, commit, buildDate string) {
	a.version = version
	a.commit = commit
	a.buildDate = buildDate
}

// GetAppInfo returns application metadata for display in the UI, derived from a
// single build-time source rather than hard-coded values.
func (a *App) GetAppInfo() map[string]string {
	return map[string]string{
		"name":      "MailMerge Go",
		"version":   firstNonEmpty(a.version, "dev"),
		"commit":    firstNonEmpty(a.commit, "unknown"),
		"buildDate": firstNonEmpty(a.buildDate, "unknown"),
		"author":    "ajbergh",
	}
}

// ================== Template Operations ==================

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

// ================== Settings Operations ==================

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
