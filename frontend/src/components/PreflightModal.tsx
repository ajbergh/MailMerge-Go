import { AlertCircle, AlertTriangle, Clock, Paperclip, Save, Send, Users, X } from 'lucide-react';
import type { campaign } from '../../wailsjs/go/models';

interface PreflightModalProps {
  isOpen: boolean;
  result: campaign.PreflightResult | null;
  draftOnly: boolean;
  totalAttachmentBytes: number;
  isSubmitting: boolean;
  onClose: () => void;
  onConfirm: () => void;
}

function formatBytes(bytes: number): string {
  if (bytes <= 0) return 'None';
  const units = ['B', 'KB', 'MB', 'GB'];
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  return `${(bytes / 1024 ** index).toFixed(index === 0 ? 0 : 1)} ${units[index]}`;
}

function formatDuration(nanoseconds: number): string {
  const seconds = Math.max(0, Math.round((nanoseconds || 0) / 1_000_000_000));
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  return `${minutes}m ${seconds % 60}s`;
}

export function PreflightModal({
  isOpen,
  result,
  draftOnly,
  totalAttachmentBytes,
  isSubmitting,
  onClose,
  onConfirm,
}: PreflightModalProps) {
  if (!isOpen || !result) return null;

  return (
    <div className="modal-overlay" role="presentation" onMouseDown={(event) => {
      if (event.target === event.currentTarget && !isSubmitting) onClose();
    }}>
      <section className="modal-content preflight-modal" role="dialog" aria-modal="true" aria-labelledby="preflight-title">
        <div className="modal-header">
          <h3 id="preflight-title"><Send size={18} /> Review campaign</h3>
          <button className="modal-close" onClick={onClose} disabled={isSubmitting} aria-label="Close campaign review">
            <X size={18} />
          </button>
        </div>
        <div className="modal-body">
          <div className="preflight-stats">
            <div><Users size={18} /><strong>{result.recipientCount}</strong><span>Recipients</span></div>
            <div><Clock size={18} /><strong>{formatDuration(result.estimatedDuration)}</strong><span>Estimated</span></div>
            <div><Paperclip size={18} /><strong>{formatBytes(totalAttachmentBytes)}</strong><span>Attachments</span></div>
          </div>

          {result.errors.length > 0 && (
            <section className="preflight-section preflight-errors" aria-labelledby="preflight-errors-title">
              <h4 id="preflight-errors-title"><AlertCircle size={16} /> Blocking issues</h4>
              <ul>{result.errors.map((issue, index) => <li key={`${issue.code}-${index}`}>{issue.message}</li>)}</ul>
            </section>
          )}
          {result.warnings.length > 0 && (
            <section className="preflight-section preflight-warnings" aria-labelledby="preflight-warnings-title">
              <h4 id="preflight-warnings-title"><AlertTriangle size={16} /> Warnings</h4>
              <ul>{result.warnings.map((issue, index) => <li key={`${issue.code}-${index}`}>{issue.message}</li>)}</ul>
            </section>
          )}
          {result.errors.length === 0 && result.warnings.length === 0 && (
            <div className="preflight-ready">Preflight passed with no warnings.</div>
          )}

          <div className="preflight-mode">
            {draftOnly ? <Save size={17} /> : <Send size={17} />}
            <div>
              <strong>{draftOnly ? 'Draft-only mode' : 'Send mode'}</strong>
              <span>{draftOnly ? 'Messages will be saved in Outlook Drafts and will not be sent.' : 'Outlook will submit messages immediately after confirmation.'}</span>
            </div>
          </div>
        </div>
        <div className="modal-footer">
          <button className="btn btn-secondary" onClick={onClose} disabled={isSubmitting}>Back</button>
          <button className="btn btn-success" onClick={onConfirm} disabled={!result.canSend || isSubmitting}>
            {draftOnly ? <Save size={16} /> : <Send size={16} />}
            {isSubmitting ? 'Starting…' : draftOnly ? `Create ${result.recipientCount} draft(s)` : `Send ${result.recipientCount} message(s)`}
          </button>
        </div>
      </section>
    </div>
  );
}
