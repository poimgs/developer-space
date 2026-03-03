import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import DateStrip, { type DateChip } from '../DateStrip';

function makeDates(dates: string[]): DateChip[] {
  return dates.map((date) => ({ date, sessionCount: 1 }));
}

describe('DateStrip', () => {
  it('renders a chip for each date', () => {
    const dates = makeDates(['2026-03-06', '2026-03-09', '2026-03-13']);
    render(<DateStrip dates={dates} selected="2026-03-06" onSelect={vi.fn()} />);

    const tabs = screen.getAllByRole('tab');
    expect(tabs).toHaveLength(3);
  });

  it('renders nothing when dates array is empty', () => {
    const { container } = render(
      <DateStrip dates={[]} selected="" onSelect={vi.fn()} />,
    );
    expect(container.firstChild).toBeNull();
  });

  it('displays day abbreviation and date number for each chip', () => {
    const dates = makeDates(['2026-03-06']); // Friday
    render(<DateStrip dates={dates} selected="2026-03-06" onSelect={vi.fn()} />);

    expect(screen.getByText('Fri')).toBeInTheDocument();
    expect(screen.getByText('6')).toBeInTheDocument();
  });

  it('marks the selected date with aria-selected=true', () => {
    const dates = makeDates(['2026-03-06', '2026-03-09']);
    render(<DateStrip dates={dates} selected="2026-03-09" onSelect={vi.fn()} />);

    const tabs = screen.getAllByRole('tab');
    expect(tabs[0]).toHaveAttribute('aria-selected', 'false');
    expect(tabs[1]).toHaveAttribute('aria-selected', 'true');
  });

  it('highlights selected chip with amber background', () => {
    const dates = makeDates(['2026-03-06', '2026-03-09']);
    render(<DateStrip dates={dates} selected="2026-03-06" onSelect={vi.fn()} />);

    const tabs = screen.getAllByRole('tab');
    expect(tabs[0].className).toContain('bg-amber-500');
    expect(tabs[1].className).not.toContain('bg-amber-500');
  });

  it('calls onSelect when a chip is clicked', async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    const dates = makeDates(['2026-03-06', '2026-03-09']);
    render(<DateStrip dates={dates} selected="2026-03-06" onSelect={onSelect} />);

    const tabs = screen.getAllByRole('tab');
    await user.click(tabs[1]);
    expect(onSelect).toHaveBeenCalledWith('2026-03-09');
  });

  it('unselected chips have stone background', () => {
    const dates = makeDates(['2026-03-06', '2026-03-09']);
    render(<DateStrip dates={dates} selected="2026-03-06" onSelect={vi.fn()} />);

    const tabs = screen.getAllByRole('tab');
    expect(tabs[1].className).toContain('bg-stone-100');
  });

  it('has tablist role for accessibility', () => {
    const dates = makeDates(['2026-03-06']);
    render(<DateStrip dates={dates} selected="2026-03-06" onSelect={vi.fn()} />);
    expect(screen.getByRole('tablist')).toBeInTheDocument();
  });

  it('renders multiple dates in order', () => {
    const dates = makeDates(['2026-03-02', '2026-03-06', '2026-03-09']); // Mon, Fri, Mon
    render(<DateStrip dates={dates} selected="2026-03-02" onSelect={vi.fn()} />);

    const tabs = screen.getAllByRole('tab');
    expect(tabs[0]).toHaveTextContent('2');
    expect(tabs[1]).toHaveTextContent('6');
    expect(tabs[2]).toHaveTextContent('9');
  });
});
