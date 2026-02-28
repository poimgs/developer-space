import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import type { SpaceSession } from '../types';
import ConfirmModal from './ConfirmModal';
import StatusBadge from './StatusBadge';

interface SessionCardProps {
  session: SpaceSession;
  onRSVP: (sessionId: string) => void;
  onCancelRSVP: (sessionId: string) => void;
  onCancelSession?: (sessionId: string) => void;
  rsvpLoading?: boolean;
}

export default function SessionCard({ session, onRSVP, onCancelRSVP, onCancelSession, rsvpLoading }: SessionCardProps) {
  const { user } = useAuth();
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [cancelSessionOpen, setCancelSessionOpen] = useState(false);

  const isCanceled = session.status === 'canceled';
  const isFull = session.rsvp_count >= session.capacity;
  const isAdmin = user?.is_admin ?? false;

  function handleRSVPClick() {
    if (session.user_rsvped) {
      setConfirmOpen(true);
    } else {
      onRSVP(session.id);
    }
  }

  function handleConfirmCancel() {
    setConfirmOpen(false);
    onCancelRSVP(session.id);
  }

  return (
    <>
      <div
        className={`rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800 ${isCanceled ? 'opacity-60' : ''}`}
      >
        <div className="flex items-start justify-between gap-2">
          <Link
            to={`/sessions/${session.id}`}
            className={`text-base font-semibold text-gray-900 hover:text-indigo-600 dark:text-gray-100 dark:hover:text-indigo-400 ${isCanceled ? 'line-through' : ''}`}
          >
            {session.title}
          </Link>
          <StatusBadge status={session.status} />
        </div>

        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {session.start_time} – {session.end_time}
        </p>

        <div className="mt-3 flex items-center justify-between">
          <span className="text-sm text-gray-600 dark:text-gray-300">
            {isFull && !session.user_rsvped ? (
              <span className="font-medium text-red-600 dark:text-red-400">Full</span>
            ) : (
              `${session.rsvp_count}/${session.capacity} spots`
            )}
          </span>

          <div className="flex items-center gap-2">
            {!isCanceled && (
              <>
                {session.user_rsvped ? (
                  <button
                    onClick={handleRSVPClick}
                    disabled={rsvpLoading}
                    className="rounded-md border border-indigo-600 px-3 py-1.5 text-sm font-medium text-indigo-600 transition-colors hover:bg-indigo-50 disabled:opacity-50 dark:border-indigo-400 dark:text-indigo-400 dark:hover:bg-indigo-900/30"
                  >
                    Cancel RSVP
                  </button>
                ) : isFull ? (
                  <span className="rounded-md bg-gray-100 px-3 py-1.5 text-sm font-medium text-gray-400 dark:bg-gray-700 dark:text-gray-500">
                    Full
                  </span>
                ) : (
                  <button
                    onClick={handleRSVPClick}
                    disabled={rsvpLoading}
                    className="rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:opacity-50"
                  >
                    RSVP
                  </button>
                )}
              </>
            )}

            {isAdmin && !isCanceled && (
              <>
                <Link
                  to={`/sessions/${session.id}/edit`}
                  className="rounded p-1 text-gray-400 transition-colors hover:text-gray-600 dark:hover:text-gray-300"
                  aria-label="Edit session"
                >
                  <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </Link>
                <button
                  onClick={() => setCancelSessionOpen(true)}
                  className="rounded p-1 text-gray-400 transition-colors hover:text-red-600 dark:hover:text-red-400"
                  aria-label="Cancel session"
                >
                  <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </>
            )}
          </div>
        </div>
      </div>

      <ConfirmModal
        open={confirmOpen}
        title="Cancel your RSVP?"
        message={`You'll lose your spot for "${session.title}" on ${formatDateLabel(session.date)}.`}
        confirmLabel="Cancel RSVP"
        onConfirm={handleConfirmCancel}
        onCancel={() => setConfirmOpen(false)}
      />

      <ConfirmModal
        open={cancelSessionOpen}
        title="Cancel this session?"
        message={`This will cancel "${session.title}" on ${formatDateLabel(session.date)}. All RSVPed members will be notified. This action cannot be undone.`}
        confirmLabel="Cancel Session"
        onConfirm={() => {
          setCancelSessionOpen(false);
          onCancelSession?.(session.id);
        }}
        onCancel={() => setCancelSessionOpen(false)}
      />
    </>
  );
}

export function formatDateLabel(dateStr: string): string {
  const date = new Date(dateStr + 'T00:00:00');
  return date.toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
  });
}
