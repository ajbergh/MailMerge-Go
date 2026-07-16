/**
 * ResultsSummary Component - Campaign Results Display
 *
 * Displays the typed CampaignResult from the backend engine: overall state
 * (completed / cancelled / preflight_failed / runtime_failed), accurate counts
 * (submitted / failed / skipped / cancelled), any fatal error or preflight
 * errors, and a per-recipient table built from attempt history. A fatal
 * failure is never shown as a zero-failure success.
 */
import { campaign, models } from '../../wailsjs/go/models';
import { CheckCircle, XCircle, Download, Users, Mail, RefreshCw, AlertTriangle, Ban } from 'lucide-react';

type CampaignResult = campaign.CampaignResult;
type RecipientResult = campaign.RecipientResult;

interface ResultsSummaryProps {
  result: CampaignResult | null;
  onExportLogs: (logs: models.EmailLog[]) => void;
  onReset: () => void;
  onRetryFailed?: () => void;
}

const stateLabel: Record<string, { text: string; className: string }> = {
  completed: { text: 'Completed', className: 'badge-success' },
  cancelled: { text: 'Cancelled', className: 'badge-warning' },
  preflight_failed: { text: 'Blocked by preflight', className: 'badge-error' },
  runtime_failed: { text: 'Failed (sender error)', className: 'badge-error' },
};

function recipientsToLogs(result: CampaignResult): models.EmailLog[] {
  return (result.recipientResults || []).map((r: RecipientResult) => {
    const last = r.attempts && r.attempts.length > 0 ? r.attempts[r.attempts.length - 1] : undefined;
    return new models.EmailLog({
      firstName: r.firstName,
      lastName: r.lastName,
      email: r.email,
      status: r.status === 'submitted' ? 'Success' : r.status === 'failed' ? 'Failure' : r.status,
      errorMessage: last?.error || '',
      timestamp: last?.timestamp,
    });
  });
}

export function ResultsSummary({ result, onExportLogs, onReset, onRetryFailed }: ResultsSummaryProps) {
  if (!result) return null;

  const badge = stateLabel[result.state] || { text: result.state, className: 'badge' };
  const logs = recipientsToLogs(result);

  const formatTimestamp = (timestamp: any) => {
    try {
      return new Date(timestamp).toLocaleString();
    } catch {
      return String(timestamp ?? '');
    }
  };

  return (
    <div className="card">
      <div className="card-header">
        <h2>
          <Mail size={20} />
          Send Results
          <span className={`badge ${badge.className}`} style={{ marginLeft: '0.75rem' }}>{badge.text}</span>
        </h2>
      </div>
      <div className="card-body">
        {/* Fatal error banner */}
        {result.fatalError && (
          <div className="alert alert-error" role="alert" style={{ marginBottom: '1rem' }}>
            <AlertTriangle size={18} />
            <span><strong>Sender error:</strong> {result.fatalError}</span>
          </div>
        )}

        {/* Preflight errors (send was blocked) */}
        {result.state === 'preflight_failed' && result.preflight && (
          <div className="alert alert-error" role="alert" style={{ marginBottom: '1rem' }}>
            <AlertTriangle size={18} />
            <div>
              <strong>The campaign was not sent. Fix these issues:</strong>
              <ul style={{ margin: '0.5rem 0 0 1rem' }}>
                {result.preflight.errors.map((e, i) => (
                  <li key={i}>{e.message}</li>
                ))}
              </ul>
            </div>
          </div>
        )}

        {/* Summary cards */}
        <div className="results-summary">
          <div className="result-card total">
            <div className="number">{result.attempted + result.skipped + result.cancelled}</div>
            <div className="label"><Users size={16} style={{ verticalAlign: 'middle', marginRight: '0.25rem' }} />Recipients</div>
          </div>
          <div className="result-card success">
            <div className="number">{result.submitted}</div>
            <div className="label"><CheckCircle size={16} style={{ verticalAlign: 'middle', marginRight: '0.25rem' }} />Submitted</div>
          </div>
          <div className="result-card failed">
            <div className="number">{result.failed}</div>
            <div className="label"><XCircle size={16} style={{ verticalAlign: 'middle', marginRight: '0.25rem' }} />Failed</div>
          </div>
          {(result.skipped > 0 || result.cancelled > 0) && (
            <div className="result-card total">
              <div className="number">{result.skipped + result.cancelled}</div>
              <div className="label"><Ban size={16} style={{ verticalAlign: 'middle', marginRight: '0.25rem' }} />Skipped/Cancelled</div>
            </div>
          )}
        </div>

        <p style={{ fontSize: '0.8125rem', color: 'var(--text-muted)', marginTop: '0.5rem' }}>
          &ldquo;Submitted&rdquo; means accepted by Outlook for sending; it does not confirm delivery.
        </p>

        {/* Per-recipient table */}
        {logs.length > 0 && (
          <div className="table-container scrollable-table" style={{ margin: '1rem 0' }}>
            <table className="table">
              <thead>
                <tr><th>Status</th><th>Name</th><th>Email</th><th>Attempts</th><th>Error</th><th>Last attempt</th></tr>
              </thead>
              <tbody>
                {result.recipientResults.map((r, index) => {
                  const last = r.attempts && r.attempts.length > 0 ? r.attempts[r.attempts.length - 1] : undefined;
                  const ok = r.status === 'submitted';
                  return (
                    <tr key={r.contactId || index}>
                      <td>
                        <span className={`badge ${ok ? 'badge-success' : r.status === 'failed' ? 'badge-error' : 'badge-warning'}`}>
                          {r.status}
                        </span>
                      </td>
                      <td>{r.firstName} {r.lastName}</td>
                      <td>{r.email}</td>
                      <td>{r.attempts ? r.attempts.length : 0}</td>
                      <td>
                        {last?.error ? (
                          <span style={{ color: 'var(--danger)', fontSize: '0.8125rem' }}>{last.error}</span>
                        ) : (
                          <span style={{ color: 'var(--gray-400)' }}>—</span>
                        )}
                      </td>
                      <td style={{ fontSize: '0.8125rem', color: 'var(--gray-500)' }}>
                        {last ? formatTimestamp(last.timestamp) : '—'}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {/* Actions */}
        <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
          <button className="btn btn-secondary" onClick={() => onExportLogs(logs)}>
            <Download size={16} />
            Export Results to CSV
          </button>
          {onRetryFailed && result.failed > 0 && (
            <button className="btn btn-warning" onClick={onRetryFailed} title={`Retry ${result.failed} failed recipients`}>
              <RefreshCw size={16} />
              Retry Failed ({result.failed})
            </button>
          )}
          <button className="btn btn-primary" onClick={onReset}>Start New Merge</button>
        </div>
      </div>
    </div>
  );
}
