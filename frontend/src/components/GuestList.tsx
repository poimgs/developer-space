import type { RSVPWithMember } from '../types';

interface GuestListProps {
  rsvps: RSVPWithMember[];
}

export default function GuestList({ rsvps }: GuestListProps) {
  if (rsvps.length === 0) {
    return (
      <p className="py-4 text-center text-sm text-gray-500 dark:text-gray-400">
        No RSVPs yet. Be the first!
      </p>
    );
  }

  return (
    <ul className="divide-y divide-gray-100 dark:divide-gray-700">
      {rsvps.map((rsvp) => (
        <li key={rsvp.id} className="flex items-center justify-between py-3">
          <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
            {rsvp.member.name}
          </span>
          {rsvp.member.telegram_handle && (
            <span className="text-sm text-gray-500 dark:text-gray-400">
              @{rsvp.member.telegram_handle.replace(/^@/, '')}
            </span>
          )}
        </li>
      ))}
    </ul>
  );
}
