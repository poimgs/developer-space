import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Member } from '../../types';
import ProfilePage from '../ProfilePage';

const mockAddToast = vi.fn();
const mockRefresh = vi.fn();
const mockPatch = vi.fn();

vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({
    user: mockUser,
    refresh: mockRefresh,
  }),
}));

vi.mock('../../context/ToastContext', () => ({
  useToast: () => ({
    addToast: mockAddToast,
  }),
}));

vi.mock('../../api/client', () => ({
  api: {
    patch: (...args: unknown[]) => mockPatch(...args),
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

let mockUser: Member | null = null;

function makeMember(overrides: Partial<Member> = {}): Member {
  return {
    id: 'member-1',
    email: 'alice@example.com',
    name: 'Alice',
    telegram_handle: '@alice',
    is_admin: false,
    is_active: true,
    bio: 'Full-stack developer',
    skills: ['react', 'go'],
    linkedin_url: 'https://linkedin.com/in/alice',
    instagram_handle: 'alice_gram',
    github_username: 'alice-dev',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('ProfilePage', () => {
  beforeEach(() => {
    mockUser = makeMember();
    mockPatch.mockResolvedValue({});
    mockRefresh.mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    mockAddToast.mockReset();
    mockPatch.mockReset();
    mockRefresh.mockReset();
  });

  it('renders all three sections', () => {
    render(<ProfilePage />);
    expect(screen.getByText('Identity')).toBeInTheDocument();
    expect(screen.getByText('About')).toBeInTheDocument();
    expect(screen.getByText('Social Links')).toBeInTheDocument();
  });

  it('displays email and admin status', () => {
    mockUser = makeMember({ is_admin: true });
    render(<ProfilePage />);
    expect(screen.getByText(/alice@example.com/)).toBeInTheDocument();
    expect(screen.getByText(/Admin/)).toBeInTheDocument();
  });

  it('populates form with user data', () => {
    render(<ProfilePage />);
    expect(screen.getByLabelText(/Name/)).toHaveValue('Alice');
    expect(screen.getByLabelText(/Telegram Handle/)).toHaveValue('@alice');
    expect(screen.getByLabelText(/Bio/)).toHaveValue('Full-stack developer');
    expect(screen.getByText('react')).toBeInTheDocument();
    expect(screen.getByText('go')).toBeInTheDocument();
    expect(screen.getByLabelText(/LinkedIn URL/)).toHaveValue('https://linkedin.com/in/alice');
    expect(screen.getByLabelText(/Instagram Handle/)).toHaveValue('alice_gram');
    expect(screen.getByLabelText(/GitHub Username/)).toHaveValue('alice-dev');
  });

  it('shows bio character counter', () => {
    mockUser = makeMember({ bio: 'Hello' });
    render(<ProfilePage />);
    expect(screen.getByText('495 characters remaining')).toBeInTheDocument();
  });

  it('shows red counter when bio nears limit', () => {
    mockUser = makeMember({ bio: 'x'.repeat(460) });
    render(<ProfilePage />);
    const counter = screen.getByText('40 characters remaining');
    expect(counter.className).toContain('text-red-600');
  });

  it('submits all fields on save', async () => {
    const user = userEvent.setup();
    render(<ProfilePage />);

    await user.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => {
      expect(mockPatch).toHaveBeenCalledWith('/api/auth/profile', {
        name: 'Alice',
        telegram_handle: '@alice',
        bio: 'Full-stack developer',
        skills: ['react', 'go'],
        linkedin_url: 'https://linkedin.com/in/alice',
        instagram_handle: 'alice_gram',
        github_username: 'alice-dev',
      });
    });
    expect(mockRefresh).toHaveBeenCalled();
    expect(mockAddToast).toHaveBeenCalledWith('Profile updated.', 'success');
  });

  it('sends null for empty optional fields', async () => {
    mockUser = makeMember({
      bio: null,
      skills: [],
      linkedin_url: null,
      instagram_handle: null,
      github_username: null,
      telegram_handle: null,
    });
    const user = userEvent.setup();
    render(<ProfilePage />);

    await user.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => {
      expect(mockPatch).toHaveBeenCalledWith('/api/auth/profile', {
        name: 'Alice',
        telegram_handle: null,
        bio: null,
        skills: [],
        linkedin_url: null,
        instagram_handle: null,
        github_username: null,
      });
    });
  });

  it('shows error toast on API failure', async () => {
    const { ApiError } = await import('../../api/client');
    mockPatch.mockRejectedValue(new ApiError(422, { error: 'Bio too long' }));
    const user = userEvent.setup();
    render(<ProfilePage />);

    await user.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => {
      expect(mockAddToast).toHaveBeenCalledWith('Bio too long', 'error');
    });
  });

  it('disables save button while saving', async () => {
    let resolvePromise: () => void;
    mockPatch.mockReturnValue(new Promise<void>((r) => { resolvePromise = r; }));
    const user = userEvent.setup();
    render(<ProfilePage />);

    await user.click(screen.getByRole('button', { name: 'Save' }));
    expect(screen.getByRole('button', { name: 'Saving...' })).toBeDisabled();

    resolvePromise!();
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Save' })).not.toBeDisabled();
    });
  });

  it('renders skills as removable tags', async () => {
    const user = userEvent.setup();
    render(<ProfilePage />);

    expect(screen.getByText('react')).toBeInTheDocument();
    expect(screen.getByLabelText('Remove react')).toBeInTheDocument();

    await user.click(screen.getByLabelText('Remove react'));
    expect(screen.queryByText('react')).toBeNull();
    expect(screen.getByText('go')).toBeInTheDocument();
  });

  it('renders nothing special when user is null', () => {
    mockUser = null;
    const { container } = render(<ProfilePage />);
    expect(container.querySelector('form')).toBeInTheDocument();
  });
});
