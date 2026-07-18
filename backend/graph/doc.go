/*
Package graph is a design-stage placeholder for a future Microsoft Graph email
sender. It intentionally contains no Graph SDK dependency yet; it documents the
intended shape and provides a stub that satisfies campaign.EmailSender so the
wiring compiles and can be feature-flagged on later.

See dev_docs/graph-sender-design.md for the full design (authentication, token
storage, Mail.Send vs draft creation, shared mailboxes, submission-vs-delivery
semantics, rate limits / Retry-After, tenant restrictions, personal vs org
accounts).

Key principle: Graph-specific code stays behind campaign.EmailSender. The
campaign runner never imports this package or any Graph SDK, exactly as it never
imports go-ole.
*/
package graph

import (
	"context"
	"errors"

	"MailMergeApp/backend/campaign"
)

// ErrNotImplemented is returned by the stub sender until Graph support lands.
var ErrNotImplemented = errors.New("microsoft Graph sender is not implemented yet")

// Config holds the (future) Graph sender configuration. Tokens are never stored
// in plaintext; TokenStore is expected to use the OS credential store / DPAPI.
type Config struct {
	TenantID    string
	ClientID    string
	Scopes      []string
	FromAddress string // shared mailbox or Send-As address, optional
	DraftOnly   bool
}

// Sender is a placeholder implementing campaign.EmailSender. It reports itself
// unavailable so it can be registered behind a feature flag without affecting
// behavior.
type Sender struct {
	cfg Config
}

// New returns a stub Graph sender.
func New(cfg Config) *Sender { return &Sender{cfg: cfg} }

// Capabilities describes what a Graph sender will support once implemented.
func (s *Sender) Capabilities(context.Context) campaign.SenderCapabilities {
	return campaign.SenderCapabilities{
		SupportsHTML:             true,
		SupportsAttachments:      true,
		SupportsMultipleAccounts: true,
		SupportsSharedMailbox:    true,
		SupportsDraftOnly:        true,
		SupportsScheduling:       false,
	}
}

// Preflight reports unavailable until the implementation lands.
func (s *Sender) Preflight(context.Context) campaign.SenderStatus {
	return campaign.SenderStatus{
		Available: false,
		State:     campaign.StateUnavailable,
		Message:   "Microsoft Graph sending is planned but not yet implemented.",
	}
}

// Send is not implemented yet.
func (s *Sender) Send(context.Context, campaign.RenderedMessage) campaign.SendReceipt {
	return campaign.SendReceipt{Submitted: false, Fatal: true, Err: ErrNotImplemented}
}

var _ campaign.EmailSender = (*Sender)(nil)
