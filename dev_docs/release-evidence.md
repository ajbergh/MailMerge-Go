# Release Evidence

Each pull request and release candidate produces durable evidence that should be reviewed before publication:

- `backend-coverage` — Go race-test coverage profile.
- `frontend-coverage` — Vitest HTML/JSON coverage output.
- `golangci-lint-diagnostics` — exact Go lint findings when the gate fails.
- `frontend-build-diagnostics` — TypeScript/Vite production compiler output.
- `govulncheck-report` — machine-readable reachable Go vulnerability findings.
- `cyclonedx-sboms` — CycloneDX SBOMs for Go and production frontend dependencies.
- `MailMergeApp-windows-unsigned` — CI-built Windows executable for smoke testing.

The manual **Release Candidate** workflow additionally packages validation evidence and a checksum-protected Windows candidate without publishing a GitHub release. Review this evidence alongside `dev_docs/release-checklist.md` before creating a `v*` tag.
