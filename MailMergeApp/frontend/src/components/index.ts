/**
 * Components Index - Central Export for UI Components
 * 
 * This barrel file exports all reusable UI components used in the MailMerge application.
 * Import components from this file for cleaner imports throughout the application.
 * 
 * Example: import { FileUpload, ContactTable, EmailEditor } from './components';
 * 
 * Phase 1 Updates (v1.2):
 *   - Added PreviewModal for email preview functionality
 * 
 * Phase 2 Updates (v1.3):
 *   - Added TemplateManager for email template management
 *   - Added SettingsModal for centralized settings panel
 */
export { FileUpload } from './FileUpload';
export { ContactTable } from './ContactTable';
export { EmailEditor } from './EmailEditor';
export { AttachmentManager } from './AttachmentManager';
export { ProgressTracker } from './ProgressTracker';
export { ResultsSummary } from './ResultsSummary';
export { PreviewModal } from './PreviewModal';
export { TemplateManager } from './TemplateManager';
export { SettingsModal } from './SettingsModal';
