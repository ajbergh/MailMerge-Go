# Outlook Compatibility

## Supported

- **Classic desktop Outlook** (the traditional Win32 app, part of Microsoft 365 /
  Office) via COM automation (`Outlook.Application`).
- Windows 10 and Windows 11.

## Not supported for sending

- **New Outlook** (the Windows Store / web-based app that replaces Mail & the
  classic client): it does **not** expose the classic COM automation surface, so
  MailMerge Go cannot drive it. The app detects this and reports an actionable
  status rather than falsely claiming support.
- Outlook on the web (OWA).

## Detection

`outlook.Sender.Preflight` attempts to reach `Outlook.Application` on the STA
worker thread and returns a `SenderStatus`:

| State                 | Meaning                                              |
|-----------------------|------------------------------------------------------|
| `classic_outlook`     | Reachable via COM; ready to send.                    |
| `unavailable`         | Could not create the COM object (New Outlook / none).|
| `com_inaccessible`    | COM present but not accessible.                      |
| `blocked_by_policy`   | Automation blocked by security policy.               |
| `no_account`          | No configured mail account.                          |

## Threading

COM requires that objects be used on the thread that created them. All COM work
runs on a single dedicated, `LockOSThread`-locked goroutine initialized once with
`CoInitializeEx(COINIT_APARTMENTTHREADED)`. This avoids the classic
"interface marshalled for a different thread" error.

## Delivery semantics

Outlook confirms **submission** (accepted for sending), not delivery. Results say
"submitted", never "delivered".

## Migration path

A future `MicrosoftGraphSender` (see [graph-sender-design.md](graph-sender-design.md))
can send without a local Outlook install, behind the same `EmailSender` interface.
