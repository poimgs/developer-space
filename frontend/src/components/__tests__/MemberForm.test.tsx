import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { Member } from '../../types';
import MemberForm from '../MemberForm';

function makeMember(overrides: Partial<Member> = {}): Member {
  return {
    id: 'member-1',
    email: 'jane@example.com',
    name: 'Jane Doe',
    telegram_handle: 'janedoe',
    is_admin: false,
    is_active: true,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('MemberForm', () => {
  describe('create mode', () => {
    it('renders all fields with empty values', () => {
      render(<MemberForm onSubmit={vi.fn()} onCancel={vi.fn()} />);
      expect(screen.getByLabelText(/^email/i)).toHaveValue('');
      expect(screen.getByLabelText(/^name/i)).toHaveValue('');
      expect(screen.getByLabelText(/telegram handle/i)).toHaveValue('');
      expect(screen.getByLabelText(/^admin$/i)).not.toBeChecked();
      expect(screen.getByLabelText(/send invitation email/i)).not.toBeChecked();
    });

    it('shows Add Member submit button', () => {
      render(<MemberForm onSubmit={vi.fn()} onCancel={vi.fn()} />);
      expect(screen.getByRole('button', { name: 'Add Member' })).toBeInTheDocument();
    });

    it('shows Cancel button', () => {
      render(<MemberForm onSubmit={vi.fn()} onCancel={vi.fn()} />);
      expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
    });

    it('calls onCancel when Cancel is clicked', async () => {
      const user = userEvent.setup();
      const onCancel = vi.fn();
      render(<MemberForm onSubmit={vi.fn()} onCancel={onCancel} />);
      await user.click(screen.getByRole('button', { name: 'Cancel' }));
      expect(onCancel).toHaveBeenCalledTimes(1);
    });

    it('validates required fields on submit', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn();
      render(<MemberForm onSubmit={onSubmit} onCancel={vi.fn()} />);
      await user.click(screen.getByRole('button', { name: 'Add Member' }));
      expect(screen.getByText('Email is required')).toBeInTheDocument();
      expect(screen.getByText('Name is required')).toBeInTheDocument();
      expect(onSubmit).not.toHaveBeenCalled();
    });

    it('validates email format', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn();
      render(<MemberForm onSubmit={onSubmit} onCancel={vi.fn()} />);

      await user.type(screen.getByLabelText(/^email/i), 'not-an-email');
      await user.type(screen.getByLabelText(/^name/i), 'Jane');
      await user.click(screen.getByRole('button', { name: 'Add Member' }));

      expect(screen.getByText('Invalid email address')).toBeInTheDocument();
      expect(onSubmit).not.toHaveBeenCalled();
    });

    it('submits valid create form data', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      render(<MemberForm onSubmit={onSubmit} onCancel={vi.fn()} />);

      await user.type(screen.getByLabelText(/^email/i), 'jane@example.com');
      await user.type(screen.getByLabelText(/^name/i), 'Jane Doe');
      await user.type(screen.getByLabelText(/telegram handle/i), 'janedoe');

      await user.click(screen.getByRole('button', { name: 'Add Member' }));

      expect(onSubmit).toHaveBeenCalledWith({
        email: 'jane@example.com',
        name: 'Jane Doe',
        telegram_handle: 'janedoe',
        is_admin: false,
        send_invite: false,
      });
    });

    it('submits with is_admin and send_invite checked', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      render(<MemberForm onSubmit={onSubmit} onCancel={vi.fn()} />);

      await user.type(screen.getByLabelText(/^email/i), 'admin@example.com');
      await user.type(screen.getByLabelText(/^name/i), 'Admin');
      await user.click(screen.getByLabelText(/^admin$/i));
      await user.click(screen.getByLabelText(/send invitation email/i));

      await user.click(screen.getByRole('button', { name: 'Add Member' }));

      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({ is_admin: true, send_invite: true }),
      );
    });

    it('does not include telegram_handle when empty', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      render(<MemberForm onSubmit={onSubmit} onCancel={vi.fn()} />);

      await user.type(screen.getByLabelText(/^email/i), 'jane@example.com');
      await user.type(screen.getByLabelText(/^name/i), 'Jane');

      await user.click(screen.getByRole('button', { name: 'Add Member' }));

      const call = onSubmit.mock.calls[0][0];
      expect(call.telegram_handle).toBeUndefined();
    });

    it('shows Saving... when loading', () => {
      render(<MemberForm onSubmit={vi.fn()} onCancel={vi.fn()} loading />);
      expect(screen.getByRole('button', { name: 'Saving...' })).toBeDisabled();
    });

    it('shows @ prefix hint in telegram field', () => {
      render(<MemberForm onSubmit={vi.fn()} onCancel={vi.fn()} />);
      expect(screen.getByText('@')).toBeInTheDocument();
    });
  });

  describe('edit mode', () => {
    it('pre-populates form with member data', () => {
      const member = makeMember();
      render(<MemberForm member={member} onSubmit={vi.fn()} onCancel={vi.fn()} />);
      expect(screen.getByLabelText(/^name/i)).toHaveValue('Jane Doe');
      expect(screen.getByLabelText(/telegram handle/i)).toHaveValue('janedoe');
      expect(screen.getByLabelText(/^admin$/i)).not.toBeChecked();
    });

    it('does not show email field in edit mode', () => {
      render(<MemberForm member={makeMember()} onSubmit={vi.fn()} onCancel={vi.fn()} />);
      expect(screen.queryByLabelText(/email/i)).toBeNull();
    });

    it('does not show send invitation checkbox in edit mode', () => {
      render(<MemberForm member={makeMember()} onSubmit={vi.fn()} onCancel={vi.fn()} />);
      expect(screen.queryByLabelText(/send invitation email/i)).toBeNull();
    });

    it('shows Save Changes submit button', () => {
      render(<MemberForm member={makeMember()} onSubmit={vi.fn()} onCancel={vi.fn()} />);
      expect(screen.getByRole('button', { name: 'Save Changes' })).toBeInTheDocument();
    });

    it('sends only changed fields on edit', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      const member = makeMember();
      render(<MemberForm member={member} onSubmit={onSubmit} onCancel={vi.fn()} />);

      const nameInput = screen.getByLabelText(/^name/i);
      await user.clear(nameInput);
      await user.type(nameInput, 'Jane Smith');

      await user.click(screen.getByRole('button', { name: 'Save Changes' }));

      expect(onSubmit).toHaveBeenCalledWith({ name: 'Jane Smith' });
    });

    it('sends empty object when nothing changed', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      render(<MemberForm member={makeMember()} onSubmit={onSubmit} onCancel={vi.fn()} />);

      await user.click(screen.getByRole('button', { name: 'Save Changes' }));

      expect(onSubmit).toHaveBeenCalledWith({});
    });

    it('detects is_admin toggle as a change', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      const member = makeMember({ is_admin: false });
      render(<MemberForm member={member} onSubmit={onSubmit} onCancel={vi.fn()} />);

      await user.click(screen.getByLabelText(/^admin$/i));
      await user.click(screen.getByRole('button', { name: 'Save Changes' }));

      expect(onSubmit).toHaveBeenCalledWith({ is_admin: true });
    });

    it('handles null telegram_handle correctly', () => {
      const member = makeMember({ telegram_handle: null });
      render(<MemberForm member={member} onSubmit={vi.fn()} onCancel={vi.fn()} />);
      expect(screen.getByLabelText(/telegram handle/i)).toHaveValue('');
    });

    it('validates name is required on edit', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn();
      render(<MemberForm member={makeMember()} onSubmit={onSubmit} onCancel={vi.fn()} />);

      const nameInput = screen.getByLabelText(/^name/i);
      await user.clear(nameInput);
      await user.click(screen.getByRole('button', { name: 'Save Changes' }));

      expect(screen.getByText('Name is required')).toBeInTheDocument();
      expect(onSubmit).not.toHaveBeenCalled();
    });
  });

  describe('required field indicators', () => {
    it('shows asterisks on required fields in create mode', () => {
      render(<MemberForm onSubmit={vi.fn()} onCancel={vi.fn()} />);
      const asterisks = document.querySelectorAll('.text-red-500');
      // Email and Name are required
      expect(asterisks.length).toBe(2);
    });

    it('shows asterisk on name field in edit mode', () => {
      render(<MemberForm member={makeMember()} onSubmit={vi.fn()} onCancel={vi.fn()} />);
      const asterisks = document.querySelectorAll('.text-red-500');
      // Only Name is required (email not shown)
      expect(asterisks.length).toBe(1);
    });
  });
});
