import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { api, ApiError } from '../api/client';
import { useAuth } from '../context/AuthContext';
import type { PublicMember } from '../types';

export default function MemberProfilePage() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const [profile, setProfile] = useState<PublicMember | null>(null);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    api
      .getPublicProfile(id)
      .then((res) => {
        setProfile(res.data);
        setNotFound(false);
      })
      .catch((err) => {
        if (err instanceof ApiError && err.status === 404) {
          setNotFound(true);
        }
      })
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="mx-auto max-w-lg px-4 py-8">
        <p className="text-sm text-stone-500 dark:text-stone-400">Loading profile…</p>
      </div>
    );
  }

  if (notFound || !profile) {
    return (
      <div className="mx-auto max-w-lg px-4 py-8 text-center">
        <h1 className="text-xl font-bold text-stone-900 dark:text-stone-100">Member not found</h1>
        <p className="mt-2 text-sm text-stone-500 dark:text-stone-400">
          This member doesn't exist or their account is inactive.
        </p>
        <Link
          to="/"
          className="mt-4 inline-block text-sm font-medium text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
        >
          Back to sessions
        </Link>
      </div>
    );
  }

  const isSelf = user?.id === profile.id;

  return (
    <div className="mx-auto max-w-lg px-4 py-8">
      <h1 className="text-2xl font-bold text-stone-900 dark:text-stone-100">{profile.name}</h1>

      {profile.telegram_handle && (
        <p className="mt-1 text-sm text-stone-500 dark:text-stone-400">
          @{profile.telegram_handle.replace(/^@/, '')}
        </p>
      )}

      {isSelf && (
        <Link
          to="/profile"
          className="mt-2 inline-block text-sm font-medium text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
        >
          Edit Profile
        </Link>
      )}

      {profile.bio && (
        <section className="mt-6">
          <h2 className="text-lg font-semibold text-stone-900 dark:text-stone-100">About</h2>
          <p className="mt-2 whitespace-pre-wrap text-sm text-stone-700 dark:text-stone-300">
            {profile.bio}
          </p>
        </section>
      )}

      {profile.skills.length > 0 && (
        <section className="mt-6">
          <h2 className="text-lg font-semibold text-stone-900 dark:text-stone-100">Skills</h2>
          <div className="mt-2 flex flex-wrap gap-2">
            {profile.skills.map((skill) => (
              <span
                key={skill}
                className="inline-flex items-center rounded-full bg-amber-100 px-3 py-1 text-sm font-medium text-amber-800 dark:bg-amber-900 dark:text-amber-200"
              >
                {skill}
              </span>
            ))}
          </div>
        </section>
      )}

      {(profile.linkedin_url || profile.instagram_handle || profile.github_username) && (
        <section className="mt-6">
          <h2 className="text-lg font-semibold text-stone-900 dark:text-stone-100">Links</h2>
          <ul className="mt-2 space-y-2">
            {profile.linkedin_url && (
              <li>
                <a
                  href={profile.linkedin_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
                >
                  LinkedIn
                </a>
              </li>
            )}
            {profile.instagram_handle && (
              <li>
                <a
                  href={`https://instagram.com/${profile.instagram_handle.replace(/^@/, '')}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
                >
                  Instagram
                </a>
              </li>
            )}
            {profile.github_username && (
              <li>
                <a
                  href={`https://github.com/${profile.github_username.replace(/^@/, '')}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
                >
                  GitHub
                </a>
              </li>
            )}
          </ul>
        </section>
      )}
    </div>
  );
}
