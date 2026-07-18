/*
Package campaign contains the platform-independent campaign engine for
MailMerge-Go: the sender abstraction, the render/preflight/run pipeline, typed
results, and a deterministic FakeSender for tests.

No code in this package imports go-ole or any Outlook COM library; the Outlook
implementation lives behind the EmailSender interface (see backend/outlook), so
the campaign runner is fully testable without Outlook installed.
*/
package campaign

import (
	"time"

	"MailMergeApp/backend/email"
	"MailMergeApp/backend/mergefield"
	"MailMergeApp/backend/models"
)

// SendOptions controls pacing and error behavior for a campaign. Every field is
// read from validated application settings; nothing is hard-coded on the send
// path.
type SendOptions struct {
	DelayBetweenMessages time.Duration `json:"delayBetweenMessages"`
	BatchSize            int           `json:"batchSize"`
	PauseBetweenBatches  time.Duration `json:"pauseBetweenBatches"`
	ConfirmBeforeSend    bool          `json:"confirmBeforeSend"`
	ContinueOnError      bool          `json:"continueOnError"`
}

// Bounds for validated options.
const (
	MinDelay = 0 * time.Millisecond
	MaxDelay = 60 * time.Second
	MaxBatch = 100000
)

// DefaultSendOptions returns safe defaults.
func DefaultSendOptions() SendOptions {
	return SendOptions{
		DelayBetweenMessages: 500 * time.Millisecond,
		BatchSize:            0,
		PauseBetweenBatches:  0,
		ConfirmBeforeSend:    true,
		ContinueOnError:      true,
	}
}

// Normalized returns a copy with values clamped to supported bounds.
func (o SendOptions) Normalized() SendOptions {
	if o.DelayBetweenMessages < MinDelay {
		o.DelayBetweenMessages = MinDelay
	}
	if o.DelayBetweenMessages > MaxDelay {
		o.DelayBetweenMessages = MaxDelay
	}
	if o.BatchSize < 0 {
		o.BatchSize = 0
	}
	if o.BatchSize > MaxBatch {
		o.BatchSize = MaxBatch
	}
	if o.PauseBetweenBatches < 0 {
		o.PauseBetweenBatches = 0
	}
	if o.PauseBetweenBatches > MaxDelay {
		o.PauseBetweenBatches = MaxDelay
	}
	return o
}

// Campaign is a fully-specified merge job.
type Campaign struct {
	Headers         []string
	Contacts        []models.Contact
	SubjectTemplate string
	BodyTemplate    string
	IsHTML          bool
	Attachments     []string
	CCTemplate      string
	BCCTemplate     string
	ToOverride      string
	OverrideEmail   bool
	DraftOnly       bool
	Options         SendOptions
	DuplicatePolicy email.Policy
	Suppressed      map[string]bool
}

// Schema builds the merge-field schema for this campaign.
func (c Campaign) Schema() *mergefield.Schema { return mergefield.NewSchema(c.Headers) }

// ResolvedAttachment is an attachment path after per-contact rendering and
// filesystem validation.
type ResolvedAttachment struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Exists bool   `json:"exists"`
	IsDir  bool   `json:"isDir"`
	Error  string `json:"error,omitempty"`
}

// RenderedMessage is the fully-rendered email for one recipient. The same
// RenderedMessage type is produced for preview, test send, and bulk send.
type RenderedMessage struct {
	To          []string                `json:"to"`
	CC          []string                `json:"cc"`
	BCC         []string                `json:"bcc"`
	Subject     string                  `json:"subject"`
	HTMLBody    string                  `json:"htmlBody"`
	TextBody    string                  `json:"textBody"`
	IsHTML      bool                    `json:"isHTML"`
	SaveAsDraft bool                    `json:"saveAsDraft"`
	Attachments []ResolvedAttachment    `json:"attachments"`
	Diagnostics []mergefield.Diagnostic `json:"diagnostics"`
}

// Body returns the body appropriate for IsHTML.
func (m RenderedMessage) Body() string {
	if m.IsHTML {
		return m.HTMLBody
	}
	return m.TextBody
}

// SenderCapabilities describes what a concrete sender supports.
type SenderCapabilities struct {
	SupportsHTML             bool `json:"supportsHTML"`
	SupportsAttachments      bool `json:"supportsAttachments"`
	SupportsMultipleAccounts bool `json:"supportsMultipleAccounts"`
	SupportsSharedMailbox    bool `json:"supportsSharedMailbox"`
	SupportsDraftOnly        bool `json:"supportsDraftOnly"`
	SupportsScheduling       bool `json:"supportsScheduling"`
}

// SenderState enumerates concrete readiness states for actionable diagnostics.
type SenderState string

const (
	StateClassicOutlook  SenderState = "classic_outlook"
	StateNewOutlook      SenderState = "new_outlook"
	StateUnavailable     SenderState = "unavailable"
	StateCOMInaccessible SenderState = "com_inaccessible"
	StateBlockedByPolicy SenderState = "blocked_by_policy"
	StateNoAccount       SenderState = "no_account"
	StateFake            SenderState = "fake"
)

// SenderStatus is a sender readiness report.
type SenderStatus struct {
	Available bool        `json:"available"`
	State     SenderState `json:"state"`
	Message   string      `json:"message"`
}

// SendErrorKind classifies a sender failure so the runner can make a safe
// campaign-level decision without depending on sender-specific error strings.
type SendErrorKind string

const (
	SendErrorRecipient SendErrorKind = "recipient"
	SendErrorTransient SendErrorKind = "transient"
	SendErrorFatal     SendErrorKind = "fatal"
	SendErrorCancelled SendErrorKind = "cancelled"
)

// SendReceipt is the outcome of a single Send call.
type SendReceipt struct {
	Submitted bool          `json:"submitted"`
	Kind      SendErrorKind `json:"kind,omitempty"`
	Fatal     bool          `json:"fatal"`
	Info      string        `json:"info,omitempty"`
	Err       error         `json:"-"`
	ErrMsg    string        `json:"error,omitempty"`
}

// CampaignState is the terminal state of a campaign run.
type CampaignState string

const (
	CampaignCompleted             CampaignState = "completed"
	CampaignCompletedWithFailures CampaignState = "completed_with_failures"
	CampaignStoppedOnFailure      CampaignState = "stopped_on_failure"
	CampaignCancelled             CampaignState = "cancelled"
	CampaignPreflightFailed       CampaignState = "preflight_failed"
	CampaignRuntimeFailed         CampaignState = "runtime_failed"
)

// RecipientStatus is the status of a single recipient attempt.
type RecipientStatus string

const (
	RecipientSubmitted RecipientStatus = "submitted"
	RecipientFailed    RecipientStatus = "failed"
	RecipientSkipped   RecipientStatus = "skipped"
	RecipientCancelled RecipientStatus = "cancelled"
)

// Attempt records one send attempt for a recipient.
type Attempt struct {
	Number    int             `json:"number"`
	Status    RecipientStatus `json:"status"`
	Error     string          `json:"error,omitempty"`
	Trigger   string          `json:"trigger,omitempty"`
	Timestamp time.Time       `json:"timestamp" ts_type:"string"`
}

// RecipientResult is the full history for a single recipient. Attempt history is
// preserved rather than flattened into ambiguous counters.
type RecipientResult struct {
	ContactIndex int             `json:"contactIndex"`
	ContactID    string          `json:"contactId,omitempty"`
	Email        string          `json:"email"`
	FirstName    string          `json:"firstName,omitempty"`
	LastName     string          `json:"lastName,omitempty"`
	Status       RecipientStatus `json:"status"`
	Attempts     []Attempt       `json:"attempts"`
}

func (r RecipientResult) lastFailed() bool { return r.Status == RecipientFailed }

// CampaignResult is the typed outcome of a campaign run. A fatal preflight or
// Outlook failure is never reported as an empty successful result.
type CampaignResult struct {
	CampaignID       string            `json:"campaignId,omitempty"`
	ParentCampaignID string            `json:"parentCampaignId,omitempty"`
	RunNumber        int               `json:"runNumber,omitempty"`
	State            CampaignState     `json:"state"`
	StartedAt        time.Time         `json:"startedAt" ts_type:"string"`
	FinishedAt       time.Time         `json:"finishedAt" ts_type:"string"`
	Duration         time.Duration     `json:"duration"`
	Attempted        int               `json:"attempted"`
	Submitted        int               `json:"submitted"`
	Failed           int               `json:"failed"`
	Skipped          int               `json:"skipped"`
	Cancelled        int               `json:"cancelled"`
	FatalError       string            `json:"fatalError,omitempty"`
	RecipientResults []RecipientResult `json:"recipientResults"`
	Preflight        *PreflightResult  `json:"preflight,omitempty"`
}
