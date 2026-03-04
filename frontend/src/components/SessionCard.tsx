import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import type { RSVPWithMember, SpaceSession } from '../types';
import ConfirmModal from './ConfirmModal';
import ProfileModal from './ProfileModal';

interface SessionCardProps {
  session: SpaceSession;
  attendees?: RSVPWithMember[];
  onRSVP: (sessionId: string) => void;
  onCancelRSVP: (sessionId: string) => void;
  onCancelSession?: (sessionId: string) => void;
  rsvpLoading?: boolean;
  variant?: 'default' | 'hero';
}

export default function SessionCard({ session, attendees, onRSVP, onCancelRSVP, onCancelSession, rsvpLoading, variant = 'default' }: SessionCardProps) {
  const { user } = useAuth();
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [cancelSessionOpen, setCancelSessionOpen] = useState(false);
  const [profileMemberId, setProfileMemberId] = useState<string | null>(null);
  const [profileModalOpen, setProfileModalOpen] = useState(false);

  const isCanceled = session.status === 'canceled';
  const isAdmin = user?.is_admin ?? false;
  const isHero = variant === 'hero';

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
        className={`mx-auto max-w-2xl rounded-xl border border-stone-200 bg-white shadow-sm dark:border-stone-700 dark:bg-stone-800 ${isCanceled ? 'opacity-60' : ''} ${!isHero ? 'p-5' : ''}`}
      >
        {/* Hero image */}
        {isHero && session.image_url && (
          <img
            src={session.image_url}
            alt={session.title}
            className="h-48 w-full rounded-t-xl object-cover"
          />
        )}

        <div className={isHero ? 'p-5' : ''}>
          {/* Header: Title + Status */}
          <div className="flex items-start justify-between gap-2">
            <div className="flex items-center gap-1.5 min-w-0">
              <Link
                to={`/sessions/${session.id}`}
                className={`font-semibold text-stone-900 hover:text-amber-600 dark:text-stone-100 dark:hover:text-amber-400 truncate ${isCanceled ? 'line-through' : ''} ${isHero ? 'text-lg' : 'text-base'}`}
              >
                {session.title}
              </Link>
              {session.series_id && (
                <span title="Recurring session">
                  <svg className="h-4 w-4 flex-shrink-0 text-amber-500 dark:text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                </span>
              )}
            </div>
          </div>

          {/* Time */}
          <p className="mt-2 text-sm text-stone-500 dark:text-stone-400">
            {session.start_time} – {session.end_time}
          </p>

          {/* Location */}
          {session.location && (
            <p className="mt-1 flex items-center gap-1 text-sm text-stone-500 dark:text-stone-400">
              <svg className="h-4 w-4 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M15 10.5a3 3 0 11-6 0 3 3 0 016 0z" />
                <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z" />
              </svg>
              {session.location}
            </p>
          )}

          {/* Description (hero only) */}
          {isHero && session.description && (
            <p className="mt-2 text-sm text-stone-600 dark:text-stone-400 line-clamp-2">
              {session.description}
            </p>
          )}

          {/* Attendance */}
          <div className="mt-2">
            <span className="text-sm text-stone-600 dark:text-stone-300">
              {session.rsvp_count} attending
            </span>
          </div>

          {/* Actions */}
          <div className="mt-3 flex items-center justify-end gap-2">
            {!isCanceled && (
              <>
                {session.user_rsvped ? (
                  <button
                    onClick={handleRSVPClick}
                    disabled={rsvpLoading}
                    className="rounded-md border border-amber-600 px-3 py-1.5 text-sm font-medium text-amber-600 transition-colors hover:bg-amber-50 disabled:opacity-50 dark:border-amber-400 dark:text-amber-400 dark:hover:bg-amber-900/30"
                  >
                    Cancel RSVP
                  </button>
                ) : (
                  <button
                    onClick={handleRSVPClick}
                    disabled={rsvpLoading}
                    className="rounded-md bg-amber-600 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-amber-700 disabled:opacity-50"
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
                  className="rounded p-1 text-stone-400 transition-colors hover:text-stone-600 dark:hover:text-stone-300"
                  aria-label="Edit session"
                >
                  <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </Link>
                <button
                  onClick={() => setCancelSessionOpen(true)}
                  className="rounded p-1 text-stone-400 transition-colors hover:text-red-600 dark:hover:text-red-400"
                  aria-label="Cancel session"
                >
                  <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </>
            )}
          </div>

          {/* Attendees */}
          {!isCanceled && attendees && attendees.length > 0 && (
            <div className="mt-3 border-t border-stone-100 pt-3 dark:border-stone-700">
              <p className="mb-1.5 text-xs font-medium text-stone-500 dark:text-stone-400">Who's going</p>
              <div className="flex flex-wrap gap-1.5">
                {attendees.slice(0, 5).map((rsvp) => (
                  <button
                    key={rsvp.id}
                    type="button"
                    title={rsvp.member.bio || rsvp.member.name}
                    onClick={() => {
                      setProfileMemberId(rsvp.member.id);
                      setProfileModalOpen(true);
                    }}
                    className="inline-flex cursor-pointer items-center rounded-full bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-900/30 dark:text-amber-300"
                  >
                    {rsvp.member.name}
                  </button>
                ))}
                {attendees.length > 5 && (
                  <span className="inline-flex items-center rounded-full bg-stone-100 px-2.5 py-0.5 text-xs font-medium text-stone-500 dark:bg-stone-700 dark:text-stone-400">
                    +{attendees.length - 5} more
                  </span>
                )}
              </div>
            </div>
          )}
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

      <ProfileModal
        open={profileModalOpen}
        memberId={profileMemberId}
        onClose={() => setProfileModalOpen(false)}
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
