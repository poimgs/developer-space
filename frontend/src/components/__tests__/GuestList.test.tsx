import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { RSVPWithMember } from '../../types';
import GuestList from '../GuestList';

function makeRSVP(overrides: Partial<RSVPWithMember> = {}): RSVPWithMember {
  return {
    id: 'rsvp-1',
    session_id: 'session-1',
    member: { id: 'member-1', name: 'Alice', telegram_handle: null },
    created_at: '2026-03-01T10:00:00Z',
    ...overrides,
  };
}

describe('GuestList', () => {
  it('renders empty state when no RSVPs', () => {
    render(<GuestList rsvps={[]} />);
    expect(screen.getByText('No RSVPs yet. Be the first!')).toBeInTheDocument();
  });

  it('renders member names', () => {
    render(<GuestList rsvps={[makeRSVP({ member: { id: 'm1', name: 'Alice', telegram_handle: null } })]} />);
    expect(screen.getByText('Alice')).toBeInTheDocument();
  });

  it('renders telegram handle without extra @', () => {
    render(
      <GuestList
        rsvps={[makeRSVP({ member: { id: 'm1', name: 'Bob', telegram_handle: '@bobdev' } })]}
      />,
    );
    expect(screen.getByText('@bobdev')).toBeInTheDocument();
  });

  it('handles telegram handle already without @', () => {
    render(
      <GuestList
        rsvps={[makeRSVP({ member: { id: 'm1', name: 'Carol', telegram_handle: 'carol' } })]}
      />,
    );
    expect(screen.getByText('@carol')).toBeInTheDocument();
  });

  it('does not show telegram handle when null', () => {
    const { container } = render(
      <GuestList
        rsvps={[makeRSVP({ member: { id: 'm1', name: 'Dave', telegram_handle: null } })]}
      />,
    );
    // Only one span in the list item (the name)
    const listItem = container.querySelector('li');
    const spans = listItem?.querySelectorAll('span') ?? [];
    expect(spans).toHaveLength(1);
    expect(spans[0].textContent).toBe('Dave');
  });

  it('renders multiple guests', () => {
    const rsvps = [
      makeRSVP({ id: 'r1', member: { id: 'm1', name: 'Alice', telegram_handle: null } }),
      makeRSVP({ id: 'r2', member: { id: 'm2', name: 'Bob', telegram_handle: '@bob' } }),
      makeRSVP({ id: 'r3', member: { id: 'm3', name: 'Carol', telegram_handle: null } }),
    ];
    render(<GuestList rsvps={rsvps} />);
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument();
    expect(screen.getByText('Carol')).toBeInTheDocument();
  });
});
