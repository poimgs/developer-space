import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
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

function renderGuestList(rsvps: RSVPWithMember[]) {
  return render(
    <MemoryRouter>
      <GuestList rsvps={rsvps} />
    </MemoryRouter>,
  );
}

describe('GuestList', () => {
  it('renders empty state when no RSVPs', () => {
    renderGuestList([]);
    expect(screen.getByText('No RSVPs yet. Be the first!')).toBeInTheDocument();
  });

  it('renders member names as links to profiles', () => {
    renderGuestList([makeRSVP({ member: { id: 'm1', name: 'Alice', telegram_handle: null } })]);
    const link = screen.getByText('Alice').closest('a');
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute('href', '/profile/m1');
  });

  it('renders telegram handle without extra @', () => {
    renderGuestList([
      makeRSVP({ member: { id: 'm1', name: 'Bob', telegram_handle: '@bobdev' } }),
    ]);
    expect(screen.getByText('@bobdev')).toBeInTheDocument();
  });

  it('handles telegram handle already without @', () => {
    renderGuestList([
      makeRSVP({ member: { id: 'm1', name: 'Carol', telegram_handle: 'carol' } }),
    ]);
    expect(screen.getByText('@carol')).toBeInTheDocument();
  });

  it('does not show telegram handle when null', () => {
    const { container } = renderGuestList([
      makeRSVP({ member: { id: 'm1', name: 'Dave', telegram_handle: null } }),
    ]);
    const listItem = container.querySelector('li');
    // One anchor (name link) and no span (no telegram)
    const spans = listItem?.querySelectorAll('span') ?? [];
    expect(spans).toHaveLength(0);
    expect(screen.getByText('Dave')).toBeInTheDocument();
  });

  it('renders multiple guests with correct profile links', () => {
    const rsvps = [
      makeRSVP({ id: 'r1', member: { id: 'm1', name: 'Alice', telegram_handle: null } }),
      makeRSVP({ id: 'r2', member: { id: 'm2', name: 'Bob', telegram_handle: '@bob' } }),
      makeRSVP({ id: 'r3', member: { id: 'm3', name: 'Carol', telegram_handle: null } }),
    ];
    renderGuestList(rsvps);

    expect(screen.getByText('Alice').closest('a')).toHaveAttribute('href', '/profile/m1');
    expect(screen.getByText('Bob').closest('a')).toHaveAttribute('href', '/profile/m2');
    expect(screen.getByText('Carol').closest('a')).toHaveAttribute('href', '/profile/m3');
  });
});
