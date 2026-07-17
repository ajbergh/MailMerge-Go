/*
Renderer for the canonical merge-field system.

A single Render path is used for subject, HTML body, plain-text body, To, CC,
BCC, and personalized attachment paths so that preview, test send, and bulk send
always produce identical output.

Rendering rules:
  - {{field_id}} and {field} both resolve through the Schema.
  - {{field_id|fallback}} supplies a literal fallback when the value is blank.
  - Unknown fields are left verbatim in the output and reported as an error
    diagnostic. They must never silently become empty strings; the campaign
    preflight blocks sending when an error diagnostic is present.
  - Known-but-blank fields (with no fallback) render empty and are reported as a
    warning.
  - When rendering into an HTML context, resolved values (and fallbacks) are
    HTML-escaped so imported data cannot inject markup. The template markup
    itself is authored content and is not escaped.
*/
package mergefield

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"MailMergeApp/backend/models"
)

// Severity classifies a diagnostic.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Diagnostic codes.
const (
	CodeUnknownField   = "unknown_field"
	CodeMissingValue   = "missing_value"
	CodeInvalidSyntax  = "invalid_syntax"
	CodeCollision      = "collision"
	CodeEmptyResult    = "empty_result"
	CodeUnsupportedXfm = "unsupported_transform"
)

// Diagnostic describes a single issue found while rendering a template.
type Diagnostic struct {
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	FieldID  string   `json:"fieldId,omitempty"`
	Token    string   `json:"token,omitempty"`
	Message  string   `json:"message"`
}

// tokenPattern matches, in priority order:
//  1. double-brace tokens with an optional fallback: {{ name | fallback }}
//  2. single-brace legacy tokens: { name }
//
// The inner content deliberately allows spaces, hyphens, punctuation, and
// Unicode so headers such as "Account-Manager" or "Customer ID" are matched.
var tokenPattern = regexp.MustCompile(`\{\{\s*([^{}|]+?)\s*(?:\|([^{}]*))?\}\}|\{\s*([^{}|]+?)\s*\}`)

// Renderer renders templates against a Schema.
type Renderer struct {
	schema *Schema
}

// NewRenderer creates a Renderer bound to a schema.
func NewRenderer(schema *Schema) *Renderer {
	if schema == nil {
		schema = NewSchema(nil)
	}
	return &Renderer{schema: schema}
}

// Schema returns the schema backing the renderer.
func (r *Renderer) Schema() *Schema { return r.schema }

// Render renders a template for a contact. If htmlContext is true, substituted
// values are HTML-escaped. It returns the rendered string and any diagnostics
// found. Diagnostics are deduplicated by (code, fieldId/token).
func (r *Renderer) Render(template string, contact models.Contact, htmlContext bool) (string, []Diagnostic) {
	var diags []Diagnostic
	seen := map[string]bool{}
	add := func(d Diagnostic) {
		k := string(d.Severity) + "|" + d.Code + "|" + d.FieldID + "|" + d.Token
		if seen[k] {
			return
		}
		seen[k] = true
		diags = append(diags, d)
	}

	out := tokenPattern.ReplaceAllStringFunc(template, func(match string) string {
		name, fallback, hasFallback := parseToken(match)
		if name == "" {
			add(Diagnostic{Severity: SeverityError, Code: CodeInvalidSyntax, Token: match,
				Message: fmt.Sprintf("invalid merge field syntax: %q", match)})
			return match
		}

		field, ok := r.schema.Resolve(name)
		if !ok {
			add(Diagnostic{Severity: SeverityError, Code: CodeUnknownField, Token: match,
				Message: fmt.Sprintf("unknown merge field %q", name)})
			return match // leave verbatim so the problem is visible; preflight blocks the send
		}

		if r.schema.HasCollision(field.ID) {
			add(Diagnostic{Severity: SeverityError, Code: CodeCollision, FieldID: field.ID, Token: match,
				Message: fmt.Sprintf("merge field %q is ambiguous (multiple columns share this canonical name)", field.ID)})
		}

		value, present := ValueFor(field, contact)
		if !present || value == "" {
			if hasFallback {
				value = fallback
			} else {
				add(Diagnostic{Severity: SeverityWarning, Code: CodeMissingValue, FieldID: field.ID, Token: match,
					Message: fmt.Sprintf("no value for %q for this recipient", field.DisplayName)})
				value = ""
			}
		}

		if htmlContext {
			return html.EscapeString(value)
		}
		return value
	})

	return out, diags
}

// parseToken extracts the field name and optional fallback from a matched token.
func parseToken(match string) (name, fallback string, hasFallback bool) {
	m := tokenPattern.FindStringSubmatch(match)
	if m == nil {
		return "", "", false
	}
	// Group 1/2 are the double-brace name/fallback; group 3 is the single-brace name.
	if m[1] != "" {
		return strings.TrimSpace(m[1]), m[2], strings.Contains(match, "|")
	}
	return strings.TrimSpace(m[3]), "", false
}

// ExtractTokens returns the raw field names referenced by a template (order
// preserved, deduplicated).
func (r *Renderer) ExtractTokens(template string) []string {
	var names []string
	seen := map[string]bool{}
	for _, m := range tokenPattern.FindAllStringSubmatch(template, -1) {
		name := strings.TrimSpace(m[1])
		if name == "" {
			name = strings.TrimSpace(m[3])
		}
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names
}

// ValidateTemplate returns diagnostics for a template without rendering against
// a specific contact. Used for editor-time validation and preflight. It reports
// unknown fields, invalid syntax, and collisions (but not missing values, which
// are recipient-specific).
func (r *Renderer) ValidateTemplate(template string) []Diagnostic {
	var diags []Diagnostic
	seen := map[string]bool{}
	add := func(d Diagnostic) {
		k := d.Code + "|" + d.FieldID + "|" + d.Token
		if seen[k] {
			return
		}
		seen[k] = true
		diags = append(diags, d)
	}
	for _, m := range tokenPattern.FindAllStringSubmatch(template, -1) {
		name := strings.TrimSpace(m[1])
		if name == "" {
			name = strings.TrimSpace(m[3])
		}
		if name == "" {
			add(Diagnostic{Severity: SeverityError, Code: CodeInvalidSyntax, Message: "empty merge field"})
			continue
		}
		field, ok := r.schema.Resolve(name)
		if !ok {
			add(Diagnostic{Severity: SeverityError, Code: CodeUnknownField, Token: name,
				Message: fmt.Sprintf("unknown merge field %q", name)})
			continue
		}
		if r.schema.HasCollision(field.ID) {
			add(Diagnostic{Severity: SeverityError, Code: CodeCollision, FieldID: field.ID,
				Message: fmt.Sprintf("merge field %q is ambiguous", field.ID)})
		}
	}
	return diags
}

// HasErrors reports whether any diagnostic is an error.
func HasErrors(diags []Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}
