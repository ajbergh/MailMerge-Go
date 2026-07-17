# Merge Fields

Implemented in `backend/mergefield`.

## Syntax

- **Canonical:** `{{field_id}}`
- **Fallback:** `{{field_id|default text}}` — the fallback is used when the value
  is blank for a recipient (suppresses the "missing value" warning).
- **Legacy:** `{Field}` (single brace) is still resolved for backward
  compatibility with older templates and the built-in templates.

Both syntaxes resolve through the same schema, so preview, test send, and bulk
send render identically.

## Canonicalization

Each imported column produces a `Field` with a stable canonical ID:

1. Trim whitespace.
2. Lower-case Unicode letters; keep Unicode digits.
3. Replace each run of non-alphanumeric characters with a single `_`.
4. Trim leading/trailing `_`.

Examples:

| Column header      | Canonical ID       | Token                   |
|--------------------|--------------------|-------------------------|
| `First Name`       | `first_name`       | `{{first_name}}`        |
| `Account-Manager`  | `account_manager`  | `{{account_manager}}`   |
| `Customer ID`      | `customer_id`      | `{{customer_id}}`       |
| `E-mail Address`   | `email` (standard) | `{{email}}`             |
| `Región`           | `región`           | `{{región}}`            |

Standard fields (`first_name`, `last_name`, `email`) always exist and carry
aliases (e.g. `firstname`, `fname`, `surname`, `email_address`).

## Collisions

If two different columns canonicalize to the same ID (e.g. `Customer ID` and
`customer-id`), it is recorded as a **collision** and referencing that field is a
**preflight error** — the app never silently picks one column.

## Unknown / missing fields

- **Unknown field** (not in the schema): reported as an error; blocks sending.
- **Known field, blank value** for a recipient: a warning, unless a fallback is
  supplied.

## HTML safety

When rendering into an HTML body, substituted values are HTML-escaped, and the
whole body is sanitized (see [security.md](security.md)). The authored template
markup is preserved; only merge values are escaped.
