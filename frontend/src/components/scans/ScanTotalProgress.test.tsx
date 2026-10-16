import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import { remainingTimeLabel, ScanTotalProgress } from './ScanTotalProgress';

describe('ETA presentation', () => {
  it('avoids false completion or countdown after the estimate expires', () => {
    expect(remainingTimeLabel({ percentage: 40, remainingSeconds: 10, receivedAt: 1000 }, 12000)).toContain('aggiornamento');
    expect(remainingTimeLabel({ percentage: 40, remainingSeconds: 100, receivedAt: 1000 }, 32000)).toContain('in calcolo');
    expect(remainingTimeLabel({ percentage: 40, remainingSeconds: 120, receivedAt: 1000 }, 1000)).toContain('indicativo');
  });
  it('keeps queued progress unknown and shows 100 only for completed scans', () => {
    const view = render(<ScanTotalProgress status="QUEUED" estimate={null} />);
    expect(screen.getByRole('progressbar').getAttribute('aria-valuenow')).toBeNull();
    view.rerender(<ScanTotalProgress status="COMPLETED" estimate={null} />);
    expect(screen.getByRole('progressbar').getAttribute('aria-valuenow')).toBe('100');
  });
});
