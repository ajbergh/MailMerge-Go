package htmlutil

import (
	"strings"
	"testing"
)

func TestSanitizeStripsScripts(t *testing.T) {
	cases := []string{
		`<script>alert('x')</script>`,
		`<img src=x onerror="alert(1)">`,
		`<a href="javascript:alert(1)">click</a>`,
		`<iframe src="http://evil"></iframe>`,
		`<object data="x"></object>`,
		`<div onclick="steal()">hi</div>`,
	}
	for _, in := range cases {
		out := Sanitize(in)
		lower := strings.ToLower(out)
		if strings.Contains(lower, "<script") || strings.Contains(lower, "onerror") ||
			strings.Contains(lower, "javascript:") || strings.Contains(lower, "<iframe") ||
			strings.Contains(lower, "<object") || strings.Contains(lower, "onclick") {
			t.Errorf("Sanitize(%q) left dangerous content: %q", in, out)
		}
	}
}

func TestSanitizePreservesSafeRichText(t *testing.T) {
	in := `<p style="text-align:center"><strong>Hi</strong> <a href="https://example.com">link</a></p><ul><li>one</li></ul>`
	out := Sanitize(in)
	for _, want := range []string{"<strong>", "<a", "https://example.com", "<ul>", "<li>", "text-align"} {
		if !strings.Contains(out, want) {
			t.Errorf("Sanitize removed safe content %q; got %q", want, out)
		}
	}
}

func TestNormalizeForEmail(t *testing.T) {
	out := NormalizeForEmail("<p>Hello</p><p><br></p>")
	if !strings.Contains(out, "margin:0") {
		t.Errorf("expected inline margin style, got %q", out)
	}
	if !strings.Contains(out, "<br>") {
		t.Errorf("expected empty paragraph collapsed to <br>, got %q", out)
	}
}
