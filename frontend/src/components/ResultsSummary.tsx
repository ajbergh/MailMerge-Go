/**
 * ResultsSummary Component - Send Operation Results Display
 * 
 * This component displays the final results after a bulk email send operation.
 * It provides a summary of successes and failures, a detailed log table,
 * and options to export logs or start a new merge.
 * 
 * Features:
 * - Summary cards showing total, successful, and failed counts
 * - Detailed log table with status, name, email, error, and timestamp
 * - Export to CSV functionality for record-keeping
 * - "Start New Merge" button to reset the application state
 * - Phase 1: "Retry Failed" button to re-send to failed contacts
 * 
 * The component is only rendered after a send operation completes.
 */
import { models } from '../../wailsjs/go/models';
import { CheckCircle, XCircle, Download, Users, Mail, RefreshCw } from 'lucide-react';

// Type aliases for cleaner code
type EmailLog = models.EmailLog;
type SendResult = models.SendResult;

/**
 * Props for the ResultsSummary component
 */
interface ResultsSummaryProps {
  /** Send result data from backend (null if not yet sent) */
  result: SendResult | null;
  /** Callback to export logs to CSV file */
  onExportLogs: (logs: EmailLog[]) => void;
  /** Callback to reset app state for new merge operation */
  onReset: () => void;
  /** Phase 1: Optional callback to retry sending to failed contacts */
  onRetryFailed?: () => void;
}

/**
 * ResultsSummary - Renders the send results card
 * 
 * Only renders if result is not null. Shows summary stats,
 * detailed log table, and action buttons.
 * Phase 1: Added retry failed functionality.
 */
export function ResultsSummary({ result, onExportLogs, onReset, onRetryFailed }: ResultsSummaryProps) {
  if (!result) return null;

  /**
   * Format timestamp for display in log table
   */
  const formatTimestamp = (timestamp: string) => {
    try {
      return new Date(timestamp).toLocaleString();
    } catch {
      return timestamp;
    }
  };

  return (
    <div className="card">
      <div className="card-header">
        <h2>
          <Mail size={20} />
          Send Results
        </h2>
      </div>
      <div className="card-body">
        {/* Summary Cards */}
        <div className="results-summary">
          <div className="result-card total">
            <div className="number">{result.totalSent + result.totalFailed}</div>
            <div className="label">
              <Users size={16} style={{ verticalAlign: 'middle', marginRight: '0.25rem' }} />
              Total Processed
            </div>
          </div>
          <div className="result-card success">
            <div className="number">{result.totalSent}</div>
            <div className="label">
              <CheckCircle size={16} style={{ verticalAlign: 'middle', marginRight: '0.25rem' }} />
              Successful
            </div>
          </div>
          <div className="result-card failed">
            <div className="number">{result.totalFailed}</div>
            <div className="label">
              <XCircle size={16} style={{ verticalAlign: 'middle', marginRight: '0.25rem' }} />
              Failed
            </div>
          </div>
        </div>

        {/* Logs Table */}
        {result.logs.length > 0 && (
          <div className="table-container scrollable-table" style={{ marginBottom: '1rem' }}>
            <table className="table">
              <thead>
                <tr>
                  <th>Status</th>
                  <th>Name</th>
                  <th>Email</th>
                  <th>Error</th>
                  <th>Timestamp</th>
                </tr>
              </thead>
              <tbody>
                {result.logs.map((log, index) => (
                  <tr key={index}>
                    <td>
                      <span className={`badge ${log.status === 'Success' ? 'badge-success' : 'badge-error'}`}>
                        {log.status === 'Success' ? (
                          <CheckCircle size={12} style={{ marginRight: '0.25rem' }} />
                        ) : (
                          <XCircle size={12} style={{ marginRight: '0.25rem' }} />
                        )}
                        {log.status}
                      </span>
                    </td>
                    <td>{log.firstName} {log.lastName}</td>
                    <td>{log.email}</td>
                    <td>
                      {log.errorMessage ? (
                        <span style={{ color: 'var(--danger)', fontSize: '0.8125rem' }}>
                          {log.errorMessage}
                        </span>
                      ) : (
                        <span style={{ color: 'var(--gray-400)' }}>—</span>
                      )}
                    </td>
                    <td style={{ fontSize: '0.8125rem', color: 'var(--gray-500)' }}>
                      {formatTimestamp(log.timestamp)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Actions */}
        <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
          <button 
            className="btn btn-secondary"
            onClick={() => onExportLogs(result.logs)}
          >
            <Download size={16} />
            Export Logs to CSV
          </button>
          {/* Phase 1: Retry Failed Button */}
          {onRetryFailed && result.totalFailed > 0 && (
            <button 
              className="btn btn-warning"
              onClick={onRetryFailed}
              title={`Retry sending to ${result.totalFailed} failed contacts`}
            >
              <RefreshCw size={16} />
              Retry Failed ({result.totalFailed})
            </button>
          )}
          <button 
            className="btn btn-primary"
            onClick={onReset}
          >
            Start New Merge
          </button>
        </div>
      </div>
    </div>
  );
}
