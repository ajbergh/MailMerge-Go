/*
Merge Service - Template Variable Replacement

This service handles mail merge functionality by replacing template variables
(merge fields) with actual contact data.

Phase 1 Updates (v1.2):
  - Now supports ANY merge field from the contact's CustomFields map
  - Fields like {Company}, {Department}, {Phone} work automatically if present in import file
  - Merge fields are still case-insensitive

Standard Merge Fields:
  - {FirstName} - Contact's first name
  - {LastName}  - Contact's last name
  - {Email}     - Contact's email address

Custom Merge Fields:
  - Any column from the imported CSV/Excel file can be used as {ColumnName}
  - Column names are matched case-insensitively

Templates can be used in both the email subject and body. Any unrecognized
merge fields are replaced with empty strings to avoid sending literal {Field}
text to recipients.
*/
package services

import (
	"MailMergeApp/backend/models"
	"regexp"
	"strings"
)

// MergeService handles template merging with contact data.
// Thread-safe - all operations are stateless after initialization.
type MergeService struct {
	// fieldPattern matches merge fields like {FirstName}, {LastName}, {Email}
	fieldPattern *regexp.Regexp
	// pTagPattern matches <p> tags to add inline styles for email compatibility
	pTagPattern *regexp.Regexp
}

// NewMergeService creates a new MergeService instance with a compiled
// regex pattern for efficient template processing.
func NewMergeService() *MergeService {
	return &MergeService{
		fieldPattern: regexp.MustCompile(`\{(\w+)\}`),
		pTagPattern:  regexp.MustCompile(`<p([^>]*)>`),
	}
}

// MergeResult contains the rendered subject and body after merge fields
// have been replaced with contact data.
type MergeResult struct {
	Subject string `json:"subject"` // Rendered email subject
	Body    string `json:"body"`    // Rendered email body
}

// NormalizeHTMLForEmail prepares HTML content for email clients by adding
// inline styles to neutralize default margins that cause spacing issues.
//
// ReactQuill generates HTML with <p> tags that have default browser margins.
// These look fine in the editor (CSS overrides them) but create extra spacing
// in email clients like Outlook. This function:
//   - Adds inline margin:0 style to all <p> tags
//   - Converts empty paragraphs (<p><br></p>) to simple <br> tags
//
// Parameters:
//   - html: Raw HTML content from the editor
//
// Returns:
//   - string: Normalized HTML suitable for email clients
func (ms *MergeService) NormalizeHTMLForEmail(html string) string {
	// Add inline style to <p> tags to remove default margins
	// Handles <p>, <p class="...">, <p style="..."> etc.
	normalized := ms.pTagPattern.ReplaceAllStringFunc(html, func(match string) string {
		// Check if it already has a style attribute
		if strings.Contains(match, "style=") {
			// Prepend margin:0 to existing style
			return strings.Replace(match, "style=\"", "style=\"margin:0;padding:0;", 1)
		}
		// Add new style attribute
		if match == "<p>" {
			return "<p style=\"margin:0;padding:0;\">"
		}
		// Has other attributes but no style
		return strings.Replace(match, "<p", "<p style=\"margin:0;padding:0;\"", 1)
	})

	// Replace empty paragraphs with single <br>
	// ReactQuill generates <p><br></p> for blank lines
	normalized = strings.ReplaceAll(normalized, "<p><br></p>", "<br>")
	normalized = strings.ReplaceAll(normalized, "<p style=\"margin:0;padding:0;\"><br></p>", "<br>")

	return normalized
}

// RenderForContact renders the subject and body templates for a specific contact.
// All merge fields in the templates are replaced with the contact's data.
// Phase 1: Now uses CustomFields for dynamic field support.
//
// Parameters:
//   - subjectTemplate: Email subject containing merge fields
//   - bodyTemplate: Email body containing merge fields
//   - contact: Contact data to use for replacement (includes CustomFields)
//
// Returns:
//   - MergeResult: Rendered subject and body
func (ms *MergeService) RenderForContact(subjectTemplate, bodyTemplate string, contact models.Contact) MergeResult {
	return MergeResult{
		Subject: ms.renderTemplate(subjectTemplate, contact),
		Body:    ms.renderTemplate(bodyTemplate, contact),
	}
}

// renderTemplate replaces all merge fields in a template with contact data.
// Phase 1: Now checks CustomFields for any field not in standard set.
// Matching is case-insensitive. Unknown fields are replaced with empty string.
func (ms *MergeService) renderTemplate(template string, contact models.Contact) string {
	// Replace all merge fields
	result := ms.fieldPattern.ReplaceAllStringFunc(template, func(match string) string {
		// Extract field name (remove braces) and normalize to lowercase
		fieldName := strings.ToLower(strings.Trim(match, "{}"))

		// Check standard fields first
		switch fieldName {
		case "firstname":
			return contact.FirstName
		case "lastname":
			return contact.LastName
		case "email":
			return contact.Email
		}

		// Check custom fields (Phase 1 feature)
		if contact.CustomFields != nil {
			if value, ok := contact.CustomFields[fieldName]; ok {
				return value
			}
		}

		// Return empty string for unknown fields
		return ""
	})

	return result
}

// PreviewMerge renders a preview of the merged email using sample data.
// If sample values are empty, defaults are used (John Doe, john.doe@example.com).
//
// Parameters:
//   - subjectTemplate: Email subject with merge fields
//   - bodyTemplate: Email body with merge fields
//   - sampleFirstName, sampleLastName, sampleEmail: Sample data for preview
//
// Returns:
//   - MergeResult: Rendered subject and body using sample data
func (ms *MergeService) PreviewMerge(subjectTemplate, bodyTemplate string, sampleFirstName, sampleLastName, sampleEmail string) MergeResult {
	sampleContact := models.Contact{
		FirstName: sampleFirstName,
		LastName:  sampleLastName,
		Email:     sampleEmail,
	}

	// Use default sample values if not provided
	if sampleContact.FirstName == "" {
		sampleContact.FirstName = "John"
	}
	if sampleContact.LastName == "" {
		sampleContact.LastName = "Doe"
	}
	if sampleContact.Email == "" {
		sampleContact.Email = "john.doe@example.com"
	}

	return ms.RenderForContact(subjectTemplate, bodyTemplate, sampleContact)
}

// PreviewMergeWithContact renders a preview using a specific contact's data.
// Phase 1: Added for the email preview modal feature.
//
// Parameters:
//   - subjectTemplate: Email subject with merge fields
//   - bodyTemplate: Email body with merge fields
//   - contact: Contact to use for preview (includes CustomFields)
//
// Returns:
//   - MergeResult: Rendered subject and body using contact's actual data
func (ms *MergeService) PreviewMergeWithContact(subjectTemplate, bodyTemplate string, contact models.Contact) MergeResult {
	return ms.RenderForContact(subjectTemplate, bodyTemplate, contact)
}

// ValidateTemplateWithHeaders checks if a template contains only valid merge fields
// based on the available headers from the imported file.
// Phase 1: Now validates against dynamic headers rather than fixed list.
//
// Parameters:
//   - template: Template string to validate
//   - availableHeaders: List of valid column headers from import file
//
// Returns:
//   - []string: List of invalid field names (fields not in headers)
func (ms *MergeService) ValidateTemplateWithHeaders(template string, availableHeaders []string) []string {
	// Build a set of valid fields (normalized to lowercase)
	validFields := make(map[string]bool)
	for _, header := range availableHeaders {
		validFields[strings.ToLower(header)] = true
	}
	// Always include standard fields
	validFields["firstname"] = true
	validFields["lastname"] = true
	validFields["email"] = true

	matches := ms.fieldPattern.FindAllString(template, -1)
	invalidFields := []string{}

	for _, match := range matches {
		fieldName := strings.ToLower(strings.Trim(match, "{}"))
		if !validFields[fieldName] {
			invalidFields = append(invalidFields, match)
		}
	}

	return invalidFields
}

// GetAvailableFields returns the list of default merge fields.
// Phase 1: For dynamic fields, use GetMergeFieldsFromHeaders instead.
func (ms *MergeService) GetAvailableFields() []string {
	return []string{"{FirstName}", "{LastName}", "{Email}"}
}

// ValidateTemplate checks if a template contains only valid merge fields
// using the default field list. For validation against custom headers,
// use ValidateTemplateWithHeaders instead.
//
// Parameters:
//   - template: Template string to validate
//
// Returns:
//   - []string: List of invalid field names (fields not in default set)
func (ms *MergeService) ValidateTemplate(template string) []string {
	return ms.ValidateTemplateWithHeaders(template, []string{"FirstName", "LastName", "Email"})
}

// GetMergeFieldsFromHeaders converts column headers to merge field format.
// Phase 1: Generates dynamic merge fields from imported file headers.
//
// Parameters:
//   - headers: Column headers from the imported file
//
// Returns:
//   - []string: Merge fields in {FieldName} format
func (ms *MergeService) GetMergeFieldsFromHeaders(headers []string) []string {
	fields := make([]string, 0, len(headers))
	seen := make(map[string]bool)

	for _, header := range headers {
		normalized := strings.TrimSpace(header)
		if normalized == "" {
			continue
		}
		// Avoid duplicates (case-insensitive)
		if seen[strings.ToLower(normalized)] {
			continue
		}
		seen[strings.ToLower(normalized)] = true
		fields = append(fields, "{"+normalized+"}")
	}

	return fields
}

// ExtractFields returns all merge fields found in a template.
// Duplicates are removed from the result.
//
// Parameters:
//   - template: Template string to scan
//
// Returns:
//   - []string: Unique list of merge fields found
func (ms *MergeService) ExtractFields(template string) []string {
	matches := ms.fieldPattern.FindAllString(template, -1)
	// Remove duplicates
	seen := make(map[string]bool)
	unique := []string{}
	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			unique = append(unique, match)
		}
	}
	return unique
}
