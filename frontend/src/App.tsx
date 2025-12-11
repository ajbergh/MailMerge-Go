/**
 * App Component - Main Application Container
 * 
 * This is the root component of the MailMerge application. It orchestrates
 * the entire email merge workflow:
 * 
 * 1. Import Contacts - Load contacts from CSV/Excel files
 * 2. Compose Email - Write subject and body with merge fields
 * 3. Manage Attachments - Add files to attach to all emails
 * 4. Send Emails - Test single emails or send to all contacts
 * 5. View Results - See success/failure summary and export logs
 * 
 * State Management:
 * - Uses React hooks for local state management
 * - Subscribes to backend events for real-time progress updates
 * - Communicates with Go backend via Wails-generated bindings
 * 
 * Components:
 * - FileUpload: Contact file selection and parsing
 * - ContactTable: Preview of loaded contacts
 * - EmailEditor: Subject/body composition with merge fields
 * - AttachmentManager: File attachment handling
 * - ProgressTracker: Real-time send progress display
 * - ResultsSummary: Final results and log export
 * 
 * Phase 2 Additions (v1.3):
 * - TemplateManager: Save/load email templates
 * - SettingsModal: Application settings panel
 * - Keyboard shortcuts for quick actions
 * - Theme management synced with settings
 */
import { useState, useEffect, useCallback, useMemo } from 'react';
import { Mail, AlertCircle, Send, FlaskConical, CheckCircle2, Sun, Moon, Eye, RefreshCw, Settings } from 'lucide-react';
import { 
  FileUpload, 
  ContactTable, 
  EmailEditor, 
  AttachmentManager, 
  ProgressTracker, 
  ResultsSummary,
  PreviewModal,
  TemplateManager,
  SettingsModal
} from './components';
import { ProgressUpdate, ParseResult, EmailTemplate, AppSettings } from './types';
import './styles/app.css';

// Import Wails bindings and models
import { 
  SelectContactFile, 
  ParseContactFile, 
  SelectAttachments,
  CheckOutlookInstalled,
  SendTestEmail,
  SendBulkEmails,
  GetMergeFields,
  GetMergeFieldsFromHeaders,
  PreviewMergeForContact,
  ExportLogsToCSV,
  // Phase 2: Template bindings
  GetAllTemplates,
  SaveTemplate,
  DeleteTemplate,
  // Phase 2: Settings bindings
  GetSettings,
  UpdateSettings,
  GetRecentFiles,
  AddRecentFile,
  ClearRecentFiles
} from '../wailsjs/go/main/App';
import { models } from '../wailsjs/go/models';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';

// Type aliases for cleaner code
type Contact = models.Contact;
type SendResult = models.SendResult;
type EmailLog = models.EmailLog;

// Use Wails-generated types for Phase 2 features (avoid type mismatches)
type WailsEmailTemplate = models.EmailTemplate;
type WailsAppSettings = models.AppSettings;

/**
 * App - Main application component
 * 
 * Manages all application state and renders the mail merge interface.
 * Handles file selection, email composition, sending, and results display.
 */
function App() {
  // ==================== State Management ====================
  
  // File and contacts state
  const [fileName, setFileName] = useState<string | null>(null);
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [parseErrors, setParseErrors] = useState<string[]>([]);
  const [parseWarnings, setParseWarnings] = useState<string[]>([]);
  const [isLoadingFile, setIsLoadingFile] = useState(false);
  
  // Phase 1: Selection and filtering state
  const [selectedIndices, setSelectedIndices] = useState<Set<number>>(new Set());
  const [headers, setHeaders] = useState<string[]>([]);
  const [duplicates, setDuplicates] = useState<string[]>([]);
  
  // Phase 1: Preview modal state
  const [showPreviewModal, setShowPreviewModal] = useState(false);

  // Email composition state
  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');
  const [isHTML, setIsHTML] = useState(true);
  const [mergeFields, setMergeFields] = useState<string[]>(['{FirstName}', '{LastName}', '{Email}']);
  
  // Phase 3: CC/BCC state
  const [cc, setCc] = useState('');
  const [bcc, setBcc] = useState('');

  // Attachments state
  const [attachments, setAttachments] = useState<string[]>([]);

  // Sending state
  const [isSending, setIsSending] = useState(false);
  const [progress, setProgress] = useState<ProgressUpdate | null>(null);
  const [progressLogs, setProgressLogs] = useState<Array<{ email: string; status: string; message?: string; timestamp: string }>>([]);
  const [sendResult, setSendResult] = useState<SendResult | null>(null);

  // Error/status state
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [outlookStatus, setOutlookStatus] = useState<'checking' | 'ok' | 'error'>('checking');

  // Theme state - initialize from localStorage or default to light
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    const saved = localStorage.getItem('theme');
    return (saved === 'dark') ? 'dark' : 'light';
  });

  // Phase 2: Templates state
  const [templates, setTemplates] = useState<WailsEmailTemplate[]>([]);
  const [isLoadingTemplates, setIsLoadingTemplates] = useState(false);

  // Phase 2: Settings state
  const [settings, setSettings] = useState<WailsAppSettings>(new models.AppSettings({
    theme: 'system',
    defaultFormat: 'html',
    sendingDelay: 500,
    confirmSend: true,
    soundEnabled: true,
    autoSaveTempls: false,
    recentFiles: []
  }));
  const [recentFiles, setRecentFiles] = useState<string[]>([]);
  const [showSettingsModal, setShowSettingsModal] = useState(false);

  // ==================== Initialization Effects ====================

  /**
   * Apply theme to document on mount and when theme changes
   * Persists theme preference to localStorage
   * Phase 2: Now syncs with settings service
   */
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
  }, [theme]);

  /**
   * Toggle between light and dark themes
   * Phase 2: Also updates settings
   */
  const toggleTheme = useCallback(() => {
    setTheme(prev => {
      const newTheme = prev === 'light' ? 'dark' : 'light';
      // Update settings with new theme
      setSettings(s => ({ ...s, theme: newTheme }));
      UpdateSettings({ ...settings, theme: newTheme }).catch(console.error);
      return newTheme;
    });
  }, [settings]);

  /**
   * Phase 2: Load templates from backend
   */
  const loadTemplates = useCallback(async () => {
    try {
      setIsLoadingTemplates(true);
      const temps = await GetAllTemplates();
      setTemplates(temps || []);
    } catch (err) {
      console.error('Failed to load templates:', err);
    } finally {
      setIsLoadingTemplates(false);
    }
  }, []);

  /**
   * Phase 2: Load settings from backend
   */
  const loadSettings = useCallback(async () => {
    try {
      const s = await GetSettings();
      if (s) {
        setSettings(s);
        // Apply theme from settings
        if (s.theme === 'dark' || (s.theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
          setTheme('dark');
        } else if (s.theme === 'light') {
          setTheme('light');
        }
        // Set default format
        if (s.defaultFormat === 'plaintext') {
          setIsHTML(false);
        }
      }
    } catch (err) {
      console.error('Failed to load settings:', err);
    }
  }, []);

  /**
   * Phase 2: Load recent files
   */
  const loadRecentFiles = useCallback(async () => {
    try {
      const files = await GetRecentFiles();
      setRecentFiles(files || []);
    } catch (err) {
      console.error('Failed to load recent files:', err);
    }
  }, []);

  /**
   * Check Outlook availability on mount
   * Also loads the list of available merge fields
   * Phase 2: Also loads templates and settings
   */
  useEffect(() => {
    const checkOutlook = async () => {
      try {
        await CheckOutlookInstalled();
        setOutlookStatus('ok');
      } catch (err: any) {
        setOutlookStatus('error');
        setError(`Outlook is not available: ${err.message || err}`);
      }
    };
    checkOutlook();

    // Get merge fields from backend
    GetMergeFields().then(setMergeFields).catch(console.error);

    // Phase 2: Load templates and settings
    loadTemplates();
    loadSettings();
    loadRecentFiles();
  }, [loadTemplates, loadSettings, loadRecentFiles]);

  /**
   * Subscribe to real-time progress events from backend
   * Handles "email:progress" events during bulk sending
   */
  useEffect(() => {
    const handleProgress = (update: ProgressUpdate) => {
      setProgress(update);
      // Add to log if not a completion event
      if (update.status !== 'complete' && update.email) {
        setProgressLogs(prev => [...prev, {
          email: update.email,
          status: update.status,
          message: update.message,
          timestamp: update.timestamp
        }]);
      }
    };

    EventsOn('email:progress', handleProgress);
    return () => {
      EventsOff('email:progress');
    };
  }, []);

  // ==================== Event Handlers ====================

  /**
   * Handle file selection and parsing
   * Opens native file dialog and parses selected CSV/Excel file
   * Phase 1: Now extracts headers, duplicates, and generates dynamic merge fields
   * Phase 2: Now adds to recent files list
   */
  const handleSelectFile = useCallback(async () => {
    try {
      setIsLoadingFile(true);
      setError(null);
      
      const filePath = await SelectContactFile();
      if (!filePath) {
        setIsLoadingFile(false);
        return;
      }

      const result = await ParseContactFile(filePath);
      setFileName(filePath.split(/[/\\]/).pop() || filePath);
      setContacts(result.contacts || []);
      setParseErrors(result.errors || []);
      setParseWarnings(result.warnings || []);
      setHeaders(result.headers || []);
      setDuplicates(result.duplicates || []);
      
      // Phase 2: Add to recent files
      try {
        await AddRecentFile(filePath);
        loadRecentFiles();
      } catch (err) {
        console.error('Failed to add to recent files:', err);
      }
      
      // Phase 1: Select all contacts by default
      if (result.contacts && result.contacts.length > 0) {
        setSelectedIndices(new Set(result.contacts.map((_, i) => i)));
      } else {
        setSelectedIndices(new Set());
      }
      
      // Phase 1: Generate merge fields from headers
      if (result.headers && result.headers.length > 0) {
        try {
          const fields = await GetMergeFieldsFromHeaders(result.headers);
          setMergeFields(fields);
        } catch (err) {
          console.error('Failed to generate merge fields:', err);
          // Fall back to default merge fields
          GetMergeFields().then(setMergeFields).catch(console.error);
        }
      }
      
      if (result.contacts.length === 0) {
        setError('No valid contacts found in the file');
      } else {
        let successMsg = `Successfully loaded ${result.contacts.length} contacts`;
        if (result.duplicates && result.duplicates.length > 0) {
          successMsg += ` (${result.duplicates.length} duplicate emails detected)`;
        }
        setSuccessMessage(successMsg);
        setTimeout(() => setSuccessMessage(null), 4000);
      }
    } catch (err: any) {
      setError(`Failed to load file: ${err.message || err}`);
      setContacts([]);
      setFileName(null);
      setHeaders([]);
      setDuplicates([]);
      setSelectedIndices(new Set());
    } finally {
      setIsLoadingFile(false);
    }
  }, [loadRecentFiles]);

  /**
   * Handle adding attachments via native file dialog
   */
  const handleAddAttachments = useCallback(async () => {
    try {
      const files = await SelectAttachments();
      if (files && files.length > 0) {
        setAttachments(prev => [...prev, ...files]);
      }
    } catch (err: any) {
      setError(`Failed to add attachments: ${err.message || err}`);
    }
  }, []);

  /**
   * Remove an attachment by index
   */
  const handleRemoveAttachment = useCallback((index: number) => {
    setAttachments(prev => prev.filter((_, i) => i !== index));
  }, []);

  /**
   * Send a test email to a specified address
   * Uses first contact's data for merge preview, or defaults
   * Phase 3: Now includes CC/BCC in test emails
   */
  const handleSendTestEmail = useCallback(async () => {
    const testAddress = prompt('Enter test email address:');
    if (!testAddress) return;

    try {
      setIsSending(true);
      setError(null);
      
      // Phase 3: Include CC/BCC in test email request
      // Determine if CC/BCC have merge fields - if so, use as templates
      const ccHasMergeFields = cc.includes('{');
      const bccHasMergeFields = bcc.includes('{');
      
      const request = new models.TestEmailRequest({
        testAddress,
        subjectTemplate: subject,
        bodyTemplate: body,
        isHTML,
        attachments,
        cc: ccHasMergeFields ? '' : cc,  // Static CC
        bcc: bccHasMergeFields ? '' : bcc,  // Static BCC
        ccTemplate: ccHasMergeFields ? cc : '',  // CC with merge fields
        bccTemplate: bccHasMergeFields ? bcc : '',  // BCC with merge fields
        sampleFirstName: contacts[0]?.firstName || 'John',
        sampleLastName: contacts[0]?.lastName || 'Doe'
      });
      
      await SendTestEmail(request);
      
      setSuccessMessage(`Test email sent successfully to ${testAddress}`);
      setTimeout(() => setSuccessMessage(null), 5000);
    } catch (err: any) {
      setError(`Failed to send test email: ${err.message || err}`);
    } finally {
      setIsSending(false);
    }
  }, [subject, body, isHTML, attachments, contacts, cc, bcc]);

  /**
   * Send emails to all loaded contacts
   * Shows confirmation dialog and tracks progress
   * Phase 1: Now sends only to selected contacts
   */
  const handleSendAll = useCallback(async () => {
    // Phase 1: Get only selected contacts
    const selectedContacts = contacts.filter((_, i) => selectedIndices.has(i));
    
    if (selectedContacts.length === 0) {
      setError('No contacts selected');
      return;
    }

    if (!subject.trim()) {
      setError('Please enter a subject');
      return;
    }

    if (!body.trim()) {
      setError('Please enter a message body');
      return;
    }

    const confirmed = window.confirm(
      `Are you sure you want to send ${selectedContacts.length} emails?\n\nThis action cannot be undone.`
    );
    if (!confirmed) return;

    try {
      setIsSending(true);
      setError(null);
      setProgress(null);
      setProgressLogs([]);
      setSendResult(null);

      // Phase 3: Include CC/BCC in bulk email request
      // Determine if CC/BCC have merge fields - if so, use as templates
      const ccHasMergeFields = cc.includes('{');
      const bccHasMergeFields = bcc.includes('{');

      const request = new models.EmailRequest({
        contacts: selectedContacts,
        subjectTemplate: subject,
        bodyTemplate: body,
        isHTML,
        attachments,
        cc: ccHasMergeFields ? '' : cc,  // Static CC
        bcc: bccHasMergeFields ? '' : bcc,  // Static BCC
        ccTemplate: ccHasMergeFields ? cc : '',  // CC with merge fields
        bccTemplate: bccHasMergeFields ? bcc : ''  // BCC with merge fields
      });

      const result = await SendBulkEmails(request);

      setSendResult(result);
    } catch (err: any) {
      setError(`Failed to send emails: ${err.message || err}`);
    } finally {
      setIsSending(false);
    }
  }, [contacts, selectedIndices, subject, body, isHTML, attachments, cc, bcc]);

  /**
   * Phase 1: Retry sending to failed contacts
   * Re-queues failed contacts for another send attempt
   */
  const handleRetryFailed = useCallback(async () => {
    if (!sendResult?.failedContacts || sendResult.failedContacts.length === 0) {
      setError('No failed contacts to retry');
      return;
    }

    const confirmed = window.confirm(
      `Retry sending to ${sendResult.failedContacts.length} failed contacts?`
    );
    if (!confirmed) return;

    try {
      setIsSending(true);
      setError(null);
      setProgress(null);
      setProgressLogs([]);

      // Phase 3: Include CC/BCC in retry request (same logic as main send)
      const ccHasMergeFields = cc.includes('{');
      const bccHasMergeFields = bcc.includes('{');

      const request = new models.EmailRequest({
        contacts: sendResult.failedContacts,
        subjectTemplate: subject,
        bodyTemplate: body,
        isHTML,
        attachments,
        cc: ccHasMergeFields ? '' : cc,
        bcc: bccHasMergeFields ? '' : bcc,
        ccTemplate: ccHasMergeFields ? cc : '',
        bccTemplate: bccHasMergeFields ? bcc : ''
      });

      const result = await SendBulkEmails(request);

      // Merge results: add successful retries to previous totals
      setSendResult(prev => {
        if (!prev) return result;
        // Create a new SendResult with merged data
        return new models.SendResult({
          totalSent: prev.totalSent + result.totalSent,
          totalFailed: result.totalFailed,
          failedContacts: result.failedContacts,
          logs: [...(prev.logs || []), ...(result.logs || [])]
        });
      });
    } catch (err: any) {
      setError(`Failed to retry emails: ${err.message || err}`);
    } finally {
      setIsSending(false);
    }
  }, [sendResult, subject, body, isHTML, attachments, cc, bcc]);

  /**
   * Phase 1: Handle preview email callback
   * Returns rendered subject and body for a specific contact
   */
  const handlePreviewEmail = useCallback(async (contact: Contact): Promise<{ subject: string; body: string } | null> => {
    try {
      const result = await PreviewMergeForContact(subject, body, contact);
      if (result) {
        return {
          subject: result.subject || '',
          body: result.body || ''
        };
      }
      return null;
    } catch (err) {
      console.error('Failed to preview email:', err);
      return null;
    }
  }, [subject, body]);

  // Export logs handler
  const handleExportLogs = useCallback(async (logs: models.EmailLog[]) => {
    try {
      const path = await ExportLogsToCSV(logs);
      if (path) {
        setSuccessMessage(`Logs exported to ${path}`);
        setTimeout(() => setSuccessMessage(null), 5000);
      }
    } catch (err: any) {
      setError(`Failed to export logs: ${err.message || err}`);
    }
  }, []);

  // Reset handler - Phase 1: Now clears all new state
  const handleReset = useCallback(() => {
    setSendResult(null);
    setProgress(null);
    setProgressLogs([]);
    setContacts([]);
    setFileName(null);
    setSubject('');
    setBody('');
    setAttachments([]);
    setParseErrors([]);
    setParseWarnings([]);
    setSelectedIndices(new Set());
    setHeaders([]);
    setDuplicates([]);
  }, []);

  /**
   * Phase 2: Load a template into the editor
   */
  const handleLoadTemplate = useCallback((template: WailsEmailTemplate) => {
    setSubject(template.subject);
    setBody(template.body);
    setIsHTML(template.isHTML);
    setSuccessMessage(`Loaded template: ${template.name}`);
    setTimeout(() => setSuccessMessage(null), 3000);
  }, []);

  /**
   * Phase 2: Save current email as a template
   */
  const handleSaveTemplate = useCallback(async (name: string) => {
    try {
      const template = new models.EmailTemplate({
        name,
        subject,
        body,
        isHTML
      });
      await SaveTemplate(template);
      await loadTemplates();
      setSuccessMessage(`Template "${name}" saved successfully`);
      setTimeout(() => setSuccessMessage(null), 3000);
    } catch (err: any) {
      setError(`Failed to save template: ${err.message || err}`);
    }
  }, [subject, body, isHTML, loadTemplates]);

  /**
   * Phase 2: Delete a template
   */
  const handleDeleteTemplate = useCallback(async (id: string) => {
    try {
      await DeleteTemplate(id);
      await loadTemplates();
      setSuccessMessage('Template deleted');
      setTimeout(() => setSuccessMessage(null), 3000);
    } catch (err: any) {
      setError(`Failed to delete template: ${err.message || err}`);
    }
  }, [loadTemplates]);

  /**
   * Phase 2: Save settings
   */
  const handleSaveSettings = useCallback(async (newSettings: WailsAppSettings) => {
    try {
      await UpdateSettings(newSettings);
      setSettings(newSettings);
      // Apply theme change
      if (newSettings.theme === 'dark' || 
          (newSettings.theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
        setTheme('dark');
      } else if (newSettings.theme === 'light') {
        setTheme('light');
      }
      setSuccessMessage('Settings saved');
      setTimeout(() => setSuccessMessage(null), 3000);
    } catch (err: any) {
      setError(`Failed to save settings: ${err.message || err}`);
    }
  }, []);

  /**
   * Phase 2: Clear recent files
   */
  const handleClearRecentFiles = useCallback(async () => {
    try {
      await ClearRecentFiles();
      setRecentFiles([]);
    } catch (err: any) {
      setError(`Failed to clear recent files: ${err.message || err}`);
    }
  }, []);

  /**
   * Phase 2: Open a recent file
   */
  const handleOpenRecentFile = useCallback(async (path: string) => {
    try {
      setIsLoadingFile(true);
      setError(null);
      
      const result = await ParseContactFile(path);
      setFileName(path.split(/[/\\]/).pop() || path);
      setContacts(result.contacts || []);
      setParseErrors(result.errors || []);
      setParseWarnings(result.warnings || []);
      setHeaders(result.headers || []);
      setDuplicates(result.duplicates || []);
      
      if (result.contacts && result.contacts.length > 0) {
        setSelectedIndices(new Set(result.contacts.map((_, i) => i)));
      } else {
        setSelectedIndices(new Set());
      }
      
      if (result.headers && result.headers.length > 0) {
        try {
          const fields = await GetMergeFieldsFromHeaders(result.headers);
          setMergeFields(fields);
        } catch (err) {
          console.error('Failed to generate merge fields:', err);
        }
      }
      
      if (result.contacts.length > 0) {
        let successMsg = `Successfully loaded ${result.contacts.length} contacts`;
        setSuccessMessage(successMsg);
        setTimeout(() => setSuccessMessage(null), 4000);
      }
    } catch (err: any) {
      setError(`Failed to load file: ${err.message || err}`);
    } finally {
      setIsLoadingFile(false);
    }
  }, []);

  // Phase 1: Selected contacts count and check
  const selectedCount = selectedIndices.size;
  const canSend = selectedCount > 0 && subject.trim() && body.trim() && !isSending && outlookStatus === 'ok';

  /**
   * Phase 2: Keyboard shortcuts handler
   * Ctrl+O: Open file, Ctrl+S: Save template, Ctrl+,: Settings, Ctrl+D: Dark mode
   */
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if in input/textarea
      const target = e.target as HTMLElement;
      const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA';
      
      if (e.ctrlKey || e.metaKey) {
        switch (e.key.toLowerCase()) {
          case 'o':
            // Ctrl+O: Open file
            if (!isInput) {
              e.preventDefault();
              handleSelectFile();
            }
            break;
          case 's':
            // Ctrl+S: Save template (show save prompt)
            e.preventDefault();
            // Trigger custom event that TemplateManager listens to
            window.dispatchEvent(new CustomEvent('save-template-shortcut'));
            break;
          case ',':
            // Ctrl+,: Open settings
            e.preventDefault();
            setShowSettingsModal(true);
            break;
          case 'd':
            // Ctrl+D: Toggle dark mode
            if (!isInput) {
              e.preventDefault();
              toggleTheme();
            }
            break;
          case 'enter':
            // Ctrl+Enter: Send test (Ctrl+Shift+Enter: Send all)
            if (!isInput) {
              e.preventDefault();
              if (e.shiftKey && canSend) {
                handleSendAll();
              } else if (!e.shiftKey && subject.trim() && body.trim() && outlookStatus === 'ok') {
                handleSendTestEmail();
              }
            }
            break;
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleSelectFile, toggleTheme, handleSendTestEmail, handleSendAll, canSend, subject, body, outlookStatus]);

  return (
    <div className="app-container">
      {/* Header */}
      <header className="app-header">
        <div className="app-header-content">
          <div className="app-header-left">
            <h1>
              <Mail size={24} />
              MailMerge Go
            </h1>
            <p className="subtitle">
              Outlook-based email mail merge for Windows
              {outlookStatus === 'ok' && (
                <span style={{ marginLeft: '1rem', color: '#86efac' }}>
                  <CheckCircle2 size={14} style={{ verticalAlign: 'middle', marginRight: '0.25rem' }} />
                  Outlook Connected
                </span>
              )}
            </p>
          </div>
          <div className="header-buttons">
            <button
              className="theme-toggle"
              onClick={() => setShowSettingsModal(true)}
              title="Settings (Ctrl+,)"
              aria-label="Open settings"
            >
              <Settings size={18} />
            </button>
            <button
              className="theme-toggle"
              onClick={toggleTheme}
              title={theme === 'light' ? 'Switch to dark mode (Ctrl+D)' : 'Switch to light mode (Ctrl+D)'}
              aria-label="Toggle theme"
            >
              {theme === 'light' ? <Moon size={18} /> : <Sun size={18} />}
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="app-main">
        {/* Error Alert */}
        {error && (
          <div className="alert alert-error">
            <AlertCircle size={20} />
            <div>
              <strong>Error:</strong> {error}
              <button 
                onClick={() => setError(null)}
                style={{ 
                  marginLeft: '1rem', 
                  background: 'none', 
                  border: 'none', 
                  textDecoration: 'underline', 
                  cursor: 'pointer',
                  color: 'inherit'
                }}
              >
                Dismiss
              </button>
            </div>
          </div>
        )}

        {/* Success Alert */}
        {successMessage && (
          <div className="alert alert-success">
            <CheckCircle2 size={20} />
            {successMessage}
          </div>
        )}

        {/* Outlook Error State */}
        {outlookStatus === 'error' && (
          <div className="alert alert-warning">
            <AlertCircle size={20} />
            <div>
              <strong>Outlook Not Available:</strong> Make sure Microsoft Outlook is installed and configured on this computer.
            </div>
          </div>
        )}

        {/* Show results if complete */}
        {sendResult ? (
          <ResultsSummary 
            result={sendResult} 
            onExportLogs={handleExportLogs}
            onReset={handleReset}
            onRetryFailed={sendResult.failedContacts && sendResult.failedContacts.length > 0 ? handleRetryFailed : undefined}
          />
        ) : (
          <>
            {/* Two Column Layout */}
            <div className="two-column">
              {/* Left Column */}
              <div>
                <FileUpload
                  onFileSelect={handleSelectFile}
                  fileName={fileName}
                  contacts={contacts}
                  errors={parseErrors}
                  isLoading={isLoadingFile}
                />
                <ContactTable 
                  contacts={contacts}
                  selectedIndices={selectedIndices}
                  onSelectionChange={setSelectedIndices}
                  duplicates={duplicates}
                />
              </div>

              {/* Right Column */}
              <div>
                {/* Phase 2: Template Manager */}
                <div className="card" style={{ marginBottom: '1rem' }}>
                  <div className="card-header">
                    <h2>Email Templates</h2>
                  </div>
                  <div className="card-body">
                    <TemplateManager
                      templates={templates}
                      onLoadTemplate={handleLoadTemplate}
                      onSaveTemplate={handleSaveTemplate}
                      onDeleteTemplate={handleDeleteTemplate}
                      onRefresh={loadTemplates}
                      currentSubject={subject}
                      currentBody={body}
                      currentIsHTML={isHTML}
                    />
                  </div>
                </div>
                <EmailEditor
                  subject={subject}
                  body={body}
                  isHTML={isHTML}
                  cc={cc}
                  bcc={bcc}
                  onSubjectChange={setSubject}
                  onBodyChange={setBody}
                  onIsHTMLChange={setIsHTML}
                  onCCChange={setCc}
                  onBCCChange={setBcc}
                  mergeFields={mergeFields}
                />
                <AttachmentManager
                  attachments={attachments}
                  onAddAttachments={handleAddAttachments}
                  onRemoveAttachment={handleRemoveAttachment}
                />
              </div>
            </div>

            {/* Progress Tracker */}
            <ProgressTracker
              progress={progress}
              logs={progressLogs}
              isRunning={isSending}
            />

            {/* Action Buttons */}
            <div className="card">
              <div className="card-header">
                <h2>
                  <span className="step-badge">4</span>
                  Send Emails
                </h2>
              </div>
              <div className="card-body">
                <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
                  {/* Phase 1: Preview Button */}
                  <button
                    className="btn btn-secondary btn-lg"
                    onClick={() => setShowPreviewModal(true)}
                    disabled={!subject.trim() || !body.trim() || contacts.length === 0}
                    title="Preview how emails will look with merge fields filled in"
                  >
                    <Eye size={18} />
                    Preview Email
                  </button>
                  <button
                    className="btn btn-secondary btn-lg"
                    onClick={handleSendTestEmail}
                    disabled={!subject.trim() || !body.trim() || isSending || outlookStatus !== 'ok'}
                  >
                    <FlaskConical size={18} />
                    Send Test Email
                  </button>
                  <button
                    className="btn btn-success btn-lg"
                    onClick={handleSendAll}
                    disabled={!canSend}
                  >
                    {isSending ? (
                      <>
                        <div className="spinner" style={{ borderTopColor: 'white' }}></div>
                        Sending...
                      </>
                    ) : (
                      <>
                        <Send size={18} />
                        Send to Selected {selectedCount > 0 && `(${selectedCount})`}
                      </>
                    )}
                  </button>
                </div>
                
                {!canSend && !isSending && (
                  <p style={{ marginTop: '1rem', color: 'var(--gray-500)', fontSize: '0.875rem' }}>
                    {outlookStatus !== 'ok' 
                      ? 'Outlook connection required to send emails'
                      : selectedCount === 0 
                        ? 'Select contacts to enable sending' 
                        : !subject.trim() 
                          ? 'Add a subject to enable sending'
                          : 'Add a message body to enable sending'
                    }
                  </p>
                )}
              </div>
            </div>
          </>
        )}

        {/* Phase 1: Preview Modal */}
        <PreviewModal
          isOpen={showPreviewModal}
          onClose={() => setShowPreviewModal(false)}
          contacts={contacts}
          subject={subject}
          body={body}
          isHTML={isHTML}
          attachments={attachments}
          onPreview={handlePreviewEmail}
        />

        {/* Phase 2: Settings Modal */}
        <SettingsModal
          isOpen={showSettingsModal}
          onClose={() => setShowSettingsModal(false)}
          settings={settings}
          onSave={handleSaveSettings}
          recentFiles={recentFiles}
          onClearRecentFiles={handleClearRecentFiles}
          onOpenRecentFile={handleOpenRecentFile}
        />
      </main>
    </div>
  );
}

export default App;
