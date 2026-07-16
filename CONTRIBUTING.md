# Contributing

Thanks for your interest in improving MailMerge Go.

## Prerequisites

- Go 1.24+
- Node.js 20+
- Wails CLI v2.11.0 (`go install github.com/wailsapp/wails/v2/cmd/wails@v2.11.0`)

## Development

```powershell
wails dev            # run the app with hot reload
```

## Before opening a pull request

Run the same checks CI runs:

```powershell
gofmt -w .
go vet ./backend/...
go test -race ./backend/...

cd frontend
npm ci
npm run test
npm run build
```

Optionally: `golangci-lint run ./backend/...`.

## Conventions

- Keep Outlook COM / `go-ole` confined to `backend/outlook`. The campaign engine
  and everything above must depend only on `campaign.EmailSender`.
- Preview, test send, and bulk send must go through the single
  `campaign.Renderer` — do not add ad-hoc frontend merge substitutions.
- Persisted files must be written atomically and validated on load.
- Add or update tests for behavior changes; the campaign engine must stay testable
  without Outlook (use `campaign.FakeSender`).
- Regenerate Wails bindings after changing bound Go signatures:
  `wails generate module`.

## Commits & branches

- Work on a feature branch; do not rewrite `main` history.
- Prefer small, focused commits grouped by concern.
