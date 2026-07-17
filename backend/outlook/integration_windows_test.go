//go:build windows && outlookintegration

/*
Opt-in Outlook integration tests. These require classic Outlook installed and a
configured mail account, so they are gated behind the `outlookintegration` build
tag and are never run by default CI.

Run on a Windows workstation with classic Outlook:

	go test -tags outlookintegration ./backend/outlook/...

These tests create Outlook items but prefer draft/inspection over sending real
email. Review each test before enabling on a machine with a live mailbox.
*/
package outlook

import (
	"context"
	"testing"
	"time"

	"MailMergeApp/backend/campaign"
)

func TestIntegrationCapabilityDetection(t *testing.T) {
	s := New()
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	status := s.Preflight(ctx)
	t.Logf("sender status: available=%v state=%s msg=%s", status.Available, status.State, status.Message)
	if !status.Available {
		t.Skipf("Outlook not available on this machine: %s", status.Message)
	}
	caps := s.Capabilities(ctx)
	if !caps.SupportsHTML || !caps.SupportsAttachments {
		t.Errorf("unexpected capabilities: %+v", caps)
	}
}

// TestIntegrationDraftRoundTrip exercises the COM path end to end. It is written
// to create + display a draft rather than send. Enable deliberately.
func TestIntegrationDraftRoundTrip(t *testing.T) {
	t.Skip("enable manually: creates an Outlook item on a machine with a live mailbox")

	s := New()
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	msg := campaign.RenderedMessage{
		To:       []string{"test@example.com"},
		Subject:  "MailMerge integration test",
		IsHTML:   true,
		HTMLBody: "<p>Hello from the integration test.</p>",
	}
	receipt := s.Send(ctx, msg)
	if receipt.Err != nil {
		t.Fatalf("send failed: %v", receipt.Err)
	}
}
