import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { api, ApiError } from '../api/client';
import SessionForm from '../components/SessionForm';
import { useToast } from '../context/ToastContext';
import type { APIResponse, SpaceSession, UpdateSessionRequest } from '../types';

export default function SessionEditPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { addToast } = useToast();

  const { data: session, isLoading } = useQuery({
    queryKey: ['session', id],
    queryFn: async () => {
      const res = await api.get<APIResponse<SpaceSession>>(`/api/sessions/${id}`);
      return res.data;
    },
    enabled: !!id,
  });

  const mutation = useMutation({
    mutationFn: (data: UpdateSessionRequest) =>
      api.patch<APIResponse<SpaceSession>>(`/api/sessions/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] });
      queryClient.invalidateQueries({ queryKey: ['session', id] });
      addToast('Session updated.', 'success');
      navigate(`/sessions/${id}`);
    },
    onError: (err) => {
      const message = err instanceof ApiError ? err.message : 'Failed to update session';
      addToast(message, 'error');
    },
  });

  if (isLoading || !session) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-indigo-600 border-t-transparent" />
      </div>
    );
  }

  if (session.status === 'canceled') {
    return (
      <div className="mx-auto max-w-xl text-center">
        <p className="text-gray-500 dark:text-gray-400">This session has been canceled and cannot be edited.</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-xl">
      <h1 className="mb-6 text-2xl font-bold text-gray-900 dark:text-gray-100">Edit Session</h1>
      <div className="rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <SessionForm
          session={session}
          onSubmit={async (data) => {
            await mutation.mutateAsync(data as UpdateSessionRequest);
          }}
          loading={mutation.isPending}
        />
      </div>
    </div>
  );
}
