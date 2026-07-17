package campaign

import (
	"context"
	"time"
)

// EmailSender abstracts a concrete delivery mechanism (classic Outlook COM, a
// fake, a draft-only sender, or a future Microsoft Graph sender). Implementations
// must be safe to call from the runner's goroutine.
type EmailSender interface {
	// Capabilities reports what the sender supports.
	Capabilities(ctx context.Context) SenderCapabilities
	// Preflight reports sender readiness (is Outlook available, etc.).
	Preflight(ctx context.Context) SenderStatus
	// Send submits a single rendered message and returns a receipt.
	Send(ctx context.Context, message RenderedMessage) SendReceipt
}

// Clock abstracts time so tests can run without real sleeping.
type Clock interface {
	Now() time.Time
	// Sleep blocks for d or until ctx is cancelled, returning ctx.Err() if so.
	Sleep(ctx context.Context, d time.Duration) error
}

// RealClock uses the wall clock and time.After.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

func (RealClock) Sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// Progress is a live update emitted during a run.
type Progress struct {
	Current int    `json:"current"`
	Total   int    `json:"total"`
	Status  string `json:"status"`
	Email   string `json:"email"`
	Message string `json:"message,omitempty"`
}

// ProgressSink receives progress updates. Implementations must be non-blocking.
type ProgressSink interface {
	Progress(p Progress)
}

// ProgressFunc adapts a function to a ProgressSink.
type ProgressFunc func(Progress)

func (f ProgressFunc) Progress(p Progress) {
	if f != nil {
		f(p)
	}
}

// nopSink discards progress.
type nopSink struct{}

func (nopSink) Progress(Progress) {}
