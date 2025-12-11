/*
Outlook Service - Microsoft Outlook COM Automation

This service handles sending emails through Microsoft Outlook using Windows
COM (Component Object Model) automation. It provides functionality for:
  - Verifying Outlook installation
  - Sending test emails
  - Sending bulk emails with progress tracking

Threading Requirements:

	COM objects must be accessed from the same OS thread that created them.
	This service uses runtime.LockOSThread() to ensure thread affinity for
	all COM operations. This is critical to prevent the error:
	"The application called an interface that was marshalled for a different thread"

Rate Limiting:

	A 500ms delay is inserted between emails to prevent overwhelming Outlook
	and to maintain COM stability during bulk operations.

Events:

	The service emits "email:progress" events for real-time UI updates during
	bulk send operations.

Phase 3 Updates (v1.4):
  - Added CC and BCC support for all email sending methods
  - CC/BCC can be static strings or template strings with merge fields
*/
package services

import (
	"MailMergeApp/backend/models"
	"context"
	"fmt"
	"os"
	goruntime "runtime"
	"sync"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// OutlookService handles Outlook COM automation for sending emails.
// All public methods are thread-safe and handle COM initialization internally.
type OutlookService struct {
	ctx          context.Context // Wails context for event emission
	mergeService *MergeService   // Template merge service
	mu           sync.Mutex      // Mutex to prevent concurrent COM operations
}

// NewOutlookService creates a new OutlookService instance with an
// internal MergeService for template rendering.
func NewOutlookService() *OutlookService {
	return &OutlookService{
		mergeService: NewMergeService(),
	}
}

// SetContext sets the Wails context for event emission.
// Must be called during app startup before sending emails.
func (s *OutlookService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// CheckOutlookInstalled verifies that Microsoft Outlook is installed and
// accessible via COM automation. Uses STA threading model.
//
// This method should be called on application startup to verify the
// environment before attempting to send emails.
//
// Returns:
//   - error: Non-nil if Outlook is not installed or not accessible
func (s *OutlookService) CheckOutlookInstalled() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Lock this goroutine to the current OS thread for COM operations
	goruntime.LockOSThread()
	defer goruntime.UnlockOSThread()

	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err != nil {
		// Already initialized is OK
		if err.(*ole.OleError).Code() != 0x80010106 && err.(*ole.OleError).Code() != 1 {
			return fmt.Errorf("failed to initialize COM: %w", err)
		}
	}
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("Outlook.Application")
	if err != nil {
		return fmt.Errorf("outlook is not installed or not accessible: %w", err)
	}
	defer unknown.Release()

	return nil
}

// SendTestEmail sends a single test email to verify configuration.
// The email is sent immediately through Outlook using COM automation.
// Phase 3: Now supports CC and BCC recipients.
//
// Parameters:
//   - testAddress: Recipient email address
//   - subject: Email subject (already rendered, no merge fields)
//   - body: Email body (already rendered, no merge fields)
//   - isHTML: True to send as HTML email, false for plain text
//   - attachments: Array of absolute file paths to attach
//   - cc: CC recipients (comma-separated email addresses)
//   - bcc: BCC recipients (comma-separated email addresses)
//
// Returns:
//   - error: Non-nil if send failed
func (s *OutlookService) SendTestEmail(testAddress, subject, body string, isHTML bool, attachments []string, cc, bcc string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Lock this goroutine to the current OS thread for COM operations
	goruntime.LockOSThread()
	defer goruntime.UnlockOSThread()

	// Initialize COM
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err != nil {
		oleErr, ok := err.(*ole.OleError)
		if !ok || (oleErr.Code() != 0x80010106 && oleErr.Code() != 1) {
			return fmt.Errorf("failed to initialize COM: %w", err)
		}
	}
	defer ole.CoUninitialize()

	// Create Outlook application
	outlook, err := s.createOutlookApp()
	if err != nil {
		return err
	}
	defer outlook.Release()

	// Send the email
	err = s.sendSingleEmail(outlook, testAddress, subject, body, isHTML, attachments, cc, bcc)
	if err != nil {
		return fmt.Errorf("failed to send test email: %w", err)
	}

	return nil
}

// SendBulkEmails sends personalized emails to all contacts with real-time progress updates.
// Each email is individually rendered using the merge service before sending.
// Phase 3: Now supports CC and BCC with optional merge field templates.
//
// Progress events are emitted for each email with status updates. The operation
// continues even if individual emails fail, collecting errors in the result.
//
// Threading: Uses LockOSThread to ensure COM thread affinity throughout the operation.
// Rate limiting: 500ms delay between emails to maintain Outlook/COM stability.
//
// Parameters:
//   - contacts: List of recipients to send to
//   - subjectTemplate: Email subject with merge fields like {FirstName}
//   - bodyTemplate: Email body with merge fields
//   - isHTML: True for HTML email, false for plain text
//   - attachments: Array of absolute file paths to attach to each email
//   - cc: Static CC recipients (same for all emails)
//   - bcc: Static BCC recipients (same for all emails)
//   - ccTemplate: CC with merge fields (personalized per contact)
//   - bccTemplate: BCC with merge fields (personalized per contact)
//
// Returns:
//   - *models.SendResult: Contains success/failure counts and detailed logs
func (s *OutlookService) SendBulkEmails(contacts []models.Contact, subjectTemplate, bodyTemplate string, isHTML bool, attachments []string, cc, bcc, ccTemplate, bccTemplate string) *models.SendResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Lock this goroutine to the current OS thread for the entire duration
	// This is CRITICAL for COM operations which require thread affinity
	goruntime.LockOSThread()
	defer goruntime.UnlockOSThread()

	result := &models.SendResult{
		Logs: []models.EmailLog{},
	}

	total := len(contacts)
	if total == 0 {
		return result
	}

	// Validate attachments first
	for _, attachment := range attachments {
		if _, err := os.Stat(attachment); os.IsNotExist(err) {
			// Emit error event
			s.emitProgress(0, total, "failure", "", fmt.Sprintf("Attachment not found: %s", attachment))
			return result
		}
	}

	// Initialize COM
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err != nil {
		oleErr, ok := err.(*ole.OleError)
		if !ok || (oleErr.Code() != 0x80010106 && oleErr.Code() != 1) {
			s.emitProgress(0, total, "failure", "", fmt.Sprintf("Failed to initialize COM: %v", err))
			return result
		}
	}
	defer ole.CoUninitialize()

	// Create Outlook application
	outlook, err := s.createOutlookApp()
	if err != nil {
		s.emitProgress(0, total, "failure", "", fmt.Sprintf("Failed to create Outlook instance: %v", err))
		return result
	}
	defer outlook.Release()

	// Send to each contact
	for i, contact := range contacts {
		// Emit sending status
		s.emitProgress(i+1, total, "sending", contact.Email, "")

		// Render templates
		merged := s.mergeService.RenderForContact(subjectTemplate, bodyTemplate, contact)

		// Render CC/BCC templates if provided
		finalCC := cc
		finalBCC := bcc
		if ccTemplate != "" {
			ccResult := s.mergeService.RenderForContact("", ccTemplate, contact)
			finalCC = ccResult.Body // Body contains the rendered template
		}
		if bccTemplate != "" {
			bccResult := s.mergeService.RenderForContact("", bccTemplate, contact)
			finalBCC = bccResult.Body
		}

		// Create log entry
		log := models.EmailLog{
			FirstName: contact.FirstName,
			LastName:  contact.LastName,
			Email:     contact.Email,
			Timestamp: time.Now(),
		}

		// Send email
		err := s.sendSingleEmail(outlook, contact.Email, merged.Subject, merged.Body, isHTML, attachments, finalCC, finalBCC)
		if err != nil {
			log.Status = "Failure"
			log.ErrorMessage = err.Error()
			result.TotalFailed++
			// Phase 1: Track failed contacts for retry functionality
			result.FailedContacts = append(result.FailedContacts, contact)
			s.emitProgress(i+1, total, "failure", contact.Email, err.Error())
		} else {
			log.Status = "Success"
			result.TotalSent++
			s.emitProgress(i+1, total, "success", contact.Email, "")
		}

		result.Logs = append(result.Logs, log)

		// Delay between emails to give Outlook time to process
		// Also helps prevent rate limiting and COM stability issues
		time.Sleep(500 * time.Millisecond)
	}

	// Emit completion
	s.emitProgress(total, total, "complete", "", fmt.Sprintf("Completed: %d sent, %d failed", result.TotalSent, result.TotalFailed))

	return result
}

// createOutlookApp creates and returns an Outlook Application COM object.
// The caller is responsible for calling Release() on the returned object.
//
// Returns:
//   - *ole.IDispatch: The Outlook.Application COM object
//   - error: Non-nil if Outlook could not be instantiated
func (s *OutlookService) createOutlookApp() (*ole.IDispatch, error) {
	unknown, err := oleutil.CreateObject("Outlook.Application")
	if err != nil {
		return nil, fmt.Errorf("failed to create Outlook application: %w", err)
	}

	outlook, err := unknown.QueryInterface(ole.IID_IDispatch)
	unknown.Release()
	if err != nil {
		return nil, fmt.Errorf("failed to get Outlook interface: %w", err)
	}

	return outlook, nil
}

// sendSingleEmail creates and sends a single email using the Outlook COM interface.
// This is an internal method used by both SendTestEmail and SendBulkEmails.
// Phase 3: Now supports CC and BCC recipients.
//
// Parameters:
//   - outlook: Active Outlook.Application COM object
//   - to: Recipient email address
//   - subject: Email subject (already rendered)
//   - body: Email body (already rendered)
//   - isHTML: True to set HTMLBody, false for plain text Body
//   - attachments: Array of file paths to attach
//   - cc: CC recipients (comma-separated, can be empty)
//   - bcc: BCC recipients (comma-separated, can be empty)
//
// Returns:
//   - error: Non-nil if any step of email creation or sending fails
func (s *OutlookService) sendSingleEmail(outlook *ole.IDispatch, to, subject, body string, isHTML bool, attachments []string, cc, bcc string) error {
	// Create MailItem (olMailItem = 0)
	mailItemVariant, err := oleutil.CallMethod(outlook, "CreateItem", 0)
	if err != nil {
		return fmt.Errorf("failed to create mail item: %w", err)
	}

	mailItem := mailItemVariant.ToIDispatch()
	defer mailItem.Release()

	// Set To
	_, err = oleutil.PutProperty(mailItem, "To", to)
	if err != nil {
		return fmt.Errorf("failed to set To: %w", err)
	}

	// Set CC (if provided)
	if cc != "" {
		_, err = oleutil.PutProperty(mailItem, "CC", cc)
		if err != nil {
			return fmt.Errorf("failed to set CC: %w", err)
		}
	}

	// Set BCC (if provided)
	if bcc != "" {
		_, err = oleutil.PutProperty(mailItem, "BCC", bcc)
		if err != nil {
			return fmt.Errorf("failed to set BCC: %w", err)
		}
	}

	// Set Subject
	_, err = oleutil.PutProperty(mailItem, "Subject", subject)
	if err != nil {
		return fmt.Errorf("failed to set Subject: %w", err)
	}

	// Set Body (HTML or plain text)
	if isHTML {
		// Normalize HTML for email clients to fix spacing issues from ReactQuill
		// This adds inline styles to <p> tags to remove default margins
		normalizedBody := s.mergeService.NormalizeHTMLForEmail(body)
		_, err = oleutil.PutProperty(mailItem, "HTMLBody", normalizedBody)
	} else {
		_, err = oleutil.PutProperty(mailItem, "Body", body)
	}
	if err != nil {
		return fmt.Errorf("failed to set body: %w", err)
	}

	// Add attachments
	for _, attachmentPath := range attachments {
		// Verify file exists
		if _, err := s.validateAttachment(attachmentPath); err != nil {
			return err
		}

		attachmentsVariant, err := oleutil.GetProperty(mailItem, "Attachments")
		if err != nil {
			return fmt.Errorf("failed to get attachments collection: %w", err)
		}
		attachmentsObj := attachmentsVariant.ToIDispatch()

		_, err = oleutil.CallMethod(attachmentsObj, "Add", attachmentPath)
		attachmentsObj.Release()
		if err != nil {
			return fmt.Errorf("failed to add attachment %s: %w", attachmentPath, err)
		}
	}

	// Send the email
	_, err = oleutil.CallMethod(mailItem, "Send")
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// validateAttachment checks if an attachment file exists and is accessible.
// Used to validate attachment paths before attempting to send.
//
// Parameters:
//   - path: Absolute path to the attachment file
//
// Returns:
//   - bool: True if file exists and is valid
//   - error: Description of why validation failed
func (s *OutlookService) validateAttachment(path string) (bool, error) {
	info, err := osStat(path)
	if osIsNotExist(err) {
		return false, fmt.Errorf("attachment not found: %s", path)
	}
	if err != nil {
		return false, fmt.Errorf("failed to access attachment: %w", err)
	}
	if info.IsDir() {
		return false, fmt.Errorf("attachment path is a directory: %s", path)
	}
	return true, nil
}

// OS functions (for easier testing - allows mocking in unit tests)
var osStat = os.Stat
var osIsNotExist = os.IsNotExist

// emitProgress sends a progress update event to the frontend.
// The event is emitted on the "email:progress" channel with a ProgressUpdate payload.
//
// Parameters:
//   - current: Current email number (1-based)
//   - total: Total number of emails to send
//   - status: Status string ("sending", "success", "failure", "complete")
//   - email: Email address being processed
//   - message: Additional message (typically error details)
func (s *OutlookService) emitProgress(current, total int, status, email, message string) {
	if s.ctx == nil {
		return
	}

	update := models.ProgressUpdate{
		Current:   current,
		Total:     total,
		Status:    status,
		Email:     email,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	runtime.EventsEmit(s.ctx, "email:progress", update)
}
