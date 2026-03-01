import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { api, ApiError } from '../api/client';
import ConfirmModal from '../components/ConfirmModal';
import EmptyState from '../components/EmptyState';
import MemberForm from '../components/MemberForm';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import type {
  APIResponse,
  CreateMemberRequest,
  Member,
  UpdateMemberRequest,
} from '../types';

type ActiveFilter = 'true' | 'false' | 'all';

export default function MembersPage() {
  const queryClient = useQueryClient();
  const { user } = useAuth();
  const { addToast } = useToast();

  const [activeFilter, setActiveFilter] = useState<ActiveFilter>('true');
  const [showForm, setShowForm] = useState(false);
  const [editingMember, setEditingMember] = useState<Member | null>(null);
  const [deletingMember, setDeletingMember] = useState<Member | null>(null);

  const { data: members = [], isLoading } = useQuery({
    queryKey: ['members', activeFilter],
    queryFn: async () => {
      const res = await api.get<APIResponse<Member[]>>(
        `/api/members?active=${activeFilter}`,
      );
      return res.data;
    },
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateMemberRequest) =>
      api.post<APIResponse<Member>>('/api/members', data),
    onSuccess: () => {
      addToast('Member added.', 'success');
      setShowForm(false);
      queryClient.invalidateQueries({ queryKey: ['members'] });
    },
    onError: (err) => {
      const message =
        err instanceof ApiError ? err.message : 'Failed to create member';
      addToast(message, 'error');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateMemberRequest }) =>
      api.patch<APIResponse<Member>>(`/api/members/${id}`, data),
    onSuccess: () => {
      addToast('Member updated.', 'success');
      setEditingMember(null);
      queryClient.invalidateQueries({ queryKey: ['members'] });
    },
    onError: (err) => {
      const message =
        err instanceof ApiError ? err.message : 'Failed to update member';
      addToast(message, 'error');
    },
  });

  const toggleActiveMutation = useMutation({
    mutationFn: ({ id, is_active }: { id: string; is_active: boolean }) =>
      api.patch<APIResponse<Member>>(`/api/members/${id}`, { is_active }),
    onSuccess: (_, variables) => {
      addToast(
        variables.is_active ? 'Member activated.' : 'Member deactivated.',
        'success',
      );
      queryClient.invalidateQueries({ queryKey: ['members'] });
    },
    onError: (err) => {
      const message =
        err instanceof ApiError ? err.message : 'Failed to update member';
      addToast(message, 'error');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/api/members/${id}`),
    onSuccess: () => {
      addToast('Member deleted.', 'success');
      setDeletingMember(null);
      queryClient.invalidateQueries({ queryKey: ['members'] });
    },
    onError: (err) => {
      setDeletingMember(null);
      const message =
        err instanceof ApiError
          ? err.message
          : 'Failed to delete member';
      addToast(message, 'error');
    },
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-indigo-600 border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header: filter + add button */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex gap-1 rounded-md border border-gray-300 p-1 dark:border-gray-600">
          {([['true', 'Active'], ['false', 'Inactive'], ['all', 'All']] as const).map(
            ([value, label]) => (
              <button
                key={value}
                onClick={() => setActiveFilter(value)}
                className={`rounded px-3 py-1 text-sm font-medium transition-colors ${
                  activeFilter === value
                    ? 'bg-indigo-600 text-white'
                    : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700'
                }`}
              >
                {label}
              </button>
            ),
          )}
        </div>
        <button
          onClick={() => {
            setEditingMember(null);
            setShowForm(true);
          }}
          className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700"
        >
          Add Member
        </button>
      </div>

      {/* Create / Edit form */}
      {(showForm || editingMember) && (
        <div className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-800">
          <h2 className="mb-4 text-lg font-semibold text-gray-900 dark:text-gray-100">
            {editingMember ? 'Edit Member' : 'Add Member'}
          </h2>
          <MemberForm
            member={editingMember ?? undefined}
            loading={createMutation.isPending || updateMutation.isPending}
            isSelf={!!editingMember && user?.id === editingMember.id}
            onCancel={() => {
              setShowForm(false);
              setEditingMember(null);
            }}
            onSubmit={async (data) => {
              if (editingMember) {
                await updateMutation.mutateAsync({
                  id: editingMember.id,
                  data: data as UpdateMemberRequest,
                });
              } else {
                await createMutation.mutateAsync(data as CreateMemberRequest);
              }
            }}
            onToggleActive={
              editingMember
                ? () => {
                    toggleActiveMutation.mutate({
                      id: editingMember.id,
                      is_active: !editingMember.is_active,
                    });
                    setEditingMember(null);
                  }
                : undefined
            }
            onDelete={
              editingMember
                ? () => {
                    setDeletingMember(editingMember);
                    setEditingMember(null);
                  }
                : undefined
            }
          />
        </div>
      )}

      {/* Empty state */}
      {members.length === 0 && !showForm && !editingMember && (
        <EmptyState
          icon={
            <svg
              className="h-16 w-16"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={1.5}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z"
              />
            </svg>
          }
          heading="No members yet"
          subtext="Add your first member to get started."
          action={{
            label: 'Add Member',
            onClick: () => setShowForm(true),
          }}
        />
      )}

      {/* Desktop table */}
      {members.length > 0 && (
        <>
          <div className="hidden overflow-hidden rounded-lg border border-gray-200 dark:border-gray-700 md:block">
            <table className="w-full text-left text-sm">
              <thead className="bg-gray-50 dark:bg-gray-800">
                <tr>
                  <th className="px-4 py-3 font-medium text-gray-600 dark:text-gray-400">Name</th>
                  <th className="px-4 py-3 font-medium text-gray-600 dark:text-gray-400">Email</th>
                  <th className="px-4 py-3 font-medium text-gray-600 dark:text-gray-400">Telegram</th>
                  <th className="px-4 py-3 font-medium text-gray-600 dark:text-gray-400">Role</th>
                  <th className="px-4 py-3 font-medium text-gray-600 dark:text-gray-400">Status</th>
                  <th className="px-4 py-3 font-medium text-gray-600 dark:text-gray-400">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {members.map((m) => (
                  <tr key={m.id} className="bg-white dark:bg-gray-900">
                    <td className="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">
                      {m.name}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{m.email}</td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                      {m.telegram_handle ? `@${m.telegram_handle}` : '—'}
                    </td>
                    <td className="px-4 py-3">
                      {m.is_admin && (
                        <span className="inline-flex rounded-full bg-indigo-100 px-2 py-0.5 text-xs font-medium text-indigo-700 dark:bg-indigo-900 dark:text-indigo-300">
                          Admin
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
                          m.is_active
                            ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                            : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
                        }`}
                      >
                        {m.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex gap-2">
                        {user?.id !== m.id && (
                          <button
                            onClick={() =>
                              toggleActiveMutation.mutate({
                                id: m.id,
                                is_active: !m.is_active,
                              })
                            }
                            className="text-gray-500 hover:text-amber-600 dark:text-gray-400 dark:hover:text-amber-400"
                            aria-label={m.is_active ? `Deactivate ${m.name}` : `Activate ${m.name}`}
                            title={m.is_active ? 'Deactivate' : 'Activate'}
                          >
                            {m.is_active ? (
                              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
                              </svg>
                            ) : (
                              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                              </svg>
                            )}
                          </button>
                        )}
                        <button
                          onClick={() => {
                            setShowForm(false);
                            setEditingMember(m);
                          }}
                          className="text-gray-500 hover:text-indigo-600 dark:text-gray-400 dark:hover:text-indigo-400"
                          aria-label={`Edit ${m.name}`}
                        >
                          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                          </svg>
                        </button>
                        {user?.id !== m.id && (
                          <button
                            onClick={() => setDeletingMember(m)}
                            className="text-gray-500 hover:text-red-600 dark:text-gray-400 dark:hover:text-red-400"
                            aria-label={`Delete ${m.name}`}
                          >
                            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                              <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                            </svg>
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Mobile card list */}
          <div className="space-y-3 md:hidden">
            {members.map((m) => (
              <div
                key={m.id}
                className="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800"
              >
                <div className="flex items-start justify-between">
                  <div>
                    <p className="font-medium text-gray-900 dark:text-gray-100">{m.name}</p>
                    <p className="text-sm text-gray-500 dark:text-gray-400">{m.email}</p>
                  </div>
                  <div className="flex gap-2">
                    {user?.id !== m.id && (
                      <button
                        onClick={() =>
                          toggleActiveMutation.mutate({
                            id: m.id,
                            is_active: !m.is_active,
                          })
                        }
                        className="text-gray-500 hover:text-amber-600 dark:text-gray-400 dark:hover:text-amber-400"
                        aria-label={m.is_active ? `Deactivate ${m.name}` : `Activate ${m.name}`}
                        title={m.is_active ? 'Deactivate' : 'Activate'}
                      >
                        {m.is_active ? (
                          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
                          </svg>
                        ) : (
                          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                          </svg>
                        )}
                      </button>
                    )}
                    <button
                      onClick={() => {
                        setShowForm(false);
                        setEditingMember(m);
                      }}
                      className="text-gray-500 hover:text-indigo-600 dark:text-gray-400 dark:hover:text-indigo-400"
                      aria-label={`Edit ${m.name}`}
                    >
                      <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                      </svg>
                    </button>
                    {user?.id !== m.id && (
                      <button
                        onClick={() => setDeletingMember(m)}
                        className="text-gray-500 hover:text-red-600 dark:text-gray-400 dark:hover:text-red-400"
                        aria-label={`Delete ${m.name}`}
                      >
                        <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                        </svg>
                      </button>
                    )}
                  </div>
                </div>
                <div className="mt-2 flex flex-wrap gap-2">
                  {m.telegram_handle && (
                    <span className="text-sm text-gray-500 dark:text-gray-400">
                      @{m.telegram_handle}
                    </span>
                  )}
                  {m.is_admin && (
                    <span className="inline-flex rounded-full bg-indigo-100 px-2 py-0.5 text-xs font-medium text-indigo-700 dark:bg-indigo-900 dark:text-indigo-300">
                      Admin
                    </span>
                  )}
                  <span
                    className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
                      m.is_active
                        ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                        : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
                    }`}
                  >
                    {m.is_active ? 'Active' : 'Inactive'}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </>
      )}

      {/* Delete confirmation modal */}
      <ConfirmModal
        open={!!deletingMember}
        title="Delete this member?"
        message={`This will permanently delete ${deletingMember?.name ?? 'this member'}. This action cannot be undone.`}
        confirmLabel="Delete"
        onConfirm={() => {
          if (deletingMember) deleteMutation.mutate(deletingMember.id);
        }}
        onCancel={() => setDeletingMember(null)}
      />
    </div>
  );
}
