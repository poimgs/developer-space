import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import StatusBadge from '../StatusBadge';

describe('StatusBadge', () => {
  it('renders "Scheduled" for scheduled status', () => {
    render(<StatusBadge status="scheduled" />);
    expect(screen.getByText('Scheduled')).toBeInTheDocument();
  });

  it('renders "Rescheduled" for shifted status', () => {
    render(<StatusBadge status="shifted" />);
    expect(screen.getByText('Rescheduled')).toBeInTheDocument();
  });

  it('renders "Canceled" for canceled status', () => {
    render(<StatusBadge status="canceled" />);
    expect(screen.getByText('Canceled')).toBeInTheDocument();
  });

  it('renders unknown status as-is', () => {
    render(<StatusBadge status="unknown" />);
    expect(screen.getByText('unknown')).toBeInTheDocument();
  });

  it('applies green styling for scheduled', () => {
    render(<StatusBadge status="scheduled" />);
    const badge = screen.getByText('Scheduled');
    expect(badge.className).toContain('text-status-scheduled');
  });

  it('applies amber styling for shifted', () => {
    render(<StatusBadge status="shifted" />);
    const badge = screen.getByText('Rescheduled');
    expect(badge.className).toContain('text-status-shifted');
  });

  it('applies red styling for canceled', () => {
    render(<StatusBadge status="canceled" />);
    const badge = screen.getByText('Canceled');
    expect(badge.className).toContain('text-status-canceled');
  });
});
