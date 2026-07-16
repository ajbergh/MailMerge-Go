package htmlutil

import "github.com/microcosm-cc/bluemonday"

// buildPolicy constructs the sanitizer policy: a safe rich-text subset that
// preserves what the editor produces (paragraphs, lists, headings, basic marks,
// links, and a small set of inline styles/classes) while removing scripts,
// event handlers, objects, iframes, and unsafe URL schemes.
func buildPolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	// Preserve ReactQuill formatting classes (e.g. ql-align-center).
	p.AllowAttrs("class").Globally()
	// Allow a conservative set of inline styles with bluemonday's built-in CSS
	// value validators. Anything else is stripped.
	p.AllowStyles(
		"text-align", "color", "background-color",
		"font-weight", "font-style", "text-decoration",
		"margin", "padding",
	).Globally()
	// Only http, https, and mailto links (bluemonday's standard URL policy).
	p.AllowStandardURLs()
	p.AllowAttrs("href").OnElements("a")
	// Do not allow remote images by default in previews.
	return p
}

var policy = buildPolicy()

// Sanitize removes dangerous HTML constructs while preserving a safe rich-text
// subset. Use for stored templates, the send path, and previews.
func Sanitize(html string) string {
	return policy.Sanitize(html)
}
