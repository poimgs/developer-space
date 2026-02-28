import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api, ApiError } from '../api/client';
import SessionCard, { formatDateLabel } from '../components/SessionCard';
import EmptyState from '../components/EmptyState';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import type { APIResponse, RSVP, SpaceSession } from '../types';

export default function SessionsPage() {
  const queryClient = useQueryClient();
  const { addToast } = useToast();
  const { user } = useAuth();
  const isAdmin = user?.is_admin ?? false;

  const { data: sessions = [], isLoading } = useQuery({
    queryKey: ['sessions'],
    queryFn: async () => {
      const res = await api.get<APIResponse<SpaceSession[]>>('/api/sessions');
      return res.data;
    },
  });

  const rsvpMutation = useMutation({
    mutationFn: (sessionId: string) =>
      api.post<APIResponse<RSVP>>(`/api/sessions/${sessionId}/rsvp`),
    onMutate: async (sessionId) => {
      await queryClient.cancelQueries({ queryKey: ['sessions'] });
      const previous = queryClient.getQueryData<SpaceSession[]>(['sessions']);
      queryClient.setQueryData<SpaceSession[]>(['sessions'], (old) =>
        old?.map((s) =>
          s.id === sessionId
            ? { ...s, user_rsvped: true, rsvp_count: s.rsvp_count + 1 }
            : s,
        ),
      );
      return { previous };
    },
    onSuccess: () => {
      addToast("You're in!", 'success');
    },
    onError: (err, _sessionId, context) => {
      queryClient.setQueryData(['sessions'], context?.previous);
      const message = err instanceof ApiError ? err.message : 'Failed to RSVP';
      addToast(message, 'error');
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] });
    },
  });

  const cancelRSVPMutation = useMutation({
    mutationFn: (sessionId: string) =>
      api.delete(`/api/sessions/${sessionId}/rsvp`),
    onMutate: async (sessionId) => {
      await queryClient.cancelQueries({ queryKey: ['sessions'] });
      const previous = queryClient.getQueryData<SpaceSession[]>(['sessions']);
      queryClient.setQueryData<SpaceSession[]>(['sessions'], (old) =>
        old?.map((s) =>
          s.id === sessionId
            ? { ...s, user_rsvped: false, rsvp_count: Math.max(0, s.rsvp_count - 1) }
            : s,
        ),
      );
      return { previous };
    },
    onSuccess: () => {
      addToast('RSVP canceled.', 'info');
    },
    onError: (err, _sessionId, context) => {
      queryClient.setQueryData(['sessions'], context?.previous);
      const message = err instanceof ApiError ? err.message : 'Failed to cancel RSVP';
      addToast(message, 'error');
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] });
    },
  });

  const cancelSessionMutation = useMutation({
    mutationFn: (sessionId: string) =>
      api.delete(`/api/sessions/${sessionId}`),
    onSuccess: () => {
      addToast('Session canceled.', 'success');
      queryClient.invalidateQueries({ queryKey: ['sessions'] });
    },
    onError: (err) => {
      const message = err instanceof ApiError ? err.message : 'Failed to cancel session';
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

  if (sessions.length === 0) {
    return (
      <div>
        {isAdmin && (
          <div className="mb-6 flex justify-end">
            <Link
              to="/sessions/new"
              className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700"
            >
              Create Session
            </Link>
          </div>
        )}
        <EmptyState
          icon={
            <svg className="h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" />
            </svg>
          }
          heading="No upcoming sessions"
          subtext={isAdmin ? 'Create the first session to get started.' : 'Check back later for new sessions.'}
        />
      </div>
    );
  }

  // Group sessions by date
  const grouped = groupByDate(sessions);

  return (
    <div className="space-y-6">
      {isAdmin && (
        <div className="flex justify-end">
          <Link
            to="/sessions/new"
            className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700"
          >
            Create Session
          </Link>
        </div>
      )}
      {grouped.map(({ date, sessions: dateSessions }) => (
        <div key={date}>
          <h2 className="sticky top-0 z-10 bg-gray-50 py-2 text-sm font-semibold text-gray-900 dark:bg-gray-900 dark:text-gray-100 md:static">
            {formatDateLabel(date)}
          </h2>
          <div className="mt-2 space-y-3">
            {dateSessions.map((session) => (
              <SessionCard
                key={session.id}
                session={session}
                onRSVP={(id) => rsvpMutation.mutate(id)}
                onCancelRSVP={(id) => cancelRSVPMutation.mutate(id)}
                onCancelSession={(id) => cancelSessionMutation.mutate(id)}
                rsvpLoading={rsvpMutation.isPending || cancelRSVPMutation.isPending}
              />
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

function groupByDate(sessions: SpaceSession[]): { date: string; sessions: SpaceSession[] }[] {
  const map = new Map<string, SpaceSession[]>();
  for (const session of sessions) {
    const existing = map.get(session.date) ?? [];
    existing.push(session);
    map.set(session.date, existing);
  }
  return Array.from(map.entries()).map(([date, sessions]) => ({ date, sessions }));
}
