package services

import (
	"os"
	"path/filepath"
	"testing"

	"MailMergeApp/backend/models"
)

// useTempConfig points the services at a temp config dir for the test.
func useTempConfig(t *testing.T) {
	t.Helper()
	testConfigDir = t.TempDir()
	t.Cleanup(func() { testConfigDir = "" })
}

// ---- Settings ----

func TestSettingsDefaultsAndPersistence(t *testing.T) {
	useTempConfig(t)
	ss, err := NewSettingsService()
	if err != nil {
		t.Fatal(err)
	}
	got := ss.GetSettings()
	if got.Theme != "light" || got.SchemaVersion != models.CurrentSettingsSchemaVersion {
		t.Errorf("unexpected defaults: %+v", got)
	}
	if err := ss.SetSendingDelay(1200); err != nil {
		t.Fatal(err)
	}
	// Reload from disk.
	ss2, err := NewSettingsService()
	if err != nil {
		t.Fatal(err)
	}
	if ss2.GetSendingDelay() != 1200 {
		t.Errorf("delay not persisted, got %d", ss2.GetSendingDelay())
	}
}

func TestSettingsValidationRejectsBad(t *testing.T) {
	useTempConfig(t)
	ss, _ := NewSettingsService()
	bad := models.DefaultSettings()
	bad.Theme = "neon"
	if err := ss.UpdateSettings(*bad); err == nil {
		t.Error("expected validation error for invalid theme")
	}
	bad = models.DefaultSettings()
	bad.SendingDelay = 999999
	if err := ss.UpdateSettings(*bad); err == nil {
		t.Error("expected validation error for out-of-range delay")
	}
}

func TestSettingsCorruptedFileFallsBack(t *testing.T) {
	useTempConfig(t)
	path := filepath.Join(testConfigDir, "settings.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0o644); err != nil {
		t.Fatal(err)
	}
	ss, err := NewSettingsService()
	if err != nil {
		t.Fatal(err)
	}
	if ss.GetSettings().Theme != "light" {
		t.Error("should fall back to defaults on corrupted file")
	}
	if _, err := os.Stat(path + ".corrupt"); err != nil {
		t.Error("corrupted file should be backed up")
	}
}

func TestGetRecentFilesReturnsCopyAndPrunes(t *testing.T) {
	useTempConfig(t)
	ss, _ := NewSettingsService()
	existing := filepath.Join(testConfigDir, "real.csv")
	os.WriteFile(existing, []byte("x"), 0o644)
	ss.AddRecentFile(existing)
	ss.AddRecentFile(filepath.Join(testConfigDir, "ghost.csv")) // does not exist

	got := ss.GetRecentFiles()
	if len(got) != 1 || got[0] != existing {
		t.Errorf("GetRecentFiles should prune missing files, got %v", got)
	}
	// Mutating the returned slice must not affect internal state.
	got[0] = "injected"
	if len(ss.GetRecentFiles()) != 1 {
		t.Error("returned slice shares internal backing storage")
	}
}

func TestAddRecentFileDedupesNormalized(t *testing.T) {
	useTempConfig(t)
	ss, _ := NewSettingsService()
	f := filepath.Join(testConfigDir, "a.csv")
	os.WriteFile(f, []byte("x"), 0o644)
	ss.AddRecentFile(f)
	ss.AddRecentFile(f) // same path again
	if n := len(ss.GetRecentFiles()); n != 1 {
		t.Errorf("expected 1 recent file after dedupe, got %d", n)
	}
}

// ---- Templates ----

func TestTemplateSaveRejectsCraftedID(t *testing.T) {
	useTempConfig(t)
	ts, err := NewTemplateService()
	if err != nil {
		t.Fatal(err)
	}
	// Path traversal / built-in overwrite attempts must be rejected.
	for _, id := range []string{"../evil", "a/b", "builtin-welcome", "..\\x"} {
		_, err := ts.SaveTemplate(models.EmailTemplate{ID: id, Name: "x", Subject: "s", Body: "b"})
		if err == nil {
			t.Errorf("SaveTemplate should reject crafted id %q", id)
		}
	}
}

func TestTemplateSaveAndUpdate(t *testing.T) {
	useTempConfig(t)
	ts, _ := NewTemplateService()
	saved, err := ts.SaveTemplate(models.EmailTemplate{Name: "Hi", Subject: "S", Body: "B"})
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID == "" || saved.IsBuiltIn {
		t.Errorf("unexpected saved template: %+v", saved)
	}
	// Update the existing template.
	saved.Subject = "S2"
	updated, err := ts.SaveTemplate(*saved)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Subject != "S2" {
		t.Error("update did not persist")
	}
	// The template file must exist on disk.
	if _, err := os.Stat(filepath.Join(testConfigDir, "templates", saved.ID+".json")); err != nil {
		t.Errorf("template file missing: %v", err)
	}
}

func TestTemplateCannotDeleteBuiltIn(t *testing.T) {
	useTempConfig(t)
	ts, _ := NewTemplateService()
	if err := ts.DeleteTemplate("builtin-welcome"); err == nil {
		t.Error("deleting a built-in template should fail")
	}
}

func TestTemplateValidationRequiresName(t *testing.T) {
	useTempConfig(t)
	ts, _ := NewTemplateService()
	if _, err := ts.SaveTemplate(models.EmailTemplate{Subject: "s", Body: "b"}); err == nil {
		t.Error("expected error for missing name")
	}
}

// ---- Suppression ----

func TestSuppressionAddContainsRemove(t *testing.T) {
	useTempConfig(t)
	s, err := NewSuppressionService()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Add("Foo@Example.com"); err != nil {
		t.Fatal(err)
	}
	if !s.Contains("foo@example.com") {
		t.Error("normalized contains should match")
	}
	if err := s.Add("not-an-email"); err == nil {
		t.Error("invalid address should be rejected")
	}
	if err := s.Remove("foo@example.com"); err != nil {
		t.Fatal(err)
	}
	if s.Contains("foo@example.com") {
		t.Error("address should be removed")
	}
}

func TestSuppressionPersistenceAndSet(t *testing.T) {
	useTempConfig(t)
	s, _ := NewSuppressionService()
	s.Add("a@x.com")
	s.Add("b@x.com")
	s2, _ := NewSuppressionService()
	set := s2.Set()
	if !set["a@x.com"] || !set["b@x.com"] {
		t.Errorf("suppression not persisted/loaded: %v", set)
	}
}
