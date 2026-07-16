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

Note: the toolchain uses Vite 3, so vitest is pinned to a compatible 0.34.x and
transforms JSX via esbuild (the React fast-refresh plugin's preamble is
incompatible with jsdom). `tsconfig.json` sets `types: []` and excludes test files
from the production `tsc` so a transitive modern `@types/node` cannot break the
Vite 3 / TypeScript 4.6 build.

## CI

`.github/workflows/ci.yml` runs all of the above plus golangci-lint and
dependency scanning on every PR and push to `main`.
