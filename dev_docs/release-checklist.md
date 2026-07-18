# MailMerge Go Release Checklist

Use this checklist for every release candidate and tagged release. A release is a **no-go** if any required item is incomplete.

## 1. Source and change control

- [ ] The release commit is on `main` and is associated with a reviewed pull request.
- [ ] Required CI checks are green: Go test/vet/lint, frontend test/build, dependency scan, SBOM generation, and Windows Wails build.
- [ ] `CHANGELOG.md` describes all user-visible, compatibility, security, and migration changes.
- [ ] `REMEDIATION_STATUS.md` does not contain an unresolved release blocker.
- [ ] The version follows semantic versioning and the release tag has the form `vX.Y.Z`.

## 2. Automated validation

- [ ] Run the **Release Candidate** workflow manually against the intended commit.
- [ ] Review backend and frontend coverage artifacts for unexpected regressions.
- [ ] Review the `govulncheck-report` artifact and confirm no reachable Go vulnerabilities.
- [ ] Confirm `npm audit --omit=dev --audit-level=high` passes.
- [ ] Confirm both CycloneDX SBOM files are generated and readable.
- [ ] Confirm the unsigned Windows release-candidate executable is produced with a SHA-256 checksum.

## 3. Windows and Outlook compatibility

Validate on a clean supported Windows environment:

- [ ] Classic Outlook is installed and has at least one configured account.
- [ ] Outlook readiness reports a clear actionable result.
- [ ] Import a CSV and an XLSX contact list.
- [ ] Preview a personalized HTML and plain-text message.
- [ ] Create a test draft and send a test message.
- [ ] Run a small campaign with attachments.
- [ ] Cancel an active campaign and verify completed results are retained.
- [ ] Retry failed recipients from campaign history after restarting the application.
- [ ] Verify New Outlook-only systems are rejected with an actionable explanation.
- [ ] Verify clean shutdown while Outlook is open and after Outlook has been closed.

## 4. Persistence and privacy

- [ ] Settings, templates, suppression records, and campaign history reload after restart.
- [ ] Campaign retries preserve immutable original snapshots and parent/child run lineage.
- [ ] Delete-one and clear-all history controls behave as described.
- [ ] Exported logs contain only the intended recipient data.
- [ ] Local data locations and retention expectations are documented.
- [ ] No credentials, certificates, contact lists, generated history, or signing material are committed.

## 5. Packaging and signing

- [ ] Build metadata displays the version, commit, and build date.
- [ ] The release manifest matches the tag and commit.
- [ ] The executable is Authenticode-signed when signing secrets are configured.
- [ ] Signature verification succeeds on the final executable.
- [ ] `SHA256SUMS.txt` includes the executable, SBOMs, and release manifest.
- [ ] Install, upgrade, launch, repair, and uninstall are smoke-tested where an installer is produced.

## 6. Publication and rollback

- [ ] Tag the exact validated commit; do not rebuild from a different commit.
- [ ] Verify the GitHub release contains the executable, checksums, SBOMs, and release manifest.
- [ ] Review generated release notes before publishing.
- [ ] Preserve the previous stable release and its checksums for rollback.
- [ ] Record known limitations, especially the requirement for classic Outlook COM automation.
- [ ] Monitor the Security tab and issue tracker after publication.

## Required branch protection

Configure `main` to require these checks before merge:

1. `Go (test, vet, lint)`
2. `Frontend (test, coverage, build)`
3. `Wails Windows build`
4. `Dependency scan`
5. `Generate SBOMs`

Also require pull requests, prevent force pushes, dismiss stale approvals when code changes, and require the branch to be current before merging.
