import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { APIResponse, Member, RSVPWithMember, SpaceSession } from '../../types';
import SessionsPage from '../SessionsPage';

// --- Mocks ---

const mockGet = vi.fn();
const mockPost = vi.fn();
const mockDelete = vi.fn();

vi.mock('../../api/client', () => ({
  api: {
    get: (...args: unknown[]) => mockGet(...args),
    post: (...args: unknown[]) => mockPost(...args),
    delete: (...args: unknown[]) => mockDelete(...args),
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

let mockUser: Partial<Member> | null = null;

vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({ user: mockUser }),
}));

const mockAddToast = vi.fn();

vi.mock('../../context/ToastContext', () => ({
  useToast: () => ({ addToast: mockAddToast }),
}));

// --- Helpers ---

function makeSession(overrides: Partial<SpaceSession> = {}): SpaceSession {
  return {
    id: 'session-1',
    title: 'Friday Coworking',
    description: null,
    date: '2026-03-06',
    start_time: '14:00',
    end_time: '18:00',
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

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  });
}

function setupApiMock(sessions: SpaceSession[], rsvps: Record<string, RSVPWithMember[]> = {}) {
  mockGet.mockImplementation((path: string) => {
    if (path === '/api/sessions') {
      return Promise.resolve({ data: sessions } as APIResponse<SpaceSession[]>);
    }
    const rsvpMatch = path.match(/^\/api\/sessions\/(.+)\/rsvps$/);
    if (rsvpMatch) {
      const sessionId = rsvpMatch[1];
      return Promise.resolve({
        data: rsvps[sessionId] ?? [],
      } as APIResponse<RSVPWithMember[]>);
    }
    return Promise.resolve({ data: [] });
  });
}

function renderPage() {
  const queryClient = createQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <SessionsPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

// --- Tests ---

describe('SessionsPage', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    vi.setSystemTime(new Date('2026-03-03T12:00:00Z'));
    mockUser = { id: 'user-1', is_admin: false };
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    mockGet.mockReset();
    mockPost.mockReset();
    mockDelete.mockReset();
    mockAddToast.mockReset();
  });

  it('shows loading spinner initially', () => {
    mockGet.mockReturnValue(new Promise(() => {})); // never resolves
    renderPage();

    const spinner = document.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
  });

  it('shows empty state when no sessions exist (non-admin)', async () => {
    setupApiMock([]);
    renderPage();

    await waitFor(() => {
      expect(screen.getByText('No upcoming sessions')).toBeInTheDocument();
    });
    expect(screen.getByText('Check back later for new sessions.')).toBeInTheDocument();
    expect(screen.queryByText('Create Session')).not.toBeInTheDocument();
  });

  it('shows empty state with Create Session button for admin', async () => {
    mockUser = { id: 'admin-1', is_admin: true };
    setupApiMock([]);
    renderPage();

    await waitFor(() => {
      expect(screen.getByText('No upcoming sessions')).toBeInTheDocument();
    });
    expect(screen.getByText('Create the first session to get started.')).toBeInTheDocument();
    const link = screen.getByText('Create Session');
    expect(link).toBeInTheDocument();
    expect(link.closest('a')).toHaveAttribute('href', '/sessions/new');
  });

  it('renders DateStrip with date chips derived from sessions', async () => {
    const sessions = [
      makeSession({ id: 's1', date: '2026-03-06', title: 'Friday Session' }),
      makeSession({ id: 's2', date: '2026-03-09', title: 'Monday Session' }),
    ];
    setupApiMock(sessions);
    renderPage();

    await waitFor(() => {
      expect(screen.getByRole('tablist')).toBeInTheDocument();
    });
    // Should have 2 date chips
    const tabs = screen.getAllByRole('tab');
    expect(tabs).toHaveLength(2);
  });

  it('defaults to nearest upcoming date', async () => {
    // Today is 2026-03-03. Sessions on 03-01 (past), 03-06 (future), 03-09 (future)
    const sessions = [
      makeSession({ id: 's-past', date: '2026-03-01', title: 'Past Session' }),
      makeSession({ id: 's-fri', date: '2026-03-06', title: 'Friday Session' }),
      makeSession({ id: 's-mon', date: '2026-03-09', title: 'Monday Session' }),
    ];
    setupApiMock(sessions);
    renderPage();

    // Should default to 2026-03-06 (nearest upcoming) and show its session
    await waitFor(() => {
      expect(screen.getByText('Friday Session')).toBeInTheDocument();
    });
    // Past session should not be visible (different date)
    expect(screen.queryByText('Past Session')).not.toBeInTheDocument();
    expect(screen.queryByText('Monday Session')).not.toBeInTheDocument();
  });

  it('falls back to last date when all dates are past', async () => {
    const sessions = [
      makeSession({ id: 's1', date: '2026-02-25', title: 'Old Session 1' }),
      makeSession({ id: 's2', date: '2026-03-01', title: 'Old Session 2' }),
    ];
    setupApiMock(sessions);
    renderPage();

    // Should default to 2026-03-01 (last date since all are past)
    await waitFor(() => {
      expect(screen.getByText('Old Session 2')).toBeInTheDocument();
    });
    expect(screen.queryByText('Old Session 1')).not.toBeInTheDocument();
  });

  it('filters sessions by selected date', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const sessions = [
      makeSession({ id: 's1', date: '2026-03-06', title: 'Friday Session' }),
      makeSession({ id: 's2', date: '2026-03-09', title: 'Monday Session' }),
    ];
    setupApiMock(sessions);
    renderPage();

    // Initially shows Friday (nearest upcoming)
    await waitFor(() => {
      expect(screen.getByText('Friday Session')).toBeInTheDocument();
    });

    // Click the second date tab (March 9)
    const tabs = screen.getAllByRole('tab');
    await user.click(tabs[1]);

    // Now should show Monday session and hide Friday session
    await waitFor(() => {
      expect(screen.getByText('Monday Session')).toBeInTheDocument();
    });
    expect(screen.queryByText('Friday Session')).not.toBeInTheDocument();
  });

  it('shows multiple sessions on the same date', async () => {
    const sessions = [
      makeSession({ id: 's1', date: '2026-03-06', title: 'Morning Session' }),
      makeSession({ id: 's2', date: '2026-03-06', title: 'Afternoon Session' }),
    ];
    setupApiMock(sessions);
    renderPage();

    await waitFor(() => {
      expect(screen.getByText('Morning Session')).toBeInTheDocument();
    });
    expect(screen.getByText('Afternoon Session')).toBeInTheDocument();
  });

  it('renders session cards in hero variant', async () => {
    const sessions = [
      makeSession({ id: 's1', date: '2026-03-06', description: 'A great coworking day' }),
    ];
    setupApiMock(sessions);
    renderPage();

    // Hero variant shows description (default variant does not)
    await waitFor(() => {
      expect(screen.getByText('A great coworking day')).toBeInTheDocument();
    });
  });

  it('shows hero image when session has image_url', async () => {
    const sessions = [
      makeSession({ id: 's1', date: '2026-03-06', image_url: '/uploads/sessions/test.jpg' }),
    ];
    setupApiMock(sessions);
    renderPage();

    await waitFor(() => {
      expect(screen.getByAltText('Friday Coworking')).toBeInTheDocument();
    });
    expect(screen.getByAltText('Friday Coworking')).toHaveAttribute(
      'src',
      '/uploads/sessions/test.jpg',
    );
  });

  it('admin sees Create Session link when sessions exist', async () => {
    mockUser = { id: 'admin-1', is_admin: true };
    setupApiMock([makeSession({ date: '2026-03-06' })]);
    renderPage();

    await waitFor(() => {
      expect(screen.getByText('Friday Coworking')).toBeInTheDocument();
    });
    const link = screen.getByText('Create Session');
    expect(link.closest('a')).toHaveAttribute('href', '/sessions/new');
  });

  it('non-admin does not see Create Session link', async () => {
    mockUser = { id: 'user-1', is_admin: false };
    setupApiMock([makeSession({ date: '2026-03-06' })]);
    renderPage();

    await waitFor(() => {
      expect(screen.getByText('Friday Coworking')).toBeInTheDocument();
    });
    expect(screen.queryByText('Create Session')).not.toBeInTheDocument();
  });

  it('aggregates session counts per date for DateStrip', async () => {
    // 2 sessions on March 6, 1 on March 9 → should produce 2 date chips
    const sessions = [
      makeSession({ id: 's1', date: '2026-03-06', title: 'Morning' }),
      makeSession({ id: 's2', date: '2026-03-06', title: 'Afternoon' }),
      makeSession({ id: 's3', date: '2026-03-09', title: 'Monday' }),
    ];
    setupApiMock(sessions);
    renderPage();

    await waitFor(() => {
      expect(screen.getByRole('tablist')).toBeInTheDocument();
    });
    const tabs = screen.getAllByRole('tab');
    expect(tabs).toHaveLength(2); // Two unique dates, not three
  });

  it('highlights the selected date in the DateStrip', async () => {
    const sessions = [
      makeSession({ id: 's1', date: '2026-03-06' }),
      makeSession({ id: 's2', date: '2026-03-09' }),
    ];
    setupApiMock(sessions);
    renderPage();

    await waitFor(() => {
      expect(screen.getByRole('tablist')).toBeInTheDocument();
    });
    const tabs = screen.getAllByRole('tab');
    // First future date (March 6) should be selected
    expect(tabs[0]).toHaveAttribute('aria-selected', 'true');
    expect(tabs[1]).toHaveAttribute('aria-selected', 'false');
  });

  it('does not fetch RSVPs for canceled sessions', async () => {
    const sessions = [
      makeSession({ id: 's-active', date: '2026-03-06', title: 'Active Session', status: 'scheduled' }),
      makeSession({ id: 's-canceled', date: '2026-03-06', title: 'Canceled Session', status: 'canceled' }),
    ];
    setupApiMock(sessions);
    renderPage();

    await waitFor(() => {
      expect(screen.getByText('Active Session')).toBeInTheDocument();
    });
    expect(screen.getByText('Canceled Session')).toBeInTheDocument();

    // api.get should have been called for sessions and for the active session's RSVPs,
    // but NOT for the canceled session's RSVPs
    const rsvpCalls = mockGet.mock.calls.filter(
      (call: unknown[]) => typeof call[0] === 'string' && (call[0] as string).includes('/rsvps'),
    );
    const canceledRsvpCall = rsvpCalls.find(
      (call: unknown[]) => (call[0] as string).includes('s-canceled'),
    );
    expect(canceledRsvpCall).toBeUndefined();
  });

  it('shows session location in hero card', async () => {
    const sessions = [
      makeSession({ id: 's1', date: '2026-03-06', location: 'WeWork Floor 3' }),
    ];
    setupApiMock(sessions);
    renderPage();

    await waitFor(() => {
      expect(screen.getByText('WeWork Floor 3')).toBeInTheDocument();
    });
  });

  it('displays month label above date strip', async () => {
    setupApiMock([makeSession({ date: '2026-03-06' })]);
    renderPage();

    await waitFor(() => {
      expect(screen.getByText('March 2026')).toBeInTheDocument();
    });
  });

  it('navigates between months and shows correct sessions', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const sessions = [
      makeSession({ id: 's-feb', date: '2026-02-25', title: 'February Session' }),
      makeSession({ id: 's-mar', date: '2026-03-06', title: 'March Session' }),
    ];
    setupApiMock(sessions);
    renderPage();

    // Default month is March (nearest upcoming date)
    await waitFor(() => {
      expect(screen.getByText('March 2026')).toBeInTheDocument();
    });
    expect(screen.getByText('March Session')).toBeInTheDocument();

    // Navigate to previous month (February)
    await user.click(screen.getByLabelText('Previous month'));

    await waitFor(() => {
      expect(screen.getByText('February 2026')).toBeInTheDocument();
    });
    expect(screen.getByText('February Session')).toBeInTheDocument();
    expect(screen.queryByText('March Session')).not.toBeInTheDocument();
  });

  it('disables month arrows at boundaries', async () => {
    setupApiMock([makeSession({ date: '2026-03-06' })]);
    renderPage();

    await waitFor(() => {
      expect(screen.getByText('March 2026')).toBeInTheDocument();
    });
    // Only one month with sessions — both arrows should be disabled
    expect(screen.getByLabelText('Previous month')).toBeDisabled();
    expect(screen.getByLabelText('Next month')).toBeDisabled();
  });
});
