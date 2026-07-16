/*
Merge Service - Template Variable Replacement (compatibility shim)

This service now delegates all merge-field parsing and rendering to the
canonical backend/mergefield package. It retains its previous method surface so
existing Wails bindings and the Outlook service keep working, but the underlying
behavior is the new deterministic renderer:

  - Fields with spaces, hyphens, punctuation, or Unicode are supported.
  - Unknown fields are reported (not silently blanked); see the campaign
    preflight for send-blocking behavior.
  - Both {{field_id}} and legacy {Field} syntaxes are accepted.

HTML escaping of merge values is applied by the campaign renderer on the send
path (see backend/campaign). NormalizeHTMLForEmail is unchanged.
*/
package services

import (
	"regexp"
	"strings"

	"MailMergeApp/backend/mergefield"
	"MailMergeApp/backend/models"
)

// MergeService handles template merging with contact data.
// Thread-safe - all operations are stateless after initialization.
type MergeService struct {
	// pTagPattern matches <p> tags to add inline styles for email compatibility
	pTagPattern *regexp.Regexp
}

// NewMergeService creates a new MergeService instance.
func NewMergeService() *MergeService {
	return &MergeService{
		pTagPattern: regexp.MustCompile(`<p([^>]*)>`),
	}
}

// MergeResult contains the rendered subject and body after merge fields
// have been replaced with contact data.
type MergeResult struct {
	Subject string `json:"subject"` // Rendered email subject
	Body    string `json:"body"`    // Rendered email body
}

// schemaForContact builds a merge-field schema from a single contact's own
// custom-field keys plus the standard fields. Suitable for one-off rendering
// where the full import schema is not available.
func schemaForContact(c models.Contact) *mergefield.Schema {
	headers := make([]string, 0, len(c.CustomFields))
	for k := range c.CustomFields {
		headers = append(headers, k)
	}
	return mergefield.NewSchema(headers)
}

// NormalizeHTMLForEmail prepares HTML content for email clients by adding
// inline styles to neutralize default margins that cause spacing issues.
func (ms *MergeService) NormalizeHTMLForEmail(html string) string {
	normalized := ms.pTagPattern.ReplaceAllStringFunc(html, func(match string) string {
		if strings.Contains(match, "style=") {
			return strings.Replace(match, "style=\"", "style=\"margin:0;padding:0;", 1)
		}
		if match == "<p>" {
			return "<p style=\"margin:0;padding:0;\">"
		}
		return strings.Replace(match, "<p", "<p style=\"margin:0;padding:0;\"", 1)
	})
	normalized = strings.ReplaceAll(normalized, "<p><br></p>", "<br>")
	normalized = strings.ReplaceAll(normalized, "<p style=\"margin:0;padding:0;\"><br></p>", "<br>")
	return normalized
}

// RenderForContact renders subject and body templates for a specific contact
// using the canonical renderer. Values are not HTML-escaped here; the campaign
// renderer performs context-aware escaping on the send path.
func (ms *MergeService) RenderForContact(subjectTemplate, bodyTemplate string, contact models.Contact) MergeResult {
	r := mergefield.NewRenderer(schemaForContact(contact))
	subject, _ := r.Render(subjectTemplate, contact, false)
	body, _ := r.Render(bodyTemplate, contact, false)
	return MergeResult{Subject: subject, Body: body}
}

// PreviewMerge renders a preview using sample data. Defaults are used for blank
// sample values.
func (ms *MergeService) PreviewMerge(subjectTemplate, bodyTemplate, sampleFirstName, sampleLastName, sampleEmail string) MergeResult {
	sample := models.Contact{FirstName: sampleFirstName, LastName: sampleLastName, Email: sampleEmail}
	if sample.FirstName == "" {
		sample.FirstName = "John"
	}
	if sample.LastName == "" {
		sample.LastName = "Doe"
	}
	if sample.Email == "" {
		sample.Email = "john.doe@example.com"
	}
	return ms.RenderForContact(subjectTemplate, bodyTemplate, sample)
}

// PreviewMergeWithContact renders a preview using a specific contact's data.
func (ms *MergeService) PreviewMergeWithContact(subjectTemplate, bodyTemplate string, contact models.Contact) MergeResult {
	return ms.RenderForContact(subjectTemplate, bodyTemplate, contact)
}

// ValidateTemplateWithHeaders returns the invalid (unknown) merge fields in a
// template given the available import headers.
func (ms *MergeService) ValidateTemplateWithHeaders(template string, availableHeaders []string) []string {
	r := mergefield.NewRenderer(mergefield.NewSchema(availableHeaders))
	var invalid []string
	for _, d := range r.ValidateTemplate(template) {
		if d.Code == mergefield.CodeUnknownField {
			invalid = append(invalid, "{"+d.Token+"}")
		}
	}
	return invalid
}

// GetAvailableFields returns the default standard merge-field tokens.
func (ms *MergeService) GetAvailableFields() []string {
	return mergefield.NewSchema(nil).InsertTokens()
}

// ValidateTemplate validates a template against only the standard fields.
func (ms *MergeService) ValidateTemplate(template string) []string {
	return ms.ValidateTemplateWithHeaders(template, nil)
}

// GetMergeFieldsFromHeaders converts import headers to canonical insertion
// tokens (e.g. {{account_manager}}), including the standard fields.
func (ms *MergeService) GetMergeFieldsFromHeaders(headers []string) []string {
	return mergefield.NewSchema(headers).InsertTokens()
}

// ExtractFields returns the unique field names referenced in a template.
func (ms *MergeService) ExtractFields(template string) []string {
	r := mergefield.NewRenderer(mergefield.NewSchema(nil))
	return r.ExtractTokens(template)
}
