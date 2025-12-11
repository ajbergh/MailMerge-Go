/**
 * SettingsModal Component - Application Settings Panel
 * 
 * Phase 2 (v1.3): New component for centralized settings management.
 * 
 * Features:
 * - Theme preference (Light/Dark/System)
 * - Default email format (HTML/Plain Text)
 * - Sending delay configuration
 * - Confirmation toggle for bulk sends
 * - Recent files management
 * - Keyboard shortcuts reference
 * 
 * Settings are persisted to disk and survive app restarts.
 */
import { useState, useEffect, useCallback } from 'react';
import { models } from '../../wailsjs/go/models';
import {
  Settings,
  X,
  Sun,
  Moon,
  Monitor,
  Clock,
  Bell,
  FileText,
  Trash2,
  Keyboard,
  HelpCircle,
  Save
} from 'lucide-react';

// Type alias for Wails-generated AppSettings
type AppSettings = models.AppSettings;

/**
 * Props for the SettingsModal component
 */
interface SettingsModalProps {
  /** Whether the modal is visible */
  isOpen: boolean;
  /** Callback to close the modal */
  onClose: () => void;
  /** Current settings from backend */
  settings: AppSettings;
  /** Callback to save updated settings */
  onSave: (settings: AppSettings) => void;
  /** Recent files list */
  recentFiles: string[];
  /** Callback to clear recent files */
  onClearRecentFiles: () => void;
  /** Callback to open a recent file */
  onOpenRecentFile: (path: string) => void;
}

/**
 * SettingsModal - Settings panel with tabs
 * 
 * Displays settings in organized tabs for easy navigation.
 * Changes are applied when the user clicks Save.
 */
export function SettingsModal({
  isOpen,
  onClose,
  settings,
  onSave,
  recentFiles,
  onClearRecentFiles,
  onOpenRecentFile
}: SettingsModalProps) {
  // Local settings state for editing
  const [localSettings, setLocalSettings] = useState<AppSettings>(settings);
  // Active tab
  const [activeTab, setActiveTab] = useState<'general' | 'sending' | 'files' | 'shortcuts'>('general');
  // Has unsaved changes
  const [hasChanges, setHasChanges] = useState(false);

  // Reset local state when modal opens
  useEffect(() => {
    if (isOpen) {
      setLocalSettings(settings);
      setHasChanges(false);
    }
  }, [isOpen, settings]);

  // Handle escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  /**
   * Update a setting and mark as changed
   */
  const updateSetting = useCallback(<K extends keyof AppSettings>(key: K, value: AppSettings[K]) => {
    setLocalSettings(prev => {
      // Create a new settings object with updated value
      const updated = new models.AppSettings({
        ...prev,
        [key]: value
      });
      return updated;
    });
    setHasChanges(true);
  }, []);

  /**
   * Save changes and close
   */
  const handleSave = useCallback(() => {
    onSave(localSettings);
    setHasChanges(false);
    onClose();
  }, [localSettings, onSave, onClose]);

  /**
   * Get file name from path
   */
  const getFileName = (path: string) => {
    return path.split(/[/\\]/).pop() || path;
  };

  if (!isOpen) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content settings-modal" onClick={e => e.stopPropagation()}>
        {/* Header */}
        <div className="modal-header">
          <h3>
            <Settings size={18} />
            Settings
          </h3>
          <button className="modal-close" onClick={onClose} aria-label="Close">
            <X size={18} />
          </button>
        </div>

        {/* Tabs */}
        <div className="settings-tabs">
          <button
            className={`settings-tab ${activeTab === 'general' ? 'active' : ''}`}
            onClick={() => setActiveTab('general')}
          >
            <Sun size={14} />
            General
          </button>
          <button
            className={`settings-tab ${activeTab === 'sending' ? 'active' : ''}`}
            onClick={() => setActiveTab('sending')}
          >
            <Clock size={14} />
            Sending
          </button>
          <button
            className={`settings-tab ${activeTab === 'files' ? 'active' : ''}`}
            onClick={() => setActiveTab('files')}
          >
            <FileText size={14} />
            Recent Files
          </button>
          <button
            className={`settings-tab ${activeTab === 'shortcuts' ? 'active' : ''}`}
            onClick={() => setActiveTab('shortcuts')}
          >
            <Keyboard size={14} />
            Shortcuts
          </button>
        </div>

        {/* Tab Content */}
        <div className="modal-body settings-body">
          {/* General Tab */}
          {activeTab === 'general' && (
            <div className="settings-section">
              <div className="setting-group">
                <label className="setting-label">
                  Theme
                  <span className="setting-hint">Choose your preferred color scheme</span>
                </label>
                <div className="theme-options">
                  <button
                    className={`theme-option ${localSettings.theme === 'light' ? 'active' : ''}`}
                    onClick={() => updateSetting('theme', 'light')}
                  >
                    <Sun size={18} />
                    Light
                  </button>
                  <button
                    className={`theme-option ${localSettings.theme === 'dark' ? 'active' : ''}`}
                    onClick={() => updateSetting('theme', 'dark')}
                  >
                    <Moon size={18} />
                    Dark
                  </button>
                  <button
                    className={`theme-option ${localSettings.theme === 'system' ? 'active' : ''}`}
                    onClick={() => updateSetting('theme', 'system')}
                  >
                    <Monitor size={18} />
                    System
                  </button>
                </div>
              </div>

              <div className="setting-group">
                <label className="setting-label">
                  Default Email Format
                  <span className="setting-hint">Format for new emails</span>
                </label>
                <select
                  className="form-select"
                  value={localSettings.defaultFormat}
                  onChange={e => updateSetting('defaultFormat', e.target.value as 'html' | 'plaintext')}
                >
                  <option value="html">HTML (Rich Text)</option>
                  <option value="plaintext">Plain Text</option>
                </select>
              </div>

              <div className="setting-group">
                <label className="setting-checkbox">
                  <input
                    type="checkbox"
                    checked={localSettings.soundEnabled}
                    onChange={e => updateSetting('soundEnabled', e.target.checked)}
                  />
                  <Bell size={16} />
                  Play sound when sending completes
                </label>
              </div>
            </div>
          )}

          {/* Sending Tab */}
          {activeTab === 'sending' && (
            <div className="settings-section">
              <div className="setting-group">
                <label className="setting-label">
                  Delay Between Emails
                  <span className="setting-hint">
                    Slower speeds are more reliable with Outlook
                  </span>
                </label>
                <div className="delay-options">
                  <button
                    className={`delay-option ${localSettings.sendingDelay === 100 ? 'active' : ''}`}
                    onClick={() => updateSetting('sendingDelay', 100)}
                  >
                    Fast (100ms)
                  </button>
                  <button
                    className={`delay-option ${localSettings.sendingDelay === 500 ? 'active' : ''}`}
                    onClick={() => updateSetting('sendingDelay', 500)}
                  >
                    Normal (500ms)
                  </button>
                  <button
                    className={`delay-option ${localSettings.sendingDelay === 2000 ? 'active' : ''}`}
                    onClick={() => updateSetting('sendingDelay', 2000)}
                  >
                    Slow (2s)
                  </button>
                </div>
                <div className="custom-delay">
                  <label>Custom delay (ms):</label>
                  <input
                    type="number"
                    className="form-input"
                    min={100}
                    max={5000}
                    step={100}
                    value={localSettings.sendingDelay}
                    onChange={e => {
                      const val = parseInt(e.target.value, 10);
                      if (val >= 100 && val <= 5000) {
                        updateSetting('sendingDelay', val);
                      }
                    }}
                  />
                </div>
              </div>

              <div className="setting-group">
                <label className="setting-checkbox">
                  <input
                    type="checkbox"
                    checked={localSettings.confirmSend}
                    onChange={e => updateSetting('confirmSend', e.target.checked)}
                  />
                  <HelpCircle size={16} />
                  Show confirmation dialog before bulk sending
                </label>
              </div>

              <div className="setting-group">
                <label className="setting-checkbox">
                  <input
                    type="checkbox"
                    checked={localSettings.autoSaveTempls}
                    onChange={e => updateSetting('autoSaveTempls', e.target.checked)}
                  />
                  <Save size={16} />
                  Auto-save templates when modified
                </label>
              </div>
            </div>
          )}

          {/* Recent Files Tab */}
          {activeTab === 'files' && (
            <div className="settings-section">
              <div className="setting-group">
                <div className="setting-label">
                  Recent Contact Files
                  <span className="setting-hint">
                    Click to open, or clear all recent files
                  </span>
                </div>
                {recentFiles.length > 0 ? (
                  <>
                    <div className="recent-files-list">
                      {recentFiles.map((path, index) => (
                        <button
                          key={index}
                          className="recent-file-item"
                          onClick={() => {
                            onOpenRecentFile(path);
                            onClose();
                          }}
                          title={path}
                        >
                          <FileText size={14} />
                          <span className="recent-file-name">{getFileName(path)}</span>
                          <span className="recent-file-path">{path}</span>
                        </button>
                      ))}
                    </div>
                    <button
                      className="btn btn-outline btn-sm"
                      onClick={onClearRecentFiles}
                      style={{ marginTop: '1rem' }}
                    >
                      <Trash2 size={14} />
                      Clear Recent Files
                    </button>
                  </>
                ) : (
                  <div className="empty-state small">
                    <FileText size={24} />
                    <p>No recent files</p>
                    <span>Files you open will appear here</span>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Shortcuts Tab */}
          {activeTab === 'shortcuts' && (
            <div className="settings-section">
              <div className="setting-group">
                <div className="setting-label">
                  Keyboard Shortcuts
                  <span className="setting-hint">
                    Speed up your workflow with these shortcuts
                  </span>
                </div>
                <div className="shortcuts-list">
                  <div className="shortcut-item">
                    <span className="shortcut-keys">
                      <kbd>Ctrl</kbd> + <kbd>O</kbd>
                    </span>
                    <span className="shortcut-desc">Open contact file</span>
                  </div>
                  <div className="shortcut-item">
                    <span className="shortcut-keys">
                      <kbd>Ctrl</kbd> + <kbd>S</kbd>
                    </span>
                    <span className="shortcut-desc">Save as template</span>
                  </div>
                  <div className="shortcut-item">
                    <span className="shortcut-keys">
                      <kbd>Ctrl</kbd> + <kbd>Enter</kbd>
                    </span>
                    <span className="shortcut-desc">Send test email</span>
                  </div>
                  <div className="shortcut-item">
                    <span className="shortcut-keys">
                      <kbd>Ctrl</kbd> + <kbd>Shift</kbd> + <kbd>Enter</kbd>
                    </span>
                    <span className="shortcut-desc">Send to all</span>
                  </div>
                  <div className="shortcut-item">
                    <span className="shortcut-keys">
                      <kbd>Ctrl</kbd> + <kbd>,</kbd>
                    </span>
                    <span className="shortcut-desc">Open settings</span>
                  </div>
                  <div className="shortcut-item">
                    <span className="shortcut-keys">
                      <kbd>Ctrl</kbd> + <kbd>D</kbd>
                    </span>
                    <span className="shortcut-desc">Toggle dark mode</span>
                  </div>
                  <div className="shortcut-item">
                    <span className="shortcut-keys">
                      <kbd>Escape</kbd>
                    </span>
                    <span className="shortcut-desc">Close modal</span>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="modal-footer">
          <button className="btn btn-secondary" onClick={onClose}>
            Cancel
          </button>
          <button
            className="btn btn-primary"
            onClick={handleSave}
            disabled={!hasChanges}
          >
            <Save size={16} />
            Save Changes
          </button>
        </div>
      </div>
    </div>
  );
}
