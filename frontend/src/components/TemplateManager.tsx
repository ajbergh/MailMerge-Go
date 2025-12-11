/**
 * TemplateManager Component - Email Template Management
 * 
 * Phase 2 (v1.3): New component for managing saved email templates.
 * 
 * Features:
 * - Template selector dropdown in the email editor
 * - Load template to populate subject and body
 * - Save current email as a new template
 * - Delete user-created templates
 * - Built-in starter templates (cannot be deleted)
 * 
 * Templates are persisted to disk and survive app restarts.
 */
import { useState, useEffect, useCallback } from 'react';
import { models } from '../../wailsjs/go/models';
import { FileText, Save, Trash2, ChevronDown, Star, Plus, X } from 'lucide-react';

// Type alias for Wails-generated EmailTemplate
type EmailTemplate = models.EmailTemplate;

/**
 * Props for the TemplateManager component
 */
interface TemplateManagerProps {
  /** All available templates (from backend) */
  templates: EmailTemplate[];
  /** Callback when a template is selected to load */
  onLoadTemplate: (template: EmailTemplate) => void;
  /** Callback to save current email as template */
  onSaveTemplate: (name: string) => void;
  /** Callback to delete a template */
  onDeleteTemplate: (id: string) => void;
  /** Callback to refresh templates list */
  onRefresh: () => void;
  /** Current subject (to show in save preview) */
  currentSubject: string;
  /** Current body (to show in save preview) */
  currentBody: string;
  /** Current HTML mode */
  currentIsHTML: boolean;
}

/**
 * TemplateManager - Template selector and management UI
 * 
 * Displays a dropdown to select templates and buttons for save/delete.
 * Integrates with the email editor to load and save templates.
 */
export function TemplateManager({
  templates,
  onLoadTemplate,
  onSaveTemplate,
  onDeleteTemplate,
  onRefresh,
  currentSubject,
  currentBody,
  currentIsHTML
}: TemplateManagerProps) {
  // Dropdown open state
  const [isOpen, setIsOpen] = useState(false);
  // Save modal state
  const [showSaveModal, setShowSaveModal] = useState(false);
  // New template name input
  const [newTemplateName, setNewTemplateName] = useState('');
  // Delete confirmation state
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      if (!target.closest('.template-dropdown')) {
        setIsOpen(false);
      }
    };
    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  }, []);

  /**
   * Handle template selection
   */
  const handleSelectTemplate = useCallback((template: EmailTemplate) => {
    onLoadTemplate(template);
    setIsOpen(false);
  }, [onLoadTemplate]);

  /**
   * Handle save template
   */
  const handleSave = useCallback(() => {
    if (!newTemplateName.trim()) return;
    onSaveTemplate(newTemplateName.trim());
    setNewTemplateName('');
    setShowSaveModal(false);
    onRefresh();
  }, [newTemplateName, onSaveTemplate, onRefresh]);

  /**
   * Handle delete with confirmation
   */
  const handleDelete = useCallback((id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (deleteConfirm === id) {
      onDeleteTemplate(id);
      setDeleteConfirm(null);
      onRefresh();
    } else {
      setDeleteConfirm(id);
      // Auto-clear confirmation after 3 seconds
      setTimeout(() => setDeleteConfirm(null), 3000);
    }
  }, [deleteConfirm, onDeleteTemplate, onRefresh]);

  // Separate built-in and user templates
  const builtInTemplates = templates.filter(t => t.isBuiltIn);
  const userTemplates = templates.filter(t => !t.isBuiltIn);

  // Can save if there's content
  const canSave = currentSubject.trim() || currentBody.trim();

  return (
    <div className="template-manager">
      <div className="template-controls">
        {/* Template Dropdown */}
        <div className="template-dropdown">
          <button
            className="btn btn-outline template-trigger"
            onClick={(e) => {
              e.stopPropagation();
              setIsOpen(!isOpen);
            }}
            title="Select a template"
          >
            <FileText size={16} />
            Templates
            <ChevronDown size={14} className={`chevron ${isOpen ? 'open' : ''}`} />
          </button>

          {isOpen && (
            <div className="template-menu">
              {/* User Templates Section */}
              {userTemplates.length > 0 && (
                <>
                  <div className="template-section-header">My Templates</div>
                  {userTemplates.map(template => (
                    <div
                      key={template.id}
                      className="template-item"
                      onClick={() => handleSelectTemplate(template)}
                    >
                      <div className="template-info">
                        <span className="template-name">{template.name}</span>
                        <span className="template-format">
                          {template.isHTML ? 'HTML' : 'Plain'}
                        </span>
                      </div>
                      <button
                        className={`template-delete ${deleteConfirm === template.id ? 'confirm' : ''}`}
                        onClick={(e) => handleDelete(template.id, e)}
                        title={deleteConfirm === template.id ? 'Click again to confirm' : 'Delete template'}
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  ))}
                </>
              )}

              {/* Built-in Templates Section */}
              {builtInTemplates.length > 0 && (
                <>
                  <div className="template-section-header">
                    <Star size={12} />
                    Built-in Templates
                  </div>
                  {builtInTemplates.map(template => (
                    <div
                      key={template.id}
                      className="template-item builtin"
                      onClick={() => handleSelectTemplate(template)}
                    >
                      <div className="template-info">
                        <span className="template-name">{template.name}</span>
                        <span className="template-format">
                          {template.isHTML ? 'HTML' : 'Plain'}
                        </span>
                      </div>
                    </div>
                  ))}
                </>
              )}

              {/* Empty State */}
              {templates.length === 0 && (
                <div className="template-empty">
                  No templates yet. Save your first one!
                </div>
              )}
            </div>
          )}
        </div>

        {/* Save Button */}
        <button
          className="btn btn-outline"
          onClick={() => setShowSaveModal(true)}
          disabled={!canSave}
          title="Save as template"
        >
          <Save size={16} />
          Save Template
        </button>
      </div>

      {/* Save Modal */}
      {showSaveModal && (
        <div className="modal-overlay" onClick={() => setShowSaveModal(false)}>
          <div className="modal-content small" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h3>
                <Plus size={18} />
                Save as Template
              </h3>
              <button
                className="modal-close"
                onClick={() => setShowSaveModal(false)}
                aria-label="Close"
              >
                <X size={18} />
              </button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label htmlFor="template-name">Template Name</label>
                <input
                  id="template-name"
                  type="text"
                  className="form-input"
                  placeholder="e.g., Weekly Newsletter"
                  value={newTemplateName}
                  onChange={e => setNewTemplateName(e.target.value)}
                  onKeyDown={e => {
                    if (e.key === 'Enter') handleSave();
                    if (e.key === 'Escape') setShowSaveModal(false);
                  }}
                  autoFocus
                />
              </div>
              <div className="template-preview">
                <div className="preview-row">
                  <span className="preview-label">Subject:</span>
                  <span className="preview-value">
                    {currentSubject || <em>No subject</em>}
                  </span>
                </div>
                <div className="preview-row">
                  <span className="preview-label">Format:</span>
                  <span className="preview-value">
                    {currentIsHTML ? 'HTML (Rich Text)' : 'Plain Text'}
                  </span>
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <button
                className="btn btn-secondary"
                onClick={() => setShowSaveModal(false)}
              >
                Cancel
              </button>
              <button
                className="btn btn-primary"
                onClick={handleSave}
                disabled={!newTemplateName.trim()}
              >
                <Save size={16} />
                Save Template
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
