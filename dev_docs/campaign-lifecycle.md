# Campaign Lifecycle

A send goes through the same pipeline whether it is a preview, a test send, or a
bulk send.

```
Build Campaign (contacts, templates, options, suppression, dedupe policy)
        │
        ▼
Preflight ──► errors? ──► CampaignResult{ state: preflight_failed }  (nothing sent)
        │ CanSend
        ▼
Runner.Run(ctx)
   for each kept recipient (post-dedupe, minus suppressed):
      • honor inter-message delay / batch pause (injectable Clock)
      • check ctx cancellation
      • render RenderedMessage
      • sender.Send → SendReceipt
      • record Attempt (submitted | failed); fatal → stop (runtime_failed)
        │
        ▼
CampaignResult{ state, submitted, failed, skipped, cancelled, recipientResults[] }
        │
        ▼
Persist to campaign history (storage repository)
```

## States

| State              | Meaning                                                        |
|--------------------|----------------------------------------------------------------|
| `completed`        | The run finished (some recipients may still have failed).      |
| `cancelled`        | The user cancelled; unattempted recipients marked `cancelled`. |
| `preflight_failed` | Blocked before sending; nothing was submitted.                 |
| `runtime_failed`   | A fatal sender error stopped the run mid-way.                  |

## Counters

Counters are derived from each recipient's **final** status, so a recipient that
failed then succeeded on retry counts once as `submitted` — retries never
double-count. Full attempt history is preserved per recipient.

## Retry

`RetryFailed` re-runs only recipients whose latest attempt failed, appending new
attempts to the existing records.

## Cancellation

`CancelCampaign` cancels the run's `context.Context`. The runner stops before the
next recipient, marks the remainder `cancelled`, and preserves completed results.
