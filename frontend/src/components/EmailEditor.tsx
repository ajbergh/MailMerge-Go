/**
 * EmailEditor Component - Email Composition Interface
 * 
 * This component provides a rich email composition experience with:
 * - Subject line input with merge field insertion
 * - CC and BCC recipient fields with merge field support
 * - Body editor with both plain text and rich text (HTML) modes
 * - Merge field helper buttons for easy personalization
 * - Preview rendered by the same backend pipeline used for sending
 * 
 * The editor uses ReactQuill for rich text editing, providing a familiar
 * word-processor-like experience with formatting options.
 * 
 * Merge fields use canonical tokens such as `{{first_name}}`, `{{last_name}}`,
 * and `{{email}}`. Legacy single-brace tokens remain supported by the backend
 * for existing templates.
 */
import { useState, useRef, useEffect } from 'react';
import ReactQuill from 'react-quill';
import 'react-quill/dist/quill.snow.css';
import { Mail, Type, FileText, Wand2, ChevronDown, ChevronRight, Users } from 'lucide-react';
import DOMPurify from 'dompurify';
import { PreviewMerge } from '../../wailsjs/go/main/App';

/**
 * Props for the email composition editor.
 */
interface EmailEditorProps {
  /** Current email subject text */
  subject: string;
  /** Current email body text (HTML or plain text depending on mode) */
  body: string;
  /** Whether body is HTML (true) or plain text (false) */
  isHTML: boolean;
  /** CC recipients (comma-separated, supports merge fields) */
  cc: string;
  /** BCC recipients (comma-separated, supports merge fields) */
  bcc: string;
  /** Callback when subject changes */
  onSubjectChange: (subject: string) => void;
  /** Callback when body changes */
  onBodyChange: (body: string) => void;
  /** Callback when HTML mode toggle changes */
  onIsHTMLChange: (isHTML: boolean) => void;
  /** Callback when CC changes */
  onCCChange: (cc: string) => void;
  /** Callback when BCC changes */
  onBCCChange: (bcc: string) => void;
  /** Array of available merge fields to display as buttons */
  mergeFields: string[];
}

/**
 * EmailEditor - Full email composition interface
 * 
 * Provides tabbed interface with Compose and Preview modes.
 * Supports inserting merge fields at the cursor in the subject, CC, BCC, or body.
 */
export function EmailEditor({
  subject,
  body,
  isHTML,
  cc,
  bcc,
  onSubjectChange,
  onBodyChange,
  onIsHTMLChange,
  onCCChange,
  onBCCChange,
  mergeFields
}: EmailEditorProps) {
  const [activeTab, setActiveTab] = useState<'compose' | 'preview'>('compose');
  const [showCcBcc, setShowCcBcc] = useState(false);
  const subjectInputRef = useRef<HTMLInputElement>(null);
  const ccInputRef = useRef<HTMLInputElement>(null);
  const bccInputRef = useRef<HTMLInputElement>(null);
  const bodyTextareaRef = useRef<HTMLTextAreaElement>(null);
  const quillRef = useRef<ReactQuill>(null);

  // Auto-expand CC/BCC section if fields have values
  const hasCcBccValues = cc.trim() !== '' || bcc.trim() !== '';

  const insertMergeField = (field: string) => {
    if (document.activeElement === subjectInputRef.current) {
      // Insert into subject
      const input = subjectInputRef.current!;
      const start = input.selectionStart || 0;
      const end = input.selectionEnd || 0;
      const newSubject = subject.slice(0, start) + field + subject.slice(end);
      onSubjectChange(newSubject);
      setTimeout(() => {
        input.focus();
        input.setSelectionRange(start + field.length, start + field.length);
      }, 0);
    } else if (document.activeElement === ccInputRef.current) {
      // Insert into CC field
      const input = ccInputRef.current!;
      const start = input.selectionStart || 0;
      const end = input.selectionEnd || 0;
      const newCC = cc.slice(0, start) + field + cc.slice(end);
      onCCChange(newCC);
      setTimeout(() => {
        input.focus();
        input.setSelectionRange(start + field.length, start + field.length);
      }, 0);
    } else if (document.activeElement === bccInputRef.current) {
      // Insert into BCC field
      const input = bccInputRef.current!;
      const start = input.selectionStart || 0;
      const end = input.selectionEnd || 0;
      const newBCC = bcc.slice(0, start) + field + bcc.slice(end);
      onBCCChange(newBCC);
      setTimeout(() => {
        input.focus();
        input.setSelectionRange(start + field.length, start + field.length);
      }, 0);
    } else if (!isHTML && document.activeElement === bodyTextareaRef.current) {
      // Insert into plain text body
      const textarea = bodyTextareaRef.current!;
      const start = textarea.selectionStart || 0;
      const end = textarea.selectionEnd || 0;
      const newBody = body.slice(0, start) + field + body.slice(end);
      onBodyChange(newBody);
      setTimeout(() => {
        textarea.focus();
        textarea.setSelectionRange(start + field.length, start + field.length);
      }, 0);
    } else if (isHTML && quillRef.current) {
      // Insert into Quill editor
      const quill = quillRef.current.getEditor();
      const range = quill.getSelection();
      if (range) {
        quill.insertText(range.index, field);
        quill.setSelection(range.index + field.length, 0);
      } else {
        quill.insertText(quill.getLength() - 1, field);
      }
    }
  };

  const quillModules = {
    toolbar: [
      [{ 'header': [1, 2, 3, false] }],
      ['bold', 'italic', 'underline', 'strike'],
      [{ 'color': [] }, { 'background': [] }],
      [{ 'list': 'ordered' }, { 'list': 'bullet' }],
      [{ 'align': [] }],
      ['link'],
      ['clean']
    ]
  };

  // Preview rendering uses the backend canonical renderer (the same one used by
  // test and bulk sends) with sample data, rather than ad-hoc frontend
  // replacements, so the preview matches what recipients receive.
  const [previewSubject, setPreviewSubject] = useState('');
  const [previewBody, setPreviewBody] = useState('');
  const [previewCc, setPreviewCc] = useState('');
  const [previewBcc, setPreviewBcc] = useState('');

  useEffect(() => {
    if (activeTab !== 'preview') return;
    let cancelled = false;
    (async () => {
      try {
        const main = await PreviewMerge(subject, body, '', '', '');
        if (cancelled) return;
        setPreviewSubject(main.subject || '');
        setPreviewBody(main.body || '');
        if (cc.trim()) {
          const r = await PreviewMerge('', cc, '', '', '');
          if (!cancelled) setPreviewCc(r.body || '');
        } else {
          setPreviewCc('');
        }
        if (bcc.trim()) {
          const r = await PreviewMerge('', bcc, '', '', '');
          if (!cancelled) setPreviewBcc(r.body || '');
        } else {
          setPreviewBcc('');
        }
      } catch (err) {
        console.error('Preview render failed:', err);
      }
    })();
    return () => { cancelled = true; };
  }, [activeTab, subject, body, cc, bcc]);

  return (
    <div className="card">
      <div className="card-header">
        <h2>
          <span className="step-badge">2</span>
          Compose Email
        </h2>
      </div>
      <div className="card-body">
        {/* Tabs */}
        <div className="tabs">
          <div 
            className={`tab ${activeTab === 'compose' ? 'active' : ''}`}
            onClick={() => setActiveTab('compose')}
          >
            <Mail size={16} style={{ marginRight: '0.5rem', verticalAlign: 'middle' }} />
            Compose
          </div>
          <div 
            className={`tab ${activeTab === 'preview' ? 'active' : ''}`}
            onClick={() => setActiveTab('preview')}
          >
            <FileText size={16} style={{ marginRight: '0.5rem', verticalAlign: 'middle' }} />
            Preview
          </div>
        </div>

        {activeTab === 'compose' ? (
          <>
            {/* Merge Fields Helper */}
            <div style={{ marginBottom: '1rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
                <Wand2 size={16} />
                <span style={{ fontSize: '0.875rem', fontWeight: 500 }}>Insert Merge Fields:</span>
              </div>
              <div className="merge-fields">
                {mergeFields.map((field) => (
                  <span 
                    key={field} 
                    className="merge-field-tag"
                    onClick={() => insertMergeField(field)}
                    title="Click to insert at cursor position"
                  >
                    {field}
                  </span>
                ))}
              </div>
            </div>

            {/* Subject */}
            <div className="form-group">
              <label className="form-label">Subject</label>
              <input
                ref={subjectInputRef}
                type="text"
                className="form-input"
                placeholder="Enter email subject..."
                value={subject}
                onChange={(e) => onSubjectChange(e.target.value)}
              />
            </div>

            {/* CC/BCC Section - Collapsible */}
            <div className="form-group">
              <div 
                style={{ 
                  display: 'flex', 
                  alignItems: 'center', 
                  gap: '0.5rem', 
                  cursor: 'pointer',
                  marginBottom: showCcBcc || hasCcBccValues ? '0.75rem' : 0
                }}
                onClick={() => setShowCcBcc(!showCcBcc)}
              >
                {showCcBcc || hasCcBccValues ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                <Users size={16} />
                <span style={{ fontSize: '0.875rem', fontWeight: 500 }}>
                  CC / BCC Recipients
                </span>
                {hasCcBccValues && (
                  <span style={{ 
                    fontSize: '0.75rem', 
                    background: 'var(--primary)', 
                    color: 'white', 
                    padding: '0.125rem 0.5rem', 
                    borderRadius: '9999px' 
                  }}>
                    {[cc, bcc].filter(v => v.trim()).length} added
                  </span>
                )}
              </div>
              
              {(showCcBcc || hasCcBccValues) && (
                <div style={{ 
                  display: 'flex', 
                  flexDirection: 'column', 
                  gap: '0.75rem',
                  padding: '0.75rem',
                  background: 'var(--gray-50)',
                  borderRadius: 'var(--border-radius)'
                }}>
                  <div>
                    <label className="form-label" style={{ fontSize: '0.8rem', marginBottom: '0.25rem' }}>
                      CC (Comma-separated emails, supports merge fields)
                    </label>
                    <input
                      ref={ccInputRef}
                      type="text"
                      className="form-input"
                      placeholder="e.g., manager@company.com, {Email}"
                      value={cc}
                      onChange={(e) => onCCChange(e.target.value)}
                      style={{ fontSize: '0.875rem' }}
                    />
                  </div>
                  <div>
                    <label className="form-label" style={{ fontSize: '0.8rem', marginBottom: '0.25rem' }}>
                      BCC (Comma-separated emails, supports merge fields)
                    </label>
                    <input
                      ref={bccInputRef}
                      type="text"
                      className="form-input"
                      placeholder="e.g., archive@company.com"
                      value={bcc}
                      onChange={(e) => onBCCChange(e.target.value)}
                      style={{ fontSize: '0.875rem' }}
                    />
                  </div>
                </div>
              )}
            </div>

            {/* Editor Type Toggle */}
            <div className="form-group">
              <label className="form-label">Body Format</label>
              <div style={{ display: 'flex', gap: '1rem' }}>
                <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer' }}>
                  <input
                    type="radio"
                    checked={!isHTML}
                    onChange={() => onIsHTMLChange(false)}
                  />
                  <Type size={16} />
                  Plain Text
                </label>
                <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer' }}>
                  <input
                    type="radio"
                    checked={isHTML}
                    onChange={() => onIsHTMLChange(true)}
                  />
                  <FileText size={16} />
                  Rich Text (HTML)
                </label>
              </div>
            </div>

            {/* Body Editor */}
            <div className="form-group">
              <label className="form-label">Message Body</label>
              {isHTML ? (
                <div className="quill-container">
                  <ReactQuill
                    ref={quillRef}
                    theme="snow"
                    value={body}
                    onChange={onBodyChange}
                    modules={quillModules}
                    placeholder="Compose your email message..."
                  />
                </div>
              ) : (
                <textarea
                  ref={bodyTextareaRef}
                  className="form-input form-textarea"
                  placeholder="Compose your email message..."
                  value={body}
                  onChange={(e) => onBodyChange(e.target.value)}
                  rows={10}
                />
              )}
            </div>
          </>
        ) : (
          /* Preview Tab */
          <div>
            <div className="alert alert-info" style={{ marginBottom: '1rem' }}>
              Preview showing sample data: John Doe (john.doe@example.com)
            </div>
            <div style={{ marginBottom: '1rem' }}>
              <strong>Subject:</strong>
              <div style={{ padding: '0.75rem', background: 'var(--gray-50)', borderRadius: 'var(--border-radius)', marginTop: '0.5rem' }}>
                {previewSubject || <span style={{ color: 'var(--gray-400)' }}>No subject</span>}
              </div>
            </div>
            {/* CC/BCC Preview - Phase 3 */}
            {(cc.trim() || bcc.trim()) && (
              <div style={{ marginBottom: '1rem', display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
                {cc.trim() && (
                  <div style={{ flex: '1 1 auto', minWidth: '200px' }}>
                    <strong>CC:</strong>
                    <div style={{ padding: '0.5rem 0.75rem', background: 'var(--gray-50)', borderRadius: 'var(--border-radius)', marginTop: '0.5rem', fontSize: '0.875rem' }}>
                      {previewCc}
                    </div>
                  </div>
                )}
                {bcc.trim() && (
                  <div style={{ flex: '1 1 auto', minWidth: '200px' }}>
                    <strong>BCC:</strong>
                    <div style={{ padding: '0.5rem 0.75rem', background: 'var(--gray-50)', borderRadius: 'var(--border-radius)', marginTop: '0.5rem', fontSize: '0.875rem' }}>
                      {previewBcc}
                    </div>
                  </div>
                )}
              </div>
            )}
            <div>
              <strong>Body:</strong>
              <div 
                style={{ 
                  padding: '1rem', 
                  background: 'var(--gray-50)', 
                  borderRadius: 'var(--border-radius)', 
                  marginTop: '0.5rem',
                  minHeight: '200px'
                }}
              >
                {isHTML ? (
                  <div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(previewBody) }} />
                ) : (
                  <pre style={{ whiteSpace: 'pre-wrap', fontFamily: 'inherit', margin: 0 }}>
                    {previewBody || <span style={{ color: 'var(--gray-400)' }}>No body content</span>}
                  </pre>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
