package campaign

import (
	"strings"
	"testing"

	"MailMergeApp/backend/email"
	"MailMergeApp/backend/models"
)

func richCampaign() Campaign {
	return Campaign{
		Headers:         []string{"Company", "Customer ID"},
		SubjectTemplate: "Hello {{first_name}} at {{company}}",
		BodyTemplate:    "<p>Hi {{first_name}}, ref {{customer_id}}</p>",
		IsHTML:          true,
		CCTemplate:      "mgr@x.com",
		Options:         DefaultSendOptions(),
		DuplicatePolicy: email.PolicyKeepFirst,
	}
}

func richContact() models.Contact {
	return models.Contact{
		ID: "1", FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com",
		CustomFields: map[string]string{"company": "Analytical", "customer id": "C-100"},
	}
}

// Acceptance criterion: preview, test send, and bulk send must generate an
// equivalent rendered subject, body, CC, BCC, and attachments for the same
// contact and template. All three use campaign.Renderer.Render.
func TestRenderEquivalencePreviewTestBulk(t *testing.T) {
	c := richCampaign()
	r := NewRenderer(c)

	preview := r.Render(c, richContact())

	// Test send: same campaign but with an explicit To override that does not
	// overwrite the {{email}} merge value.
	test := c
	test.ToOverride = "qa@example.com"
	testMsg := NewRenderer(test).Render(test, richContact())

	bulk := r.Render(c, richContact())

	if preview.Subject != bulk.Subject || preview.Subject != testMsg.Subject {
		t.Errorf("subjects differ: %q / %q / %q", preview.Subject, testMsg.Subject, bulk.Subject)
	}
	if preview.HTMLBody != bulk.HTMLBody || preview.HTMLBody != testMsg.HTMLBody {
		t.Errorf("bodies differ")
	}
	if strings.Join(preview.CC, ",") != strings.Join(bulk.CC, ",") {
		t.Errorf("CC differ")
	}
	// The test send goes to the override, but merge data (body) is unchanged.
	if len(testMsg.To) == 0 || !strings.Contains(testMsg.To[0], "qa@example.com") {
		t.Errorf("test To = %v, want qa override", testMsg.To)
	}
}

func TestRenderTestSendDoesNotOverwriteEmailByDefault(t *testing.T) {
	c := richCampaign()
	c.BodyTemplate = "Your email is {{email}}"
	c.ToOverride = "qa@example.com" // OverrideEmail defaults to false
	msg := NewRenderer(c).Render(c, richContact())
	if !strings.Contains(msg.TextBody, "ada@example.com") {
		t.Errorf("{{email}} should still render the contact address, got %q", msg.TextBody)
	}
	if !strings.Contains(msg.To[0], "qa@example.com") {
		t.Errorf("To should be the override, got %v", msg.To)
	}
}

func TestRenderTestSendOverwriteEmailWhenRequested(t *testing.T) {
	c := richCampaign()
	c.BodyTemplate = "Your email is {{email}}"
	c.ToOverride = "qa@example.com"
	c.OverrideEmail = true
	msg := NewRenderer(c).Render(c, richContact())
	if !strings.Contains(msg.TextBody, "qa@example.com") {
		t.Errorf("with OverrideEmail, {{email}} should render override, got %q", msg.TextBody)
	}
}

func TestRenderHTMLBodyEscapesMergeValues(t *testing.T) {
	c := richCampaign()
	ct := richContact()
	ct.FirstName = "<b>x</b>"
	msg := NewRenderer(c).Render(c, ct)
	if strings.Contains(msg.HTMLBody, "<b>x</b>") {
		t.Errorf("merge value not escaped in HTML body: %q", msg.HTMLBody)
	}
	if !strings.Contains(msg.HTMLBody, "&lt;b&gt;") {
		t.Errorf("expected escaped value in HTML body: %q", msg.HTMLBody)
	}
	// The authored <p> markup is preserved.
	if !strings.Contains(msg.HTMLBody, "<p>") {
		t.Errorf("template markup should be preserved: %q", msg.HTMLBody)
	}
}
