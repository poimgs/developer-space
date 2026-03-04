import { useState } from 'react';
import type { RSVPWithMember } from '../types';
import ProfileModal from './ProfileModal';

interface GuestListProps {
  rsvps: RSVPWithMember[];
}

export default function GuestList({ rsvps }: GuestListProps) {
  const [selectedMemberId, setSelectedMemberId] = useState<string | null>(null);
  const [modalOpen, setModalOpen] = useState(false);

  if (rsvps.length === 0) {
    return (
      <p className="py-4 text-center text-sm text-stone-500 dark:text-stone-400">
        No RSVPs yet. Be the first!
      </p>
    );
  }

  return (
    <>
      <ul className="divide-y divide-stone-100 dark:divide-stone-700">
        {rsvps.map((rsvp) => (
          <li key={rsvp.id} className="flex items-center justify-between py-3">
            <button
              type="button"
              onClick={() => {
                setSelectedMemberId(rsvp.member.id);
                setModalOpen(true);
              }}
              className="text-sm font-medium text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
            >
              {rsvp.member.name}
            </button>
            {rsvp.member.bio && (
              <span className="text-sm text-stone-500 dark:text-stone-400 truncate max-w-[200px]">
                {rsvp.member.bio}
              </span>
            )}
          </li>
        ))}
      </ul>
      <ProfileModal
        open={modalOpen}
        memberId={selectedMemberId}
        onClose={() => setModalOpen(false)}
      />
    </>
  );
}
