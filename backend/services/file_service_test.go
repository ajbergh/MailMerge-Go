package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParseCSVStandard(t *testing.T) {
	fs := NewFileService()
	p := writeTemp(t, "c.csv", "FirstName,LastName,Email\nAda,Lovelace,ada@example.com\n")
	res, err := fs.ParseContactFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Contacts) != 1 {
		t.Fatalf("got %d contacts", len(res.Contacts))
	}
	c := res.Contacts[0]
	if c.FirstName != "Ada" || c.Email != "ada@example.com" {
		t.Errorf("bad contact: %+v", c)
	}
	if c.ID == "" {
		t.Error("contact should have a stable ID")
	}
}

func TestParseCSVBOM(t *testing.T) {
	fs := NewFileService()
	p := writeTemp(t, "bom.csv", "\xef\xbb\xbfFirstName,Email\nAda,ada@example.com\n")
	res, err := fs.ParseContactFile(p)
	if err != nil {
		t.Fatalf("BOM header should be handled: %v", err)
	}
	if len(res.Contacts) != 1 || res.Headers[0] != "FirstName" {
		t.Errorf("BOM not stripped: headers=%v", res.Headers)
	}
}

func TestParseCSVCustomFieldsAndUnicode(t *testing.T) {
	fs := NewFileService()
	p := writeTemp(t, "u.csv", "Email,Account-Manager,Región\na@example.com,Bob,Norte\n")
	res, err := fs.ParseContactFile(p)
	if err != nil {
		t.Fatal(err)
	}
	c := res.Contacts[0]
	if c.CustomFields["account-manager"] != "Bob" {
		t.Errorf("custom field not captured: %+v", c.CustomFields)
	}
	if c.CustomFields["región"] != "Norte" {
		t.Errorf("unicode header field not captured: %+v", c.CustomFields)
	}
}

func TestParseCSVMissingEmailColumn(t *testing.T) {
	fs := NewFileService()
	p := writeTemp(t, "noemail.csv", "FirstName,LastName\nAda,Lovelace\n")
	if _, err := fs.ParseContactFile(p); err == nil {
		t.Error("expected error for missing Email column")
	}
}

func TestParseCSVInvalidEmailAndBlankRows(t *testing.T) {
	fs := NewFileService()
	p := writeTemp(t, "mix.csv", "Email\ngood@example.com\nnot-an-email\n\n   \nbad@\n")
	res, err := fs.ParseContactFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Contacts) != 1 {
		t.Errorf("only the valid address should be imported, got %d", len(res.Contacts))
	}
	if len(res.Errors) == 0 {
		t.Error("expected row-level errors for invalid emails")
	}
}

func TestParseCSVAltEmailHeader(t *testing.T) {
	fs := NewFileService()
	p := writeTemp(t, "alt.csv", "E-mail Address\nada@example.com\n")
	res, err := fs.ParseContactFile(p)
	if err != nil {
		t.Fatalf("alternate email header should be recognized: %v", err)
	}
	if len(res.Contacts) != 1 {
		t.Errorf("got %d contacts", len(res.Contacts))
	}
}

func TestParseCSVDuplicateHeaderWarns(t *testing.T) {
	fs := NewFileService()
	p := writeTemp(t, "dup.csv", "Email,Company,company\na@example.com,X,Y\n")
	res, err := fs.ParseContactFile(p)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, w := range res.Warnings {
		if strings.Contains(strings.ToLower(w), "duplicate column") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected duplicate-header warning, got %v", res.Warnings)
	}
}

func TestParseCSVDuplicateEmailsDetected(t *testing.T) {
	fs := NewFileService()
	p := writeTemp(t, "dupe.csv", "Email\na@example.com\nA@example.com\nb@example.com\n")
	res, err := fs.ParseContactFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Duplicates) != 1 {
		t.Errorf("expected 1 duplicate address group, got %v", res.Duplicates)
	}
}

func TestValidateFileRejectsUnknownExt(t *testing.T) {
	fs := NewFileService()
	p := writeTemp(t, "x.txt", "hello")
	if err := fs.ValidateFile(p); err == nil {
		t.Error("expected error for unsupported extension")
	}
}
