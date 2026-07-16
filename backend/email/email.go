/*
Package email provides address validation, normalization, and duplicate
resolution for MailMerge-Go.

It replaces the previous naive validation (a bare check for "@" and ".") with
Go's net/mail parser, and implements a documented duplicate-normalization policy
that does NOT over-normalize provider-specific addresses (dots and plus tags are
preserved).
*/
package email

import (
	"fmt"
	"net/mail"
	"strings"
)

// Address is a parsed, validated email address.
type Address struct {
	// Name is the optional display name ("Ada Lovelace").
	Name string
	// Address is the addr-spec ("ada@example.com"), trimmed but case-preserved.
	Address string
}

// String renders the address for use in an Outlook recipient field.
func (a Address) String() string {
	if a.Name != "" {
		return fmt.Sprintf("%s <%s>", a.Name, a.Address)
	}
	return a.Address
}

// Parse validates and parses a single address using net/mail. Surrounding
// whitespace is normalized. Display names are preserved when present.
func Parse(raw string) (Address, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return Address{}, fmt.Errorf("empty address")
	}
	m, err := mail.ParseAddress(s)
	if err != nil {
		return Address{}, fmt.Errorf("invalid email address %q: %w", raw, err)
	}
	// net/mail accepts local-only addresses like "root@localhost"; require a dot
	// in the domain to reject obvious typos while still allowing valid TLDs.
	at := strings.LastIndex(m.Address, "@")
	if at < 0 || !strings.Contains(m.Address[at+1:], ".") {
		return Address{}, fmt.Errorf("invalid email address %q: missing domain", raw)
	}
	return Address{Name: strings.TrimSpace(m.Name), Address: strings.TrimSpace(m.Address)}, nil
}

// IsValid reports whether raw is a single valid address.
func IsValid(raw string) bool {
	_, err := Parse(raw)
	return err == nil
}

// ParseList parses a recipient list delimited by commas or semicolons (Outlook
// accepts both). It returns the valid addresses and the raw entries that failed.
func ParseList(raw string) (valid []Address, invalid []string) {
	for _, part := range splitList(raw) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if a, err := Parse(part); err == nil {
			valid = append(valid, a)
		} else {
			invalid = append(invalid, part)
		}
	}
	return valid, invalid
}

// splitList splits on commas and semicolons that are not inside a quoted display
// name or angle brackets.
func splitList(raw string) []string {
	var parts []string
	var b strings.Builder
	inQuote := false
	inAngle := false
	for _, r := range raw {
		switch r {
		case '"':
			inQuote = !inQuote
			b.WriteRune(r)
		case '<':
			inAngle = true
			b.WriteRune(r)
		case '>':
			inAngle = false
			b.WriteRune(r)
		case ',', ';':
			if inQuote || inAngle {
				b.WriteRune(r)
			} else {
				parts = append(parts, b.String())
				b.Reset()
			}
		default:
			b.WriteRune(r)
		}
	}
	if b.Len() > 0 {
		parts = append(parts, b.String())
	}
	return parts
}

// NormalizeKey returns the canonical key used for duplicate detection.
//
// Policy (documented and intentionally conservative):
//   - Trim whitespace.
//   - Lower-case the address (email addresses are case-insensitive in practice).
//   - Do NOT strip dots from the local part.
//   - Do NOT strip "+tag" suffixes.
//
// This means alice@example.com and Alice@Example.com are duplicates, but
// alice+news@example.com and alice@example.com are treated as distinct, because
// they may be intentionally different deliverable addresses.
func NormalizeKey(raw string) string {
	s := strings.TrimSpace(raw)
	if a, err := Parse(s); err == nil {
		s = a.Address
	}
	return strings.ToLower(s)
}
