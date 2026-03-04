import { useMemo, useState } from 'react';
import { useMutation, useQueries, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api, ApiError } from '../api/client';
import DateStrip, { type DateChip } from '../components/DateStrip';
import SessionCard from '../components/SessionCard';
import EmptyState from '../components/EmptyState';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import type { APIResponse, RSVP, RSVPWithMember, SpaceSession } from '../types';

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

  // Derive date chips and default selection
  const dateChips = useMemo(() => {
    const map = new Map<string, number>();
    for (const s of sessions) {
      map.set(s.date, (map.get(s.date) ?? 0) + 1);
    }
    return Array.from(map.entries())
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([date, sessionCount]): DateChip => ({ date, sessionCount }));
  }, [sessions]);

  const defaultDate = useMemo(() => {
    if (dateChips.length === 0) return '';
    const today = new Date().toISOString().split('T')[0];
    const upcoming = dateChips.find((c) => c.date >= today);
    return upcoming?.date ?? dateChips[dateChips.length - 1].date;
  }, [dateChips]);

  const [selectedDate, setSelectedDate] = useState<string>('');
  const activeDate = selectedDate || defaultDate;

  // Derive available months from date chips
  const availableMonths = useMemo(() => {
    const months = new Set<string>();
    for (const chip of dateChips) {
      months.add(chip.date.slice(0, 7));
    }
    return Array.from(months).sort();
  }, [dateChips]);

  const defaultMonth = useMemo(() => defaultDate.slice(0, 7), [defaultDate]);
  const [selectedMonth, setSelectedMonth] = useState<string>('');
  const activeMonth = selectedMonth || defaultMonth;

  // Filter date chips to active month
  const monthDateChips = useMemo(
    () => dateChips.filter((c) => c.date.startsWith(activeMonth)),
    [dateChips, activeMonth],
  );

  // Format month label (e.g., "March 2026")
  const monthLabel = useMemo(() => {
    if (!activeMonth) return '';
    const d = new Date(activeMonth + '-01T00:00:00');
    return d.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });
  }, [activeMonth]);

  const monthIndex = availableMonths.indexOf(activeMonth);

  function handlePrevMonth() {
    if (monthIndex > 0) {
      const newMonth = availableMonths[monthIndex - 1];
      setSelectedMonth(newMonth);
      const firstDate = dateChips.find((c) => c.date.startsWith(newMonth));
      if (firstDate) setSelectedDate(firstDate.date);
    }
  }

  function handleNextMonth() {
    if (monthIndex < availableMonths.length - 1) {
      const newMonth = availableMonths[monthIndex + 1];
      setSelectedMonth(newMonth);
      const firstDate = dateChips.find((c) => c.date.startsWith(newMonth));
      if (firstDate) setSelectedDate(firstDate.date);
    }
  }

  // Sessions for the selected date
  const selectedSessions = useMemo(
    () => sessions.filter((s) => s.date === activeDate),
    [sessions, activeDate],
  );

  // Fetch RSVPs for all displayed sessions in parallel
  const rsvpQueries = useQueries({
    queries: sessions.map((session) => ({
      queryKey: ['rsvps', session.id],
      queryFn: async () => {
        const res = await api.get<APIResponse<RSVPWithMember[]>>(`/api/sessions/${session.id}/rsvps`);
        return res.data;
      },
      staleTime: 60_000,
      enabled: session.status !== 'canceled',
    })),
  });

  // Build attendees lookup map
  const attendeesMap = new Map<string, RSVPWithMember[]>();
  sessions.forEach((session, i) => {
    const query = rsvpQueries[i];
    if (query?.data) {
      attendeesMap.set(session.id, query.data);
    }
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
    onSettled: (_data, _err, sessionId) => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] });
      queryClient.invalidateQueries({ queryKey: ['rsvps', sessionId] });
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
    onSettled: (_data, _err, sessionId) => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] });
      queryClient.invalidateQueries({ queryKey: ['rsvps', sessionId] });
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
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-amber-600 border-t-transparent" />
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
              className="rounded-md bg-amber-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-amber-700"
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

  return (
    <div className="space-y-4">
      {isAdmin && (
        <div className="flex justify-end">
          <Link
            to="/sessions/new"
            className="rounded-md bg-amber-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-amber-700"
          >
            Create Session
          </Link>
        </div>
      )}

      <DateStrip
        dates={monthDateChips}
        selected={activeDate}
        onSelect={setSelectedDate}
        monthLabel={monthLabel}
        onPrevMonth={handlePrevMonth}
        onNextMonth={handleNextMonth}
        prevDisabled={monthIndex <= 0}
        nextDisabled={monthIndex >= availableMonths.length - 1}
      />

      <div className="space-y-4">
        {selectedSessions.map((session) => (
          <SessionCard
            key={session.id}
            session={session}
            attendees={attendeesMap.get(session.id)}
            onRSVP={(id) => rsvpMutation.mutate(id)}
            onCancelRSVP={(id) => cancelRSVPMutation.mutate(id)}
            onCancelSession={(id) => cancelSessionMutation.mutate(id)}
            rsvpLoading={rsvpMutation.isPending || cancelRSVPMutation.isPending}
            variant="hero"
          />
        ))}
      </div>
    </div>
  );
}
