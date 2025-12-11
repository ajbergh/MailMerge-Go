/*
Template Service - Email Template Persistence

Phase 2 (v1.3): New service for managing saved email templates.

This service handles:
  - Loading templates from disk on startup
  - Saving templates to user's app data folder
  - CRUD operations for templates
  - Built-in starter templates

Templates are stored as JSON files in the user's app data directory.
*/
package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"MailMergeApp/backend/models"

	"github.com/google/uuid"
)

// TemplateService manages email templates with persistence to disk.
// All templates are stored in the user's app data folder.
type TemplateService struct {
	templates   map[string]*models.EmailTemplate // In-memory template cache
	storagePath string                           // Path to templates directory
}

// NewTemplateService creates a new TemplateService and loads existing templates.
// Creates the storage directory if it doesn't exist.
func NewTemplateService() (*TemplateService, error) {
	// Get user's app data directory
	appData, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	storagePath := filepath.Join(appData, "MailMergeGo", "templates")

	// Create storage directory if needed
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}

	ts := &TemplateService{
		templates:   make(map[string]*models.EmailTemplate),
		storagePath: storagePath,
	}

	// Load existing templates
	if err := ts.loadTemplates(); err != nil {
		// Non-fatal: log error but continue with empty templates
		fmt.Printf("Warning: failed to load templates: %v\n", err)
	}

	// Add built-in templates if none exist
	if len(ts.templates) == 0 {
		ts.addBuiltInTemplates()
	}

	return ts, nil
}

// loadTemplates reads all template files from disk into memory.
func (ts *TemplateService) loadTemplates() error {
	files, err := os.ReadDir(ts.storagePath)
	if err != nil {
		return fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(ts.storagePath, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Warning: failed to read template file %s: %v\n", file.Name(), err)
			continue
		}

		var template models.EmailTemplate
		if err := json.Unmarshal(data, &template); err != nil {
			fmt.Printf("Warning: failed to parse template file %s: %v\n", file.Name(), err)
			continue
		}

		ts.templates[template.ID] = &template
	}

	return nil
}

// addBuiltInTemplates adds default starter templates on first run.
func (ts *TemplateService) addBuiltInTemplates() {
	builtIns := []models.EmailTemplate{
		{
			ID:        "builtin-welcome",
			Name:      "Welcome Email",
			Subject:   "Welcome to {Company}, {FirstName}!",
			Body:      "<h2>Welcome, {FirstName}!</h2><p>We're excited to have you join us. If you have any questions, please don't hesitate to reach out.</p><p>Best regards,<br>The Team</p>",
			IsHTML:    true,
			IsBuiltIn: true,
		},
		{
			ID:        "builtin-newsletter",
			Name:      "Newsletter Template",
			Subject:   "Monthly Newsletter - {Month} Update",
			Body:      "<h1>Monthly Update</h1><p>Hi {FirstName},</p><p>Here's what's new this month...</p><h2>Highlights</h2><ul><li>Feature 1</li><li>Feature 2</li><li>Feature 3</li></ul><p>Thanks for being with us!</p>",
			IsHTML:    true,
			IsBuiltIn: true,
		},
		{
			ID:        "builtin-meeting",
			Name:      "Meeting Request",
			Subject:   "Meeting Request: {Topic}",
			Body:      "<p>Hi {FirstName},</p><p>I'd like to schedule a meeting to discuss {Topic}.</p><p><strong>Proposed Time:</strong> {Date} at {Time}</p><p><strong>Location:</strong> {Location}</p><p>Please let me know if this works for you.</p><p>Best,<br>{SenderName}</p>",
			IsHTML:    true,
			IsBuiltIn: true,
		},
		{
			ID:        "builtin-followup",
			Name:      "Follow-up Email",
			Subject:   "Following up on our conversation",
			Body:      "<p>Hi {FirstName},</p><p>I wanted to follow up on our recent conversation about {Topic}.</p><p>Please let me know if you have any questions or if there's anything else I can help with.</p><p>Best regards,<br>{SenderName}</p>",
			IsHTML:    true,
			IsBuiltIn: true,
		},
		{
			ID:        "builtin-thankyou",
			Name:      "Thank You Note",
			Subject:   "Thank you, {FirstName}!",
			Body:      "<p>Dear {FirstName},</p><p>Thank you so much for {Reason}. We truly appreciate your support!</p><p>Warm regards,<br>{SenderName}</p>",
			IsHTML:    true,
			IsBuiltIn: true,
		},
		{
			ID:        "builtin-plain",
			Name:      "Simple Plain Text",
			Subject:   "Hi {FirstName}",
			Body:      "Hi {FirstName},\n\nI hope this email finds you well.\n\n[Your message here]\n\nBest regards,\n{SenderName}",
			IsHTML:    false,
			IsBuiltIn: true,
		},
	}

	now := time.Now()
	for i := range builtIns {
		builtIns[i].CreatedAt = now
		builtIns[i].UpdatedAt = now
		ts.templates[builtIns[i].ID] = &builtIns[i]
	}

	// Save built-in templates to disk
	ts.saveAllTemplates()
}

// GetAllTemplates returns all templates sorted by name.
// Built-in templates appear first, then user templates alphabetically.
func (ts *TemplateService) GetAllTemplates() []models.EmailTemplate {
	templates := make([]models.EmailTemplate, 0, len(ts.templates))
	for _, t := range ts.templates {
		templates = append(templates, *t)
	}

	// Sort: built-in first, then alphabetically by name
	sort.Slice(templates, func(i, j int) bool {
		if templates[i].IsBuiltIn != templates[j].IsBuiltIn {
			return templates[i].IsBuiltIn
		}
		return templates[i].Name < templates[j].Name
	})

	return templates
}

// GetTemplate returns a template by ID, or nil if not found.
func (ts *TemplateService) GetTemplate(id string) *models.EmailTemplate {
	if t, ok := ts.templates[id]; ok {
		copy := *t
		return &copy
	}
	return nil
}

// SaveTemplate creates or updates a template.
// Returns the saved template with updated timestamps.
func (ts *TemplateService) SaveTemplate(template models.EmailTemplate) (*models.EmailTemplate, error) {
	now := time.Now()

	// Generate ID if new template
	if template.ID == "" {
		template.ID = uuid.New().String()
		template.CreatedAt = now
	}
	template.UpdatedAt = now
	template.IsBuiltIn = false // User templates are never built-in

	// Save to memory
	ts.templates[template.ID] = &template

	// Save to disk
	if err := ts.saveTemplate(&template); err != nil {
		return nil, fmt.Errorf("failed to save template: %w", err)
	}

	return &template, nil
}

// DeleteTemplate removes a template by ID.
// Built-in templates cannot be deleted, only hidden (not implemented yet).
func (ts *TemplateService) DeleteTemplate(id string) error {
	template, ok := ts.templates[id]
	if !ok {
		return fmt.Errorf("template not found: %s", id)
	}

	if template.IsBuiltIn {
		return fmt.Errorf("cannot delete built-in template")
	}

	// Remove from memory
	delete(ts.templates, id)

	// Remove file from disk
	filePath := filepath.Join(ts.storagePath, id+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete template file: %w", err)
	}

	return nil
}

// saveTemplate writes a single template to disk.
func (ts *TemplateService) saveTemplate(template *models.EmailTemplate) error {
	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	filePath := filepath.Join(ts.storagePath, template.ID+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

// saveAllTemplates writes all templates to disk.
func (ts *TemplateService) saveAllTemplates() {
	for _, template := range ts.templates {
		if err := ts.saveTemplate(template); err != nil {
			fmt.Printf("Warning: failed to save template %s: %v\n", template.ID, err)
		}
	}
}
