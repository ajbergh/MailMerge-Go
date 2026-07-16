/*
Package htmlutil contains HTML helpers shared by the renderer and senders:

  - NormalizeForEmail neutralizes ReactQuill's default <p> margins so bodies
    render consistently in email clients (notably Outlook).
  - Sanitize strips dangerous constructs (scripts, event handlers, unsafe URLs)
    from authored template HTML. See sanitize.go.
*/
package htmlutil

import (
	"regexp"
	"strings"
)

var pTagPattern = regexp.MustCompile(`<p([^>]*)>`)

// NormalizeForEmail adds inline margin:0 styles to <p> tags and collapses empty
// paragraphs to <br>, matching the previous MergeService behavior.
func NormalizeForEmail(html string) string {
	normalized := pTagPattern.ReplaceAllStringFunc(html, func(match string) string {
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
