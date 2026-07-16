# Security Policy

## Reporting a vulnerability

Please report suspected vulnerabilities privately via GitHub Security Advisories
("Report a vulnerability" on the repository's **Security** tab) rather than
opening a public issue. Include reproduction steps and affected version/commit.

We aim to acknowledge reports promptly and to coordinate disclosure once a fix is
available.

## Scope & posture

MailMerge Go is a local Windows desktop application. It:

- sends email through the user's **local classic Outlook** profile (no credentials
  or tokens are stored by this app);
- stores settings, templates, a suppression list, and campaign history locally
  under `%AppData%\MailMergeGo` (written atomically);
- HTML-escapes untrusted imported contact data and sanitizes authored HTML bodies;
- renders previews sanitized and sandboxed;
- adds no tracking pixels or unsubscribe links, and sends no data off-device.

See [dev_docs/security.md](dev_docs/security.md) for details.

## Supported versions

Security fixes target the latest `main` and the most recent tagged release.
