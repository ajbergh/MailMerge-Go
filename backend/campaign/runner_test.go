package campaign

import (
	"context"
	"sync"
	"testing"
	"time"

	"MailMergeApp/backend/email"
	"MailMergeApp/backend/models"
)

// fakeClock records sleeps instead of actually sleeping.
type fakeClock struct {
	mu     sync.Mutex
	now    time.Time
	sleeps []time.Duration
}

func newFakeClock() *fakeClock { return &fakeClock{now: time.Unix(0, 0)} }

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(time.Millisecond)
	return c.now
}

func (c *fakeClock) Sleep(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	c.sleeps = append(c.sleeps, d)
	c.mu.Unlock()
	return nil
}

func (c *fakeClock) totalSleep() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	var t time.Duration
	for _, s := range c.sleeps {
		t += s
	}
	return t
}

func contacts(emails ...string) []models.Contact {
	out := make([]models.Contact, len(emails))
	for i, e := range emails {
		out[i] = models.Contact{ID: e, FirstName: "F", LastName: "L", Email: e}
	}
	return out
}

func baseCampaign(emails ...string) Campaign {
	return Campaign{
		Contacts:        contacts(emails...),
		SubjectTemplate: "Hi {{first_name}}",
		BodyTemplate:    "Hello {{first_name}}",
		IsHTML:          false,
		Options:         SendOptions{DelayBetweenMessages: 500 * time.Millisecond, ContinueOnError: true},
		DuplicatePolicy: email.PolicyKeepFirst,
	}
}

func newTestRunner(sender EmailSender, clk Clock) *Runner {
	return NewRunner(sender, WithClock(clk))
}

func TestRunCompleteSuccess(t *testing.T) {
	fs := NewFakeSender()
	clk := newFakeClock()
	r := newTestRunner(fs, clk)

	res := r.Run(context.Background(), baseCampaign("a@x.com", "b@x.com", "c@x.com"), nil)

	if res.State != CampaignCompleted {
		t.Fatalf("state = %v, want completed", res.State)
	}
	if res.Submitted != 3 || res.Failed != 0 || res.Attempted != 3 {
		t.Errorf("counts: submitted=%d failed=%d attempted=%d", res.Submitted, res.Failed, res.Attempted)
	}
	if len(fs.Sent) != 3 {
		t.Errorf("fake sent %d, want 3", len(fs.Sent))
	}
	// Delay applied between messages: 2 sleeps of 500ms (not before the first).
	if clk.totalSleep() != time.Second {
		t.Errorf("total sleep = %v, want 1s", clk.totalSleep())
	}
}

func TestRunPartialFailure(t *testing.T) {
	fs := NewFakeSender()
	fs.FailFor = map[string]bool{"b@x.com": true}
	r := newTestRunner(fs, newFakeClock())

	res := r.Run(context.Background(), baseCampaign("a@x.com", "b@x.com", "c@x.com"), nil)
	if res.State != CampaignCompleted {
		t.Fatalf("state = %v", res.State)
	}
	if res.Submitted != 2 || res.Failed != 1 {
		t.Errorf("submitted=%d failed=%d, want 2/1", res.Submitted, res.Failed)
	}
}

func TestRunFatalSenderFailure(t *testing.T) {
	fs := &FakeSender{Behavior: FakeFatal}
	r := newTestRunner(fs, newFakeClock())

	res := r.Run(context.Background(), baseCampaign("a@x.com", "b@x.com"), nil)
	if res.State != CampaignRuntimeFailed {
		t.Fatalf("state = %v, want runtime_failed", res.State)
	}
	if res.FatalError == "" {
		t.Error("expected fatal error message")
	}
	// Fatal must not look like a zero-failure success.
	if res.Submitted != 0 {
		t.Errorf("submitted = %d, want 0", res.Submitted)
	}
	if res.Failed != 1 || res.Cancelled != 1 {
		t.Errorf("failed=%d cancelled=%d, want 1/1", res.Failed, res.Cancelled)
	}
}

func TestRunCancellation(t *testing.T) {
	fs := NewFakeSender()
	clk := newFakeClock()
	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after the first send via a progress sink.
	var n int
	sink := ProgressFunc(func(p Progress) {
		if p.Status == "success" {
			n++
			if n == 1 {
				cancel()
			}
		}
	})
	r := newTestRunner(fs, clk)
	res := r.Run(ctx, baseCampaign("a@x.com", "b@x.com", "c@x.com"), sink)

	if res.State != CampaignCancelled {
		t.Fatalf("state = %v, want cancelled", res.State)
	}
	if res.Submitted != 1 {
		t.Errorf("submitted = %d, want 1", res.Submitted)
	}
	if res.Cancelled != 2 {
		t.Errorf("cancelled = %d, want 2", res.Cancelled)
	}
}

func TestRetryOnlyFailedNoDoubleCount(t *testing.T) {
	fs := NewFakeSender()
	fs.FailFor = map[string]bool{"b@x.com": true}
	clk := newFakeClock()
	r := newTestRunner(fs, clk)

	c := baseCampaign("a@x.com", "b@x.com", "c@x.com")
	first := r.Run(context.Background(), c, nil)
	if first.Failed != 1 {
		t.Fatalf("first run failed = %d", first.Failed)
	}

	// Now let b succeed and retry.
	fs.FailFor = nil
	retried := r.Retry(context.Background(), c, first, nil, nil)

	if retried.Submitted != 3 {
		t.Errorf("after retry submitted = %d, want 3 (no double count)", retried.Submitted)
	}
	if retried.Failed != 0 {
		t.Errorf("after retry failed = %d, want 0", retried.Failed)
	}
	// The successful recipients must not have been re-sent.
	// Only b was retried => one additional Send beyond the original 3.
	if len(fs.All) != 4 {
		t.Errorf("total Send calls = %d, want 4 (3 initial + 1 retry)", len(fs.All))
	}
	// Attempt history preserved for b (failed then submitted).
	for _, rr := range retried.RecipientResults {
		if rr.Email == "b@x.com" {
			if len(rr.Attempts) != 2 {
				t.Errorf("b attempts = %d, want 2", len(rr.Attempts))
			}
		}
	}
}

func TestDuplicateExclusion(t *testing.T) {
	fs := NewFakeSender()
	r := newTestRunner(fs, newFakeClock())
	c := baseCampaign("a@x.com", "A@x.com", "b@x.com")
	res := r.Run(context.Background(), c, nil)
	if res.Submitted != 2 {
		t.Errorf("submitted = %d, want 2 (duplicate keep-first)", res.Submitted)
	}
	if res.Preflight.DuplicateDropped != 1 {
		t.Errorf("duplicate dropped = %d, want 1", res.Preflight.DuplicateDropped)
	}
}

func TestSuppression(t *testing.T) {
	fs := NewFakeSender()
	r := newTestRunner(fs, newFakeClock())
	c := baseCampaign("a@x.com", "b@x.com")
	c.Suppressed = map[string]bool{"b@x.com": true}
	res := r.Run(context.Background(), c, nil)
	if res.Submitted != 1 {
		t.Errorf("submitted = %d, want 1", res.Submitted)
	}
	if res.Preflight.SuppressedCount != 1 {
		t.Errorf("suppressed = %d, want 1", res.Preflight.SuppressedCount)
	}
}

func TestPreflightBlocksUnknownField(t *testing.T) {
	fs := NewFakeSender()
	r := newTestRunner(fs, newFakeClock())
	c := baseCampaign("a@x.com")
	c.BodyTemplate = "Hi {{does_not_exist}}"
	res := r.Run(context.Background(), c, nil)
	if res.State != CampaignPreflightFailed {
		t.Fatalf("state = %v, want preflight_failed", res.State)
	}
	if len(fs.All) != 0 {
		t.Error("nothing should be sent when preflight fails")
	}
}

func TestPreflightBlocksInvalidRecipient(t *testing.T) {
	fs := NewFakeSender()
	r := newTestRunner(fs, newFakeClock())
	c := baseCampaign("not-an-email", "b@x.com")
	res := r.Run(context.Background(), c, nil)
	if res.State != CampaignPreflightFailed {
		t.Fatalf("state = %v, want preflight_failed", res.State)
	}
}

func TestBatchPauses(t *testing.T) {
	fs := NewFakeSender()
	clk := newFakeClock()
	r := newTestRunner(fs, clk)
	c := baseCampaign("a@x.com", "b@x.com", "c@x.com", "d@x.com")
	c.Options = SendOptions{DelayBetweenMessages: 0, BatchSize: 2, PauseBetweenBatches: time.Second, ContinueOnError: true}
	res := r.Run(context.Background(), c, nil)
	if res.Submitted != 4 {
		t.Fatalf("submitted = %d", res.Submitted)
	}
	// Batch pause fires at i==2 (start of the 2nd batch): 1 pause of 1s.
	if clk.totalSleep() != time.Second {
		t.Errorf("total sleep = %v, want 1s", clk.totalSleep())
	}
}

func TestSenderUnavailableFailsPreflight(t *testing.T) {
	fs := &FakeSender{Behavior: FakeUnavailable}
	r := newTestRunner(fs, newFakeClock())
	res := r.Run(context.Background(), baseCampaign("a@x.com"), nil)
	if res.State != CampaignPreflightFailed {
		t.Fatalf("state = %v, want preflight_failed", res.State)
	}
}

func TestContinueOnErrorFalseStops(t *testing.T) {
	fs := NewFakeSender()
	fs.FailFor = map[string]bool{"a@x.com": true}
	r := newTestRunner(fs, newFakeClock())
	c := baseCampaign("a@x.com", "b@x.com", "c@x.com")
	c.Options.ContinueOnError = false
	res := r.Run(context.Background(), c, nil)
	if res.Failed != 1 {
		t.Errorf("failed = %d, want 1", res.Failed)
	}
	if res.Skipped != 2 {
		t.Errorf("skipped = %d, want 2 (stopped after first failure)", res.Skipped)
	}
}

func TestProgressOrdering(t *testing.T) {
	fs := NewFakeSender()
	r := newTestRunner(fs, newFakeClock())
	var order []int
	sink := ProgressFunc(func(p Progress) {
		if p.Status == "sending" {
			order = append(order, p.Current)
		}
	})
	r.Run(context.Background(), baseCampaign("a@x.com", "b@x.com", "c@x.com"), sink)
	for i, v := range order {
		if v != i+1 {
			t.Errorf("progress out of order at %d: %v", i, order)
			break
		}
	}
}
