/**
 * Type Definitions - Frontend Type Aliases and Interfaces
 * 
 * This module provides TypeScript type definitions for the MailMerge application.
 * It includes both re-exports of Wails-generated types and custom frontend-only types.
 * 
 * For types that are bound from the Go backend (Contact, EmailLog, SendResult, etc.),
 * use the types from '../wailsjs/go/models' directly. These are automatically generated
 * by Wails during the build process.
 * 
 * This file defines additional types used only in the frontend, such as ProgressUpdate
 * for real-time event handling.
 * 
 * Phase 1 Updates (v1.2):
 *   - Contact now includes customFields for dynamic merge field support
 *   - ParseResult includes warnings, duplicates, and headers
 *   - SendResult includes failedContacts for retry functionality
 * 
 * Phase 2 Updates (v1.3):
 *   - EmailTemplate for saved email templates
 *   - AppSettings for user preferences
 */

// Re-export Wails types for convenience
export { models } from '../../wailsjs/go/models';

/**
 * Contact type alias - represents a single email recipient
 * Mirrors the Go models.Contact struct
 * Phase 1: Added customFields for dynamic merge field support
 */
export type Contact = {
  firstName: string;                    // Recipient's first name (for personalization)
  lastName: string;                     // Recipient's last name (for personalization)
  email: string;                        // Recipient's email address (required)
  customFields?: Record<string, string>; // Additional columns from import file
};

/**
 * ParseResult type alias - result of parsing a contact file
 * Phase 1: Added warnings, duplicates, and headers
 */
export type ParseResult = {
  contacts: Contact[];      // Successfully parsed contacts
  errors?: string[];        // Row-level parsing errors
  warnings?: string[];      // Non-fatal warnings (e.g., duplicates found)
  total: number;            // Total number of valid contacts
  headers?: string[];       // All detected column headers from file
  duplicates?: string[];    // List of duplicate email addresses
};

/**
 * EmailLog type alias - represents the result of sending a single email
 * Used for tracking success/failure and generating export reports
 */
export type EmailLog = {
  firstName: string;       // Contact's first name
  lastName: string;        // Contact's last name
  email: string;           // Contact's email address
  status: string;          // "Success" or "Failure"
  errorMessage?: string;   // Error details if failed (omitted on success)
  timestamp: any;          // ISO 8601 timestamp of send attempt
};

/**
 * SendResult type alias - aggregate result of bulk send operation
 * Contains summary counts and detailed per-email logs
 * Phase 1: Added failedContacts for retry functionality
 */
export type SendResult = {
  totalSent: number;          // Number of emails sent successfully
  totalFailed: number;        // Number of emails that failed
  logs: EmailLog[];           // Detailed log for each email attempt
  failedContacts?: Contact[]; // Contacts that failed, for retry
};

/**
 * ProgressUpdate - real-time progress event payload
 * 
 * This interface is used for events emitted by the backend during bulk
 * email sending. The frontend subscribes to "email:progress" events to
 * receive these updates and display live progress to the user.
 */
export interface ProgressUpdate {
  current: number;       // Current email number being processed (1-based)
  total: number;         // Total number of emails to send
  status: 'sending' | 'success' | 'failure' | 'complete';  // Current operation status
  email: string;         // Email address currently being processed
  message?: string;      // Optional message (typically error details)
  timestamp: string;     // ISO 8601 timestamp of this update
}

// ==================== Phase 2 Types ====================

/**
 * EmailTemplate - saved email template for reuse
 * Phase 2: New type for template persistence
 */
export type EmailTemplate = {
  id: string;           // Unique identifier (UUID)
  name: string;         // User-friendly template name
  subject: string;      // Subject line with merge fields
  body: string;         // Body content with merge fields
  isHTML: boolean;      // True for HTML format, false for plain text
  isBuiltIn: boolean;   // True if this is a built-in template
  createdAt: string;    // ISO 8601 timestamp when created
  updatedAt: string;    // ISO 8601 timestamp when last modified
};

/**
 * AppSettings - user preferences and application settings
 * Phase 2: New type for centralized settings management
 */
export type AppSettings = {
  // Display Settings
  theme: 'light' | 'dark' | 'system';  // Theme preference
  defaultFormat: 'html' | 'plaintext'; // Default email format

  // Sending Settings
  sendingDelay: number;    // Delay between emails in milliseconds
  confirmSend: boolean;    // Show confirmation before sending
  soundEnabled: boolean;   // Play sound on completion
  autoSaveTempls: boolean; // Auto-save templates on exit

  // Recent Files
  recentFiles: string[];   // Last 10 opened contact files
};
