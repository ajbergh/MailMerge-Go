package campaign

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"MailMergeApp/backend/email"
	"MailMergeApp/backend/mergefield"
)

// Preflight issue codes.
const (
	CodeZeroRecipients    = "zero_recipients"
	CodeSenderUnavailable = "sender_unavailable"
	CodeUnknownField      = "unknown_field"
	CodeInvalidSyntax     = "invalid_syntax"
	CodeMissingValue      = "missing_value"
	CodeBlankSubject      = "blank_subject"
	CodeBlankBody         = "blank_body"
	CodeInvalidTo         = "invalid_to"
	CodeInvalidCC         = "invalid_cc"
	CodeInvalidBCC        = "invalid_bcc"
	CodeRecipientInCC     = "recipient_in_cc_bcc"
	CodeMissingAttach     = "missing_attachment"
	CodeAttachIsDir       = "attachment_is_directory"
	CodeOversizeAttach    = "oversized_attachment"
	CodeDuplicateAttach   = "duplicate_attachment"
	CodeUnsupportedFeat   = "unsupported_feature"
	CodeSuppressed        = "suppressed_recipient"
	CodeDuplicateGroup    = "duplicate_recipients"
)

// PreflightIssue is a single problem found during preflight.
type PreflightIssue struct {
	Severity     string `json:"severity"` // "error" | "warning"
	Code         string `json:"code"`
	Message      string `json:"message"`
	ContactIndex *int   `json:"contactIndex,omitempty"`
	Email        string `json:"email,omitempty"`
	FieldID      string `json:"fieldId,omitempty"`
	Count        int    `json:"count,omitempty"` // for aggregated issues
}

// PreflightResult is the structured outcome of a preflight.
type PreflightResult struct {
	CanSend           bool               `json:"canSend"`
	Errors            []PreflightIssue   `json:"errors"`
	Warnings          []PreflightIssue   `json:"warnings"`
	RecipientCount    int                `json:"recipientCount"`
	SuppressedCount   int                `json:"suppressedCount"`
	DuplicateDropped  int                `json:"duplicateDropped"`
	DuplicateGroups   []email.Group      `json:"duplicateGroups,omitempty"`
	EstimatedAttempts int                `json:"estimatedAttempts"`
	EstimatedDuration time.Duration      `json:"estimatedDuration"`
	Capabilities      SenderCapabilities `json:"capabilities"`
	SenderStatus      SenderStatus       `json:"senderStatus"`
	// Recipients is the resolved list of contact indices to send to, in order.
	Recipients []int `json:"recipients"`
}

func (p *PreflightResult) addError(i PreflightIssue) {
	i.Severity = "error"
	p.Errors = append(p.Errors, i)
}

func (p *PreflightResult) addWarning(i PreflightIssue) {
	i.Severity = "warning"
	p.Warnings = append(p.Warnings, i)
}

// Preflighter runs the pre-send validation checklist.
type Preflighter struct {
	// WarnAttachmentBytes triggers a warning when total attachment size exceeds
	// this soft limit (0 = default 20MB).
	WarnAttachmentBytes int64
	// MaxAttachmentBytes is a hard per-message limit (0 = no hard limit).
	MaxAttachmentBytes int64
	stat               func(string) (os.FileInfo, error)
}

// NewPreflighter builds a Preflighter with default limits.
func NewPreflighter() *Preflighter {
	return &Preflighter{
		WarnAttachmentBytes: 20 * 1024 * 1024,
		stat:                os.Stat,
	}
}

// Preflight validates a campaign against a sender and returns a structured
// result. Bulk sending must be blocked while CanSend is false.
func (pf *Preflighter) Preflight(ctx context.Context, c Campaign, sender EmailSender) PreflightResult {
	res := PreflightResult{CanSend: true}
	renderer := NewRenderer(c)
	if pf.stat != nil {
		renderer.withStat(pf.stat)
	}

	// 1. Sender readiness and capabilities.
	res.SenderStatus = sender.Preflight(ctx)
	res.Capabilities = sender.Capabilities(ctx)
	if !res.SenderStatus.Available {
		res.addError(PreflightIssue{Code: CodeSenderUnavailable, Message: senderMessage(res.SenderStatus)})
	}

	// 2. Duplicate resolution (applied before recipient counts).
	emails := make([]string, len(c.Contacts))
	for i, ct := range c.Contacts {
		emails[i] = ct.Email
	}
	dd := email.ResolveDuplicates(emails, c.DuplicatePolicy)
	res.DuplicateGroups = dd.Groups
	res.DuplicateDropped = len(dd.Dropped)
	if len(dd.Groups) > 0 {
		res.addWarning(PreflightIssue{Code: CodeDuplicateGroup, Count: len(dd.Groups),
			Message: fmt.Sprintf("%d duplicate address group(s) detected; policy %q applied", len(dd.Groups), effectivePolicy(c.DuplicatePolicy))})
	}

	// 3. Suppression filtering.
	kept := dd.Keep[:0:0]
	for _, idx := range dd.Keep {
		key := email.NormalizeKey(c.Contacts[idx].Email)
		if c.Suppressed != nil && c.Suppressed[key] {
			res.SuppressedCount++
			continue
		}
		kept = append(kept, idx)
	}
	if res.SuppressedCount > 0 {
		res.addWarning(PreflightIssue{Code: CodeSuppressed, Count: res.SuppressedCount,
			Message: fmt.Sprintf("%d recipient(s) skipped: on suppression list", res.SuppressedCount)})
	}
	res.Recipients = kept
	res.RecipientCount = len(kept)
	res.EstimatedAttempts = len(kept)

	if res.RecipientCount == 0 {
		res.addError(PreflightIssue{Code: CodeZeroRecipients, Message: "no recipients selected to send to"})
	}

	// 4. Template-level validation (once).
	for _, d := range renderer.mf.ValidateTemplate(c.SubjectTemplate) {
		pf.addTemplateDiag(&res, d, "subject")
	}
	for _, d := range renderer.mf.ValidateTemplate(c.BodyTemplate) {
		pf.addTemplateDiag(&res, d, "body")
	}

	// 5. Per-recipient rendering checks (aggregate missing values).
	missingByField := map[string]int{}
	var blankSubject, blankBody int
	for _, idx := range kept {
		msg := renderer.Render(c, c.Contacts[idx])
		i := idx
		if strings.TrimSpace(msg.Subject) == "" {
			blankSubject++
		}
		if strings.TrimSpace(stripTags(msg.Body())) == "" {
			blankBody++
		}
		// Address validation.
		pf.checkAddresses(&res, c, msg, i)
		// Missing values.
		for _, d := range msg.Diagnostics {
			if d.Code == mergefield.CodeMissingValue {
				missingByField[d.FieldID]++
			}
		}
		// Attachments.
		pf.checkAttachments(&res, msg, i)
	}
	if blankSubject > 0 {
		res.addWarning(PreflightIssue{Code: CodeBlankSubject, Count: blankSubject,
			Message: fmt.Sprintf("%d recipient(s) would have a blank subject", blankSubject)})
	}
	if blankBody > 0 {
		res.addError(PreflightIssue{Code: CodeBlankBody, Count: blankBody,
			Message: fmt.Sprintf("%d recipient(s) would have a blank body", blankBody)})
	}
	for field, n := range missingByField {
		res.addWarning(PreflightIssue{Code: CodeMissingValue, FieldID: field, Count: n,
			Message: fmt.Sprintf("field %q is blank for %d recipient(s)", field, n)})
	}

	// 6. Capability gating.
	if c.IsHTML && !res.Capabilities.SupportsHTML {
		res.addError(PreflightIssue{Code: CodeUnsupportedFeat, Message: "selected sender does not support HTML email"})
	}
	if len(c.Attachments) > 0 && !res.Capabilities.SupportsAttachments {
		res.addError(PreflightIssue{Code: CodeUnsupportedFeat, Message: "selected sender does not support attachments"})
	}

	// 7. Estimated duration.
	opts := c.Options.Normalized()
	res.EstimatedDuration = estimateDuration(res.RecipientCount, opts)

	res.CanSend = len(res.Errors) == 0
	return res
}

func (pf *Preflighter) addTemplateDiag(res *PreflightResult, d mergefield.Diagnostic, where string) {
	switch d.Code {
	case mergefield.CodeUnknownField:
		res.addError(PreflightIssue{Code: CodeUnknownField, FieldID: d.Token,
			Message: fmt.Sprintf("%s references unknown merge field %q", where, d.Token)})
	case mergefield.CodeInvalidSyntax:
		res.addError(PreflightIssue{Code: CodeInvalidSyntax, Message: fmt.Sprintf("%s: %s", where, d.Message)})
	case mergefield.CodeCollision:
		res.addError(PreflightIssue{Code: CodeUnknownField, FieldID: d.FieldID,
			Message: fmt.Sprintf("%s references ambiguous merge field %q", where, d.FieldID)})
	}
}

func (pf *Preflighter) checkAddresses(res *PreflightResult, c Campaign, msg RenderedMessage, idx int) {
	i := idx
	// To
	toKeys := map[string]bool{}
	if len(msg.To) == 0 {
		res.addError(PreflightIssue{Code: CodeInvalidTo, ContactIndex: &i, Message: "recipient has no valid To address"})
	}
	for _, a := range msg.To {
		if !email.IsValid(a) {
			res.addError(PreflightIssue{Code: CodeInvalidTo, ContactIndex: &i, Email: a,
				Message: fmt.Sprintf("invalid To address: %s", a)})
		} else {
			toKeys[email.NormalizeKey(a)] = true
		}
	}
	for _, a := range msg.CC {
		if !email.IsValid(a) {
			res.addError(PreflightIssue{Code: CodeInvalidCC, ContactIndex: &i, Email: a,
				Message: fmt.Sprintf("invalid CC address: %s", a)})
		} else if toKeys[email.NormalizeKey(a)] {
			res.addWarning(PreflightIssue{Code: CodeRecipientInCC, ContactIndex: &i, Email: a,
				Message: fmt.Sprintf("%s appears in both To and CC", a)})
		}
	}
	for _, a := range msg.BCC {
		if !email.IsValid(a) {
			res.addError(PreflightIssue{Code: CodeInvalidBCC, ContactIndex: &i, Email: a,
				Message: fmt.Sprintf("invalid BCC address: %s", a)})
		} else if toKeys[email.NormalizeKey(a)] {
			res.addWarning(PreflightIssue{Code: CodeRecipientInCC, ContactIndex: &i, Email: a,
				Message: fmt.Sprintf("%s appears in both To and BCC", a)})
		}
	}
}

func (pf *Preflighter) checkAttachments(res *PreflightResult, msg RenderedMessage, idx int) {
	i := idx
	var total int64
	seen := map[string]bool{}
	for _, a := range msg.Attachments {
		key := strings.ToLower(a.Path)
		if seen[key] {
			res.addWarning(PreflightIssue{Code: CodeDuplicateAttach, ContactIndex: &i, Message: fmt.Sprintf("duplicate attachment: %s", a.Path)})
			continue
		}
		seen[key] = true
		if !a.Exists {
			res.addError(PreflightIssue{Code: CodeMissingAttach, ContactIndex: &i, Message: fmt.Sprintf("attachment not found: %s", a.Path)})
			continue
		}
		if a.IsDir {
			res.addError(PreflightIssue{Code: CodeAttachIsDir, ContactIndex: &i, Message: fmt.Sprintf("attachment is a directory: %s", a.Path)})
			continue
		}
		total += a.Size
	}
	if pf.MaxAttachmentBytes > 0 && total > pf.MaxAttachmentBytes {
		res.addError(PreflightIssue{Code: CodeOversizeAttach, ContactIndex: &i,
			Message: fmt.Sprintf("attachments (%d bytes) exceed hard limit %d", total, pf.MaxAttachmentBytes)})
	} else if pf.WarnAttachmentBytes > 0 && total > pf.WarnAttachmentBytes {
		res.addWarning(PreflightIssue{Code: CodeOversizeAttach, ContactIndex: &i,
			Message: fmt.Sprintf("attachments (%d bytes) exceed recommended limit %d", total, pf.WarnAttachmentBytes)})
	}
}

func estimateDuration(n int, o SendOptions) time.Duration {
	if n <= 0 {
		return 0
	}
	d := time.Duration(n) * o.DelayBetweenMessages
	if o.BatchSize > 0 && o.PauseBetweenBatches > 0 {
		batches := (n - 1) / o.BatchSize
		d += time.Duration(batches) * o.PauseBetweenBatches
	}
	return d
}

func senderMessage(s SenderStatus) string {
	if s.Message != "" {
		return s.Message
	}
	return "email sender is not available"
}

func effectivePolicy(p email.Policy) email.Policy {
	if !p.Valid() {
		return email.DefaultPolicy
	}
	return p
}

// stripTags removes HTML tags for the purpose of blank-body detection.
func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				b.WriteRune(r)
			}
		}
	}
	return strings.ReplaceAll(strings.ReplaceAll(b.String(), "&nbsp;", ""), " ", "")
}
