/**
 * App Component - Main Application Container
 * 
 * This is the root component of the MailMerge application. It orchestrates
 * the entire email merge workflow:
 * 
 * 1. Import Contacts - Load contacts from CSV/Excel files
 * 2. Compose Email - Write subject and body with merge fields
 * 3. Manage Attachments - Add files to attach to all emails
 * 4. Send Emails - Test a representative contact or send to selected contacts
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
 * It also coordinates persisted templates and settings, preflight feedback,
 * cancellation/retry, keyboard shortcuts, and theme state.
 */
import { useState, useEffect, useCallback, useMemo } from 'react';
import { Mail, AlertCircle, Send, FlaskConical, CheckCircle2, Sun, Moon, Eye, Settings, History, Save } from 'lucide-react';
import { 
  FileUpload, 
  ContactTable, 
  EmailEditor, 
  AttachmentManager, 
  ProgressTracker, 
  ResultsSummary,
  PreviewModal,
  TemplateManager,
  SettingsModal,
  TestSendModal,
  PreflightModal,
  CampaignHistoryModal
} from './components';
import { ProgressUpdate } from './types';
import type { TestSendOptions } from './components';
import './styles/app.css';

// Import Wails bindings and models
import {
  SelectContactFile,
  ParseContactFile,
  SelectAttachments,
  GetFileInfo,
  GetOutlookStatus,
  SendTestEmail,
  SendBulkEmails,
  PreflightCampaign,
  RetryFailed,
  RetryCampaign,
  CancelCampaign,
  GetCampaignHistory,
  DeleteCampaign,
  ClearCampaignHistory,
  GetMergeFields,
  GetMergeFieldsFromHeaders,
  PreviewMergeForContact,
  ExportLogsToCSV,
  // Template bindings
  GetAllTemplates,
  SaveTemplate,
  DeleteTemplate,
  // Settings bindings
  GetSettings,
  UpdateSettings,
  GetRecentFiles,
  AddRecentFile,
  ClearRecentFiles
} from '../wailsjs/go/main/App';
import { models, campaign, storage } from '../wailsjs/go/models';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';

// Type aliases for cleaner code
type Contact = models.Contact;
type CampaignResult = campaign.CampaignResult;
type PreflightResult = campaign.PreflightResult;
type FileInfo = models.FileInfo;
type CampaignRecord = storage.CampaignRecord;

// Use Wails-generated types at the API boundary to avoid type drift.
type WailsEmailTemplate = models.EmailTemplate;
type WailsAppSettings = models.AppSettings;

/**
 * App - Main application component
 * 
 * Manages all application state and renders the mail merge interface.
 * Handles file selection, email composition, sending, and results display.
 */
function normalizedWindowsPath(path: string): string {
  return path.trim().replaceAll('/', '\\').toLocaleLowerCase();
}

function appendUniquePaths(existing: string[], additions: string[]): string[] {
  const seen = new Set(existing.map(normalizedWindowsPath));
  const result = [...existing];
  for (const path of additions) {
    const key = normalizedWindowsPath(path);
    if (!key || seen.has(key)) continue;
    seen.add(key);
    result.push(path);
  }
  return result;
}

function App() {
  // ==================== State Management ====================
  
  // File and contacts state
  const [fileName, setFileName] = useState<string | null>(null);
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [parseErrors, setParseErrors] = useState<string[]>([]);
  const [parseWarnings, setParseWarnings] = useState<string[]>([]);
  const [isLoadingFile, setIsLoadingFile] = useState(false);
  
  // Selection is keyed by stable contact ID (survives filtering/sorting).
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [headers, setHeaders] = useState<string[]>([]);
  const [duplicates, setDuplicates] = useState<string[]>([]);
  
  // Preview modal state
  const [showPreviewModal, setShowPreviewModal] = useState(false);

  // Email composition state
  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');
  const [isHTML, setIsHTML] = useState(true);
  const [mergeFields, setMergeFields] = useState<string[]>(['{FirstName}', '{LastName}', '{Email}']);
  
  // CC/BCC state
  const [cc, setCc] = useState('');
  const [bcc, setBcc] = useState('');

  // Attachments state
  const [attachments, setAttachments] = useState<string[]>([]);
  const [attachmentInfo, setAttachmentInfo] = useState<FileInfo[]>([]);

  // Sending state
  const [isSending, setIsSending] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);
  const [progress, setProgress] = useState<ProgressUpdate | null>(null);
  const [progressLogs, setProgressLogs] = useState<Array<{ email: string; status: string; message?: string; timestamp: string }>>([]);
  const [sendResult, setSendResult] = useState<CampaignResult | null>(null);
  const [preflight, setPreflight] = useState<PreflightResult | null>(null);
  const [lastRequest, setLastRequest] = useState<models.EmailRequest | null>(null);
  const [draftOnly, setDraftOnly] = useState(false);
  const [showTestSendModal, setShowTestSendModal] = useState(false);
  const [showPreflightModal, setShowPreflightModal] = useState(false);
  const [pendingRequest, setPendingRequest] = useState<models.EmailRequest | null>(null);

  // Campaign history state
  const [showHistoryModal, setShowHistoryModal] = useState(false);
  const [campaignHistory, setCampaignHistory] = useState<CampaignRecord[]>([]);
  const [isLoadingHistory, setIsLoadingHistory] = useState(false);

  // Error/status state
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [outlookStatus, setOutlookStatus] = useState<'checking' | 'ok' | 'error'>('checking');

  // Theme state - initialize from localStorage or default to light
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    const saved = localStorage.getItem('theme');
    return (saved === 'dark') ? 'dark' : 'light';
  });

  // Templates state
  const [templates, setTemplates] = useState<WailsEmailTemplate[]>([]);
  const [isLoadingTemplates, setIsLoadingTemplates] = useState(false);

  // Settings state
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
   * Applies the selected theme to the document and keeps the local UI preference
   * available before persisted settings finish loading.
   */
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
  }, [theme]);

  /**
   * Toggles the local theme and persists the resulting setting.
   */
  const toggleTheme = useCallback(() => {
    setTheme(prev => {
      const newTheme = prev === 'light' ? 'dark' : 'light';
      setSettings(current => {
        const next = new models.AppSettings({ ...current, theme: newTheme });
        UpdateSettings(next).catch(console.error);
        return next;
      });
      return newTheme;
    });
  }, []);

  /**
   * Loads saved and built-in templates from the backend.
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
   * Loads persisted settings from the backend.
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
   * Loads the persisted recent-contact-file list.
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
   * Also loads templates and settings needed by the initial UI.
   */
  useEffect(() => {
    const checkOutlook = async () => {
      try {
        const status = await GetOutlookStatus();
        if (status.available) {
          setOutlookStatus('ok');
        } else {
          setOutlookStatus('error');
          setError(status.message || 'Outlook is not available');
        }
      } catch (err: any) {
        setOutlookStatus('error');
        setError(`Outlook is not available: ${err.message || err}`);
      }
    };
    checkOutlook();

    // Get merge fields from backend
    GetMergeFields().then(setMergeFields).catch(console.error);

    // Load templates and settings after startup.
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
   * Parses the selected file, derives canonical merge tokens, and records it in
   * the recent-file list.
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
      
      // Persist the successfully imported file in the recent-file list.
      try {
        await AddRecentFile(filePath);
        loadRecentFiles();
      } catch (err) {
        console.error('Failed to add to recent files:', err);
      }
      
      // Select every imported contact by default using stable IDs.
      if (result.contacts && result.contacts.length > 0) {
        setSelectedIds(new Set(result.contacts.map((c) => c.id)));
      } else {
        setSelectedIds(new Set());
      }
      
      // Generate canonical merge tokens from the imported headers.
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
      setSelectedIds(new Set());
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
        setAttachments(prev => appendUniquePaths(prev, files));
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
   * Handle native file drop onto the attachment area. Wails exposes real file
   * paths via the drop event's file objects (path property); fall back to the
   * dialog when paths are unavailable.
   */
  const handleFilesDropped = useCallback((files: FileList) => {
    const paths: string[] = [];
    for (let i = 0; i < files.length; i++) {
      const p = (files[i] as any).path as string | undefined;
      if (p) paths.push(p);
    }
    if (paths.length > 0) {
      setAttachments(prev => appendUniquePaths(prev, paths));
    }
  }, []);

  /**
   * Fetch backend file metadata (name/size) whenever the attachment list
   * changes, so sizes and the total-size warning work reliably.
   */
  useEffect(() => {
    let cancelled = false;
    (async () => {
      const infos = await Promise.all(
        attachments.map(async (path) => {
          try {
            return await GetFileInfo(path);
          } catch {
            return new models.FileInfo({ name: path.split(/[/\\]/).pop() || path, path, size: 0 });
          }
        })
      );
      if (!cancelled) setAttachmentInfo(infos);
    })();
    return () => { cancelled = true; };
  }, [attachments]);

  /**
   * Send a test email to a specified address
   * Uses first contact's data for merge preview, or defaults
   * Sends a test message through the same campaign pipeline, including CC/BCC.
   */
  // Split CC/BCC into static vs template fields based on merge-field presence.
  const splitCcBcc = useCallback(() => {
    const ccHasMerge = cc.includes('{');
    const bccHasMerge = bcc.includes('{');
    return {
      cc: ccHasMerge ? '' : cc,
      bcc: bccHasMerge ? '' : bcc,
      ccTemplate: ccHasMerge ? cc : '',
      bccTemplate: bccHasMerge ? bcc : '',
    };
  }, [cc, bcc]);

  const selectedContacts = useMemo(
    () => contacts.filter((c) => selectedIds.has(c.id)),
    [contacts, selectedIds]
  );

  const handleSendTestEmail = useCallback(() => {
    if (contacts.length === 0) {
      setError('Import at least one contact before sending a test message.');
      return;
    }
    setShowTestSendModal(true);
  }, [contacts.length]);

  const handleSubmitTestEmail = useCallback(async (options: TestSendOptions) => {
    try {
      setIsSending(true);
      setError(null);

      const parts = splitCcBcc();
      const request = new models.TestEmailRequest({
        testAddress: options.testAddress,
        subjectTemplate: subject,
        bodyTemplate: body,
        isHTML,
        attachments,
        ...parts,
        contact: options.contact,
        overwriteEmail: options.overwriteEmail,
        draftOnly: options.draftOnly,
        sampleFirstName: options.contact.firstName || 'John',
        sampleLastName: options.contact.lastName || 'Doe',
      });

      await SendTestEmail(request);
      setShowTestSendModal(false);
      setSuccessMessage(options.draftOnly
        ? `Test draft created for ${options.testAddress}`
        : `Test email sent successfully to ${options.testAddress}`);
      setTimeout(() => setSuccessMessage(null), 5000);
    } catch (err: any) {
      setError(`Failed to process test email: ${err.message || err}`);
    } finally {
      setIsSending(false);
    }
  }, [subject, body, isHTML, attachments, splitCcBcc]);

  const executeBulkSend = useCallback(async (request: models.EmailRequest) => {
    try {
      setIsSending(true);
      setIsCancelling(false);
      setError(null);
      setProgress(null);
      setProgressLogs([]);
      setSendResult(null);
      setLastRequest(request);
      setShowPreflightModal(false);

      const result = await SendBulkEmails(request);
      setSendResult(result);
    } catch (err: any) {
      setError(`Failed to process campaign: ${err.message || err}`);
    } finally {
      setIsSending(false);
      setIsCancelling(false);
      setPendingRequest(null);
    }
  }, []);

  /**
   * Send to selected contacts through the backend campaign engine, which
   * preflights, dedupes, applies suppression, and returns a typed result.
   */
  const handleSendAll = useCallback(async () => {
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

    const request = new models.EmailRequest({
      contacts: selectedContacts,
      subjectTemplate: subject,
      bodyTemplate: body,
      isHTML,
      attachments,
      draftOnly,
      ...splitCcBcc(),
    });

    // Preflight first; block sending on any error.
    let pf: PreflightResult;
    try {
      pf = await PreflightCampaign(request);
    } catch (err: any) {
      setError(`Preflight failed: ${err.message || err}`);
      return;
    }
    setPreflight(pf);
    if (!pf.canSend) {
      setError(`Cannot send: ${pf.errors.map((e) => e.message).join('; ')}`);
      return;
    }

    setPendingRequest(request);
    if (settings.confirmSend) {
      setShowPreflightModal(true);
      return;
    }
    await executeBulkSend(request);
  }, [selectedContacts, subject, body, isHTML, attachments, draftOnly, splitCcBcc, settings.confirmSend, executeBulkSend]);

  const handleConfirmSend = useCallback(() => {
    if (pendingRequest) void executeBulkSend(pendingRequest);
  }, [pendingRequest, executeBulkSend]);

  /**
   * Retry only failed recipients via the backend, which appends attempts
   * without double-counting successes.
   */
  const handleRetryFailed = useCallback(async () => {
    if (!sendResult || sendResult.failed === 0 || (!sendResult.campaignId && !lastRequest)) {
      setError('No failed recipients to retry');
      return;
    }
    try {
      setIsSending(true);
      setError(null);
      setProgress(null);
      setProgressLogs([]);
      const result = sendResult.campaignId
        ? await RetryCampaign(sendResult.campaignId)
        : await RetryFailed(lastRequest!);
      setSendResult(result);
    } catch (err: any) {
      setError(`Failed to retry emails: ${err.message || err}`);
    } finally {
      setIsSending(false);
    }
  }, [lastRequest, sendResult]);

  /**
   * Cancel the in-flight campaign. Unattempted recipients are marked cancelled.
   */
  const handleCancel = useCallback(async () => {
    setIsCancelling(true);
    try {
      await CancelCampaign();
    } catch (err) {
      console.error('Cancel failed:', err);
    }
  }, []);

  /**
   * Renders a preview for the contact selected in the preview modal.
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

  // Reset campaign-specific UI state after a completed or dismissed run.
  const handleReset = useCallback(() => {
    setSendResult(null);
    setProgress(null);
    setProgressLogs([]);
    setContacts([]);
    setFileName(null);
    setSubject('');
    setBody('');
    setAttachments([]);
    setAttachmentInfo([]);
    setParseErrors([]);
    setParseWarnings([]);
    setSelectedIds(new Set());
    setHeaders([]);
    setDuplicates([]);
    setPreflight(null);
    setLastRequest(null);
    setDraftOnly(false);
    setPendingRequest(null);
    setShowPreflightModal(false);
  }, []);

  /**
   * Loads a saved template into the editor.
   */
  const handleLoadTemplate = useCallback((template: WailsEmailTemplate) => {
    setSubject(template.subject);
    setBody(template.body);
    setIsHTML(template.isHTML);
    setSuccessMessage(`Loaded template: ${template.name}`);
    setTimeout(() => setSuccessMessage(null), 3000);
  }, []);

  /**
   * Saves the current composition as a reusable template.
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
   * Deletes a user-created template and refreshes the list.
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
   * Persists validated settings and updates local UI state.
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
   * Clears the persisted recent-file list.
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
   * Re-imports a selected recent contact file.
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
        setSelectedIds(new Set(result.contacts.map((c) => c.id)));
      } else {
        setSelectedIds(new Set());
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

  const loadCampaignHistory = useCallback(async () => {
    try {
      setIsLoadingHistory(true);
      setCampaignHistory(await GetCampaignHistory());
    } catch (err: any) {
      setError(`Failed to load campaign history: ${err.message || err}`);
    } finally {
      setIsLoadingHistory(false);
    }
  }, []);

  const openCampaignHistory = useCallback(() => {
    setShowHistoryModal(true);
    void loadCampaignHistory();
  }, [loadCampaignHistory]);

  const handleHistoryRetry = useCallback(async (campaignId: string) => {
    try {
      setShowHistoryModal(false);
      setIsSending(true);
      setError(null);
      setProgress(null);
      setProgressLogs([]);
      const result = await RetryCampaign(campaignId);
      setSendResult(result);
    } catch (err: any) {
      setError(`Failed to retry campaign: ${err.message || err}`);
    } finally {
      setIsSending(false);
    }
  }, []);

  const handleHistoryDelete = useCallback(async (campaignId: string) => {
    try {
      await DeleteCampaign(campaignId);
      await loadCampaignHistory();
    } catch (err: any) {
      setError(`Failed to delete campaign: ${err.message || err}`);
    }
  }, [loadCampaignHistory]);

  const handleHistoryClear = useCallback(async () => {
    try {
      await ClearCampaignHistory();
      await loadCampaignHistory();
    } catch (err: any) {
      setError(`Failed to clear campaign history: ${err.message || err}`);
    }
  }, [loadCampaignHistory]);

  // Derive the selected-recipient count and send eligibility.
  const selectedCount = selectedIds.size;
  const canSend = selectedCount > 0 && subject.trim() && body.trim() && !isSending && outlookStatus === 'ok';

  /**
   * Registers keyboard shortcuts while the application is mounted.
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
              onClick={openCampaignHistory}
              title="Campaign history"
              aria-label="Open campaign history"
            >
              <History size={18} />
            </button>
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
            onRetryFailed={sendResult.failed > 0 ? handleRetryFailed : undefined}
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
                  selectedIds={selectedIds}
                  onSelectionChange={setSelectedIds}
                  duplicates={duplicates}
                />
              </div>

              {/* Right Column */}
              <div>
                {/* Template manager */}
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
                  onFilesDropped={handleFilesDropped}
                  attachmentInfo={attachmentInfo}
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
                <label className="send-mode-option">
                  <input
                    type="checkbox"
                    checked={draftOnly}
                    onChange={(event) => setDraftOnly(event.target.checked)}
                    disabled={isSending}
                  />
                  <Save size={17} />
                  Save campaign messages to Outlook Drafts instead of sending
                </label>
                <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', marginTop: '1rem' }}>
                  {/* Preview button */}
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
                        {draftOnly ? <Save size={18} /> : <Send size={18} />}
                        {draftOnly ? 'Create Drafts' : 'Send to Selected'} {selectedCount > 0 && `(${selectedCount})`}
                      </>
                    )}
                  </button>
                  {isSending && (
                    <button
                      className="btn btn-warning btn-lg"
                      onClick={handleCancel}
                      disabled={isCancelling}
                      aria-label="Cancel the current send"
                    >
                      {isCancelling ? 'Cancelling…' : 'Cancel'}
                    </button>
                  )}
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

        {/* Preview modal */}
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

        <TestSendModal
          isOpen={showTestSendModal}
          contacts={selectedContacts.length > 0 ? selectedContacts : contacts}
          subject={subject}
          body={body}
          isHTML={isHTML}
          isSubmitting={isSending}
          onClose={() => setShowTestSendModal(false)}
          onPreview={handlePreviewEmail}
          onSubmit={handleSubmitTestEmail}
        />

        <PreflightModal
          isOpen={showPreflightModal}
          result={preflight}
          draftOnly={draftOnly}
          totalAttachmentBytes={attachmentInfo.reduce((total, info) => total + info.size, 0)}
          isSubmitting={isSending}
          onClose={() => { setShowPreflightModal(false); setPendingRequest(null); }}
          onConfirm={handleConfirmSend}
        />

        <CampaignHistoryModal
          isOpen={showHistoryModal}
          records={campaignHistory}
          isLoading={isLoadingHistory}
          onClose={() => setShowHistoryModal(false)}
          onRefresh={loadCampaignHistory}
          onRetry={handleHistoryRetry}
          onDelete={handleHistoryDelete}
          onClear={handleHistoryClear}
        />

        {/* Settings modal */}
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
