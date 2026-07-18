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
	FakeAlwaysSucceed FakeBehavior = iota
	FakeAlwaysFail
	FakeFatal
	FakeUnavailable
)

// FakeSender is a deterministic EmailSender for tests. It records every message
// it is asked to send and supports selected failing recipients, delayed
// responses, and cancellation.
type FakeSender struct {
	mu sync.Mutex

	Behavior FakeBehavior
	FailFor  map[string]bool
	Delay    time.Duration
	Caps     *SenderCapabilities

	Sent []RenderedMessage
	All  []RenderedMessage

	sendCount int
}

func NewFakeSender() *FakeSender { return &FakeSender{Behavior: FakeAlwaysSucceed} }

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

func (f *FakeSender) Preflight(context.Context) SenderStatus {
	if f.Behavior == FakeUnavailable {
		return SenderStatus{Available: false, State: StateUnavailable, Message: "fake sender is unavailable"}
	}
	return SenderStatus{Available: true, State: StateFake, Message: "fake sender ready"}
}

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
			return SendReceipt{Submitted: false, Kind: SendErrorCancelled, Err: ctx.Err()}
		case <-t.C:
		}
	}
	if err := ctx.Err(); err != nil {
		return SendReceipt{Submitted: false, Kind: SendErrorCancelled, Err: err}
	}

	switch f.Behavior {
	case FakeAlwaysFail:
		return SendReceipt{Submitted: false, Kind: SendErrorRecipient, Err: fmt.Errorf("fake failure")}
	case FakeFatal:
		if n == 1 {
			return SendReceipt{Submitted: false, Kind: SendErrorFatal, Fatal: true, Err: fmt.Errorf("fake fatal error")}
		}
	}

	if f.FailFor != nil {
		for _, to := range msg.To {
			if f.FailFor[strings.ToLower(to)] {
				return SendReceipt{Submitted: false, Kind: SendErrorRecipient, Err: fmt.Errorf("fake failure for %s", to)}
			}
		}
	}

	f.mu.Lock()
	f.Sent = append(f.Sent, msg)
	f.mu.Unlock()
	info := "fake submitted"
	if msg.SaveAsDraft {
		info = "fake draft saved"
	}
	return SendReceipt{Submitted: true, Info: info}
}
