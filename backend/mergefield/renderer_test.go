package mergefield

import (
	"strings"
	"testing"

	"MailMergeApp/backend/models"
)

func contact() models.Contact {
	return models.Contact{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     "ada@example.com",
		CustomFields: map[string]string{
			"account-manager": "Bob",
			"company":         "Analytical Engines",
			"customer id":     "C-100",
		},
	}
}

func TestRenderStandardAndCustom(t *testing.T) {
	s := NewSchema([]string{"Account-Manager", "Company", "Customer ID"})
	r := NewRenderer(s)

	out, diags := r.Render("Hi {{first_name}} from {{company}} (mgr {{account_manager}})", contact(), false)
	if out != "Hi Ada from Analytical Engines (mgr Bob)" {
		t.Errorf("got %q", out)
	}
	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %+v", diags)
	}
}

func TestRenderLegacySingleBrace(t *testing.T) {
	s := NewSchema([]string{"Company"})
	r := NewRenderer(s)
	// Built-in templates use single-brace fields like {FirstName} and {Company}.
	out, diags := r.Render("Welcome to {Company}, {FirstName}!", contact(), false)
	if out != "Welcome to Analytical Engines, Ada!" {
		t.Errorf("got %q", out)
	}
	if HasErrors(diags) {
		t.Errorf("unexpected errors: %+v", diags)
	}
}

func TestRenderHeaderWithSpaceViaSingleBrace(t *testing.T) {
	s := NewSchema([]string{"Account-Manager", "Customer ID"})
	r := NewRenderer(s)
	out, _ := r.Render("Ref {Customer ID} handled by {Account-Manager}", contact(), false)
	if out != "Ref C-100 handled by Bob" {
		t.Errorf("got %q", out)
	}
}

func TestRenderUnknownFieldIsErrorAndVerbatim(t *testing.T) {
	s := NewSchema([]string{"Company"})
	r := NewRenderer(s)
	out, diags := r.Render("Hello {{nope}}", contact(), false)
	if !strings.Contains(out, "{{nope}}") {
		t.Errorf("unknown field should be left verbatim, got %q", out)
	}
	if !HasErrors(diags) {
		t.Errorf("expected an error diagnostic for unknown field")
	}
	if diags[0].Code != CodeUnknownField {
		t.Errorf("code = %q, want %q", diags[0].Code, CodeUnknownField)
	}
}

func TestRenderMissingValueWarns(t *testing.T) {
	s := NewSchema([]string{"Company", "Region"})
	r := NewRenderer(s)
	out, diags := r.Render("Region: {{region}}", contact(), false)
	if out != "Region: " {
		t.Errorf("got %q", out)
	}
	if HasErrors(diags) {
		t.Errorf("missing value should be a warning, not error: %+v", diags)
	}
	found := false
	for _, d := range diags {
		if d.Code == CodeMissingValue && d.Severity == SeverityWarning {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing_value warning, got %+v", diags)
	}
}

func TestRenderFallback(t *testing.T) {
	s := NewSchema([]string{"Region"})
	r := NewRenderer(s)
	out, diags := r.Render("Hi {{first_name|there}}, region {{region|Global}}", contact(), false)
	if out != "Hi Ada, region Global" {
		t.Errorf("got %q", out)
	}
	// Fallback used => no missing-value warning.
	for _, d := range diags {
		if d.Code == CodeMissingValue {
			t.Errorf("fallback should suppress missing-value warning: %+v", diags)
		}
	}
}

func TestRenderHTMLEscaping(t *testing.T) {
	c := contact()
	c.FirstName = `<script>alert('x')</script>`
	s := NewSchema(nil)
	r := NewRenderer(s)

	out, _ := r.Render("<p>Hi {{first_name}}</p>", c, true)
	if strings.Contains(out, "<script>") {
		t.Errorf("merge value not escaped in HTML context: %q", out)
	}
	if !strings.Contains(out, "&lt;script&gt;") {
		t.Errorf("expected escaped value, got %q", out)
	}
	// Template markup itself must be preserved.
	if !strings.HasPrefix(out, "<p>Hi ") {
		t.Errorf("template markup should not be escaped: %q", out)
	}

	// In plain-text context, no escaping.
	outText, _ := r.Render("Hi {{first_name}}", c, false)
	if !strings.Contains(outText, "<script>") {
		t.Errorf("plain-text context should not escape: %q", outText)
	}
}

func TestRenderCollisionIsError(t *testing.T) {
	s := NewSchema([]string{"Customer ID", "customer-id"})
	r := NewRenderer(s)
	_, diags := r.Render("{{customer_id}}", contact(), false)
	if !HasErrors(diags) {
		t.Errorf("collision should produce an error, got %+v", diags)
	}
}

func TestValidateTemplate(t *testing.T) {
	s := NewSchema([]string{"Company"})
	r := NewRenderer(s)
	diags := r.ValidateTemplate("{{first_name}} at {{company}} — {{unknown_x}}")
	if !HasErrors(diags) {
		t.Fatal("expected error for unknown field")
	}
	errCount := 0
	for _, d := range diags {
		if d.Severity == SeverityError {
			errCount++
		}
	}
	if errCount != 1 {
		t.Errorf("expected exactly 1 error, got %d: %+v", errCount, diags)
	}
}

func TestExtractTokens(t *testing.T) {
	s := NewSchema([]string{"Company"})
	r := NewRenderer(s)
	got := r.ExtractTokens("{{first_name}} {Company} {{first_name}}")
	if len(got) != 2 {
		t.Errorf("expected 2 unique tokens, got %v", got)
	}
}

func TestRenderEquivalenceAcrossCalls(t *testing.T) {
	// Acceptance criterion: same contact + template => identical output every time.
	s := NewSchema([]string{"Company"})
	r := NewRenderer(s)
	tpl := "Hi {{first_name}} <b>{{company}}</b>"
	a, _ := r.Render(tpl, contact(), true)
	b, _ := r.Render(tpl, contact(), true)
	if a != b {
		t.Errorf("render not deterministic: %q vs %q", a, b)
	}
}
