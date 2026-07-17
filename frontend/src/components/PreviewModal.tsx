/**
 * PreviewModal Component - Email Preview Modal
 * 
 * Features:
 * - Contact selector dropdown to preview for any contact
 * - Shows subject and body rendered by the backend merge pipeline
 * - Displays HTML emails in an iframe for accurate preview
 * - Shows attachments that will be included
 * - Close button and click-outside-to-close functionality
 */
import { useState, useEffect, useMemo } from 'react';
import { Contact } from '../types';
import { X, Eye, User, Paperclip, ChevronDown } from 'lucide-react';
import DOMPurify from 'dompurify';

/**
 * Props for the PreviewModal component
 */
interface PreviewModalProps {
  /** Whether the modal is visible */
  isOpen: boolean;
  /** Callback to close the modal */
  onClose: () => void;
  /** List of contacts to choose from for preview */
  contacts: Contact[];
  /** Email subject template with merge fields (for display context) */
  subject: string;
  /** Email body template with merge fields (for display context) */
  body: string;
  /** Whether the body is HTML */
  isHTML: boolean;
  /** List of attachment file paths */
  attachments: string[];
  /** Function to preview merge for a contact (calls backend) */
  onPreview: (contact: Contact) => Promise<{ subject: string; body: string } | null>;
}

/**
 * PreviewModal - Shows a preview of the email for a selected contact
 * 
 * Allows users to see exactly how their email will look with merge fields
 * replaced with actual contact data before sending.
 */
export function PreviewModal({
  isOpen,
  onClose,
  contacts,
  subject,
  body,
  isHTML,
  attachments,
  onPreview
}: PreviewModalProps) {
  // Selected contact index for preview
  const [selectedIndex, setSelectedIndex] = useState(0);
  // Rendered preview content
  const [previewSubject, setPreviewSubject] = useState('');
  const [previewBody, setPreviewBody] = useState('');
  // Loading state
  const [isLoading, setIsLoading] = useState(false);

  // Get the currently selected contact
  const selectedContact = useMemo(() => 
    contacts[selectedIndex] || null,
    [contacts, selectedIndex]
  );

  // Update preview when contact or templates change
  useEffect(() => {
    if (!isOpen || !selectedContact) return;

    const updatePreview = async () => {
      setIsLoading(true);
      try {
        const result = await onPreview(selectedContact);
        if (result) {
          setPreviewSubject(result.subject);
          setPreviewBody(result.body);
        } else {
          setPreviewSubject(subject);
          setPreviewBody(body);
        }
      } catch (error) {
        console.error('Failed to generate preview:', error);
        setPreviewSubject(subject);
        setPreviewBody(body);
      } finally {
        setIsLoading(false);
      }
    };

    updatePreview();
  }, [isOpen, selectedContact, subject, body, onPreview]);

  // Reset to first contact when modal opens
  useEffect(() => {
    if (isOpen && contacts.length > 0) {
      setSelectedIndex(0);
    }
  }, [isOpen, contacts.length]);

  // Extract file name from path
  const getFileName = (path: string): string => {
    const parts = path.split(/[/\\]/);
    return parts[parts.length - 1] || path;
  };

  // Wrap HTML content with proper styling for iframe display
  // Also clean up ReactQuill's HTML output to prevent double spacing
  const getStyledHtmlContent = (html: string): string => {
    // Clean up ReactQuill's HTML output:
    // 1. Remove empty paragraphs that only contain <br> (ReactQuill adds these for blank lines)
    // 2. Replace consecutive closing/opening p tags with line breaks to avoid double spacing
    // Sanitize before display (defense in depth; the iframe is also sandboxed
    // without allow-scripts).
    let cleanedHtml = DOMPurify.sanitize(html)
      // Convert <p><br></p> (empty lines from ReactQuill) to just <br> for single line break
      .replace(/<p><br><\/p>/gi, '<br>')
      // Handle <p><br/></p> variant
      .replace(/<p><br\s*\/?><\/p>/gi, '<br>')
      // Remove trailing <br> inside paragraphs before closing tag (ReactQuill artifact)
      .replace(/<br><\/p>/gi, '</p>')
      .replace(/<br\s*\/?><\/p>/gi, '</p>');
    
    return `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
      font-size: 14px;
      line-height: 1.6;
      color: #333;
      margin: 16px;
      padding: 0;
    }
    p {
      margin: 0;
      padding: 0;
      min-height: 1.6em;
    }
    p + p {
      margin-top: 0;
    }
    p:empty {
      min-height: 1.6em;
    }
    br {
      line-height: 1.6;
    }
    a {
      color: #0066cc;
    }
    h1, h2, h3, h4, h5, h6 {
      margin: 0 0 0.5em 0;
    }
    ul, ol {
      margin: 0 0 1em 0;
      padding-left: 1.5em;
    }
    blockquote {
      margin: 0 0 1em 0;
      padding-left: 1em;
      border-left: 3px solid #ccc;
    }
    strong, b {
      font-weight: bold;
    }
    em, i {
      font-style: italic;
    }
    u {
      text-decoration: underline;
    }
    s, strike {
      text-decoration: line-through;
    }
  </style>
</head>
<body>${cleanedHtml}</body>
</html>`;
  };

  if (!isOpen) return null;

  return (
    <div 
      className="modal-overlay"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'rgba(0, 0, 0, 0.5)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 1000,
        padding: '2rem'
      }}
    >
      <div 
        className="modal-content"
        style={{
          backgroundColor: 'var(--card-bg)',
          borderRadius: 'var(--border-radius)',
          boxShadow: 'var(--shadow-lg)',
          width: '100%',
          maxWidth: '700px',
          maxHeight: '90vh',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden'
        }}
      >
        {/* Modal Header */}
        <div 
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            padding: '1rem 1.5rem',
            borderBottom: '1px solid var(--border-color)',
            backgroundColor: 'var(--card-header-bg)'
          }}
        >
          <h2 style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', margin: 0 }}>
            <Eye size={20} />
            Email Preview
          </h2>
          <button
            onClick={onClose}
            className="btn btn-sm btn-outline"
            style={{ padding: '0.25rem' }}
            title="Close preview"
          >
            <X size={20} />
          </button>
        </div>

        {/* Contact Selector */}
        <div style={{ padding: '1rem 1.5rem', borderBottom: '1px solid var(--border-color)' }}>
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
            <User size={16} />
            <span>Preview for contact:</span>
          </label>
          <div style={{ position: 'relative' }}>
            <select
              value={selectedIndex}
              onChange={(e) => setSelectedIndex(Number(e.target.value))}
              className="form-input"
              style={{ 
                appearance: 'none', 
                paddingRight: '2rem',
                cursor: 'pointer'
              }}
            >
              {contacts.map((contact, index) => (
                <option key={index} value={index}>
                  {contact.firstName || ''} {contact.lastName || ''} ({contact.email})
                </option>
              ))}
            </select>
            <ChevronDown 
              size={16} 
              style={{ 
                position: 'absolute', 
                right: '0.75rem', 
                top: '50%', 
                transform: 'translateY(-50%)',
                pointerEvents: 'none',
                color: 'var(--text-muted)'
              }} 
            />
          </div>
        </div>

        {/* Preview Content */}
        <div style={{ flex: 1, overflow: 'auto', padding: '1.5rem' }}>
          {isLoading ? (
            <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-muted)' }}>
              <div className="spinner" style={{ margin: '0 auto 1rem' }}></div>
              <p>Generating preview...</p>
            </div>
          ) : (
            <>
              {/* Subject */}
              <div style={{ marginBottom: '1.5rem' }}>
                <label style={{ 
                  display: 'block', 
                  fontWeight: 600, 
                  marginBottom: '0.5rem',
                  color: 'var(--text-secondary)'
                }}>
                  Subject:
                </label>
                <div style={{ 
                  padding: '0.75rem 1rem',
                  backgroundColor: 'var(--bg-tertiary)',
                  borderRadius: 'var(--border-radius)',
                  border: '1px solid var(--border-color)',
                  fontWeight: 500
                }}>
                  {previewSubject || <span style={{ color: 'var(--text-muted)', fontStyle: 'italic' }}>(No subject)</span>}
                </div>
              </div>

              {/* Body */}
              <div style={{ marginBottom: '1.5rem' }}>
                <label style={{ 
                  display: 'block', 
                  fontWeight: 600, 
                  marginBottom: '0.5rem',
                  color: 'var(--text-secondary)'
                }}>
                  Body:
                </label>
                {isHTML ? (
                  <div 
                    style={{ 
                      border: '1px solid var(--border-color)',
                      borderRadius: 'var(--border-radius)',
                      overflow: 'hidden',
                      backgroundColor: '#fff'
                    }}
                  >
                    <iframe
                      srcDoc={getStyledHtmlContent(previewBody)}
                      title="Email Preview"
                      style={{
                        width: '100%',
                        minHeight: '300px',
                        border: 'none'
                      }}
                      sandbox="allow-same-origin"
                    />
                  </div>
                ) : (
                  <div style={{ 
                    padding: '1rem',
                    backgroundColor: 'var(--bg-tertiary)',
                    borderRadius: 'var(--border-radius)',
                    border: '1px solid var(--border-color)',
                    whiteSpace: 'pre-wrap',
                    fontFamily: 'monospace',
                    minHeight: '200px'
                  }}>
                    {previewBody || <span style={{ color: 'var(--text-muted)', fontStyle: 'italic' }}>(No body)</span>}
                  </div>
                )}
              </div>

              {/* Attachments */}
              {attachments.length > 0 && (
                <div>
                  <label style={{ 
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    fontWeight: 600, 
                    marginBottom: '0.5rem',
                    color: 'var(--text-secondary)'
                  }}>
                    <Paperclip size={16} />
                    Attachments ({attachments.length}):
                  </label>
                  <div style={{ 
                    display: 'flex',
                    flexWrap: 'wrap',
                    gap: '0.5rem'
                  }}>
                    {attachments.map((path, index) => (
                      <span 
                        key={index}
                        style={{
                          padding: '0.25rem 0.75rem',
                          backgroundColor: 'var(--bg-tertiary)',
                          borderRadius: '9999px',
                          fontSize: '0.875rem',
                          border: '1px solid var(--border-color)'
                        }}
                      >
                        {getFileName(path)}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </>
          )}
        </div>

        {/* Modal Footer */}
        <div 
          style={{
            padding: '1rem 1.5rem',
            borderTop: '1px solid var(--border-color)',
            display: 'flex',
            justifyContent: 'flex-end'
          }}
        >
          <button onClick={onClose} className="btn btn-secondary">
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
