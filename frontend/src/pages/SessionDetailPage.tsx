import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { api, ApiError } from '../api/client';
import ConfirmModal from '../components/ConfirmModal';
import GuestList from '../components/GuestList';
import ImageUpload from '../components/ImageUpload';
import StatusBadge from '../components/StatusBadge';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import type { APIResponse, RSVP, RSVPWithMember, SpaceSession } from '../types';
import { formatDateLabel } from '../components/SessionCard';

export default function SessionDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { addToast } = useToast();
  const { user } = useAuth();
  const isAdmin = user?.is_admin ?? false;
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [cancelSessionOpen, setCancelSessionOpen] = useState(false);
  const [stopSeriesOpen, setStopSeriesOpen] = useState(false);

  const { data: session, isLoading } = useQuery({
    queryKey: ['session', id],
    queryFn: async () => {
      const res = await api.get<APIResponse<SpaceSession>>(`/api/sessions/${id}`);
      return res.data;
    },
    enabled: !!id,
  });

  const { data: rsvps = [] } = useQuery({
    queryKey: ['rsvps', id],
    queryFn: async () => {
      const res = await api.get<APIResponse<RSVPWithMember[]>>(`/api/sessions/${id}/rsvps`);
      return res.data;
    },
    enabled: !!id,
  });

  const rsvpMutation = useMutation({
    mutationFn: () => api.post<APIResponse<RSVP>>(`/api/sessions/${id}/rsvp`),
    onSuccess: () => {
      addToast("You're in!", 'success');
      queryClient.invalidateQueries({ queryKey: ['session', id] });
      queryClient.invalidateQueries({ queryKey: ['rsvps', id] });
      queryClient.invalidateQueries({ queryKey: ['sessions'] });
    },
    onError: (err) => {
      const message = err instanceof ApiError ? err.message : 'Failed to RSVP';
      addToast(message, 'error');
    },
  });

  const cancelRSVPMutation = useMutation({
    mutationFn: () => api.delete(`/api/sessions/${id}/rsvp`),
    onSuccess: () => {
      addToast('RSVP canceled.', 'info');
      queryClient.invalidateQueries({ queryKey: ['session', id] });
      queryClient.invalidateQueries({ queryKey: ['rsvps', id] });
      queryClient.invalidateQueries({ queryKey: ['sessions'] });
    },
    onError: (err) => {
      const message = err instanceof ApiError ? err.message : 'Failed to cancel RSVP';
      addToast(message, 'error');
    },
  });

  const cancelSessionMutation = useMutation({
    mutationFn: () => api.delete(`/api/sessions/${id}`),
    onSuccess: () => {
      addToast('Session canceled.', 'success');
      queryClient.invalidateQueries({ queryKey: ['sessions'] });
      navigate('/');
    },
    onError: (err) => {
      const message = err instanceof ApiError ? err.message : 'Failed to cancel session';
      addToast(message, 'error');
    },
  });

  const stopSeriesMutation = useMutation({
    mutationFn: (seriesId: string) => api.delete(`/api/sessions/series/${seriesId}`),
    onSuccess: () => {
      addToast('Recurring series stopped. No new sessions will be generated.', 'success');
      queryClient.invalidateQueries({ queryKey: ['session', id] });
      queryClient.invalidateQueries({ queryKey: ['sessions'] });
    },
    onError: (err) => {
      const message = err instanceof ApiError ? err.message : 'Failed to stop series';
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

  const isCanceled = session.status === 'canceled';
  const isFull = session.rsvp_count >= session.capacity;

  return (
    <div className="mx-auto max-w-2xl">
      <Link
        to="/"
        className="mb-4 inline-flex items-center text-sm text-stone-500 hover:text-stone-700 dark:text-stone-400 dark:hover:text-stone-200"
      >
        <svg className="mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M15 19l-7-7 7-7" />
        </svg>
        Back to sessions
      </Link>

      <div className={`rounded-xl border border-stone-200 bg-white dark:border-stone-700 dark:bg-stone-800 ${isCanceled ? 'opacity-60' : ''}`}>
        {/* Hero image */}
        {session.image_url && (
          <img
            src={session.image_url}
            alt={session.title}
            className="h-48 w-full rounded-t-xl object-cover"
          />
        )}

        <div className="p-6">
          <div className="flex items-start justify-between gap-2">
            <h1 className={`text-xl font-bold text-stone-900 dark:text-stone-100 ${isCanceled ? 'line-through' : ''}`}>
              {session.title}
            </h1>
            <div className="flex items-center gap-2">
              {session.series_id && (
                <span className="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-900/30 dark:text-amber-300">
                  <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Recurring
                </span>
              )}
              <StatusBadge status={session.status} />
            </div>
          </div>

          {session.description && (
            <p className="mt-2 text-sm text-stone-600 dark:text-stone-400">{session.description}</p>
          )}

          <div className="mt-4 space-y-2 text-sm text-stone-600 dark:text-stone-300">
            <p>{formatDateLabel(session.date)}</p>
            <p>{session.start_time} – {session.end_time}</p>
            {session.location && (
              <p className="flex items-center gap-1">
                <svg className="h-4 w-4 flex-shrink-0 text-stone-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M15 10.5a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z" />
                </svg>
                {session.location}
              </p>
            )}
            <p>
              {isFull && !session.user_rsvped ? (
                <span className="font-medium text-red-600 dark:text-red-400">Full</span>
              ) : (
                `${session.rsvp_count}/${session.capacity} spots`
              )}
            </p>
          </div>

          {!isCanceled && (
            <div className="mt-6 flex flex-wrap items-center gap-3">
              {session.user_rsvped ? (
                <button
                  onClick={() => setConfirmOpen(true)}
                  disabled={cancelRSVPMutation.isPending}
                  className="rounded-md border border-amber-600 px-4 py-2 text-sm font-medium text-amber-600 transition-colors hover:bg-amber-50 disabled:opacity-50 dark:border-amber-400 dark:text-amber-400 dark:hover:bg-amber-900/30"
                >
                  Cancel RSVP
                </button>
              ) : isFull ? (
                <span className="inline-block rounded-md bg-stone-100 px-4 py-2 text-sm font-medium text-stone-400 dark:bg-stone-700 dark:text-stone-500">
                  Full
                </span>
              ) : (
                <button
                  onClick={() => rsvpMutation.mutate()}
                  disabled={rsvpMutation.isPending}
                  className="rounded-md bg-amber-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-amber-700 disabled:opacity-50"
                >
                  RSVP
                </button>
              )}

              {isAdmin && (
                <>
                  <Link
                    to={`/sessions/${session.id}/edit`}
                    className="rounded-md border border-stone-300 px-4 py-2 text-sm font-medium text-stone-700 transition-colors hover:bg-stone-50 dark:border-stone-600 dark:text-stone-300 dark:hover:bg-stone-700"
                  >
                    Edit
                  </Link>
                  <button
                    onClick={() => setCancelSessionOpen(true)}
                    className="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-red-700"
                  >
                    Cancel Session
                  </button>
                  {session.series_id && (
                    <button
                      onClick={() => setStopSeriesOpen(true)}
                      disabled={stopSeriesMutation.isPending}
                      className="rounded-md border border-orange-500 px-4 py-2 text-sm font-medium text-orange-600 transition-colors hover:bg-orange-50 disabled:opacity-50 dark:border-orange-400 dark:text-orange-400 dark:hover:bg-orange-900/30"
                    >
                      Stop Recurring
                    </button>
                  )}
                </>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Admin image upload */}
      {isAdmin && !isCanceled && (
        <div className="mt-6 rounded-xl border border-stone-200 bg-white p-6 dark:border-stone-700 dark:bg-stone-800">
          <h2 className="mb-3 text-lg font-semibold text-stone-900 dark:text-stone-100">Session Image</h2>
          <ImageUpload
            sessionId={session.id}
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
      )}

      <div className="mt-6 rounded-xl border border-stone-200 bg-white p-6 dark:border-stone-700 dark:bg-stone-800">
        <h2 className="text-lg font-semibold text-stone-900 dark:text-stone-100">Guest List</h2>
        <GuestList rsvps={rsvps} />
      </div>

      <ConfirmModal
        open={confirmOpen}
        title="Cancel your RSVP?"
        message={`You'll lose your spot for "${session.title}" on ${formatDateLabel(session.date)}.`}
        confirmLabel="Cancel RSVP"
        onConfirm={() => {
          setConfirmOpen(false);
          cancelRSVPMutation.mutate();
        }}
        onCancel={() => setConfirmOpen(false)}
      />

      <ConfirmModal
        open={cancelSessionOpen}
        title="Cancel this session?"
        message={`This will cancel "${session.title}" on ${formatDateLabel(session.date)}. All RSVPed members will be notified. This action cannot be undone.`}
        confirmLabel="Cancel Session"
        onConfirm={() => {
          setCancelSessionOpen(false);
          cancelSessionMutation.mutate();
        }}
        onCancel={() => setCancelSessionOpen(false)}
      />

      <ConfirmModal
        open={stopSeriesOpen}
        title="Stop recurring series?"
        message={`This will stop generating new sessions for "${session.title}". Existing sessions will remain and can still be individually managed.`}
        confirmLabel="Stop Recurring"
        onConfirm={() => {
          setStopSeriesOpen(false);
          if (session.series_id) {
            stopSeriesMutation.mutate(session.series_id);
          }
        }}
        onCancel={() => setStopSeriesOpen(false)}
      />
    </div>
  );
}
