/**
 * ProgressTracker Component - Real-time Send Progress Display
 * 
 * This component displays real-time progress during bulk email sending.
 * It shows a progress bar, current/total count, and a scrollable log
 * of individual email send statuses.
 * 
 * Features:
 * - Animated progress bar with percentage
 * - Real-time status log with timestamps
 * - Success/failure icons for each email
 * - Completion state when all emails processed
 * 
 * The component subscribes to "email:progress" events from the backend
 * to receive live updates during the send operation.
 */
import { ProgressUpdate } from '../types';
import { Clock, CheckCircle, XCircle, Loader } from 'lucide-react';

/**
 * Props for the ProgressTracker component
 */
interface ProgressTrackerProps {
  /** Current progress update from backend (null if not started) */
  progress: ProgressUpdate | null;
  /** Array of status log entries for display */
  logs: Array<{ email: string; status: string; message?: string; timestamp: string }>;
  /** Whether the send operation is currently running */
  isRunning: boolean;
}

/**
 * ProgressTracker - Renders progress bar and status log during sending
 * 
 * Only renders when either isRunning is true or there are log entries.
 * Shows different header based on whether operation is in progress or complete.
 */
export function ProgressTracker({ progress, logs, isRunning }: ProgressTrackerProps) {
  /**
   * Calculate progress percentage (0-100)
   */
  const getProgressPercentage = () => {
    if (!progress || progress.total === 0) return 0;
    return Math.round((progress.current / progress.total) * 100);
  };

  /**
   * Get the appropriate status icon based on email status
   */
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle size={14} style={{ color: 'var(--success)' }} />;
      case 'failure':
        return <XCircle size={14} style={{ color: 'var(--danger)' }} />;
      case 'sending':
        return <Loader size={14} style={{ color: 'var(--primary)' }} className="spinner" />;
      default:
        return <Clock size={14} style={{ color: 'var(--gray-400)' }} />;
    }
  };

  const formatTime = (timestamp: string) => {
    try {
      return new Date(timestamp).toLocaleTimeString();
    } catch {
      return timestamp;
    }
  };

  if (!isRunning && logs.length === 0) {
    return null;
  }

  return (
    <div className="card">
      <div className="card-header">
        <h2>
          {isRunning ? (
            <>
              <div className="spinner" style={{ marginRight: '0.5rem' }}></div>
              Sending Progress
            </>
          ) : (
            <>
              <CheckCircle size={20} style={{ color: 'var(--success)' }} />
              Send Complete
            </>
          )}
        </h2>
      </div>
      <div className="card-body">
        {/* Progress Bar */}
        {progress && (
          <div className="progress-container">
            <div className="progress-header">
              <span>
                {isRunning ? 'Sending emails...' : 'Complete'}
              </span>
              <span>
                {progress.current} / {progress.total} ({getProgressPercentage()}%)
              </span>
            </div>
            <div className="progress-bar">
              <div 
                className={`progress-bar-fill ${!isRunning ? 'success' : ''}`}
                style={{ width: `${getProgressPercentage()}%` }}
              />
            </div>
          </div>
        )}

        {/* Status Log */}
        {logs.length > 0 && (
          <div className="status-log">
            {logs.map((log, index) => (
              <div 
                key={index} 
                className={`status-log-entry ${log.status}`}
              >
                <span className="timestamp">{formatTime(log.timestamp)}</span>
                {getStatusIcon(log.status)}
                <span className="message">
                  {log.email}
                  {log.message && (
                    <span style={{ color: 'var(--gray-500)', marginLeft: '0.5rem' }}>
                      — {log.message}
                    </span>
                  )}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
