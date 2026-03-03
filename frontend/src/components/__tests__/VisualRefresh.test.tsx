import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import type { RSVPWithMember, SpaceSession } from '../../types';
import ConfirmModal from '../ConfirmModal';
import SessionCard from '../SessionCard';
import Toast from '../Toast';

// Mock useAuth for SessionCard
vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({ user: { id: 'user-1', is_admin: true } }),
}));

// Mock useToast to return a controlled list of toasts
const mockRemoveToast = vi.fn();
vi.mock('../../context/ToastContext', () => ({
  useToast: () => ({
    toasts: [
      { id: '1', message: 'Info message', type: 'info' },
      { id: '2', message: 'Success message', type: 'success' },
    ],
    removeToast: mockRemoveToast,
  }),
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
    image_url: null,
    location: null,
    series_id: null,
    created_by: 'admin-1',
    created_at: '2026-03-01T00:00:00Z',
    updated_at: '2026-03-01T00:00:00Z',
    rsvp_count: 3,
    user_rsvped: false,
    ...overrides,
  };
}

/** Collect all class names from an element and its descendants */
function getAllClasses(container: HTMLElement): string {
  const classes: string[] = [];
  container.querySelectorAll('*').forEach((el) => {
    if (el.className && typeof el.className === 'string') {
      classes.push(el.className);
    }
  });
  return classes.join(' ');
}

describe('Visual Refresh — Warm Palette', () => {
  describe('SessionCard uses amber accent and stone neutrals', () => {
    it('uses amber accent for RSVP button', () => {
      render(
        <MemoryRouter>
          <SessionCard
            session={makeSession()}
            onRSVP={vi.fn()}
            onCancelRSVP={vi.fn()}
          />
        </MemoryRouter>,
      );
      const rsvpButton = screen.getByRole('button', { name: 'RSVP' });
      expect(rsvpButton.className).toContain('bg-amber-600');
      expect(rsvpButton.className).toContain('hover:bg-amber-700');
      expect(rsvpButton.className).not.toMatch(/indigo/);
    });

    it('uses stone neutrals for card container', () => {
      const { container } = render(
        <MemoryRouter>
          <SessionCard
            session={makeSession()}
            onRSVP={vi.fn()}
            onCancelRSVP={vi.fn()}
          />
        </MemoryRouter>,
      );
      const allClasses = getAllClasses(container);
      expect(allClasses).toContain('border-stone-200');
      expect(allClasses).toContain('dark:bg-stone-800');
      expect(allClasses).not.toMatch(/\bgray-\d/);
    });

    it('uses rounded-xl with shadow-sm on card', () => {
      const { container } = render(
        <MemoryRouter>
          <SessionCard
            session={makeSession()}
            onRSVP={vi.fn()}
            onCancelRSVP={vi.fn()}
          />
        </MemoryRouter>,
      );
      // The card is the first real div child
      const card = container.querySelector('[class*="rounded-xl"]');
      expect(card).not.toBeNull();
      expect(card!.className).toContain('shadow-sm');
    });

    it('uses amber for Cancel RSVP button', () => {
      render(
        <MemoryRouter>
          <SessionCard
            session={makeSession({ user_rsvped: true })}
            onRSVP={vi.fn()}
            onCancelRSVP={vi.fn()}
          />
        </MemoryRouter>,
      );
      const cancelBtn = screen.getByRole('button', { name: 'Cancel RSVP' });
      expect(cancelBtn.className).toContain('border-amber-600');
      expect(cancelBtn.className).toContain('text-amber-600');
      expect(cancelBtn.className).not.toMatch(/indigo/);
    });

    it('uses amber for attendee pills', () => {
      const attendees: RSVPWithMember[] = [
        { id: 'r1', session_id: 's1', member: { id: 'm1', name: 'Alice', telegram_handle: null }, created_at: '' },
      ];
      render(
        <MemoryRouter>
          <SessionCard
            session={makeSession()}
            attendees={attendees}
            onRSVP={vi.fn()}
            onCancelRSVP={vi.fn()}
          />
        </MemoryRouter>,
      );
      const pill = screen.getByText('Alice');
      expect(pill.className).toContain('bg-amber-50');
      expect(pill.className).toContain('text-amber-700');
      expect(pill.className).not.toMatch(/indigo/);
    });

    it('uses amber for capacity fill bar', () => {
      const { container } = render(
        <MemoryRouter>
          <SessionCard
            session={makeSession({ rsvp_count: 3, capacity: 8 })}
            onRSVP={vi.fn()}
            onCancelRSVP={vi.fn()}
          />
        </MemoryRouter>,
      );
      const fillBar = container.querySelector('[class*="bg-amber-500"]');
      expect(fillBar).not.toBeNull();
    });

    it('has no indigo classes anywhere in the card', () => {
      const attendees: RSVPWithMember[] = [
        { id: 'r1', session_id: 's1', member: { id: 'm1', name: 'Alice', telegram_handle: null }, created_at: '' },
      ];
      const { container } = render(
        <MemoryRouter>
          <SessionCard
            session={makeSession({ user_rsvped: true, series_id: 'series-1' })}
            attendees={attendees}
            onRSVP={vi.fn()}
            onCancelRSVP={vi.fn()}
          />
        </MemoryRouter>,
      );
      const allClasses = getAllClasses(container);
      expect(allClasses).not.toMatch(/indigo/);
      expect(allClasses).not.toMatch(/\bgray-\d/);
    });
  });

  describe('Toast uses amber for info type', () => {
    it('renders info toast with amber border', () => {
      render(<Toast />);
      const infoToast = screen.getByText('Info message').closest('div');
      expect(infoToast?.className).toContain('border-amber-500');
      expect(infoToast?.className).toContain('bg-amber-50');
      expect(infoToast?.className).not.toMatch(/indigo/);
    });
  });

  describe('ConfirmModal uses stone neutrals', () => {
    it('uses stone colors and rounded-xl', () => {
      const { container } = render(
        <ConfirmModal
          open={true}
          title="Confirm action"
          message="Are you sure?"
          onConfirm={vi.fn()}
          onCancel={vi.fn()}
        />,
      );
      const allClasses = getAllClasses(container);
      expect(allClasses).toContain('dark:bg-stone-800');
      expect(allClasses).toContain('text-stone-900');
      expect(allClasses).toContain('rounded-xl');
      expect(allClasses).not.toMatch(/indigo/);
      expect(allClasses).not.toMatch(/\bgray-\d/);
    });

    it('uses stone for cancel button border', () => {
      render(
        <ConfirmModal
          open={true}
          title="Confirm action"
          message="Are you sure?"
          onConfirm={vi.fn()}
          onCancel={vi.fn()}
        />,
      );
      const cancelBtn = screen.getByRole('button', { name: 'Cancel' });
      expect(cancelBtn.className).toContain('border-stone-300');
      expect(cancelBtn.className).toContain('text-stone-700');
    });
  });
});
