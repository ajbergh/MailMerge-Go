package campaign

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"MailMergeApp/backend/email"
	"MailMergeApp/backend/models"
)

func TestPreflightMissingAttachmentBlocks(t *testing.T) {
	c := baseCampaign("a@x.com")
	c.Attachments = []string{filepath.Join(t.TempDir(), "does-not-exist.pdf")}
	pf := NewPreflighter().Preflight(context.Background(), c, NewFakeSender())
	if pf.CanSend {
		t.Fatal("missing attachment should block send")
	}
	foundMissing := false
	for _, e := range pf.Errors {
		if e.Code == CodeMissingAttach {
			foundMissing = true
		}
	}
	if !foundMissing {
		t.Errorf("expected missing_attachment error, got %+v", pf.Errors)
	}
}

func TestPreflightDirectoryAttachmentBlocks(t *testing.T) {
	dir := t.TempDir()
	c := baseCampaign("a@x.com")
	c.Attachments = []string{dir}
	pf := NewPreflighter().Preflight(context.Background(), c, NewFakeSender())
	if pf.CanSend {
		t.Fatal("directory attachment should block send")
	}
}

func TestPreflightValidAttachmentPasses(t *testing.T) {
	f := filepath.Join(t.TempDir(), "report.pdf")
	if err := os.WriteFile(f, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := baseCampaign("a@x.com")
	c.Attachments = []string{f}
	pf := NewPreflighter().Preflight(context.Background(), c, NewFakeSender())
	if !pf.CanSend {
		t.Fatalf("valid attachment should pass, errors: %+v", pf.Errors)
	}
}

func TestPreflightPersonalizedAttachment(t *testing.T) {
	dir := t.TempDir()
	// Personalized path C:\...\{{customer_id}}.pdf resolved per contact.
	if err := os.WriteFile(filepath.Join(dir, "C-1.pdf"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := Campaign{
		Headers:         []string{"Customer ID"},
		Contacts:        []models.Contact{{ID: "1", Email: "a@x.com", FirstName: "A", CustomFields: map[string]string{"customer id": "C-1"}}},
		SubjectTemplate: "Hi",
		BodyTemplate:    "Body",
		Attachments:     []string{filepath.Join(dir, "{{customer_id}}.pdf")},
		Options:         DefaultSendOptions(),
		DuplicatePolicy: email.PolicyKeepFirst,
	}
	pf := NewPreflighter().Preflight(context.Background(), c, NewFakeSender())
	if !pf.CanSend {
		t.Fatalf("personalized attachment should resolve and pass, errors: %+v", pf.Errors)
	}
}

func TestPreflightBlankBodyBlocks(t *testing.T) {
	c := baseCampaign("a@x.com")
	c.BodyTemplate = "   "
	pf := NewPreflighter().Preflight(context.Background(), c, NewFakeSender())
	if pf.CanSend {
		t.Fatal("blank body should block send")
	}
}

func TestPreflightEstimatedDuration(t *testing.T) {
	c := baseCampaign("a@x.com", "b@x.com", "c@x.com")
	pf := NewPreflighter().Preflight(context.Background(), c, NewFakeSender())
	if pf.EstimatedDuration <= 0 {
		t.Errorf("expected positive estimated duration, got %v", pf.EstimatedDuration)
	}
	if pf.RecipientCount != 3 {
		t.Errorf("recipient count = %d, want 3", pf.RecipientCount)
	}
}
