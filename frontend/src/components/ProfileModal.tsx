import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import type { PublicMember } from '../types';

interface ProfileModalProps {
  open: boolean;
  memberId: string | null;
  onClose: () => void;
}

export default function ProfileModal({ open, memberId, onClose }: ProfileModalProps) {
  const [profile, setProfile] = useState<PublicMember | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!open || !memberId) return;
    setLoading(true);
    setProfile(null);
    api
      .getPublicProfile(memberId)
      .then((res) => setProfile(res.data))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [open, memberId]);

  useEffect(() => {
    if (!open) return;
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', handleKey);
    return () => document.removeEventListener('keydown', handleKey);
  }, [open, onClose]);

  if (!open) return null;

  const telegramHandle = profile?.telegram_handle?.replace(/^@/, '');

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-label="Member profile"
    >
      <div
        className="mx-4 w-full max-w-md rounded-xl bg-white p-6 shadow-xl dark:bg-stone-800"
        onClick={(e) => e.stopPropagation()}
      >
        {loading && (
          <div className="flex justify-center py-8">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-amber-600 border-t-transparent" />
          </div>
        )}

        {!loading && profile && (
          <>
            <h2 className="text-lg font-semibold text-stone-900 dark:text-stone-100">{profile.name}</h2>

            {profile.bio && (
              <p className="mt-2 text-sm text-stone-600 dark:text-stone-400 line-clamp-3">
                {profile.bio}
              </p>
            )}

            {profile.skills.length > 0 && (
              <div className="mt-3 flex flex-wrap gap-1.5">
                {profile.skills.map((skill) => (
                  <span
                    key={skill}
                    className="inline-flex items-center rounded-full bg-amber-100 px-2.5 py-0.5 text-xs font-medium text-amber-800 dark:bg-amber-900/30 dark:text-amber-300"
                  >
                    {skill}
                  </span>
                ))}
              </div>
            )}

            {(profile.github_username || profile.linkedin_url || telegramHandle || profile.instagram_handle) && (
              <div className="mt-3 flex flex-wrap gap-3">
                {profile.github_username && (
                  <a
                    href={`https://github.com/${profile.github_username.replace(/^@/, '')}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
                  >
                    GitHub
                  </a>
                )}
                {profile.linkedin_url && (
                  <a
                    href={profile.linkedin_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
                  >
                    LinkedIn
                  </a>
                )}
                {telegramHandle && (
                  <a
                    href={`https://t.me/${telegramHandle}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
                  >
                    Telegram
                  </a>
                )}
                {profile.instagram_handle && (
                  <a
                    href={`https://instagram.com/${profile.instagram_handle.replace(/^@/, '')}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
                  >
                    Instagram
                  </a>
                )}
              </div>
            )}

            <div className="mt-4 border-t border-stone-100 pt-4 dark:border-stone-700">
              <Link
                to={`/profile/${profile.id}`}
                onClick={onClose}
                className="text-sm font-medium text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
              >
                View full profile
              </Link>
            </div>
          </>
        )}

        {!loading && !profile && (
          <p className="py-4 text-center text-sm text-stone-500 dark:text-stone-400">
            Could not load profile.
          </p>
        )}
      </div>
    </div>
  );
}
