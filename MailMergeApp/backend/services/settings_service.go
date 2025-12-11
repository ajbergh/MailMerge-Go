/*
Settings Service - Application Settings Persistence

Phase 2 (v1.3): New service for managing user preferences and app settings.

This service handles:
  - Loading settings from disk on startup
  - Saving settings when changed
  - Default settings initialization
  - Recent files management

Settings are stored as a JSON file in the user's app data directory.
*/
package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"MailMergeApp/backend/models"
)

// SettingsService manages application settings with persistence to disk.
type SettingsService struct {
	settings     *models.AppSettings // Current settings
	settingsPath string              // Path to settings file
}

// NewSettingsService creates a new SettingsService and loads existing settings.
// Creates the storage directory and default settings if they don't exist.
func NewSettingsService() (*SettingsService, error) {
	// Get user's app data directory
	appData, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	configDir := filepath.Join(appData, "MailMergeGo")

	// Create config directory if needed
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	ss := &SettingsService{
		settings:     models.DefaultSettings(),
		settingsPath: filepath.Join(configDir, "settings.json"),
	}

	// Load existing settings if available
	if err := ss.loadSettings(); err != nil {
		// Non-fatal: use default settings
		fmt.Printf("Using default settings: %v\n", err)
	}

	return ss, nil
}

// loadSettings reads settings from disk.
func (ss *SettingsService) loadSettings() error {
	data, err := os.ReadFile(ss.settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No settings file yet, save defaults
			return ss.saveSettings()
		}
		return fmt.Errorf("failed to read settings: %w", err)
	}

	var settings models.AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	ss.settings = &settings
	return nil
}

// saveSettings writes current settings to disk.
func (ss *SettingsService) saveSettings() error {
	data, err := json.MarshalIndent(ss.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(ss.settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	return nil
}

// GetSettings returns the current application settings.
func (ss *SettingsService) GetSettings() *models.AppSettings {
	// Return a copy to prevent external modification
	copy := *ss.settings
	return &copy
}

// UpdateSettings updates all settings and saves to disk.
func (ss *SettingsService) UpdateSettings(settings models.AppSettings) error {
	ss.settings = &settings
	return ss.saveSettings()
}

// GetTheme returns the current theme setting.
func (ss *SettingsService) GetTheme() string {
	return ss.settings.Theme
}

// SetTheme updates the theme and saves to disk.
func (ss *SettingsService) SetTheme(theme string) error {
	if theme != "light" && theme != "dark" && theme != "system" {
		return fmt.Errorf("invalid theme: %s (must be light, dark, or system)", theme)
	}
	ss.settings.Theme = theme
	return ss.saveSettings()
}

// GetSendingDelay returns the delay between emails in milliseconds.
func (ss *SettingsService) GetSendingDelay() int {
	return ss.settings.SendingDelay
}

// SetSendingDelay updates the sending delay and saves to disk.
func (ss *SettingsService) SetSendingDelay(delay int) error {
	if delay < 100 || delay > 5000 {
		return fmt.Errorf("invalid delay: %d (must be 100-5000ms)", delay)
	}
	ss.settings.SendingDelay = delay
	return ss.saveSettings()
}

// GetRecentFiles returns the list of recently opened contact files.
func (ss *SettingsService) GetRecentFiles() []string {
	return ss.settings.RecentFiles
}

// AddRecentFile adds a file to the recent files list.
// Maintains a maximum of 10 files, removing the oldest if needed.
func (ss *SettingsService) AddRecentFile(filePath string) error {
	// Remove if already in list (to move to front)
	recentFiles := make([]string, 0, len(ss.settings.RecentFiles))
	for _, f := range ss.settings.RecentFiles {
		if f != filePath {
			recentFiles = append(recentFiles, f)
		}
	}

	// Add to front of list
	recentFiles = append([]string{filePath}, recentFiles...)

	// Keep only last 10
	if len(recentFiles) > 10 {
		recentFiles = recentFiles[:10]
	}

	ss.settings.RecentFiles = recentFiles
	return ss.saveSettings()
}

// ClearRecentFiles removes all recent files.
func (ss *SettingsService) ClearRecentFiles() error {
	ss.settings.RecentFiles = []string{}
	return ss.saveSettings()
}
