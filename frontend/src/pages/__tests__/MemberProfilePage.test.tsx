import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Member, PublicMember } from '../../types';
import MemberProfilePage from '../MemberProfilePage';

const mockGetPublicProfile = vi.fn();
let mockUser: Member | null = null;

vi.mock('../../api/client', () => ({
  api: {
    getPublicProfile: (...args: unknown[]) => mockGetPublicProfile(...args),
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

vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({
    user: mockUser,
  }),
}));

function makePublicMember(overrides: Partial<PublicMember> = {}): PublicMember {
  return {
    id: 'member-1',
    name: 'Alice',
    telegram_handle: '@alice',
    bio: 'Full-stack developer who loves building things.',
    skills: ['react', 'go', 'postgresql'],
    linkedin_url: 'https://linkedin.com/in/alice',
    instagram_handle: 'alice_gram',
    github_username: 'alice-dev',
    ...overrides,
  };
}

function makeUser(overrides: Partial<Member> = {}): Member {
  return {
    id: 'member-1',
    email: 'alice@example.com',
    name: 'Alice',
    telegram_handle: '@alice',
    is_admin: false,
    is_active: true,
    bio: null,
    skills: [],
    linkedin_url: null,
    instagram_handle: null,
    github_username: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function renderWithRouter(memberId: string) {
  return render(
    <MemoryRouter initialEntries={[`/profile/${memberId}`]}>
      <Routes>
        <Route path="/profile/:id" element={<MemberProfilePage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('MemberProfilePage', () => {
  beforeEach(() => {
    mockUser = makeUser({ id: 'other-member' });
    mockGetPublicProfile.mockResolvedValue({ data: makePublicMember() });
  });

  afterEach(() => {
    vi.restoreAllMocks();
    mockGetPublicProfile.mockReset();
  });

  it('renders profile data after loading', async () => {
    renderWithRouter('member-1');

    await waitFor(() => {
      expect(screen.getByText('Alice')).toBeInTheDocument();
    });
    expect(screen.getByText('Full-stack developer who loves building things.')).toBeInTheDocument();
    expect(screen.getByText('react')).toBeInTheDocument();
    expect(screen.getByText('go')).toBeInTheDocument();
    expect(screen.getByText('postgresql')).toBeInTheDocument();
  });

  it('calls getPublicProfile with the route parameter', async () => {
    renderWithRouter('member-42');

    await waitFor(() => {
      expect(mockGetPublicProfile).toHaveBeenCalledWith('member-42');
    });
  });

  it('renders social links with correct URLs including Telegram', async () => {
    renderWithRouter('member-1');

    await waitFor(() => {
      expect(screen.getByText('Telegram')).toBeInTheDocument();
    });
    expect(screen.getByText('Telegram').closest('a')).toHaveAttribute(
      'href',
      'https://t.me/alice',
    );
    expect(screen.getByText('LinkedIn').closest('a')).toHaveAttribute(
      'href',
      'https://linkedin.com/in/alice',
    );
    expect(screen.getByText('Instagram').closest('a')).toHaveAttribute(
      'href',
      'https://instagram.com/alice_gram',
    );
    expect(screen.getByText('GitHub').closest('a')).toHaveAttribute(
      'href',
      'https://github.com/alice-dev',
    );
  });

  it('shows Edit Profile link when viewing own profile', async () => {
    mockUser = makeUser({ id: 'member-1' });
    renderWithRouter('member-1');

    await waitFor(() => {
      expect(screen.getByText('Edit Profile')).toBeInTheDocument();
    });
    expect(screen.getByText('Edit Profile').closest('a')).toHaveAttribute('href', '/profile');
  });

  it('does not show Edit Profile link for other members', async () => {
    mockUser = makeUser({ id: 'other-member' });
    renderWithRouter('member-1');

    await waitFor(() => {
      expect(screen.getByText('Alice')).toBeInTheDocument();
    });
    expect(screen.queryByText('Edit Profile')).not.toBeInTheDocument();
  });

  it('shows not found state on 404', async () => {
    const { ApiError } = await import('../../api/client');
    mockGetPublicProfile.mockRejectedValue(new ApiError(404, { error: 'Not found' }));

    renderWithRouter('nonexistent');

    await waitFor(() => {
      expect(screen.getByText('Member not found')).toBeInTheDocument();
    });
    expect(screen.getByText("This member doesn't exist or their account is inactive.")).toBeInTheDocument();
    expect(screen.getByText('Back to sessions').closest('a')).toHaveAttribute('href', '/');
  });

  it('hides optional sections when data is null/empty', async () => {
    mockGetPublicProfile.mockResolvedValue({
      data: makePublicMember({
        bio: null,
        skills: [],
        linkedin_url: null,
        instagram_handle: null,
        github_username: null,
        telegram_handle: null,
      }),
    });

    renderWithRouter('member-1');

    await waitFor(() => {
      expect(screen.getByText('Alice')).toBeInTheDocument();
    });
    expect(screen.queryByText('About')).not.toBeInTheDocument();
    expect(screen.queryByText('Skills')).not.toBeInTheDocument();
    expect(screen.queryByText('Social Links')).not.toBeInTheDocument();
  });

  it('shows loading state initially', () => {
    mockGetPublicProfile.mockReturnValue(new Promise(() => {}));
    renderWithRouter('member-1');

    expect(screen.getByText('Loading profile…')).toBeInTheDocument();
  });

  it('shows Social Links section heading', async () => {
    renderWithRouter('member-1');

    await waitFor(() => {
      expect(screen.getByText('Social Links')).toBeInTheDocument();
    });
  });

  it('shows Telegram as a t.me link in Social Links', async () => {
    renderWithRouter('member-1');

    await waitFor(() => {
      expect(screen.getByText('Telegram')).toBeInTheDocument();
    });
    const link = screen.getByText('Telegram').closest('a');
    expect(link).toHaveAttribute('href', 'https://t.me/alice');
    expect(link).toHaveAttribute('target', '_blank');
  });

  it('shows Social Links when only telegram is present', async () => {
    mockGetPublicProfile.mockResolvedValue({
      data: makePublicMember({
        linkedin_url: null,
        instagram_handle: null,
        github_username: null,
        telegram_handle: 'alice',
      }),
    });

    renderWithRouter('member-1');

    await waitFor(() => {
      expect(screen.getByText('Social Links')).toBeInTheDocument();
    });
    expect(screen.getByText('Telegram')).toBeInTheDocument();
  });

  it('strips @ from instagram handle in URL', async () => {
    mockGetPublicProfile.mockResolvedValue({
      data: makePublicMember({ instagram_handle: '@alice_gram' }),
    });

    renderWithRouter('member-1');

    await waitFor(() => {
      expect(screen.getByText('Instagram')).toBeInTheDocument();
    });
    expect(screen.getByText('Instagram').closest('a')).toHaveAttribute(
      'href',
      'https://instagram.com/alice_gram',
    );
  });

  it('opens social links in new tab', async () => {
    renderWithRouter('member-1');

    await waitFor(() => {
      expect(screen.getByText('LinkedIn')).toBeInTheDocument();
    });
    for (const linkText of ['GitHub', 'LinkedIn', 'Telegram', 'Instagram']) {
      const anchor = screen.getByText(linkText).closest('a');
      expect(anchor).toHaveAttribute('target', '_blank');
      expect(anchor).toHaveAttribute('rel', 'noopener noreferrer');
    }
  });
});
