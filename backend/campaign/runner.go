package campaign

import (
	"context"
	"os"
)

// Runner executes campaigns against an EmailSender. It uses context.Context for
// cancellation and an injectable Clock for delays so tests never sleep for real.
type Runner struct {
	sender    EmailSender
	clock     Clock
	preflight *Preflighter
	stat      func(string) (os.FileInfo, error)
}

// RunnerOption customizes a Runner.
type RunnerOption func(*Runner)

// WithClock sets the clock (default RealClock).
func WithClock(c Clock) RunnerOption { return func(r *Runner) { r.clock = c } }

// WithPreflighter sets the preflighter (default NewPreflighter).
func WithPreflighter(p *Preflighter) RunnerOption { return func(r *Runner) { r.preflight = p } }

// withStat overrides filesystem stat (tests).
func withStat(fn func(string) (os.FileInfo, error)) RunnerOption {
	return func(r *Runner) { r.stat = fn }
}

// NewRunner builds a Runner.
func NewRunner(sender EmailSender, opts ...RunnerOption) *Runner {
	r := &Runner{sender: sender, clock: RealClock{}, preflight: NewPreflighter(), stat: os.Stat}
	for _, o := range opts {
		o(r)
	}
	if r.preflight.stat == nil {
		r.preflight.stat = r.stat
	}
	return r
}

// Run preflights and executes a campaign, returning a typed result. A fatal
// preflight or sender failure is never reported as an empty success.
func (r *Runner) Run(ctx context.Context, c Campaign, sink ProgressSink) CampaignResult {
	if sink == nil {
		sink = nopSink{}
	}
	pf := r.preflight.Preflight(ctx, c, r.sender)
	if !pf.CanSend {
		return CampaignResult{State: CampaignPreflightFailed, Preflight: &pf}
	}

	// Seed recipient results in send order.
	results := make([]RecipientResult, 0, len(pf.Recipients))
	for _, idx := range pf.Recipients {
		ct := c.Contacts[idx]
		results = append(results, RecipientResult{
			ContactIndex: idx,
			ContactID:    ct.ID,
			Email:        ct.Email,
			FirstName:    ct.FirstName,
			LastName:     ct.LastName,
		})
	}

	final := r.execute(ctx, c, results, allIndices(len(results)), sink)
	final.Preflight = &pf
	return final
}

// Retry re-sends to recipients whose latest attempt failed (or a provided
// subset of contact indices), appending new attempt records without
// double-counting prior successes.
func (r *Runner) Retry(ctx context.Context, c Campaign, prev CampaignResult, only []int, sink ProgressSink) CampaignResult {
	if sink == nil {
		sink = nopSink{}
	}
	results := make([]RecipientResult, len(prev.RecipientResults))
	copy(results, prev.RecipientResults)

	onlySet := map[int]bool{}
	for _, i := range only {
		onlySet[i] = true
	}

	var targets []int
	for pos, rr := range results {
		if !rr.lastFailed() {
			continue
		}
		if len(onlySet) > 0 && !onlySet[rr.ContactIndex] {
			continue
		}
		targets = append(targets, pos)
	}

	final := r.execute(ctx, c, results, targets, sink)
	final.Preflight = prev.Preflight
	return final
}

// execute sends to the recipient positions in targets, recording attempts.
func (r *Runner) execute(ctx context.Context, c Campaign, results []RecipientResult, targets []int, sink ProgressSink) CampaignResult {
	renderer := NewRenderer(c)
	if r.stat != nil {
		renderer.withStat(r.stat)
	}
	opts := c.Options.Normalized()
	total := len(targets)
	state := CampaignCompleted
	var fatalErr string

	cancelRemaining := func(from int, status RecipientStatus) {
		for j := from; j < len(targets); j++ {
			pos := targets[j]
			r.appendAttempt(&results[pos], status, "")
		}
	}

	for i, pos := range targets {
		if err := ctx.Err(); err != nil {
			cancelRemaining(i, RecipientCancelled)
			state = CampaignCancelled
			break
		}

		// Inter-message delay (skipped before the very first send).
		if i > 0 && opts.DelayBetweenMessages > 0 {
			if err := r.clock.Sleep(ctx, opts.DelayBetweenMessages); err != nil {
				cancelRemaining(i, RecipientCancelled)
				state = CampaignCancelled
				break
			}
		}
		// Batch pause.
		if opts.BatchSize > 0 && i > 0 && i%opts.BatchSize == 0 && opts.PauseBetweenBatches > 0 {
			if err := r.clock.Sleep(ctx, opts.PauseBetweenBatches); err != nil {
				cancelRemaining(i, RecipientCancelled)
				state = CampaignCancelled
				break
			}
		}

		rr := &results[pos]
		sink.Progress(Progress{Current: i + 1, Total: total, Status: "sending", Email: rr.Email})

		msg := renderer.Render(c, c.Contacts[rr.ContactIndex])
		receipt := r.sender.Send(ctx, msg)

		if receipt.Fatal {
			errMsg := receiptError(receipt)
			r.appendAttempt(rr, RecipientFailed, errMsg)
			sink.Progress(Progress{Current: i + 1, Total: total, Status: "failure", Email: rr.Email, Message: errMsg})
			cancelRemaining(i+1, RecipientCancelled)
			state = CampaignRuntimeFailed
			fatalErr = errMsg
			break
		}

		if receipt.Submitted {
			r.appendAttempt(rr, RecipientSubmitted, "")
			sink.Progress(Progress{Current: i + 1, Total: total, Status: "success", Email: rr.Email})
		} else {
			errMsg := receiptError(receipt)
			r.appendAttempt(rr, RecipientFailed, errMsg)
			sink.Progress(Progress{Current: i + 1, Total: total, Status: "failure", Email: rr.Email, Message: errMsg})
			if !opts.ContinueOnError {
				cancelRemaining(i+1, RecipientSkipped)
				break
			}
		}
	}

	out := CampaignResult{State: state, FatalError: fatalErr, RecipientResults: results}
	tally(&out)
	sink.Progress(Progress{Current: total, Total: total, Status: "complete",
		Message: summaryMessage(out)})
	return out
}

func (r *Runner) appendAttempt(rr *RecipientResult, status RecipientStatus, errMsg string) {
	rr.Attempts = append(rr.Attempts, Attempt{
		Number:    len(rr.Attempts) + 1,
		Status:    status,
		Error:     errMsg,
		Timestamp: r.clock.Now(),
	})
	rr.Status = status
}

// tally recomputes counters from final recipient statuses so retries never
// double-count.
func tally(out *CampaignResult) {
	out.Submitted, out.Failed, out.Skipped, out.Cancelled, out.Attempted = 0, 0, 0, 0, 0
	for _, rr := range out.RecipientResults {
		switch rr.Status {
		case RecipientSubmitted:
			out.Submitted++
			out.Attempted++
		case RecipientFailed:
			out.Failed++
			out.Attempted++
		case RecipientSkipped:
			out.Skipped++
		case RecipientCancelled:
			out.Cancelled++
		}
	}
}

func allIndices(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}

func receiptError(r SendReceipt) string {
	if r.Err != nil {
		return r.Err.Error()
	}
	if r.ErrMsg != "" {
		return r.ErrMsg
	}
	return "send failed"
}

func summaryMessage(r CampaignResult) string {
	return string(r.State)
}
