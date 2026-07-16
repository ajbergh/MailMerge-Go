/*
Settings Service - Application Settings Persistence

Settings are stored as a JSON file in the user's app data directory. This
service:
  - Merges stored settings over defaults so new fields adopt sane values.
  - Validates settings before they replace the current object.
  - Versions the schema and migrates older files forward.
  - Writes atomically (temp file + rename).
  - Returns copies of slices/maps so callers cannot mutate internal state.
*/
package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"MailMergeApp/backend/models"
)

// SettingsService manages application settings with persistence to disk.
type SettingsService struct {
	mu           sync.RWMutex
	settings     *models.AppSettings
	settingsPath string
}

// NewSettingsService loads (or initializes) settings.
func NewSettingsService() (*SettingsService, error) {
	dir, err := configDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	ss := &SettingsService{
		settings:     models.DefaultSettings(),
		settingsPath: filepath.Join(dir, "settings.json"),
	}
	if err := ss.load(); err != nil {
		fmt.Printf("Using default settings: %v\n", err)
	}
	return ss, nil
}

// load reads settings, merging over defaults and migrating older schemas.
func (ss *SettingsService) load() error {
	data, err := os.ReadFile(ss.settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ss.persist() // write defaults
		}
		return fmt.Errorf("failed to read settings: %w", err)
	}

	// Merge over defaults: fields absent in the file keep their default value.
	merged := models.DefaultSettings()
	if err := json.Unmarshal(data, merged); err != nil {
		// Corrupted file: back it up and fall back to defaults rather than crash.
		_ = os.Rename(ss.settingsPath, ss.settingsPath+".corrupt")
		ss.settings = models.DefaultSettings()
		_ = ss.persist()
		return fmt.Errorf("corrupted settings file (backed up as .corrupt): %w", err)
	}

	migrated := migrateSettings(merged)
	sanitizeSettings(migrated)
	ss.settings = migrated

	// Persist if we migrated or repaired anything.
	if migrated.SchemaVersion != merged.SchemaVersion {
		return ss.persist()
	}
	return nil
}

// migrateSettings brings older schema versions up to date.
func migrateSettings(s *models.AppSettings) *models.AppSettings {
	if s.SchemaVersion < 1 {
		if s.DuplicatePolicy == "" {
			s.DuplicatePolicy = "keep_first"
		}
		s.SchemaVersion = 1
	}
	// Future migrations: if s.SchemaVersion < 2 { ... }
	s.SchemaVersion = models.CurrentSettingsSchemaVersion
	return s
}

// persist writes the current settings atomically.
func (ss *SettingsService) persist() error {
	return atomicWriteJSON(ss.settingsPath, ss.settings)
}

// validateSettings returns an error if any field is out of range/invalid.
func validateSettings(s models.AppSettings) error {
	switch s.Theme {
	case "light", "dark", "system":
	default:
		return fmt.Errorf("invalid theme: %s", s.Theme)
	}
	switch s.DefaultFormat {
	case "html", "plaintext":
	default:
		return fmt.Errorf("invalid default format: %s", s.DefaultFormat)
	}
	if s.SendingDelay < 0 || s.SendingDelay > 60000 {
		return fmt.Errorf("sending delay out of range (0-60000ms): %d", s.SendingDelay)
	}
	switch s.DuplicatePolicy {
	case "", "keep_first", "keep_last", "exclude_all", "keep_all", "manual":
	default:
		return fmt.Errorf("invalid duplicate policy: %s", s.DuplicatePolicy)
	}
	return nil
}

// sanitizeSettings clamps/normalizes values in place (best-effort repair).
func sanitizeSettings(s *models.AppSettings) {
	if s.Theme != "light" && s.Theme != "dark" && s.Theme != "system" {
		s.Theme = "light"
	}
	if s.DefaultFormat != "html" && s.DefaultFormat != "plaintext" {
		s.DefaultFormat = "html"
	}
	if s.SendingDelay < 0 {
		s.SendingDelay = 0
	}
	if s.SendingDelay > 60000 {
		s.SendingDelay = 60000
	}
	if s.DuplicatePolicy == "" {
		s.DuplicatePolicy = "keep_first"
	}
	if s.RecentFiles == nil {
		s.RecentFiles = []string{}
	}
}

// GetSettings returns a deep copy of the current settings.
func (ss *SettingsService) GetSettings() *models.AppSettings {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	cp := *ss.settings
	cp.RecentFiles = append([]string(nil), ss.settings.RecentFiles...)
	return &cp
}

// UpdateSettings validates and replaces the settings.
func (ss *SettingsService) UpdateSettings(settings models.AppSettings) error {
	if err := validateSettings(settings); err != nil {
		return err
	}
	sanitizeSettings(&settings)
	settings.SchemaVersion = models.CurrentSettingsSchemaVersion

	ss.mu.Lock()
	defer ss.mu.Unlock()
	// Preserve recent files if the incoming payload omitted them.
	if settings.RecentFiles == nil {
		settings.RecentFiles = append([]string(nil), ss.settings.RecentFiles...)
	}
	ss.settings = &settings
	return ss.persist()
}

// GetTheme returns the current theme setting.
func (ss *SettingsService) GetTheme() string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.settings.Theme
}

// SetTheme updates the theme and saves.
func (ss *SettingsService) SetTheme(theme string) error {
	if theme != "light" && theme != "dark" && theme != "system" {
		return fmt.Errorf("invalid theme: %s (must be light, dark, or system)", theme)
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.settings.Theme = theme
	return ss.persist()
}

// GetSendingDelay returns the delay between emails in milliseconds.
func (ss *SettingsService) GetSendingDelay() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.settings.SendingDelay
}

// SetSendingDelay updates the sending delay and saves.
func (ss *SettingsService) SetSendingDelay(delay int) error {
	if delay < 0 || delay > 60000 {
		return fmt.Errorf("invalid delay: %d (must be 0-60000ms)", delay)
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.settings.SendingDelay = delay
	return ss.persist()
}

// pathKey normalizes a file path for duplicate comparison (case-insensitive on
// Windows).
func pathKey(p string) string {
	c := filepath.Clean(p)
	if filepath.Separator == '\\' {
		return strings.ToLower(c)
	}
	return c
}

// GetRecentFiles returns a copy of the recent files that still exist on disk.
func (ss *SettingsService) GetRecentFiles() []string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	out := make([]string, 0, len(ss.settings.RecentFiles))
	for _, f := range ss.settings.RecentFiles {
		if _, err := os.Stat(f); err == nil {
			out = append(out, f)
		}
	}
	return out
}

// AddRecentFile adds a file to the front of the recent list (max 10), removing
// any prior occurrence (normalized comparison).
func (ss *SettingsService) AddRecentFile(filePath string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	key := pathKey(filePath)
	recent := make([]string, 0, len(ss.settings.RecentFiles)+1)
	recent = append(recent, filePath)
	for _, f := range ss.settings.RecentFiles {
		if pathKey(f) != key {
			recent = append(recent, f)
		}
	}
	if len(recent) > 10 {
		recent = recent[:10]
	}
	ss.settings.RecentFiles = recent
	return ss.persist()
}

// ClearRecentFiles empties the recent files list.
func (ss *SettingsService) ClearRecentFiles() error {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.settings.RecentFiles = []string{}
	return ss.persist()
}
