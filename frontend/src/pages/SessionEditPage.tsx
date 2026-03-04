import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { api, ApiError } from '../api/client';
import EditScopeModal from '../components/EditScopeModal';
import ImageUpload from '../components/ImageUpload';
import SessionForm from '../components/SessionForm';
import { useToast } from '../context/ToastContext';
import type { APIResponse, SpaceSession, UpdateSessionRequest, UpdateSeriesRequest } from '../types';

export default function SessionEditPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { addToast } = useToast();
  const [scopeModalOpen, setScopeModalOpen] = useState(false);
  const [editScope, setEditScope] = useState<'single' | 'series' | null>(null);
  const [pendingData, setPendingData] = useState<UpdateSessionRequest | null>(null);

  const { data: session, isLoading } = useQuery({
    queryKey: ['session', id],
    queryFn: async () => {
      const res = await api.get<APIResponse<SpaceSession>>(`/api/sessions/${id}`);
      return res.data;
    },
    enabled: !!id,
  });

  const singleMutation = useMutation({
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

  const seriesMutation = useMutation({
    mutationFn: (data: UpdateSeriesRequest) =>
      api.updateSeries(session!.series_id!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] });
      queryClient.invalidateQueries({ queryKey: ['session', id] });
      addToast('All future sessions updated.', 'success');
      navigate(`/sessions/${id}`);
    },
    onError: (err) => {
      const message = err instanceof ApiError ? err.message : 'Failed to update series';
      addToast(message, 'error');
    },
  });

  if (isLoading || !session) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-amber-600 border-t-transparent" />
      </div>
    );
  }

  if (session.status === 'canceled') {
    return (
      <div className="mx-auto max-w-xl text-center">
        <p className="text-stone-500 dark:text-stone-400">This session has been canceled and cannot be edited.</p>
      </div>
    );
  }

  // If editing all sessions, hide the date field (each session has a different date)
  const isSeriesEdit = editScope === 'series';

  async function handleSubmit(data: UpdateSessionRequest | UpdateSeriesRequest) {
    if (session!.series_id && editScope === null) {
      // Show scope modal
      setPendingData(data as UpdateSessionRequest);
      setScopeModalOpen(true);
      return;
    }

    if (isSeriesEdit) {
      // Convert to series request (no date field)
      const seriesData: UpdateSeriesRequest = {};
      const d = data as UpdateSessionRequest;
      if (d.title) seriesData.title = d.title;
      if (d.description !== undefined) seriesData.description = d.description;
      if (d.start_time) seriesData.start_time = d.start_time;
      if (d.end_time) seriesData.end_time = d.end_time;
      if (d.location !== undefined) seriesData.location = d.location;
      if (session!.image_url) seriesData.image_url = session!.image_url;
      await seriesMutation.mutateAsync(seriesData);
    } else {
      await singleMutation.mutateAsync(data as UpdateSessionRequest);
    }
  }

  return (
    <div className="mx-auto max-w-xl">
      <h1 className="mb-6 text-2xl font-bold text-stone-900 dark:text-stone-100">
        {isSeriesEdit ? 'Edit All Future Sessions' : 'Edit Session'}
      </h1>
      <div className="rounded-xl border border-stone-200 bg-white p-6 dark:border-stone-700 dark:bg-stone-800">
        <SessionForm
          session={session}
          onSubmit={handleSubmit}
          loading={singleMutation.isPending || seriesMutation.isPending}
          hideDate={isSeriesEdit}
        >
          <div>
            <label className="block text-sm font-medium text-stone-700 dark:text-stone-300">Image</label>
            <div className="mt-1">
              <ImageUpload
                sessionId={isSeriesEdit ? undefined : session.id}
                seriesId={isSeriesEdit ? session.series_id! : undefined}
                currentImageUrl={session.image_url}
                onUpload={(imageUrl) => {
                  queryClient.setQueryData<SpaceSession>(['session', id], (old) =>
                    old ? { ...old, image_url: imageUrl } : old,
                  );
                  queryClient.invalidateQueries({ queryKey: ['sessions'] });
                }}
                onRemove={() => {
                  queryClient.setQueryData<SpaceSession>(['session', id], (old) =>
                    old ? { ...old, image_url: null } : old,
                  );
                  queryClient.invalidateQueries({ queryKey: ['sessions'] });
                }}
              />
            </div>
          </div>
        </SessionForm>
      </div>

      <EditScopeModal
        open={scopeModalOpen}
        onThisOnly={() => {
          setScopeModalOpen(false);
          setEditScope('single');
          if (pendingData) singleMutation.mutate(pendingData);
        }}
        onAllSessions={() => {
          setScopeModalOpen(false);
          setEditScope('series');
          // Re-render form in series mode — user will re-submit
        }}
        onCancel={() => {
          setScopeModalOpen(false);
          setPendingData(null);
        }}
      />
    </div>
  );
}
