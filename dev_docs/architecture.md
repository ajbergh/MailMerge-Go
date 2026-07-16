# Architecture

MailMerge Go is a Wails v2 desktop app: a Go backend bound to a React/TypeScript
frontend. The backend is layered so the campaign engine is platform-independent
and testable without Outlook.

```
UI (React) ── Wails bindings (app.go)
                    │
        Campaign application service (app.go)
                    │
   ┌────────────────┴─────────────────────────────┐
   │ mergefield (canonical fields + renderer)      │
   │ email (validation + duplicate policy)         │
   │ htmlutil (normalize + sanitize)               │
   │ campaign (preflight, runner, typed results)   │
   │ services (import, templates, settings, supp.) │
   │ storage (campaign history repository)         │
   └────────────────┬─────────────────────────────┘
                    │
            campaign.EmailSender (interface)
                    │
   ┌────────────────┼───────────────────────────────┐
   │ outlook.Sender  │ campaign.FakeSender │ graph.Sender │
   │ (classic COM,   │ (tests)             │ (design stub)│
   │  STA worker)    │                     │              │
   └─────────────────┴─────────────────────┴──────────────┘
```

## Key decisions

- **Sender abstraction (`campaign.EmailSender`).** The runner depends only on the
  interface. `go-ole` lives solely in `backend/outlook` (Windows-tagged); a
  non-Windows stub keeps the module cross-compiling and CI green on Linux.
- **Dedicated STA worker.** All COM calls run on one `runtime.LockOSThread`-locked
  goroutine with a single `CoInitializeEx`. Commands are serialized over a
  channel; the `Outlook.Application` object is created once and reused;
  `CoUninitialize` runs only when a COM reference was actually acquired
  (`S_OK`/`S_FALSE`, never after `RPC_E_CHANGED_MODE`).
- **One renderer.** `campaign.Renderer` produces a single `RenderedMessage` used
  by preview, test send, and bulk send, guaranteeing equivalent output.
- **Preflight gates sending.** `campaign.Preflighter` returns a structured
  `PreflightResult`; the UI cannot send while `CanSend` is false.
- **Typed results.** `CampaignResult` carries an explicit state
  (`completed`/`cancelled`/`preflight_failed`/`runtime_failed`) and per-recipient
  attempt history, so a fatal failure is never a zero-failure "success".
- **Injectable Clock.** Delays/batch pauses use a `Clock`; tests use a fake clock
  and never sleep.
- **Dependency injection.** `NewApp` constructs services and the sender and wires
  them into the runner; nothing constructs platform services ad hoc deep in the
  call graph.

## Persistence

JSON files under `%AppData%\MailMergeGo`, all written atomically (temp + rename).
Campaign history sits behind a `storage.CampaignRepository` interface so a SQLite
implementation can replace the JSON store without touching the app layer.
