/**
 * FileUpload Component - Contact File Import Interface
 * 
 * This component provides the UI for importing contact lists from CSV or Excel files.
 * It displays the current file status, loading states, and any parsing errors.
 * 
 * Features:
 * - Click-to-upload interaction
 * - Loading spinner during file parsing
 * - Success state showing file name and contact count
 * - Warning display for row-level parsing errors
 * - Option to select a different file after loading
 */
import { Upload, FileSpreadsheet, CheckCircle } from 'lucide-react';
import { Contact } from '../types';

/**
 * Props for the FileUpload component
 */
interface FileUploadProps {
  /** Callback triggered when user clicks to select a file */
  onFileSelect: () => void;
  /** Name of the currently loaded file, or null if no file loaded */
  fileName: string | null;
  /** Array of contacts parsed from the file */
  contacts: Contact[];
  /** Array of parsing errors/warnings to display */
  errors: string[];
  /** Whether file is currently being loaded/parsed */
  isLoading: boolean;
}

/**
 * FileUpload - Renders the contact file import card
 * 
 * Displays different states:
 * - Empty: Shows upload prompt with supported format info
 * - Loading: Shows spinner and "Loading contacts..." message
 * - Loaded: Shows success icon, filename, contact count, and change file button
 */
export function FileUpload({ onFileSelect, fileName, contacts, errors, isLoading }: FileUploadProps) {
  const hasFile = fileName !== null;

  return (
    <div className="card">
      <div className="card-header">
        <h2>
          <span className="step-badge">1</span>
          Import Contacts
        </h2>
      </div>
      <div className="card-body">
        <div 
          className={`file-upload-area ${hasFile ? 'has-file' : ''}`}
          onClick={onFileSelect}
        >
          {isLoading ? (
            <>
              <div className="spinner" style={{ margin: '0 auto 1rem' }}></div>
              <p>Loading contacts...</p>
            </>
          ) : hasFile ? (
            <>
              <CheckCircle className="file-upload-icon" style={{ color: 'var(--success)' }} />
              <p style={{ fontWeight: 500, color: 'var(--success)' }}>{fileName}</p>
              <p style={{ fontSize: '0.875rem', color: 'var(--gray-600)', marginTop: '0.5rem' }}>
                {contacts.length} contacts loaded
              </p>
              <button 
                className="btn btn-secondary btn-sm" 
                style={{ marginTop: '1rem' }}
                onClick={(e) => { e.stopPropagation(); onFileSelect(); }}
              >
                Choose Different File
              </button>
            </>
          ) : (
            <>
              <Upload className="file-upload-icon" />
              <p style={{ fontWeight: 500 }}>Click to select a file</p>
              <p style={{ fontSize: '0.875rem', color: 'var(--gray-500)', marginTop: '0.5rem' }}>
                Supports CSV and Excel (.xlsx) files
              </p>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', justifyContent: 'center', marginTop: '1rem' }}>
                <FileSpreadsheet size={16} />
                <span style={{ fontSize: '0.8125rem', color: 'var(--gray-500)' }}>
                  Required columns: FirstName, LastName, Email
                </span>
              </div>
            </>
          )}
        </div>

        {errors.length > 0 && (
          <div className="alert alert-warning" style={{ marginTop: '1rem' }}>
            <div>
              <strong>Import Warnings:</strong>
              <ul style={{ margin: '0.5rem 0 0 1rem', padding: 0 }}>
                {errors.slice(0, 5).map((error, index) => (
                  <li key={index}>{error}</li>
                ))}
                {errors.length > 5 && <li>...and {errors.length - 5} more</li>}
              </ul>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
