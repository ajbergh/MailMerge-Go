package campaign

import (
	"context"
	"os"
	"time"

	"MailMergeApp/backend/models"
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
	started := r.clock.Now()
	if sink == nil {
		sink = nopSink{}
	}
	pf := r.preflight.Preflight(ctx, c, r.sender)
	if !pf.CanSend {
		return r.withTiming(CampaignResult{State: CampaignPreflightFailed, Preflight: &pf}, started)
	}

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

	final := r.execute(ctx, c, results, allIndices(len(results)), sink, "initial")
	final.Preflight = &pf
	return r.withTiming(final, started)
}

// Retry re-sends recipients whose latest attempt failed. It always performs a
// fresh preflight against only the actual retry target set before sending. This
// catches sender, suppression, attachment, template, and address changes that
// occurred after the original run while preserving historical attempts.
func (r *Runner) Retry(ctx context.Context, c Campaign, prev CampaignResult, only []int, sink ProgressSink) CampaignResult {
	started := r.clock.Now()
	if sink == nil {
		sink = nopSink{}
	}

	results := make([]RecipientResult, len(prev.RecipientResults))
	copy(results, prev.RecipientResults)

	onlySet := map[int]bool{}
	for _, i := range only {
		onlySet[i] = true
	}

	candidatePositions := make([]int, 0)
	retryContacts := make([]models.Contact, 0)
	for pos, rr := range results {
		if !rr.lastFailed() {
			continue
		}
		if len(onlySet) > 0 && !onlySet[rr.ContactIndex] {
			continue
		}
		if rr.ContactIndex < 0 || rr.ContactIndex >= len(c.Contacts) {
			continue
		}
		candidatePositions = append(candidatePositions, pos)
		retryContacts = append(retryContacts, c.Contacts[rr.ContactIndex])
	}

	if len(candidatePositions) == 0 {
		out := CampaignResult{
			State:            CampaignPreflightFailed,
			FatalError:       "no failed recipients are eligible for retry",
			RecipientResults: results,
		}
		tally(&out)
		return r.withTiming(out, started)
	}

	retryCampaign := c
	retryCampaign.Contacts = retryContacts
	pf := r.preflight.Preflight(ctx, retryCampaign, r.sender)
	if !pf.CanSend {
		out := CampaignResult{
			State:            CampaignPreflightFailed,
			RecipientResults: results,
			Preflight:        &pf,
		}
		tally(&out)
		return r.withTiming(out, started)
	}

	targets := make([]int, 0, len(pf.Recipients))
	for _, retryIdx := range pf.Recipients {
		if retryIdx >= 0 && retryIdx < len(candidatePositions) {
			targets = append(targets, candidatePositions[retryIdx])
		}
	}

	final := r.execute(ctx, c, results, targets, sink, "retry")
	final.Preflight = &pf
	return r.withTiming(final, started)
}

// execute sends to the recipient positions in targets, recording attempts.
func (r *Runner) execute(ctx context.Context, c Campaign, results []RecipientResult, targets []int, sink ProgressSink, trigger string) CampaignResult {
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
			r.appendAttempt(&results[pos], status, "", trigger)
		}
	}

	for i, pos := range targets {
		if err := ctx.Err(); err != nil {
			cancelRemaining(i, RecipientCancelled)
			state = CampaignCancelled
			break
		}

		if i > 0 && opts.DelayBetweenMessages > 0 {
			if err := r.clock.Sleep(ctx, opts.DelayBetweenMessages); err != nil {
				cancelRemaining(i, RecipientCancelled)
				state = CampaignCancelled
				break
			}
		}
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

		if receipt.Kind == SendErrorCancelled {
			r.appendAttempt(rr, RecipientCancelled, receiptError(receipt), trigger)
			cancelRemaining(i+1, RecipientCancelled)
			state = CampaignCancelled
			break
		}

		if receipt.Fatal || receipt.Kind == SendErrorFatal {
			errMsg := receiptError(receipt)
			r.appendAttempt(rr, RecipientFailed, errMsg, trigger)
			sink.Progress(Progress{Current: i + 1, Total: total, Status: "failure", Email: rr.Email, Message: errMsg})
			cancelRemaining(i+1, RecipientCancelled)
			state = CampaignRuntimeFailed
			fatalErr = errMsg
			break
		}

		if receipt.Submitted {
			r.appendAttempt(rr, RecipientSubmitted, "", trigger)
			sink.Progress(Progress{Current: i + 1, Total: total, Status: "success", Email: rr.Email})
		} else {
			errMsg := receiptError(receipt)
			r.appendAttempt(rr, RecipientFailed, errMsg, trigger)
			sink.Progress(Progress{Current: i + 1, Total: total, Status: "failure", Email: rr.Email, Message: errMsg})
			if !opts.ContinueOnError {
				cancelRemaining(i+1, RecipientSkipped)
				state = CampaignStoppedOnFailure
				break
			}
		}
	}

	out := CampaignResult{State: state, FatalError: fatalErr, RecipientResults: results}
	tally(&out)
	if out.State == CampaignCompleted && out.Failed > 0 {
		out.State = CampaignCompletedWithFailures
	}
	sink.Progress(Progress{Current: total, Total: total, Status: "complete", Message: summaryMessage(out)})
	return out
}

func (r *Runner) appendAttempt(rr *RecipientResult, status RecipientStatus, errMsg, trigger string) {
	rr.Attempts = append(rr.Attempts, Attempt{
		Number:    len(rr.Attempts) + 1,
		Status:    status,
		Error:     errMsg,
		Trigger:   trigger,
		Timestamp: r.clock.Now(),
	})
	rr.Status = status
}

func (r *Runner) withTiming(out CampaignResult, started time.Time) CampaignResult {
	finished := r.clock.Now()
	out.StartedAt = started
	out.FinishedAt = finished
	out.Duration = finished.Sub(started)
	return out
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
