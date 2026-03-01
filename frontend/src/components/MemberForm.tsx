import { useState, type FormEvent } from 'react';
import type { CreateMemberRequest, Member, UpdateMemberRequest } from '../types';

interface MemberFormProps {
  member?: Member;
  onSubmit: (data: CreateMemberRequest | UpdateMemberRequest) => Promise<void>;
  onCancel: () => void;
  loading?: boolean;
  onToggleActive?: () => void;
  onDelete?: () => void;
  isSelf?: boolean;
}

interface FormErrors {
  email?: string;
  name?: string;
}

export default function MemberForm({ member, onSubmit, onCancel, loading, onToggleActive, onDelete, isSelf }: MemberFormProps) {
  const isEdit = !!member;

  const [email, setEmail] = useState(member?.email ?? '');
  const [name, setName] = useState(member?.name ?? '');
  const [telegramHandle, setTelegramHandle] = useState(member?.telegram_handle ?? '');
  const [isAdmin, setIsAdmin] = useState(member?.is_admin ?? false);
  const [sendInvite, setSendInvite] = useState(false);
  const [errors, setErrors] = useState<FormErrors>({});

  function validate(): FormErrors {
    const errs: FormErrors = {};
    if (!isEdit && !email.trim()) errs.email = 'Email is required';
    if (!isEdit && email.trim() && !email.includes('@')) errs.email = 'Invalid email address';
    if (!name.trim()) errs.name = 'Name is required';
    return errs;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const errs = validate();
    if (Object.keys(errs).length > 0) {
      setErrors(errs);
      return;
    }
    setErrors({});

    if (isEdit) {
      const data: UpdateMemberRequest = {};
      if (name.trim() !== member.name) data.name = name.trim();
      const handle = telegramHandle.trim() || null;
      if (handle !== (member.telegram_handle ?? null)) data.telegram_handle = handle;
      if (isAdmin !== member.is_admin) data.is_admin = isAdmin;
      await onSubmit(data);
    } else {
      const data: CreateMemberRequest = {
        email: email.trim(),
        name: name.trim(),
        is_admin: isAdmin,
        send_invite: sendInvite,
      };
      if (telegramHandle.trim()) data.telegram_handle = telegramHandle.trim();
      await onSubmit(data);
    }
  }

  return (
    <form onSubmit={handleSubmit} noValidate className="space-y-4">
      {!isEdit && (
        <div>
          <label htmlFor="member-email" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
            Email <span className="text-red-500">*</span>
          </label>
          <input
            id="member-email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-gray-900 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
          />
          {errors.email && <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.email}</p>}
        </div>
      )}

      <div>
        <label htmlFor="member-name" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
          Name <span className="text-red-500">*</span>
        </label>
        <input
          id="member-name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-gray-900 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
        />
        {errors.name && <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.name}</p>}
      </div>

      <div>
        <label htmlFor="member-telegram" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
          Telegram Handle
        </label>
        <div className="relative mt-1">
          <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500">
            @
          </span>
          <input
            id="member-telegram"
            type="text"
            value={telegramHandle}
            onChange={(e) => setTelegramHandle(e.target.value)}
            placeholder="username"
            className="block w-full rounded-md border border-gray-300 py-2 pl-8 pr-3 text-gray-900 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
          />
        </div>
      </div>

      <div className="space-y-2">
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={isAdmin}
            onChange={(e) => setIsAdmin(e.target.checked)}
            className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
          />
          <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Admin</span>
        </label>

        {!isEdit && (
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={sendInvite}
              onChange={(e) => setSendInvite(e.target.checked)}
              className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
            />
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Send invitation email</span>
          </label>
        )}
      </div>

      {isEdit && !isSelf && (onToggleActive || onDelete) && (
        <div className="flex gap-3 border-t border-gray-200 pt-4 dark:border-gray-700">
          {onToggleActive && (
            <button
              type="button"
              onClick={onToggleActive}
              className="rounded-md border border-amber-300 px-4 py-2 text-sm font-medium text-amber-700 transition-colors hover:bg-amber-50 dark:border-amber-600 dark:text-amber-400 dark:hover:bg-amber-900/20"
            >
              {member?.is_active ? 'Deactivate' : 'Activate'}
            </button>
          )}
          {onDelete && (
            <button
              type="button"
              onClick={onDelete}
              className="rounded-md border border-red-300 px-4 py-2 text-sm font-medium text-red-700 transition-colors hover:bg-red-50 dark:border-red-600 dark:text-red-400 dark:hover:bg-red-900/20"
            >
              Delete
            </button>
          )}
        </div>
      )}

      <div className="flex justify-end gap-3 pt-2">
        <button
          type="button"
          onClick={onCancel}
          className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={loading}
          className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? 'Saving...' : isEdit ? 'Save Changes' : 'Add Member'}
        </button>
      </div>
    </form>
  );
}
