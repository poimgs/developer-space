import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { SpaceSession } from '../../types';
import SessionForm from '../SessionForm';

function makeSession(overrides: Partial<SpaceSession> = {}): SpaceSession {
  return {
    id: 'session-1',
    title: 'Friday Coworking',
    description: 'A fun session',
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

describe('SessionForm', () => {
  describe('create mode', () => {
    it('renders all form fields with empty values', () => {
      render(<SessionForm onSubmit={vi.fn()} />);
      expect(screen.getByLabelText(/title/i)).toHaveValue('');
      expect(screen.getByLabelText(/description/i)).toHaveValue('');
      expect(screen.getByLabelText(/date/i)).toHaveValue('');
      expect(screen.getByLabelText(/start time/i)).toHaveValue('');
      expect(screen.getByLabelText(/end time/i)).toHaveValue('');
      expect(screen.getByLabelText(/capacity/i)).toHaveValue(null);
    });

    it('shows Create Session submit button', () => {
      render(<SessionForm onSubmit={vi.fn()} />);
      expect(screen.getByRole('button', { name: 'Create Session' })).toBeInTheDocument();
    });

    it('shows repeat options', () => {
      render(<SessionForm onSubmit={vi.fn()} />);
      expect(screen.getByLabelText(/no repeat/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/repeat weekly for/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/repeat forever/i)).toBeInTheDocument();
    });

    it('shows week count input when repeat weekly selected', async () => {
      const user = userEvent.setup();
      render(<SessionForm onSubmit={vi.fn()} />);
      await user.click(screen.getByLabelText(/repeat weekly for/i));
      expect(screen.getByDisplayValue('1')).toBeInTheDocument();
      expect(screen.getByText('weeks')).toBeInTheDocument();
    });

    it('validates required fields on submit', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn();
      render(<SessionForm onSubmit={onSubmit} />);
      await user.click(screen.getByRole('button', { name: 'Create Session' }));
      expect(screen.getByText('Title is required')).toBeInTheDocument();
      expect(screen.getByText('Date is required')).toBeInTheDocument();
      expect(screen.getByText('Start time is required')).toBeInTheDocument();
      expect(screen.getByText('End time is required')).toBeInTheDocument();
      expect(screen.getByText('Capacity must be at least 1')).toBeInTheDocument();
      expect(onSubmit).not.toHaveBeenCalled();
    });

    it('validates end time after start time', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn();
      render(<SessionForm onSubmit={onSubmit} />);

      await user.type(screen.getByLabelText(/title/i), 'Test');
      await user.type(screen.getByLabelText(/date/i), '2026-04-01');
      // Type time values using the input
      const startInput = screen.getByLabelText(/start time/i);
      const endInput = screen.getByLabelText(/end time/i);
      await user.clear(startInput);
      await user.type(startInput, '18:00');
      await user.clear(endInput);
      await user.type(endInput, '14:00');
      await user.clear(screen.getByLabelText(/capacity/i));
      await user.type(screen.getByLabelText(/capacity/i), '8');

      await user.click(screen.getByRole('button', { name: 'Create Session' }));
      expect(screen.getByText('End time must be after start time')).toBeInTheDocument();
      expect(onSubmit).not.toHaveBeenCalled();
    });

    it('submits valid create form data', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      render(<SessionForm onSubmit={onSubmit} />);

      await user.type(screen.getByLabelText(/title/i), 'Friday Coworking');
      await user.type(screen.getByLabelText(/description/i), 'A fun session');
      await user.type(screen.getByLabelText(/date/i), '2026-04-01');
      await user.type(screen.getByLabelText(/start time/i), '14:00');
      await user.type(screen.getByLabelText(/end time/i), '18:00');
      await user.type(screen.getByLabelText(/capacity/i), '8');

      await user.click(screen.getByRole('button', { name: 'Create Session' }));

      expect(onSubmit).toHaveBeenCalledWith({
        title: 'Friday Coworking',
        description: 'A fun session',
        date: '2026-04-01',
        start_time: '14:00',
        end_time: '18:00',
        capacity: 8,
      });
    });

    it('submits with repeat_weekly when selected', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      render(<SessionForm onSubmit={onSubmit} />);

      await user.type(screen.getByLabelText(/title/i), 'Weekly');
      await user.type(screen.getByLabelText(/date/i), '2026-04-01');
      await user.type(screen.getByLabelText(/start time/i), '14:00');
      await user.type(screen.getByLabelText(/end time/i), '18:00');
      await user.type(screen.getByLabelText(/capacity/i), '8');

      await user.click(screen.getByLabelText(/repeat weekly for/i));
      // Default repeat count is 1
      await user.click(screen.getByRole('button', { name: 'Create Session' }));

      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({ repeat_weekly: 1 }),
      );
    });

    it('shows Saving... when loading', () => {
      render(<SessionForm onSubmit={vi.fn()} loading />);
      expect(screen.getByRole('button', { name: 'Saving...' })).toBeDisabled();
    });
  });

  describe('edit mode', () => {
    it('pre-populates form with session data', () => {
      const session = makeSession();
      render(<SessionForm session={session} onSubmit={vi.fn()} />);
      expect(screen.getByLabelText(/title/i)).toHaveValue('Friday Coworking');
      expect(screen.getByLabelText(/description/i)).toHaveValue('A fun session');
      expect(screen.getByLabelText(/date/i)).toHaveValue('2026-03-06');
      expect(screen.getByLabelText(/start time/i)).toHaveValue('14:00');
      expect(screen.getByLabelText(/end time/i)).toHaveValue('18:00');
      expect(screen.getByLabelText(/capacity/i)).toHaveValue(8);
    });

    it('shows Save Changes submit button', () => {
      render(<SessionForm session={makeSession()} onSubmit={vi.fn()} />);
      expect(screen.getByRole('button', { name: 'Save Changes' })).toBeInTheDocument();
    });

    it('does not show repeat options in edit mode', () => {
      render(<SessionForm session={makeSession()} onSubmit={vi.fn()} />);
      expect(screen.queryByLabelText(/no repeat/i)).toBeNull();
      expect(screen.queryByLabelText(/repeat forever/i)).toBeNull();
    });

    it('sends only changed fields on edit', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      const session = makeSession();
      render(<SessionForm session={session} onSubmit={onSubmit} />);

      const titleInput = screen.getByLabelText(/title/i);
      await user.clear(titleInput);
      await user.type(titleInput, 'New Title');

      await user.click(screen.getByRole('button', { name: 'Save Changes' }));

      expect(onSubmit).toHaveBeenCalledWith({ title: 'New Title' });
    });

    it('sends empty object when nothing changed', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      const session = makeSession();
      render(<SessionForm session={session} onSubmit={onSubmit} />);

      await user.click(screen.getByRole('button', { name: 'Save Changes' }));

      expect(onSubmit).toHaveBeenCalledWith({});
    });

    it('handles null description correctly in edit', () => {
      const session = makeSession({ description: null });
      render(<SessionForm session={session} onSubmit={vi.fn()} />);
      expect(screen.getByLabelText(/description/i)).toHaveValue('');
    });
  });

  describe('location field', () => {
    it('renders location field in create mode', () => {
      render(<SessionForm onSubmit={vi.fn()} />);
      expect(screen.getByLabelText(/location/i)).toBeInTheDocument();
    });

    it('includes location in create form data', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      render(<SessionForm onSubmit={onSubmit} />);

      await user.type(screen.getByLabelText(/title/i), 'Test');
      await user.type(screen.getByLabelText(/date/i), '2026-04-01');
      await user.type(screen.getByLabelText(/start time/i), '14:00');
      await user.type(screen.getByLabelText(/end time/i), '18:00');
      await user.type(screen.getByLabelText(/capacity/i), '8');
      await user.type(screen.getByLabelText(/location/i), 'Room 42');

      await user.click(screen.getByRole('button', { name: 'Create Session' }));

      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({ location: 'Room 42' }),
      );
    });

    it('pre-populates location in edit mode', () => {
      const session = makeSession({ location: 'Room 42' });
      render(<SessionForm session={session} onSubmit={vi.fn()} />);
      expect(screen.getByLabelText(/location/i)).toHaveValue('Room 42');
    });

    it('sends location change on edit', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      const session = makeSession({ location: 'Room 42' });
      render(<SessionForm session={session} onSubmit={onSubmit} />);

      const locationInput = screen.getByLabelText(/location/i);
      await user.clear(locationInput);
      await user.type(locationInput, 'Room 99');

      await user.click(screen.getByRole('button', { name: 'Save Changes' }));

      expect(onSubmit).toHaveBeenCalledWith({ location: 'Room 99' });
    });
  });

  describe('required field indicators', () => {
    it('shows asterisks on required fields', () => {
      render(<SessionForm onSubmit={vi.fn()} />);
      // There should be red asterisks for Title, Date, Start time, End time, Capacity
      const asterisks = document.querySelectorAll('.text-red-500');
      expect(asterisks.length).toBe(5);
    });
  });
});
