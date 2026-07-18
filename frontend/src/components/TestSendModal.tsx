import { useEffect, useMemo, useRef, useState } from 'react';
import DOMPurify from 'dompurify';
import { FlaskConical, Save, Send, X } from 'lucide-react';
import type { Contact } from '../types';

export interface TestSendOptions {
  testAddress: string;
  contact: Contact;
  overwriteEmail: boolean;
  draftOnly: boolean;
}

interface TestSendModalProps {
  isOpen: boolean;
  contacts: Contact[];
  subject: string;
  body: string;
  isHTML: boolean;
  isSubmitting: boolean;
  onClose: () => void;
  onPreview: (contact: Contact) => Promise<{ subject: string; body: string } | null>;
  onSubmit: (options: TestSendOptions) => Promise<void>;
}

export function TestSendModal({
  isOpen,
  contacts,
  subject,
  body,
  isHTML,
  isSubmitting,
  onClose,
  onPreview,
  onSubmit,
}: TestSendModalProps) {
  const [testAddress, setTestAddress] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [overwriteEmail, setOverwriteEmail] = useState(false);
  const [draftOnly, setDraftOnly] = useState(false);
  const [previewSubject, setPreviewSubject] = useState(subject);
  const [previewBody, setPreviewBody] = useState(body);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [validationError, setValidationError] = useState('');
  const addressRef = useRef<HTMLInputElement>(null);

  const selectedContact = useMemo(
    () => contacts[selectedIndex] || contacts[0],
    [contacts, selectedIndex],
  );

  useEffect(() => {
    if (!isOpen) return;
    setSelectedIndex(0);
    setOverwriteEmail(false);
    setDraftOnly(false);
    setValidationError('');
    window.setTimeout(() => addressRef.current?.focus(), 0);
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen || !selectedContact) {
      setPreviewSubject(subject);
      setPreviewBody(body);
      return;
    }
    let cancelled = false;
    setPreviewLoading(true);
    onPreview(selectedContact)
      .then((result) => {
        if (cancelled) return;
        setPreviewSubject(result?.subject ?? subject);
        setPreviewBody(result?.body ?? body);
      })
      .catch(() => {
        if (cancelled) return;
        setPreviewSubject(subject);
        setPreviewBody(body);
      })
      .finally(() => {
        if (!cancelled) setPreviewLoading(false);
      });
    return () => { cancelled = true; };
  }, [isOpen, selectedContact, subject, body, onPreview]);

  useEffect(() => {
    if (!isOpen) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && !isSubmitting) onClose();
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [isOpen, isSubmitting, onClose]);

  if (!isOpen) return null;

  const submit = async () => {
    const address = testAddress.trim();
    if (!/^\S+@\S+\.\S+$/.test(address)) {
      setValidationError('Enter a valid test email address.');
      addressRef.current?.focus();
      return;
    }
    if (!selectedContact) {
      setValidationError('Import or select a contact to provide merge data.');
      return;
    }
    setValidationError('');
    await onSubmit({ testAddress: address, contact: selectedContact, overwriteEmail, draftOnly });
  };

  return (
    <div className="modal-overlay" role="presentation" onMouseDown={(event) => {
      if (event.target === event.currentTarget && !isSubmitting) onClose();
    }}>
      <section className="modal-content test-send-modal" role="dialog" aria-modal="true" aria-labelledby="test-send-title">
        <div className="modal-header">
          <h3 id="test-send-title"><FlaskConical size={18} /> Test message</h3>
          <button className="modal-close" onClick={onClose} disabled={isSubmitting} aria-label="Close test message dialog">
            <X size={18} />
          </button>
        </div>
        <div className="modal-body modal-grid">
          <div className="modal-form-column">
            <label className="form-label" htmlFor="test-address">Destination</label>
            <input
              ref={addressRef}
              id="test-address"
              className="form-input"
              type="email"
              value={testAddress}
              onChange={(event) => setTestAddress(event.target.value)}
              placeholder="you@example.com"
              autoComplete="email"
            />

            <label className="form-label" htmlFor="merge-contact">Merge data contact</label>
            <select
              id="merge-contact"
              className="form-input"
              value={selectedIndex}
              onChange={(event) => setSelectedIndex(Number(event.target.value))}
            >
              {contacts.map((contact, index) => (
                <option key={contact.id || `${contact.email}-${index}`} value={index}>
                  {[contact.firstName, contact.lastName].filter(Boolean).join(' ') || contact.email} — {contact.email}
                </option>
              ))}
            </select>

            <label className="setting-checkbox">
              <input type="checkbox" checked={overwriteEmail} onChange={(event) => setOverwriteEmail(event.target.checked)} />
              Replace <code>{'{{email}}'}</code> with the test destination
            </label>
            <label className="setting-checkbox">
              <input type="checkbox" checked={draftOnly} onChange={(event) => setDraftOnly(event.target.checked)} />
              Save the test message to Outlook Drafts instead of sending
            </label>
            {validationError && <div className="inline-error" role="alert">{validationError}</div>}
          </div>

          <div className="test-preview" aria-live="polite">
            <div className="test-preview-label">Rendered preview</div>
            {previewLoading ? (
              <div className="test-preview-loading">Rendering…</div>
            ) : (
              <>
                <div className="test-preview-subject"><strong>Subject:</strong> {previewSubject}</div>
                {isHTML ? (
                  <div className="test-preview-body" dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(previewBody) }} />
                ) : (
                  <pre className="test-preview-body test-preview-plain">{previewBody}</pre>
                )}
              </>
            )}
          </div>
        </div>
        <div className="modal-footer">
          <button className="btn btn-secondary" onClick={onClose} disabled={isSubmitting}>Cancel</button>
          <button className="btn btn-success" onClick={submit} disabled={isSubmitting || contacts.length === 0}>
            {draftOnly ? <Save size={16} /> : <Send size={16} />}
            {isSubmitting ? 'Working…' : draftOnly ? 'Save test draft' : 'Send test message'}
          </button>
        </div>
      </section>
    </div>
  );
}
