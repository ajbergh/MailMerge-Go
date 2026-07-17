/**
 * AttachmentManager Component - Email Attachment Handler
 * 
 * This component provides the UI for managing email attachments.
 * Users can add multiple files and remove individual attachments.
 * 
 * Features:
 * - Add multiple attachments via native file dialog
 * - Drag-and-drop files directly onto the attachment area
 * - Display list of attached files with filenames
 * - Remove individual attachments with one click
 * - Empty state when no attachments added
 * 
 * Note: Attachments are stored as full file paths and validated
 * by the backend before sending.
 */
import { useState, useCallback, DragEvent } from 'react';
import { Paperclip, X, Plus, File, Upload, AlertTriangle } from 'lucide-react';

/**
 * FileInfo type for attachment metadata
 */
interface FileInfo {
  path: string;
  name: string;
  size?: number;
}

/**
 * Props for the AttachmentManager component
 */
interface AttachmentManagerProps {
  /** Array of attachment file paths */
  attachments: string[];
  /** Callback to open file dialog and add attachments */
  onAddAttachments: () => void;
  /** Callback to remove an attachment by index */
  onRemoveAttachment: (index: number) => void;
  /** Optional callback that adds files dropped on the attachment area */
  onFilesDropped?: (files: FileList) => void;
  /** Optional file info with sizes */
  attachmentInfo?: FileInfo[];
}

// Maximum total attachment size in bytes (20MB)
const MAX_TOTAL_SIZE = 20 * 1024 * 1024;

/**
 * AttachmentManager - Renders the attachment management card with drag-and-drop
 * 
 * Displays an "Add Attachments" button, drag-and-drop zone, and a list of 
 * currently attached files with remove buttons.
 */
export function AttachmentManager({ 
  attachments, 
  onAddAttachments, 
  onRemoveAttachment,
  onFilesDropped,
  attachmentInfo
}: AttachmentManagerProps) {
  // Drag state for visual feedback
  const [isDragging, setIsDragging] = useState(false);

  /**
   * Extract filename from a full file path
   * Handles both forward and back slashes for cross-platform compatibility
   */
  const getFileName = (path: string): string => {
    return path.split(/[/\\]/).pop() || path;
  };

  /**
   * Format file size in human-readable format
   */
  const formatSize = (bytes: number): string => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  };

  /**
   * Calculate total attachment size
   */
  const totalSize = attachmentInfo?.reduce((acc, info) => acc + (info.size || 0), 0) || 0;
  const sizeExceedsLimit = totalSize > MAX_TOTAL_SIZE;

  /**
   * Handle drag enter event
   */
  const handleDragEnter = useCallback((e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.dataTransfer.types.includes('Files')) {
      setIsDragging(true);
    }
  }, []);

  /**
   * Handle drag leave event
   */
  const handleDragLeave = useCallback((e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    // Only set to false if we're leaving the drop zone entirely
    const rect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX;
    const y = e.clientY;
    if (x < rect.left || x >= rect.right || y < rect.top || y >= rect.bottom) {
      setIsDragging(false);
    }
  }, []);

  /**
   * Handle drag over event (required to allow drop)
   */
  const handleDragOver = useCallback((e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  /**
   * Handle file drop event
   */
  const handleDrop = useCallback((e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    if (onFilesDropped && e.dataTransfer.files.length > 0) {
      onFilesDropped(e.dataTransfer.files);
    }
  }, [onFilesDropped]);

  return (
    <div className="card">
      <div className="card-header">
        <h2>
          <span className="step-badge">3</span>
          Attachments
          {attachments.length > 0 && (
            <span style={{ fontWeight: 'normal', fontSize: '0.875rem', marginLeft: '0.5rem', color: 'var(--text-muted)' }}>
              ({attachments.length} file{attachments.length !== 1 ? 's' : ''})
            </span>
          )}
        </h2>
      </div>
      <div className="card-body">
        {/* Drag-and-drop zone */}
        <div
          onDragEnter={handleDragEnter}
          onDragLeave={handleDragLeave}
          onDragOver={handleDragOver}
          onDrop={handleDrop}
          onClick={onAddAttachments}
          style={{
            border: `2px dashed ${isDragging ? 'var(--primary)' : 'var(--border-color-strong)'}`,
            borderRadius: 'var(--border-radius)',
            padding: '1.5rem',
            textAlign: 'center',
            cursor: 'pointer',
            backgroundColor: isDragging ? 'var(--primary-light)' : 'var(--bg-secondary)',
            transition: 'all 0.2s ease',
            marginBottom: attachments.length > 0 ? '1rem' : 0
          }}
        >
          <Upload 
            size={24} 
            style={{ 
              margin: '0 auto 0.5rem', 
              color: isDragging ? 'var(--primary)' : 'var(--text-muted)' 
            }} 
          />
          <p style={{ margin: 0, color: isDragging ? 'var(--primary)' : 'var(--text-secondary)' }}>
            {isDragging 
              ? 'Drop files here...' 
              : 'Click to add attachments or drag and drop files here'
            }
          </p>
        </div>

        {/* Size warning */}
        {sizeExceedsLimit && (
          <div className="alert alert-warning" style={{ marginBottom: '1rem' }}>
            <AlertTriangle size={16} />
            <span>
              Total attachment size ({formatSize(totalSize)}) exceeds recommended limit of 20MB. 
              Some email servers may reject large attachments.
            </span>
          </div>
        )}

        {/* Attachment list */}
        {attachments.length > 0 && (
          <div className="attachments-list">
            {attachments.map((attachment, index) => {
              const info = attachmentInfo?.find(i => i.path === attachment);
              return (
                <div key={index} className="attachment-item">
                  <div className="file-name">
                    <File size={16} style={{ color: 'var(--primary)' }} />
                    <span>{getFileName(attachment)}</span>
                    {info?.size && (
                      <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem', marginLeft: '0.5rem' }}>
                        ({formatSize(info.size)})
                      </span>
                    )}
                  </div>
                  <button
                    className="remove-btn"
                    onClick={() => onRemoveAttachment(index)}
                    title="Remove attachment"
                  >
                    <X size={16} />
                  </button>
                </div>
              );
            })}
            {/* Total size */}
            {totalSize > 0 && (
              <div style={{ 
                marginTop: '0.5rem', 
                fontSize: '0.8125rem', 
                color: sizeExceedsLimit ? 'var(--danger)' : 'var(--text-muted)',
                textAlign: 'right'
              }}>
                Total: {formatSize(totalSize)}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
