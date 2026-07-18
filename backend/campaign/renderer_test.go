package campaign

import (
	"os"
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
	if len(testMsg.To) == 0 || !strings.Contains(testMsg.To[0], "qa@example.com") {
		t.Errorf("test To = %v, want qa override", testMsg.To)
	}
}

func TestRenderTestSendDoesNotOverwriteEmailByDefault(t *testing.T) {
	c := richCampaign()
	c.BodyTemplate = "Your email is {{email}}"
	c.ToOverride = "qa@example.com"
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

func TestRenderHTMLBodySanitizesAuthoredTemplate(t *testing.T) {
	c := richCampaign()
	c.BodyTemplate = `<p>Hi {{first_name}}</p><script>steal()</script><a href="javascript:evil()">x</a>`
	msg := NewRenderer(c).Render(c, richContact())
	if strings.Contains(strings.ToLower(msg.HTMLBody), "<script") {
		t.Errorf("script not stripped: %q", msg.HTMLBody)
	}
	if strings.Contains(strings.ToLower(msg.HTMLBody), "javascript:") {
		t.Errorf("javascript URL not stripped: %q", msg.HTMLBody)
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
	if !strings.Contains(msg.HTMLBody, "<p") {
		t.Errorf("template markup should be preserved: %q", msg.HTMLBody)
	}
}

func TestRendererCachesResolvedAttachmentMetadata(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "static-attachment-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	c := richCampaign()
	c.Attachments = []string{path}
	calls := 0
	r := NewRenderer(c).withStat(func(p string) (os.FileInfo, error) {
		calls++
		return os.Stat(p)
	})

	r.Render(c, richContact())
	ct2 := richContact()
	ct2.ID = "2"
	ct2.Email = "grace@example.com"
	r.Render(c, ct2)

	if calls != 1 {
		t.Fatalf("static attachment stat calls = %d, want 1", calls)
	}
}
