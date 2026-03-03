import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { api, ApiError } from '../api/client';
import SessionForm from '../components/SessionForm';
import { useToast } from '../context/ToastContext';
import type { APIResponse, CreateSessionRequest, SpaceSession } from '../types';

export default function SessionCreatePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { addToast } = useToast();

  const mutation = useMutation({
    mutationFn: (data: CreateSessionRequest) =>
      api.post<APIResponse<SpaceSession | SpaceSession[]>>('/api/sessions', data),
    onSuccess: (res, variables) => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] });
      if (variables.repeat_weekly && variables.repeat_weekly > 0) {
        const sessions = Array.isArray(res.data) ? res.data : [res.data];
        addToast(`Created ${sessions.length} sessions.`, 'success');
      } else {
        addToast('Session created.', 'success');
      }
      navigate('/');
    },
    onError: (err) => {
      const message = err instanceof ApiError ? err.message : 'Failed to create session';
      addToast(message, 'error');
    },
  });

  return (
    <div className="mx-auto max-w-xl">
      <h1 className="mb-6 text-2xl font-bold text-stone-900 dark:text-stone-100">Create Session</h1>
      <div className="rounded-xl border border-stone-200 bg-white p-6 dark:border-stone-700 dark:bg-stone-800">
        <SessionForm
          onSubmit={async (data) => {
            await mutation.mutateAsync(data as CreateSessionRequest);
          }}
          loading={mutation.isPending}
        />
      </div>
    </div>
  );
}
