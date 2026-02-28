import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import EmptyState from '../EmptyState';

describe('EmptyState', () => {
  it('renders heading', () => {
    render(<EmptyState heading="No sessions" />);
    expect(screen.getByText('No sessions')).toBeInTheDocument();
  });

  it('renders subtext', () => {
    render(<EmptyState heading="No sessions" subtext="Create your first session" />);
    expect(screen.getByText('Create your first session')).toBeInTheDocument();
  });

  it('does not render subtext when not provided', () => {
    const { container } = render(<EmptyState heading="Empty" />);
    expect(container.querySelectorAll('p')).toHaveLength(0);
  });

  it('renders CTA button when action provided', () => {
    const onClick = vi.fn();
    render(<EmptyState heading="Empty" action={{ label: 'Create', onClick }} />);
    expect(screen.getByRole('button', { name: 'Create' })).toBeInTheDocument();
  });

  it('calls action onClick when button clicked', async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    render(<EmptyState heading="Empty" action={{ label: 'Create', onClick }} />);
    await user.click(screen.getByRole('button', { name: 'Create' }));
    expect(onClick).toHaveBeenCalledOnce();
  });

  it('does not render button when no action', () => {
    render(<EmptyState heading="Empty" />);
    expect(screen.queryByRole('button')).toBeNull();
  });

  it('renders icon when provided', () => {
    render(<EmptyState heading="Empty" icon={<span data-testid="icon">Icon</span>} />);
    expect(screen.getByTestId('icon')).toBeInTheDocument();
  });
});
