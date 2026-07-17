# Security

## Trust boundaries

- **Imported contact data is untrusted.** It may contain HTML/script. Merge values
  are HTML-escaped before insertion into HTML bodies (`mergefield` renderer), so
  imported data cannot inject markup.
- **Authored template HTML is semi-trusted.** The rendered HTML body is sanitized
  with a well-maintained sanitizer (bluemonday, `backend/htmlutil`): scripts,
  event handlers (`onclick`, …), `javascript:`/unsafe URLs, `<iframe>`, and
  `<object>`/embeds are removed; a safe rich-text subset (paragraphs, lists,
  headings, marks, links, a few inline styles/classes) is preserved.
- **Previews** are additionally sanitized client-side with DOMPurify and rendered
  in a sandboxed `<iframe>` (`sandbox="allow-same-origin"`, no scripts).
- Plain-text messages never pass through HTML rendering.

## Persistence safety

- Template IDs are backend-generated (UUIDs); IDs with path separators or other
  unsafe characters are rejected, so a crafted ID cannot traverse directories.
- Built-in templates cannot be overwritten via a crafted ID.
- Settings are validated before replacing the live object; corrupted settings
  files are backed up (`.corrupt`) and defaults restored.
- Settings, templates, suppression list, and campaign history are written
  atomically (temp file + `rename`).
- Recent-files getters return copies (no shared mutable backing storage) and prune
  entries that no longer exist.

## Sending safety

- A campaign **preflight** blocks sending on unknown/ambiguous merge fields,
  invalid To/CC/BCC addresses, blank bodies, missing/oversized/directory
  attachments, zero recipients, and unsupported sender features.
- Duplicate recipients are resolved by an explicit policy (default keep-first),
  normalized without stripping dots/plus tags.
- A local **suppression list** blocks suppressed recipients during preflight.
- Results distinguish **submitted** (accepted by Outlook) from delivered — the app
  never claims confirmed delivery.

## What is not stored / sent

- No credentials or tokens are stored (sending uses the local Outlook profile).
- No tracking pixels or unsubscribe links are injected.
- No contact data or email content leaves the machine via this application.

## Reporting

See [SECURITY.md](../SECURITY.md) for how to report vulnerabilities.
