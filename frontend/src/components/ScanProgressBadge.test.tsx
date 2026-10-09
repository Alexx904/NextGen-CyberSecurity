import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ScanProgressBadge } from './ScanProgressBadge';

describe('ScanProgressBadge Component', () => {
  it('renders COMPLETED badge', () => {
    render(<ScanProgressBadge status="COMPLETED" />);
    expect(screen.getByText('Completata')).toBeDefined();
  });

  it('renders basic RUNNING badge without phase/percentage', () => {
    render(<ScanProgressBadge status="RUNNING" />);
    expect(screen.getByText('In Esecuzione')).toBeDefined();
  });

  it('renders advanced RUNNING badge with phase and percentage', () => {
    render(<ScanProgressBadge status="RUNNING" phase="PORT_SCAN" percentage={35} />);
    expect(screen.getByText('PORT_SCAN (35%)')).toBeDefined();
  });

  it('renders RUNNING badge with only phase', () => {
    render(<ScanProgressBadge status="RUNNING" phase="HOST_DISCOVERY" />);
    expect(screen.getByText('HOST_DISCOVERY')).toBeDefined();
  });

  it('renders QUEUED, FAILED and CANCELLED statuses', () => {
    const { rerender } = render(<ScanProgressBadge status="QUEUED" />);
    expect(screen.getByText('In Coda')).toBeDefined();

    rerender(<ScanProgressBadge status="FAILED" />);
    expect(screen.getByText('Fallita')).toBeDefined();

    rerender(<ScanProgressBadge status="CANCELLED" />);
    expect(screen.getByText('Annullata')).toBeDefined();
  });
});

