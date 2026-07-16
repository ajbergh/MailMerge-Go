//go:build !windows

/*
Non-Windows stub for the Outlook COM sender. Classic Outlook COM automation is
Windows-only, so on other platforms the sender reports itself unavailable. This
lets the whole module build and the campaign engine's tests run on Linux/macOS
CI without go-ole or Outlook.
*/
package outlook

import (
	"context"

	"MailMergeApp/backend/campaign"
)

// Sender is a no-op sender on non-Windows platforms.
type Sender struct{}

// New returns a stub sender.
func New() *Sender { return &Sender{} }

// Close is a no-op.
func (s *Sender) Close() {}

// Capabilities reports no support on non-Windows platforms.
func (s *Sender) Capabilities(context.Context) campaign.SenderCapabilities {
	return campaign.SenderCapabilities{}
}

// Preflight always reports unavailable off Windows.
func (s *Sender) Preflight(context.Context) campaign.SenderStatus {
	return campaign.SenderStatus{
		Available: false,
		State:     campaign.StateUnavailable,
		Message:   "Outlook COM automation is only available on Windows.",
	}
}

// Send always fails off Windows.
func (s *Sender) Send(context.Context, campaign.RenderedMessage) campaign.SendReceipt {
	return campaign.SendReceipt{Submitted: false, Fatal: true, ErrMsg: "Outlook is only available on Windows"}
}
