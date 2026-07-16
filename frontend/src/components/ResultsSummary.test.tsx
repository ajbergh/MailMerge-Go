import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { ResultsSummary } from './ResultsSummary';
import { campaign } from '../../wailsjs/go/models';

function result(partial: Partial<campaign.CampaignResult>): campaign.CampaignResult {
  return new campaign.CampaignResult({
    state: 'completed',
    attempted: 0,
    submitted: 0,
    failed: 0,
    skipped: 0,
    cancelled: 0,
    recipientResults: [],
    ...partial,
  } as any);
}

describe('ResultsSummary', () => {
  it('shows preflight errors when a send was blocked (not a zero-failure success)', () => {
    const r = result({
      state: 'preflight_failed',
      preflight: new campaign.PreflightResult({
        canSend: false,
        errors: [new campaign.PreflightIssue({ severity: 'error', code: 'unknown_field', message: 'subject references unknown merge field "nope"' } as any)],
        warnings: [],
        recipientCount: 0,
      } as any),
    });
    render(<ResultsSummary result={r} onExportLogs={() => {}} onReset={() => {}} />);
    expect(screen.getByText(/Blocked by preflight/i)).toBeInTheDocument();
    expect(screen.getByText(/unknown merge field/i)).toBeInTheDocument();
  });

  it('surfaces a fatal runtime error instead of success', () => {
    const r = result({ state: 'runtime_failed', failed: 1, fatalError: 'Outlook not reachable' });
    render(<ResultsSummary result={r} onExportLogs={() => {}} onReset={() => {}} />);
    expect(screen.getByText(/Outlook not reachable/i)).toBeInTheDocument();
    expect(screen.getByText(/Failed \(sender error\)/i)).toBeInTheDocument();
  });

  it('shows a retry button only when there are failures', () => {
    const withFail = result({ state: 'completed', submitted: 2, failed: 1, attempted: 3 });
    const { rerender } = render(
      <ResultsSummary result={withFail} onExportLogs={() => {}} onReset={() => {}} onRetryFailed={() => {}} />
    );
    expect(screen.getByRole('button', { name: /Retry Failed/i })).toBeInTheDocument();

    const noFail = result({ state: 'completed', submitted: 3, failed: 0, attempted: 3 });
    rerender(<ResultsSummary result={noFail} onExportLogs={() => {}} onReset={() => {}} onRetryFailed={() => {}} />);
    expect(screen.queryByRole('button', { name: /Retry Failed/i })).not.toBeInTheDocument();
  });
});
