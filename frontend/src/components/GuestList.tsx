import { Link } from 'react-router-dom';
import type { RSVPWithMember } from '../types';

interface GuestListProps {
  rsvps: RSVPWithMember[];
}

export default function GuestList({ rsvps }: GuestListProps) {
  if (rsvps.length === 0) {
    return (
      <p className="py-4 text-center text-sm text-stone-500 dark:text-stone-400">
        No RSVPs yet. Be the first!
      </p>
    );
  }

  return (
    <ul className="divide-y divide-stone-100 dark:divide-stone-700">
      {rsvps.map((rsvp) => (
        <li key={rsvp.id} className="flex items-center justify-between py-3">
          <Link
            to={`/profile/${rsvp.member.id}`}
            className="text-sm font-medium text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
          >
            {rsvp.member.name}
          </Link>
          {rsvp.member.telegram_handle && (
            <span className="text-sm text-stone-500 dark:text-stone-400">
              @{rsvp.member.telegram_handle.replace(/^@/, '')}
            </span>
          )}
        </li>
      ))}
    </ul>
  );
}
