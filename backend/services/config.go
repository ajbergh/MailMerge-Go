package services

import (
	"os"
	"path/filepath"
)

// testConfigDir, when non-empty, overrides the application config directory.
// Tests set it (same package) to a temp dir so persistence can be exercised
// without touching the real user profile.
var testConfigDir string

// configDir returns the MailMergeGo config directory, creating it if needed.
func configDir() (string, error) {
	base := testConfigDir
	if base == "" {
		appData, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(appData, "MailMergeGo")
	}
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", err
	}
	return base, nil
}
