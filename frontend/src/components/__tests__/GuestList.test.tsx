import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import type { RSVPWithMember } from '../../types';
import GuestList from '../GuestList';

vi.mock('../../api/client', () => ({
  api: {
    getPublicProfile: vi.fn().mockResolvedValue({ data: null }),
  },
  ApiError: class ApiError extends Error {
    status: number;
    body: { error: string };
    constructor(status: number, body: { error: string }) {
      super(body.error);
      this.status = status;
      this.body = body;
    }
  },
}));

function makeRSVP(overrides: Partial<RSVPWithMember> = {}): RSVPWithMember {
  return {
    id: 'rsvp-1',
    session_id: 'session-1',
    member: { id: 'member-1', name: 'Alice', telegram_handle: null, bio: null },
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

  it('renders member names as clickable buttons', () => {
    renderGuestList([makeRSVP({ member: { id: 'm1', name: 'Alice', telegram_handle: null, bio: null } })]);
    const button = screen.getByRole('button', { name: 'Alice' });
    expect(button).toBeInTheDocument();
  });

  it('renders bio when present', () => {
    renderGuestList([
      makeRSVP({ member: { id: 'm1', name: 'Bob', telegram_handle: null, bio: 'Full-stack developer' } }),
    ]);
    expect(screen.getByText('Full-stack developer')).toBeInTheDocument();
  });

  it('does not show bio when null', () => {
    const { container } = renderGuestList([
      makeRSVP({ member: { id: 'm1', name: 'Dave', telegram_handle: null, bio: null } }),
    ]);
    const listItem = container.querySelector('li');
    const spans = listItem?.querySelectorAll('span') ?? [];
    expect(spans).toHaveLength(0);
    expect(screen.getByText('Dave')).toBeInTheDocument();
  });

  it('renders multiple guests as buttons', () => {
    const rsvps = [
      makeRSVP({ id: 'r1', member: { id: 'm1', name: 'Alice', telegram_handle: null, bio: null } }),
      makeRSVP({ id: 'r2', member: { id: 'm2', name: 'Bob', telegram_handle: null, bio: 'Designer' } }),
      makeRSVP({ id: 'r3', member: { id: 'm3', name: 'Carol', telegram_handle: null, bio: null } }),
    ];
    renderGuestList(rsvps);

    expect(screen.getByRole('button', { name: 'Alice' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Bob' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Carol' })).toBeInTheDocument();
  });
});
