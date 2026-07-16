/*
Package mergefield implements the canonical merge-field model for MailMerge-Go.

It replaces the previous `\{(\w+)\}` parsing approach, which could not represent
column headers containing spaces, hyphens, punctuation, or non-ASCII characters
and silently rendered unknown fields as empty strings.

Design
  - Every imported column is represented by a Field with a stable canonical ID,
    the original display name, a normalized lookup key (matching the key used in
    Contact.CustomFields), and optional aliases.
  - A Schema is the set of fields available for a given import. It always includes
    the three standard fields (first_name, last_name, email) and detects
    canonical-name collisions between imported columns.
  - Canonicalization is deterministic and documented (see Canonicalize).

Supported template syntax (see renderer.go):
  - Canonical:      {{field_id}}
  - With fallback:  {{field_id|default value}}
  - Legacy:         {Field}   (single-brace, no fallback) for backward compatibility

Both syntaxes resolve through the same Schema, so preview, test send, and bulk
send always produce identical output.
*/
package mergefield

import (
	"sort"
	"strings"
	"unicode"

	"MailMergeApp/backend/models"
)

// Canonical IDs for the three standard fields that always exist.
const (
	StandardFirstName = "first_name"
	StandardLastName  = "last_name"
	StandardEmail     = "email"
)

// Field describes a single merge field derived from an imported column or a
// built-in standard field.
type Field struct {
	ID          string   `json:"id"`          // Canonical identifier, e.g. "account_manager"
	DisplayName string   `json:"displayName"` // Original column header, e.g. "Account-Manager"
	LookupKey   string   `json:"lookupKey"`   // Key used to read Contact.CustomFields (lowercased header)
	Aliases     []string `json:"aliases,omitempty"`
	Standard    bool     `json:"standard"`           // True for the three built-in fields
	Required    bool     `json:"required,omitempty"` // True if a value is required (email)
}

// Token returns the canonical insertion token for the field, e.g. "{{email}}".
func (f Field) Token() string { return "{{" + f.ID + "}}" }

// Collision records two or more display names that canonicalize to the same ID.
type Collision struct {
	ID           string   `json:"id"`
	DisplayNames []string `json:"displayNames"`
}

// Schema is the resolved set of merge fields available for an import.
type Schema struct {
	Fields     []Field     `json:"fields"`
	Collisions []Collision `json:"collisions,omitempty"`

	byKey map[string]int // normalized alias/id/lookupkey/displayname -> index into Fields
}

// standardFields returns freshly-built standard field definitions.
func standardFields() []Field {
	return []Field{
		{ID: StandardFirstName, DisplayName: "First Name", LookupKey: "firstname",
			Aliases: []string{"firstname", "first", "fname", "givenname", "given_name"}, Standard: true},
		{ID: StandardLastName, DisplayName: "Last Name", LookupKey: "lastname",
			Aliases: []string{"lastname", "last", "lname", "surname", "familyname", "family_name"}, Standard: true},
		{ID: StandardEmail, DisplayName: "Email", LookupKey: "email",
			Aliases: []string{"email", "emailaddress", "email_address", "mail", "e_mail", "e_mail_address"},
			Standard: true, Required: true},
	}
}

// Canonicalize converts an arbitrary display name into a deterministic canonical
// identifier.
//
// Algorithm:
//  1. Trim surrounding whitespace.
//  2. Lower-case every Unicode letter and keep Unicode digits.
//  3. Replace every run of non-alphanumeric characters with a single "_".
//  4. Trim leading/trailing "_".
//
// Examples:
//
//	"First Name"      -> "first_name"
//	"Account-Manager" -> "account_manager"
//	"Customer ID"     -> "customer_id"
//	"E-mail Address"  -> "e_mail_address"
//	"Präsident"       -> "präsident"
func Canonicalize(s string) string {
	var b strings.Builder
	pendingSep := false
	for _, r := range strings.TrimSpace(s) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if pendingSep && b.Len() > 0 {
				b.WriteRune('_')
			}
			pendingSep = false
			b.WriteRune(unicode.ToLower(r))
		default:
			pendingSep = true
		}
	}
	return b.String()
}

// NewSchema builds a Schema from imported column headers. The three standard
// fields are always included. Headers whose canonical ID matches a standard
// field (or one of its aliases) are folded into that standard field. Remaining
// headers become custom fields; any two custom headers that canonicalize to the
// same ID are recorded as a Collision and are NOT silently merged.
func NewSchema(headers []string) *Schema {
	s := &Schema{Fields: standardFields()}

	// Map canonical alias/id -> standard field index for folding.
	standardByCanon := map[string]int{}
	for i, f := range s.Fields {
		standardByCanon[f.ID] = i
		for _, a := range f.Aliases {
			standardByCanon[Canonicalize(a)] = i
		}
	}

	// Track custom fields by canonical ID so we can detect collisions.
	customIndex := map[string]int{}          // id -> index into s.Fields
	collisionNames := map[string][]string{}  // id -> display names contributing to a collision

	for _, h := range headers {
		display := strings.TrimSpace(h)
		if display == "" {
			continue
		}
		id := Canonicalize(display)
		if id == "" {
			continue
		}
		lookup := strings.ToLower(display)

		// Fold headers that map onto a standard field.
		if si, ok := standardByCanon[id]; ok {
			// Record the real column header as an alias/lookup for the standard field.
			addAlias(&s.Fields[si], lookup)
			continue
		}

		if existing, ok := customIndex[id]; ok {
			// Collision: a different display name produced the same canonical ID.
			if s.Fields[existing].DisplayName != display {
				if len(collisionNames[id]) == 0 {
					collisionNames[id] = []string{s.Fields[existing].DisplayName}
				}
				collisionNames[id] = appendUnique(collisionNames[id], display)
			}
			addAlias(&s.Fields[existing], lookup)
			continue
		}

		customIndex[id] = len(s.Fields)
		s.Fields = append(s.Fields, Field{
			ID:          id,
			DisplayName: display,
			LookupKey:   lookup,
		})
	}

	// Materialize collisions in a stable order.
	ids := make([]string, 0, len(collisionNames))
	for id := range collisionNames {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		names := collisionNames[id]
		sort.Strings(names)
		s.Collisions = append(s.Collisions, Collision{ID: id, DisplayNames: names})
	}

	s.rebuildIndex()
	return s
}

func addAlias(f *Field, alias string) {
	alias = strings.TrimSpace(alias)
	if alias == "" || alias == f.LookupKey {
		return
	}
	for _, a := range f.Aliases {
		if a == alias {
			return
		}
	}
	f.Aliases = append(f.Aliases, alias)
}

func appendUnique(s []string, v string) []string {
	for _, x := range s {
		if x == v {
			return s
		}
	}
	return append(s, v)
}

// rebuildIndex constructs the fast lookup table used by Resolve.
func (s *Schema) rebuildIndex() {
	s.byKey = map[string]int{}
	register := func(key string, idx int) {
		key = strings.ToLower(strings.TrimSpace(key))
		if key == "" {
			return
		}
		// First registration wins so standard fields take precedence.
		if _, exists := s.byKey[key]; !exists {
			s.byKey[key] = idx
		}
		if c := Canonicalize(key); c != "" {
			if _, exists := s.byKey[c]; !exists {
				s.byKey[c] = idx
			}
		}
	}
	for i, f := range s.Fields {
		register(f.ID, i)
		register(f.LookupKey, i)
		register(f.DisplayName, i)
		for _, a := range f.Aliases {
			register(a, i)
		}
	}
}

// Resolve looks up a field by a token name using its canonical ID, lookup key,
// display name, or any alias (all matched case-insensitively and via
// canonicalization). The boolean is false if no field matches.
func (s *Schema) Resolve(name string) (Field, bool) {
	if s.byKey == nil {
		s.rebuildIndex()
	}
	key := strings.ToLower(strings.TrimSpace(name))
	if idx, ok := s.byKey[key]; ok {
		return s.Fields[idx], true
	}
	if idx, ok := s.byKey[Canonicalize(name)]; ok {
		return s.Fields[idx], true
	}
	return Field{}, false
}

// HasCollision reports whether a canonical ID is subject to a collision.
func (s *Schema) HasCollision(id string) bool {
	for _, c := range s.Collisions {
		if c.ID == id {
			return true
		}
	}
	return false
}

// ValueFor returns the value of a field for a contact and whether it was present.
func ValueFor(f Field, c models.Contact) (string, bool) {
	switch f.ID {
	case StandardFirstName:
		return c.FirstName, c.FirstName != ""
	case StandardLastName:
		return c.LastName, c.LastName != ""
	case StandardEmail:
		return c.Email, c.Email != ""
	}
	if c.CustomFields != nil {
		if v, ok := c.CustomFields[f.LookupKey]; ok {
			return v, v != ""
		}
	}
	return "", false
}

// InsertTokens returns the canonical insertion tokens for all fields, in schema
// order. Used to populate the frontend merge-field buttons.
func (s *Schema) InsertTokens() []string {
	tokens := make([]string, 0, len(s.Fields))
	for _, f := range s.Fields {
		tokens = append(tokens, f.Token())
	}
	return tokens
}
