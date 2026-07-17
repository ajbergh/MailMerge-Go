package campaign

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// FakeBehavior selects how a FakeSender responds.
type FakeBehavior int

const (
	// FakeAlwaysSucceed submits every message.
	FakeAlwaysSucceed FakeBehavior = iota
	// FakeAlwaysFail fails every message (non-fatal).
	FakeAlwaysFail
	// FakeFatal fails the first send with a fatal receipt (simulates Outlook
	// dying / creation failure).
	FakeFatal
	// FakeUnavailable reports the sender as unavailable in Preflight.
	FakeUnavailable
)

// FakeSender is a deterministic EmailSender for tests. It records every message
// it is asked to send and supports selected failing recipients, delayed
// responses, and cancellation.
type FakeSender struct {
	mu sync.Mutex

	Behavior FakeBehavior
	// FailFor causes Send to fail for any recipient whose To contains one of
	// these addresses (matched case-insensitively).
	FailFor map[string]bool
	// Delay is applied inside Send (respecting ctx) to simulate slow sends.
	Delay time.Duration
	// Caps overrides reported capabilities (defaults to full support).
	Caps *SenderCapabilities

	Sent []RenderedMessage // messages submitted successfully
	All  []RenderedMessage // every message Send was called with

	sendCount int
}

// NewFakeSender returns an always-succeeding fake sender.
func NewFakeSender() *FakeSender { return &FakeSender{Behavior: FakeAlwaysSucceed} }

// Capabilities reports full capabilities unless overridden.
func (f *FakeSender) Capabilities(context.Context) SenderCapabilities {
	if f.Caps != nil {
		return *f.Caps
	}
	return SenderCapabilities{
		SupportsHTML:             true,
		SupportsAttachments:      true,
		SupportsMultipleAccounts: true,
		SupportsSharedMailbox:    true,
		SupportsDraftOnly:        true,
		SupportsScheduling:       false,
	}
}

// Preflight reports readiness based on Behavior.
func (f *FakeSender) Preflight(context.Context) SenderStatus {
	if f.Behavior == FakeUnavailable {
		return SenderStatus{Available: false, State: StateUnavailable, Message: "fake sender is unavailable"}
	}
	return SenderStatus{Available: true, State: StateFake, Message: "fake sender ready"}
}

// Send records and responds to a message.
func (f *FakeSender) Send(ctx context.Context, msg RenderedMessage) SendReceipt {
	f.mu.Lock()
	f.All = append(f.All, msg)
	f.sendCount++
	n := f.sendCount
	f.mu.Unlock()

	if f.Delay > 0 {
		t := time.NewTimer(f.Delay)
		defer t.Stop()
		select {
		case <-ctx.Done():
			return SendReceipt{Submitted: false, Err: ctx.Err()}
		case <-t.C:
		}
	}
	if err := ctx.Err(); err != nil {
		return SendReceipt{Submitted: false, Err: err}
	}

	switch f.Behavior {
	case FakeAlwaysFail:
		return SendReceipt{Submitted: false, Err: fmt.Errorf("fake failure")}
	case FakeFatal:
		if n == 1 {
			return SendReceipt{Submitted: false, Fatal: true, Err: fmt.Errorf("fake fatal error")}
		}
	}

	if f.FailFor != nil {
		for _, to := range msg.To {
			if f.FailFor[strings.ToLower(to)] {
				return SendReceipt{Submitted: false, Err: fmt.Errorf("fake failure for %s", to)}
			}
		}
	}

	f.mu.Lock()
	f.Sent = append(f.Sent, msg)
	f.mu.Unlock()
	return SendReceipt{Submitted: true, Info: "fake submitted"}
}
