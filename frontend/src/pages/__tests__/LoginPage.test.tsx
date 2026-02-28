import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { type ReactNode } from 'react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import * as client from '../../api/client';
import { AuthProvider } from '../../context/AuthContext';
import { ToastProvider } from '../../context/ToastContext';
import LoginPage from '../LoginPage';

// Mock the API client
vi.mock('../../api/client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
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

function Wrapper({ children }: { children: ReactNode }) {
  return (
    <MemoryRouter>
      <AuthProvider>
        <ToastProvider>{children}</ToastProvider>
      </AuthProvider>
    </MemoryRouter>
  );
}

describe('LoginPage', () => {
  beforeEach(() => {
    // Default: not authenticated
    (client.api.get as ReturnType<typeof vi.fn>).mockRejectedValue(
      new client.ApiError(401, { error: 'unauthorized' }),
    );
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders email input and submit button', async () => {
    render(<LoginPage />, { wrapper: Wrapper });
    await waitFor(() => {
      expect(screen.getByLabelText('Email')).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: 'Send magic link' })).toBeInTheDocument();
  });

  it('shows confirmation after submitting email', async () => {
    (client.api.post as ReturnType<typeof vi.fn>).mockResolvedValue({});
    const user = userEvent.setup();
    render(<LoginPage />, { wrapper: Wrapper });

    await waitFor(() => {
      expect(screen.getByLabelText('Email')).toBeInTheDocument();
    });
    await user.type(screen.getByLabelText('Email'), 'test@example.com');
    await user.click(screen.getByRole('button', { name: 'Send magic link' }));

    await waitFor(() => {
      expect(screen.getByText('Check your email')).toBeInTheDocument();
    });
    expect(screen.getByText(/test@example.com/)).toBeInTheDocument();
  });

  it('shows error on rate limit (429)', async () => {
    (client.api.post as ReturnType<typeof vi.fn>).mockRejectedValue(
      new client.ApiError(429, { error: 'rate limited' }),
    );
    const user = userEvent.setup();
    render(<LoginPage />, { wrapper: Wrapper });

    await waitFor(() => {
      expect(screen.getByLabelText('Email')).toBeInTheDocument();
    });
    await user.type(screen.getByLabelText('Email'), 'test@example.com');
    await user.click(screen.getByRole('button', { name: 'Send magic link' }));

    await waitFor(() => {
      expect(screen.getByText('Too many requests. Please try again later.')).toBeInTheDocument();
    });
  });

  it('allows returning to email form from confirmation', async () => {
    (client.api.post as ReturnType<typeof vi.fn>).mockResolvedValue({});
    const user = userEvent.setup();
    render(<LoginPage />, { wrapper: Wrapper });

    await waitFor(() => {
      expect(screen.getByLabelText('Email')).toBeInTheDocument();
    });
    await user.type(screen.getByLabelText('Email'), 'test@example.com');
    await user.click(screen.getByRole('button', { name: 'Send magic link' }));

    await waitFor(() => {
      expect(screen.getByText('Check your email')).toBeInTheDocument();
    });
    await user.click(screen.getByText('Use a different email'));
    expect(screen.getByLabelText('Email')).toBeInTheDocument();
  });
});
