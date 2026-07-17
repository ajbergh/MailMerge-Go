# Microsoft Graph Sender — Design

Status: **design only** (not implemented). Tracked as a future phase; classic
Outlook COM remains the default and only shipping sender.

## Goal

Add a `MicrosoftGraphSender` that implements the existing
`campaign.EmailSender` interface so it can be selected as an alternative to the
classic Outlook COM sender without changing the campaign engine. Graph-specific
code stays out of the campaign runner (the same rule that keeps go-ole out of
it today).

## Interface reuse

`campaign.EmailSender` already models everything the runner needs:

```go
Capabilities(ctx) SenderCapabilities
Preflight(ctx) SenderStatus
Send(ctx, RenderedMessage) SendReceipt
```

The stub lives in `backend/graph` and compiles today; it reports itself
unavailable and returns `ErrNotImplemented` from `Send`.

## Authentication

- **Delegated auth** via MSAL (Authorization Code + PKCE) for interactive users;
  optionally client-credentials for unattended/service scenarios.
- Scopes: `Mail.Send`, `Mail.ReadWrite` (drafts), and `offline_access` for
  refresh tokens.
- **Personal vs organizational accounts**: personal Microsoft accounts support a
  reduced surface; detect the account type and disable unsupported features
  (shared mailbox, Send-As) accordingly.

## Token storage

- Never store tokens in plaintext. Use the OS credential store (Windows
  Credential Manager / DPAPI) via a `TokenStore` abstraction.
- Cache refresh tokens; acquire access tokens silently; fall back to interactive
  only when needed.

## Sending

- `POST /me/sendMail` (or `/users/{id}/sendMail` for Send-As / shared mailbox).
- **Draft-first mode**: create a draft (`POST /me/messages`), attach files, then
  send — mirrors the safer `DraftOnlySender` idea and enables review.
- Attachments: inline (<3 MB) vs large-attachment upload sessions.

## Submission vs delivery

Like COM, Graph confirms **submission**, not delivery. `SendReceipt.Submitted`
must keep meaning "accepted for sending", and the UI must not claim confirmed
delivery.

## Rate limits & retries

- Honor `429` + `Retry-After`; exponential backoff with jitter.
- Respect per-mailbox Graph throttling limits; surface an actionable status when
  throttled. The campaign runner's existing pacing (`SendOptions`) still applies.

## Tenant restrictions

- Admins can disable app access or restrict `Mail.Send`. `Preflight` must return
  an actionable `SenderStatus` (e.g. `blocked_by_policy`) rather than a generic
  failure.

## Testing

- Unit-test request building and Retry-After handling with a fake HTTP client.
- Integration tests behind a build tag using a dedicated test tenant; never send
  to real recipients by default (prefer draft creation).
- Ship behind a feature flag with auth, security, and integration tests before
  enabling.
