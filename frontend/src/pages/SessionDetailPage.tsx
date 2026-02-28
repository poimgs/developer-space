import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { api, ApiError } from '../api/client';
import ConfirmModal from '../components/ConfirmModal';
import GuestList from '../components/GuestList';
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

  if (isLoading || !session) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-indigo-600 border-t-transparent" />
      </div>
    );
  }

  const isCanceled = session.status === 'canceled';
  const isFull = session.rsvp_count >= session.capacity;

  return (
    <div className="mx-auto max-w-2xl">
      <Link
        to="/"
        className="mb-4 inline-flex items-center text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
      >
        <svg className="mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M15 19l-7-7 7-7" />
        </svg>
        Back to sessions
      </Link>

      <div className={`rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800 ${isCanceled ? 'opacity-60' : ''}`}>
        <div className="flex items-start justify-between gap-2">
          <h1 className={`text-xl font-bold text-gray-900 dark:text-gray-100 ${isCanceled ? 'line-through' : ''}`}>
            {session.title}
          </h1>
          <StatusBadge status={session.status} />
        </div>

        {session.description && (
          <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">{session.description}</p>
        )}

        <div className="mt-4 space-y-2 text-sm text-gray-600 dark:text-gray-300">
          <p>{formatDateLabel(session.date)}</p>
          <p>{session.start_time} – {session.end_time}</p>
          <p>
            {isFull && !session.user_rsvped ? (
              <span className="font-medium text-red-600 dark:text-red-400">Full</span>
            ) : (
              `${session.rsvp_count}/${session.capacity} spots`
            )}
          </p>
        </div>

        {!isCanceled && (
          <div className="mt-6 flex items-center gap-3">
            {session.user_rsvped ? (
              <button
                onClick={() => setConfirmOpen(true)}
                disabled={cancelRSVPMutation.isPending}
                className="rounded-md border border-indigo-600 px-4 py-2 text-sm font-medium text-indigo-600 transition-colors hover:bg-indigo-50 disabled:opacity-50 dark:border-indigo-400 dark:text-indigo-400 dark:hover:bg-indigo-900/30"
              >
                Cancel RSVP
              </button>
            ) : isFull ? (
              <span className="inline-block rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-400 dark:bg-gray-700 dark:text-gray-500">
                Full
              </span>
            ) : (
              <button
                onClick={() => rsvpMutation.mutate()}
                disabled={rsvpMutation.isPending}
                className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:opacity-50"
              >
                RSVP
              </button>
            )}

            {isAdmin && (
              <>
                <Link
                  to={`/sessions/${session.id}/edit`}
                  className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
                >
                  Edit
                </Link>
                <button
                  onClick={() => setCancelSessionOpen(true)}
                  className="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-red-700"
                >
                  Cancel Session
                </button>
              </>
            )}
          </div>
        )}
      </div>

      <div className="mt-6 rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Guest List</h2>
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
    </div>
  );
}
