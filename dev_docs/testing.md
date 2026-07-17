# Testing

## Go (no Outlook required)

```powershell
go test ./backend/...          # unit tests
go test -race ./backend/...    # with the race detector
go vet ./backend/...
gofmt -l .                     # should print nothing
```

Coverage highlights:

- `mergefield` — canonicalization, collisions, fallbacks, HTML escaping, unknown
  vs missing fields, legacy syntax.
- `email` — `net/mail` validation, comma/semicolon lists, duplicate policy
  (dots/plus tags preserved).
- `campaign` — success, partial failure, fatal (runtime_failed), cancellation,
  retry-without-double-count, dedupe, suppression, delay/batch pacing via a fake
  clock, preflight blocking, preview/test/bulk render equivalence, HTML sanitize.
- `htmlutil` — script/handler/URL stripping; safe rich-text preserved.
- `services` — settings validation/migration/corruption, template ID safety +
  atomic writes, suppression, CSV/BOM/blank-row/duplicate-header import.
- `storage` — campaign history repository.

The campaign engine is exercised entirely through `campaign.FakeSender`, so no
Outlook install is needed.

## Windows / Outlook integration tests (opt-in)

Guarded by a build tag so default CI never needs Outlook:

```powershell
go test -tags outlookintegration ./backend/outlook/...
```

Run on a Windows workstation with **classic** Outlook and a configured account.
These tests prefer creating/inspecting items over sending real email; review each
before enabling on a machine with a live mailbox.

## Frontend

```powershell
cd frontend
npm ci
npm run test     # vitest (jsdom)
npm run build    # tsc + vite build
```

The frontend uses Vite 8, Vitest 4, and TypeScript 6. `vitest.config.ts` runs
tests in jsdom and explicitly uses esbuild's automatic JSX runtime because React
Fast Refresh is not needed in tests. `tsconfig.json` excludes test files from the
production type check.

## CI

`.github/workflows/ci.yml` runs all of the above plus golangci-lint and
dependency scanning on every PR and push to `main`.
