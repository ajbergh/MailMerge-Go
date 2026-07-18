import React from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';

interface ErrorBoundaryState {
  hasError: boolean;
  message: string;
}

export class ErrorBoundary extends React.Component<React.PropsWithChildren<Record<string, never>>, ErrorBoundaryState> {
  state: ErrorBoundaryState = { hasError: false, message: '' };

  static getDerivedStateFromError(error: unknown): ErrorBoundaryState {
    return {
      hasError: true,
      message: error instanceof Error ? error.message : 'An unexpected interface error occurred.',
    };
  }

  componentDidCatch(error: unknown, info: React.ErrorInfo) {
    console.error('MailMerge Go UI error', error, info.componentStack);
  }

  render() {
    if (!this.state.hasError) return this.props.children;
    return (
      <main className="fatal-error" role="alert">
        <AlertTriangle size={36} />
        <h1>MailMerge Go could not display this screen</h1>
        <p>{this.state.message}</p>
        <button className="btn btn-primary" onClick={() => window.location.reload()}>
          <RefreshCw size={16} /> Reload application
        </button>
      </main>
    );
  }
}
