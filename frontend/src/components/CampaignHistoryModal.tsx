import { useState } from 'react';
import { History, RefreshCw, Trash2, X } from 'lucide-react';
import type { storage } from '../../wailsjs/go/models';

interface CampaignHistoryModalProps {
  isOpen: boolean;
  records: storage.CampaignRecord[];
  isLoading: boolean;
  onClose: () => void;
  onRefresh: () => void;
  onRetry: (campaignId: string) => void;
  onDelete: (campaignId: string) => void;
  onClear: () => void;
}

export function CampaignHistoryModal({
  isOpen,
  records,
  isLoading,
  onClose,
  onRefresh,
  onRetry,
  onDelete,
  onClear,
}: CampaignHistoryModalProps) {
  const [confirmClear, setConfirmClear] = useState(false);
  if (!isOpen) return null;
  return (
    <div className="modal-overlay" role="presentation" onMouseDown={(event) => {
      if (event.target === event.currentTarget) onClose();
    }}>
      <section className="modal-content history-modal" role="dialog" aria-modal="true" aria-labelledby="history-title">
        <div className="modal-header">
          <h3 id="history-title"><History size={18} /> Campaign history</h3>
          <button className="modal-close" onClick={onClose} aria-label="Close campaign history"><X size={18} /></button>
        </div>
        <div className="modal-body">
          <div className="history-actions">
            <button className="btn btn-secondary btn-sm" onClick={onRefresh} disabled={isLoading}><RefreshCw size={14} /> Refresh</button>
            {confirmClear ? (
              <span className="history-clear-confirm">
                <span>Delete all history?</span>
                <button className="btn btn-danger btn-sm" onClick={() => { setConfirmClear(false); onClear(); }}>Delete all</button>
                <button className="btn btn-secondary btn-sm" onClick={() => setConfirmClear(false)}>Cancel</button>
              </span>
            ) : (
              <button className="btn btn-danger btn-sm" onClick={() => setConfirmClear(true)} disabled={records.length === 0}><Trash2 size={14} /> Clear all</button>
            )}
          </div>
          {isLoading ? <p>Loading campaign history…</p> : records.length === 0 ? (
            <p className="empty-state">No campaign runs have been recorded.</p>
          ) : (
            <div className="history-list">
              {records.map((record) => (
                <article className="history-row" key={record.id}>
                  <div className="history-row-main">
                    <strong>{record.subject || '(No subject)'}</strong>
                    <span>{new Date(record.startedAt || record.createdAt).toLocaleString()} · Run {record.runNumber || 1}</span>
                    <span>{record.recipientCount} recipient(s) · {record.state.replaceAll('_', ' ')}</span>
                  </div>
                  <div className="history-row-actions">
                    {record.result?.failed > 0 && <button className="btn btn-secondary btn-sm" onClick={() => onRetry(record.id)}>Retry failed</button>}
                    <button className="btn btn-outline btn-sm" onClick={() => onDelete(record.id)} aria-label={`Delete campaign ${record.subject}`}><Trash2 size={14} /></button>
                  </div>
                </article>
              ))}
            </div>
          )}
        </div>
      </section>
    </div>
  );
}
