import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import type { SpaceSession } from '../../types';
import SessionCard, { formatDateLabel } from '../SessionCard';

// Mock useAuth
vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({ user: { id: 'user-1', is_admin: false } }),
}));

function makeSession(overrides: Partial<SpaceSession> = {}): SpaceSession {
  return {
    id: 'session-1',
    title: 'Friday Coworking',
    description: null,
    date: '2026-03-06',
    start_time: '14:00',
    end_time: '18:00',
    capacity: 8,
    status: 'scheduled',
    series_id: null,
    created_by: 'admin-1',
    created_at: '2026-03-01T00:00:00Z',
    updated_at: '2026-03-01T00:00:00Z',
    rsvp_count: 3,
    user_rsvped: false,
    ...overrides,
  };
}

function renderCard(session: SpaceSession, onRSVP = vi.fn(), onCancelRSVP = vi.fn()) {
  return render(
    <MemoryRouter>
      <SessionCard session={session} onRSVP={onRSVP} onCancelRSVP={onCancelRSVP} />
    </MemoryRouter>,
  );
}

describe('SessionCard', () => {
  it('renders session title and time range', () => {
    renderCard(makeSession());
    expect(screen.getByText('Friday Coworking')).toBeInTheDocument();
    expect(screen.getByText('14:00 – 18:00')).toBeInTheDocument();
  });

  it('renders spot count', () => {
    renderCard(makeSession({ rsvp_count: 3, capacity: 8 }));
    expect(screen.getByText('3/8')).toBeInTheDocument();
  });

  it('renders status badge', () => {
    renderCard(makeSession({ status: 'shifted' }));
    expect(screen.getByText('Rescheduled')).toBeInTheDocument();
  });

  it('shows RSVP button when not RSVPed and not full', () => {
    renderCard(makeSession({ user_rsvped: false, rsvp_count: 3, capacity: 8 }));
    expect(screen.getByRole('button', { name: 'RSVP' })).toBeInTheDocument();
  });

  it('calls onRSVP when RSVP button clicked', async () => {
    const user = userEvent.setup();
    const onRSVP = vi.fn();
    renderCard(makeSession({ user_rsvped: false }), onRSVP);
    await user.click(screen.getByRole('button', { name: 'RSVP' }));
    expect(onRSVP).toHaveBeenCalledWith('session-1');
  });

  it('shows Cancel RSVP button when already RSVPed', () => {
    renderCard(makeSession({ user_rsvped: true }));
    expect(screen.getByRole('button', { name: 'Cancel RSVP' })).toBeInTheDocument();
  });

  it('opens confirm modal when Cancel RSVP clicked', async () => {
    const user = userEvent.setup();
    renderCard(makeSession({ user_rsvped: true }));
    await user.click(screen.getByRole('button', { name: 'Cancel RSVP' }));
    expect(screen.getByText('Cancel your RSVP?')).toBeInTheDocument();
  });

  it('calls onCancelRSVP after confirming cancel', async () => {
    const user = userEvent.setup();
    const onCancelRSVP = vi.fn();
    renderCard(makeSession({ user_rsvped: true }), vi.fn(), onCancelRSVP);
    await user.click(screen.getByRole('button', { name: 'Cancel RSVP' }));
    // Click confirm in modal
    const buttons = screen.getAllByRole('button', { name: 'Cancel RSVP' });
    await user.click(buttons[buttons.length - 1]); // the modal confirm button
    expect(onCancelRSVP).toHaveBeenCalledWith('session-1');
  });

  it('shows Full badge when session is full and user not RSVPed', () => {
    renderCard(makeSession({ rsvp_count: 8, capacity: 8, user_rsvped: false }));
    // Both the spot count area and the disabled badge show "Full"
    const fullElements = screen.getAllByText('Full');
    expect(fullElements.length).toBeGreaterThanOrEqual(1);
  });

  it('does not show RSVP buttons for canceled sessions', () => {
    renderCard(makeSession({ status: 'canceled' }));
    expect(screen.queryByRole('button', { name: 'RSVP' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Cancel RSVP' })).toBeNull();
  });

  it('applies line-through and opacity for canceled sessions', () => {
    const { container } = renderCard(makeSession({ status: 'canceled' }));
    const card = container.firstChild as HTMLElement;
    expect(card.className).toContain('opacity-60');
    expect(screen.getByText('Friday Coworking').className).toContain('line-through');
  });

  it('shows Cancel RSVP even when full if user is RSVPed', () => {
    renderCard(makeSession({ rsvp_count: 8, capacity: 8, user_rsvped: true }));
    expect(screen.getByRole('button', { name: 'Cancel RSVP' })).toBeInTheDocument();
  });
});

describe('formatDateLabel', () => {
  it('formats a date as "Weekday, Month Day"', () => {
    const result = formatDateLabel('2026-03-06');
    expect(result).toBe('Friday, March 6');
  });

  it('formats another date correctly', () => {
    const result = formatDateLabel('2026-03-09');
    expect(result).toBe('Monday, March 9');
  });
});
