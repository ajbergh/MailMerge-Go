# Contributing

Thanks for your interest in improving MailMerge Go.

## Prerequisites

- Go 1.24+
- Node.js 22.23.1 (the version pinned in CI and release workflows)
- Wails CLI v2.11.0 (`go install github.com/wailsapp/wails/v2/cmd/wails@v2.11.0`)
- golangci-lint v2.11.4 when running lint locally

## Development

```powershell
wails dev
```

## Before opening a pull request

Run the same core checks CI runs:

```powershell
gofmt -w .
go vet ./backend/...
go test -race -covermode=atomic -coverprofile=backend-coverage.out ./backend/...
golangci-lint run ./backend/...

cd frontend
npm ci
npm run test -- --coverage
npm run build
npm audit --omit=dev --audit-level=high
```

Go vulnerability check:

```powershell
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./backend/...
```

To verify the packaged Windows application locally, run
`./scripts/build-windows.ps1`, then run `go test .` on Windows to compile and test
the root Wails-bound application APIs. Linux and macOS scripts are intended for
their respective native hosts.

## Release candidates

Maintainers should run the **Release Candidate** GitHub Actions workflow before
creating a version tag. It produces coverage evidence, vulnerability-scan results,
CycloneDX SBOMs, a Windows executable, and checksums without publishing a release.
Follow [`dev_docs/release-checklist.md`](dev_docs/release-checklist.md).

## Conventions

- Keep Outlook COM / `go-ole` confined to `backend/outlook`. The campaign engine
  and everything above it must depend only on `campaign.EmailSender`.
- Preview, test send, and bulk send must use the single `campaign.Renderer`; do not
  add ad-hoc frontend merge substitutions.
- Persisted campaign runs are immutable snapshots. Retries create linked child runs
  rather than rewriting the original result.
- Persisted files must be written atomically and validated/normalized on load.
- Add or update tests for behavior changes; the campaign engine must remain testable
  without Outlook by using `campaign.FakeSender`.
- Regenerate Wails bindings after changing bound Go signatures with
  `wails generate module`, and commit the generated bindings with the source change.
- Never commit contact data, generated campaign history, signing certificates,
  signing passwords, access tokens, or other secrets.

## Commits and branches

- Work on a feature branch; do not rewrite `main` history.
- Prefer small, focused commits grouped by concern.
- Keep pull requests in draft while required CI checks are red.
- Do not merge by bypassing the required release and security gates.
